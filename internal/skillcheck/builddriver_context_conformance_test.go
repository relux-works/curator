package skillcheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

// TestBuildRootContentInContextVector proves the authoritative
// build-root-content-in-context rejection reaches a stable Curator issue code:
// prompt-visible text that names a build-root input is reported before any
// installation, and Curator never copies that input into agent context.
func TestBuildRootContentInContextVector(t *testing.T) {
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
		RejectionCases []struct {
			Name     string `json:"name"`
			Boundary string `json:"boundary"`
			Expected struct {
				Result           string `json:"result"`
				Error            string `json:"error"`
				Reuse            bool   `json:"reuse"`
				ArtifactExecuted bool   `json:"artifact_executed"`
			} `json:"expected"`
		} `json:"rejection_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, testCase := range vectors.RejectionCases {
		if testCase.Name != "build-root-content-in-context" {
			continue
		}
		if testCase.Boundary != "context" || testCase.Expected.Result != "reject" ||
			testCase.Expected.Reuse || testCase.Expected.ArtifactExecuted {
			t.Fatalf("vector no longer fails closed: %+v", testCase)
		}
		found = true
	}
	if !found {
		t.Skipf("%s publishes no build-root context-exposure rejection", root)
	}

	// The skill is materialised from the authoritative fixture layout, then one
	// prompt-visible document is made to reference the build root.
	dir := copyConformanceFixture(t, filepath.Join(root, "fixtures", "go-build-skill"))
	spec, err := skillspec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.BuildRoots) == 0 {
		t.Fatal("the fixture manifest declares no build roots")
	}
	buildRoot := spec.BuildRoots[0]

	clean := buildRootReferenceWarnings(dir, spec)
	for _, issue := range clean {
		if issue.Code == "skill.build_root_in_prompt_context" {
			t.Fatalf("the untouched fixture already exposes a build root: %s", issue.Message)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"),
		[]byte("# golden\n\nRun `"+buildRoot+"/cmd/golden-tool/main.go` directly.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var codes []string
	for _, issue := range buildRootReferenceWarnings(dir, spec) {
		codes = append(codes, issue.Code)
		if issue.Code == "skill.build_root_in_prompt_context" {
			if !strings.Contains(issue.Message, buildRoot) {
				t.Fatalf("issue does not name the build root: %s", issue.Message)
			}
			return
		}
	}
	t.Fatalf("build-root exposure produced %v, want skill.build_root_in_prompt_context", codes)
}

// copyConformanceFixture materialises a writable copy of a suite fixture.
func copyConformanceFixture(t *testing.T, source string) string {
	t.Helper()
	destination := t.TempDir()
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return destination
}
