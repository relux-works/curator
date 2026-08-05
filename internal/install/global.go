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
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/mcp"
	"github.com/relux-works/curator/internal/runtimestore"
)

// GlobalRoot returns the global scope directory under the machine home.
func GlobalRoot(home string) string { return filepath.Join(home, "global") }

// Global installs the machine-wide scope: the global Skillfile into
// global/skills with shims in global/bin and home-level adapters
// (Spec §9.2). userHome receives the adapter mirrors.
//
// The lifecycle matches Project: the global root is the canonical operation
// identity, incomplete journals are recovered before any network or compiler
// work, and every shared target is published through one manager-home-locked
// transaction. The global scope registers no runtime consumer, so its commit
// carries no consumer-ledger class.
func Global(cfg *config.Config, userHome string, opts Options) Result {
	if opts.DryRun {
		result, _ := globalAttempt(cfg, userHome, opts, CommitDeps{})
		return result
	}
	commit, err := opts.Commit.resolve(cfg.Home(), opts.Build.Cache)
	if err != nil {
		return failedResult("global", GlobalRoot(cfg.Home()), err)
	}
	ctx := opts.context()
	locks, err := commit.Locks.AcquireProjects(ctx, GlobalRoot(cfg.Home()))
	if err != nil {
		return failedResult("global", GlobalRoot(cfg.Home()),
			fmt.Errorf("acquire the global operation lock: %w", err))
	}
	defer func() { _ = locks.Close() }()
	if err := recoverJournals(ctx, commit); err != nil {
		return failedResult("global", GlobalRoot(cfg.Home()), err)
	}
	return runWithRestarts("global", commit, func() (Result, *restartError) {
		return globalAttempt(cfg, userHome, opts, commit)
	})
}

