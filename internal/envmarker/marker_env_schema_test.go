package envmarker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// schemaIndexEntry is one row of the published schema-cases index.
type schemaIndexEntry struct {
	Instance string `json:"instance"`
	Schema   string `json:"schema"`
	Valid    bool   `json:"valid"`
}

// schemaCaseValidity loads the published index and reports the recorded
// validity of each instance in the family. The index — not the file-name
// prefix — decides: an unindexed file fails the test instead of silently
// passing or skipping.
func schemaCaseValidity(t *testing.T, root, family string) map[string]bool {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var entries []schemaIndexEntry
	if err := json.Unmarshal(payload, &entries); err != nil {
		t.Fatal(err)
	}
	validity := map[string]bool{}
	indexed := map[string]bool{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Instance, family+"/") {
			continue
		}
		name := strings.TrimPrefix(entry.Instance, family+"/")
		validity[name] = entry.Valid
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
	return validity
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
	validity := schemaCaseValidity(t, root, "agent-environment-marker-v1")
	for name, wantValid := range validity {
		t.Run(name, func(t *testing.T) {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "agent-environment-marker-v1", name)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			_, err = Parse(payload)
			if wantValid && err != nil {
				t.Fatalf("valid case rejected: %v", err)
			}
			if !wantValid && err == nil {
				t.Fatalf("invalid case accepted")
			}
		})
	}
}
