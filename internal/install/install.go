// Package install orchestrates a project installation in the normative
// phase order of Spec §8.1.
//
// Gates that belong to later plan phases (hybrid scope, MCP verification,
// source audit, registry resolution) plug into the marked hook points; the
// core order, materialization, and cleanup are complete here.
package install

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/envfiles"
	"github.com/relux-works/curator/internal/gitignore"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/mcp"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/skillcheck"
	"github.com/relux-works/curator/internal/skillspec"
)

// Options control one installation run.
type Options struct {
	DryRun       bool
	Fetch        bool
	FixGitignore bool
	StrictTags   bool
	Verbose      bool
	Platform     string // "" resolves to the current platform
	// FetchedRepos deduplicates closure-scoped upgrades across selected
	// projects. Callers may leave it nil for a single installation.
	FetchedRepos map[string]bool
	// Context bounds toolchain probing and compilation. Nil means background.
	Context context.Context
	// Build injects the narrow toolchain, protected-cache, builder, clock, and
	// generation-read boundaries. The zero value resolves the real ones.
	Build BuildDeps
	// External injects operator-owned external repository acquisition, audit,
	// protected-store, and signer policy. Package manifests cannot populate it.
	External ExternalDeps
	// BuildSSH is the run-wide operator SSH selection this command line
	// carries. It takes precedence over every configured build_ssh scope.
	BuildSSH BuildSSHFlags
	// Commit injects the manager locks, transaction journal, cache publisher,
	// and post-commit collector of the serialized commit phase, plus the fault
	// hooks a rollback test drives. The zero value resolves the real ones.
	Commit CommitDeps
	// OnStaged observes the operation-private staging result after every build
	// succeeded and before the plan releases it. Publishing staged outputs to
	// the protected cache and to live installation targets belongs to the
	// atomic commit phase, not to this one.
	OnStaged func(Staged) error
	// Hooks for later phases. Each may be nil.
	VerifyMcp     func(nodes []*closure.Node) (map[string]map[string][]string, []string, error)
	AuditGate     func(nodes []*closure.Node) (warnings []string, errs []string)
	ResolveAttest func(nodes []*closure.Node) (map[string]*marker.Attestation, []string, error)
}

func (o Options) context() context.Context {
	if o.Context == nil {
		return context.Background()
	}
	return o.Context
}

// Result reports one project installation.
type Result struct {
	Alias    string
	Path     string
	Status   string // ok | skipped | failed
	Messages []string
	Errors   []string
	// Builds is the immutable build plan derived before any mutation.
	Builds []PlannedBuild
	// BuildsComplete reports whether Builds describes every compiled command the
	// closure activates. It can be true for a failed run, because a refusal is
	// itself a per-command verdict; a read-only reporting caller uses it to tell
	// a complete diagnostic from a silently partial one.
	BuildsComplete bool
	// Staged describes the operation-private outputs a real run produced. The
	// staging itself is already released when the result is returned.
	Staged []StagedBuild
	// BuildDiagnostic is the stable go-v1 boundary code when toolchain
	// selection, planning, or staging failed at a driver trust boundary, and is
	// empty otherwise. It carries no message text: the human-readable failure is
	// already in Errors, and a presentation layer branches on the code to add
	// operator guidance without parsing that text.
	BuildDiagnostic string
	// BuildCacheRetained marks a failed run that rebuilt protected cache state
	// and did not put the previous state back. There is more than one way to get
	// there — an incomplete transaction may still need what this run published,
	// a publication may have failed without being able to compensate itself, or
	// the reversal may not have completed — and the presentation layer treats
	// them alike, because they differ only in why. It is false for every run
	// that changed no cache entry and for every run that restored the ones it
	// changed, so a presentation layer can state which of the two happened
	// instead of claiming the live cache is always unchanged after a failure.
	// The warnings of the run say which case it was.
	BuildCacheRetained bool
}

