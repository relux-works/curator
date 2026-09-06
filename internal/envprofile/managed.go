// Package envprofile managed homes (environments §8.1, §7.4, §10.1): one manager-owned home
// per profile × environment below the environments root, provisioned on
// the first env resolve --repair naming it, verified lock-free on every
// resolve, and repaired under the manager-home mutation lock. Managed
// surfaces link into the immutable profile store with copies where a
// surface requires bytes (the always-copied claude_code root-context file,
// §8.1); every surface is recorded in the environment marker with its
// content hash, its form, and every copy's reason. Credential passthrough
// entries and provisioning seeds are recorded but never hashed; the tool
// owns them after provisioning.
package envprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Diagnostics for managed homes and resolution (environments §7.7, §8.5,
// §9.7, §10.4).
const (
	DiagHomeStale       = "environment_home_stale"
	DiagRepairFailed    = "environment_repair_failed"
	DiagForeignManager  = "environment_foreign_manager_detected"
	DiagForeignSuspect  = "environment_foreign_manager_suspected"
	DiagProfileUnknown  = "profile_unknown"
	DiagReservedCommand = "environment_reserved_command_name"
)

// EnvRootName is the manager-owned environments root below the manager
// home (environments §8.1).
const EnvRootName = "environments"

// EnvRoot returns the manager-owned environments root.
func EnvRoot(home string) string { return filepath.Join(home, EnvRootName) }

// ManagedParent returns the fragment env-value directory for a profile ×
// environment: the tool's home, except for opencode where the variable
// names the managed XDG parent and the tool reads the opencode child
// (environments §7.1).
func ManagedParent(home, profile, envID string) string {
	if envID == envregistry.OpenCode {
		return filepath.Join(EnvRoot(home), profile, envID)
	}
	return filepath.Join(EnvRoot(home), profile, envID)
}

// ManagedHomeDir returns the tool's home directory.
func ManagedHomeDir(home, profile, envID string) string {
	if envID == envregistry.OpenCode {
		return filepath.Join(ManagedParent(home, profile, envID), "opencode")
	}
	return ManagedParent(home, profile, envID)
}

// ResolveRequest is one env resolve or repair operation.
type ResolveRequest struct {
	Home      string
	Profile   string
	EnvID     string
	LaunchDir string
	Machine   envregistry.MachineConfig
	Repair    bool
	Format    string
	// Detect maps an adapter to its detected tool release, or "unknown".
	// Nil means probe the machine read-only.
	Detect func(adapter envregistry.Adapter) string
	// NativeHomeOf resolves native homes; nil means the process homes.
	// A test seam: production never sets it.
	NativeHomeOf func(id string) (string, error)
	// OperatorXDG is the operator's effective XDG config home; empty
	// means resolve it from the process environment.
	OperatorXDG string
}

// ResolveResult is the outcome: exactly one of Document and StaleErr is
// set; a first provisioning additionally carries the first-resolve notice.
type ResolveResult struct {
	Document     []byte
	Notice       string
	Warnings     []string
	Provisioned  bool
	StaleReasons []string
}

// versionToken finds the first dotted-decimal token in tool output.
var versionToken = regexp.MustCompile(`\d+(?:\.\d+)+`)

// detectRelease probes the adapter's tool read-only for its release
// (environments §7.9). Any failure reports "unknown", never a match.
func detectRelease(probe []string) string {
	if len(probe) == 0 {
		return "unknown"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, probe[0], probe[1:]...).Output() // #nosec G204 -- probe argv is registry data
	if err != nil {
		return "unknown"
	}
	if token := versionToken.FindString(string(out)); token != "" {
		return token
	}
	return "unknown"
}

// nativeHome resolves the adapter's native home through the seam.
func (req *ResolveRequest) nativeHome(id string) (string, error) {
	if req.NativeHomeOf != nil {
		return req.NativeHomeOf(id)
	}
	adapter, ok := adapterByID(id)
	if !ok {
		return "", fmt.Errorf("%s: unregistered environment %q", envregistry.DiagUnknown, id)
	}
	return NativeHome(adapter)
}

// detected resolves the tool release through the seam.
func (req *ResolveRequest) detected(adapter envregistry.Adapter) string {
	if req.Detect != nil {
		return req.Detect(adapter)
	}
	return detectRelease(adapter.Probe)
}

// operatorXDG resolves the operator's effective XDG config home.
func (req *ResolveRequest) operatorXDG() string {
	if req.OperatorXDG != "" {
		return req.OperatorXDG
	}
	if value := os.Getenv("XDG_CONFIG_HOME"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}

// mcpsOf loads the lock's MCP members from the store as resolved servers.
func mcpsOf(home string, manager *gitManager, lock *contextlock.Lock) ([]contextmaterialize.MCPServer, error) {
	var servers []contextmaterialize.MCPServer
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindMCP {
			continue
		}
		entry := manager.entryPath(home, resolvedOf(member))
		manifest, err := contextpkg.LoadMCP(packageRoot(entry, member.Directory))
		if err != nil {
			return nil, fmt.Errorf("%s: member %s %v", DiagRepairFailed, member.Name, err)
		}
		servers = append(servers, contextmaterialize.MCPServer{
			Name:         member.Name,
			Transport:    manifest.Server.Transport,
			Command:      manifest.Server.Command,
			Args:         manifest.Server.Args,
			URL:          manifest.Server.URL,
			EnvNames:     manifest.Server.EnvNames,
			Environments: manifest.Server.Environments,
		})
	}
	return servers, nil
}

// skillOf is one lock skill member with its store package root and content
// hash.
type skillOf struct {
	name string
	root string
	hash string
}

// skillsOf loads the lock's skill members: entry package roots with their
// store content hashes. A skill whose materialized name would collide with
// the umbrella discovery namespace is refused with
// environment_reserved_command_name (environments §9.4): profile-
// materialized files must not be able to poison the PATH the §11 dispatch
// trusts.
func skillsOf(home string, manager *gitManager, lock *contextlock.Lock) ([]skillOf, error) {
	var skills []skillOf
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindSkill {
			continue
		}
		if strings.HasPrefix(member.Name, "curator-") {
			return nil, fmt.Errorf("%s: skill %q collides with the curator-* name reservation", DiagReservedCommand, member.Name)
		}
		entry := manager.entryPath(home, resolvedOf(member))
		root := packageRoot(entry, member.Directory)
		hash, err := contextstore.ContentHash(root)
		if err != nil {
			return nil, fmt.Errorf("%s: member %s %v", DiagRepairFailed, member.Name, err)
		}
		skills = append(skills, skillOf{name: member.Name, root: root, hash: hash})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].name < skills[j].name })
	return skills, nil
}

// storeDocPath names the deterministic store document for manager-authored
// bytes: the same inputs always name the same path, so lock-free
// verification recomputes the expected link target without writing.
func storeDocPath(home, profile, envID, rel string) string {
	return filepath.Join(ProfilesDir(home), profile, "rendered", envID, filepath.FromSlash(rel))
}

// homePlan is the fully assembled managed home: every write, link, and
// removal, with the marker recording them.
type homePlan struct {
	adapter    envregistry.Adapter
	parent     string
	homeDir    string
	form       string
	isolation  string
	copies     map[string][]byte
	links      map[string]string
	docs       map[string][]byte
	marker     *envmarker.Marker
	fileHashes map[string][]byte
	warnings   []string
}

