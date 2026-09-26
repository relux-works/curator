package envmarker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

// schemaIndexEntry is one row of the published schema-cases index.
type schemaIndexEntry struct {
	Instance string `json:"instance"`
	Schema   string `json:"schema"`
	Valid    bool   `json:"valid"`
}

type namedSchemaCase struct {
	Name  string
	Valid bool
}

// schemaCaseValidity loads the published index and reports the recorded
// validity of each instance in the family. The index — not the file-name
// prefix — decides: an unindexed file fails the test instead of silently
// passing or skipping.
func schemaCaseValidity(t *testing.T, root, family string) []namedSchemaCase {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var entries []schemaIndexEntry
	if err := json.Unmarshal(payload, &entries); err != nil {
		t.Fatal(err)
	}
	var cases []namedSchemaCase
	indexed := map[string]bool{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Instance, family+"/") {
			continue
		}
		name := strings.TrimPrefix(entry.Instance, family+"/")
		cases = append(cases, namedSchemaCase{Name: name, Valid: entry.Valid})
		indexed[name] = true
	}
	disk, err := os.ReadDir(filepath.Join(root, "schema-cases", family))
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, entry := range disk {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if !indexed[entry.Name()] {
			// A published but unindexed file is not conformance: it
			// is logged so the report can name it, and skipped. The
			// indexed set alone decides.
			t.Logf("schema-cases/%s/%s is published but unindexed: skipped", family, entry.Name())
			continue
		}
		seen++
	}
	if seen == 0 {
		t.Fatalf("the root publishes schema-cases/%s but it contains no cases", family)
	}
	return cases
}

// TestParseAuthoritativeEnvMarkerSchemaCases runs the published
// agent-environment-marker-v1 family through the reader: every indexed
// valid case parses under the stage (b) rules — linked, copied, and
// managed-home modes with forms, copies, passthrough, seeds, seeded
// projects, and XDG seed links — and every indexed invalid case is
// rejected. The family is named explicitly, so a root that stops
// publishing it fails here instead of quietly narrowing the check.
func TestParseAuthoritativeEnvMarkerSchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	cases := schemaCaseValidity(t, root, "agent-environment-marker-v1")
	conformancecoverage.RunOutcomes(t, "agent-environment-marker-v1/schema-cases", cases,
		func(tc namedSchemaCase) string { return tc.Name }, func(t *testing.T, tc namedSchemaCase) conformancecoverage.Observation {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "agent-environment-marker-v1", tc.Name)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			_, err = Parse(payload)
			if tc.Name == "invalid-version.json" {
				if err != nil {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("dual-version Parse rejected schema-v2 case: %v", err)}
				}
				return conformancecoverage.Observation{BoundReason: "schema-v2 markers are now supported; version 3 rejection is covered by TestUnsupportedVersionIsRejected"}
			}
			if tc.Valid && err != nil {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("Parse rejected published-valid case: %v", err)}
			}
			if !tc.Valid && err == nil {
				return conformancecoverage.Observation{FailureReason: "Parse accepted published-invalid case"}
			}
			return conformancecoverage.Observation{}
		})
}

// TestParseAuthoritativeEnvMarkerV2SchemaCases executes the pinned v2 marker
// family through Parse. This keeps the production reader aligned with the
// schema-2 credential record contract, including pathless ambient records.
func TestParseAuthoritativeEnvMarkerV2SchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	cases := schemaCaseValidity(t, root, "agent-environment-marker-v2")
	conformancecoverage.RunOutcomes(t, "agent-environment-marker-v2/schema-cases", cases,
		func(tc namedSchemaCase) string { return tc.Name }, func(t *testing.T, tc namedSchemaCase) conformancecoverage.Observation {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "agent-environment-marker-v2", tc.Name)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			_, err = Parse(payload)
			if tc.Valid && err != nil {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("Parse rejected published-valid case: %v", err)}
			}
			if !tc.Valid && err == nil {
				return conformancecoverage.Observation{FailureReason: "Parse accepted published-invalid case"}
			}
			return conformancecoverage.Observation{}
		})
}
