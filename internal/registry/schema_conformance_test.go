package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

type frozenRegistrySchemaIndexEntry struct {
	Instance string `json:"instance"`
	Valid    bool   `json:"valid"`
}

func readFrozenRegistrySchemaCase(t *testing.T, root, family, caseName string) ([]byte, bool) {
	t.Helper()
	indexBytes, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []frozenRegistrySchemaIndexEntry
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("parse schema-case index: %v", err)
	}
	var count int
	var target *frozenRegistrySchemaIndexEntry
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
	path := filepath.Join(root, "schema-cases", family, caseName)
	if target == nil {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("frozen case exists on disk but is absent from the published index: %s", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect frozen case %s: %v", path, err)
		}
		if suiteID, err := conformancecoverage.SelectedSuiteManifestSHA256(); err != nil {
			t.Fatal(err)
		} else if suiteID == conformancecoverage.ContentHashV2CandidateManifestSHA256 {
			t.Fatalf("candidate suite is missing frozen case %s/%s", family, caseName)
		}
		return nil, false
	}
	if target.Valid {
		t.Fatalf("frozen case %s/%s is published as valid", family, caseName)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return payload, true
}

func TestParseRecordRejectsHashVersionOnFrozenAuditRecord(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, present := readFrozenRegistrySchemaCase(t, root, "audit-record-v1", "invalid-hash-version-on-frozen-record.json")
	if !present {
		return
	}
	var record map[string]any
	if err := decodeJSON(payload, &record); err != nil {
		t.Fatal(err)
	}
	if record["schema_version"] != json.Number("1") || record["hash_version"] != json.Number("2") {
		t.Fatalf("frozen audit record versions = schema:%v hash:%v", record["schema_version"], record["hash_version"])
	}
	mutant := make(map[string]any, len(record)-1)
	for key, value := range record {
		if key != "hash_version" {
			mutant[key] = value
		}
	}
	if _, err := ParseRecord(mutant); err != nil {
		t.Fatalf("audit record without the unknown hash_version is not a valid control: %v", err)
	}
	if _, err := ParseRecord(record); err == nil {
		t.Fatal("ParseRecord accepted hash_version on a frozen audit-record-v1 document")
	}
}

func TestParseRecordRejectsV2RecordInsideFrozenRegistryLogEntry(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, present := readFrozenRegistrySchemaCase(t, root, "registry-log-entry-v1", "invalid-v2-record-in-frozen-entry.json")
	if !present {
		return
	}
	var entry map[string]any
	if err := decodeJSON(payload, &entry); err != nil {
		t.Fatal(err)
	}
	record, ok := entry["record"].(map[string]any)
	if !ok {
		t.Fatalf("frozen registry log entry has no record object: %T", entry["record"])
	}
	if record["schema_version"] != json.Number("2") || record["hash_version"] != json.Number("2") {
		t.Fatalf("embedded record versions = schema:%v hash:%v", record["schema_version"], record["hash_version"])
	}
	if _, err := ParseRecord(record); err == nil {
		t.Fatal("ParseRecord accepted a v2 record inside a frozen registry-log-entry-v1 case")
	}
}
