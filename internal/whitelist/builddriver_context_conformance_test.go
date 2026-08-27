package whitelist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

// TestBuildRootExcludedFromAgentContextVector is the executable assertion for
// the authoritative positive context vector: the agent context carries exactly
// the published prompt-visible members and none of the build-root or runtime
// inputs, and no Go child is involved in producing it.
func TestBuildRootExcludedFromAgentContextVector(t *testing.T) {
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
		PositiveCases []struct {
			Name                    string   `json:"name"`
			Result                  string   `json:"result"`
			ExpectedContextFiles    []string `json:"expected_context_files"`
			ExpectedExcludedFiles   []string `json:"expected_excluded_files"`
			CacheHitGoCommands      []string `json:"cache_hit_source_aware_go_commands"`
			DryRunSourceAwareGoCmds []string `json:"dry_run_source_aware_go_commands"`
		} `json:"positive_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	var wantContext, wantExcluded []string
	var found bool
	for _, testCase := range vectors.PositiveCases {
		if testCase.Name != "build-root-excluded-from-agent-context" {
			continue
		}
		if testCase.Result != "accepted" {
			t.Fatalf("vector result = %q, want accepted", testCase.Result)
		}
		if len(testCase.CacheHitGoCommands) != 0 || len(testCase.DryRunSourceAwareGoCmds) != 0 {
			t.Fatalf("vector allows source-aware Go commands: %+v", testCase)
		}
		wantContext, wantExcluded, found = testCase.ExpectedContextFiles, testCase.ExpectedExcludedFiles, true
	}
	if !found {
		t.Skipf("%s publishes no build-root context-exclusion vector", root)
	}

	fixture := filepath.Join(root, "fixtures", "go-build-skill")
	spec, err := skillspec.Load(fixture)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "context")
	files, err := CopyContext(fixture, destination,
		len(spec.Commands) == 0, ContextExcludedRoots(spec.RuntimeRoots, spec.BuildRoots))
	if err != nil {
		t.Fatal(err)
	}
	got := append([]string(nil), files...)
	sort.Strings(got)
	want := append([]string(nil), wantContext...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("agent context = %v, the vector publishes %v", got, want)
	}
	for _, excluded := range wantExcluded {
		if _, err := os.Lstat(filepath.Join(destination, filepath.FromSlash(excluded))); !os.IsNotExist(err) {
			t.Fatalf("excluded input %q reached the agent context (%v)", excluded, err)
		}
	}
}
