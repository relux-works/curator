// Package install orchestrates a project installation in the normative
// phase order of Spec §8.1.
//
// Gates that belong to later plan phases (hybrid scope, MCP verification,
// source audit, registry resolution) plug into the marked hook points; the
// core order, materialization, and cleanup are complete here.
package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/envfiles"
	"github.com/relux-works/curator/internal/gitignore"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/locale"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/mcp"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
)

// Options control one installation run.
type Options struct {
	DryRun       bool
	FixGitignore bool
	StrictTags   bool
	Verbose      bool
	Platform     string // "" resolves to the current platform
	// Hooks for later phases. Each may be nil.
	VerifyMcp     func(nodes []*closure.Node) (map[string]map[string][]string, []string, error)
	AuditGate     func(nodes []*closure.Node) (warnings []string, errs []string)
	ResolveAttest func(nodes []*closure.Node) (map[string]*marker.Attestation, []string, error)
}

// Result reports one project installation.
type Result struct {
	Alias    string
	Path     string
	Status   string // ok | skipped | failed
	Messages []string
	Errors   []string
}

func (r *Result) failf(format string, args ...any) {
	r.Status = "failed"
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

// Project installs one project per Spec §8.1.
func Project(cfg *config.Config, projectRoot, alias string, opts Options) Result {
	result := Result{Alias: alias, Path: projectRoot, Status: "ok"}
	platform := opts.Platform
	if platform == "" {
		platform = runtimestore.Platform()
	}

	// 1. Load the manifest; absent means skipped.
	projectManifest, err := manifest.Load(projectRoot)
	if err != nil {
		result.failf("%v", err)
		return result
	}
	if projectManifest == nil {
		result.Status = "skipped"
		result.Messages = append(result.Messages, alias+": Skillfile.json not found; skipped")
		return result
	}

	// 2. Effective agents.
	agents := projectManifest.Agents
	if len(agents) == 0 {
		if project, known := cfg.Projects[alias]; known && len(project.Agents) > 0 {
			agents = project.Agents
		} else {
			agents = cfg.DefaultAgents
		}
	}
	if unknown := adapters.UnknownAgents(agents); len(unknown) > 0 {
		result.Messages = append(result.Messages, fmt.Sprintf(
			"%s: warning: unknown agent(s) ignored: %s", alias, strings.Join(unknown, ", ")))
	}

	// 3. Managed .gitignore gate.
	required := adapters.RequiredGitignoreEntries(agents)
	if err := gitignore.Ensure(projectRoot, required, opts.FixGitignore && !opts.DryRun); err != nil {
		result.Status = "skipped"
		result.Messages = append(result.Messages, fmt.Sprintf("%s: %v; skipped", alias, err))
		return result
	}

	// 4. Dev substitutions.
	substitutions, err := devsub.Load(projectRoot)
	if err != nil {
		result.failf("%v", err)
		return result
	}
	if len(substitutions) > 0 {
		if cfg.Audit.Enabled && cfg.Audit.Mode == "strict" {
			result.failf("dev substitutions are active in %s; strict audit refuses substituted installs", devsub.Name)
			return result
		}
		if err := gitignore.Ensure(projectRoot, []string{devsub.Name}, opts.FixGitignore && !opts.DryRun); err != nil {
			result.Status = "skipped"
			result.Messages = append(result.Messages, fmt.Sprintf("%s: %v; skipped", alias, err))
			return result
		}
		var names []string
		for name := range substitutions {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			result.Messages = append(result.Messages, fmt.Sprintf(
				"%s: SUBSTITUTION %s -> %s", alias, name, substitutions[name].Describe()))
		}
	}

	// 5. Hybrid scope: applicable machine-level declarations join the
	// manifest; a project declaration of the same name shadows the hybrid
	// entry (Spec §9.3).
	hybridDecls, err := scopes.LoadHybridDecls(cfg.Home())
	if err != nil {
		result.failf("%v", err)
		return result
	}
	projectDeclared := map[string]bool{}
	for _, decl := range projectManifest.Skills {
		projectDeclared[decl.Name] = true
	}
	aliases := []string{alias}
	if project, known := cfg.Projects[alias]; known {
		aliases = append(aliases, project.ProjectAlias)
	}
	if projectManifest.ProjectAlias != "" {
		aliases = append(aliases, projectManifest.ProjectAlias)
	}
	var hybridDirect []manifest.Decl
	for _, hd := range hybridDecls {
		if !scopes.AppliesToProject(hd, aliases, projectRoot) {
			continue
		}
		if projectDeclared[hd.Decl.Name] {
			result.Messages = append(result.Messages, fmt.Sprintf(
				"%s: hybrid skill %s is shadowed by the project declaration", alias, hd.Decl.Name))
			continue
		}
		hybridDirect = append(hybridDirect, hd.Decl)
	}
	effectiveManifest := projectManifest
	if len(hybridDirect) > 0 {
		merged := *projectManifest
		merged.Skills = append(append([]manifest.Decl{}, projectManifest.Skills...), hybridDirect...)
		effectiveManifest = &merged
	}

	// 6. Effective locale.
	effectiveLocale := projectManifest.Locale
	if effectiveLocale == "" {
		effectiveLocale = cfg.PreferredLocale
	}

	// 7. Closure resolution.
	nodes, err := closure.Build(closure.Options{
		SkillsRoot:     cfg.SkillsRoot,
		Home:           cfg.Home(),
		AllowedSources: cfg.AllowedSources,
	}, effectiveManifest, substitutions)
	if err != nil {
		result.failf("%v", err)
		return result
	}

	// 8. Skill validation happens during closure manifest parsing; locale
	// analysis warnings surface per node during materialization.

	// 9. Active command collisions.
	if err := closure.DetectActiveCommandCollisions(nodes); err != nil {
		result.failf("%v", err)
		return result
	}

	// 10. Declared dependency checks.
	if err := checkSystemCommands(nodes); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := checkLegacySkillDependencies(nodes); err != nil {
		result.failf("%v", err)
		return result
	}

	// 11. MCP verification (Spec §11); Options.VerifyMcp overrides for tests.
	verifyMcp := opts.VerifyMcp
	if verifyMcp == nil {
		verifyMcp = func(nodes []*closure.Node) (map[string]map[string][]string, []string, error) {
			userHome, err := os.UserHomeDir()
			if err != nil {
				userHome = ""
			}
			requirements := map[string]map[string]skillspec.McpServer{}
			for _, node := range nodes {
				if len(node.Spec.McpServers) > 0 {
					requirements[node.Name] = node.Spec.McpServers
				}
			}
			findings, warnings, err := mcp.Verify(mcp.Env{ProjectRoot: projectRoot, UserHome: userHome}, agents, requirements)
			found := map[string]map[string][]string{}
			for name, finding := range findings {
				found[name] = finding.FoundIn
			}
			prefixed := make([]string, 0, len(warnings))
			for _, warning := range warnings {
				prefixed = append(prefixed, alias+": "+warning)
			}
			return found, prefixed, err
		}
	}
	mcpFound, mcpWarnings, mcpErr := verifyMcp(nodes)
	result.Messages = append(result.Messages, mcpWarnings...)
	if mcpErr != nil {
		result.failf("%v", mcpErr)
		return result
	}

	// 12. Migration warnings.
	for _, node := range nodes {
		for _, dependency := range node.Spec.Dependencies {
			if dependency.Type == "skill" {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"%s: %s uses dependencies.commands with type 'skill'; migrate to csk-skill.json schema v4 dependencies.skills",
					alias, node.Name))
				break
			}
		}
	}

	// 13. Audit gate (Spec §12); Options.AuditGate overrides for tests.
	auditGate := opts.AuditGate
	if auditGate == nil {
		auditGate = func(nodes []*closure.Node) ([]string, []string) {
			subjects := make([]audit.Subject, 0, len(nodes))
			for _, node := range nodes {
				subjects = append(subjects, audit.Subject{
					Name: node.Name, Source: node.Decl.Source, Git: node.Decl.Git,
					Commit: node.Resolved.Commit, Snapshot: node.Snapshot,
					SchemaVersion: node.Spec.SchemaVersion, Capabilities: node.Spec.Capabilities,
				})
			}
			warnings, errs := audit.Gate(cfg, subjects)
			prefixed := make([]string, 0, len(warnings))
			for _, warning := range warnings {
				prefixed = append(prefixed, alias+": "+warning)
			}
			return prefixed, errs
		}
	}
	auditWarnings, auditErrs := auditGate(nodes)
	result.Messages = append(result.Messages, auditWarnings...)
	if len(auditErrs) > 0 {
		result.failf("%s", strings.Join(auditErrs, "; "))
		return result
	}

	// 14. Registry resolution (Spec §13); Options.ResolveAttest overrides.
	resolveAttest := opts.ResolveAttest
	if resolveAttest == nil {
		resolveAttest = func(nodes []*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return resolveRegistries(cfg, nodes, alias)
		}
	}
	attestations, regWarnings, regErr := resolveAttest(nodes)
	result.Messages = append(result.Messages, regWarnings...)
	if regErr != nil {
		result.failf("%v", regErr)
		return result
	}

	// 15. Moved tags.
	movedTags := detectMovedTags(projectRoot, nodes)
	if len(movedTags) > 0 {
		if opts.StrictTags {
			result.failf("%s", strings.Join(movedTags, "; "))
			return result
		}
		for _, warning := range movedTags {
			result.Messages = append(result.Messages, alias+": "+warning)
		}
	}

	// 16. Dry run stops before any file changes.
	if opts.DryRun {
		for _, node := range nodes {
			result.Messages = append(result.Messages, fmt.Sprintf("%s: %s (planned)", alias, nodeSummary(node)))
		}
		result.Messages = append(result.Messages, alias+": dry-run; no files modified")
		return result
	}

	// 17. Record the checkout as a runtime consumer for GC (Spec §8.7).
	if err := scopes.RecordConsumer(cfg.Home(), projectRoot); err != nil {
		result.failf("%v", err)
		return result
	}

	// 18. Materialize every node in provider-first order.
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	binDir := filepath.Join(projectRoot, ".agents", "bin")
	hybridStore := scopes.HybridSkillsRoot(cfg.Home())
	hybridNames := hybridStoreNames(nodes, projectDeclared)
	expectedCommands := map[string]bool{}
	var contextNames []string
	var hybridContextNames []string
	expectedSkills := map[string]bool{}
	for _, node := range nodes {
		if !hybridNames[node.Name] {
			expectedSkills[node.Name] = true
		}
		active := node.ActiveCommands()
		var activeSorted []string
		for name := range active {
			activeSorted = append(activeSorted, name)
		}
		sort.Strings(activeSorted)

		if len(active) > 0 {
			commandNames, err := installRuntime(cfg.Home(), binDir, node, active, platform)
			if err != nil {
				result.failf("%s: %v", node.Name, err)
				return result
			}
			for name := range commandNames {
				expectedCommands[name] = true
			}
		}

		isHybrid := hybridNames[node.Name]
		targetDir := skillsDir
		nodeLocale := effectiveLocale
		nodeAgents := agents
		if isHybrid {
			// Hybrid context renders once per machine with the machine locale;
			// per-project variance stays out of the shared marker (Spec §9.3).
			targetDir = hybridStore
			nodeLocale = cfg.PreferredLocale
			nodeAgents = []string{}
		}
		expected := buildMarker(node, nodeLocale, nodeAgents, activeSorted, mcpFound[node.Name], attestations[node.Name])
		var status string
		var installErr error
		if node.ContextActive() {
			status, installErr = installContext(targetDir, node, nodeLocale, expected, &result, alias)
			if isHybrid {
				hybridContextNames = append(hybridContextNames, node.Name)
			} else {
				contextNames = append(contextNames, node.Name)
			}
		} else {
			status, installErr = installMarkerOnly(targetDir, node, expected)
		}
		if installErr != nil {
			result.failf("%s: %v", node.Name, installErr)
			return result
		}
		suffix := ""
		if isHybrid {
			suffix = " (hybrid)"
		}
		result.Messages = append(result.Messages, fmt.Sprintf("%s: %s%s %s", alias, nodeSummary(node), suffix, status))
		if opts.Verbose {
			result.Messages = append(result.Messages, fmt.Sprintf("%s: %s commit %s", alias, node.Name, node.Resolved.Commit))
		}
	}

	// 19. Cleanup, shims, env files, adapters.
	if err := cleanupRemoved(skillsDir, expectedSkills); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := runtimestore.RemoveStaleShims(binDir, expectedCommands, platform); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := envfiles.WriteProject(projectRoot); err != nil {
		result.failf("%v", err)
		return result
	}
	sort.Strings(contextNames)
	sort.Strings(hybridContextNames)
	if err := adapters.RefreshProject(projectRoot, agents, []adapters.Group{
		{Root: skillsDir, Skills: contextNames},
		{Root: hybridStore, Skills: hybridContextNames},
	}, cfg.AdapterMode); err != nil {
		result.failf("%v", err)
		return result
	}

	// 20. Garbage-collect unreferenced runtime entries (Spec §8.7).
	if _, err := scopes.CollectRuntime(cfg.Home()); err != nil {
		result.failf("%v", err)
		return result
	}
	return result
}

