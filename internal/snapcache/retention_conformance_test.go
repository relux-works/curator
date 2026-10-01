package snapcache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

type retentionVector struct {
	Capability string          `json:"capability"`
	Reasons    []string        `json:"reasons"`
	Cases      []retentionCase `json:"cases"`
}

type retentionCase struct {
	Name                string `json:"name"`
	Now                 string `json:"now"`
	GraceSeconds        int64  `json:"grace_seconds"`
	ReferenceSetCertain bool   `json:"reference_set_certain"`
	Policy              struct {
		KeepLast         *int   `json:"keep_last"`
		OlderThanSeconds *int64 `json:"older_than_seconds"`
	} `json:"policy"`
	Entries []struct {
		Source     string `json:"source"`
		Commit     string `json:"commit"`
		LastUsedAt string `json:"last_used_at"`
		Reachable  bool   `json:"reachable"`
	} `json:"entries"`
	Expected []retentionOutcome `json:"expected"`
}

type retentionOutcome struct {
	Source string `json:"source"`
	Commit string `json:"commit"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// TestSnapshotRetentionDecisionCases drives every published
// snapshot-retention decision case (manager profile section 10.1) through
// Plan, the function `curator cache prune` applies. Suites that predate the
// family publish no vector; the snapshot-retention candidate must publish it.
func TestSnapshotRetentionDecisionCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	suiteID, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "vectors", "snapshot-retention.json"))
	if os.IsNotExist(err) {
		if suiteID == conformancecoverage.SnapshotRetentionCandidateManifestSHA256 {
			t.Fatal("snapshot-retention candidate suite lost vectors/snapshot-retention.json")
		}
		return
	}
	if err != nil {
		t.Fatalf("read snapshot-retention vector: %v", err)
	}
	var vector retentionVector
	if err := json.Unmarshal(data, &vector); err != nil {
		t.Fatalf("parse snapshot-retention vector: %v", err)
	}
	if vector.Capability != "snapshot-retention" {
		t.Fatalf("capability = %q", vector.Capability)
	}
	want := []string{ReasonReachable, ReasonReferenceUncertain, ReasonGrace, ReasonKeepLast, ReasonNewerThan, ReasonUnreachable}
	if len(vector.Reasons) != len(want) {
		t.Fatalf("reasons = %q, Curator decides %q", vector.Reasons, want)
	}
	for index := range want {
		if vector.Reasons[index] != want[index] {
			t.Fatalf("reasons = %q, Curator decides %q", vector.Reasons, want)
		}
	}
	conformancecoverage.Run(t, "snapshot-retention/decision-cases", vector.Cases,
		func(testCase retentionCase) string { return testCase.Name },
		func(t *testing.T, testCase retentionCase) {
			now := mustTime(t, testCase.Now)
			policy := Policy{KeepLast: testCase.Policy.KeepLast, Grace: time.Duration(testCase.GraceSeconds) * time.Second}
			if testCase.Policy.OlderThanSeconds != nil {
				olderThan := time.Duration(*testCase.Policy.OlderThanSeconds) * time.Second
				policy.OlderThan = &olderThan
			}
			entries := make([]Entry, 0, len(testCase.Entries))
			for _, entry := range testCase.Entries {
				entries = append(entries, Entry{
					Source: entry.Source, Commit: entry.Commit,
					LastUsed: mustTime(t, entry.LastUsedAt), Reachable: entry.Reachable,
				})
			}
			got := Plan(entries, testCase.ReferenceSetCertain, policy, now)
			if len(got) != len(testCase.Expected) {
				t.Fatalf("plan has %d entries, want %d", len(got), len(testCase.Expected))
			}
			for index, decision := range got {
				observed := retentionOutcome{Source: decision.Source, Commit: decision.Commit, Action: decision.Action, Reason: decision.Reason}
				if observed != testCase.Expected[index] {
					t.Errorf("entry %d = %+v, want %+v", index, observed, testCase.Expected[index])
				}
			}
		})
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}
