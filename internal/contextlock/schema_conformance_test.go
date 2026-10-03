package contextlock

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

type frozenLockSchemaIndexEntry struct {
	Instance string `json:"instance"`
	Valid    bool   `json:"valid"`
}

func TestContextLockV2SchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	suiteID, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	if !conformancecoverage.IsImplementedContentHashV2Suite(suiteID) {
		return
	}
	indexBytes, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []frozenLockSchemaIndexEntry
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("parse schema-case index: %v", err)
	}
	const family = "context-lock-v2"
	prefix := family + "/"
	var cases []frozenLockSchemaIndexEntry
	for _, entry := range index {
		if strings.HasPrefix(entry.Instance, prefix) {
			cases = append(cases, entry)
		}
	}
	conformancecoverage.RequirePublishedCount(t, family+"/schema-cases", len(cases))
	conformancecoverage.RunOutcomes(t, family+"/schema-cases", cases,
		func(entry frozenLockSchemaIndexEntry) string { return strings.TrimPrefix(entry.Instance, prefix) },
		func(t *testing.T, entry frozenLockSchemaIndexEntry) conformancecoverage.Observation {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", filepath.FromSlash(entry.Instance)))
			if err != nil {
				t.Fatal(err)
			}
			lock, parseErr := Parse(payload)
			if (parseErr == nil) != entry.Valid {
				t.Fatalf("valid=%v: Parse err=%v", entry.Valid, parseErr)
			}
			if entry.Valid && (lock.SchemaVersion != SchemaVersion2 || lock.HashVersion != 2) {
				t.Fatalf("parsed versions = schema:%d hash:%d, want 2/2", lock.SchemaVersion, lock.HashVersion)
			}
			return conformancecoverage.Observation{}
		})
}

func TestParseRejectsHashVersionOnFrozenContextLock(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	const family = "context-lock-v1"
	const caseName = "invalid-hash-version-on-frozen-lock.json"

	indexBytes, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []frozenLockSchemaIndexEntry
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("parse schema-case index: %v", err)
	}
	var count int
	var target *frozenLockSchemaIndexEntry
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
	conformancecoverage.RequirePublishedCount(t, family+"/schema-cases", count)
	if target == nil {
		if _, err := os.Stat(filepath.Join(root, "schema-cases", family, caseName)); err == nil {
			t.Fatal("frozen context-lock case exists on disk but is absent from the published index")
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect frozen context-lock case: %v", err)
		}
		if suiteID, err := conformancecoverage.SelectedSuiteManifestSHA256(); err != nil {
			t.Fatal(err)
		} else if conformancecoverage.IsContentHashV2Candidate(suiteID) {
			t.Fatal("candidate suite is missing the frozen context-lock hash-version case")
		}
		return
	}
	if target.Valid {
		t.Fatal("the frozen context-lock hash-version case is published as valid")
	}
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", family, caseName))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(payload); err == nil {
		t.Fatal("Parse accepted the frozen context-lock with hash_version")
	}
	var control map[string]json.RawMessage
	if err := json.Unmarshal(payload, &control); err != nil {
		t.Fatalf("decode frozen context-lock: %v", err)
	}
	if _, present := control["hash_version"]; !present {
		t.Fatal("frozen context-lock case does not contain hash_version")
	}
	delete(control, "hash_version")
	controlPayload, err := json.Marshal(control)
	if err != nil {
		t.Fatalf("encode frozen context-lock control: %v", err)
	}
	if _, err := Parse(controlPayload); err != nil {
		t.Fatalf("context-lock without unknown hash_version is not a valid control: %v", err)
	}
}
