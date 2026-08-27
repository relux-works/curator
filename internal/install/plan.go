package install

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/skillspec"
)

// BuildOutcome is the stable per-command planner vocabulary. The values match
// the read-only protected-cache inspection exactly, so a plan never invents an
// outcome the cache did not report.
type BuildOutcome string

const (
	// BuildCacheHit reuses a protected, exact cache entry. No source-aware Go
	// command runs for a hit.
	BuildCacheHit BuildOutcome = "cache-hit"
	// BuildWouldPreflightAndBuild is a cold miss: a real run preflights and
	// builds it privately.
	BuildWouldPreflightAndBuild BuildOutcome = "would-preflight-and-build"
	// BuildWouldRebuildUntrustedCache marks an entry whose protected provenance
	// no longer holds. A real run rebuilds it privately and never reuses it.
	BuildWouldRebuildUntrustedCache BuildOutcome = "would-rebuild-untrusted-cache"
	// BuildCorrupt marks a protected entry that cannot be interpreted. It is
	// never reused; a real run rebuilds it privately, and publication quarantines
	// the unusable entry rather than adopting or repairing it.
	BuildCorrupt BuildOutcome = "corrupt"
	// BuildUnsupported marks a platform that cannot protect the build cache. It
	// fails closed: without a provable boundary there is nowhere safe to publish
	// a rebuild either.
	BuildUnsupported BuildOutcome = "unsupported"
	// BuildToolchainUnavailable marks a command whose logical identity could not
	// be derived at all because the trusted Go toolchain could not be resolved or
	// verified. It fails closed and carries no target, key, or cache verdict.
	BuildToolchainUnavailable BuildOutcome = "toolchain-unavailable"
)

// buildable reports whether a real run must produce this command privately.
// Every entry the manager refuses to reuse but can replace is buildable: that
// is the whole repair path, because Curator has no separate repair command.
func (outcome BuildOutcome) buildable() bool {
	return outcome == BuildWouldPreflightAndBuild ||
		outcome == BuildWouldRebuildUntrustedCache ||
		outcome == BuildCorrupt
}

// blocking reports whether the outcome fails the installation before any
// persistent mutation. Only states a rebuild cannot resolve block.
func (outcome BuildOutcome) blocking() bool {
	return outcome == BuildUnsupported || outcome == BuildToolchainUnavailable
}

// PlannedBuild is one immutable planned build command. Every field is read
// through an accessor so a plan cannot be edited after the gates ran.
type PlannedBuild struct {
	skill         string
	command       string
	commandObject map[string]any
	buildRoot     string
	sourceDir     string
	source        buildsource.Identity
	target        buildmeta.Target
	input         buildmeta.Input
	key           buildmeta.CacheKey
	outcome       BuildOutcome
	reason        string
	diagnostic    string
	artifactPath  string
	receiptHash   buildmeta.ReceiptHash
	artifact      buildmeta.Artifact
}

// Skill is the closure node that declared the command.
func (build PlannedBuild) Skill() string { return build.skill }

// Command is the exported command name.
func (build PlannedBuild) Command() string { return build.command }

// Source is the logical build-source identity of the frozen skill snapshot.
func (build PlannedBuild) Source() buildsource.Identity { return build.source }

// BuildRoot is the protocol-relative module root of the command.
func (build PlannedBuild) BuildRoot() string { return build.buildRoot }

// SourceDir is the protocol-relative package directory of the command.
func (build PlannedBuild) SourceDir() string { return build.sourceDir }

// Target is the frozen native build target. The tuning map is copied so a
// caller cannot reach back into the plan.
func (build PlannedBuild) Target() buildmeta.Target {
	copied := buildmeta.Target{GOOS: build.target.GOOS, GOARCH: build.target.GOARCH}
	copied.Tuning = make(map[string]string, len(build.target.Tuning))
	for key, value := range build.target.Tuning {
		copied.Tuning[key] = value
	}
	return copied
}

// Driver is the closed build driver the command is planned for.
func (build PlannedBuild) Driver() string { return build.input.Driver }

// CacheKey is the derived protected build-cache key.
func (build PlannedBuild) CacheKey() buildmeta.CacheKey { return build.key }

