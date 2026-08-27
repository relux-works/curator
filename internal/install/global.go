package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/envfiles"
	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/mcp"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
)

// GlobalRoot returns the global scope directory under the machine home.
func GlobalRoot(home string) string { return filepath.Join(home, "global") }

// Global installs the machine-wide scope: the global Skillfile into
// global/skills with shims in global/bin and home-level adapters
// (Spec §9.2). userHome receives the adapter mirrors.
func Global(cfg *config.Config, userHome string, opts Options) (result Result) {
	home := cfg.Home()
	result = Result{Alias: "global", Path: GlobalRoot(home), Status: "ok"}
	platform := opts.Platform
	if platform == "" {
		platform = runtimestore.Platform()
	}

	globalManifest, err := manifest.Load(GlobalRoot(home))
	if err != nil {
		result.failf("%v", err)
		return result
	}
	if globalManifest == nil {
		result.Status = "skipped"
		result.Messages = append(result.Messages, "global: Skillfile.json not found; run 'curator global init' first")
		return result
	}

	agents := globalManifest.Agents
	if len(agents) == 0 {
		agents = cfg.DefaultAgents
	}
	if unknown := adapters.UnknownAgents(agents); len(unknown) > 0 {
		result.Messages = append(result.Messages, fmt.Sprintf(
			"global: warning: unknown agent(s) ignored: %s", strings.Join(unknown, ", ")))
	}
	effectiveLocale := globalManifest.Locale
	if effectiveLocale == "" {
		effectiveLocale = cfg.PreferredLocale
	}

	// One operation-private root serves the whole run: the read-only closure
	// workspace of a dry run and, later, the trusted toolchain's probe or build
	// base all live inside it. It is released last, after the plan dropped the
	// staging it owns.
	private := &privateRoot{prefix: operationPrivatePrefix}
	defer func() { releasePrivateRoot(&result, "global", private) }()
	scratchRoot := ""
	if opts.DryRun {
		scratchRoot, err = private.dir("closure-")
		if err != nil {
			result.failf("could not create the read-only dry-run workspace: %v", err)
			return result
		}
	}
	if opts.FetchedRepos == nil {
		opts.FetchedRepos = map[string]bool{}
	}
	fetchedBefore := copySet(opts.FetchedRepos)
	nodes, err := closure.Build(closure.Options{
		SkillsRoot:     cfg.SkillsRoot,
		Home:           home,
		AllowedSources: cfg.AllowedSources,
		FetchExisting:  opts.Fetch && !opts.DryRun,
		FetchedRepos:   opts.FetchedRepos,
		ScratchRoot:    scratchRoot,
	}, globalManifest, map[string]devsub.Substitution{})
	if err != nil {
		result.failf("%v", err)
		return result
	}
	for _, repo := range newSetEntries(fetchedBefore, opts.FetchedRepos) {
		result.Messages = append(result.Messages, "global: fetched "+filepath.Base(repo))
	}
	if !validateNodes(nodes, effectiveLocale, "global", &result) {
		return result
	}
	if err := closure.DetectActiveCommandCollisions(nodes); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := checkSystemCommands(nodes); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := checkLegacySkillDependencies(nodes); err != nil {
		result.failf("%v", err)
		return result
	}

	// MCP verification (Spec §11). The machine-wide scope runs the same gate as
	// a project install: its own root is the project-level configuration
	// surface and userHome is the user-level one. Options.VerifyMcp overrides.
	verifyMcp := opts.VerifyMcp
	if verifyMcp == nil {
		verifyMcp = mcpVerifier(mcp.Env{ProjectRoot: GlobalRoot(home), UserHome: userHome}, agents, "global")
	}
	mcpFound, mcpWarnings, mcpErr := verifyMcp(nodes)
	result.Messages = append(result.Messages, mcpWarnings...)
	if mcpErr != nil {
		result.failf("%v", mcpErr)
		return result
	}

	for _, node := range nodes {
		for _, dependency := range node.Spec.Dependencies {
			if dependency.Type == "skill" {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"global: %s uses dependencies.commands with type 'skill'; migrate to agent-skill.json schema v4 dependencies.skills",
					node.Name))
				break
			}
		}
	}

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
			var warnings, errs []string
			if opts.DryRun {
				warnings, errs = audit.GateReadOnly(cfg, subjects)
			} else {
				warnings, errs = audit.Gate(cfg, subjects)
			}
			for index := range warnings {
				warnings[index] = "global: " + warnings[index]
			}
			return warnings, errs
		}
	}
	warnings, auditErrors := auditGate(nodes)
	result.Messages = append(result.Messages, warnings...)
	if len(auditErrors) > 0 {
		result.failf("%s", strings.Join(auditErrors, "; "))
		return result
	}

	// Registry resolution (Spec §13); Options.ResolveAttest overrides. A
	// revoked or unaudited artifact fails the global scope here, before any
	// toolchain, cache, or compiler work.
	resolveAttest := opts.ResolveAttest
	if resolveAttest == nil {
		resolveAttest = func(nodes []*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return resolveRegistries(cfg, nodes, "global", !opts.DryRun)
		}
	}
	attestations, regWarnings, regErr := resolveAttest(nodes)
	result.Messages = append(result.Messages, regWarnings...)
	if regErr != nil {
		result.failf("%v", regErr)
		return result
	}

	// Narrow boundaries for the remaining read-only gates. Operation-private
	// toolchain state must never land in the global scope, the runtime store,
	// or a skill repository.
	deps, err := opts.Build.resolve(home, private, []string{
		GlobalRoot(home), filepath.Join(home, "runtime"), cfg.SkillsRoot,
	})
	if err != nil {
		result.failf("%v", err)
		return result
	}

	movedTags := detectMovedTagsIn(filepath.Join(GlobalRoot(home), "skills"), nodes, deps.Generation)
	if len(movedTags) > 0 {
		if opts.StrictTags {
			result.failf("%s", strings.Join(movedTags, "; "))
			return result
		}
		for _, warning := range movedTags {
			result.Messages = append(result.Messages, "global: "+warning)
		}
	}

	// Build planning is the last read-only phase: it resolves the trusted
	// toolchain identity and inspects protected cache state, but runs no go
	// list or go build and writes no persistent state.
	plan, planErr := planBuilds(opts.context(), buildPlanRequest{
		scope: "global", nodes: nodes, deps: deps, dryRun: opts.DryRun,
	})
	defer func() { releasePlan(&result, plan) }()
	result.Messages = append(result.Messages, plan.Lines()...)
	result.Builds = plan.Builds()
	if planErr != nil {
		result.failf("%v", planErr)
		return result
	}

	if opts.DryRun {
		for _, node := range nodes {
			result.Messages = append(result.Messages, fmt.Sprintf("global: %s (planned)", nodeSummary(node)))
		}
		result.Messages = append(result.Messages, "global: dry-run; no files modified")
		return result
	}

	// Stage every build miss privately and finalize toolchain and build-source
	// trust, all before the first live mutation below.
	staged, stageErr := stageBuilds(opts.context(), plan, deps)
	if stageErr != nil {
		result.failf("%v", stageErr)
		return result
	}
	result.Messages = append(result.Messages, staged.Lines()...)
	result.Staged = staged.Builds()
	if opts.OnStaged != nil {
		if err := opts.OnStaged(staged); err != nil {
			result.failf("%v", err)
			return result
		}
	}

	skillsDir := filepath.Join(GlobalRoot(home), "skills")
	binDir := filepath.Join(GlobalRoot(home), "bin")
	expectedCommands := map[string]bool{}
	var contextNames []string
	expectedSkills := map[string]bool{}
	for _, node := range nodes {
		expectedSkills[node.Name] = true
		active := node.ActiveCommands()
		activeSorted := node.ActiveCommandNames()
		if len(active) > 0 {
			commandNames, err := installRuntime(home, binDir, node, active, platform)
			if err != nil {
				result.failf("%s: %v", node.Name, err)
				return result
			}
			for name := range commandNames {
				expectedCommands[name] = true
			}
		}
		expected := buildMarker(node, effectiveLocale, agents, activeSorted, mcpFound[node.Name], attestations[node.Name])
		var status string
		var installErr error
		if node.ContextActive() {
			status, installErr = installContext(skillsDir, node, effectiveLocale, expected, deps.Clock)
			contextNames = append(contextNames, node.Name)
		} else {
			status, installErr = installMarkerOnly(skillsDir, node, expected, deps.Clock)
		}
		if installErr != nil {
			result.failf("%s: %v", node.Name, installErr)
			return result
		}
		result.Messages = append(result.Messages, fmt.Sprintf("global: %s %s", nodeSummary(node), status))
	}

	if err := cleanupRemoved(skillsDir, expectedSkills); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := runtimestore.RemoveStaleShims(binDir, expectedCommands, platform); err != nil {
		result.failf("%v", err)
		return result
	}
	result.Messages = append(result.Messages, globalbins.Refresh(home, expectedCommands, platform, nil, userHome)...)
	if err := envfiles.WriteGlobal(home); err != nil {
		result.failf("%v", err)
		return result
	}
	sort.Strings(contextNames)
	if err := adapters.RefreshGlobal(home, userHome, agents, contextNames, cfg.AdapterMode); err != nil {
		result.failf("%v", err)
		return result
	}
	if _, err := scopes.CollectRuntime(home); err != nil {
		result.failf("%v", err)
		return result
	}
	return result
}

// GlobalInit creates an empty global Skillfile.
func GlobalInit(home string) (string, error) {
	root := GlobalRoot(home)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return manifest.EnsureEmpty(root)
}