// assembleHome builds the desired managed home purely from the lock, the
// store entries it names, and machine configuration: no read of the home
// itself beyond the fallback and merge inputs the caller supplies.
func assembleHome(req *ResolveRequest, source Source, lock *contextlock.Lock, hash string, precedence contextmaterialize.Precedence, order []contextlock.Member, adapter envregistry.Adapter, prior *envmarker.Marker) (*homePlan, error) {
	manager := newGitManager(req.Home)
	packages, err := loadMaterial(req.Home, manager, lock)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	servers, err := mcpsOf(req.Home, manager, lock)
	if err != nil {
		return nil, err
	}
	skills, err := skillsOf(req.Home, manager, lock)
	if err != nil {
		return nil, err
	}
	form, err := req.Machine.EffectiveForm(adapter)
	if err != nil {
		return nil, err
	}
	// An unknown release warns (environment_tool_version_unverified, emitted
	// by verification) rather than blocks: the selection scheme stands on
	// the bundle evidence, so resolution proceeds as at-or-above.
	atAbove := true
	if ok, known := adapter.AtOrAbovePinned(req.detected(adapter)); known {
		atAbove = ok
	}
	isolation, err := req.Machine.EffectiveIsolation(req.Profile, adapter, atAbove)
	if err != nil {
		return nil, err
	}
	plan := &homePlan{
		adapter:    adapter,
		parent:     ManagedParent(req.Home, req.Profile, adapter.ID),
		homeDir:    ManagedHomeDir(req.Home, req.Profile, adapter.ID),
		form:       form,
		isolation:  isolation,
		copies:     map[string][]byte{},
		links:      map[string]string{},
		docs:       map[string][]byte{},
		fileHashes: map[string][]byte{},
	}
	if form == envregistry.FormReferenced && adapter.ID == envregistry.OpenCode {
		if blocked, err := referencedBlocked(plan.homeDir, prior); err != nil {
			return nil, err
		} else if blocked {
			plan.warnings = append(plan.warnings, contextmaterialize.DiagFormUnavailable+": opencode.json exists and no marker records it; monolithic emitted")
			form = envregistry.FormMonolithic
			plan.form = form
		}
	}
	surfaces := map[string]envmarker.Surface{}
	if err := plan.rootContext(req, lock, hash, precedence, packages, surfaces); err != nil {
		return nil, err
	}
	if err := plan.systemPrompt(req, lock, precedence, packages, surfaces); err != nil {
		return nil, err
	}
	plan.skills(req, skills, surfaces)
	if err := plan.mcp(req, servers, surfaces); err != nil {
		return nil, err
	}
	marker := &envmarker.Marker{
		Version: 1,
		Profile: envmarker.Profile{
			Name: req.Profile, Root: lock.Root, Kind: markerKind(source),
			LockSHA256: strings.TrimPrefix(hash, "sha256:"),
			Source:     markerSource(source), Requirement: markerRequirement(source),
			Directory: source.Directory, SourcePath: markerSourcePath(source),
		},
		Precedence: envmarker.Precedence{Winner: precedence.Winner, Placement: precedence.Placement},
		Mode:       envmarker.ModeManagedHome,
		Surfaces:   surfaces,
	}
	for _, member := range order {
		entry := envmarker.Member{Name: member.Name, Version: member.Version, Weight: member.Weight, Overlay: member.Overlay}
		if member.Commit != "" {
			entry.Commit = member.Commit
		} else {
			entry.StateSHA256 = member.StateHash
		}
		marker.Members = append(marker.Members, entry)
	}
	plan.marker = marker
	return plan, nil
}

// referencedBlocked reports whether an unmanaged opencode.json blocks the
// referenced form: the file exists and the preceding marker does not
// record it, so the adapter warns environment_form_unavailable and
// materializes monolithic instead of editing the unmanaged file (§5.3).
func referencedBlocked(homeDir string, prior *envmarker.Marker) (bool, error) {
	recorded := false
	if prior != nil {
		for _, surface := range prior.Surfaces {
			for _, path := range surface.Paths {
				if path == contextmaterialize.OpenCodeConfigName || path == "opencode.json" {
					recorded = true
				}
			}
		}
	}
	if recorded {
		return false, nil
	}
	_, err := os.Lstat(filepath.Join(homeDir, "opencode.json"))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// rootContext assembles the root-context surface: monolithic bytes or the
// referenced file set, linked into the store with copies where the surface
// requires bytes. The claude_code root file is always a copied regular
// file (§8.1); referenced module files link into the immutable store
// entries so link-target identity verifies them lock-free (§10.1).
func (p *homePlan) rootContext(req *ResolveRequest, lock *contextlock.Lock, hash string, precedence contextmaterialize.Precedence, packages map[string]contextmaterialize.Package, surfaces map[string]envmarker.Surface) error {
	target := p.adapter.RootTarget
	if p.form == envregistry.FormReferenced {
		files, written, err := contextmaterialize.Referenced(lock, hash, precedence, p.adapter.ID, packages)
		if err != nil {
			return err
		}
		if !written {
			return nil
		}
		paths := make([]string, 0, len(files))
		for path := range files {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		copies := []envmarker.Copy{}
		for _, path := range paths {
			document := files[path]
			p.fileHashes[path] = document
			if path == target && p.adapter.ID == envregistry.ClaudeCode {
				p.copies[path] = document
				copies = append(copies, envmarker.Copy{Path: path, Reason: envmarker.ReasonClaudeCodeRootContext})
				continue
			}
			if path == target || path == contextmaterialize.OpenCodeConfigName {
				storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, path)
				p.docs[storeDoc] = document
				p.links[path] = storeDoc
				continue
			}
			member, module := splitModulePath(path)
			link, err := storeModuleLink(req, lock, member, module)
			if err != nil {
				return err
			}
			p.links[path] = link
		}
		ordered := append([]string{}, paths...)
		surfaces[envmarker.SurfaceRootContext] = envmarker.Surface{
			Paths:         ordered,
			Form:          p.form,
			ContentSHA256: contextmaterialize.SurfaceHash(p.fileHashesFor(paths)),
			Copies:        &copies,
		}
		return nil
	}
	document, written, err := contextmaterialize.Monolithic(lock, hash, precedence, p.adapter.ID, packages)
	if err != nil {
		return err
	}
	if !written {
		return nil
	}
	p.fileHashes[target] = document
	copies := []envmarker.Copy{}
	if p.adapter.ID == envregistry.ClaudeCode {
		p.copies[target] = document
		copies = append(copies, envmarker.Copy{Path: target, Reason: envmarker.ReasonClaudeCodeRootContext})
	} else {
		storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, target)
		p.docs[storeDoc] = document
		p.links[target] = storeDoc
	}
	// The size advisory warns but changes nothing about the bytes (§5).
	if int64(len(document)) > p.adapter.SizeAdvisoryBytes {
		p.warnings = append(p.warnings, fmt.Sprintf("%s: root context %d bytes exceeds the %s advisory of %d", envregistry.DiagSizeExceeded, len(document), p.adapter.ID, p.adapter.SizeAdvisoryBytes))
	}
	surfaces[envmarker.SurfaceRootContext] = envmarker.Surface{
		Paths:         []string{target},
		Form:          p.form,
		ContentSHA256: contextmaterialize.SurfaceHash(map[string][]byte{target: document}),
		Copies:        &copies,
	}
	return nil
}

// fileHashesFor projects the plan's known bytes onto paths.
func (p *homePlan) fileHashesFor(paths []string) map[string][]byte {
	out := map[string][]byte{}
	for _, path := range paths {
		out[path] = p.fileHashes[path]
	}
	return out
}

// splitModulePath cuts a module file path into its package and module
// segments.
func splitModulePath(path string) (string, string) {
	rest, _ := strings.CutPrefix(path, contextmaterialize.ModulesDir+"/")
	name, module, _ := strings.Cut(rest, "/")
	return name, module
}