// Expectation is the read-only protected-cache lookup this row's verdict was
// taken with, and is empty for a command whose logical identity was never
// derived.
//
// It exists so a read-only reporting caller can re-take the same lookup after
// it has finished classifying, and report that compiled state moved rather than
// publishing a verdict that was already stale. Nothing here can mutate the
// plan: the returned input is a copy, and every inspector this expectation is
// handed to is strictly read-only.
func (build PlannedBuild) Expectation() buildcache.Expectation {
	if build.key == "" {
		return buildcache.Expectation{}
	}
	input := build.input
	input.Target = build.Target()
	return buildcache.Expectation{Input: input}
}

// ReceiptSHA256 is the identity of the canonical receipt the protected cache
// currently holds for this command. It is set only for a hit, so a caller
// comparing it with a recorded install marker learns nothing from an entry the
// inspection already refused.
func (build PlannedBuild) ReceiptSHA256() buildmeta.ReceiptHash { return build.receiptHash }

// Artifact is the artifact metadata the protected cache currently holds:
// the manager-derived protocol-relative path, its hash, and its size. It is
// set only for a hit and never names a manager-private absolute path.
func (build PlannedBuild) Artifact() buildmeta.Artifact { return build.artifact }

// Outcome is the read-only protected-cache inspection result.
func (build PlannedBuild) Outcome() BuildOutcome { return build.outcome }

// Reason explains a non-reusable outcome; it is empty for a plain cold miss.
func (build PlannedBuild) Reason() string { return build.reason }

// DiagnosticCode is the stable go-v1 boundary code that prevented this command
// from being planned, and is empty for every command the plan could derive.
func (build PlannedBuild) DiagnosticCode() string { return build.diagnostic }

// ArtifactPath is the protected cache artifact and is set only for a hit.
func (build PlannedBuild) ArtifactPath() string { return build.artifactPath }

// Describe renders the exact logical source, target, key, and outcome of one
// planned command.
//
// A command the plan could not derive at all — the toolchain boundary failed
// before any identity existed — reports the fields it does have and omits the
// rest rather than publishing empty identities that read like real ones.
func (build PlannedBuild) Describe() string {
	line := fmt.Sprintf("%s.%s build", build.skill, build.command)
	for _, field := range []struct{ name, value string }{
		{"source", describeSource(build.source)},
		{"root", build.buildRoot},
		{"dir", build.sourceDir},
		{"target", describeTarget(build.target)},
		{"key", string(build.key)},
	} {
		if field.value != "" {
			line += " " + field.name + "=" + field.value
		}
	}
	line += " outcome=" + string(build.outcome)
	if build.reason != "" && build.outcome != BuildWouldPreflightAndBuild {
		line += fmt.Sprintf(" reason=%q", RedactDiagnostic(build.reason))
	}
	return line
}

func describeSource(source buildsource.Identity) string {
	if source.Algorithm == "" && source.ContentSHA256 == "" {
		return ""
	}
	return source.Algorithm + ":" + source.ContentSHA256
}

func describeTarget(target buildmeta.Target) string {
	if target.GOOS == "" && target.GOARCH == "" {
		return ""
	}
	rendered := target.GOOS + "/" + target.GOARCH
	keys := make([]string, 0, len(target.Tuning))
	for key := range target.Tuning {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		rendered += "+" + key + "=" + target.Tuning[key]
	}
	return rendered
}

// BuildPlan is the immutable outcome of the read-only planning phase. It owns
// the frozen build sources and, on a real run, the trusted Go session; Close
// releases both without touching any installation target.
type BuildPlan struct {
	scope    string
	builds   []PlannedBuild
	sources  map[string]*buildsource.Token
	session  BuildSession
	complete bool
}

// Builds returns the planned commands in provider-first, command-lexical order.
func (plan BuildPlan) Builds() []PlannedBuild {
	return append([]PlannedBuild(nil), plan.builds...)
}

// Empty reports whether the closure activates no build command at all. An
// empty plan performs no toolchain, cache, or Go work.
func (plan BuildPlan) Empty() bool { return len(plan.builds) == 0 }

