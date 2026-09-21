package crossconformance

import (
	"sync"
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
	resetSemanticOutcomes()
	for _, c := range cases {
		c := c
		t.Run(c.ID, func(t *testing.T) {
			// The cleanup observes the final subtest state even
			// when the driver stops early (t.Skip/t.FailNow), so
			// every iterated case is classified exactly once.
			t.Cleanup(func() {
				if t.Skipped() {
					recordSemanticSkipped(c.ID)
					return
				}
				recordSemanticOutcomeDefault(c.ID, semanticOutcomeDriven)
			})
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
	driven, knownGap, bound, skipped := tallySemanticOutcomes()
	t.Logf("semantic cases: %d driven, %d known-gap, %d bound, %d skipped, %d total",
		driven, knownGap, bound, skipped, len(cases))
	if executed := driven + knownGap + bound + skipped; executed != len(cases) || len(cases) != wantSemanticCases {
		t.Fatalf("executed(%d) != total(%d, want %d); run the full matrix without -run subtest filters",
			executed, len(cases), wantSemanticCases)
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
	recordSemanticOutcome(c.ID, semanticOutcomeBound)
	t.Logf("BOUND %s: expected %q: %s", c.ID, c.Expected, reason)
}

// Semantic outcome classes, one recorded per executed case id.
const (
	semanticOutcomeDriven   = "driven"
	semanticOutcomeKnownGap = "known-gap"
	semanticOutcomeBound    = "bound"
	semanticOutcomeSkipped  = "skipped"
)

// draftSemanticOutcomes classifies every case the runner iterates. The
// subtests run sequentially, but the registry is still mutex-guarded so
// a future t.Parallel cannot corrupt the tally.
var draftSemanticOutcomes = struct {
	sync.Mutex
	byID map[string]string
}{byID: map[string]string{}}

func resetSemanticOutcomes() {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	draftSemanticOutcomes.byID = map[string]string{}
}

// recordSemanticOutcome classifies one case explicitly (bound or
// known-gap). A conflicting second classification panics: the harness
// must never report one row under two classes.
func recordSemanticOutcome(id, class string) {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	if prev, ok := draftSemanticOutcomes.byID[id]; ok && prev != class {
		panic("semantic case " + id + " classified twice: " + prev + " then " + class)
	}
	draftSemanticOutcomes.byID[id] = class
}

// recordSemanticOutcomeDefault classifies one case only when no
// explicit class was recorded: a completed row without a bound or gap
// marker is a driven pass.
func recordSemanticOutcomeDefault(id, class string) {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	if _, ok := draftSemanticOutcomes.byID[id]; !ok {
		draftSemanticOutcomes.byID[id] = class
	}
}

// recordSemanticSkipped reports the row as skipped. Skip is the final
// Go-reported state, so it wins over an earlier explicit marker when a
// row records one and then skips at a later layer.
func recordSemanticSkipped(id string) {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	draftSemanticOutcomes.byID[id] = semanticOutcomeSkipped
}

func tallySemanticOutcomes() (driven, knownGap, bound, skipped int) {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	for _, class := range draftSemanticOutcomes.byID {
		switch class {
		case semanticOutcomeDriven:
			driven++
		case semanticOutcomeKnownGap:
			knownGap++
		case semanticOutcomeBound:
			bound++
		case semanticOutcomeSkipped:
			skipped++
		}
	}
	return driven, knownGap, bound, skipped
}