// hybridStoreNames returns nodes unreachable from project declarations: they
// materialize in the machine-level hybrid store (Spec §9.3).
func hybridStoreNames(nodes []*closure.Node, projectDeclared map[string]bool) map[string]bool {
	byName := map[string]*closure.Node{}
	for _, node := range nodes {
		byName[node.Name] = node
	}
	reachable := map[string]bool{}
	var stack []string
	for name := range projectDeclared {
		if _, present := byName[name]; present {
			stack = append(stack, name)
		}
	}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable[current] {
			continue
		}
		reachable[current] = true
		for name := range byName[current].Spec.Requirements {
			if _, present := byName[name]; present && !reachable[name] {
				stack = append(stack, name)
			}
		}
	}
	hybrid := map[string]bool{}
	for name := range byName {
		if !reachable[name] {
			hybrid[name] = true
		}
	}
	return hybrid
}

func installRuntime(home, binDir string, node *closure.Node, active map[string]bool, platform string) (map[string]bool, error) {
	installed := map[string]bool{}
	if len(node.Spec.RuntimeRoots) > 0 {
		if _, err := runtimestore.InstallRuntimeRoots(home, node.Name, node.Resolved.Commit, node.Snapshot, node.Spec.RuntimeRoots); err != nil {
			return nil, err
		}
	}
	var names []string
	for name := range node.Spec.Commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		command := node.Spec.Commands[name]
		if command.Type != "script" || !active[name] {
			continue
		}
		var runtimePath string
		var err error
		if len(node.Spec.RuntimeRoots) > 0 {
			runtimePath, err = runtimestore.RuntimeCommandPath(home, node.Name, node.Resolved.Commit, command, platform)
		} else {
			runtimePath, err = runtimestore.InstallSingleCommand(home, node.Name, node.Resolved.Commit, node.Snapshot, command, platform)
		}
		if err != nil {
			return nil, err
		}
		if _, err := runtimestore.WriteBinShim(binDir, name, runtimePath, platform); err != nil {
			return nil, err
		}
		installed[name] = true
	}
	return installed, nil
}

