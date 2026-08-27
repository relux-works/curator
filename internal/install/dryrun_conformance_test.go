package install

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envfiles"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/scopes"
)

// authoritativeDryRunCase is the published no-mutation contract. Only the field
// names are mirrored; the forbidden surfaces themselves are read from the
// authoritative document at run time.
//
// The fields below ForbiddenPersistentEffects are published only by a
// compiled-build case. They are zero for a case that names no compiled command,
// and every binding that reads one says what it does when it is zero, so a root
// that publishes fewer fields never turns an assertion into a silent pass.
type authoritativeDryRunCase struct {
	Name                       string   `json:"name"`
	Scope                      string   `json:"scope"`
	ForbiddenPersistentEffects []string `json:"forbidden_persistent_effects"`
	AllowedGoCommands          []string `json:"allowed_go_commands"`
	ForbiddenGoCommands        []string `json:"forbidden_go_commands"`
	ArtifactExecuted           bool     `json:"artifact_executed"`
	LogicalCacheKey            string   `json:"logical_cache_key"`
	OperationPrivateStateAfter string   `json:"operation_private_state_after"`
	ReportedBuildOutcomes      []string `json:"reported_build_outcomes"`
}

// authoritativeCompiledFixture is the published compiled-build identity the
// multi-project dry-run case shares its logical cache key with.
type authoritativeCompiledFixture struct {
	CacheKey        string `json:"cache_key"`
	ExecutionPolicy string `json:"execution_policy"`
	LogicalCommand  string `json:"logical_command"`
}

// authoritativeLifecycleDocument is the part of the published manager-lifecycle
// suite this file binds.
type authoritativeLifecycleDocument struct {
	DryRunCases          []authoritativeDryRunCase    `json:"dry_run_cases"`
	CompiledBuildFixture authoritativeCompiledFixture `json:"compiled_build_fixture"`
}

func authoritativeLifecycle(t *testing.T) authoritativeLifecycleDocument {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-lifecycle.json")) // #nosec G304 -- explicit authoritative conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document authoritativeLifecycleDocument
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.DryRunCases) == 0 {
		t.Fatal("the authoritative suite publishes no dry-run case to bind")
	}
	return document
}

// dryRunBaseline is everything a published forbidden effect can be measured
// against: the surfaces that must stay absent and the bytes that must not move.
type dryRunBaseline struct {
	env           *env
	fetchable     string
	clonedName    string
	clonedOrigin  string
	configBytes   []byte
	registryCalls func() int
	// projects is every project root the case planned, so a project-scoped
	// forbidden effect is asserted for all of them and not only for the first.
	projects []string
	// tempBase is the isolated process temporary directory the run allocated its
	// operation-private roots inside.
	tempBase string
	// private is the operation-private root the real trusted-toolchain boundary
	// probed inside, and is nil for a case that activates no compiled command.
	private   *privateRoot
	toolchain *observedToolchain
	builder   *refusingBuilder
}

// dryRunGitignore is the ignore file every scope in this file needs: a dry run
// still runs the gitignore gate, and a project that would commit its generated
// context fails that gate before anything is planned.
const dryRunGitignore = ".agents/\n.claude/skills/\nSkillfile.dev.json\n"

// The compiled skill the multi-project case declares in every project, so both
// projects derive one shared logical cache key against one protected boundary.
const (
	compiledDryRunSkill   = "compiled-skill"
	compiledDryRunCommand = "shared-tool"
)

// observedToolchain wraps the real trusted-toolchain boundary and counts the two
// calls a plan can make against it. Probing is the package-independent form a
// dry run may take; establishing a session is the form that would make a
// source-aware Go command, a module cache and a build cache reachable at all.
type observedToolchain struct {
	inner       Toolchain
	probes      int
	establishes int
}

func (observed *observedToolchain) Probe(ctx context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	observed.probes++
	return observed.inner.Probe(ctx)
}

func (observed *observedToolchain) Establish(ctx context.Context) (BuildSession, error) {
	observed.establishes++
	return observed.inner.Establish(ctx)
}

// refusingBuilder is the source-aware boundary of a plan: the only thing in an
// installation that runs `go list` or `go build`. A dry run must never reach it,
// so every call is recorded and refused instead of served.
type refusingBuilder struct{ calls []string }

func (builder *refusingBuilder) Stage(_ context.Context, request StageRequest) (StagedArtifact, error) {
	builder.calls = append(builder.calls, request.Command)
	return StagedArtifact{}, fmt.Errorf("a dry run must not run a source-aware Go command")
}

