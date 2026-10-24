package projectlease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
)

type assertOnlyLock struct{}

func (assertOnlyLock) AssertHeld() error { return nil }

// GC roots require a real acquisition for exactly this home. The valid
// same-home witness reads roots; assertion-only, foreign, released/reacquired
// and replaced witnesses refuse instead of exposing an unprotected mark set.
func TestGCReferenceWitness(t *testing.T) {
	f := setup(t)
	acquire(t, f, "one")
	lm, err := managerlock.New(f.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := lm.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if roots, err := ReferencesUnderLock(f.home, lock); err != nil || len(roots.Homes) != 1 {
		t.Fatal("valid witness refused roots", err)
	}
	if _, err = ReferencesUnderLock(f.home, assertOnlyLock{}); !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("assertion-only GC witness = %v", err)
	}
	foreign, err := managerlock.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	other, err := foreign.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ReferencesUnderLock(f.home, other); !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("foreign GC witness = %v", err)
	}
	_ = other.Close()
	_ = lock.Close()
	reacquired, err := lm.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reacquired.Close() }()
	if _, err = ReferencesUnderLock(f.home, lock); !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("released GC witness = %v", err)
	}
	// Replace the named lock only in this isolated fixture. The old open
	// descriptor remains locked but is no longer the home's current witness.
	// This fixture has acquired only the home lock, so its lock root has
	// exactly one lock file. Do not duplicate the manager's private lock hash.
	paths, err := filepath.Glob(filepath.Join(f.home, "state", "locks", "v1", "*.lock"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("home lock fixture paths = %v: %v", paths, err)
	}
	path := paths[0]
	if err = os.Rename(path, path+".old"); err != nil {
		if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(32)) {
			t.Skip("host refuses replacement of an open lock: sharing violation")
		}
		t.Fatal(err)
	}
	if err = os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ReferencesUnderLock(f.home, reacquired); !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("replaced GC witness = %v", err)
	}
}
