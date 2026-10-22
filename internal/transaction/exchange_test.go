//go:build darwin || linux

package transaction

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each real subprocess exits at the named durable boundary; restart sees a
// complete old/new tree, recovers in the recorded direction, retains exact
// permissions, and removes the journal and sidecars. Prepared/commit recovery
// needs no source bytes; interrupted preparation retains its source witness.
// The ordinary commit is a neighboring positive control for these crash cases.
func TestExchangeCrashRecovery(t *testing.T) {
	for _, direction := range []string{"commit", "rollback"} {
		points := []Point{PointAfterStagingChunkSync, PointDuringStagingCopy, PointPrepared, PointBeforeBackup, PointAfterBackup, PointBeforeInstall, PointAfterInstall, PointTargetCommitted, PointBeforeCleanup, PointAfterCleanupRename, PointDuringCleanupRemoval}
		if direction == "rollback" {
			points = []Point{PointAfterRestore, PointTargetRolledBack, PointAfterCleanupRename, PointDuringCleanupRemoval}
		}
		for _, point := range append([]Point{"success"}, points...) {
			t.Run(direction+"/"+string(point), func(t *testing.T) {
				root := t.TempDir()
				live, source := filepath.Join(root, "live"), filepath.Join(root, "source")
				for _, dir := range []string{live, source} {
					mustMkdirAll(t, filepath.Join(dir, "nested"))
					mustWrite(t, filepath.Join(dir, "nested", "stable.md"), "stable")
				}
				mustWrite(t, filepath.Join(live, "SKILL.md"), "old")
				mustWrite(t, filepath.Join(source, "SKILL.md"), "new")
				old, err := DigestPath(live)
				if err != nil {
					t.Fatal(err)
				}
				desired, err := DigestPath(source)
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(os.Args[0], "-test.run=^TestExchangeCrashHelper$")
				cmd.Env = append(os.Environ(), "CURATOR_EXCHANGE_TEST_ROOT="+root, "CURATOR_EXCHANGE_TEST_POINT="+string(point), "CURATOR_EXCHANGE_TEST_DIRECTION="+direction)
				output, err := cmd.CombinedOutput()
				if point == "success" {
					if err != nil {
						t.Fatalf("positive control: %v\n%s", err, output)
					}
				} else {
					var exit *exec.ExitError
					if !errors.As(err, &exit) || exit.ExitCode() != 86 {
						t.Fatalf("crash control missed: %v\n%s", err, output)
					}
				}
				before, err := DigestPath(live)
				if err != nil || (before != old && before != desired) {
					t.Fatalf("crash exposed missing/mixed tree: %s %v", before, err)
				}
				if point != PointAfterStagingChunkSync && point != PointDuringStagingCopy {
					if err := os.RemoveAll(source); err != nil {
						t.Fatal(err)
					}
				}
				engine := mustEngine(t, filepath.Join(root, "home"))
				if err := engine.Recover(testLock{}); err != nil {
					t.Fatal(err)
				}
				want := desired
				if direction == "rollback" || point == PointAfterStagingChunkSync || point == PointDuringStagingCopy {
					want = old
				}
				if got, err := DigestPath(live); err != nil || got != want {
					t.Fatalf("recovery bytes/modes = %s %v, want %s", got, err, want)
				}
				if pending, err := engine.PendingRecovery(testLock{}); err != nil || pending {
					t.Fatalf("recovery journal remains: %v %t", err, pending)
				}
				matches, err := filepath.Glob(filepath.Join(root, ".curator-txn-*"))
				if err != nil || len(matches) != 0 {
					t.Fatalf("sidecars remain: %v %v", matches, err)
				}
			})
		}
	}
}

