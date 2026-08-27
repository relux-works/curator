package whitelist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/skillspec"
)

func TestBuildRootContextExclusionConformanceFixture(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	fixture := filepath.Join(root, "fixtures", "go-build-skill")
	spec, err := skillspec.Load(fixture)
	if err != nil {
		t.Fatal(err)
	}
	wantPayload, err := os.ReadFile(filepath.Join(root, "expected", "build-driver", "context_files.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wantFiles []string
	if err := json.Unmarshal(wantPayload, &wantFiles); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "context")
	excludeRoots := ContextExcludedRoots(spec.RuntimeRoots, spec.BuildRoots)
	files, err := CopyContext(fixture, destination, len(spec.Commands) == 0, excludeRoots)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, wantFiles) {
		t.Fatalf("context files = %v, want %v", files, wantFiles)
	}
	gotHash, err := hashing.ContentSHA256(destination, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantHashPayload, err := os.ReadFile(filepath.Join(root, "expected", "build-driver", "context_sha256.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if wantHash := strings.TrimSpace(string(wantHashPayload)); gotHash != wantHash {
		t.Fatalf("context hash = %s, want %s", gotHash, wantHash)
	}
}