// storeModuleLink resolves the immutable store file a referenced module
// links into.
func storeModuleLink(req *ResolveRequest, lock *contextlock.Lock, member, module string) (string, error) {
	manager := newGitManager(req.Home)
	found, ok := lock.Find(contextlock.KindContext, member)
	if !ok {
		return "", fmt.Errorf("%s: lock carries no context member %s", DiagRepairFailed, member)
	}
	entry := manager.entryPath(req.Home, resolvedOf(found))
	return filepath.Join(packageRoot(entry, found.Directory), contextpkg.ContextDir, filepath.FromSlash(module)), nil
}

// publishDocs publishes the plan's manager-authored store documents. It
// runs only under the mutation lock: assembly itself is pure so that
// lock-free verification recomputes expected link targets without
// writing (environments §10.1, §12).
func (p *homePlan) publishDocs() error {
	paths := make([]string, 0, len(p.docs))
	for path := range p.docs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := writeStoreDoc(path, p.docs[path]); err != nil {
			return err
		}
	}
	return nil
}

// effectivePassthrough computes the passthrough links a managed home
// records: home-relative path to native absolute target, with the strategy
// (§7.4). isolated homes link nothing; keyring-backed codex homes are
// ambient and link nothing; every file-shaped strategy is watched by the
// liveness row. The native config.toml read is read-only: an absent file
// means the default file store, and an unreadable one is treated as file
// with no warning — the liveness row still watches the link.
func (req *ResolveRequest) effectivePassthrough(adapter envregistry.Adapter, isolation string) (map[string]string, map[string]string, error) {
	links := map[string]string{}
	strategies := map[string]string{}
	if isolation == envregistry.IsolationIsolated {
		return links, strategies, nil
	}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range adapter.PassthroughFor(runtime.GOOS) {
		switch entry.Strategy {
		case envregistry.StrategyPerHomeKeychain, envregistry.StrategyAmbient:
			continue
		case envregistry.StrategyKeyringPreferred:
			if codexKeyring(native) {
				continue
			}
			links[entry.Path] = filepath.Join(native, entry.FileLinkTarget)
			strategies[entry.Path] = envregistry.StrategyFileLink
		case envregistry.StrategyFileLink:
			links[entry.Path] = filepath.Join(native, entry.FileLinkTarget)
			strategies[entry.Path] = envregistry.StrategyFileLink
		}
	}
	return links, strategies, nil
}

// codexKeyring reports whether the operator's native config.toml selects
// the keyring credential store, in which case the credential is ambient
// and no entry is linked (§7.4).
func codexKeyring(native string) bool {
	payload, err := os.ReadFile(filepath.Join(native, "config.toml")) // #nosec G304 -- native home resolved from the registry
	if err != nil {
		return false
	}
	matches := keyringSetting.FindSubmatch(payload)
	return len(matches) == 2 && string(matches[1]) == "keyring"
}

var keyringSetting = regexp.MustCompile(`(?m)^\s*cli_auth_credentials_store\s*=\s*"([^"]*)"`)

// seedBundle is the one-time provisioning class (§7.4): non-credential
// files copied from the native home exactly once, at provisioning, never
// refreshed, never hashed.
type seedBundle struct {
	files map[string][]byte
	// xdg maps parent-relative names to operator absolute targets.
	xdg map[string]string
	// claudeInit marks the written (not copied) .claude.json seed.
	claudeInit bool
}

// gatherSeeds reads every seed upfront so that an unreadable seed stops
// provisioning before the first write. A seed absent in the native home
// is simply not seeded; absence and unreadability stay different facts
// (§8.4). Copy seeds are gathered at provisioning only: repair never
// refreshes them, so an unreadable native copy must not fail a repair
// that would never read it.
func (req *ResolveRequest) gatherSeeds(adapter envregistry.Adapter, provision bool) (*seedBundle, error) {
	bundle := &seedBundle{files: map[string][]byte{}, xdg: map[string]string{}}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return nil, err
	}
	if provision {
		for _, seed := range adapter.Seeds {
			if adapter.SeedWritten[seed] {
				bundle.claudeInit = true
				continue
			}
			payload, err := os.ReadFile(filepath.Join(native, filepath.FromSlash(seed))) // #nosec G304 -- seed names are registry data
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("%s: seed %s: %v", envregistry.DiagSeedUnreadable, seed, err)
			}
			bundle.files[seed] = payload
		}
	} else if adapter.ID == envregistry.ClaudeCode {
		bundle.claudeInit = true
	}
	if adapter.ID == envregistry.OpenCode {
		xdg := req.operatorXDG()
		if xdg == "" {
			return bundle, nil
		}
		for _, name := range req.Machine.XDGSeedAllowlist {
			if name == "opencode" {
				continue
			}
			target := filepath.Join(xdg, name)
			if _, err := os.Stat(target); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("%s: xdg seed %s: %v", envregistry.DiagSeedUnreadable, name, err)
			}
			bundle.xdg[name] = target
		}
	}
	return bundle, nil
}

// writeStoreDoc publishes manager-authored bytes as a store document.
func writeStoreDoc(path string, document []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, document, 0o644)
}

// claudeSeed merges the .claude.json provisioning seed: the exact object
// {"hasCompletedOnboarding":true,"projects":{}} at provisioning, plus one
// project entry per launch directory holding hasTrustDialogAccepted and —
// under the referenced form — hasClaudeMdExternalIncludesApproved (§7.4).
// Later tool writes are its own state: repair only adds the launch
// directory's entry and never rewrites anything else.
func claudeSeed(homeDir, launchDir, form string) (seeded []string, err error) {
	path := filepath.Join(homeDir, ".claude.json")
	object := map[string]any{"hasCompletedOnboarding": true, "projects": map[string]any{}}
	if payload, err := os.ReadFile(path); err == nil { // #nosec G304 -- managed .claude.json below the resolved home
		parsed, ok := decodeJSONObject(payload)
		if !ok {
			return nil, fmt.Errorf("%s: managed .claude.json is not an object", envregistry.DiagSeedUnreadable)
		}
		object = parsed
		if _, ok := object["projects"].(map[string]any); !ok {
			object["projects"] = map[string]any{}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("%s: managed .claude.json: %v", envregistry.DiagSeedUnreadable, err)
	}
	projects := object["projects"].(map[string]any)
	for name := range projects {
		seeded = append(seeded, name)
	}
	entry, _ := projects[launchDir].(map[string]any)
	if entry == nil {
		entry = map[string]any{}
	}
	entry["hasTrustDialogAccepted"] = true
	if form == envregistry.FormReferenced {
		entry["hasClaudeMdExternalIncludesApproved"] = true
	}
	projects[launchDir] = entry
	found := false
	for _, name := range seeded {
		if name == launchDir {
			found = true
		}
	}
	if !found {
		seeded = append(seeded, launchDir)
	}
	sort.Strings(seeded)
	payload, err := protocoljsonMarshal(object)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return nil, err
	}
	return seeded, nil
}

// decodeJSONObject decodes one JSON object value.
func decodeJSONObject(payload []byte) (map[string]any, bool) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, false
	}
	return object, true
}

// protocoljsonMarshal renders manager-authored JSON state as CCJ-1 plus
// one trailing LF.
func protocoljsonMarshal(object map[string]any) ([]byte, error) {
	document, err := protocoljson.MarshalCanonical(object)
	if err != nil {
		return nil, err
	}
	return append(document, '\n'), nil
}

// claudeProjects reads the managed .claude.json project entries for
// verification. A missing file means no entry; an unreadable or invalid
// file is reported, never treated as absence (§8.4).
func claudeProjects(homeDir string) (map[string]bool, error) {
	payload, err := os.ReadFile(filepath.Join(homeDir, ".claude.json")) // #nosec G304 -- managed .claude.json below the resolved home
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	parsed, ok := decodeJSONObject(payload)
	if !ok {
		return nil, fmt.Errorf("managed .claude.json is not an object")
	}
	entries := map[string]bool{}
	if projects, ok := parsed["projects"].(map[string]any); ok {
		for name := range projects {
			entries[name] = true
		}
	}
	return entries, nil
}

