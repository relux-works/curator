package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/hashing"
)

type frozenRegistrySchemaIndexEntry struct {
	Instance string `json:"instance"`
	Valid    bool   `json:"valid"`
}

func TestAuditRecordV2SchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	suiteID, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	if suiteID != conformancecoverage.ContentHashV2CandidateManifestSHA256 {
		return
	}
	indexBytes, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []frozenRegistrySchemaIndexEntry
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("parse schema-case index: %v", err)
	}
	const family = "audit-record-v2"
	prefix := family + "/"
	var cases []frozenRegistrySchemaIndexEntry
	for _, entry := range index {
		if strings.HasPrefix(entry.Instance, prefix) {
			cases = append(cases, entry)
		}
	}
	conformancecoverage.RequirePublishedCount(t, family+"/schema-cases", len(cases))
	conformancecoverage.RunOutcomes(t, family+"/schema-cases", cases,
		func(entry frozenRegistrySchemaIndexEntry) string { return strings.TrimPrefix(entry.Instance, prefix) },
		func(t *testing.T, entry frozenRegistrySchemaIndexEntry) conformancecoverage.Observation {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", filepath.FromSlash(entry.Instance)))
			if err != nil {
				t.Fatal(err)
			}
			var object map[string]any
			if err := decodeJSON(payload, &object); err != nil {
				t.Fatalf("decode audit-record-v2 case %s: %v", entry.Instance, err)
			}
			_, parseErr := ParseRecordVersioned(object)
			if (parseErr == nil) != entry.Valid {
				t.Fatalf("valid=%v: ParseRecord err=%v", entry.Valid, parseErr)
			}
			return conformancecoverage.Observation{}
		})
}

func TestRegistryV2CarrierSchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	suiteID, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	if suiteID != conformancecoverage.ContentHashV2CandidateManifestSHA256 {
		return
	}
	indexBytes, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []frozenRegistrySchemaIndexEntry
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("parse schema-case index: %v", err)
	}
	families := []struct {
		name  string
		parse func(map[string]any) error
	}{
		{name: "registry-log-entry-v2", parse: func(value map[string]any) error { _, err := ParseLogEntryVersioned(value); return err }},
		{name: "registry-bundle-v2", parse: func(value map[string]any) error { _, err := ParseBundleV2(value); return err }},
		{name: "log-response-v3", parse: func(value map[string]any) error { _, err := ParseLogResponseV3(value); return err }},
	}
	for _, family := range families {
		family := family
		prefix := family.name + "/"
		var cases []frozenRegistrySchemaIndexEntry
		for _, entry := range index {
			if strings.HasPrefix(entry.Instance, prefix) {
				cases = append(cases, entry)
			}
		}
		coverageID := family.name + "/schema-cases"
		conformancecoverage.RequirePublishedCount(t, coverageID, len(cases))
		conformancecoverage.RunOutcomes(t, coverageID, cases,
			func(entry frozenRegistrySchemaIndexEntry) string { return strings.TrimPrefix(entry.Instance, prefix) },
			func(t *testing.T, entry frozenRegistrySchemaIndexEntry) conformancecoverage.Observation {
				payload, err := os.ReadFile(filepath.Join(root, "schema-cases", filepath.FromSlash(entry.Instance)))
				if err != nil {
					t.Fatal(err)
				}
				var object map[string]any
				if err := decodeJSON(payload, &object); err != nil {
					t.Fatalf("decode %s case %s: %v", family.name, entry.Instance, err)
				}
				parseErr := family.parse(object)
				if (parseErr == nil) != entry.Valid {
					t.Fatalf("valid=%v: parse err=%v", entry.Valid, parseErr)
				}
				return conformancecoverage.Observation{}
			})
	}
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
		} else if conformancecoverage.IsContentHashV2Candidate(suiteID) {
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

// TestParseLogEntryRejectsV2RecordInsideFrozenRegistryLogEntry drives the
// same frozen case through the entry readers: the frozen v1 entry reader
// refuses the nested v2 record, the versioned reader admits it, and the same
// entry with a v1 record is a valid frozen control.
func TestParseLogEntryRejectsV2RecordInsideFrozenRegistryLogEntry(t *testing.T) {
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
	if _, err := ParseLogEntry(entry); err == nil {
		t.Fatal("ParseLogEntry accepted a v2 record inside a frozen registry-log-entry-v1 case")
	}
	if parsed, err := ParseLogEntryVersioned(entry); err != nil || parsed.Record.HashVersion != hashing.VersionV2 {
		t.Fatalf("versioned entry reader = %+v, %v; want the nested hash-v2 record", parsed, err)
	}
	v1Record := make(map[string]any, len(record)-1)
	for key, value := range record {
		if key != "hash_version" {
			v1Record[key] = value
		}
	}
	v1Record["schema_version"] = json.Number("1")
	control := map[string]any{"seq": entry["seq"], "entry_hash": entry["entry_hash"], "prev_hash": entry["prev_hash"], "record": v1Record}
	if parsed, err := ParseLogEntry(control); err != nil || parsed.Record.HashVersion != hashing.VersionV1 {
		t.Fatalf("frozen entry reader rejected the v1-record control: %+v, %v", parsed, err)
	}
}
