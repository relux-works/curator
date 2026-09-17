package install

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/staging"
)

// boundaryCommitRequest drives the real serialized publisher with caller
// supplied scope targets. The journal is a recording stub so the test can
// prove the recheck runs before journaling; locks are real.
func boundaryCommitRequest(t *testing.T, home string, journal *stubJournal, idCalls *int, targets scopeTargets) commitRequest {
	t.Helper()
	return commitRequest{
		scope: "test",
		home:  home,
		commit: CommitDeps{
			Locks:   realLocks(t, home),
			Journal: journal,
			NewTransactionID: func() (string, error) {
				*idCalls++
				return "test-boundary-txn", nil
			},
		},
		stageTargets: func(scopeCommit) (scopeTargets, error) { return targets, nil },
	}
}

func TestRunCommitRecheckRefusesDestinationInsideAdmitted(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(admitted, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Planning missed the overlap; the publisher must not.
	var plan staging.Plan
	plan.Replace(staging.ClassContext, "a", filepath.Join(admitted, "entry"), filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	journal := &stubJournal{}
	var idCalls int
	_, err = runCommit(context.Background(), boundaryCommitRequest(t, home, journal, &idCalls, scopeTargets{
		plan:       plan,
		boundaries: &staging.Boundaries{Snapshot: snapshot, Admitted: []string{admitted}},
	}))
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("runCommit err = %v, want source_output_overlap", err)
	}
	if journal.prepared != 0 || journal.commits != 0 || idCalls != 0 {
		t.Fatalf("refused commit reached journaling: prepared=%d commits=%d ids=%d", journal.prepared, journal.commits, idCalls)
	}
	if _, statErr := os.Lstat(filepath.Join(admitted, "entry")); !os.IsNotExist(statErr) {
		t.Fatalf("refused commit touched the live destination: %v", statErr)
	}
	payload, err := os.ReadFile(filepath.Join(admitted, "SKILL.md"))
	if err != nil || string(payload) != "x" {
		t.Fatalf("admitted input changed by refused commit: %q, %v", payload, err)
	}
}

func TestRunCommitRecheckRefusesSwappedParent(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan staging.Plan
	live := filepath.Join(parent, "entry")
	plan.Replace(staging.ClassContext, "a", live, filepath.Join(root, "staged"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	// Retarget the parent between planning and publication: same
	// spelling, different object.
	if err := os.Rename(parent, filepath.Join(root, "parent-old")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	journal := &stubJournal{}
	var idCalls int
	_, err = runCommit(context.Background(), boundaryCommitRequest(t, home, journal, &idCalls, scopeTargets{
		plan:       plan,
		boundaries: &staging.Boundaries{Snapshot: snapshot, Admitted: []string{admitted}},
	}))
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("runCommit err = %v, want source_output_overlap", err)
	}
	if journal.prepared != 0 || journal.commits != 0 || idCalls != 0 {
		t.Fatalf("refused commit reached journaling: prepared=%d commits=%d ids=%d", journal.prepared, journal.commits, idCalls)
	}
}

func TestRunCommitRecheckPassesUnchangedBoundaries(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	root := t.TempDir()
	admitted := filepath.Join(root, "authored")
	if err := os.MkdirAll(admitted, 0o755); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(root, "managed")
	if err := os.MkdirAll(managed, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan staging.Plan
	// A removal target stages no bytes, so the pass-through proof needs
	// no staged fixtures and mutates no live state.
	plan.Remove("stale", filepath.Join(managed, "stale"))
	snapshot, err := plan.Snapshot([]string{admitted})
	if err != nil {
		t.Fatal(err)
	}
	journal := &stubJournal{}
	var idCalls int
	_, err = runCommit(context.Background(), boundaryCommitRequest(t, home, journal, &idCalls, scopeTargets{
		plan:       plan,
		boundaries: &staging.Boundaries{Snapshot: snapshot, Admitted: []string{admitted}},
	}))
	if err != nil {
		t.Fatalf("runCommit with unchanged boundaries: %v", err)
	}
	if journal.prepared != 1 || journal.commits != 1 {
		t.Fatalf("prepared=%d commits=%d, want one journaled commit", journal.prepared, journal.commits)
	}
}

func TestRunCommitWithoutBoundariesSkipsRecheck(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	root := t.TempDir()
	managed := filepath.Join(root, "managed")
	if err := os.MkdirAll(managed, 0o755); err != nil {
		t.Fatal(err)
	}
	var plan staging.Plan
	plan.Remove("stale", filepath.Join(managed, "stale"))
	journal := &stubJournal{}
	var idCalls int
	// Legacy scopes carry no boundary record: identical behavior, no gate.
	_, err := runCommit(context.Background(), boundaryCommitRequest(t, home, journal, &idCalls, scopeTargets{plan: plan}))
	if err != nil {
		t.Fatalf("runCommit without boundaries: %v", err)
	}
	if journal.prepared != 1 || journal.commits != 1 {
		t.Fatalf("prepared=%d commits=%d, want one journaled commit", journal.prepared, journal.commits)
	}
}