// dotfileStateDirs is the closed heuristic list for the §9.5 inventory: a
// well-known dotfile-manager state location elevates the notice for plain
// unmanaged files to environment_foreign_manager_suspected. The heuristic
// never blocks.
var dotfileStateDirs = [][2]string{
	{".local/share/chezmoi", "chezmoi"},
	{".config/home-manager", "home-manager"},
}

// foreignManagerHint scans the operator home for dotfile-manager state.
func foreignManagerHint() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, candidate := range dotfileStateDirs {
		if info, err := os.Stat(filepath.Join(home, filepath.FromSlash(candidate[0]))); err == nil && info.IsDir() {
			return candidate[1]
		}
	}
	return ""
}

// inventoryUnmanaged walks the desired home paths before the first write:
// a managed-surface path that is already a symlink pointing outside the
// manager's store is evidence of another manager and stops the operation
// with environment_foreign_manager_detected; any other unmanaged file
// fails with environment_surface_unmanaged_conflict (§8.3, §9.5).
func inventoryUnmanaged(homeDir string, want map[string]bool, storeRoot string) error {
	paths := make([]string, 0, len(want))
	for path := range want {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		full := filepath.Join(homeDir, filepath.FromSlash(path))
		info, err := os.Lstat(full)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("%s: inventory of %s: %v", DiagUnmanagedConflict, path, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(full)
			if err == nil && !sameStoreTree(target, storeRoot) {
				return fmt.Errorf("%s: %s is a symlink outside the manager store; abort, or take over with backup", DiagForeignManager, path)
			}
		}
		return fmt.Errorf("%s: %s exists and no marker records it", DiagUnmanagedConflict, path)
	}
	return nil
}

// sameStoreTree reports whether the link target lives below the store.
func sameStoreTree(target, storeRoot string) bool {
	if !filepath.IsAbs(target) {
		return false
	}
	root := filepath.Clean(storeRoot)
	return target == root || strings.HasPrefix(target, root+string(filepath.Separator))
}

// applyPlan writes the plan into the home under the held mutation lock:
// store documents, copies, links (with copy fallback for files), seed
// files, and the marker. Managed paths the plan no longer wants are
// removed; files the marker does not record are never touched.
func applyPlan(req *ResolveRequest, plan *homePlan, seeds *seedBundle, recorded map[string]bool, prior *envmarker.Marker, provisioned bool) error {
	if err := os.MkdirAll(plan.homeDir, 0o755); err != nil {
		return err
	}
	if plan.adapter.ID == envregistry.OpenCode {
		if err := os.MkdirAll(plan.parent, 0o755); err != nil {
			return err
		}
	}
	if err := plan.publishDocs(); err != nil {
		return err
	}
	want := map[string]bool{}
	for path := range plan.copies {
		want[path] = true
	}
	for path := range plan.links {
		want[path] = true
	}
	if seeds.claudeInit {
		want[".claude.json"] = true
	}
	for seed := range seeds.files {
		want[seed] = true
	}
	if backup := !provisioned; backup {
		// Versioned generations preserve replaced file bytes (§8.3).
		// Symlinks are skipped: they are manager-derived and re-derive
		// from the lock and the immutable store, while openBackup
		// follows links and fails on directory links (skills trees).
		replacing := map[string]bool{}
		for path := range want {
			info, err := os.Lstat(filepath.Join(plan.homeDir, filepath.FromSlash(path)))
			if err != nil || info.Mode()&os.ModeSymlink != 0 {
				continue
			}
			replacing[path] = true
		}
		if len(replacing) > 0 {
			if _, err := openBackup(plan.homeDir, replacing); err != nil {
				return err
			}
		}
	}
	// Remove recorded files the plan no longer wants.
	for path := range recorded {
		if !want[path] && !isSeedPath(plan, path) {
			_ = os.Remove(filepath.Join(plan.homeDir, filepath.FromSlash(path)))
		}
	}
	// Writes run in sorted path order: the fresh-home provisioning order
	// (§8.1) is deterministic even though the plan accumulates maps.
	for _, path := range sortedKeysBytes(plan.copies) {
		document := plan.copies[path]
		if err := os.MkdirAll(filepath.Dir(filepath.Join(plan.homeDir, filepath.FromSlash(path))), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(plan.homeDir, filepath.FromSlash(path)), document, 0o644); err != nil {
			return err
		}
	}
	for _, path := range sortedKeys(plan.links) {
		target := plan.links[path]
		full := filepath.Join(plan.homeDir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		_ = os.Remove(full)
		if err := os.Symlink(target, full); err != nil {
			// Copy fallback where the platform takes no symlink: record
			// the copy with its reason so hash drift applies to it.
			if err := copyLinkFallback(full, target, path, plan); err != nil {
				return err
			}
		}
	}
	// Copy seeds are written at provisioning only: after that the tool
	// owns them, so repair never refreshes them (§7.4).
	if provisioned {
		for seed, payload := range seeds.files {
			if err := os.WriteFile(filepath.Join(plan.homeDir, filepath.FromSlash(seed)), payload, 0o644); err != nil {
				return err
			}
		}
	}
	marker, err := finalizeMarker(req, plan, seeds, prior)
	if err != nil {
		return err
	}
	payload, err := marker.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(plan.homeDir, envmarker.Name), payload, 0o644)
}

// isSeedPath reports whether a recorded path is seed-owned tool state the
// repair must not remove.
func isSeedPath(plan *homePlan, path string) bool {
	if plan.marker.Seeds == nil {
		return false
	}
	for _, seed := range *plan.marker.Seeds {
		if seed == path {
			return true
		}
	}
	return false
}

// finalizeMarker records the passthrough entries, provisioning seeds,
// seeded launch-directory projects, and XDG seed links on the assembled
// marker, then reconciles the opencode XDG links: newly present allowlisted
// entries are seeded and recorded, recorded seeds whose target is gone are
// removed, and unrecorded entries shadowing allowlisted operator entries
// warn environment_seed_shadowed and are never touched (§7.1, §7.4).
func finalizeMarker(req *ResolveRequest, plan *homePlan, seeds *seedBundle, prior *envmarker.Marker) (*envmarker.Marker, error) {
	marker := plan.marker
	links, strategies, err := req.effectivePassthrough(plan.adapter, plan.isolation)
	if err != nil {
		return nil, err
	}
	passthrough := []envmarker.Passthrough{}
	for _, path := range sortedKeys(links) {
		full := filepath.Join(plan.homeDir, filepath.FromSlash(path))
		_ = os.Remove(full)
		if err := os.Symlink(links[path], full); err != nil {
			return nil, fmt.Errorf("passthrough %s: %v", path, err)
		}
		passthrough = append(passthrough, envmarker.Passthrough{Path: path, Strategy: strategies[path]})
	}
	marker.Passthrough = &passthrough
	// Copy-seed records survive repair: the tool owns the files, and
	// repair never refreshes them, so the record unions prior with new.
	recordedSeeds := []string{}
	seenSeeds := map[string]bool{}
	if prior != nil && prior.Seeds != nil {
		for _, seed := range *prior.Seeds {
			if !seenSeeds[seed] {
				seenSeeds[seed] = true
				recordedSeeds = append(recordedSeeds, seed)
			}
		}
	}
	for seed := range seeds.files {
		if !seenSeeds[seed] {
			seenSeeds[seed] = true
			recordedSeeds = append(recordedSeeds, seed)
		}
	}
	if seeds.claudeInit || plan.adapter.ID == envregistry.ClaudeCode {
		seeded, err := claudeSeed(plan.homeDir, req.LaunchDir, plan.form)
		if err != nil {
			return nil, err
		}
		marker.SeededProjects = seeded
		recordedSeeds = append(recordedSeeds, ".claude.json")
	}
	sort.Strings(recordedSeeds)
	marker.Seeds = &recordedSeeds
	seedLinks := []string{}
	if plan.adapter.ID == envregistry.OpenCode {
		for _, name := range sortedKeys(seeds.xdg) {
			full := filepath.Join(plan.parent, name)
			_ = os.Remove(full)
			if err := os.Symlink(seeds.xdg[name], full); err != nil {
				return nil, fmt.Errorf("xdg seed %s: %v", name, err)
			}
			seedLinks = append(seedLinks, name)
		}
		// A recorded seed whose target no longer exists is removed (§7.1).
		if prior != nil {
			for _, name := range prior.SeedLinks {
				if _, ok := seeds.xdg[name]; !ok {
					_ = os.Remove(filepath.Join(plan.parent, name))
				}
			}
		}
		plan.warnings = append(plan.warnings, shadowWarnings(req, plan, seedLinks)...)
	}
	marker.SeedLinks = seedLinks
	return marker, nil
}

