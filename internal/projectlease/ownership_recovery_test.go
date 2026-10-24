package projectlease

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// Once a new acquisition settles admission intent, the stale acquisition must
// neither recreate its journal nor roll back the recovered binding. Adjacent
// backup/install boundaries and unchanged-lock admissions are positive controls.
func TestAcquireCannotUndoRecoveredOwner(t *testing.T) {
	for _, point := range []transaction.Point{transaction.PointBeforeBackup, transaction.PointBeforeInstall, transaction.PointAfterInstall, transaction.PointBeforeCleanup, transaction.PointAfterCleanupRename, transaction.PointDuringCleanupRemoval} {
		for _, replace := range []bool{false, true} {
			t.Run(string(point)+map[bool]string{false: "/held", true: "/replaced"}[replace], func(t *testing.T) {
				f := setup(t)
				acquire(t, f, "first")
				request := nextAdmission(t, &f)
				reached := false
				f.m.hooks = transaction.Hooks{Observe: func(e transaction.Event) {
					if reached || e.Point != point {
						return
					}
					reached = true
					if !replace {
						return
					}
					paths, err := filepath.Glob(filepath.Join(f.home, "state", "locks", "v1", "*.lock"))
					if err != nil || len(paths) != 1 {
						t.Fatalf("lock fixture: %v %v", paths, err)
					}
					if err = os.Rename(paths[0], paths[0]+".old"); err != nil {
						if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(32)) {
							t.Skip("host refuses open-lock replacement")
						}
						t.Fatal(err)
					}
					if err = os.WriteFile(paths[0], nil, 0600); err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(os.Args[0], "-test.run=^TestLeaseRecoveryOwnerHelper$", "-test.v")
					cmd.Env = append(os.Environ(), "CURATOR_TEST_RECOVERY_OWNER_HOME="+f.home)
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Fatalf("new owner recovery: %v\n%s", err, out)
					}
				}}
				_, err := f.m.Acquire(context.Background(), request)
				if !reached {
					t.Fatal("boundary not reached")
				}
				if !replace && err != nil {
					t.Fatal("held owner refused", err)
				}
				statuses, statusErr := f.m.Status(context.Background())
				if statusErr != nil || len(statuses) != 2 {
					t.Fatalf("stale owner damaged recovered state: %v %v", statuses, statusErr)
				}
				if _, err := os.Stat(bindingPath(f.home, "next")); err != nil {
					t.Fatal("stale owner removed recovered binding", err)
				}
				if replace && !errors.Is(err, managerlock.ErrNotHeld) {
					t.Fatalf("stale owner refusal: %v", err)
				}
				lm, err := managerlock.New(f.home)
				if err != nil {
					t.Fatal(err)
				}
				lock, err := lm.AcquireHomeOnly(t.Context(), false)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = lock.Close() }()
				engine, _ := transaction.New(f.home)
				if pending, err := engine.PendingRecovery(lock); err != nil || pending {
					t.Fatalf("stale owner recreated settled intent: %v %v", pending, err)
				}
			})
		}
	}
}

// The subprocess uses a distinct real lock acquisition and verifies the
// recovered ledger before the parent resumes its stale in-memory transaction.
func TestLeaseRecoveryOwnerHelper(t *testing.T) {
	home := os.Getenv("CURATOR_TEST_RECOVERY_OWNER_HOME")
	if home == "" {
		return
	}
	f := atHome(t, home)
	if err := f.m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if statuses, err := f.m.Status(context.Background()); err != nil || len(statuses) != 2 {
		t.Fatalf("new owner status: %v %v", statuses, err)
	}
}
