package harnessboundary

import "testing"

func TestMuseBranchNegativeControl(t *testing.T) {
	root := t.TempDir()
	rel := "internal/shared/plant.go"
	plant(t, root, rel, "package shared\nfunc selected(id string) bool { return id == \"muse\" }\n")
	findings := scan(t, root)
	requireFinding(t, findings, rel, 2)
	unmatched := ApplyExceptions(findings, ExceptionSet{})
	if len(unmatched) != 1 || unmatched[0].HarnessID != "muse" {
		t.Fatalf("Muse branch was not rejected without an allowance: %v", unmatched)
	}
}