// shadowWarnings reports unrecorded parent entries shadowing allowlisted
// operator entries as environment_seed_shadowed, left as they are (§7.1).
func shadowWarnings(req *ResolveRequest, plan *homePlan, recorded []string) []string {
	known := map[string]bool{}
	for _, name := range recorded {
		known[name] = true
	}
	xdg := req.operatorXDG()
	var warnings []string
	for _, name := range req.Machine.XDGSeedAllowlist {
		if name == "opencode" || known[name] {
			continue
		}
		if xdg == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(xdg, name)); err != nil {
			continue
		}
		if _, err := os.Lstat(filepath.Join(plan.parent, name)); err == nil {
			warnings = append(warnings, fmt.Sprintf("%s: %s shadows the allowlisted operator entry and is left as it is", envregistry.DiagSeedShadowed, name))
		}
	}
	return warnings
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysBytes(values map[string][]byte) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// verification is the lock-free currency verdict over exactly the surfaces
// the marker records (§10.1).
type verification struct {
	current  bool
	reasons  []string
	warnings []string
	// surfaceState carries the per-surface outcome for status rows: ""
	// (current), environment_surface_drift, environment_surface_missing,
	// or environment_surface_unreadable.
	surfaceState map[string]string
	plan         *homePlan
	marker       *envmarker.Marker
	hash         string
	order        []contextlock.Member
	servers      []contextmaterialize.MCPServer
}

// SurfaceDiagnostics for status rows (environments §8.5).
const (
	DiagSurfaceDrift      = "environment_surface_drift"
	DiagSurfaceMissing    = "environment_surface_missing"
	DiagSurfaceUnreadable = "environment_surface_unreadable"
)

// verifyHome verifies the managed home lock-free: it reads the marker and
// covers exactly the surfaces the marker records — no more. For a symlinked
// surface into an immutable store entry, link-target identity is sufficient
// currency (§10.1); a copied surface is verified by its recorded content
// hash. Assembly errors from the store are reported as staleness reasons
// here (fail-closed, no fragment); the same error under the repair lock
// surfaces as environment_repair_failed.
func verifyHome(req *ResolveRequest, adapter envregistry.Adapter, source Source, lock *contextlock.Lock, hash string) *verification {
	verdict := &verification{surfaceState: map[string]string{}}
	precedence := contextmaterialize.DefaultPrecedence
	order, err := contextmaterialize.EmittedOrder(lock, precedence)
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("emitted order: %v", err))
		return verdict
	}
	verdict.order = order
	marker, err := envmarker.Read(ManagedHomeDir(req.Home, req.Profile, adapter.ID))
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("marker invalid: %v", err))
		return verdict
	}
	if marker == nil {
		verdict.reasons = append(verdict.reasons, "home unprovisioned")
		return verdict
	}
	verdict.marker = marker
	manager := newGitManager(req.Home)
	servers, err := mcpsOf(req.Home, manager, lock)
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("mcp members: %v", err))
		return verdict
	}
	verdict.servers = servers
	plan, err := assembleHome(req, source, lock, hash, precedence, order, adapter, marker)
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("expected home: %v", err))
		return verdict
	}
	verdict.plan = plan
	verdict.hash = hash
	verdict.warnings = append(verdict.warnings, plan.warnings...)
	if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, "sha256:") {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("lock %s is not the current %s", marker.Profile.LockSHA256, strings.TrimPrefix(hash, "sha256:")))
	}
	if marker.Mode != envmarker.ModeManagedHome {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("mode %s is not managed-home", marker.Mode))
	}
	if marker.Precedence.Winner != precedence.Winner || marker.Precedence.Placement != precedence.Placement {
		verdict.reasons = append(verdict.reasons, "precedence does not match the effective policy")
	}
	if len(marker.Members) != len(order) {
		verdict.reasons = append(verdict.reasons, "member list does not match the lock")
	} else {
		for i, member := range order {
			recorded := marker.Members[i]
			pin := member.Commit
			if pin == "" {
				pin = member.StateHash
			}
			have := recorded.Commit
			if have == "" {
				have = recorded.StateSHA256
			}
			if recorded.Name != member.Name || recorded.Weight != member.Weight || recorded.Overlay != member.Overlay || have != pin {
				verdict.reasons = append(verdict.reasons, fmt.Sprintf("member %s does not match the lock", member.Name))
				break
			}
		}
	}
	verdict.checkSurfaces(req, plan, marker)
	verdict.checkPassthrough(req, plan, marker)
	verdict.checkClaudeProject(req, plan)
	verdict.checkXDG(req, plan, marker)
	verdict.checkShadows(plan, marker)
	verdict.checkToolRelease(req, adapter)
	verdict.current = len(verdict.reasons) == 0
	return verdict
}

// checkSurfaces covers exactly the recorded surfaces.
func (v *verification) checkSurfaces(req *ResolveRequest, plan *homePlan, marker *envmarker.Marker) {
	_ = req
	want := plan.marker.Surfaces
	for key, recorded := range marker.Surfaces {
		expected, ok := want[key]
		if !ok {
			v.reasons = append(v.reasons, fmt.Sprintf("surface %s is no longer materialized", key))
			v.mark(key, DiagSurfaceMissing)
			continue
		}
		if !equalStrings(recorded.Paths, expected.Paths) || recorded.ContentSHA256 != expected.ContentSHA256 {
			v.reasons = append(v.reasons, fmt.Sprintf("surface %s record does not match", key))
			v.mark(key, DiagSurfaceDrift)
			continue
		}
		if key == envmarker.SurfaceRootContext && recorded.Form != plan.form {
			v.reasons = append(v.reasons, fmt.Sprintf("root-context form %s is not the effective %s", recorded.Form, plan.form))
			v.mark(key, DiagSurfaceDrift)
		}
		copiedPaths := map[string]bool{}
		if recorded.Copies != nil {
			for _, copy := range *recorded.Copies {
				copiedPaths[copy.Path] = true
			}
		}
		for _, path := range recorded.Paths {
			full := filepath.Join(plan.homeDir, filepath.FromSlash(path))
			target, linked := plan.links[path]
			if linked && !copiedPaths[path] {
				info, err := os.Lstat(full)
				if err != nil {
					if os.IsNotExist(err) {
						v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is missing", path))
						v.mark(key, DiagSurfaceMissing)
					} else {
						v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is unreadable: %v", path, err))
						v.mark(key, DiagSurfaceUnreadable)
					}
					continue
				}
				if info.Mode()&os.ModeSymlink == 0 {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: expected a store link", path))
					v.mark(key, DiagSurfaceDrift)
					continue
				}
				got, err := os.Readlink(full)
				if err != nil || got != target {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: link target changed", path))
					v.mark(key, DiagSurfaceDrift)
				}
				continue
			}
			// A copied surface — the claude_code root file, a pi live
			// channel, or a recorded symlink-fallback copy — is verified
			// by the content hash (§10.1), read against the plan bytes
			// or, for a fallback copy, against the store source.
			if document, ok := plan.copies[path]; ok {
				v.mark(key, v.checkCopy(path, full, document))
				continue
			}
			if linked {
				v.mark(key, v.checkFallbackCopy(path, full, target))
				continue
			}
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is not in the plan", path))
			v.mark(key, DiagSurfaceDrift)
		}
	}
	for key := range want {
		if _, ok := marker.Surfaces[key]; !ok {
			v.reasons = append(v.reasons, fmt.Sprintf("surface %s is missing", key))
			v.mark(key, DiagSurfaceMissing)
		}
	}
}