// armedDryRunEnv builds a machine where every published persistent surface is
// genuinely reachable: one skill whose repository can really be fetched, one
// skill that can only be resolved by cloning, the audit gate armed, and a
// trusted registry that answers over loopback. Without arming them, "absent
// afterwards" would prove nothing.
func armedDryRunEnv(t *testing.T) *dryRunBaseline {
	t.Helper()
	e := newEnv(t)

	// A repository that a real run would fetch: its origin is a working local
	// upstream, so a fetch would succeed and leave FETCH_HEAD behind.
	e.skill("fetchable")
	upstream := filepath.Join(t.TempDir(), "upstream")
	e.git(t.TempDir(), "clone", "-q", "--bare", filepath.Join(e.skillsRoot, "fetchable"), upstream)
	e.git(filepath.Join(e.skillsRoot, "fetchable"), "remote", "add", "origin", upstream)

	// A skill that is absent from the skills root, so resolution has to clone.
	e.skill("clonable")
	movedOrigin := e.relocate("clonable", filepath.Join(t.TempDir(), "outside"))

	e.cfg.Audit.Enabled = true
	calls := 0
	server, pinned := loopbackRegistry(t, &calls)
	t.Cleanup(server.Close)
	e.cfg.AuditRegistries = []config.Registry{
		{Name: "loopback", URL: server.URL, PublicKeys: []string{pinned}, Enabled: true},
	}

	// config.Bootstrap writes a real configuration document, so "configuration"
	// can be asserted byte for byte rather than by absence.
	if err := config.Bootstrap(e.cfg.Path, e.skillsRoot, "", []string{"claude_code"}, false); err != nil {
		t.Fatal(err)
	}
	configBytes, err := os.ReadFile(e.cfg.Path) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	return &dryRunBaseline{
		env: e, fetchable: filepath.Join(e.skillsRoot, "fetchable"),
		clonedName: "clonable", clonedOrigin: movedOrigin,
		configBytes: configBytes, registryCalls: func() int { return calls },
		projects: []string{e.project},
	}
}

// loopbackRegistry is a trusted registry that answers over loopback and counts
// the requests it served, so a response cache can only stay empty because the
// run refused to persist one — not because nothing was ever fetched.
func loopbackRegistry(t *testing.T, calls *int) (*httptest.Server, string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pinned := "ed25519:" + base64.StdEncoding.EncodeToString(public)
	sign := func(body map[string]any) map[string]any {
		signature := ed25519.Sign(private, registry.CanonicalBytes(body))
		body["sig"] = map[string]any{
			"key_id": registry.KeyID(public), "algorithm": "ed25519",
			"signature": base64.StdEncoding.EncodeToString(signature),
		}
		return body
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/snapshot"):
			_ = json.NewEncoder(w).Encode(sign(map[string]any{
				"schema_version": 1, "merkle_root": strings.Repeat("a", 64), "log_size": 1,
				"head": strings.Repeat("b", 64), "version": 1,
				"created_at": time.Now().UTC().Format(time.RFC3339),
			}))
		case strings.HasSuffix(r.URL.Path, "/v1/records"):
			_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{}, "next_cursor": nil})
		default:
			http.NotFound(w, r)
		}
	}))
	return server, pinned
}

// declareEvery writes one scope manifest naming the fetchable skill, the skill
// that can only be cloned, and any further skill the case armed.
func (baseline *dryRunBaseline) declareEvery(t *testing.T, root string, extra ...string) {
	t.Helper()
	skills := []map[string]any{
		{"name": "fetchable", "tag": "v1"},
		{"name": baseline.clonedName, "git": baseline.clonedOrigin, "tag": "v1"},
	}
	for _, name := range extra {
		skills = append(skills, map[string]any{"name": name, "tag": "v1"})
	}
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"claude_code"},
		"skills":         skills,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	baseline.env.write(root, manifest.Name, string(payload))
}

