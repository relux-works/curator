package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
)

// The published build-driver document is the only source of expected values in
// this file. Only field names are mirrored.
type authoritativePositiveCase struct {
	Name                       string   `json:"name"`
	Result                     string   `json:"result"`
	SourceAwareGoCommands      []string `json:"source_aware_go_commands"`
	PackageIndependentCommands []string `json:"package_independent_go_commands"`
	PersistentEffects          []string `json:"persistent_effects"`
	ArtifactExecuted           bool     `json:"artifact_executed"`
	ProtectedBoundaryVerified  bool     `json:"protected_boundary_verified"`
}

type authoritativeRejectionCase struct {
	Name     string `json:"name"`
	Boundary string `json:"boundary"`
	Expected struct {
		Result           string `json:"result"`
		Error            string `json:"error"`
		Reuse            bool   `json:"reuse"`
		ArtifactExecuted bool   `json:"artifact_executed"`
	} `json:"expected"`
}

type authoritativeDriverDocument struct {
	PositiveCases  []authoritativePositiveCase  `json:"positive_cases"`
	RejectionCases []authoritativeRejectionCase `json:"rejection_cases"`
}

func authoritativeDriverCases(t *testing.T) authoritativeDriverDocument {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json")) // #nosec G304 -- explicit authoritative conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document authoritativeDriverDocument
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

// TestAuthoritativeCacheOutcomesDriveInstallation binds every published cache
// outcome to a real installation plan. The planner's outcome vocabulary is the
// published vocabulary, so the published value is compared directly rather than
// through a private translation table.
func TestAuthoritativeCacheOutcomesDriveInstallation(t *testing.T) {
	t.Parallel()
	document := authoritativeDriverCases(t)
	bound := 0
	for _, published := range document.PositiveCases {
		published := published
		switch published.Result {
		case "":
			continue
		case "accepted":
			// Plain acceptance is the identity, environment, argv, and context
			// cluster, owned by the driver and manifest consumers. It carries no
			// cache verdict for an installation to reproduce.
			continue
		case string(BuildCacheHit):
			bound++
			t.Run(published.Name, func(t *testing.T) { runPublishedCacheHit(t, published) })
		case string(BuildWouldPreflightAndBuild):
			bound++
			t.Run(published.Name, func(t *testing.T) { runPublishedCompilerFreeMiss(t, published) })
		default:
			t.Fatalf("published cache outcome %q has no executable binding", published.Result)
		}
	}
	if bound == 0 {
		t.Fatal("the authoritative suite published no cache outcome to bind")
	}
}

// runPublishedCacheHit proves a protected hit is reused without any
// source-aware Go command and without staging or executing an artifact.
func runPublishedCacheHit(t *testing.T, published authoritativePositiveCase) {
	t.Helper()
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)
	cache.seedHit("alpha")

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if got := string(result.Builds[0].Outcome()); got != published.Result {
		t.Fatalf("outcome = %q, want the published %q", got, published.Result)
	}
	// The published case names no source-aware Go command, and the builder is
	// the only thing in this plan that would run one.
	if len(published.SourceAwareGoCommands) != 0 {
		t.Fatalf("published hit names source-aware commands %v", published.SourceAwareGoCommands)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("a protected hit ran source-aware Go commands: %v", builder.calls)
	}
	if !published.ArtifactExecuted && len(result.Staged) != 0 {
		t.Fatalf("a protected hit staged an artifact: %+v", result.Staged)
	}
	if published.ProtectedBoundaryVerified && !containsString(cache.inspected, "alpha") {
		t.Fatalf("the protected boundary was never inspected: %v", cache.inspected)
	}
	if got := result.Builds[0].ArtifactPath(); got != cache.byCommand["alpha"].ArtifactPath {
		t.Fatalf("hit artifact = %q, want the protected entry the boundary reported", got)
	}
}