func installContext(skillsDir string, node *closure.Node, effectiveLocale string, expected *marker.Marker, result *Result, alias string) (string, error) {
	target := filepath.Join(skillsDir, node.Name)

	// Locale warnings must surface even for up-to-date installs.
	analysis := locale.Analyze(node.Snapshot, effectiveLocale)
	for _, issue := range analysis.Issues {
		if issue.Severity == "warning" {
			result.Messages = append(result.Messages, fmt.Sprintf("%s: %s: %s: %s", alias, node.Name, issue.Code, issue.Message))
		}
	}
	if analysis.Failed() {
		return "", fmt.Errorf("%s", analysis.Issues[0].Message)
	}

	current, err := marker.Current(target, expected)
	if err != nil {
		return "", err
	}
	if current {
		return "up-to-date", nil
	}

	tmp := filepath.Join(skillsDir, fmt.Sprintf(".%s.tmp-%d", node.Name, os.Getpid()))
	if err := os.RemoveAll(tmp); err != nil {
		return "", err
	}
	includeScripts := len(node.Spec.Commands) == 0
	if includeScripts {
		if _, err := os.Stat(filepath.Join(node.Snapshot, "scripts")); err != nil {
			includeScripts = false
		}
	}
	files, err := whitelist.CopyContext(node.Snapshot, tmp, includeScripts, node.Spec.RuntimeRoots)
	if err != nil {
		return "", err
	}
	if _, err := locale.Render(node.Snapshot, tmp, effectiveLocale); err != nil {
		return "", err
	}
	contentHash, err := hashing.ContentSHA256(tmp, nil)
	if err != nil {
		return "", err
	}
	expected.ContentSHA256 = contentHash
	expected.Files = files
	expected.InstalledAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	if err := marker.Write(tmp, expected); err != nil {
		return "", err
	}
	if err := marker.ReplaceDir(tmp, target); err != nil {
		return "", err
	}
	return "installed", nil
}

