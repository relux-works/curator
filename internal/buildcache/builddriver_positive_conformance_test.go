package buildcache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/closureexec"
)

// cachePositiveCase is one authoritative accepted cache-boundary vector.
type cachePositiveCase struct {
	Name                     string   `json:"name"`
	Result                   string   `json:"result"`
	CacheKey                 string   `json:"cache_key"`
	ReceiptSHA256            string   `json:"receipt_sha256"`
	ArtifactExecuted         bool     `json:"artifact_executed"`
	ProtectedBoundaryChecked bool     `json:"protected_boundary_verified"`
	SourceAwareGoCommands    []string `json:"source_aware_go_commands"`
	PackageIndependentGoCmds []string `json:"package_independent_go_commands"`
	PersistentEffects        []string `json:"persistent_effects"`
}

func loadCachePositiveCases(t *testing.T) map[string]cachePositiveCase {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json")) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-drivers vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		PositiveCases []cachePositiveCase `json:"positive_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	indexed := make(map[string]cachePositiveCase, len(vectors.PositiveCases))
	for _, testCase := range vectors.PositiveCases {
		indexed[testCase.Name] = testCase
	}
	return indexed
}

// TestProtectedCacheHitVector proves an exact protected entry is reusable, that
// the reuse is decided entirely from the protected boundary with no Go child,
// and that the artifact is handed back as a path rather than started.
func TestProtectedCacheHitVector(t *testing.T) {
	testCase, ok := loadCachePositiveCases(t)["protected-cache-hit"]
	if !ok {
		t.Skip("this conformance root publishes no protected cache-hit vector")
	}
	if testCase.Result != "cache-hit" || testCase.ArtifactExecuted ||
		!testCase.ProtectedBoundaryChecked || len(testCase.SourceAwareGoCommands) != 0 {
		t.Fatalf("vector = %+v", testCase)
	}

	store := newTestStore(t)
	publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	published, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != Published {
		t.Fatalf("publication status = %q", published.Status)
	}
	result := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash, Assurance: publication.Assurance})
	if result.Status != Hit {
		t.Fatalf("exact protected entry = %+v", result)
	}
	if result.DryRunOutcome() != "cache-hit" || result.DryRunOutcome() != testCase.Result {
		t.Fatalf("dry-run outcome = %q, the suite publishes %q", result.DryRunOutcome(), testCase.Result)
	}
	if result.ReceiptHash != receiptHash {
		t.Fatalf("receipt identity = %q, want %q", result.ReceiptHash, receiptHash)
	}
	if result.ArtifactPath == "" {
		t.Fatal("a hit must name the protected artifact for the caller to install")
	}
	// The reusable artifact is returned as a path only. Nothing in this package
	// starts it, and the entry keeps its published bytes.
	info, err := os.Lstat(result.ArtifactPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("reusable artifact mode = %s", info.Mode())
	}
}

// TestCompilerFreeDryRunMissVector proves a cold protected cache reports the
// published plan outcome without any source-aware Go command and without
// creating persistent state.
func TestCompilerFreeDryRunMissVector(t *testing.T) {
	testCase, ok := loadCachePositiveCases(t)["compiler-free-dry-run-miss"]
	if !ok {
		t.Skip("this conformance root publishes no compiler-free dry-run vector")
	}
	if len(testCase.SourceAwareGoCommands) != 0 || len(testCase.PersistentEffects) != 0 {
		t.Fatalf("vector = %+v", testCase)
	}
	want := []string{"telemetry-off", "version", "env"}
	if len(testCase.PackageIndependentGoCmds) != len(want) {
		t.Fatalf("vector publishes %d package-independent commands", len(testCase.PackageIndependentGoCmds))
	}
	for index, name := range want {
		if testCase.PackageIndependentGoCmds[index] != name {
			t.Fatalf("package-independent command %d = %q, want %q", index, testCase.PackageIndependentGoCmds[index], name)
		}
	}

	store := newTestStore(t)
	input := testInput("tool")
	before := storeMembers(t, store.Home())
	result := store.Inspect(Expectation{Input: input, Assurance: closureexec.PortableAssuranceBinding()})
	if result.Status != Miss {
		t.Fatalf("cold protected cache = %+v", result)
	}
	if result.DryRunOutcome() != testCase.Result {
		t.Fatalf("dry-run outcome = %q, the suite publishes %q", result.DryRunOutcome(), testCase.Result)
	}
	if result.ArtifactPath != "" {
		t.Fatalf("a miss named an artifact: %s", result.ArtifactPath)
	}
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	if string(key) == "" {
		t.Fatal("a dry-run miss must still derive a cache key")
	}
	if after := storeMembers(t, store.Home()); after != before {
		t.Fatalf("a read-only inspection created persistent state: %d members before, %d after", before, after)
	}
}

// storeMembers counts every member below the manager home.
func storeMembers(t *testing.T, home string) int {
	t.Helper()
	var count int
	err := filepath.WalkDir(home, func(_ string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return count
}