// checkCopy verifies one copied file by its content hash (§10.1). A
// missing file and an unreadable file are different facts (§8.4). It
// returns the surface state for the status row.
func (v *verification) checkCopy(path, full string, document []byte) string {
	payload, err := os.ReadFile(full) // #nosec G304 -- home path from the verified plan
	if err != nil {
		if os.IsNotExist(err) {
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is missing", path))
			return DiagSurfaceMissing
		}
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is unreadable: %v", path, err))
		return DiagSurfaceUnreadable
	}
	if contextmaterialize.FileHash(payload) != contextmaterialize.FileHash(document) {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: content hash differs", path))
		return DiagSurfaceDrift
	}
	return ""
}

// checkFallbackCopy verifies a recorded symlink-fallback copy against its
// store source: file bytes for files, the store content hash for trees.
func (v *verification) checkFallbackCopy(path, full, source string) string {
	info, err := os.Stat(source)
	if err != nil {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s store source is unreadable: %v", path, err))
		return DiagSurfaceUnreadable
	}
	if info.IsDir() {
		want, err := contextstore.ContentHash(source)
		if err != nil {
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s store source is unreadable: %v", path, err))
			return DiagSurfaceUnreadable
		}
		got, err := contextstore.ContentHash(full)
		if err != nil {
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is unreadable: %v", path, err))
			return DiagSurfaceUnreadable
		}
		if got != want {
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: content hash differs", path))
			return DiagSurfaceDrift
		}
		return ""
	}
	payload, err := os.ReadFile(source) // #nosec G304 -- store source from the verified plan
	if err != nil {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s store source is unreadable: %v", path, err))
		return DiagSurfaceUnreadable
	}
	return v.checkCopy(path, full, payload)
}

