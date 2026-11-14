package envprofile

import (
	"encoding/json"
	"strings"
	"testing"
)

// Decode-level controls reach all four migration gates beyond the immutable
// pin seam. Valid disjoint successors cover all levels; each defect has its
// named refusal and never reaches a runtime fixture.
func TestStoreVerificationMigrationCoverage(t *testing.T) {
	const raw = `{"migration":[{"group":"resolve_cases","name":"old","successor_case":"legacy","verification_levels":[3,4],"level_successors":["low1","low2"]}]}`
	for _, plant := range []string{"valid", "scope", "invalid", "overlap", "incomplete"} {
		t.Run(plant, func(t *testing.T) {
			var m storeVerificationManifest
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				t.Fatal(err)
			}
			base := storeVerificationCase{Name: "legacy", Operation: "resolve", Member: "context", Source: "git", Home: "current", Levels: []int{3, 4}, Predecessor: &storeVerificationPredecessor{Group: "resolve_cases", Name: "old"}}
			one, two := base, base
			one.Name, one.Levels, one.Predecessor = "low1", []int{1}, nil
			two.Name, two.Levels, two.Predecessor = "low2", []int{2}, nil
			want := ""
			switch plant {
			case "scope":
				m.Migration[0].Levels = []int{3}
				want = "predecessor level scope mismatch"
			case "invalid":
				one.Operation = "repair"
				want = "invalid level successor"
			case "overlap":
				one.Levels = []int{3}
				want = "overlapping level successor"
			case "incomplete":
				m.Migration[0].LevelSuccessors = []string{"low1"}
				want = "incomplete predecessor level coverage"
			}
			err := validateStoreVerificationMigration(m, []storeVerificationCase{base, one, two})
			if want == "" {
				if err != nil {
					t.Fatalf("valid migration neighbor refused: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("migration gate accepted %s: %v", plant, err)
			}
		})
	}
}

// Disabling incomplete-level admission must fail its intended assertion.
func TestStoreVerificationMigrationCoverageControl(t *testing.T) {
	runObservationControl(t, "store_verification_reader_test.go", "if len(levels) != 4 {", "if false && len(levels) != 4 {", "TestStoreVerificationMigrationCoverage", "migration gate accepted incomplete")
}
