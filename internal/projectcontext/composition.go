package projectcontext

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	sourceidentity "github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pkgversion"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
)

type compositionSkill struct {
	EffectiveSkill
	member contextlock.Member
	spec   *skillspec.Spec
	files  map[string][]byte
}
type compositionRequirement struct {
	kind, name, source, revision, tag, versionRange, directory string
	commands                                                   []string
	transport                                                  string
	environment                                                string
	lockedCommit                                               string
}
type compositionPlan struct {
	Preview         CompositionPreview
	RuntimeEntries  []string
	BuildKeys       []string
	ConcurrentState bool
}

// PreviewComposition performs only protected input reads. It acquires no lease
// and publishes neither a composition record nor generated agent-visible files.
func (s Store) PreviewComposition(ctx context.Context, req ComposeRequest) (CompositionPreview, error) {
	op, err := s.begin(false)
	if err != nil {
		return CompositionPreview{}, err
	}
	defer op.close()
	plan, err := op.compose(ctx, req)
	return plan.Preview, err
}
func (op *operation) compose(ctx context.Context, req ComposeRequest) (compositionPlan, error) {
	var result compositionPlan
	if err := ctx.Err(); err != nil {
		return result, err
	}
	r, err := op.registered(req.CheckoutID)
	if err != nil {
		return result, err
	}
	if err = req.Adapter.validate(r.Record.Environment); err != nil {
		return result, err
	}
	if req.Policy.Version != 1 || req.Policy.Memory != "off" {
		return result, fail("project_context_invalid", "effective policy must be version 1 with memory off")
	}
	if _, err = req.Policy.System.Policy(); err != nil {
		return result, err
	}
	admitted, err := op.admitted(req.CheckoutID, req.Selections)
	if err != nil {
		return result, err
	}
	out := CompositionPreview{Files: map[string][]byte{}, Modules: []OrderedModule{}, Skills: []EffectiveSkill{}}
	generation := Generation{"project-generation/1", "registration:" + r.Record.CheckoutID, identity(r.Record), []GenerationItem{}}
	for _, c := range admitted.Items {
		a := op.state.Approvals[r.Record.CheckoutID+":"+c.Item.ID]
		generation.Items = append(generation.Items, GenerationItem{c.Item.ID, c.ItemSHA256, "approval:" + hashID(identity(a.Record))})
	}
	out.Generation = generation
	precedence := contextmaterialize.Precedence{Winner: r.Precedence.Winner, Placement: r.Precedence.Placement}
	skills := map[string]compositionSkill{}
	servers := map[string]contextmaterialize.MCPServer{}
	mcpMembers := map[string]contextlock.Member{}
	requirements := []compositionRequirement{}
	packages := map[string]contextmaterialize.Package{}
	var baseLock *contextlock.Lock
	runtimeEntries := map[string]bool{}
	home := filepath.Dir(op.root)
	err = envprofile.VisitCompositionProfile(home, r.Record.Profile.Name, r.Record.Profile.LockSHA256, func(lock *contextlock.Lock, m contextlock.Member, root string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		baseLock = lock
		dir, err := os.OpenRoot(root)
		if err != nil {
			return err
		}
		defer dir.Close()
		capt, err := captureTree(dir, "", false)
		if err != nil {
			return err
		}
		// State pins bind captured bytes as well as the surrounding store read.
		if m.StateHash != "" && (m.Directory == "" || m.Directory == ".") {
			h, e := hashing.ContentSHA256Files(capt.Bytes, lock.ContentHashVersion())
			if e != nil || strings.TrimPrefix(h, "sha256:") != m.StateHash {
				return fail("project_context_drift", "captured profile member differs from pin")
			}
		}
		if m.Kind == "skill" && m.Commit != "" {
			runtimeEntries[m.Name+"/"+m.Commit] = true
		}
		switch m.Kind {
		case "context":
			man, e := contextpkg.ParseManifest(capt.Bytes["agent-context.json"])
			if e != nil {
				return e
			}
			if man.Name != m.Name {
				return fail("project_context_drift", "profile context identity changed")
			}
			packages[m.Name] = materialPackage(man, capt.Bytes)
			requirements = append(requirements, contextRequirements(man)...)
		case "skill":
			spec, e := skillspec.Load(root)
			if e != nil {
				return e
			}
			contentPin, e := hashing.ContentSHA256Files(capt.Bytes, hashing.VersionV2)
			if e != nil {
				return e
			}
			skills[m.Name] = compositionSkill{EffectiveSkill{"profile", m.Name, "state " + contentPin, m.Name, identity(capt.Manifest)}, m, spec, capt.Bytes}
		case "mcp":
			mcpMembers[m.Name] = m
			man, e := contextpkg.ParseMCP(capt.Bytes[contextpkg.MCPManifestName])
			if e != nil {
				return e
			}
			if man.Name != m.Name {
				return fail("project_context_drift", "profile MCP identity changed")
			}
			server := contextmaterialize.MCPServer{Name: m.Name, Transport: man.Server.Transport, Command: man.Server.Command, Args: man.Server.Args, URL: man.Server.URL, EnvNames: man.Server.EnvNames, Environments: man.Server.Environments}
			if e = mergeCompositionMCP(servers, server); e != nil {
				return e
			}
		default:
			return fail("project_context_invalid", "unknown profile member kind")
		}
		return nil
	})
	if err != nil {
		return result, fail("project_context_drift", "selected profile closure is unavailable or changed: "+err.Error())
	}
	if baseLock == nil {
		return result, fail("project_context_drift", "empty selected profile closure")
	}
	profileRoot, _, err := contextmaterialize.Monolithic(baseLock, "sha256:"+r.Record.Profile.LockSHA256, precedence, r.Record.Environment, packages)
	if err != nil {
		return result, err
	}
	profileSystem, _, _, err := contextmaterialize.SystemPrompt(baseLock, precedence, r.Record.Environment, packages, req.Policy.System)
	if err != nil {
		return result, err
	}
	if err = appendModules(&out, "profile", baseLock, packages, precedence, r.Record.Environment, req.Policy.System); err != nil {
		return result, err
	}
	projectRoots := [][]byte{}
	systemDocs := [][]byte{}
	if len(profileSystem) > 0 {
		systemDocs = append(systemDocs, profileSystem)
	}
	for _, c := range admitted.Items {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		switch c.Item.Type {
		case "skill":
			capt, e := op.get(c.Item.Payload.SnapshotRef, c.Item.Payload.SnapshotSHA256)
			if e != nil {
				return result, e
			}
			root := filepath.Join(op.root, "snapshots", c.Item.Payload.SnapshotSHA256, "files")
			spec, e := skillspec.Load(root)
			if e != nil {
				return result, e
			}
			next := compositionSkill{EffectiveSkill{"project", c.Item.ID, c.Members[0].Pin, c.Item.ID, identity(capt.Manifest)}, contextlock.Member{Kind: "skill", Name: c.Item.ID, StateHash: strings.TrimPrefix(c.Members[0].Pin, "state sha256:")}, spec, capt.Bytes}
			// Assignment substitutes the entire skill: no old context, command,
			// dependency or provider survives a different pin of the same identity.
			if old, ok := skills[c.Item.ID]; !ok || old.Pin != next.Pin {
				skills[c.Item.ID] = next
			}
			// Project snapshots remain protected through the retained generation; they are not runtime-cache entries.
		case "context":
			b, e := op.lockBytes(c.Item.Payload.LockRef, c.Item.Payload.LockSHA256)
			if e != nil {
				return result, e
			}
			lock, e := contextlock.Parse(b)
			if e != nil {
				return result, e
			}
			pkgs := map[string]contextmaterialize.Package{}
			for _, m := range c.Members {
				capt, e := op.get(m.SnapshotRef, m.SnapshotSHA256)
				if e != nil {
					return result, e
				}
				man, e := manifest(capt)
				if e != nil {
					return result, e
				}
				pkgs[m.Name] = materialPackage(man, capt.Bytes)
				// Protected context snapshots are retained by their immutable member records.
			}
			doc, _, e := contextmaterialize.Monolithic(lock, "sha256:"+c.Item.Payload.LockSHA256, precedence, r.Record.Environment, pkgs)
			if e != nil {
				return result, e
			}
			projectRoots = append(projectRoots, doc)
			sys, _, _, e := contextmaterialize.SystemPrompt(lock, precedence, r.Record.Environment, pkgs, req.Policy.System)
			if e != nil {
				return result, e
			}
			if len(sys) > 0 {
				systemDocs = append(systemDocs, sys)
			}
			if e = appendModules(&out, c.Item.ID, lock, pkgs, precedence, r.Record.Environment, req.Policy.System); e != nil {
				return result, e
			}
		default:
			return result, fail("project_context_unsupported", "item has no admitted composition arm")
		}
	}
	names := sortedKeys(skills)
	for _, name := range names {
		requirements = append(requirements, skillRequirements(skills[name].spec)...)
	}
	for i := range requirements {
		requirements[i].environment = r.Record.Environment
		for _, member := range baseLock.Members {
			if member.Kind == requirements[i].kind && member.Name == requirements[i].name {
				requirements[i].lockedCommit = member.Commit
			}
		}
	}
	if err = checkCompositionRequirements(requirements, skills, servers, mcpMembers); err != nil {
		return result, err
	}
	owners, err := compositionCommands(skills)
	if err != nil {
		return result, err
	}
	if len(owners) > 1 {
		return result, fail("project_context_unsupported", "leaf 1 enables at most one command")
	}
	adapter, err := envregistry.ByID(r.Record.Environment)
	if err != nil {
		return result, err
	}
	hasSurface := func(name string) bool {
		for _, v := range req.Adapter.Surfaces {
			if v == name {
				return true
			}
		}
		return false
	}
	root := JoinRootDocuments(profileRoot, projectRoots...)
	if len(root) > 0 {
		out.Files[adapter.RootTarget] = root
	}
	if len(systemDocs) > 0 {
		if r.Record.Environment == "muse" {
			return result, fail("project_context_unsupported", "adapter has no additive system channel")
		}
		out.Files[".agent-context/system-prompt.md"] = contextmaterialize.Join(systemDocs)
	}
	for _, name := range names {
		skill := skills[name]
		if !hasSurface("skills") {
			return result, fail("project_context_unsupported", "adapter has no skill surface")
		}
		if _, ok := skill.files["SKILL.md"]; !ok {
			return result, fail("project_context_drift", "skill has no SKILL.md")
		}
		excludes := whitelist.ContextExcludedRoots(skill.spec.RuntimeRoots, skill.spec.BuildRoots)
		for p, b := range skill.files {
			if whitelist.ContextPathIncluded(p, len(skill.spec.Commands) == 0, excludes) {
				out.Files[path.Join(adapter.SkillsDir, name, p)] = b
			}
		}
		out.Skills = append(out.Skills, skill.EffectiveSkill)
	}
	commandModes := map[string]bool{}
	for _, owner := range owners {
		if !hasSurface("commands") {
			return result, fail("project_context_unsupported", "adapter has no checked command surface")
		}
		skill := skills[owner.ItemID]
		command := skill.spec.Commands[owner.Name]
		// Build/system activation must supply its owning immutable artifact. This
		// composition lane neither builds binaries nor resolves machine PATH.
		if command.Type != "script" {
			return result, fail("project_context_unsupported", "command requires owning build/system artifact admission")
		}
		payload, ok := skill.files[command.UnixPath]
		if !ok {
			return result, fail("project_context_drift", "command artifact missing from approved skill")
		}
		p := path.Join(".agent-context/commands", owner.Name)
		out.Files[p] = payload
		commandModes[p] = true
	}
	mcpSet := []contextmaterialize.MCPServer{}
	for _, name := range sortedKeys(servers) {
		mcpSet = append(mcpSet, servers[name])
	}
	if len(mcpSet) > 0 {
		if !hasSurface("mcp") || !contextmaterialize.SupportsMCPFile(r.Record.Environment) {
			return result, fail("project_context_unsupported", "adapter cannot reproduce MCP surface")
		}
		p, b, w, e := contextmaterialize.MCPFile(r.Record.Environment, mcpSet)
		if e != nil {
			return result, e
		}
		if w {
			out.Files[p] = b
		}
	}
	policyObject := struct {
		Policy          CompositionPolicy
		Precedence      contextlock.Precedence
		Modules         []OrderedModule
		Skills          []EffectiveSkill
		Commands        []CommandOwner
		ConcurrentState bool
	}{req.Policy, r.Precedence, out.Modules, out.Skills, owners, req.ConcurrentState}
	policyBytes, err := canonical(policyObject)
	if err != nil {
		return result, err
	}
	out.PolicyRecord = policyBytes
	policyHash := digest(policyBytes)
	receipt, err := compositionReceipt(out.Files, commandModes, identity(req.Adapter), policyHash, r.Record.WorktreeRef)
	if err != nil {
		return result, err
	}

	omitted := []CompositionOmission{}
	for _, v := range admitted.Omitted {
		omitted = append(omitted, CompositionOmission{v.ItemID, v.Reason})
	}
	record := CompositionRecord{"project-composition/1", generation.RegistrationRef, generation.RegistrationSHA256, identity(generation), r.Record.Profile, r.Record.Environment, "adapter:" + hashID(identity(req.Adapter)), identity(req.Adapter), "policy:" + hashID(policyHash), policyHash, "off", r.Record.HomeRef, r.Record.WorktreeRef, identity(receipt), owners, omitted}
	b, err := canonical(record)
	if err != nil {
		return result, err
	}
	if _, err = ParseCompositionRecord(b); err != nil {
		return result, err
	}
	out.Record = record
	out.Receipt = receipt
	out.SHA256 = digest(b)
	out.Ref = "composition:" + hashID(out.SHA256)
	if _, err = op.admitted(req.CheckoutID, req.Selections); err != nil {
		return result, err
	}
	if err = verifyBinding(r); err != nil {
		return result, err
	}
	if err = op.policyFresh(r); err != nil {
		return result, err
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	result.Preview = out
	result.RuntimeEntries = sortedKeys(runtimeEntries)
	result.BuildKeys = []string{}
	result.ConcurrentState = req.ConcurrentState
	return result, nil
}
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func materialPackage(man *contextpkg.Manifest, files map[string][]byte) contextmaterialize.Package {
	p := contextmaterialize.Package{HasContext: man.HasContext}
	for _, m := range man.Modules {
		p.Modules = append(p.Modules, contextmaterialize.Module{Module: m, Bytes: files[path.Join("context", m.Path)]})
	}
	return p
}
func appendModules(out *CompositionPreview, item string, lock *contextlock.Lock, pkgs map[string]contextmaterialize.Package, p contextmaterialize.Precedence, env string, a contextmaterialize.Admission) error {
	order, err := contextmaterialize.EmittedOrder(lock, p)
	if err != nil {
		return err
	}
	return appendOrderedModules(out, item, order, pkgs, env, a, lock)
}
func appendOrderedModules(out *CompositionPreview, item string, order []contextlock.Member, pkgs map[string]contextmaterialize.Package, env string, a contextmaterialize.Admission, lock *contextlock.Lock) error {
	system, _ := contextmaterialize.ClassifySystemModules(lock, order, pkgs, env, a)
	allowed := map[string]bool{}
	for _, m := range system {
		allowed[m.Package+"/"+m.Module.Path] = true
	}
	for _, m := range order {
		for _, mod := range pkgs[m.Name].Modules {
			if !mod.Applies(env) {
				continue
			}
			if mod.Class == "system" && !allowed[m.Name+"/"+mod.Path] {
				continue
			}
			out.Modules = append(out.Modules, OrderedModule{item, m.Name, mod.Path, mod.Class, digest(mod.Bytes)})
		}
	}
	return nil
}
func contextRequirements(m *contextpkg.Manifest) []compositionRequirement {
	out := []compositionRequirement{}
	for kind, set := range map[string]map[string]contextpkg.Requirement{"skill": m.Skills, "mcp": m.MCP} {
		for name, r := range set {
			source, _ := sourceidentity.Parse(r.Git)
			out = append(out, compositionRequirement{kind: kind, name: name, source: source, revision: r.Revision, tag: r.Tag, versionRange: r.Range, directory: r.Directory, commands: r.Commands})
		}
	}
	return out
}
func skillRequirements(s *skillspec.Spec) []compositionRequirement {
	out := []compositionRequirement{}
	for name, r := range s.Requirements {
		source, _ := sourceidentity.Parse(r.Git)
		q := compositionRequirement{kind: "skill", name: name, source: source, directory: r.Directory, commands: r.Commands}
		if r.RefKind == "tag" {
			q.tag = r.RefValue
		} else {
			q.revision = r.RefValue
		}
		out = append(out, q)
	}
	for name := range s.McpServers {
		out = append(out, compositionRequirement{kind: "mcp", name: name, transport: s.McpServers[name].Transport})
	}
	for _, r := range s.Dependencies {
		if r.Type == "skill" {
			out = append(out, compositionRequirement{kind: "skill", name: r.Skill, commands: []string{r.Command}})
		}
	}
	return out
}
func checkCompositionRequirements(reqs []compositionRequirement, skills map[string]compositionSkill, servers map[string]contextmaterialize.MCPServer, mcpMembers map[string]contextlock.Member) error {
	for _, r := range reqs {

		var member contextlock.Member
		var skill compositionSkill
		if r.kind == "mcp" {
			server, ok := servers[r.name]
			if !ok {
				return fail("project_context_unsupported", "required MCP provider missing: "+r.name)
			}
			if r.transport != "" && r.transport != server.Transport {
				return fail("project_context_unsupported", "required MCP transport differs: "+r.name)
			}
			if r.environment != "" && !contextmaterialize.MCPApplies(server, r.environment) {
				return fail("project_context_unsupported", "required MCP environment differs: "+r.name)
			}
			member = mcpMembers[r.name]
		} else {
			var ok bool
			skill, ok = skills[r.name]
			if !ok {
				return fail("project_context_unsupported", "required skill provider missing: "+r.name)
			}
			member = skill.member
		}

		if r.source != "" && r.source != member.Source {
			return fail("project_context_unsupported", "required provider source differs: "+r.name)
		}
		dir := r.directory
		if dir == "" {
			dir = "."
		}
		got := member.Directory
		if got == "" {
			got = "."
		}
		if dir != got || (r.revision != "" && member.Commit != r.revision) {
			return fail("project_context_unsupported", "required provider pin differs: "+r.name)
		}
		if r.tag != "" {
			if r.lockedCommit == "" || member.Commit != r.lockedCommit {
				return fail("project_context_unsupported", "required provider tag differs: "+r.name)
			}
		}
		if r.versionRange != "" {
			v, e := pkgversion.ParseVersion(member.Version)
			constraint, e2 := pkgversion.ParseRange(r.versionRange)
			if e != nil || e2 != nil || !constraint.Satisfies(v) {
				return fail("project_context_unsupported", "required provider range differs: "+r.name)
			}
		}
		for _, c := range r.commands {
			if _, ok := skill.spec.Commands[c]; !ok {
				return fail("project_context_unsupported", "required provider command missing: "+c)
			}
		}
	}
	return nil
}
func compositionCommands(skills map[string]compositionSkill) ([]CommandOwner, error) {
	owners := map[string]CommandOwner{}
	for _, name := range sortedKeys(skills) {
		s := skills[name]
		for _, command := range sortedKeys(s.spec.Commands) {
			if prior, ok := owners[command]; ok && prior.ItemID != name {
				return nil, fail("project_context_unsupported", "command collision: "+command)
			}
			artifact := s.ArtifactSHA256
			if c := s.spec.Commands[command]; c.Type == "script" {
				if c.UnixPath == "" {
					return nil, fail("project_context_unsupported", "command has no applicable Unix artifact")
				}
				b, ok := s.files[c.UnixPath]
				if !ok {
					return nil, fail("project_context_drift", "command artifact missing")
				}
				artifact = digest(b)
			}
			owners[command] = CommandOwner{command, name, artifact}
		}
	}
	out := []CommandOwner{}
	for _, name := range sortedKeys(owners) {
		out = append(out, owners[name])
	}
	return out, nil
}
func mergeCompositionMCP(set map[string]contextmaterialize.MCPServer, next contextmaterialize.MCPServer) error {
	next.Args = append([]string{}, next.Args...)
	next.EnvNames = append([]string{}, next.EnvNames...)
	sort.Strings(next.EnvNames)
	if prior, ok := set[next.Name]; ok && identity(prior) != identity(next) {
		return fail("project_context_unsupported", "same-ID MCP replacement has no explicit approval: "+next.Name)
	}
	set[next.Name] = next
	return nil
}

func compositionReceipt(files map[string][]byte, executable map[string]bool, adapterHash, policyHash, worktree string) (MaterializationReceipt, error) {
	receipt := MaterializationReceipt{1, []File{}, adapterHash, policyHash, worktree}
	if len(files) > MaxFiles {
		return receipt, fail("project_context_invalid", "generated file bound exceeded")
	}
	for _, p := range sortedKeys(files) {
		b := files[p]
		mode := "0600"
		if executable[p] {
			mode = "0700"
		}
		receipt.Files = append(receipt.Files, File{p, int64(len(b)), digest(b), mode})
	}
	if err := validateInventory(captured{Manifest: Snapshot{1, receipt.Files}, Bytes: files}); err != nil {
		return MaterializationReceipt{}, err
	}
	return receipt, nil
}