// mark records the per-surface outcome, keeping the worst state: missing
// beats unreadable beats drifted beats current.
func (v *verification) mark(key, state string) {
	rank := map[string]int{"": 0, DiagSurfaceDrift: 1, DiagSurfaceUnreadable: 2, DiagSurfaceMissing: 3}
	if rank[state] > rank[v.surfaceState[key]] {
		v.surfaceState[key] = state
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// checkPassthrough runs the liveness row: every recorded entry must still
// be a symlink targeting the native entry, and the recorded set must match
// the effective one (§7.4).
func (v *verification) checkPassthrough(req *ResolveRequest, plan *homePlan, marker *envmarker.Marker) {
	links, strategies, err := req.effectivePassthrough(plan.adapter, plan.isolation)
	if err != nil {
		v.reasons = append(v.reasons, fmt.Sprintf("passthrough: %v", err))
		return
	}
	recorded := map[string]string{}
	if marker.Passthrough != nil {
		for _, entry := range *marker.Passthrough {
			recorded[entry.Path] = entry.Strategy
		}
	}
	if len(recorded) != len(links) {
		v.reasons = append(v.reasons, "passthrough entries do not match the effective set")
		return
	}
	for path, target := range links {
		strategy, ok := recorded[path]
		if !ok || strategy != strategies[path] {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is not recorded", path))
			continue
		}
		full := filepath.Join(plan.homeDir, filepath.FromSlash(path))
		got, err := os.Readlink(full)
		if err != nil || got != target {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is detached", path))
		}
	}
}

// checkClaudeProject enforces the per-launch-directory project entry under
// the referenced form: a launch directory without its entry makes the home
// stale for that directory (§7.4, §10.1). Under monolithic the entry is
// still merged on repair, but its absence is not staleness.
func (v *verification) checkClaudeProject(req *ResolveRequest, plan *homePlan) {
	if plan.adapter.ID != envregistry.ClaudeCode {
		return
	}
	entries, err := claudeProjects(plan.homeDir)
	if err != nil {
		v.reasons = append(v.reasons, fmt.Sprintf("managed .claude.json is unreadable: %v", err))
		return
	}
	if plan.form == envregistry.FormReferenced && !entries[req.LaunchDir] {
		v.reasons = append(v.reasons, fmt.Sprintf("launch directory %s has no project entry", req.LaunchDir))
	}
}

// checkXDG verifies the recorded XDG seed links and warns on shadows.
func (v *verification) checkXDG(req *ResolveRequest, plan *homePlan, marker *envmarker.Marker) {
	if plan.adapter.ID != envregistry.OpenCode {
		return
	}
	xdg := req.operatorXDG()
	expected := map[string]string{}
	if xdg != "" {
		for _, name := range req.Machine.XDGSeedAllowlist {
			if name == "opencode" {
				continue
			}
			if _, err := os.Stat(filepath.Join(xdg, name)); err == nil {
				expected[name] = filepath.Join(xdg, name)
			}
		}
	}
	recorded := map[string]bool{}
	for _, name := range marker.SeedLinks {
		recorded[name] = true
		target, ok := expected[name]
		if !ok {
			v.reasons = append(v.reasons, fmt.Sprintf("xdg seed %s target is gone", name))
			continue
		}
		got, err := os.Readlink(filepath.Join(plan.parent, name))
		if err != nil || got != target {
			v.reasons = append(v.reasons, fmt.Sprintf("xdg seed %s is detached", name))
		}
	}
	for name := range expected {
		if !recorded[name] {
			// A newly present allowlisted entry is reconciliation
			// input for the next repair, not staleness: the home is
			// still exactly as recorded (§7.1).
			v.warnings = append(v.warnings, fmt.Sprintf("xdg seed %s newly present: repair to seed", name))
		}
	}
	v.warnings = append(v.warnings, shadowWarnings(req, plan, marker.SeedLinks)...)
}

// checkShadows reports declared shadowing paths. The surface is genuinely
// inert so the status row is non-current by default; resolution warns and
// leaves staleness to the recorded surfaces (§7.5).
func (v *verification) checkShadows(plan *homePlan, marker *envmarker.Marker) {
	_ = marker
	for _, shadow := range plan.adapter.Shadows {
		if _, err := os.Lstat(filepath.Join(plan.homeDir, filepath.FromSlash(shadow.Path))); err != nil {
			continue
		}
		v.warnings = append(v.warnings, fmt.Sprintf("%s: %s exists and makes %s inert", envregistry.DiagShadowingPresent, shadow.Path, shadow.Surface))
	}
}

// checkToolRelease warns when the detected release differs from the
// recorded one. An undetectable tool reports unknown, never matching.
func (v *verification) checkToolRelease(req *ResolveRequest, adapter envregistry.Adapter) {
	detected := req.detected(adapter)
	if detected == "unknown" || detected == "" {
		v.warnings = append(v.warnings, fmt.Sprintf("%s: %s detected release unknown, recorded %s", envregistry.DiagToolVersionUnverifed, adapter.ID, recordedRelease(adapter)))
		return
	}
	if adapter.VerifiedRelease != "" && detected != adapter.VerifiedRelease {
		if comparison, ok := envregistry.CompareVersions(detected, adapter.VerifiedRelease); !ok || comparison != 0 {
			v.warnings = append(v.warnings, fmt.Sprintf("%s: %s detected %s, recorded %s", envregistry.DiagToolVersionUnverifed, adapter.ID, detected, adapter.VerifiedRelease))
		}
	}
}

func recordedRelease(adapter envregistry.Adapter) string {
	if adapter.VerifiedRelease == "" {
		return "unrecorded"
	}
	return adapter.VerifiedRelease
}

// buildFragment assembles the launch fragment from a current home (§10.2):
// profile.lock_sha256, the precedence object, env, the system_prompt
// section exactly when the home carries the inert file, and the mcp
// section exactly when the home carries the channel file, with the sorted
// env_names union under the §10.3 double bound. path_prepend is never
// emitted in revision 1.
func buildFragment(req *ResolveRequest, adapter envregistry.Adapter, verdict *verification) (*envfragment.Fragment, error) {
	plan := verdict.plan
	fragment := &envfragment.Fragment{
		Environment: adapter.ID,
		Profile:     req.Profile,
		LockSHA256:  strings.TrimPrefix(verdict.hash, "sha256:"),
		Winner:      contextmaterialize.DefaultPrecedence.Winner,
		Placement:   contextmaterialize.DefaultPrecedence.Placement,
		Env:         map[string]string{adapter.EnvVar: plan.parent},
	}
	if _, ok := plan.marker.Surfaces[envmarker.SurfaceSystemPrompt]; ok {
		fragment.SystemPrompt = &envfragment.SystemPrompt{
			Path:     filepath.Join(plan.homeDir, ".agent-context", "system-prompt.md"),
			Channels: adapter.SystemPrompt,
		}
	}
	if surface, ok := plan.marker.Surfaces[envmarker.SurfaceMCP]; ok && len(surface.Paths) == 1 {
		set, err := contextmaterialize.MCPSet(verdict.servers, adapter.ID)
		if err != nil {
			return nil, err
		}
		if adapter.MCP == nil {
			return nil, fmt.Errorf("managed home carries an mcp surface the %s adapter declares no channel for", adapter.ID)
		}
		fragment.MCP = &envfragment.MCP{
			Path:     filepath.Join(plan.homeDir, filepath.FromSlash(surface.Paths[0])),
			EnvNames: envfragment.BoundEnvNames(contextmaterialize.MCPEnvNames(set), req.Machine.PassableEnvNames),
			Channels: []envregistry.Channel{*adapter.MCP},
		}
	}
	if err := envfragment.CheckBoundary(adapter, EnvRoot(req.Home), fragment); err != nil {
		return nil, err
	}
	return fragment, nil
}

// renderFragment renders the fragment in the requested format.
func renderFragment(fragment *envfragment.Fragment, adapter envregistry.Adapter, format string) ([]byte, error) {
	switch format {
	case "", "json":
		return fragment.JSON()
	case "env":
		return fragment.EnvFormat(adapter), nil
	case "shell":
		return fragment.ShellFormat(adapter), nil
	default:
		return nil, fmt.Errorf("format %q is not json, env, or shell", format)
	}
}

// firstResolveNotice accompanies the first provisioning and every first
// resolve of a home (§8.1): the managed-home path, the own-state-root
// statement, and the first-run steps the seeds do not cover. Informative,
// never suppressed by configuration.
func firstResolveNotice(adapter envregistry.Adapter, plan *homePlan) string {
	var steps string
	switch adapter.ID {
	case envregistry.ClaudeCode:
		steps = "Log in inside this home on first use; accept the project trust dialog on the first interactive launch; approve MCP servers when prompted."
	case envregistry.CodexCLI:
		steps = "Log in unless the native credential is shared through; pass --skip-git-repo-check outside a git repository."
	case envregistry.OpenCode:
		steps = "Apply the resolved fragment by hand (OPENCODE_CONFIG names the MCP file); skills come from the machine-current profile, split-brain by construction."
	case envregistry.Pi:
		steps = "No trust wall; authentication is shared from the native home."
	}
	return fmt.Sprintf("Managed home provisioned at %s.\nThe tool treats this home as its own state root: sessions, trust records, and approvals accrue here, not in the native home.\nFirst-run steps the seeds do not cover: %s", plan.homeDir, steps)
}

// loadResolveInputs reads the profile source and lock for resolution. A
// missing profile is profile_unknown (§10.4).
func loadResolveInputs(home, profile string) (Source, *contextlock.Lock, string, error) {
	source, err := readSource(home, profile)
	if err != nil {
		return Source{}, nil, "", fmt.Errorf("%s: profile %q is not installed", DiagProfileUnknown, profile)
	}
	lock, hash, err := readLock(home, profile)
	if err != nil {
		return Source{}, nil, "", fmt.Errorf("%s: profile %q is not installed", DiagProfileUnknown, profile)
	}
	return source, lock, hash, nil
}

// currentProfileFor resolves the profile operand: the named profile, else
// the scoped current for the environment, else the machine current
// (§9.3, §10.1).
func currentProfileFor(home, envID, named string) (string, error) {
	if named != "" {
		return named, nil
	}
	scoped, err := ScopedCurrents(home)
	if err == nil {
		if profile, ok := scoped["env:"+envID]; ok && profile != "" {
			return profile, nil
		}
	}
	current, err := Current(home)
	if err != nil || current == "" {
		return "", fmt.Errorf("%s: no profile is current", DiagProfileUnknown)
	}
	return current, nil
}

// Resolve verifies the managed home lock-free and, when current, emits the
// launch fragment without touching any lock. A stale home emits no
// fragment without repair (§10.1: environment_home_stale with the
// reasons); under repair the home is provisioned or repaired from the
// store under the mutation lock and the fragment is emitted.
func Resolve(req ResolveRequest) (*ResolveResult, error) {
	adapter, err := envregistry.ByID(req.EnvID)
	if err != nil {
		return nil, err
	}
	profile, err := currentProfileFor(req.Home, req.EnvID, req.Profile)
	if err != nil {
		return nil, err
	}
	req.Profile = profile
	if !identifiers.Valid(profile) {
		return nil, fmt.Errorf("%s: profile %q is not installed", DiagProfileUnknown, profile)
	}
	source, lock, hash, err := loadResolveInputs(req.Home, profile)
	if err != nil {
		return nil, err
	}
	verdict := verifyHome(&req, adapter, source, lock, hash)
	if verdict.current {
		fragment, err := buildFragment(&req, adapter, verdict)
		if err != nil {
			return nil, err
		}
		document, err := renderFragment(fragment, adapter, req.Format)
		if err != nil {
			return nil, err
		}
		return &ResolveResult{Document: document, Warnings: verdict.warnings}, nil
	}
	if !req.Repair {
		return &ResolveResult{Warnings: verdict.warnings, StaleReasons: verdict.reasons}, fmt.Errorf("%s: %s", DiagHomeStale, strings.Join(verdict.reasons, "; "))
	}
	return repairUnderLock(&req, adapter, source, lock, hash)
}

// repairUnderLock takes the mutation lock with the bounded wait and a
// distinct lock-acquisition diagnostic, provisions or repairs the home
// from the store entries the lock names as one transaction, then emits
// the fragment (§10.1). A current home emits without touching any state;
// repair restores managed bytes from the store and never adopts candidate
// bytes found in the home.
func repairUnderLock(req *ResolveRequest, adapter envregistry.Adapter, source Source, lock *contextlock.Lock, hash string) (*ResolveResult, error) {
	op, err := beginOperation(req.Home)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()
	verdict := verifyHome(req, adapter, source, lock, hash)
	if verdict.current {
		fragment, err := buildFragment(req, adapter, verdict)
		if err != nil {
			return nil, err
		}
		document, err := renderFragment(fragment, adapter, req.Format)
		if err != nil {
			return nil, err
		}
		return &ResolveResult{Document: document, Warnings: verdict.warnings}, nil
	}
	precedence := contextmaterialize.DefaultPrecedence
	order, err := contextmaterialize.EmittedOrder(lock, precedence)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	provisioned := verdict.marker == nil
	if provisioned {
		if err := checkProfileCollision(req.Home, req.Profile); err != nil {
			return nil, err
		}
	}
	plan, err := assembleHome(req, source, lock, hash, precedence, order, adapter, verdict.marker)
	if err != nil {
		return nil, err
	}
	seeds, err := req.gatherSeeds(adapter, provisioned)
	if err != nil {
		return nil, err
	}
	if provisioned {
		want := map[string]bool{}
		for path := range plan.copies {
			want[path] = true
		}
		for path := range plan.links {
			want[path] = true
		}
		if err := inventoryUnmanaged(plan.homeDir, want, contextstore.Root(req.Home)); err != nil {
			return nil, err
		}
		if hint := foreignManagerHint(); hint != "" {
			plan.warnings = append(plan.warnings, fmt.Sprintf("%s: %s appears to manage this machine and will overwrite managed surfaces on its next apply", DiagForeignSuspect, hint))
		}
	}
	recorded := map[string]bool{}
	if verdict.marker != nil {
		for _, surface := range verdict.marker.Surfaces {
			for _, path := range surface.Paths {
				recorded[path] = true
			}
		}
	}
	if err := applyPlan(req, plan, seeds, recorded, verdict.marker, provisioned); err != nil {
		if provisioned {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	after := verifyHome(req, adapter, source, lock, hash)
	if !after.current {
		return nil, fmt.Errorf("%s: %s", DiagRepairFailed, strings.Join(after.reasons, "; "))
	}
	fragment, err := buildFragment(req, adapter, after)
	if err != nil {
		return nil, err
	}
	document, err := renderFragment(fragment, adapter, req.Format)
	if err != nil {
		return nil, err
	}
	result := &ResolveResult{Document: document, Warnings: after.warnings, Provisioned: provisioned}
	if provisioned {
		result.Notice = firstResolveNotice(adapter, plan)
	}
	return result, nil
}

// checkProfileCollision fails provisioning with environment_path_collision
// when two profile names map to one platform path below the environments
// root (§5).
func checkProfileCollision(home, profile string) error {
	entries, err := os.ReadDir(EnvRoot(home))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.Name() != profile && strings.EqualFold(entry.Name(), profile) {
			return fmt.Errorf("%s: profile names %q and %q map to one platform path", contextmaterialize.DiagPathCollision, entry.Name(), profile)
		}
	}
	return nil
}

// copyLinkFallback materializes a file link as a copy when the platform
// takes no symlink, recording the manager §5 fallback reason.
func copyLinkFallback(full, target, path string, plan *homePlan) error {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("link %s: %v", path, err)
	}
	if info.IsDir() {
		if err := copyTree(target, full); err != nil {
			return fmt.Errorf("link %s: %v", path, err)
		}
	} else {
		payload, err := os.ReadFile(target) // #nosec G304 -- link target resolved from the store plan
		if err != nil {
			return fmt.Errorf("link %s: %v", path, err)
		}
		if err := os.WriteFile(full, payload, 0o644); err != nil {
			return err
		}
	}
	for key, surface := range plan.marker.Surfaces {
		for _, candidate := range surface.Paths {
			if candidate == path {
				copies := append([]envmarker.Copy{}, *surface.Copies...)
				copies = append(copies, envmarker.Copy{Path: path, Reason: envmarker.ReasonSymlinkFallback})
				surface.Copies = &copies
				plan.marker.Surfaces[key] = surface
			}
		}
	}
	delete(plan.links, path)
	return nil
}

// copyTree copies a directory tree for symlink-fallback copies.
func copyTree(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		out := filepath.Join(destination, rel)
		if info.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		payload, err := os.ReadFile(path) // #nosec G304 -- store tree walked from the plan
		if err != nil {
			return err
		}
		return os.WriteFile(out, payload, info.Mode().Perm())
	})
}

