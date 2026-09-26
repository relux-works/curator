package crossconformance

import (
	"hash/crc32"
	"strconv"
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

// The released semantic matrix is divided into five stable batches so each
// bounded go test invocation stays below the local runner budget. CRC32 keeps
// the assignment deterministic and spreads unrelated production paths across
// batches.
const skillfileSourcesSemanticBatchCount = 5

func TestDraftSourcesSemanticCasesBatch0(t *testing.T) { runDraftSourcesSemanticBatch(t, 0) }
func TestDraftSourcesSemanticCasesBatch1(t *testing.T) { runDraftSourcesSemanticBatch(t, 1) }
func TestDraftSourcesSemanticCasesBatch2(t *testing.T) { runDraftSourcesSemanticBatch(t, 2) }
func TestDraftSourcesSemanticCasesBatch3(t *testing.T) { runDraftSourcesSemanticBatch(t, 3) }
func TestDraftSourcesSemanticCasesBatch4(t *testing.T) { runDraftSourcesSemanticBatch(t, 4) }

func draftSemanticBatch(id string) int {
	return int(crc32.ChecksumIEEE([]byte(id)) % skillfileSourcesSemanticBatchCount)
}

func runDraftSourcesSemanticBatch(t *testing.T, batch int) {
	cases := loadDraftSemantic(t)
	selected := make([]draftSemanticCase, 0, len(cases)/skillfileSourcesSemanticBatchCount+1)
	for _, c := range cases {
		if draftSemanticBatch(c.ID) == batch {
			selected = append(selected, c)
		}
	}
	family := "skillfile-sources-v1/semantic-cases/" + strconv.Itoa(batch)
	conformancecoverage.RunOutcomesParallel(t, family, selected,
		func(c draftSemanticCase) string { return c.ID }, func(t *testing.T, c draftSemanticCase) conformancecoverage.Observation {
			drive, ok := draftSemanticDrivers[c.ID]
			if !ok {
				t.Fatalf("semantic case %q has no production-entry row", c.ID)
			}
			if c.Expected == "" {
				t.Fatalf("case %q carries no expected outcome", c.ID)
			}
			drive(t, c)
			return conformancecoverage.Observation{}
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
		if batch := draftSemanticBatch(c.ID); batch < 0 || batch >= skillfileSourcesSemanticBatchCount {
			t.Errorf("semantic case %q has invalid batch %d", c.ID, batch)
		}
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