func globalAttempt(cfg *config.Config, userHome string, opts Options, commit CommitDeps) (result Result, restart *restartError) {
	home := cfg.Home()
	result = Result{Alias: "global", Path: GlobalRoot(home), Status: "ok"}
	platform := opts.Platform
	if platform == "" {
		platform = runtimestore.Platform()
	}

	// The optimistic observation set opens before the first declaration input is
	// read and is revalidated under the manager-home lock in the commit phase.
	//
	// The machine-wide manifest lives under GlobalRoot(home), which is this
	// run's own operation identity, but that proves nothing about it:
	// `curator global add` and `curator global remove` rewrite it through
	// manifest.AddDecl and manifest.RemoveDecl without taking that lock. It is
	// therefore observed exactly like a project manifest: one read, and the
	// generation recorded for it is the generation of the bytes that read
	// returned, so a declaration that moves during the run restarts closure
	// resolution instead of committing the closure that was planned.
	observed := newObservations()

	globalManifest, globalManifestGeneration, err := readManifestDocument(GlobalRoot(home))
	if err != nil {
		result.failf("%v", err)
		return result, nil
	}
	observed.observeDocument(globalManifestKey, manifest.PathIn(GlobalRoot(home)), globalManifestGeneration)
	if globalManifest == nil {
		result.Status = "skipped"
		result.Messages = append(result.Messages, "global: Skillfile.json not found; run 'curator global init' first")
		return result, nil
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
			return result, nil
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
		return result, nil
	}
	for _, repo := range newSetEntries(fetchedBefore, opts.FetchedRepos) {
		result.Messages = append(result.Messages, "global: fetched "+filepath.Base(repo))
	}
	if !validateNodes(nodes, effectiveLocale, "global", &result) {
		return result, nil
	}
	if err := closure.DetectActiveCommandCollisions(nodes); err != nil {
		result.failf("%v", err)
		return result, nil
	}
	if err := checkSystemCommands(nodes); err != nil {
		result.failf("%v", err)
		return result, nil
	}
	if err := checkLegacySkillDependencies(nodes); err != nil {
		result.failf("%v", err)
		return result, nil
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
		return result, nil
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
		return result, nil
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
		return result, nil
	}

	// Narrow boundaries for the remaining read-only gates. Operation-private
	// toolchain state must never land in the global scope, the runtime store,
	// or a skill repository.
	deps, err := opts.Build.resolve(home, private, []string{
		GlobalRoot(home), filepath.Join(home, "runtime"), cfg.SkillsRoot,
	})
	if err != nil {
		result.failf("%v", err)
		return result, nil
	}

	skillsDir := filepath.Join(GlobalRoot(home), "skills")
	binDir := filepath.Join(GlobalRoot(home), "bin")

	// The moved-tag gate reads installed generations, so every marker it
	// consults joins the same optimistic observation set the machine-wide
	// manifest already entered, and the commit phase revalidates all of them
	// under the manager-home lock.
	//
	// The machine-wide scope consults no hybrid activation manifest: hybrid
	// declarations activate against a project, never against the global scope.
	// It also reads no development substitution manifest, so the only
	// declaration input it has is the manifest observed above.
	for _, node := range nodes {
		observed.observe("marker/global/"+node.Name, filepath.Join(skillsDir, node.Name, marker.Name))
	}
	movedTags := detectMovedTagsIn(skillsDir, nodes, deps.Generation)
	if len(movedTags) > 0 {
		if opts.StrictTags {
			result.failf("%s", strings.Join(movedTags, "; "))
			return result, nil
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
	result.BuildsComplete = plan.Complete()
	if planErr != nil {
		result.BuildDiagnostic = godriver.DiagnosticCode(planErr)
		result.failBuild(planErr)
		return result, nil
	}
	externalPlan, externalPlanErr := planExternalBuilds(opts.context(), "global", "global", home, nodes, nil, deps.Toolchain, opts.External, opts.DryRun)
	if externalPlanErr != nil {
		result.failBuild(externalPlanErr)
		return result, nil
	}
	for _, row := range externalPlan.rows {
		for _, warning := range row.result.Warnings {
			result.Messages = append(result.Messages, "global: warning: "+warning)
		}
		result.Messages = append(result.Messages, fmt.Sprintf("global: %s.%s external build key=%s outcome=%s", row.node.Name, row.command.Name, row.result.CacheKey, row.result.State))
	}
	for _, build := range plan.builds {
		observed.outcomes[build.skill+"."+build.command] = build.outcome
	}

	if opts.DryRun {
		for _, node := range nodes {
			result.Messages = append(result.Messages, fmt.Sprintf("global: %s (planned)", nodeSummary(node)))
		}
		result.Messages = append(result.Messages, "global: dry-run; no files modified")
		return result, nil
	}

	// Stage every build miss privately and finalize toolchain and build-source
	// trust, all before the first live mutation below.
	staged, stageErr := stageBuilds(opts.context(), plan, deps)
	if stageErr != nil {
		result.BuildDiagnostic = godriver.DiagnosticCode(stageErr)
		result.failBuild(stageErr)
		return result, nil
	}
	result.Messages = append(result.Messages, staged.Lines()...)
	result.Staged = staged.Builds()
	if opts.OnStaged != nil {
		if err := opts.OnStaged(staged); err != nil {
			result.failf("%v", err)
			return result, nil
		}
	}
	externalStaged, externalStageErr := stageExternalBuilds(opts.context(), externalPlan, deps.Toolchain, deps.Builder, private)
	if externalStageErr != nil {
		result.failBuild(externalStageErr)
		return result, nil
	}

	outcome, commitErr := runCommit(opts.context(), commitRequest{
		scope:    "global",
		home:     home,
		deps:     deps,
		commit:   commit,
		plan:     plan,
		staged:   staged,
		observed: observed,
		stageTargets: func(scoped scopeCommit) (scopeTargets, error) {
			return stageGlobalTargets(globalTargetRequest{
				cfg: cfg, home: home, userHome: userHome, platform: platform,
				nodes: nodes, agents: agents, effectiveLocale: effectiveLocale,
				mcpFound: mcpFound, attestations: attestations,
				skillsDir: skillsDir, binDir: binDir,
				plan: plan, deps: deps, scoped: scoped,
				external: externalStaged, externalStoreRoot: externalPlan.deps.StoreRoot,
			})
		},
	})
	result.Messages = append(result.Messages, outcome.messages...)
	result.Messages = append(result.Messages, outcome.warnings...)
	result.BuildCacheRetained = outcome.retainedBuilds
	if commitErr != nil {
		return result.failCommit(commitErr)
	}
	return result, nil
}

// globalTargetRequest is the scope-specific input of global target staging.
type globalTargetRequest struct {
	cfg               *config.Config
	home              string
	userHome          string
	platform          string
	nodes             []*closure.Node
	agents            []string
	effectiveLocale   string
	mcpFound          map[string]map[string][]string
	attestations      map[string]*marker.Attestation
	skillsDir         string
	binDir            string
	plan              BuildPlan
	deps              BuildDeps
	scoped            scopeCommit
	external          stagedExternal
	externalStoreRoot string
}

// stageGlobalTargets derives the complete desired machine-wide state under the
// manager-home lock, including the safe user-bin forwarding mirror.
func stageGlobalTargets(request globalTargetRequest) (scopeTargets, error) {
	var targets scopeTargets
	stageRoot := request.scoped.stageRoot

	runtime, err := stageRuntimeAndShims(
		stageRoot, request.home, request.binDir, request.nodes,
		runtimestore.GlobalCanonicalShim, request.platform, request.scoped, request.plan.plannedInputs(), request.external.entries, request.externalStoreRoot,
	)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(runtime.plan)
	targets.plan.Merge(request.external.transactionPlan(request.externalStoreRoot))
	targets.referencedKeys = runtime.referencedKeys()

	expectedSkills := map[string]bool{}
	var contextNames []string
	for _, node := range request.nodes {
		expectedSkills[node.Name] = true
		expected := buildMarker(node, request.effectiveLocale, request.agents, node.ActiveCommandNames(),
			request.mcpFound[node.Name], request.attestations[node.Name],
			runtime.builds[node.Name], request.plan.sourceIdentity(node.Name))
		nodePlan, status, err := stageNode(stageRoot, nodeInstall{
			node: node, store: request.skillsDir, kind: "global",
			locale: request.effectiveLocale, agents: request.agents, expected: expected,
		}, request.deps.Clock)
		if err != nil {
			return scopeTargets{}, fmt.Errorf("%s: %w", node.Name, err)
		}
		targets.plan.Merge(nodePlan)
		if node.ContextActive() {
			contextNames = append(contextNames, node.Name)
		}
		targets.messages = append(targets.messages, fmt.Sprintf("global: %s %s", nodeSummary(node), status))
	}

	staleSkills, err := stageStaleSkillRemovals(request.skillsDir, "global", expectedSkills)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(staleSkills)

	forwarding, err := globalbins.StageForwarding(
		stageRoot, request.home, runtime.commands, request.platform, nil, request.userHome)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(forwarding.Plan())
	targets.messages = append(targets.messages, forwarding.Messages...)

	envPlan, err := envfiles.StageGlobal(stageRoot, request.home)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(envPlan)

	sort.Strings(contextNames)
	mirror, err := adapters.StageGlobal(
		stageRoot, request.home, request.userHome, request.agents, contextNames, request.cfg.AdapterMode,
		contextSources(targets.plan, request.skillsDir, contextNames))
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(mirror.Plan())
	for _, message := range mirror.Messages {
		targets.messages = append(targets.messages, "global: "+message)
	}
	return targets, nil
}

// GlobalInit creates an empty global Skillfile.
func GlobalInit(home string) (string, error) {
	root := GlobalRoot(home)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return manifest.EnsureEmpty(root)
}