// TestAuthoritativeDryRunCasesMutateNothingPersistent binds each published
// dry-run case to a real planning run in the scope it names, then proves every
// published forbidden effect individually. An effect with no binding fails the
// test rather than passing unasserted.
//
// The test is deliberately not parallel. Every case isolates the process
// temporary directory so the operation-private roots under test are the only
// ones in it, and `t.Setenv` — the only portable way to do that — forbids a
// parallel ancestor. Nothing about the assertions depends on it: the cases run
// in sequence and each still arms, plans, and measures its own machine.
func TestAuthoritativeDryRunCasesMutateNothingPersistent(t *testing.T) {
	document := authoritativeLifecycle(t)
	for _, published := range document.DryRunCases {
		published := published
		t.Run(published.Name, func(t *testing.T) {
			if len(published.ForbiddenPersistentEffects) == 0 {
				t.Fatalf("published dry-run case %q forbids nothing", published.Name)
			}
			baseline := armedDryRunEnv(t)
			e := baseline.env

			var results []Result
			switch published.Scope {
			case "project":
				baseline.declareEvery(t, e.project)
				e.write(e.project, ".gitignore", dryRunGitignore)
				baseline.tempBase = isolateTempDir(t)
				results = []Result{Project(e.cfg, e.project, "test",
					Options{Platform: "unix", DryRun: true, Fetch: true})}
			case "global":
				if _, err := GlobalInit(e.home); err != nil {
					t.Fatal(err)
				}
				baseline.declareEvery(t, GlobalRoot(e.home))
				userHome := t.TempDir()
				baseline.tempBase = isolateTempDir(t)
				results = []Result{Global(e.cfg, userHome,
					Options{Platform: "unix", DryRun: true, Fetch: true})}
			case "multi-project":
				results = baseline.planEveryProject(t, published, document)
			default:
				t.Fatalf("published dry-run scope %q has no executable binding", published.Scope)
			}
			for index, result := range results {
				if result.Status != "ok" {
					t.Fatalf("the %s dry run of project %d did not plan: %+v",
						published.Scope, index, result)
				}
				// The plan really did the work whose side effects are forbidden: it
				// resolved every declared skill, one of which only exists remotely.
				if !strings.Contains(strings.Join(result.Messages, "\n"), baseline.clonedName) {
					t.Fatalf("the dry run never resolved the skill it had to clone:\n%s",
						strings.Join(result.Messages, "\n"))
				}
			}
			// The registry surfaces can only be proven unwritten if the run
			// actually reached a registry.
			for _, effect := range published.ForbiddenPersistentEffects {
				if effect != "response-cache" && effect != "registry-state" {
					continue
				}
				if baseline.registryCalls() == 0 {
					t.Fatalf("no registry was contacted, so %q proves nothing", effect)
				}
				break
			}
			for _, effect := range published.ForbiddenPersistentEffects {
				baseline.assertNoEffect(t, effect)
			}
			baseline.assertNoOperationPrivateState(t, published)
		})
	}
}

// planEveryProject binds the published multi-project scope to Curator's own
// multi-project operation. `curator install --all` and `curator upgrade --all`
// plan every selected project target in one run — one manager home, one skills
// root, one shared fetch-deduplication set — and profiles/manager.md §2.5 takes
// those targets in the unsigned byte order of their canonical project
// identities, which is the order used here.
//
// Both projects declare the same compiled command, so both derive one shared
// logical cache key and both miss the same protected cache entry. That shared
// miss is the surface the case exists to protect: two projects looking at one
// absent entry must still leave the whole machine byte-for-byte alone.
//
// The trusted toolchain and the protected build cache are the real ones, so the
// compiled surfaces the case forbids are genuinely reachable rather than
// vacuously absent. The real probe does run `go telemetry off`, `go version` and
// `go env` inside a private base, and the real cache boundary is really
// inspected for the shared key. Only the source-aware builder is replaced, by
// one that records and refuses every call a dry run must never make.
func (baseline *dryRunBaseline) planEveryProject(
	t *testing.T, published authoritativeDryRunCase, document authoritativeLifecycleDocument,
) []Result {
	t.Helper()
	e := baseline.env
	// The published key is derived from the published toolchain and target, which
	// this host does not have; internal/buildmeta binds that derivation directly
	// against expected/build-driver/cache-key.txt. What belongs here is that the
	// case and the fixture name one build, and that the two projects planned
	// against one key of their own — the shared entry, not its absolute value.
	fixture := document.CompiledBuildFixture
	if published.LogicalCacheKey == "" || published.LogicalCacheKey != fixture.CacheKey {
		t.Fatalf("the published multi-project key %q is not the published compiled fixture key %q, so the two documents describe different builds",
			published.LogicalCacheKey, fixture.CacheKey)
	}
	if fixture.ExecutionPolicy != buildmeta.ExecutionPolicy {
		t.Fatalf("the published shared key is derived under execution policy %q, which is not the %q Curator implements",
			fixture.ExecutionPolicy, buildmeta.ExecutionPolicy)
	}

	e.buildSkill(compiledDryRunSkill, compiledDryRunCommand)
	second := t.TempDir()
	e.git(second, "init", "-q")
	baseline.projects = append(baseline.projects, second)
	for _, project := range baseline.projects {
		baseline.declareEvery(t, project, compiledDryRunSkill)
		e.write(project, ".gitignore", dryRunGitignore)
	}
	ordered := canonicalProjectOrder(t, baseline.projects)

	// The real trusted-toolchain boundary, over a private root this test owns so
	// the probe's own base can be proven gone afterwards. The forbidden roots are
	// the production ones — every project checkout, the runtime store and the
	// skills root — widened to cover both projects of the operation.
	forbidden := append([]string{filepath.Join(e.home, "runtime"), e.skillsRoot}, baseline.projects...)
	baseline.private = &privateRoot{prefix: operationPrivatePrefix}
	baseline.toolchain = &observedToolchain{inner: &goToolchain{private: baseline.private, forbidden: forbidden}}
	baseline.builder = &refusingBuilder{}
	baseline.tempBase = isolateTempDir(t)

	// One fetch-deduplication set for the whole operation, exactly as the
	// `--all` caller wires it.
	fetched := map[string]bool{}
	results := make([]Result, 0, len(ordered))
	for _, project := range ordered {
		results = append(results, Project(e.cfg, project, filepath.Base(project), Options{
			Platform: "unix", DryRun: true, Fetch: true, FetchedRepos: fetched,
			Build: BuildDeps{Toolchain: baseline.toolchain, Builder: baseline.builder},
		}))
	}
	baseline.assertCompiledCase(t, published, results)
	return results
}