func installMarkerOnly(skillsDir string, node *closure.Node, expected *marker.Marker) (string, error) {
	target := filepath.Join(skillsDir, node.Name)
	current, err := marker.Current(target, expected)
	if err != nil {
		return "", err
	}
	if current {
		return "up-to-date", nil
	}
	tmp := filepath.Join(skillsDir, fmt.Sprintf(".%s.tmp-%d", node.Name, os.Getpid()))
	if err := os.RemoveAll(tmp); err != nil {
		return "", err
	}
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return "", err
	}
	contentHash, err := hashing.ContentSHA256(tmp, nil)
	if err != nil {
		return "", err
	}
	expected.ContentSHA256 = contentHash
	expected.Files = []string{}
	expected.Locale = ""
	expected.Agents = []string{}
	expected.InstalledAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	if err := marker.Write(tmp, expected); err != nil {
		return "", err
	}
	if err := marker.ReplaceDir(tmp, target); err != nil {
		return "", err
	}
	return "installed", nil
}

func buildMarker(node *closure.Node, effectiveLocale string, agents []string, activeCommands []string, mcp map[string][]string, attestation *marker.Attestation) *marker.Marker {
	var commands []string
	for name, command := range node.Spec.Commands {
		if command.Type == "script" {
			commands = append(commands, name)
		}
	}
	var dependencies []string
	for name := range node.Spec.Dependencies {
		dependencies = append(dependencies, name)
	}
	var requirements []string
	for name := range node.Spec.Requirements {
		requirements = append(requirements, name)
	}
	expected := &marker.Marker{
		Name:               node.Name,
		Source:             node.Decl.Source,
		RefKind:            node.Resolved.Kind,
		Ref:                node.Resolved.Ref,
		Commit:             node.Resolved.Commit,
		Locale:             effectiveLocale,
		Agents:             agents,
		Commands:           commands,
		Dependencies:       dependencies,
		SkillSchemaVersion: node.Spec.SchemaVersion,
		RuntimeRoots:       node.Spec.RuntimeRoots,
		Git:                node.Decl.Git,
		Requirements:       requirements,
		McpServers:         mcp,
		Attestation:        attestation,
		Activation:         &marker.Activation{Context: node.ContextActive(), Commands: activeCommands},
		Requirers:          node.Consumers(),
		Substituted:        node.Substituted,
	}
	if !node.ContextActive() {
		expected.Locale = ""
		expected.Agents = []string{}
	}
	return expected
}

