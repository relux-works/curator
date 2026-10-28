package hookapproval

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// Planning preserves unrelated operator approvals and rejects malformed state
// or additions before producing bytes. This pure codec test is not an executor
// rollback or live shell-hook witness.
func TestPlanUpserts(t *testing.T) {
	stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	prior := []byte("/fixture/operator\t" + strings.Repeat("a", 64) + "\toperator\t2026-01-01T00:00:00Z\n")
	original := append([]byte{}, prior...)
	addition := Record{Path: "/fixture/env.sh", SHA256: strings.Repeat("b", 64), ApprovedBy: ApprovedByManager, ApprovedAt: stamp}
	for _, bad := range []bool{false, true} {
		record := addition
		if bad {
			record.Path = "/fixture/env\t.sh"
		}
		got, err := PlanUpserts(prior, []Record{record})
		if bad {
			if err == nil || got != nil {
				t.Fatal("malformed addition produced authority")
			}
		} else if err != nil || !bytes.Contains(got, prior) || !bytes.Contains(got, []byte("/fixture/env.sh\t"+addition.SHA256+"\tmanager\t")) {
			t.Fatalf("planned records: %s %v", got, err)
		}
		if !bytes.Equal(prior, original) {
			t.Fatal("planning changed caller preimage")
		}
	}
	if got, err := PlanUpserts([]byte("malformed\n"), []Record{addition}); err == nil || got != nil {
		t.Fatal("malformed prior state was discarded")
	}
}