// canonicalProjectOrder returns the project roots in the unsigned byte order of
// their canonical project identities — the order profiles/manager.md §2.5
// requires a multi-project operation to take its targets in. The roots
// themselves are returned rather than the identities, so the surfaces this test
// arms and measures stay the ones it created.
func canonicalProjectOrder(t *testing.T, projects []string) []string {
	t.Helper()
	identities, err := managerlock.CanonicalProjects(projects...)
	if err != nil {
		t.Fatal(err)
	}
	byIdentity := make(map[string]string, len(projects))
	for _, project := range projects {
		identity, identityErr := managerlock.CanonicalProject(project)
		if identityErr != nil {
			t.Fatal(identityErr)
		}
		byIdentity[string(identity)] = project
	}
	ordered := make([]string, 0, len(identities))
	for _, identity := range identities {
		project, ok := byIdentity[string(identity)]
		if !ok {
			t.Fatalf("canonical identity %q belongs to no project of this operation", identity)
		}
		ordered = append(ordered, project)
	}
	return ordered
}

// assertCompiledCase proves the published compiled-build fields of one dry-run
// case: the outcome vocabulary it may report, the one logical key its projects
// share, the artifact it must not produce, and the Go commands it may and may
// not run.
func (baseline *dryRunBaseline) assertCompiledCase(
	t *testing.T, published authoritativeDryRunCase, results []Result,
) {
	t.Helper()
	if published.ArtifactExecuted {
		t.Fatalf("published case %q executes the artifact a dry run only plans", published.Name)
	}
	reported := map[string]bool{}
	for _, outcome := range published.ReportedBuildOutcomes {
		reported[outcome] = true
	}
	if len(reported) == 0 {
		t.Fatalf("published case %q reports no build outcome to bind", published.Name)
	}
	keys := make([]string, 0, len(results))
	for index, result := range results {
		if !result.BuildsComplete {
			t.Fatalf("project %d described only part of the compiled closure: %+v", index, result)
		}
		if len(result.Builds) != 1 {
			t.Fatalf("project %d planned %d compiled commands, want the one shared command",
				index, len(result.Builds))
		}
		build := result.Builds[0]
		if !reported[string(build.Outcome())] {
			t.Fatalf("project %d reported build outcome %q, which the published case does not admit: %v",
				index, build.Outcome(), published.ReportedBuildOutcomes)
		}
		// The shared protected entry is absent, and the case name says what that
		// has to report: a miss, planned without preflighting or building.
		if build.Outcome() != BuildWouldPreflightAndBuild {
			t.Fatalf("project %d reported %q against an empty protected cache, want %q",
				index, build.Outcome(), BuildWouldPreflightAndBuild)
		}
		if len(result.Staged) != 0 {
			t.Fatalf("project %d staged an artifact during a dry run: %+v", index, result.Staged)
		}
		if got := build.ArtifactPath(); got != "" {
			t.Fatalf("project %d reported the artifact %q for a cache miss", index, got)
		}
		keys = append(keys, string(build.CacheKey()))
	}
	if len(keys) < 2 {
		t.Fatalf("the multi-project binding planned %d projects, want at least two", len(keys))
	}
	for _, key := range keys[1:] {
		if key == "" || key != keys[0] {
			t.Fatalf("the projects derived the different logical cache keys %v, so no shared entry was under test", keys)
		}
	}
	// The published allowed commands are exactly the package-independent probe
	// the real toolchain boundary takes: `go telemetry off`, `go version` and
	// `go env`. Taking it once per project is what makes the probe surfaces
	// reachable, so "absent afterwards" says something.
	if len(published.AllowedGoCommands) == 0 {
		t.Fatalf("published case %q allows no package-independent Go command", published.Name)
	}
	if baseline.toolchain.probes != len(results) {
		t.Fatalf("the allowed commands %v were taken %d times, want once per project (%d)",
			published.AllowedGoCommands, baseline.toolchain.probes, len(results))
	}
	// The forbidden commands are source-aware. A dry run reaches them only by
	// establishing a build session and handing a miss to the builder.
	if len(published.ForbiddenGoCommands) == 0 {
		t.Fatalf("published case %q forbids no source-aware Go command", published.Name)
	}
	if baseline.toolchain.establishes != 0 {
		t.Fatalf("a dry run established %d build sessions, which is what makes %v reachable",
			baseline.toolchain.establishes, published.ForbiddenGoCommands)
	}
	if len(baseline.builder.calls) != 0 {
		t.Fatalf("a dry run reached the source-aware builder for %v, so %v ran",
			baseline.builder.calls, published.ForbiddenGoCommands)
	}
}

