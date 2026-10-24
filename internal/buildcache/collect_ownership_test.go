package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
)

type retainedSweepLock struct {
	held managerlock.Held
	home string
}

func (lock retainedSweepLock) AssertHeld() error { return lock.held.Validate(lock.home) }

// A loss after receipt classification or immediately before retirement must
// preserve the physical candidate. Unchanged-lock siblings still collect it.
// The supplied witness is the same retained-acquisition contract Collect uses.
func TestSweepRetainsOwnershipAfterClassification(t *testing.T) {
	for _, point := range []string{"classification", "retirement"} {
		for _, replace := range []bool{false, true} {
			t.Run(point+map[bool]string{false: "/held", true: "/replaced"}[replace], func(t *testing.T) {
				store := newSweepStore(t)
				key := publishTestEntry(t, store, "next", "artifact")
				backdate(t, entryPathOf(store, key), 30*24*time.Hour)
				store.ExecutablePaths = func() ([]string, error) { return nil, nil }
				lm, err := managerlock.New(store.Home())
				if err != nil {
					t.Fatal(err)
				}
				lock, err := lm.AcquireHomeOnly(t.Context(), false)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = lock.Close() }()
				reached := false
				hook := func() {
					reached = true
					if !replace {
						return
					}
					paths, err := filepath.Glob(filepath.Join(store.Home(), "state", "locks", "v1", "*.lock"))
					if err != nil || len(paths) != 1 {
						t.Fatal("lock fixture", paths, err)
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
					if err = lock.AssertHeld(); err != nil {
						t.Fatal("fixture unexpectedly dropped descriptor lock", err)
					}
				}
				if point == "classification" {
					beforeClassifyForTests = hook
				} else {
					beforeRetireForTests = hook
				}
				defer func() { beforeClassifyForTests = nil; beforeRetireForTests = nil }()
				result, err := store.Sweep(SweepRequest{}, retainedSweepLock{held: lock.Held(), home: store.Home()})
				if !reached {
					t.Fatal("late sweep boundary not reached")
				}
				if replace {
					if !errors.Is(err, managerlock.ErrNotHeld) || !entryExists(t, store, key) || len(result.Removed) != 0 {
						t.Fatalf("late stale sweep removed candidate: %+v %v", result, err)
					}
				} else if err != nil || entryExists(t, store, key) || len(result.Removed) != 1 {
					t.Fatalf("held sweep control: %+v %v", result, err)
				}
			})
		}
	}
}
