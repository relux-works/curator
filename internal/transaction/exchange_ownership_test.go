//go:build darwin || linux

package transaction

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
)

// The real Engine must not publish after its boundary callback returns under a
// replaced acquisition. A separate process holds the new real home lock. Bytes
// targets and unchanged exchange acquisitions are neighboring positive controls.
func TestExchangeOwnershipAtBoundary(t *testing.T) {
	for _, point := range []Point{PointBeforeBackup, PointBeforeInstall} {
		for _, kind := range []TargetKind{KindBytes, KindExchange} {
			for _, replace := range []bool{false, true} {
				name := string(kind)
				if kind == KindBytes {
					name = "bytes"
				}
				t.Run(string(point)+"/"+name+map[bool]string{false: "/held", true: "/replaced"}[replace], func(t *testing.T) {
					root := t.TempDir()
					home := filepath.Join(root, "home")
					manager, err := managerlock.New(home)
					if err != nil {
						t.Fatal(err)
					}
					lock, err := manager.AcquireHomeOnly(context.Background(), false)
					if err != nil {
						t.Fatal(err)
					}
					defer lock.Close()
					target := fileTarget(t, "projection", "fixture", filepath.Join(root, "live"), filepath.Join(root, "source"), "old", "new")
					target.Kind = kind
					installing := false
					reached := false
					engine := mustEngine(t, home, WithHooks(Hooks{Observe: func(e Event) {
						if e.Point == point {
							installing = true
						}
					}}))
					plan := Plan{TransactionID: "exchange-ownership", ProjectIdentity: "projection", Targets: []Target{target}, BoundaryCheck: func(_ int, _ string) error {
						if !installing || reached {
							return nil
						}
						reached = true
						if !replace {
							return nil
						}
						replaceExchangeOwner(t, home, lock, false)
						return nil
					}}
					prepared, err := engine.Prepare(lock, plan)
					if err != nil {
						t.Fatal(err)
					}
					err = engine.Commit(lock, plan.TransactionID)
					if !reached {
						t.Fatal("boundary missed")
					}
					if !replace {
						if err != nil || mustRead(t, target.LivePath) != "new" {
							t.Fatalf("positive control %v", err)
						}
						return
					}
					if !errors.Is(err, managerlock.ErrNotHeld) {
						t.Fatalf("wrong refusal: %v", err)
					}
					got, readErr := os.ReadFile(target.LivePath)
					if kind == KindBytes && point == PointBeforeInstall {
						if !os.IsNotExist(readErr) {
							t.Fatalf("stale bytes owner wrote live: %q %v", got, readErr)
						}
					} else if readErr != nil || string(got) != "old" {
						t.Fatalf("stale exchange owner published after replacement: got %q error %v; commit %v", got, readErr, err)
					}
					if point == PointBeforeBackup && kind == KindExchange {
						record := prepared.Targets[0]
						if got, err := DigestPath(record.StagedPath); err != nil || got != record.DesiredDigest {
							t.Fatalf("stale exchange owner moved staging: %s %v", got, err)
						}
						if got, err := DigestPath(record.BackupPath); err != nil || got != DigestAbsent {
							t.Fatalf("stale exchange owner created backup: %s %v", got, err)
						}
					}
				})
			}
		}
	}
}

// Child helper holds a distinct acquisition until the parent closes stdin.
func TestExchangeNewOwnerHelper(t *testing.T) {
	home := os.Getenv("CURATOR_EXCHANGE_OWNER_HOME")
	if home == "" {
		return
	}
	m, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := m.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if os.Getenv("CURATOR_EXCHANGE_OWNER_RECOVER") == "true" {
		if err := mustEngine(t, home).Recover(lock); err != nil {
			t.Fatal(err)
		}
	}
	if err = lock.Held().Validate(home); err != nil {
		t.Fatal(err)
	}
	fmt.Println("held")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

// replaceExchangeOwner acquires a real replacement lock in a bounded subprocess.
// Its stdin lifetime keeps that acquisition held throughout the stale attempt.
func replaceExchangeOwner(t *testing.T, home string, old *managerlock.HomeLock, recover bool) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, "state", "locks", "v1", "*.lock"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("lock paths %v %v", paths, err)
	}
	if err := os.Rename(paths[0], paths[0]+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths[0], nil, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestExchangeNewOwnerHelper$")
	child.Env = append(os.Environ(), "CURATOR_EXCHANGE_OWNER_HOME="+home, fmt.Sprintf("CURATOR_EXCHANGE_OWNER_RECOVER=%t", recover))
	child.Stderr = os.Stderr
	input, err := child.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	output, err := child.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	started := false
	t.Cleanup(func() {
		defer cancel()
		_ = input.Close()
		if started {
			if err := child.Wait(); err != nil {
				t.Error("new owner", err)
			}
		}
	})
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	started = true
	t.Logf("new owner PID %d", child.Process.Pid)
	line, err := bufio.NewReader(output).ReadString('\n')
	if err != nil || line != "held\n" {
		t.Fatalf("new owner not held %q %v", line, err)
	}
	if err := old.Held().Validate(home); !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("old acquisition still valid: %v", err)
	}
}

