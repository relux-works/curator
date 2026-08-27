package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRollbackCorruptionBranchesNeverOverwriteUnknownState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *TargetRecord)
		check  func(*testing.T, *TargetRecord)
	}{
		{
			name:   "unknown live bytes",
			mutate: func(t *testing.T, target *TargetRecord) { mustWrite(t, target.LivePath, "unknown-live") },
			check: func(t *testing.T, target *TargetRecord) {
				if got := mustRead(t, target.LivePath); got != "unknown-live" {
					t.Fatalf("unknown live = %q", got)
				}
			},
		},
		{
			name: "committed target disappeared",
			mutate: func(t *testing.T, target *TargetRecord) {
				if err := os.RemoveAll(target.LivePath); err != nil {
					t.Fatal(err)
				}
			},
			check: func(t *testing.T, target *TargetRecord) {
				if _, err := os.Lstat(target.LivePath); !os.IsNotExist(err) {
					t.Fatalf("missing live replaced: %v", err)
				}
			},
		},
		{
			name: "unknown captured rollback bytes",
			mutate: func(t *testing.T, target *TargetRecord) {
				if err := durableRenameNoReplace(target.LivePath, target.RollbackPath); err != nil {
					t.Fatal(err)
				}
				mustWrite(t, target.RollbackPath, "unknown-rollback")
			},
			check: func(t *testing.T, target *TargetRecord) {
				if got := mustRead(t, target.LivePath); got != "unknown-rollback" {
					t.Fatalf("unknown rollback = %q", got)
				}
			},
		},
		{
			name: "live and rollback both exist",
			mutate: func(t *testing.T, target *TargetRecord) {
				if err := copyTarget(target.LivePath, target.RollbackPath); err != nil {
					t.Fatal(err)
				}
			},
			check: func(t *testing.T, target *TargetRecord) {
				if got := mustRead(t, target.LivePath); got != "new" {
					t.Fatalf("live = %q", got)
				}
			},
		},
		{
			name:   "backup changed",
			mutate: func(t *testing.T, target *TargetRecord) { mustWrite(t, target.BackupPath, "unknown-backup") },
			check: func(t *testing.T, target *TargetRecord) {
				if got := mustRead(t, target.BackupPath); got != "unknown-backup" {
					t.Fatalf("backup = %q", got)
				}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			engine, journal := committedCleanupFixture(t)
			journal.Phase = PhaseRollingBack
			if err := engine.saveJournal(journal); err != nil {
				t.Fatal(err)
			}
			testCase.mutate(t, &journal.Targets[0])
			err := engine.rollback(journal)
			if !errors.Is(err, ErrImplementationCorruption) {
				t.Fatalf("rollback error = %v, want implementation-corruption", err)
			}
			testCase.check(t, &journal.Targets[0])
		})
	}
}

func TestCommitRecoveryRejectsAmbiguousPendingBackups(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*testing.T, *TargetRecord)
	}{
		{
			name: "live and backup",
			mutate: func(t *testing.T, target *TargetRecord) {
				if err := copyTarget(target.LivePath, target.BackupPath); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "wrong backup digest",
			mutate: func(t *testing.T, target *TargetRecord) {
				if err := durableRenameNoReplace(target.LivePath, target.BackupPath); err != nil {
					t.Fatal(err)
				}
				mustWrite(t, target.BackupPath, "unknown-backup")
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			mustMkdirAll(t, filepath.Join(root, "live"))
			mustMkdirAll(t, filepath.Join(root, "stage"))
			target := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
			engine := mustEngine(t, filepath.Join(root, "home"))
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-pending", ProjectIdentity: "/project", Targets: []Target{target}})
			if err != nil {
				t.Fatal(err)
			}
			testCase.mutate(t, &journal.Targets[0])
			err = engine.Commit(testLock{}, journal.TransactionID)
			if !errors.Is(err, ErrImplementationCorruption) {
				t.Fatalf("commit error = %v, want implementation-corruption", err)
			}
		})
	}
}

func TestLoadJournalRejectsNoncanonicalUnknownAndMultipleJSON(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(string) string
	}{
		{name: "noncanonical", mutate: func(payload string) string { return " " + payload }},
		{name: "unknown field", mutate: func(payload string) string {
			return strings.Replace(payload, `{"schema":`, `{"unknown":true,"schema":`, 1)
		}},
		{name: "multiple values", mutate: func(payload string) string { return payload + "{}\n" }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			engine, journal := preparedFixture(t)
			path := engine.journalPath(journal.TransactionID)
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mustWrite(t, path, testCase.mutate(string(payload)))
			if _, err := engine.loadJournal(journal.TransactionID); err == nil {
				t.Fatal("invalid journal was loaded")
			}
		})
	}

	t.Run("symlink", func(t *testing.T) {
		engine, journal := preparedFixture(t)
		path := engine.journalPath(journal.TransactionID)
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		external := filepath.Join(t.TempDir(), "external.json")
		mustWrite(t, external, string(payload))
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(external, path); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		if _, err := engine.loadJournal(journal.TransactionID); err == nil {
			t.Fatal("journal symlink was loaded")
		}
	})
}

func committedCleanupFixture(t *testing.T) (*Engine, *Journal) {
	t.Helper()
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	target := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
	stop := errors.New("stop at cleanup")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointBeforeCleanup {
			return stop
		}
		return nil
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-corruption", ProjectIdentity: "/project", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); !errors.Is(err, stop) {
		t.Fatalf("commit = %v", err)
	}
	journal, err = engine.loadJournal(journal.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	engine.hooks = Hooks{}
	return engine, journal
}

func preparedFixture(t *testing.T) (*Engine, *Journal) {
	t.Helper()
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	target := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
	engine := mustEngine(t, filepath.Join(root, "home"))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-load", ProjectIdentity: "/project", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	return engine, journal
}