// assertNoOperationPrivateState proves the published operation-private contract.
// The base the real toolchain probed in is gone, and once the private root this
// test owns is removed nothing named after an operation-private root is left in
// the isolated temporary directory at all — so every root the run allocated for
// itself was released too.
func (baseline *dryRunBaseline) assertNoOperationPrivateState(
	t *testing.T, published authoritativeDryRunCase,
) {
	t.Helper()
	switch published.OperationPrivateStateAfter {
	case "", "absent":
	default:
		t.Fatalf("published case %q leaves operation-private state %q behind, want absent",
			published.Name, published.OperationPrivateStateAfter)
	}
	if baseline.private != nil {
		requireNoToolchainBases(t, baseline.private, "a "+published.Scope+" dry run")
		if err := baseline.private.remove(); err != nil {
			t.Fatal(err)
		}
	}
	requireNoOperationPrivateRoots(t, baseline.tempBase, "the "+published.Scope+" dry run")
}

func (baseline *dryRunBaseline) assertNoEffect(t *testing.T, effect string) {
	t.Helper()
	if visible := baseline.effectSurfaces(t, effect); len(visible) > 0 {
		t.Fatalf("the dry run produced the forbidden persistent effect %q at %s",
			effect, strings.Join(visible, ", "))
	}
}