// Both inverse layouts retain every byte and journal byte under a stale real
// acquisition. Held controls restore the old live bytes despite a denied forward
// boundary, proving rollback is gated by ownership rather than admission policy.
// These helper-level cases isolate the checks after the rollback digest reads.
func TestExchangeRollbackOwnership(t *testing.T) {
	for _, installed := range []bool{false, true} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("installed=%t/replaced=%t", installed, replace), func(t *testing.T) {
				root := t.TempDir()
				home := filepath.Join(root, "home")
				manager, err := managerlock.New(home)
				if err != nil {
					t.Fatal(err)
				}
				lock, err := manager.AcquireHomeOnly(t.Context(), false)
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Close()
				target := fileTarget(t, "projection", "fixture", filepath.Join(root, "live"), filepath.Join(root, "source"), "old", "new")
				target.Kind = KindExchange
				engine := mustEngine(t, home)
				journal, err := engine.Prepare(lock, Plan{TransactionID: "exchange-rollback-owner", ProjectIdentity: "projection", Targets: []Target{target}})
				if err != nil {
					t.Fatal(err)
				}
				record := &journal.Targets[0]
				if err := durableRenameNoReplace(record.StagedPath, record.BackupPath); err != nil {
					t.Fatal(err)
				}
				if installed {
					if err := durableExchange(record.BackupPath, record.LivePath); err != nil {
						t.Fatal(err)
					}
				}
				journal.Phase = PhaseRollingBack
				record.State = StateBackedUp
				record.BackupDigest = record.PreimageDigest
				if err := engine.saveJournal(journal); err != nil {
					t.Fatal(err)
				}
				paths := []string{record.LivePath, record.BackupPath, record.RollbackPath, engine.journalPath(journal.TransactionID)}
				before := make([]string, len(paths))
				for i, path := range paths {
					before[i], err = DigestPath(path)
					if err != nil {
						t.Fatal(err)
					}
				}
				if replace {
					replaceExchangeOwner(t, home, lock, false)
				}
				policyCalls := 0
				engine.guards[journal.TransactionID] = func(int, string) error { policyCalls++; return errors.New("forward policy refused") }
				engine.commitLock = lock
				engine.mu.Lock()
				err = engine.rollbackExchange(journal, 0)
				engine.mu.Unlock()
				engine.commitLock = nil
				if policyCalls != 0 {
					t.Fatal("rollback consulted forward policy")
				}
				if replace {
					if !errors.Is(err, managerlock.ErrNotHeld) {
						t.Fatalf("stale rollback refusal: %v", err)
					}
					for i, path := range paths {
						if got, err := DigestPath(path); err != nil || got != before[i] {
							t.Fatalf("stale rollback changed %s: %s %v", filepath.Base(path), got, err)
						}
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if mustRead(t, record.LivePath) != "old" || mustRead(t, record.RollbackPath) != "new" {
						t.Fatal("held rollback did not restore the preimage")
					}
					if got, err := DigestPath(record.BackupPath); err != nil || got != DigestAbsent {
						t.Fatalf("held rollback backup: %s %v", got, err)
					}
				}
			})
		}
	}
}

// At forward and cleanup boundaries a distinct owner recovers the real journal
// before the stale process resumes. The recovered bytes and settled journal
// survive that stale return; unchanged acquisitions complete as positive controls.
func TestExchangeRecoveredOwnerPreserved(t *testing.T) {
	for _, point := range []Point{PointBeforeInstall, PointAfterInstall, PointBeforeCleanup, PointAfterCleanupRename} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/replaced=%t", point, replace), func(t *testing.T) {
				root := t.TempDir()
				home := filepath.Join(root, "home")
				manager, err := managerlock.New(home)
				if err != nil {
					t.Fatal(err)
				}
				lock, err := manager.AcquireHomeOnly(t.Context(), false)
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Close()
				target := fileTarget(t, "projection", "fixture", filepath.Join(root, "live"), filepath.Join(root, "source"), "old", "new")
				target.Kind = KindExchange
				reached := false
				engine := mustEngine(t, home, WithHooks(Hooks{Observe: func(e Event) {
					if reached || e.Point != point {
						return
					}
					reached = true
					if replace {
						replaceExchangeOwner(t, home, lock, true)
					}
				}}))
				plan := Plan{TransactionID: "exchange-recovered-owner", ProjectIdentity: "projection", Targets: []Target{target}}
				if _, err := engine.Prepare(lock, plan); err != nil {
					t.Fatal(err)
				}
				err = engine.Commit(lock, plan.TransactionID)
				if !reached {
					t.Fatal("recovery boundary missed")
				}
				if replace {
					if !errors.Is(err, managerlock.ErrNotHeld) {
						t.Fatalf("stale recovery refusal: %v", err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				if mustRead(t, target.LivePath) != "new" {
					t.Fatal("stale owner changed recovered bytes")
				}
				if _, err := os.Lstat(engine.journalPath(plan.TransactionID)); !os.IsNotExist(err) {
					t.Fatalf("stale owner recreated settled journal: %v", err)
				}
				paths, err := filepath.Glob(filepath.Join(root, ".curator-txn-*"))
				if err != nil || len(paths) != 0 {
					t.Fatalf("recovery left sidecars: %v %v", paths, err)
				}
			})
		}
	}
}