func checkSystemCommands(nodes []*closure.Node) error {
	for _, node := range nodes {
		var checks []struct{ name, binary, hint string }
		for _, command := range node.Spec.Commands {
			if command.Type == "system" {
				checks = append(checks, struct{ name, binary, hint string }{command.Name, command.Command, command.Hint})
			}
		}
		for _, dependency := range node.Spec.Dependencies {
			if dependency.Type == "system" {
				checks = append(checks, struct{ name, binary, hint string }{dependency.Name, dependency.Command, dependency.Hint})
			}
		}
		sort.Slice(checks, func(i, j int) bool { return checks[i].name < checks[j].name })
		for _, check := range checks {
			if _, err := exec.LookPath(check.binary); err != nil {
				hint := ""
				if check.hint != "" {
					hint = " Hint: " + check.hint
				}
				return fmt.Errorf("missing system command %q for %s.%s", check.binary, node.Name, hint)
			}
		}
	}
	return nil
}

func checkLegacySkillDependencies(nodes []*closure.Node) error {
	byName := map[string]*closure.Node{}
	for _, node := range nodes {
		byName[node.Name] = node
	}
	var problems []string
	for _, node := range nodes {
		var names []string
		for name := range node.Spec.Dependencies {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			dependency := node.Spec.Dependencies[name]
			if dependency.Type != "skill" {
				continue
			}
			provider, present := byName[dependency.Skill]
			if !present {
				hint := ""
				if dependency.Hint != "" {
					hint = " Hint: " + dependency.Hint
				}
				problems = append(problems, fmt.Sprintf(
					"missing skill dependency %q for %s; add %s to Skillfile.json.%s",
					dependency.Skill, node.Name, dependency.Skill, hint))
				continue
			}
			provided, exported := provider.Spec.Commands[dependency.Command]
			if !exported || provided.Type != "script" {
				problems = append(problems, fmt.Sprintf(
					"skill dependency %s requires %s.%s, but %s does not export a script command named %q",
					node.Name, dependency.Skill, dependency.Command, dependency.Skill, dependency.Command))
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return nil
}

func detectMovedTags(projectRoot string, nodes []*closure.Node) []string {
	var warnings []string
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	for _, node := range nodes {
		if node.Resolved.Kind != "tag" {
			continue
		}
		recorded := marker.Read(filepath.Join(skillsDir, node.Name))
		if recorded == nil {
			continue
		}
		if recorded.RefKind == "tag" && recorded.Ref == node.Resolved.Ref && recorded.Commit != node.Resolved.Commit {
			warnings = append(warnings, fmt.Sprintf(
				"moved tag for %s: %s %s -> %s", node.Name, node.Resolved.Ref, recorded.Commit, node.Resolved.Commit))
		}
	}
	return warnings
}

func cleanupRemoved(skillsDir string, expected map[string]bool) error {
	entries, err := os.ReadDir(skillsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if !expected[name] {
			if err := os.RemoveAll(filepath.Join(skillsDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func nodeSummary(node *closure.Node) string {
	var active []string
	for name := range node.ActiveCommands() {
		active = append(active, name)
	}
	sort.Strings(active)
	summary := fmt.Sprintf("%s %s %s %s context=%s commands=[%s] via=%s",
		node.Name, node.Resolved.Kind, node.Resolved.Ref, shortCommit(node.Resolved.Commit),
		yesNo(node.ContextActive()), strings.Join(active, ","), strings.Join(node.Consumers(), ","))
	if node.Substituted != "" {
		summary += " SUBSTITUTED (" + node.Substituted + ")"
	}
	return summary
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func shortCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

// resolveRegistries applies the audit registry gate (Spec §13.3, §13.4):
// snapshot verification excludes tampered registries (all tampered fails),
// a verified revocation denies, strict policy fails unknown artifacts, and
// the authorizing attestation lands in the marker.
func resolveRegistries(cfg *config.Config, nodes []*closure.Node, alias string) (map[string]*marker.Attestation, []string, error) {
	trusted := cfg.TrustedRegistries()
	if len(trusted) == 0 {
		return map[string]*marker.Attestation{}, nil, nil
	}
	registries := make([]registry.Registry, 0, len(trusted))
	for _, entry := range trusted {
		registries = append(registries, registry.Registry{Name: entry.Name, URL: entry.URL, PublicKeys: entry.PublicKeys})
	}
	cacheDir := filepath.Join(cfg.Home(), "cache", "registry")
	var warnings []string
	tampered, snapshotWarnings := registry.CheckSnapshots(registries, cacheDir, registry.HTTPGetSnapshot, time.Now(), 0)
	for _, warning := range snapshotWarnings {
		warnings = append(warnings, alias+": registry: "+warning)
	}
	var usable []registry.Registry
	for _, reg := range registries {
		if !tampered[reg.URL] {
			usable = append(usable, reg)
		}
	}
	if len(usable) == 0 {
		return nil, warnings, fmt.Errorf("every trusted audit registry served a tampered snapshot")
	}
	fetch := registry.NewHTTPFetch(cacheDir, 0, 0, nil)
	strict := cfg.Audit.RegistryPolicy == "strict"
	attestations := map[string]*marker.Attestation{}
	var problems []string
	for _, node := range nodes {
		if node.Identity == "" {
			continue
		}
		contentHash, err := hashing.ContentSHA256(node.Snapshot, nil)
		if err != nil {
			return nil, warnings, err
		}
		resolution := registry.Resolve(usable, node.Identity, node.Resolved.Commit, contentHash, fetch)
		for _, warning := range resolution.Warnings {
			warnings = append(warnings, alias+": registry: "+warning)
		}
		switch resolution.Result {
		case registry.ResultRevoked:
			name := "a trusted registry"
			if resolution.Attestation != nil {
				name = resolution.Attestation.Registry
			}
			problems = append(problems, fmt.Sprintf("%s is revoked by %s", node.Name, name))
		case registry.ResultDeprecated:
			warnings = append(warnings, fmt.Sprintf("%s: registry: %s is marked deprecated", alias, node.Name))
		case registry.ResultUnknown:
			if strict {
				problems = append(problems, fmt.Sprintf(
					"%s is not audited by any trusted registry (registry_policy is strict)", node.Name))
			}
		}
		if resolution.Attestation != nil && resolution.Result != registry.ResultRevoked {
			attestations[node.Name] = &marker.Attestation{
				Registry: resolution.Attestation.Registry,
				Status:   resolution.Attestation.Status,
				KeyID:    resolution.Attestation.KeyID,
			}
		}
	}
	if len(problems) > 0 {
		return nil, warnings, fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return attestations, warnings, nil
}
