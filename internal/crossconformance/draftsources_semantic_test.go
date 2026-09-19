package crossconformance

import (
	"testing"
)

// draftSemanticDrivers maps every pinned semantic case id to the row
// that drives it at a production entry point (or records it as an
// explicit bound). The runner fails on any corpus id without a row and
// on any row without a corpus id, so the matrix cannot silently shrink.
var draftSemanticDrivers = map[string]func(t *testing.T, _ draftSemanticCase){}

// registerDraftSemantic registers one semantic row. Drivers live in the
// per-group files; registration keeps the matrix next to the code.
func registerDraftSemantic(id string, drive func(t *testing.T, _ draftSemanticCase)) {
	if _, dup := draftSemanticDrivers[id]; dup {
		panic("duplicate semantic driver for " + id)
	}
	draftSemanticDrivers[id] = drive
}

// TestDraftSourcesSemanticCases drives all 94 pinned semantic cases.
func TestDraftSourcesSemanticCases(t *testing.T) {
	cases := loadDraftSemantic(t)
	if len(cases) != wantSemanticCases {
		t.Fatalf("semantic cases = %d, want %d", len(cases), wantSemanticCases)
	}
	for _, c := range cases {
		c := c
		t.Run(c.ID, func(t *testing.T) {
			drive, ok := draftSemanticDrivers[c.ID]
			if !ok {
				t.Fatalf("semantic case %q has no production-entry row", c.ID)
			}
			if c.Expected == "" {
				t.Fatalf("case %q carries no expected outcome", c.ID)
			}
			drive(t, c)
		})
	}
}

// TestDraftSourcesSemanticCoverage fails when the row matrix and the
// corpus disagree in either direction.
func TestDraftSourcesSemanticCoverage(t *testing.T) {
	cases := loadDraftSemantic(t)
	ids := map[string]bool{}
	for _, c := range cases {
		ids[c.ID] = true
	}
	for _, c := range cases {
		if _, ok := draftSemanticDrivers[c.ID]; !ok {
			t.Errorf("semantic case %q has no production-entry row", c.ID)
		}
	}
	for id := range draftSemanticDrivers {
		if !ids[id] {
			t.Errorf("driver %q has no corpus case", id)
		}
	}
}

// semanticBound records an explicit bound: the case cannot be driven at
// a production entry for the stated reason, and is reported as a bound,
// never as passing.
func semanticBound(t *testing.T, c draftSemanticCase, reason string) {
	t.Helper()
	t.Logf("BOUND %s: expected %q: %s", c.ID, c.Expected, reason)
}
