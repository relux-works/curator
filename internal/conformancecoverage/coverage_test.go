package conformancecoverage

import (
	"strings"
	"testing"
)

func TestPublishedCaseTallyRejectsMissingClassification(t *testing.T) {
	_, err := Check("fixture/family", []string{"case-a", "case-b"}, []Result{
		{CaseID: "case-a"},
	}, nil, 2)
	if err == nil || !strings.Contains(err.Error(), "want published count 2") {
		t.Fatalf("Check error = %v, want tally mismatch for omitted case-b", err)
	}
}

func TestUnlistedFailingCaseIsRejected(t *testing.T) {
	_, err := Check("fixture/family", []string{"case-a"}, []Result{
		{CaseID: "case-a", FailureReason: "production entry rejected the published case"},
	}, nil, 1)
	if err == nil || !strings.Contains(err.Error(), "not listed in the gap ledger") {
		t.Fatalf("Check error = %v, want unlisted failing case refusal", err)
	}
}

func TestPassingKnownGapIsRejected(t *testing.T) {
	_, err := Check("fixture/family", []string{"case-a"}, []Result{
		{CaseID: "case-a"},
	}, []Gap{{Family: "fixture/family", CaseID: "case-a", Owner: "TASK-1", Reason: "owed"}}, 1)
	if err == nil || !strings.Contains(err.Error(), "now passes; remove its ledger row") {
		t.Fatalf("Check error = %v, want stale passing gap refusal", err)
	}
}

func TestVanishedGapCaseIsRejected(t *testing.T) {
	_, err := Check("fixture/family", []string{"case-b"}, []Result{
		{CaseID: "case-b"},
	}, []Gap{{Family: "fixture/family", CaseID: "case-a", Owner: "TASK-1", Reason: "owed"}}, 1)
	if err == nil || !strings.Contains(err.Error(), "case-a vanished") {
		t.Fatalf("Check error = %v, want vanished ledger case refusal", err)
	}
}

func TestPinnedCountRejectsVanishedPublishedCase(t *testing.T) {
	_, err := Check("fixture/family", []string{"case-a"}, []Result{
		{CaseID: "case-a"},
	}, nil, 2)
	if err == nil || !strings.Contains(err.Error(), "want pinned count 2") {
		t.Fatalf("Check error = %v, want vanished published case refusal", err)
	}
}

func TestCoverageCountsEveryOutcomeClassOnce(t *testing.T) {
	ids := []string{"driven", "gap", "bound", "skip"}
	results := []Result{
		{CaseID: "driven"},
		{CaseID: "gap", FailureReason: "still fails"},
		{CaseID: "bound", BoundReason: "no production entry exists"},
		{CaseID: "skip", SkipReason: "platform does not expose this property"},
	}
	gaps := []Gap{{Family: "fixture/family", CaseID: "gap", Owner: "TASK-1", Reason: "owed"}}
	tally, err := Check("fixture/family", ids, results, gaps, len(ids))
	if err != nil {
		t.Fatal(err)
	}
	want := (Tally{Driven: 1, KnownGap: 1, Bound: 1, Skipped: 1})
	if tally != want {
		t.Fatalf("tally = %+v, want %+v", tally, want)
	}
}

func TestCommittedCoveragePolicyParses(t *testing.T) {
	counts, gaps, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := counts["manager-config-v2/vectors"]; !ok {
		t.Fatal("manager-config-v2 vector count pin is missing")
	}
	for _, gap := range gaps {
		if counts[gap.Family] == 0 || gap.CaseID == "" || gap.Owner == "" || gap.Reason == "" {
			t.Fatalf("malformed committed gap row: %+v", gap)
		}
	}
}
