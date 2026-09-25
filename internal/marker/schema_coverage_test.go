package marker

import (
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
			return conformancecoverage.Observation{}
		})
}

func TestReadAuthoritativeMarkerV2SchemaCases(t *testing.T) {
	runMarkerSchemaCases(t, "install-marker-v2")
}

func TestReadAuthoritativeMarkerV4SchemaCases(t *testing.T) {
	runMarkerSchemaCases(t, "install-marker-v4")
}