// Complete reports whether the plan derived a row for every compiled command
// the closure activates — including a plan that then failed, because a refusal
// is itself a per-command verdict.
//
// It exists for the read-only reporting caller. A caller that only describes
// state may trust a complete plan to be the whole picture even when planning
// returned an error; a caller that mutates must still honour that error. An
// incomplete plan means some active command has no row at all, so a report
// derived from it would be silently partial.
func (plan BuildPlan) Complete() bool { return plan.complete }

// Misses returns the planned commands a real run must build privately.
func (plan BuildPlan) Misses() []PlannedBuild {
	var misses []PlannedBuild
	for _, build := range plan.builds {
		if build.outcome.buildable() {
			misses = append(misses, build)
		}
	}
	return misses
}

// Lines renders one scope-prefixed report line per planned command.
func (plan BuildPlan) Lines() []string {
	lines := make([]string, 0, len(plan.builds))
	for _, build := range plan.builds {
		lines = append(lines, plan.scope+": "+build.Describe())
	}
	return lines
}

// Verify re-establishes trust in everything the plan froze: the trusted
// toolchain identity through the last build child, and every build-source
// snapshot the plan inspected or compiled.
//
// Callers run it after the final build and before a staged output is handed
// off or any shared state changes, so drift that appeared during planning,
// cache reuse, or staging fails the installation while the prior installation
// and the live cache are still untouched.
func (plan BuildPlan) Verify(ctx context.Context) error {
	if len(plan.builds) == 0 {
		return nil
	}
	var errs []error
	if plan.session != nil {
		if err := plan.session.VerifyToolchain(ctx); err != nil {
			errs = append(errs, fmt.Errorf("trusted toolchain: %w", err))
		}
	}
	errs = append(errs, plan.recheckSources())
	return errors.Join(errs...)
}