// effectSurfaces is the executable binding of one published forbidden effect: it
// reports every surface where that effect is currently visible.
//
// The binding is a probe rather than an assertion so it can be run in both
// directions. After a dry run it must report nothing, which is the published
// contract. After a real operation that genuinely produces the effect it must
// report something, which is what
// TestDryRunEffectBindingsSeeWhatARealOperationWrites proves — a binding that
// cannot see its own surface is indistinguishable from no binding at all.
//
// An effect the manager has no surface for is a fatal gap, not an empty result.
func (baseline *dryRunBaseline) effectSurfaces(t *testing.T, effect string) []string {
	t.Helper()
	e := baseline.env
	// The trees a name-scoped or prefix-scoped binding scans: every project the
	// operation planned, the manager home, and the skills root. The isolated
	// temporary directory joins them for a surface that could survive there
	// instead of in persistent state.
	trees := append(append([]string{}, baseline.projects...), e.home, e.skillsRoot)
	ephemeral := trees
	if baseline.tempBase != "" {
		ephemeral = append(append([]string{}, trees...), baseline.tempBase)
	}
	// The protected build cache is one root: the compiled artifacts, the
	// quarantine namespace an unusable predecessor is moved into, and the only
	// modes a manager ever establishes or repairs all live below it.
	protectedCache := filepath.Join(e.home, "cache", "build")
	lockRoot := filepath.Join(e.home, "state", "locks")
	switch effect {
	case "source-fetch":
		return existingPaths(t, filepath.Join(baseline.fetchable, ".git", "FETCH_HEAD"))
	case "source-clone", "source-checkout":
		// rc.3 calls it a clone and the rc.6 compiled case calls it a checkout.
		// Both name the persistent destinations a real resolution would
		// materialise the missing repository into.
		return existingPaths(t, filepath.Join(e.skillsRoot, baseline.clonedName),
			filepath.Join(e.home, "dev"))
	case "snapshot-cache":
		return existingPaths(t, filepath.Join(e.home, "cache"))
	case "response-cache":
		return existingPaths(t, filepath.Join(e.home, "cache", "registry"))
	case "toolchain-probe-memo":
		// Curator memoises no toolchain identity anywhere persistent. The real
		// probe allocates its base inside the operation-private root and removes
		// it before returning, so no base may survive anywhere either.
		return pathsPrefixed(t, []string{"go-probe-base-"}, ephemeral)
	case "module-cache":
		// GOPATH and GOMODCACHE exist only inside an established session's private
		// base, and a dry run establishes none.
		return pathsNamed(t, []string{"gopath", "gomodcache"}, ephemeral)
	case "go-build-cache":
		return append(pathsNamed(t, []string{"gocache"}, ephemeral),
			pathsPrefixed(t, []string{"go-build-base-"}, ephemeral)...)
	case "compiled-artifact-cache":
		return existingPaths(t, protectedCache)
	case "quarantine":
		return append(existingPaths(t, protectedCache),
			pathsPrefixed(t, []string{quarantinedEntryPrefix}, trees)...)
	case "permission-repair":
		// The protected cache boundary is the only surface whose modes a manager
		// establishes or repairs; an untrusted entry is quarantined rather than
		// permission-repaired. An absent boundary was never created, so no mode on
		// it was ever set or corrected.
		return existingPaths(t, protectedCache)
	case "audit-state":
		return existingPaths(t, filepath.Join(e.home, "audit"))
	case "registry-state":
		return existingPaths(t, filepath.Join(e.home, "state", "registry"))
	case "revocation-state":
		// Curator keeps no separate revocation store: a revocation is decided from
		// signed registry records and the operator's configuration. The only
		// persistent state such a decision could leave is the registry rollback
		// state below the manager's state root, so the whole root must be absent.
		return existingPaths(t, filepath.Join(e.home, "state"))
	case "configuration":
		payload, err := os.ReadFile(e.cfg.Path) // #nosec G304 -- test fixture path
		if err != nil {
			return []string{e.cfg.Path + ": unreadable: " + err.Error()}
		}
		if string(payload) != string(baseline.configBytes) {
			return []string{e.cfg.Path + ": rewritten"}
		}
		return nil
	case "project-lock":
		return append(existingPaths(t, filepath.Join(lockRoot, "v1", "projects")),
			pathsSuffixed(t, []string{lockFileSuffix}, baseline.projects)...)
	case "cache-build-lock":
		return existingPaths(t, filepath.Join(lockRoot, "v1", "build"))
	case "manager-home-lock":
		// The home lock file sits directly in the versioned lock root, so the whole
		// namespace must be absent: no lock of any class was created.
		return existingPaths(t, lockRoot)
	case "journal":
		return existingPaths(t, filepath.Join(e.home, "state", "transactions"))
	case "backup":
		// A transaction writes its backup as a sidecar next to every live target it
		// is about to swap, under one reserved name prefix.
		return pathsPrefixed(t, []string{transactionSidecarPrefix}, trees)
	case "runtime", "runtime-tree":
		return existingPaths(t, filepath.Join(e.home, "runtime"))
	case "context-tree":
		var contexts []string
		for _, project := range baseline.projects {
			contexts = append(contexts, filepath.Join(project, ".agents", "skills"))
			for _, agent := range e.cfg.DefaultAgents {
				contexts = append(contexts,
					filepath.Join(project, filepath.FromSlash(adapters.AgentPaths[agent])))
			}
		}
		return existingPaths(t, contexts...)
	case "environment-file":
		envDirs := []string{envfiles.GlobalDir(e.home)}
		for _, project := range baseline.projects {
			envDirs = append(envDirs, envfiles.ProjectDir(project))
		}
		var helpers []string
		for _, dir := range envDirs {
			helpers = append(helpers,
				filepath.Join(dir, envfiles.ShellName), filepath.Join(dir, envfiles.PowerShellName))
		}
		return existingPaths(t, helpers...)
	case "install-marker":
		return pathsNamed(t, []string{marker.Name}, trees)
	case "command-shim":
		shims := []string{filepath.Join(e.home, "global", "bin")}
		for _, project := range baseline.projects {
			shims = append(shims, filepath.Join(project, ".agents", "bin"))
		}
		return existingPaths(t, shims...)
	case "adapter-ledger":
		return pathsNamed(t, []string{adapters.LedgerName}, trees)
	case "adapter-mirror":
		return existingPaths(t, filepath.Join(e.home, "global", "skills"),
			scopes.HybridSkillsRoot(e.home))
	case "consumer-ledger":
		return existingPaths(t, filepath.Join(e.home, scopes.ConsumersName))
	case "gc-metadata":
		// Collection keeps no metadata of its own: it reads the consumer ledger and
		// sweeps the runtime store and the protected build cache in place. Those
		// three are its entire persistent footprint.
		return existingPaths(t, filepath.Join(e.home, scopes.ConsumersName),
			filepath.Join(e.home, "runtime"), protectedCache)
	case "project-artifacts":
		var artifacts []string
		for _, project := range baseline.projects {
			artifacts = append(artifacts,
				filepath.Join(project, ".agents"), filepath.Join(project, ".claude"))
		}
		return existingPaths(t, artifacts...)
	case "global-artifacts":
		return existingPaths(t, filepath.Join(e.home, "global", "skills"),
			filepath.Join(e.home, "global", "bin"))
	default:
		t.Fatalf("published forbidden effect %q has no executable binding", effect)
		return nil
	}
}

