package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareSyncsStagedParentBeforePreparedJournal(t *testing.T) {
	for _, testCase := range preparationDurabilityCases() {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			target := testCase.target(t, root)
			engine := mustEngine(t, home)
			observed := false
			engine.syncStagedParent = func(parent string) error {
				observed = true
				journal, err := engine.loadJournal("txn-stage-durable")
				if err != nil {
					t.Fatalf("load journal at staged-parent sync: %v", err)
				}
				if journal.Phase != PhasePreparing {
					t.Fatalf("journal phase at staged-parent sync = %q, want %q", journal.Phase, PhasePreparing)
				}
				staged := journal.Targets[0].StagedPath
				if parent != filepath.Dir(staged) {
					t.Fatalf("staged parent sync path = %q, want %q", parent, filepath.Dir(staged))
				}
				got, err := DigestPath(staged)
				if err != nil {
					t.Fatalf("digest staged target before parent sync: %v", err)
				}
				if got != journal.Targets[0].DesiredDigest {
					t.Fatalf("staged digest before parent sync = %q, want %q", got, journal.Targets[0].DesiredDigest)
				}
				return syncDirectory(parent)
			}

			journal, err := engine.Prepare(testLock{}, Plan{
				TransactionID:   "txn-stage-durable",
				ProjectIdentity: "/project",
				Targets:         []Target{target},
			})
			if err != nil {
				t.Fatal(err)
			}
			if !observed {
				t.Fatal("staged parent durability primitive was not called")
			}
			if journal.Phase != PhasePrepared {
				t.Fatalf("returned journal phase = %q, want %q", journal.Phase, PhasePrepared)
			}
		})
	}
}

func TestPrepareStagedParentSyncFailureDoesNotPublishPreparedJournal(t *testing.T) {
	for _, testCase := range preparationDurabilityCases() {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			target := testCase.target(t, root)
			preimage, err := DigestPath(target.LivePath)
			if err != nil {
				t.Fatal(err)
			}
			engine := mustEngine(t, home)
			injected := errors.New("staged parent sync failed")
			called := false
			engine.syncStagedParent = func(string) error {
				called = true
				journal, err := engine.loadJournal("txn-stage-sync-failure")
				if err != nil {
					t.Fatalf("load journal at failed staged-parent sync: %v", err)
				}
				if journal.Phase != PhasePreparing {
					t.Fatalf("journal phase at failed staged-parent sync = %q, want %q", journal.Phase, PhasePreparing)
				}
				return injected
			}

			_, err = engine.Prepare(testLock{}, Plan{
				TransactionID:   "txn-stage-sync-failure",
				ProjectIdentity: "/project",
				Targets:         []Target{target},
			})
			if !errors.Is(err, injected) {
				t.Fatalf("prepare error = %v, want staged parent sync failure", err)
			}
			if !called {
				t.Fatal("staged parent durability primitive was not called")
			}
			if _, err := os.Lstat(engine.journalPath("txn-stage-sync-failure")); !os.IsNotExist(err) {
				t.Fatalf("failed staging exposed a journal: %v", err)
			}
			if got, err := DigestPath(target.LivePath); err != nil || got != preimage {
				t.Fatalf("failed staging changed live target: digest %q, error %v; want %q", got, err, preimage)
			}
			staged := sidecarPath(target.LivePath, "txn-stage-sync-failure", 0, "desired")
			if _, err := os.Lstat(staged); !os.IsNotExist(err) {
				t.Fatalf("failed staging left sidecar %q: %v", staged, err)
			}
		})
	}
}

type preparationDurabilityCase struct {
	name   string
	target func(*testing.T, string) Target
}

func preparationDurabilityCases() []preparationDurabilityCase {
	return []preparationDurabilityCase{
		{
			name: "regular file",
			target: func(t *testing.T, root string) Target {
				return fileTarget(t, "class", "file", filepath.Join(root, "live", "file"), filepath.Join(root, "stage", "file"), "old", "new")
			},
		},
		{
			name: "directory",
			target: func(t *testing.T, root string) Target {
				live := filepath.Join(root, "live", "directory")
				source := filepath.Join(root, "stage", "directory")
				mustWrite(t, filepath.Join(live, "bin", "tool"), "old")
				mustWrite(t, filepath.Join(source, "bin", "tool"), "new")
				preimage, err := DigestPath(live)
				if err != nil {
					t.Fatal(err)
				}
				return Target{Class: "class", Identifier: "directory", LivePath: live, StagedSource: source, PreimageDigest: preimage}
			},
		},
	}
}
