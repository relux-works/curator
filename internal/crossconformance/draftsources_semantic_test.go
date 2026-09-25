package crossconformance

import (
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
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

// TestDraftSourcesSemanticCases drives every pinned semantic case.
func TestDraftSourcesSemanticCases(t *testing.T) {
	cases := loadDraftSemantic(t)
	resetSemanticOutcomes()
	conformancecoverage.RunOutcomesParallel(t, "draft-sources-v1/semantic-cases", cases,
		func(c draftSemanticCase) string { return c.ID }, func(t *testing.T, c draftSemanticCase) conformancecoverage.Observation {
			drive, ok := draftSemanticDrivers[c.ID]
			if !ok {
				t.Fatalf("semantic case %q has no production-entry row", c.ID)
			}
			if c.Expected == "" {
				t.Fatalf("case %q carries no expected outcome", c.ID)
			}
			drive(t, c)
			return semanticObservation(c.ID)
		})
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
	recordSemanticBound(c.ID, reason)
	t.Logf("BOUND %s: expected %q: %s", c.ID, c.Expected, reason)
}

// draftSemanticBounds carries only explicit bounds. Known gaps are read from
// the committed ledger by conformancecoverage.Run; no driver may declare one.
var draftSemanticOutcomes = struct {
	sync.Mutex
	byID map[string]conformancecoverage.Observation
}{byID: map[string]conformancecoverage.Observation{}}

func resetSemanticOutcomes() {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	draftSemanticOutcomes.byID = map[string]conformancecoverage.Observation{}
}

func recordSemanticBound(id, reason string) {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	if prev, ok := draftSemanticOutcomes.byID[id]; ok && prev.BoundReason != reason {
		panic("semantic case " + id + " classified twice: " + prev.BoundReason + " then " + reason)
	}
	draftSemanticOutcomes.byID[id] = conformancecoverage.Observation{BoundReason: reason}
}

func semanticObservation(id string) conformancecoverage.Observation {
	draftSemanticOutcomes.Lock()
	defer draftSemanticOutcomes.Unlock()
	return draftSemanticOutcomes.byID[id]
}
