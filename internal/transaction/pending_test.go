package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// PendingRecovery preserves canonical records in every phase, including cleanup
// tombs, while the held-witness and invalid-evidence controls fail closed.
func TestPendingRecoveryReadOnlyPhasesAndTombs(t *testing.T) {
	for _, phase := range []Phase{PhasePrepared, PhaseCommitting, PhaseCleanup, PhaseRollingBack, PhaseRollbackCleanup} {
		t.Run(string(phase), func(t *testing.T) {
			root := t.TempDir()
			engine := mustEngine(t, filepath.Join(root, "home"))
			if pending, err := engine.PendingRecovery(testLock{}); err != nil || pending {
				t.Fatalf("absent inventory = %v, %v", pending, err)
			}
			target := fileTarget(t, "profile", "one", filepath.Join(root, "live"), filepath.Join(root, "stage"), "old", "new")
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "pending", ProjectIdentity: "capture", Targets: []Target{target}})
			if err != nil {
				t.Fatal(err)
			}
			journal.Phase = phase
			if err := engine.saveJournal(journal); err != nil {
				t.Fatal(err)
			}
			path := engine.journalPath("pending")
			if phase == PhaseCleanup || phase == PhaseRollbackCleanup {
				if err := os.Rename(path, path+".delete"); err != nil {
					t.Fatal(err)
				}
				path += ".delete"
			}
			before := mustRead(t, path)
			if pending, err := engine.PendingRecovery(testLock{}); err != nil || !pending {
				t.Fatalf("pending inventory = %v, %v", pending, err)
			}
			// Observation exposes detached intent, never live mutable engine state.
			observed, err := engine.PendingJournals(testLock{})
			if err != nil || len(observed) != 1 || observed[0].TransactionID != "pending" || observed[0].Phase != phase {
				t.Fatalf("observed journal = %v, %v", observed, err)
			}
			observed[0].Targets[0].LivePath = "caller mutation"
			again, err := engine.PendingJournals(testLock{})
			if err != nil || len(again) != 1 || again[0].Targets[0].LivePath != target.LivePath {
				t.Fatalf("caller mutation changed durable observation: %v", err)
			}
			if mustRead(t, path) != before || mustRead(t, target.LivePath) != "old" {
				t.Fatal("observation mutated recovery state")
			}
			if _, err := engine.PendingRecovery(testLock{err: errors.New("released")}); err == nil {
				t.Fatal("accepted released witness")
			}
			if observed, err := engine.PendingJournals(testLock{err: errors.New("released")}); err == nil || observed != nil {
				t.Fatal("journal projection accepted released witness")
			}
			mustWrite(t, engine.journalPath("z-corrupt"), "{")
			if pending, err := engine.PendingRecovery(testLock{}); err == nil || pending {
				t.Fatalf("valid record masked corrupt evidence: %v, %v", pending, err)
			}
			if observed, err := engine.PendingJournals(testLock{}); err == nil || observed != nil {
				t.Fatal("projection returned partial intent despite corrupt inventory")
			}
		})
	}
}

// A linked inventory cannot be treated as an empty or valid recovery set.
func TestPendingRecoveryRejectsLinkedInventory(t *testing.T) {
	home := t.TempDir()
	engine := mustEngine(t, home)
	if err := os.MkdirAll(filepath.Dir(engine.journalRoot), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), engine.journalRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if pending, err := engine.PendingRecovery(testLock{}); err == nil || pending {
		t.Fatalf("linked inventory = %v, %v", pending, err)
	}
}

// The instrumentation detects Recover even with no transactions, so a zero
// recovery count at capture is evidence of non-entry, not merely no writes.
func TestRecoveryObserverDetectsEmptyRecovery(t *testing.T) {
	calls := 0
	engine := mustEngine(t, t.TempDir(), WithHooks(Hooks{ObserveRecovery: func() { calls++ }}))
	if err := engine.Recover(testLock{}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("recovery entry count = %d", calls)
	}
}