func (r *Result) failf(format string, args ...any) {
	r.Status = "failed"
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

// failBuild records a failure of the build phase.
//
// Everything the planning and staging phases can fail on carries untrusted or
// manager-private detail: cache reasons and receipt bytes are filesystem
// controlled, compiler and driver text is toolchain controlled, and probe,
// staging, and cache locations are operation-private. The rendered message
// therefore goes through the one bounded, path-redacted rendering before it
// reaches an operator or a machine-readable field. The stable boundary code is
// carried separately in BuildDiagnostic and is never redacted.
func (r *Result) failBuild(err error) {
	r.Status = "failed"
	r.Errors = append(r.Errors, RedactDiagnostic(err.Error()))
}

// Project installs one project per Spec §8.1.
//
// A real operation holds the canonical project operation lock from before
// planning until after handoff, recovers every incomplete transaction journal
// under a brief manager-home lock before any network, audit, fingerprinting, or
// compilation work, and commits through one manager-home-locked transaction. A
// dry run acquires no lock, recovers nothing, and returns before staging.
func Project(cfg *config.Config, projectRoot, alias string, opts Options) Result {
	if opts.DryRun {
		result, _ := projectAttempt(cfg, projectRoot, alias, opts, CommitDeps{})
		return result
	}
	commit, err := opts.Commit.resolve(cfg.Home(), opts.Build.Cache)
	if err != nil {
		return failedResult(alias, projectRoot, err)
	}
	ctx := opts.context()
	locks, err := commit.Locks.AcquireProjects(ctx, projectRoot)
	if err != nil {
		return failedResult(alias, projectRoot, fmt.Errorf("acquire the project operation lock: %w", err))
	}
	defer func() { _ = locks.Close() }()
	if err := recoverJournals(ctx, commit); err != nil {
		return failedResult(alias, projectRoot, err)
	}
	return runWithRestarts(alias, commit, func() (Result, *restartError) {
		return projectAttempt(cfg, projectRoot, alias, opts, commit)
	})
}

// runWithRestarts re-runs one scope while revalidation keeps reporting that
// shared state moved. Restarting is always preferable to applying a stale plan;
// the bound turns a livelock into a reported failure instead of an endless loop.
func runWithRestarts(alias string, commit CommitDeps, attempt func() (Result, *restartError)) Result {
	var carried []string
	for tries := 1; ; tries++ {
		result, restart := attempt()
		if restart == nil {
			result.Messages = append(carried, result.Messages...)
			return result
		}
		if tries >= commit.MaxRestarts {
			result.Messages = append(carried, result.Messages...)
			result.failf("%v (gave up after %d attempts)", restart, tries)
			return result
		}
		carried = append(carried, fmt.Sprintf("%s: %v", alias, restart))
	}
}

func failedResult(alias, path string, err error) Result {
	result := Result{Alias: alias, Path: path, Status: "ok"}
	result.failf("%v", err)
	return result
}

func projectAttempt(cfg *config.Config, projectRoot, alias string, opts Options, commit CommitDeps) (result Result, restart *restartError) {
	result = Result{Alias: alias, Path: projectRoot, Status: "ok"}
	platform := opts.Platform
	if platform == "" {
		platform = runtimestore.Platform()
	}

	// The optimistic observation set opens here, before the first declaration
	// input is read, and stays open until build planning. Every entry is
	// re-read under the manager-home lock in the commit phase, and a difference
	// restarts instead of committing what was planned.
	//
	// Every declaration document below is read exactly once, and the generation
	// recorded for it is the generation of the bytes that read returned — the
	// same bytes the parser consumes. A write that races the read therefore
	// leaves a recorded generation that no longer matches the settled file, so
	// the recheck restarts; digesting the path as a second, separate operation
	// would let an A -> B -> A rewrite pass revalidation while the closure came
	// from the transient B (see generation.go). A spurious restart is bounded
	// and always preferable.
	observed := newObservations()

	// 1. Load the manifest; absent means skipped.
	//
	// The manifest is machine-mutable declaration state, not a stable input of
	// the locked checkout: `curator add` and `curator remove` rewrite it while
	// holding no operation lock, so it can gain, lose, or retarget a
	// declaration at any point during this run.
	projectManifest, projectManifestGeneration, err := readManifestDocument(projectRoot)
	if err != nil {
		result.failf("%v", err)
		return result, nil
	}
	observed.observeDocument(projectManifestKey, manifest.PathIn(projectRoot), projectManifestGeneration)
	if projectManifest == nil {
		result.Status = "skipped"
		result.Messages = append(result.Messages, alias+": Skillfile.json not found; skipped")
		return result, nil
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
		return result, nil
	}

	// 4. Dev substitutions. No Curator command writes this file, so nothing
	// serializes it against an installation at all; it is edited by hand while
	// a run may already be resolving. It redirects a declaration at a local
	// checkout or another ref and it decides the strict-audit refusal below, so
	// it selects installed content exactly like the manifest and is observed
	// the same way.
	substitutionManifest, substitutionsGeneration, err := readSubstitutionsDocument(projectRoot)
	if err != nil {
		result.failf("%v", err)
		return result, nil
	}
	observed.observeDocument(substitutionsKey, devsub.PathIn(projectRoot), substitutionsGeneration)
	substitutions := substitutionManifest.Substitutions
	buildRepositorySubstitutions := substitutionManifest.BuildRepositorySubstitutions
	if len(substitutions) > 0 || len(buildRepositorySubstitutions) > 0 {
		if cfg.Audit.Enabled && cfg.Audit.Mode == "strict" {
			result.failf("dev substitutions are active in %s; strict audit refuses substituted installs", devsub.Name)
			return result, nil
		}
		if err := gitignore.Ensure(projectRoot, []string{devsub.Name}, opts.FixGitignore && !opts.DryRun); err != nil {
			result.Status = "skipped"
			result.Messages = append(result.Messages, fmt.Sprintf("%s: %v; skipped", alias, err))
			return result, nil
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
		for _, skillName := range sortedBuildSubstitutionSkills(buildRepositorySubstitutions) {
			for _, repositoryName := range sortedBuildSubstitutionRepositories(buildRepositorySubstitutions[skillName]) {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"%s: BUILD REPOSITORY SUBSTITUTION %s.%s", alias, skillName, repositoryName))
			}
		}
	}

	// 5. Hybrid scope: applicable machine-level declarations join the
	// manifest; a project declaration of the same name shadows the hybrid
	// entry (Spec §9.3).
	//
	// The hybrid manifest is machine-home activation state, not project state:
	// `curator hybrid add|rm` rewrites it while holding no project operation
	// lock, so a declaration can appear, disappear, or retarget while this run
	// resolves its closure and stages builds. It is therefore an optimistic
	// observation of the exact activation bytes this closure was resolved from,
	// and the commit phase restarts closure resolution when they moved.
	hybridDecls, hybridGeneration, err := readHybridDocument(cfg.Home())
	if err != nil {
		result.failf("%v", err)
		return result, nil
	}
	observed.observeDocument(hybridActivationKey, scopes.HybridManifestPath(cfg.Home()), hybridGeneration)
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

	// 7. Closure resolution. One operation-private root serves the whole run:
	// the read-only closure workspace of a dry run and, later, the trusted
	// toolchain's probe or build base all live inside it. It is released last,
	// after the plan dropped the staging it owns.
	private := &privateRoot{prefix: operationPrivatePrefix}
	defer func() { releasePrivateRoot(&result, alias, private) }()
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
		Home:           cfg.Home(),
		AllowedSources: cfg.AllowedSources,
		FetchExisting:  opts.Fetch && !opts.DryRun,
		FetchedRepos:   opts.FetchedRepos,
		ScratchRoot:    scratchRoot,
	}, effectiveManifest, substitutions)
	if err != nil {
		result.failf("%v", err)
		return result, nil
	}
	for _, repo := range newSetEntries(fetchedBefore, opts.FetchedRepos) {
		result.Messages = append(result.Messages, fmt.Sprintf("%s: fetched %s", alias, filepath.Base(repo)))
	}

	// 8. Validate every node with the same rules as `skill check`. Manifest
	// parsing alone is insufficient because runtime-only nodes still require
	// SKILL.md and must pass locale consistency checks (Spec §4, §8.1).
	if !validateNodes(nodes, effectiveLocale, alias, &result) {
		return result, nil
	}

	// 9. Active command collisions.
	if err := closure.DetectActiveCommandCollisions(nodes); err != nil {
		result.failf("%v", err)
		return result, nil
	}

	// 10. Declared dependency checks.
	if err := checkSystemCommands(nodes); err != nil {
		result.failf("%v", err)
		return result, nil
	}
	if err := checkLegacySkillDependencies(nodes); err != nil {
		result.failf("%v", err)
		return result, nil
	}

	// 11. MCP verification (Spec §11); Options.VerifyMcp overrides for tests.
	verifyMcp := opts.VerifyMcp
	if verifyMcp == nil {
		userHome, homeErr := os.UserHomeDir()
		if homeErr != nil {
			userHome = ""
		}
		verifyMcp = mcpVerifier(mcp.Env{ProjectRoot: projectRoot, UserHome: userHome}, agents, alias)
	}
	mcpFound, mcpWarnings, mcpErr := verifyMcp(nodes)
	result.Messages = append(result.Messages, mcpWarnings...)
	if mcpErr != nil {
		result.failf("%v", mcpErr)
		return result, nil
	}

	// 12. Migration warnings.
	for _, node := range nodes {
		for _, dependency := range node.Spec.Dependencies {
			if dependency.Type == "skill" {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"%s: %s uses dependencies.commands with type 'skill'; migrate to agent-skill.json schema v4 dependencies.skills",
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
			var warnings, errs []string
			if opts.DryRun {
				warnings, errs = audit.GateReadOnly(cfg, subjects)
			} else {
				warnings, errs = audit.Gate(cfg, subjects)
			}
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
		return result, nil
	}

	// 14. Registry resolution (Spec §13); Options.ResolveAttest overrides.
	resolveAttest := opts.ResolveAttest
	if resolveAttest == nil {
		resolveAttest = func(nodes []*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return resolveRegistries(cfg, nodes, alias, !opts.DryRun)
		}
	}
	attestations, regWarnings, regErr := resolveAttest(nodes)
	result.Messages = append(result.Messages, regWarnings...)
	if regErr != nil {
		result.failf("%v", regErr)
		return result, nil
	}

	// 15. Narrow boundaries for the remaining read-only gates. Operation-private
	// toolchain state must never land in the checkout, the runtime store, or a
	// skill repository.
	deps, err := opts.Build.resolve(cfg.Home(), private, []string{
		projectRoot, filepath.Join(cfg.Home(), "runtime"), cfg.SkillsRoot,
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

	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	binDir := filepath.Join(projectRoot, ".agents", "bin")
	hybridStore := scopes.HybridSkillsRoot(cfg.Home())
	hybridNames := hybridStoreNames(nodes, projectDeclared)

	// 16. Moved tags. This gate reads installed generations, so every marker it
	// consults joins the same optimistic observation set the declaration inputs
	// already entered in steps 1, 4, and 5, and all of them are revalidated
	// under the manager-home lock before the plan is allowed to commit.
	for _, node := range nodes {
		store, kind := skillsDir, "project"
		if hybridNames[node.Name] {
			store, kind = hybridStore, "hybrid"
		}
		observed.observe("marker/"+kind+"/"+node.Name, filepath.Join(store, node.Name, marker.Name))
	}
	movedTags := detectMovedTags(projectRoot, nodes, deps.Generation)
	if len(movedTags) > 0 {
		if opts.StrictTags {
			result.failf("%s", strings.Join(movedTags, "; "))
			return result, nil
		}
		for _, warning := range movedTags {
			result.Messages = append(result.Messages, alias+": "+warning)
		}
	}

	// 17. Build planning. This is the last read-only phase: it resolves the
	// trusted toolchain identity and inspects protected cache state, but runs
	// no go list or go build and writes no persistent state.
	plan, planErr := planBuilds(opts.context(), buildPlanRequest{
		scope: alias, nodes: nodes, deps: deps, dryRun: opts.DryRun,
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
	externalPlan, externalPlanErr := planExternalBuilds(opts.context(), alias, alias, cfg.Home(), nodes,
		buildRepositorySubstitutions, deps.Toolchain, opts.External, deps.Assurance, opts.DryRun)
	if externalPlanErr != nil {
		result.failBuild(externalPlanErr)
		return result, nil
	}
	result.Messages = append(result.Messages, externalPlan.credentialReport(alias)...)
	for _, row := range externalPlan.rows {
		for _, warning := range row.result.Warnings {
			result.Messages = append(result.Messages, alias+": warning: "+warning)
		}
		result.Messages = append(result.Messages, fmt.Sprintf("%s: %s.%s external build key=%s outcome=%s",
			alias, row.node.Name, row.command.Name, row.result.CacheKey, row.result.State))
	}
	for _, build := range plan.builds {
		observed.outcomes[build.skill+"."+build.command] = build.outcome
	}

	// 18. Dry run stops before any file changes.
	if opts.DryRun {
		for _, node := range nodes {
			result.Messages = append(result.Messages, fmt.Sprintf("%s: %s (planned)", alias, nodeSummary(node)))
		}
		result.Messages = append(result.Messages, alias+": dry-run; no files modified")
		return result, nil
	}

	// 19. Stage every build miss privately, then finalize toolchain and
	// build-source trust. Nothing below has run yet, so a staging or trust
	// failure leaves the prior installation and the live build cache
	// byte-for-byte unchanged once the plan releases its private root.
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

	// 20. Serialized publication and commit. Everything below the private build
	// staging is derived and published under one manager-home mutation lock, in
	// deterministic classes, with the machine-wide consumer ledger last.
	outcome, commitErr := runCommit(opts.context(), commitRequest{
		scope:       alias,
		home:        cfg.Home(),
		projectRoot: projectRoot,
		deps:        deps,
		commit:      commit,
		plan:        plan,
		staged:      staged,
		observed:    observed,
		stageTargets: func(scoped scopeCommit) (scopeTargets, error) {
			return stageProjectTargets(projectTargetRequest{
				cfg: cfg, alias: alias, projectRoot: projectRoot, platform: platform,
				nodes: nodes, hybridNames: hybridNames, agents: agents,
				effectiveLocale: effectiveLocale, mcpFound: mcpFound, attestations: attestations,
				skillsDir: skillsDir, binDir: binDir, hybridStore: hybridStore,
				plan: plan, deps: deps, scoped: scoped, verbose: opts.Verbose,
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

// failCommit records one commit-phase failure of a scope.
//
// A stale observation is not a failure at all: it restarts the attempt. A
// failure at the protected build cache boundary carries untrusted or
// manager-private detail exactly like the planning and staging phases do, so it
// goes through the same bounded, redacted rendering and carries the same stable
// boundary code. Everything else is target, journal, or ledger state, whose
// paths are the operator's own and are the actionable part of the message.
func (r *Result) failCommit(commitErr error) (Result, *restartError) {
	var restart *restartError
	if errors.As(commitErr, &restart) {
		return *r, restart
	}
	var buildFailure *buildPhaseError
	if errors.As(commitErr, &buildFailure) {
		if code := godriver.DiagnosticCode(commitErr); code != "" {
			r.BuildDiagnostic = code
		}
		r.failBuild(commitErr)
		return *r, nil
	}
	r.failf("%v", commitErr)
	return *r, nil
}

// projectTargetRequest is the scope-specific input of project target staging.
type projectTargetRequest struct {
	cfg               *config.Config
	alias             string
	projectRoot       string
	platform          string
	nodes             []*closure.Node
	hybridNames       map[string]bool
	agents            []string
	effectiveLocale   string
	mcpFound          map[string]map[string][]string
	attestations      map[string]*marker.Attestation
	skillsDir         string
	binDir            string
	hybridStore       string
	plan              BuildPlan
	deps              BuildDeps
	scoped            scopeCommit
	verbose           bool
	external          stagedExternal
	externalStoreRoot string
}

// stageProjectTargets derives the complete desired state of one project under
// the manager-home lock. Nothing here writes a live path: every result is a
// staged replacement or an explicit managed removal for the transaction layer.
func stageProjectTargets(request projectTargetRequest) (scopeTargets, error) {
	var targets scopeTargets
	stageRoot := request.scoped.stageRoot

	// Runtime trees and canonical launchers come first: an install marker
	// records the published build identity of every compiled command, so the
	// protected cache entry must already be resolved when the marker is built.
	runtime, err := stageRuntimeAndShims(
		stageRoot, request.cfg.Home(), request.binDir, request.nodes,
		runtimestore.ProjectShim, request.platform, request.scoped, request.plan.plannedInputs(), request.external.entries, request.externalStoreRoot,
	)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(runtime.plan)
	targets.plan.Merge(request.external.transactionPlan(request.externalStoreRoot))
	targets.referencedKeys = runtime.referencedKeys()

	expectedSkills := map[string]bool{}
	expectedHybrid := map[string]bool{}
	var contextNames, hybridContextNames []string
	for _, node := range request.nodes {
		isHybrid := request.hybridNames[node.Name]
		store, kind := request.skillsDir, "project"
		nodeLocale, nodeAgents := request.effectiveLocale, request.agents
		if isHybrid {
			// Hybrid context renders once per machine with the machine locale;
			// per-project variance stays out of the shared marker (Spec §9.3).
			store, kind = request.hybridStore, "hybrid"
			nodeLocale, nodeAgents = request.cfg.PreferredLocale, []string{}
			expectedHybrid[node.Name] = true
		} else {
			expectedSkills[node.Name] = true
		}
		expected := buildMarker(node, nodeLocale, nodeAgents, node.ActiveCommandNames(),
			request.mcpFound[node.Name], request.attestations[node.Name],
			runtime.builds[node.Name], request.plan.sourceIdentity(node.Name))
		nodePlan, status, err := stageNode(stageRoot, nodeInstall{
			node: node, store: store, kind: kind, locale: nodeLocale, agents: nodeAgents, expected: expected,
		}, request.deps.Clock)
		if err != nil {
			return scopeTargets{}, fmt.Errorf("%s: %w", node.Name, err)
		}
		targets.plan.Merge(nodePlan)
		if node.ContextActive() {
			if isHybrid {
				hybridContextNames = append(hybridContextNames, node.Name)
			} else {
				contextNames = append(contextNames, node.Name)
			}
		}
		suffix := ""
		if isHybrid {
			suffix = " (hybrid)"
		}
		targets.messages = append(targets.messages,
			fmt.Sprintf("%s: %s%s %s", request.alias, nodeSummary(node), suffix, status))
		if request.verbose {
			targets.messages = append(targets.messages,
				fmt.Sprintf("%s: %s commit %s", request.alias, node.Name, node.Resolved.Commit))
		}
	}

	staleSkills, err := stageStaleSkillRemovals(request.skillsDir, "project", expectedSkills)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(staleSkills)

	envPlan, err := envfiles.StageProject(stageRoot, request.projectRoot)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(envPlan)

	sort.Strings(contextNames)
	sort.Strings(hybridContextNames)
	mirror, err := adapters.StageProject(stageRoot, request.projectRoot, request.agents, []adapters.Group{
		{
			Root:    request.skillsDir,
			Skills:  contextNames,
			Sources: contextSources(targets.plan, request.skillsDir, contextNames),
		},
		{
			Root:    request.hybridStore,
			Skills:  hybridContextNames,
			Sources: contextSources(targets.plan, request.hybridStore, hybridContextNames),
		},
	}, request.cfg.AdapterMode)
	if err != nil {
		return scopeTargets{}, err
	}
	targets.plan.Merge(mirror.Plan())
	for _, message := range mirror.Messages {
		targets.messages = append(targets.messages, request.alias+": "+message)
	}

	if len(runtime.commands) > 0 && !directoryOnPath(request.binDir, os.Getenv("PATH"), request.platform) {
		targets.messages = append(targets.messages, fmt.Sprintf(
			"%s: commands are installed in %s; agent skills can invoke that directory explicitly (.cmd on Windows). Optional bare commands for interactive use: curator shell-init --install",
			request.alias, request.binDir,
		))
	}
	return targets, nil
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

func runtimePathEntries(node *closure.Node, binDir, platform string) ([]string, error) {
	absoluteBin, err := filepath.Abs(binDir)
	if err != nil {
		return nil, err
	}
	candidates := []string{absoluteBin}
	type systemCommand struct{ name, binary string }
	var commands []systemCommand
	for _, command := range node.Spec.Commands {
		if command.Type == "system" {
			commands = append(commands, systemCommand{command.Name, command.Command})
		}
	}
	for _, dependency := range node.Spec.Dependencies {
		if dependency.Type == "system" {
			commands = append(commands, systemCommand{dependency.Name, dependency.Command})
		}
	}
	sort.Slice(commands, func(i, j int) bool { return commands[i].name < commands[j].name })
	for _, command := range commands {
		resolved, lookErr := exec.LookPath(command.binary)
		if lookErr != nil {
			return nil, fmt.Errorf("missing system command %q for %s", command.binary, node.Name)
		}
		if absolute, absErr := filepath.Abs(resolved); absErr == nil {
			resolved = absolute
		}
		if evaluated, evalErr := filepath.EvalSymlinks(resolved); evalErr == nil {
			resolved = evaluated
		}
		candidates = append(candidates, filepath.Dir(resolved))
	}
	seen := map[string]bool{}
	entries := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		key := filepath.Clean(candidate)
		if platform == "windows" {
			key = strings.ToLower(key)
		}
		if !seen[key] {
			seen[key] = true
			entries = append(entries, candidate)
		}
	}
	return entries, nil
}

func copySet(values map[string]bool) map[string]bool {
	cloned := make(map[string]bool, len(values))
	for value, present := range values {
		if present {
			cloned[value] = true
		}
	}
	return cloned
}

func newSetEntries(before, after map[string]bool) []string {
	var entries []string
	for value, present := range after {
		if present && !before[value] {
			entries = append(entries, value)
		}
	}
	sort.Strings(entries)
	return entries
}

func directoryOnPath(directory, pathValue, platform string) bool {
	separator := string(os.PathListSeparator)
	if platform == "windows" {
		separator = ";"
	}
	for _, entry := range strings.Split(pathValue, separator) {
		if entry == "" {
			continue
		}
		if left, leftErr := os.Stat(directory); leftErr == nil {
			if right, rightErr := os.Stat(entry); rightErr == nil && os.SameFile(left, right) {
				return true
			}
		}
		expected, expectedErr := filepath.Abs(directory)
		actual, actualErr := filepath.Abs(entry)
		if expectedErr != nil || actualErr != nil {
			continue
		}
		if platform == "windows" {
			if strings.EqualFold(filepath.Clean(expected), filepath.Clean(actual)) {
				return true
			}
		} else if filepath.Clean(expected) == filepath.Clean(actual) {
			return true
		}
	}
	return false
}

func validateNodes(nodes []*closure.Node, localeValue, alias string, result *Result) bool {
	valid := true
	for _, node := range nodes {
		for _, issue := range skillcheck.Validate(node.Snapshot, localeValue) {
			message := fmt.Sprintf("%s: %s: %s", alias, node.Name, skillcheck.Format(issue))
			if issue.Severity == "error" {
				result.failf("%s", message)
				valid = false
			} else {
				result.Messages = append(result.Messages, message)
			}
		}
	}
	return valid
}

func buildMarker(
	node *closure.Node,
	effectiveLocale string,
	agents []string,
	activeCommands []string,
	mcp map[string][]string,
	attestation *marker.Attestation,
	builds map[string]marker.Build,
	source *buildsource.Identity,
) *marker.Marker {
	// Commands lists every command the marker may carry build state for, so a
	// compiled command is exported here exactly like a script one.
	var commands []string
	for name, command := range node.Spec.Commands {
		if command.Type == "script" || command.Type == "build" {
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
	// Build state is all-or-nothing: a marker either records every compiled
	// command with the frozen source they were built from, or records neither.
	// A node with no published build must not carry a build source at all.
	if builds == nil {
		builds = map[string]marker.Build{}
	}
	// Marker v3 and v4 record an explicit receipt schema version and an
	// explicit execution policy for every local go-v1 build; marker v2 records
	// neither. The band is the manifest schema, not one exact version.
	if node.Spec.SchemaVersion >= 7 {
		upgraded := make(map[string]marker.Build, len(builds))
		for command, build := range builds {
			if build.Driver == "go-v1" {
				build.ReceiptSchemaVersion = 1
				build.ExecutionPolicy = "manager-worker-v1"
			}
			upgraded[command] = build
		}
		builds = upgraded
	}
	if len(builds) == 0 {
		source = nil
	}
	buildRoots := append([]string(nil), node.Spec.BuildRoots...)
	sort.Strings(buildRoots)
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
		BuildRoots:         buildRoots,
		BuildSource:        source,
		Builds:             builds,
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

func detectMovedTags(projectRoot string, nodes []*closure.Node, generation GenerationReader) []string {
	return detectMovedTagsIn(filepath.Join(projectRoot, ".agents", "skills"), nodes, generation)
}

// detectMovedTagsIn is a read-only gate: it only reads the recorded
// installation generation of each node through the injected reader.
func detectMovedTagsIn(skillsDir string, nodes []*closure.Node, generation GenerationReader) []string {
	var warnings []string
	for _, node := range nodes {
		if node.Resolved.Kind != "tag" {
			continue
		}
		recorded := generation.InstalledMarker(filepath.Join(skillsDir, node.Name))
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

// mcpVerifier is the default MCP requirement gate of one scope (Spec §11).
// env fixes the two configuration surfaces the check reads: the project-level
// one — a checkout for a project install, the global scope root for the
// machine-wide one — and the user home. Both scopes run the identical gate, so
// neither can reach build planning with an unproven MCP requirement.
func mcpVerifier(env mcp.Env, agents []string, scope string) func([]*closure.Node) (map[string]map[string][]string, []string, error) {
	return func(nodes []*closure.Node) (map[string]map[string][]string, []string, error) {
		requirements := map[string]map[string]skillspec.McpServer{}
		for _, node := range nodes {
			if len(node.Spec.McpServers) > 0 {
				requirements[node.Name] = node.Spec.McpServers
			}
		}
		findings, warnings, err := mcp.Verify(env, agents, requirements)
		found := map[string]map[string][]string{}
		for name, finding := range findings {
			found[name] = finding.FoundIn
		}
		prefixed := make([]string, 0, len(warnings))
		for _, warning := range warnings {
			prefixed = append(prefixed, scope+": "+warning)
		}
		return found, prefixed, err
	}
}

// resolveRegistries applies the audit registry gate (Spec §13.3, §13.4):
// snapshot verification excludes tampered registries (all tampered fails),
// a verified revocation denies, strict policy fails unknown artifacts, and
// the authorizing attestation lands in the marker.
func resolveRegistries(cfg *config.Config, nodes []*closure.Node, alias string, persist bool) (map[string]*marker.Attestation, []string, error) {
	trusted := cfg.TrustedRegistries()
	if len(trusted) == 0 {
		return map[string]*marker.Attestation{}, nil, nil
	}
	registries := make([]registry.Registry, 0, len(trusted))
	for _, entry := range trusted {
		registries = append(registries, registry.Registry{Name: entry.Name, URL: entry.URL, PublicKeys: entry.PublicKeys})
	}
	cacheDir := filepath.Join(cfg.Home(), "cache", "registry")
	stateDir := filepath.Join(cfg.Home(), "state", "registry")
	snapshotStateDir := stateDir
	if persist {
		if err := registry.MigrateSnapshotStates(cacheDir, stateDir); err != nil {
			return nil, nil, fmt.Errorf("audit registry rollback state migration failed: %w", err)
		}
	} else if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		// A pre-migration installation may still hold protected snapshot state in
		// the legacy cache. Reading it is safer than treating the registry as a
		// first use, and the read-only checker will fail closed if its catalog is
		// incomplete.
		snapshotStateDir = cacheDir
	}
	var warnings []string
	var tampered map[string]bool
	var snapshotWarnings []string
	if persist {
		tampered, snapshotWarnings = registry.CheckSnapshotsWithPolicy(
			registries, snapshotStateDir, registry.HTTPGetSnapshot, time.Now(),
			time.Duration(cfg.Audit.SnapshotMaxAgeSeconds)*time.Second,
			time.Duration(cfg.Audit.SnapshotClockSkewSeconds)*time.Second,
		)
	} else {
		tampered, snapshotWarnings = registry.CheckSnapshotsWithPolicyReadOnly(
			registries, snapshotStateDir, registry.HTTPGetSnapshot, time.Now(),
			time.Duration(cfg.Audit.SnapshotMaxAgeSeconds)*time.Second,
			time.Duration(cfg.Audit.SnapshotClockSkewSeconds)*time.Second,
		)
	}
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
	var fetch registry.FetchFn
	if persist {
		fetch = registry.NewHTTPFetchWithPolicy(
			cacheDir,
			time.Duration(cfg.Audit.CacheTTLSeconds)*time.Second,
			time.Duration(cfg.Audit.OfflineGraceSeconds)*time.Second,
			nil,
		)
	} else {
		fetch = registry.NewHTTPFetchWithPolicyReadOnly(
			cacheDir,
			time.Duration(cfg.Audit.CacheTTLSeconds)*time.Second,
			time.Duration(cfg.Audit.OfflineGraceSeconds)*time.Second,
			nil,
		)
	}
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