// The reserved name shapes a binding matches on. Each one is owned by the
// package that writes it; repeating it here is what lets a whole-tree scan find
// the surface wherever a manager chose to put it.
const (
	// quarantinedEntryPrefix mirrors internal/buildcache's quarantinePrefix.
	quarantinedEntryPrefix = ".quarantine-"
	// transactionSidecarPrefix mirrors the sidecar names internal/transaction
	// writes beside every live target it is about to swap.
	transactionSidecarPrefix = ".curator-txn-"
	// lockFileSuffix is the extension every cross-process lock file carries.
	lockFileSuffix = ".lock"
)

// existingPaths reports which of these paths exist. A path that cannot be
// stat'ed at all is reported too: an unreadable surface is not an absent one.
func existingPaths(t *testing.T, paths ...string) []string {
	t.Helper()
	var visible []string
	for _, path := range paths {
		_, err := os.Lstat(path)
		switch {
		case err == nil:
			visible = append(visible, path)
		case !os.IsNotExist(err):
			visible = append(visible, path+": "+err.Error())
		}
	}
	return visible
}

// pathsNamed reports everything below the roots carrying one of these exact base
// names. It binds a surface identified by the name a manager gives it rather
// than by one fixed location.
func pathsNamed(t *testing.T, names []string, roots []string) []string {
	t.Helper()
	return pathsMatching(t, roots, func(name string) bool {
		for _, wanted := range names {
			if name == wanted {
				return true
			}
		}
		return false
	})
}

// pathsPrefixed is pathsNamed for a reserved name prefix: the per-operation
// suffix differs on every run, the prefix never does.
func pathsPrefixed(t *testing.T, prefixes []string, roots []string) []string {
	t.Helper()
	return pathsMatching(t, roots, func(name string) bool {
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				return true
			}
		}
		return false
	})
}

// pathsSuffixed is pathsNamed for a reserved extension.
func pathsSuffixed(t *testing.T, suffixes []string, roots []string) []string {
	t.Helper()
	return pathsMatching(t, roots, func(name string) bool {
		for _, suffix := range suffixes {
			if strings.HasSuffix(name, suffix) {
				return true
			}
		}
		return false
	})
}

