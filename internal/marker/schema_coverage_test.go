package marker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

type markerSchemaCase struct {
	Name  string
	Valid bool
	Bytes []byte
}

func loadMarkerSchemaCases(t *testing.T, family string) []markerSchemaCase {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	casesDir := filepath.Join(root, "schema-cases", family)
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatal(err)
	}
	var cases []markerSchemaCase
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(casesDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, markerSchemaCase{
			Name:  entry.Name(),
			Valid: strings.HasPrefix(entry.Name(), "valid"),
			Bytes: payload,
		})
	}
	if len(cases) == 0 {
		t.Fatalf("the root publishes schema-cases/%s but it contains no cases", family)
	}
	return cases
}

func runMarkerSchemaCases(t *testing.T, family string) {
	t.Helper()
	cases := loadMarkerSchemaCases(t, family)
	conformancecoverage.RunOutcomes(t, "marker/"+family+"/schema-cases", cases,
		func(tc markerSchemaCase) string { return tc.Name }, func(t *testing.T, tc markerSchemaCase) conformancecoverage.Observation {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, Name), tc.Bytes, 0o644); err != nil {
				t.Fatal(err)
			}
			gotValid := Read(dir) != nil
			if gotValid != tc.Valid {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("Read valid=%v, suite expects %v", gotValid, tc.Valid)}
			}
			if tc.Name == "invalid-hash-version-on-frozen-marker.json" && !tc.Valid {
				controlPayload := withoutHashVersion(t, tc.Bytes)
				controlDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(controlDir, Name), controlPayload, 0o600); err != nil {
					t.Fatal(err)
				}
				if Read(controlDir) == nil {
					return conformancecoverage.Observation{FailureReason: "the frozen marker is not accepted after removing its unknown hash_version"}
				}
			}
			return conformancecoverage.Observation{}
		})
}

func withoutHashVersion(t *testing.T, payload []byte) []byte {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("decode frozen marker: %v", err)
	}
	if _, present := raw["hash_version"]; !present {
		t.Fatal("frozen marker case does not contain hash_version")
	}
	delete(raw, "hash_version")
	control, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("encode frozen marker control: %v", err)
	}
	return control
}

func TestReadAuthoritativeMarkerV2SchemaCases(t *testing.T) {
	runMarkerSchemaCases(t, "install-marker-v2")
}

func TestReadAuthoritativeMarkerV4SchemaCases(t *testing.T) {
	runMarkerSchemaCases(t, "install-marker-v4")
}

func TestReadAuthoritativeMarkerV5SchemaCases(t *testing.T) {
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
	runMarkerSchemaCases(t, "install-marker-v5")
}