// runPublishedCompilerFreeMiss proves a cold miss is planned without starting a
// compiler and without leaving any persistent effect behind.
func runPublishedCompilerFreeMiss(t *testing.T, published authoritativePositiveCase) {
	t.Helper()
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, toolchain, cache, builder := newFakeDeps(t)

	result := e.install(Options{DryRun: true, Build: deps})
	if result.Status != "ok" {
		t.Fatalf("dry run failed: %+v", result)
	}
	if got := string(result.Builds[0].Outcome()); got != published.Result {
		t.Fatalf("outcome = %q, want the published %q", got, published.Result)
	}
	if len(published.SourceAwareGoCommands) != 0 {
		t.Fatalf("published miss names source-aware commands %v", published.SourceAwareGoCommands)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("a compiler-free miss ran source-aware Go commands: %v", builder.calls)
	}
	// The published package-independent commands are the toolchain probe; a
	// dry run takes it and never establishes a build session on top of it.
	if len(published.PackageIndependentCommands) > 0 && toolchain.probes == 0 {
		t.Fatalf("the package-independent commands %v were never taken", published.PackageIndependentCommands)
	}
	if toolchain.establishes != 0 {
		t.Fatalf("a compiler-free miss established %d build sessions", toolchain.establishes)
	}
	if !containsString(cache.inspected, "alpha") {
		t.Fatalf("the miss never inspected the protected boundary: %v", cache.inspected)
	}
	if len(published.PersistentEffects) != 0 {
		t.Fatalf("published miss names persistent effects %v", published.PersistentEffects)
	}
	requireAbsent(t, e.persistentPaths())
}

// nonReusableStatuses are the protected-boundary verdicts that refuse an entry
// but can still be replaced. Every published cache rejection must reach one of
// them, and an installation must rebuild rather than adopt whichever it sees.
var nonReusableStatuses = []buildcache.Status{buildcache.Corrupt, buildcache.UntrustedProvenance}

// TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted binds every published
// cache-boundary rejection to a real installation. Each case is driven against
// one of the non-reusable verdicts the boundary produces — assigned by position
// so the whole set is exercised deterministically — and the installation must
// rebuild the command privately, never adopt the refused artifact, and never
// report the refused entry as a hit.
func TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted(t *testing.T) {
	t.Parallel()
	document := authoritativeDriverCases(t)
	var cases []authoritativeRejectionCase
	for _, published := range document.RejectionCases {
		if published.Boundary == "cache" {
			cases = append(cases, published)
		}
	}
	if len(cases) == 0 {
		t.Fatal("the authoritative suite published no cache-boundary rejection to bind")
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].Name < cases[j].Name })

	exercised := map[buildcache.Status]bool{}
	for index, published := range cases {
		published := published
		status := nonReusableStatuses[index%len(nonReusableStatuses)]
		exercised[status] = true
		t.Run(published.Name, func(t *testing.T) {
			if published.Expected.Result != "reject" || published.Expected.Reuse || published.Expected.ArtifactExecuted {
				t.Fatalf("published rejection %q no longer fails closed: %+v", published.Name, published.Expected)
			}
			runRefusedCacheEntry(t, published, status)
		})
	}
	for _, status := range nonReusableStatuses {
		if !exercised[status] {
			t.Fatalf("the non-reusable verdict %q was never exercised across the cluster", status)
		}
	}
}

func runRefusedCacheEntry(t *testing.T, published authoritativeRejectionCase, status buildcache.Status) {
	t.Helper()
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)

	// The refused entry still advertises an artifact path. Adopting it is
	// exactly what the published expectation forbids.
	refused := filepath.Join(t.TempDir(), "refused-artifact")
	if err := os.WriteFile(refused, []byte("refused"), 0o600); err != nil {
		t.Fatal(err)
	}
	cache.byCommand["alpha"] = buildcache.Result{
		Status: status, Reason: published.Name + ": " + published.Expected.Error, ArtifactPath: refused,
	}

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("the refused entry was not repaired: %+v", result)
	}
	if got := string(result.Builds[0].Outcome()); got == string(BuildCacheHit) {
		t.Fatalf("a refused entry was adopted as a cache hit")
	}
	if len(builder.calls) != 1 || builder.calls[0] != "alpha" {
		t.Fatalf("the refused entry was not rebuilt privately: %v", builder.calls)
	}
	if got := result.Builds[0].ArtifactPath(); got == refused {
		t.Fatalf("the installation adopted the refused artifact %q", got)
	}
	for _, staged := range result.Staged {
		if staged.Receipt().CacheKey != result.Builds[0].CacheKey() {
			t.Fatalf("the rebuild published a receipt for another key: %+v", staged.Receipt())
		}
	}
}