// Unsupported absence/removal/generation exchange contracts refuse before
// staging or journal creation; the present/present digest contract succeeds.
func TestExchangePreparationContract(t *testing.T) {
	for _, mode := range []string{"valid", "absent", "removal", "generation"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			target := fileTarget(t, "projection", "fixture", filepath.Join(root, "live"), filepath.Join(root, "source"), "old", "new")
			target.Kind = KindExchange
			switch mode {
			case "absent":
				target.PreimageDigest = DigestAbsent
			case "removal":
				target.StagedSource = ""
			case "generation":
				target.PreimageDigest = ""
				target.ExpectedGeneration = "old"
			}
			engine := mustEngine(t, filepath.Join(root, "home"))
			_, err := engine.Prepare(testLock{}, Plan{TransactionID: "exchange-contract", ProjectIdentity: "projection", Targets: []Target{target}})
			if mode == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				if err := engine.Commit(testLock{}, "exchange-contract"); err != nil {
					t.Fatal(err)
				}
				if got := mustRead(t, target.LivePath); got != "new" {
					t.Fatal("valid exchange did not install")
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), "exchange requires supported present byte preimages and replacements") {
					t.Fatalf("wrong contract refusal: %v", err)
				}
				if got := mustRead(t, target.LivePath); got != "old" {
					t.Fatal("refusal changed live bytes")
				}
				if _, err := os.Lstat(engine.journalPath("exchange-contract")); !os.IsNotExist(err) {
					t.Fatalf("refusal wrote journal: %v", err)
				}
				matches, err := filepath.Glob(filepath.Join(root, ".curator-txn-*"))
				if err != nil || len(matches) != 0 {
					t.Fatalf("refusal wrote sidecars: %v %v", matches, err)
				}
			}
		})
	}
}

// Unknown backup/staging/rollback bytes are not exchanged or deleted; the clean
// layout is a positive control. Refusal preserves published old bytes and the
// journal plus unknown sidecar for reconciliation.
func TestExchangeUnknownBackupPreserved(t *testing.T) {
	for _, damage := range []string{"clean", "backup", "desired", "rollback"} {
		t.Run(damage, func(t *testing.T) {
			root := t.TempDir()
			target := fileTarget(t, "projection", "fixture", filepath.Join(root, "live"), filepath.Join(root, "source"), "old", "new")
			target.Kind = KindExchange
			unknown := sidecarPath(target.LivePath, "exchange-unknown", 0, damage)
			engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Observe: func(event Event) {
				if damage != "clean" && event.Point == PointBeforeInstall {
					mustWrite(t, unknown, "unknown")
				}
			}}))
			if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "exchange-unknown", ProjectIdentity: "projection", Targets: []Target{target}}); err != nil {
				t.Fatal(err)
			}
			err := engine.Commit(testLock{}, "exchange-unknown")
			if damage == "clean" {
				if err != nil {
					t.Fatal(err)
				}
				if got := mustRead(t, target.LivePath); got != "new" {
					t.Fatal("control did not install")
				}
				return
			}
			if !errors.Is(err, ErrImplementationCorruption) {
				t.Fatalf("wrong unknown-backup refusal: %v", err)
			}
			if mustRead(t, target.LivePath) != "old" || mustRead(t, unknown) != "unknown" {
				t.Fatal("refusal changed published or unknown bytes")
			}
			if pending, err := engine.PendingRecovery(testLock{}); err != nil || !pending {
				t.Fatalf("refusal lost journal: %v %t", err, pending)
			}
		})
	}
}

// Invoked only by the crash test with paths created in its disposable fixture.
func TestExchangeCrashHelper(t *testing.T) {
	root := os.Getenv("CURATOR_EXCHANGE_TEST_ROOT")
	if root == "" {
		return
	}
	point := Point(os.Getenv("CURATOR_EXCHANGE_TEST_POINT"))
	rollback := os.Getenv("CURATOR_EXCHANGE_TEST_DIRECTION") == "rollback"
	fault := errors.New("exchange rollback control")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == point {
			os.Exit(86)
		}
		if rollback && event.Point == PointAfterInstall {
			return fault
		}
		return nil
	}}))
	live := filepath.Join(root, "live")
	preimage, err := DigestPath(live)
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Prepare(testLock{}, Plan{TransactionID: "exchange-crash", ProjectIdentity: "projection", Targets: []Target{{Class: "skill-projection", Identifier: "fixture", Kind: KindExchange, LivePath: live, StagedSource: filepath.Join(root, "source"), PreimageDigest: preimage}}})
	if err != nil {
		t.Fatal(err)
	}
	err = engine.Commit(testLock{}, "exchange-crash")
	if rollback {
		if !errors.Is(err, fault) {
			t.Fatalf("rollback control: %v", err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	if point != "success" {
		t.Fatal("intended crash boundary was not reached")
	}
}