func pathsMatching(t *testing.T, roots []string, match func(name string) bool) []string {
	t.Helper()
	var visible []string
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir // a source repository is not manager state
			}
			if match(info.Name()) {
				visible = append(visible, path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	return visible
}

// requireNoOperationPrivateRoots asserts the isolated temporary directory holds
// no operation-private root of this manager at all. It is narrower than
// requireTempDirEmpty on purpose: an unrelated temporary file from a child
// process is not a persistent effect of the run under test.
func requireNoOperationPrivateRoots(t *testing.T, base, context string) {
	t.Helper()
	for _, name := range tempEntries(t, base) {
		if strings.HasPrefix(name, operationPrivatePrefix) {
			t.Fatalf("%s left the operation-private root %s behind", context, name)
		}
	}
}

// TestDryRunEffectBindingsSeeWhatARealOperationWrites is the anti-vacuity
// regression for the bindings of
// TestAuthoritativeDryRunCasesMutateNothingPersistent. Every published forbidden
// effect is a claim that a surface stayed absent, and a binding that could never
// see its surface would make that claim unfalsifiable — the exact way a
// no-mutation contract passes without proving anything.
//
// The witness is a real operation in the same armed machine. A real project
// install, a real global install and a real locked operation genuinely produce
// most of these surfaces in their production locations, and every binding must
// then report them. The few surfaces no compiler-free operation here can reach
// are produced through the same allocators and reserved names their owning
// packages write them under, with the reason recorded beside each.
//
// Completeness is part of the test: every effect any published case names must
// be witnessed. A future effect that gains a binding without gaining a witness
// fails here instead of passing silently.
func TestDryRunEffectBindingsSeeWhatARealOperationWrites(t *testing.T) {
	document := authoritativeLifecycle(t)
	published := map[string]bool{}
	for _, testCase := range document.DryRunCases {
		for _, effect := range testCase.ForbiddenPersistentEffects {
			published[effect] = true
		}
	}
	if len(published) == 0 {
		t.Fatal("the authoritative suite publishes no forbidden effect to witness")
	}

	baseline := armedDryRunEnv(t)
	e := baseline.env
	baseline.declareEvery(t, e.project)
	e.write(e.project, ".gitignore", dryRunGitignore)
	if result := Project(e.cfg, e.project, "test",
		Options{Platform: "unix", Fetch: true}); result.Status != "ok" {
		t.Fatalf("the real project install failed: %+v", result)
	}
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	baseline.declareEvery(t, GlobalRoot(e.home))
	if result := Global(e.cfg, t.TempDir(), Options{Platform: "unix", Fetch: true}); result.Status != "ok" {
		t.Fatalf("the real global install failed: %+v", result)
	}
	takeEveryLockClass(t, e.home, e.project)
	// Isolation comes last: the surfaces produced below are allocated inside the
	// operation-private root, and the binding for them scans this base.
	baseline.tempBase = isolateTempDir(t)
	baseline.produceSurfacesNoRealRunHereReaches(t)

	effects := make([]string, 0, len(published))
	for effect := range published {
		effects = append(effects, effect)
	}
	sort.Strings(effects)
	for _, effect := range effects {
		if visible := baseline.effectSurfaces(t, effect); len(visible) == 0 {
			t.Fatalf("the binding for %q reported nothing after a real operation produced it, so its absence after a dry run proves nothing",
				effect)
		}
	}
}

// takeEveryLockClass acquires and releases one lock of every class a real
// operation can take, so the lock bindings are witnessed against the real
// manager-lock layout rather than a guessed one.
func takeEveryLockClass(t *testing.T, home, project string) {
	t.Helper()
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(false)
	defer func() {
		if closeErr := operation.Close(); closeErr != nil {
			t.Errorf("release the manager locks: %v", closeErr)
		}
	}()
	if _, err := operation.AcquireProjects(context.Background(), project); err != nil {
		t.Fatal(err)
	}
	if err := operation.AcquireBuildKey(context.Background(), "sha256:"+strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
}

// produceSurfacesNoRealRunHereReaches creates the remaining published surfaces.
// Each one is unreachable from a compiler-free operation in this package, and
// each is produced the way its owning package writes it.
func (baseline *dryRunBaseline) produceSurfacesNoRealRunHereReaches(t *testing.T) {
	t.Helper()
	e := baseline.env

	// Configuration. A real install of an already-registered project rewrites
	// nothing, so the byte-for-byte binding is witnessed by changing the bytes.
	if err := os.WriteFile(e.cfg.Path, append(baseline.configBytes, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	// The protected build cache root and one quarantined predecessor below it.
	// Publishing a real entry needs a compiled artifact and a real receipt;
	// internal/buildcache owns both paths and asserts them directly.
	e.seedLiveCache()
	if err := os.MkdirAll(filepath.Join(e.home, "cache", "build", buildmeta.DriverGoV1,
		quarantinedEntryPrefix+"displaced"), 0o700); err != nil {
		t.Fatal(err)
	}

	// One transaction backup sidecar beside a live target. A committed
	// transaction removes its own sidecars, so no completed real run can leave
	// one behind; internal/transaction asserts the mid-commit state directly.
	sidecar := filepath.Join(e.project, ".agents", transactionSidecarPrefix+"aaaaaaaa-000.backup")
	if err := os.WriteFile(sidecar, []byte("prior"), 0o600); err != nil {
		t.Fatal(err)
	}

	// The operation-private toolchain bases, and the Go caches the driver lays
	// out inside an established build session. Both are allocated through the
	// same operation-private allocator production uses; a dry run establishes no
	// session, so the session base and its caches have no real path here at all.
	baseline.private = &privateRoot{prefix: operationPrivatePrefix}
	t.Cleanup(func() { _ = baseline.private.remove() })
	if _, err := baseline.private.dir("go-probe-base-"); err != nil {
		t.Fatal(err)
	}
	session, err := baseline.private.dir("go-build-base-")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"gopath", "gomodcache", "gocache"} {
		if err := os.MkdirAll(filepath.Join(session, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
}
