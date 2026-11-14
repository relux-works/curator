package envprofile

import "testing"

// Expected IDs come independently from the pinned manifest, while executed
// IDs/levels are recorded only inside the actual runtime child invocations.
func storeVerificationRuntimeRoster(t *testing.T, m storeVerificationManifest, rows []storeVerificationCase, group string) map[string]map[int]bool {
	t.Helper()
	byName := map[string]storeVerificationCase{}
	for _, row := range rows {
		byName[row.Name] = row
	}
	expected := map[string]map[int]bool{}
	slots := 0
	for _, old := range m.Migration {
		if old.Group != group {
			continue
		}
		for _, id := range append([]string{old.Successor}, old.LevelSuccessors...) {
			row, ok := byName[id]
			if !ok || expected[id] != nil {
				t.Fatalf("runtime manifest ID missing or repeated: %s", id)
			}
			expected[id] = map[int]bool{}
			for _, level := range row.Levels {
				expected[id][level] = true
				slots++
			}
		}
	}
	counts := map[string][2]int{"resolve_cases": {31, 80}, "dry_run_cases": {4, 16}, "repair_cases": {16, 40}, "status_cases": {4, 16}}
	want, ok := counts[group]
	if !ok || len(expected) != want[0] || slots != want[1] {
		t.Fatalf("runtime manifest scope drift: %s rows=%d slots=%d", group, len(expected), slots)
	}
	return expected
}

// Dropping all but the first row must fail the exact runtime roster guard.
func TestStoreVerificationRuntimeRosterControl(t *testing.T) {
	t.Run("row", func(t *testing.T) {
		runObservationControl(t, "store_verification_runtime_test.go", `if name == "" {`, `if name == "" || count >= 1 {`, "TestStoreBoundaryStatusVectorsDriveStatus", "runtime roster mismatch")
	})
	t.Run("level", func(t *testing.T) {
		runObservationControl(t, "store_verification_runtime_test.go", "for _, level := range row.Levels {", "for _, level := range row.Levels[:1] {", "TestStoreBoundaryStatusVectorsDriveStatus", "runtime roster mismatch")
	})
}
