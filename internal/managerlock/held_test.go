package managerlock

import (
	"context"
	"errors"
	"os"
	"runtime"
	"syscall"
	"testing"
)

// A witness proves only its live acquisition's exact home. Zero, another
// home's lock and a released acquisition refuse; reacquisition cannot revive it.
// Both writer/observer paths and ordered project operations mint valid witnesses.
func TestHeldWitnessScopeAndLifetime(t *testing.T) {
	home, other := t.TempDir(), t.TempDir()
	if !errors.Is((Held{}).Validate(home), ErrNotHeld) {
		t.Fatal("zero witness accepted")
	}
	if !errors.Is((&HomeLock{}).Held().Validate(home), ErrNotHeld) {
		t.Fatal("unacquired witness accepted")
	}
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"writer", "observer", "ordered"} {
		t.Run(mode, func(t *testing.T) {
			var lock *HomeLock
			var err error
			switch mode {
			case "ordered":
				op := manager.NewOperation(false)
				defer func() { _ = op.Close() }()
				if _, err = op.AcquireProjects(context.Background(), t.TempDir()); err != nil {
					t.Fatal(err)
				}
				lock, err = op.AcquireHome(context.Background())
			case "observer":
				lock, err = manager.AcquireExistingHomeOnly(context.Background())
			default:
				lock, err = manager.AcquireHomeOnly(context.Background(), false)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = lock.Close() }()
			held := lock.Held()
			if err := held.Validate(home); err != nil {
				t.Fatalf("live same-home witness: %v", err)
			}
			if !errors.Is(held.Validate(other), ErrNotHeld) {
				t.Fatal("wrong-home witness accepted")
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			if !errors.Is(held.Validate(home), ErrNotHeld) {
				t.Fatal("stale witness accepted")
			}
			replacement, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = replacement.Close() }()
			if err := replacement.Held().Validate(home); err != nil {
				t.Fatal(err)
			}
			if !errors.Is(held.Validate(home), ErrNotHeld) {
				t.Fatal("reacquisition revived stale witness")
			}
		})
	}
}

// The witness revalidates the stable lock name, rather than trusting a cached
// held flag. Removing that name cannot authorize a read beneath a new owner.
func TestHeldWitnessRejectsRemovedLock(t *testing.T) {
	manager, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if err := lock.Held().Validate(manager.Home()); err != nil {
		t.Fatal(err)
	}
	path := lock.lock.file.Name()
	if err := os.Remove(path); err != nil {
		if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(32)) {
			t.Skip("host refuses unlink of an open lock: sharing violation")
		}
		t.Fatal(err)
	}
	if err := lock.Held().Validate(manager.Home()); err == nil {
		t.Fatal("removed lock witness accepted")
	}
}
