package managerlock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

// The observation path never creates an absent home or missing lock state;
// the normal writer creates the same lock which observation subsequently uses.
func TestAcquireExistingHomeOnlyDoesNotProvision(t *testing.T) {
	home := filepath.Join(t.TempDir(), "absent")
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AcquireExistingHomeOnly(context.Background()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing lock = %v", err)
	}
	if _, err := os.Lstat(home); !os.IsNotExist(err) {
		t.Fatalf("created home: %v", err)
	}
	if err := os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AcquireExistingHomeOnly(context.Background()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing lock state = %v", err)
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("created lock state: %v, %v", entries, err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	lock, err = manager.AcquireExistingHomeOnly(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.AssertCurrent(); err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(lock.AssertCurrent(), ErrNotHeld) {
		t.Fatal("accepted released lock")
	}
}

// An externally replaced stable name invalidates the held capture witness;
// retaining an OS lock on an unlinked inode does not serialize later writers.
func TestCaptureLockRejectsReplacedIdentity(t *testing.T) {
	manager, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if err := lock.AssertCurrent(); err != nil {
		t.Fatal(err)
	}
	path := lock.lock.file.Name()
	if err := os.Rename(path, path+".old"); err != nil {
		if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(32)) { // ERROR_SHARING_VIOLATION
			t.Skip("this host cannot create replacement of an open capture lock: sharing violation")
		}
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(lock.AssertCurrent(), ErrNotHeld) {
		t.Fatal("accepted replacement lock identity")
	}
}

// Both acquisition paths contend on the same OS lock and honor the timeout.
func TestCaptureLockSharesWriterLock(t *testing.T) {
	manager, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := manager.AcquireExistingHomeOnly(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock wait = %v", err)
	}
}