// recheckSources rechecks every frozen snapshot in one deterministic pass, so
// a mutation of any planned source is reported the same way on every platform.
func (plan BuildPlan) recheckSources() error {
	var errs []error
	for _, name := range plan.sourceNames() {
		if err := plan.sources[name].Recheck(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}
	return errors.Join(errs...)
}

// sourceIdentity returns the frozen build-source identity of one node, or nil
// when the node declares no build command. The install marker records it so a
// later currentness check can compare the installed identity without reaching
// into the plan.
func (plan BuildPlan) sourceIdentity(name string) *buildsource.Identity {
	token := plan.sources[name]
	if token == nil {
		return nil
	}
	identity := token.Identity()
	return &identity
}

// sourceNames orders the frozen snapshots lexically by node name.
func (plan BuildPlan) sourceNames() []string {
	names := make([]string, 0, len(plan.sources))
	for name := range plan.sources {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Release drops the frozen snapshots and the operation-private toolchain
// session, deleting every staged output that session owns.
//
// It is teardown only. Nothing here revalidates the toolchain or rechecks a
// source: Verify owns both verdicts and has already taken them by the time a
// plan is released. The only failure Release can report is a private path it
// could not remove.
func (plan BuildPlan) Release() error {
	var errs []error
	if plan.session != nil {
		errs = append(errs, plan.session.Release())
	}
	for _, name := range plan.sourceNames() {
		errs = append(errs, plan.sources[name].Close())
	}
	return errors.Join(errs...)
}

// releasePlan drops the frozen snapshots and the operation-private staging root
// of one scope. It is deferred across every phase below planning, so it runs
// only once the installation has already failed or committed.
//
// That position is exactly why it cannot decide the outcome. Trust is finalized
// by Verify before the handoff and before the first persistent write, and a
// private root that outlived its operation is an operator warning — not a
// verdict on live state that has already changed.
func releasePlan(result *Result, plan BuildPlan) {
	if err := plan.Release(); err != nil {
		result.Messages = append(result.Messages, fmt.Sprintf(
			"%s: could not release the operation-private build root: %s",
			plan.scope, RedactDiagnostic(err.Error())))
	}
}

// buildPlanRequest is the read-only input of one planning phase.
type buildPlanRequest struct {
	scope  string
	nodes  []*closure.Node
	deps   BuildDeps
	dryRun bool
}

// planBuilds derives the immutable build plan after every manifest, closure,
// collision, requirement, audit, registry, and moved-tag gate has passed.
//
// Planning is read-only: it validates frozen sources, resolves the trusted
// toolchain identity (a probe on a dry run), and inspects protected cache
// state. It never runs go list or go build and never writes persistent state.
// The plan is returned even on failure so the caller can report every outcome
// it managed to derive.
func planBuilds(ctx context.Context, request buildPlanRequest) (BuildPlan, error) {
	plan := BuildPlan{scope: request.scope}
	planned := plannedCommands(request.nodes)
	if len(planned) == 0 {
		return plan, nil
	}

	plan.sources = map[string]*buildsource.Token{}
	for _, node := range buildNodes(request.nodes) {
		token, err := buildsource.Validate(node.Snapshot)
		if err != nil {
			return plan, fmt.Errorf("%s: %w", node.Name, err)
		}
		plan.sources[node.Name] = token
	}

	var target buildmeta.Target
	var toolchain buildmeta.Toolchain
	if request.dryRun {
		probedTarget, probedToolchain, err := request.deps.Toolchain.Probe(ctx)
		if err != nil {
			plan.builds, plan.complete = toolchainInventory(planned, plan.sources, err), true
			return plan, err
		}
		target, toolchain = probedTarget, probedToolchain
	} else {
		session, err := request.deps.Toolchain.Establish(ctx)
		if err != nil {
			plan.builds, plan.complete = toolchainInventory(planned, plan.sources, err), true
			return plan, err
		}
		plan.session = session
		target, toolchain = session.Target(), session.Toolchain()
	}

	var blocked []string
	for _, item := range planned {
		build, err := planOne(item, plan.sources[item.node.Name], target, toolchain, request.deps.Cache)
		if err != nil {
			return plan, err
		}
		plan.builds = append(plan.builds, build)
		if build.outcome.blocking() {
			blocked = append(blocked, fmt.Sprintf("%s.%s is %s: %s",
				build.skill, build.command, build.outcome, RedactDiagnostic(build.reason)))
		}
	}
	plan.complete = true
	if len(blocked) > 0 {
		return plan, fmt.Errorf("build cache refused reuse: %s", strings.Join(blocked, "; "))
	}
	return plan, nil
}

// toolchainInventory records every command the closure activates when the
// trusted toolchain boundary failed before a single logical identity could be
// derived.
//
// The plan still fails; the inventory exists so a read-only currentness check
// can report one diagnostic per active compiled command — with the stable
// boundary code that refused it — instead of silently reporting nothing at all
// about compiled state.
//
// Every identity that was already established when the toolchain was consulted
// is carried: the closed build driver the schema admits for the command, the
// build root and package directory the declaration names, and the build-source
// digest of the frozen snapshot, which is validated before the toolchain is
// touched. Only what genuinely depends on the toolchain — the native target,
// the logical key, and any cache verdict — stays empty, so a row never
// publishes an identity nothing derived.
func toolchainInventory(
	planned []plannedCommand,
	sources map[string]*buildsource.Token,
	err error,
) []PlannedBuild {
	code := godriver.DiagnosticCode(err)
	builds := make([]PlannedBuild, 0, len(planned))
	for _, item := range planned {
		// A build root that cannot be resolved is reported as absent rather than
		// guessed: this row exists to name the command, not to assert identity.
		buildRoot, rootErr := buildRootFor(item.node.Spec, item.command)
		if rootErr != nil {
			buildRoot = ""
		}
		var source buildsource.Identity
		if token := sources[item.node.Name]; token != nil {
			source = token.Identity()
		}
		builds = append(builds, PlannedBuild{
			skill:         item.node.Name,
			command:       item.command.Name,
			commandObject: commandObject(item.command),
			buildRoot:     buildRoot,
			sourceDir:     item.command.SourceDir,
			source:        source,
			// The driver is the closed go-v1 boundary the parser already admitted
			// this command under; only the input it would have keyed is unknown.
			input:      buildmeta.Input{Driver: item.command.Driver},
			outcome:    BuildToolchainUnavailable,
			reason:     err.Error(),
			diagnostic: code,
		})
	}
	return builds
}

func planOne(
	item plannedCommand,
	source *buildsource.Token,
	target buildmeta.Target,
	toolchain buildmeta.Toolchain,
	cache CacheInspector,
) (PlannedBuild, error) {
	buildRoot, err := buildRootFor(item.node.Spec, item.command)
	if err != nil {
		return PlannedBuild{}, fmt.Errorf("%s.%s: %w", item.node.Name, item.command.Name, err)
	}
	if source == nil {
		return PlannedBuild{}, fmt.Errorf("%s.%s: build source was not validated", item.node.Name, item.command.Name)
	}
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion,
		Driver:        buildmeta.DriverGoV1,
		BuildSource:   source.Identity(),
		BuildRoot:     buildRoot,
		Command:       item.command.Name,
		SourceDir:     item.command.SourceDir,
		Target:        target,
		Toolchain:     toolchain,
		Policy:        buildmeta.FixedPolicy(),
	}
	if err := input.Validate(); err != nil {
		return PlannedBuild{}, fmt.Errorf("%s.%s: %w", item.node.Name, item.command.Name, err)
	}
	key, err := input.CacheKey()
	if err != nil {
		return PlannedBuild{}, fmt.Errorf("%s.%s: %w", item.node.Name, item.command.Name, err)
	}
	// The cache decision is bracketed by the frozen source: a reuse verdict is
	// only meaningful if the identity it was taken for still holds immediately
	// before and immediately after the lookup.
	var inspection buildcache.Result
	if err := source.Use(func(*buildsource.Token) error {
		inspection = cache.Inspect(buildcache.Expectation{Input: input})
		return nil
	}); err != nil {
		return PlannedBuild{}, fmt.Errorf("%s.%s: %w", item.node.Name, item.command.Name, err)
	}
	build := PlannedBuild{
		skill:         item.node.Name,
		command:       item.command.Name,
		commandObject: commandObject(item.command),
		buildRoot:     buildRoot,
		sourceDir:     item.command.SourceDir,
		source:        input.BuildSource,
		target:        target,
		input:         input,
		key:           key,
		outcome:       BuildOutcome(inspection.DryRunOutcome()),
		reason:        inspection.Reason,
	}
	if build.outcome == BuildCacheHit {
		build.artifactPath = inspection.ArtifactPath
		build.receiptHash = inspection.ReceiptHash
		build.artifact = inspection.Receipt.Artifact
	}
	return build, nil
}

// commandObject reproduces the exact package-declared build-command surface.
// The parser admits only these three fields for a build command, so anything
// else in the manifest has already been rejected.
func commandObject(command skillspec.Command) map[string]any {
	return map[string]any{
		"type":       "build",
		"driver":     command.Driver,
		"source_dir": command.SourceDir,
	}
}

// buildRootFor returns the single build root that contains the command's
// package directory. The schema v6 parser already proved the containment is
// unique, so more or fewer matches is a manager defect.
func buildRootFor(spec *skillspec.Spec, command skillspec.Command) (string, error) {
	var containing []string
	for _, root := range spec.BuildRoots {
		if pathContains(root, command.SourceDir) {
			containing = append(containing, root)
		}
	}
	if len(containing) != 1 {
		return "", fmt.Errorf("source_dir %q is not below exactly one build root", command.SourceDir)
	}
	return containing[0], nil
}

func pathContains(root, rel string) bool {
	return rel == root || strings.HasPrefix(rel, root+"/")
}

// plannedCommand pairs one node with one of its active build commands.
type plannedCommand struct {
	node    *closure.Node
	command skillspec.Command
}

// plannedCommands enumerates active build commands provider-first across the
// closure order and command-lexical within each node.
func plannedCommands(nodes []*closure.Node) []plannedCommand {
	var planned []plannedCommand
	for _, node := range nodes {
		active := node.ActiveCommands()
		for _, name := range node.ActiveCommandNames() {
			command := node.Spec.Commands[name]
			if command.Type != "build" || !active[name] {
				continue
			}
			planned = append(planned, plannedCommand{node: node, command: command})
		}
	}
	return planned
}

// buildNodes returns the closure nodes with at least one active build command,
// preserving provider-first order.
func buildNodes(nodes []*closure.Node) []*closure.Node {
	var selected []*closure.Node
	seen := map[string]bool{}
	for _, item := range plannedCommands(nodes) {
		if seen[item.node.Name] {
			continue
		}
		seen[item.node.Name] = true
		selected = append(selected, item.node)
	}
	return selected
}
