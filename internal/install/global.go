package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/envfiles"
	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/hashing"
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

	globalManifest, globalManifestGeneration, globalManifestPayload, err := readManifestDocument(GlobalRoot(home))
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
	if globalManifest.SchemaVersion != manifest.SchemaVersion {
		result.failf("global: Skillfile schema %d is unsupported; global scope requires schema %d", globalManifest.SchemaVersion, manifest.SchemaVersion)
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
	// Legacy-lane package migration, exactly like the project scope: the
	// global scope has no draft lane, so every schema-9 or
	// subdirectory-selected installation escalates through the one staged
	// migration the marker, runtime, receipts, and audit record share.
	// An unstatable package fails here, before any audit, registry, or
	// compiler work.
	// A status plan stages the effective lock even though it is a dry
	// run, exactly like the project scope: the read-only currentness
	// check compares the recorded package and lock against this staged
	// expectation (skillfile-sources §4).
	legacy, err := planLegacyLane(globalManifestPayload, globalManifest, nodes, !opts.DryRun || opts.Operation == OperationStatus)
	if err != nil {
		result.failf("%v", err)
		return result, nil
	}
	if legacy != nil {
		result.LegacyPackages = legacy.Packages
		result.LegacyLockSHA256 = legacy.LockSHA256
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
				packageIdentity, _ := auditPackageForNode(node, nil)
				subjects = append(subjects, audit.Subject{
					Name: node.Name, Source: node.Decl.Source, Git: node.Decl.Git,
					Commit: node.Resolved.Commit, Snapshot: node.Snapshot,
					SchemaVersion: node.Spec.SchemaVersion, Capabilities: node.Spec.Capabilities,
					Directory: normalizedNodeDirectory(node),
					Package:   packageIdentity,
					Commands:  node.Spec.Commands,
					// The global scope has no draft lane: the
					// audited identity is the writer version.
					HashVersion: hashing.WriteVersion(),
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

	// Registry resolution (Spec §13) and the operation-level
	// unreachable-registry gate (registry §4, manager §7.1) both run before any
	// toolchain, cache, or compiler work.
	registryResult, regErr := resolveRegistryEvidence(cfg, nodes, "global", !opts.DryRun, false, opts)
	result.Messages = append(result.Messages, registryResult.Warnings...)
	result.Attestations = registryResult.Attestations
	registryRefused := false
	if opts.Operation != OperationStatus {
		registryRefused = addRegistryDiagnostic(&result, cfg, registryResult)
	}
	if registryRefused {
		if regErr != nil {
			result.Errors = append(result.Errors, regErr.Error())
		}
		return result, nil
	}
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
	if !opts.DryRun {
		commit, err = bindCommitPublisher(deps.Assurance, commit)
		if err != nil {
			result.failf("%v", err)
			return result, nil
		}
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
	movedTags, movedTagsErr := detectMovedTagsIn(skillsDir, nodes, deps.Generation, true, observed)
	if movedTagsErr != nil {
		result.failf("inspect installed markers for moved tags: %v", movedTagsErr)
		return result, nil
	}
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
	// list or go build and writes no persistent state. Migrated nodes
	// record their builds under the receipt-3 package wrapper on both
	// arms; every other member keeps the legacy receipts, keys and
	// namespaces byte-identical.
	var buildPackages map[string]*buildmeta.Package
	if legacy != nil {
		buildPackages = legacy.BuildPackages
	}
	plan, planErr := planBuilds(opts.context(), buildPlanRequest{
		scope: "global", nodes: nodes, deps: deps, dryRun: opts.DryRun, packages: buildPackages,
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
	externalPlan, externalPlanErr := planExternalBuilds(opts.context(), "global", "global", home, nodes, nil, buildPackages, deps.Toolchain, opts.External, deps.Assurance, opts.DryRun)
	if externalPlanErr != nil {
		result.failBuild(externalPlanErr)
		return result, nil
	}
	result.Messages = append(result.Messages, externalPlan.credentialReport("global")...)
	for _, row := range externalPlan.rows {
		for _, warning := range row.result.Warnings {
			result.Messages = append(result.Messages, "global: warning: "+warning)
		}
		result.Messages = append(result.Messages, fmt.Sprintf("global: %s.%s external build key=%s outcome=%s", row.node.Name, row.command.Name, row.result.CacheKey, row.result.State))
	}
	for _, build := range plan.builds {
		observed.outcomes[build.skill+"."+build.command] = build.outcome
	}

	// A status plan reports on installed trees, so it verifies them: a
	// recorded v1 identity over a NUL-bearing tree refuses here with the
	// opaque finding instead of the plan silently reporting ready. This
	// runs after build planning so the refusal still carries the complete
	// per-command verdict a read-only reporting caller needs. The global
	// scope has one store and no hybrid activation.
	if opts.Operation == OperationStatus {
		if err := refuseInstalledNULForStatus(nodes, func(string) string { return skillsDir }); err != nil {
			result.failf("%v", err)
			return result, nil
		}
	}

	if opts.DryRun {
		for _, node := range nodes {
			result.Messages = append(result.Messages, fmt.Sprintf("global: %s (planned)", nodeSummary(node)))
		}
		result.Messages = append(result.Messages, "global: dry-run; no files modified")
		return result, nil
	}

	// The mixed plan stages external commands before local go-v1 commands, in
	// the same command and receipt order as project installs. Both arms remain
	// private until the serialized commit below.
	externalStaged, externalStageErr := stageExternalBuilds(opts.context(), externalPlan, deps.Toolchain, deps.Builder, private)
	if externalStageErr != nil {
		result.failBuild(externalStageErr)
		return result, nil
	}
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
	outcome, commitErr := runCommit(opts.context(), commitRequest{
		scope:    "global",
		home:     home,
		deps:     deps,
		commit:   commit,
		plan:     plan,
		staged:   staged,
		observed: observed,
		stageTargets: func(scoped scopeCommit) (scopeTargets, error) {
			// A migrated node publishes its script runtime under the
			// source-v1 key of the same staged package its marker
			// records; every other node keeps the commit-keyed leaf.
			var runtimeKeys map[string]string
			if legacy != nil {
				runtimeKeys = legacy.RuntimeKeys
			}
			return stageGlobalTargets(globalTargetRequest{
				cfg: cfg, home: home, userHome: userHome, platform: platform,
				nodes: nodes, agents: agents, effectiveLocale: effectiveLocale,
				mcpFound: mcpFound, attestations: registryResult.Attestations,
				skillsDir: skillsDir, binDir: binDir,
				plan: plan, deps: deps, scoped: scoped,
				external: externalStaged, externalStoreRoot: externalPlan.deps.StoreRoot,
				runtimeKeys: runtimeKeys, legacy: legacy,
			})
		},
	})
	result.Messages = append(result.Messages, outcome.messages...)
	result.Messages = append(result.Messages, outcome.warnings...)
	result.BuildCacheRetained = outcome.retainedBuilds
	if commitErr != nil {
		return result.failCommit(externalTransactionFailure(commitErr, externalStaged))
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
	// runtimeKeys overrides the commit-keyed runtime leaf per migrated
	// node with the staged source-v1 key; legacy is the staged
	// package-marker migration, nil when no node migrates.
	runtimeKeys map[string]string
	legacy      *legacyLanePlan
}

// stageGlobalTargets derives the complete desired machine-wide state under the
// manager-home lock, including the safe user-bin forwarding mirror.
func stageGlobalTargets(request globalTargetRequest) (scopeTargets, error) {
	var targets scopeTargets
	stageRoot := request.scoped.stageRoot

	// The node snapshots are the admitted inputs of this scope: no
	// adapter destination may overwrite them, and the publication
	// guard pins their identity for the per-write recheck. A node
	// without one cannot stage and fails before anything is derived.
	admitted, err := nodeSnapshots(request.nodes)
	if err != nil {
		return scopeTargets{}, err
	}

	// The global scope has no draft lane: a nil key map keeps the
	// resolved-commit runtime behavior byte-identically. Migrated nodes
	// publish under their staged source-v1 keys instead.
	runtime, err := stageRuntimeAndShims(
		stageRoot, request.home, request.binDir, request.nodes,
		runtimestore.GlobalCanonicalShim, request.platform, request.scoped, request.plan.plannedInputs(), request.external.entries, request.externalStoreRoot,
		request.runtimeKeys, "",
	)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(runtime.plan)
	targets.messages = append(targets.messages, runtime.messages...)
	targets.plan.Merge(request.external.transactionPlan(request.externalStoreRoot))
	targets.adoptions = request.external.adoptions(request.externalStoreRoot)
	targets.referencedKeys = runtime.referencedKeys()

	expectedSkills := map[string]bool{}
	var contextNames []string
	for _, node := range request.nodes {
		expectedSkills[node.Name] = true
		expected, err := buildMarker(node, request.effectiveLocale, request.agents, node.ActiveCommandNames(),
			request.mcpFound[node.Name], request.attestations[node.Name],
			runtime.builds[node.Name], request.plan.sourceIdentity(node.Name), request.legacy)
		if err != nil {
			return scopeTargets{}, fmt.Errorf("%s: %w", node.Name, err)
		}
		nodePlan, status, err := stageNode(stageRoot, nodeInstall{
			node: node, store: request.skillsDir, kind: "global",
			locale: request.effectiveLocale, agents: request.agents, expected: expected,
		}, request.deps.Clock)
		if err != nil {
			return scopeTargets{}, fmt.Errorf("%s: %w", node.Name, err)
		}
		targets.plan.Merge(nodePlan)
		generationPlan, err := stageLegacyGeneration(stageRoot, request.skillsDir, "global", node.Name, request.legacy)
		if err != nil {
			return scopeTargets{}, err
		}
		targets.plan.Merge(generationPlan)
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

	// Enforced commands stay out of the user-bin forwarding mirror: a
	// forwarding shim is a shell wrapper, and no shell wrapper may stand
	// in front of an enforced launcher. The canonical native launcher is
	// the only published entry point.
	forwardable := map[string]bool{}
	for name := range runtime.commands {
		if !runtime.enforced[name] {
			forwardable[name] = true
		}
	}
	forwarding, err := globalbins.StageForwarding(
		stageRoot, request.home, forwardable, request.platform, nil, request.userHome)
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
		contextSources(targets.plan, request.skillsDir, contextNames), admitted)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(mirror.Plan())
	for _, message := range mirror.Messages {
		targets.messages = append(targets.messages, "global: "+message)
	}
	if err := targets.attachBoundaries(admitted); err != nil {
		return scopeTargets{}, err
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