// systemPrompt assembles the system-prompt surface into managed homes only
// (§5.5): the inert .agent-context/system-prompt.md, linked from the
// store, plus the pi live channels machine configuration explicitly
// materializes (off by default, so a plain launch carries no active
// system prompt).
func (p *homePlan) systemPrompt(req *ResolveRequest, lock *contextlock.Lock, precedence contextmaterialize.Precedence, packages map[string]contextmaterialize.Package, surfaces map[string]envmarker.Surface) error {
	document, written, err := contextmaterialize.SystemPrompt(lock, precedence, p.adapter.ID, packages)
	if err != nil {
		return err
	}
	if !written {
		return nil
	}
	const inert = ".agent-context/system-prompt.md"
	storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, inert)
	p.docs[storeDoc] = document
	p.links[inert] = storeDoc
	p.fileHashes[inert] = document
	paths := []string{inert}
	if p.adapter.ID == envregistry.Pi {
		switch req.Machine.SystemPromptFiles[req.Profile] {
		case "append":
			p.copies["APPEND_SYSTEM.md"] = document
			p.fileHashes["APPEND_SYSTEM.md"] = document
			paths = append(paths, "APPEND_SYSTEM.md")
		case "replace":
			p.copies["SYSTEM.md"] = document
			p.fileHashes["SYSTEM.md"] = document
			paths = append(paths, "SYSTEM.md")
		}
	}
	copies := []envmarker.Copy{}
	surfaces[envmarker.SurfaceSystemPrompt] = envmarker.Surface{
		Paths:         paths,
		ContentSHA256: contextmaterialize.SurfaceHash(p.fileHashesFor(paths)),
		Copies:        &copies,
	}
	return nil
}

// skills assembles the per-profile skills surface: one symlink per lock
// skill member into its immutable store entry (§7.1, §7.4 matrix). The
// surface hash binds each link to its entry's store content hash, so
// identical inputs hash identically on every machine (§5.6). opencode
// carries no skills surface: its skills target is the machine-global
// native surface, split-brain by construction (§7.1).
func (p *homePlan) skills(req *ResolveRequest, skills []skillOf, surfaces map[string]envmarker.Surface) {
	_ = req
	if p.adapter.SkillsDir == "" {
		return
	}
	hashes := map[string][]byte{}
	paths := []string{}
	copies := []envmarker.Copy{}
	for _, skill := range skills {
		path := p.adapter.SkillsDir + "/" + skill.name
		paths = append(paths, path)
		p.links[path] = skill.root
		hashes[path] = []byte(skill.hash)
	}
	surfaces[envmarker.SurfaceSkills] = envmarker.Surface{
		Paths:         paths,
		ContentSHA256: contextmaterialize.SurfaceHash(hashes),
		Copies:        &copies,
	}
}

// mcp assembles the §5.8 channel file surface: one inert file per
// adapter, managed homes only, with the sorted env_names union carried
// for the fragment.
func (p *homePlan) mcp(req *ResolveRequest, servers []contextmaterialize.MCPServer, surfaces map[string]envmarker.Surface) error {
	path, document, written, err := contextmaterialize.MCPFile(p.adapter.ID, servers)
	if err != nil {
		return err
	}
	if !written {
		return nil
	}
	storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, path)
	p.docs[storeDoc] = document
	p.links[path] = storeDoc
	p.fileHashes[path] = document
	copies := []envmarker.Copy{}
	surfaces[envmarker.SurfaceMCP] = envmarker.Surface{
		Paths:         []string{path},
		ContentSHA256: contextmaterialize.SurfaceHash(map[string][]byte{path: document}),
		Copies:        &copies,
	}
	return nil
}
