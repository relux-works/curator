package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

type frozenMarkerSchemaIndexEntry struct {
	Instance string `json:"instance"`
	Valid    bool   `json:"valid"`
}

func TestReadRejectsHashVersionOnFrozenMarkerV3(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	const family = "install-marker-v3"
	const caseName = "invalid-hash-version-on-frozen-marker.json"
	coverageFamily := "marker/" + family + "/schema-cases"

	indexBytes, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []frozenMarkerSchemaIndexEntry
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("parse schema-case index: %v", err)
	}
	var count int
	var target *frozenMarkerSchemaIndexEntry
	prefix := family + "/"
	for _, entry := range index {
		if !strings.HasPrefix(entry.Instance, prefix) {
			continue
		}
		count++
		if strings.TrimPrefix(entry.Instance, prefix) == caseName {
			copyEntry := entry
			target = &copyEntry
		}
	}
	conformancecoverage.RequirePublishedCount(t, coverageFamily, count)
	if target == nil {
		if _, err := os.Stat(filepath.Join(root, "schema-cases", family, caseName)); err == nil {
			t.Fatal("frozen marker case exists on disk but is absent from the published index")
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect frozen marker case: %v", err)
		}
		if suiteID, err := conformancecoverage.SelectedSuiteManifestSHA256(); err != nil {
			t.Fatal(err)
		} else if suiteID == conformancecoverage.ContentHashV2CandidateManifestSHA256 {
			t.Fatal("candidate suite is missing the frozen marker v3 hash-version case")
		}
		return
	}
	if target.Valid {
		t.Fatal("the frozen marker v3 hash-version case is published as valid")
	}
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", family, caseName))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, Name), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := Read(dir); got != nil {
		t.Fatalf("Read accepted the frozen marker with hash_version: %+v", got)
	}
	controlDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(controlDir, Name), withoutHashVersion(t, payload), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := Read(controlDir); got == nil {
		t.Fatal("frozen marker v3 is not accepted after removing its unknown hash_version")
	}
}
