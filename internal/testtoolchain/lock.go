// Package testtoolchain coordinates tests that hold a whole-tree fingerprint
// of the host Go installation with tests that execute tools from that tree.
package testtoolchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
)

// This bounds each test's queue, not a legitimate package-wide hold: observed
// macOS race runs can exhaust it while cmd/curator is still running tests.
// Keep the refusal until acquire/release evidence supports a scope/budget change.
const hostGOROOTLockTimeout = 20 * time.Minute

type hostGOROOTOwner struct {
	Name       string    `json:"name"`
	PID        int       `json:"pid"`
	Process    string    `json:"process"`
	AcquiredAt time.Time `json:"acquired_at"`
}

// HostGOROOTLock adds diagnostic metadata to the unchanged manager-home lock.
// Metadata never authorizes access; only the operating-system lock does.
type HostGOROOTLock struct {
	lock  *managerlock.HomeLock
	owner hostGOROOTOwner
	path  string
	logf  func(string, ...any)
	once  sync.Once
	err   error
}

type hostGOROOTWaitError struct {
	cause  error
	waiter string
	waited time.Duration
	holder *hostGOROOTOwner
	held   time.Duration
}

func (err *hostGOROOTWaitError) Unwrap() error { return err.cause }

func (err *hostGOROOTWaitError) Error() string {
	detail := "holder=unknown (metadata unavailable)"
	if err.holder != nil {
		detail = fmt.Sprintf("holder=%q pid=%d process=%q held=%s (unverified record; current holder may differ)",
			err.holder.Name, err.holder.PID, err.holder.Process, err.held.Round(time.Millisecond))
	}
	return fmt.Sprintf("%v; waiter=%q waited=%s; %s", err.cause, err.waiter, err.waited.Round(time.Millisecond), detail)
}

// AcquireHostGOROOT acquires the cross-process macOS host-toolchain lock for a
// named test process. The caller retains its previous timeout and lock scope.
func AcquireHostGOROOT(ctx context.Context, owner string) (*HostGOROOTLock, error) {
	if runtime.GOOS != "darwin" {
		return nil, nil
	}
	return acquireHostGOROOT(ctx, filepath.Join(os.TempDir(), "curator-host-goroot-test-lock-v1"), owner,
		func(format string, args ...any) { _, _ = fmt.Fprintf(os.Stderr, format+"\n", args...) })
}

func acquireHostGOROOT(ctx context.Context, root, owner string, logf func(string, ...any)) (*HostGOROOTLock, error) {
	started := time.Now()
	logf("host GOROOT test lock waiting: owner=%q pid=%d", owner, os.Getpid())
	manager, err := managerlock.New(root)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(root, "holder.json")
	lock, err := manager.AcquireHomeOnly(ctx, false)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			// An owner may release or change while this snapshot is read. Missing
			// or invalid metadata is unknown, never evidence of an unlocked tree.
			holder := readHostGOROOTOwner(path)
			failure := &hostGOROOTWaitError{cause: err, waiter: owner, waited: time.Since(started), holder: holder}
			if holder != nil {
				failure.held = time.Since(holder.AcquiredAt)
			}
			return nil, failure
		}
		return nil, err
	}
	record := hostGOROOTOwner{Name: owner, PID: os.Getpid(), Process: filepath.Base(os.Args[0]), AcquiredAt: time.Now().UTC()}
	if err := writeHostGOROOTOwner(path, record); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "warning: publish host GOROOT holder record failed; continuing with OS lock")
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			_, _ = fmt.Fprintln(os.Stderr, "warning: remove old host GOROOT holder record failed; record may be stale")
		}
	}
	logf("host GOROOT test lock acquired: owner=%q pid=%d waited=%s", owner, record.PID, time.Since(started).Round(time.Millisecond))
	return &HostGOROOTLock{lock: lock, owner: record, path: path, logf: logf}, nil
}

func writeHostGOROOTOwner(path string, record hostGOROOTOwner) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".holder-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func readHostGOROOTOwner(path string) *hostGOROOTOwner {
	file, err := openHostGOROOTRecord(path)
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	data, err := readHostGOROOTRecord(file)
	var owner hostGOROOTOwner
	if err != nil || len(data) > 4096 || json.Unmarshal(data, &owner) != nil || owner.Name == "" || owner.PID <= 0 || owner.Process == "" || owner.AcquiredAt.IsZero() || owner.AcquiredAt.After(time.Now()) {
		return nil
	}
	return &owner
}

// Close removes the diagnostic record while still owning the OS lock, then
// releases that lock. The stable manager lock file is never removed.
func (lock *HostGOROOTLock) Close() error {
	if lock == nil {
		return nil
	}
	lock.once.Do(func() {
		if err := os.Remove(lock.path); err != nil && !errors.Is(err, os.ErrNotExist) {
			_, _ = fmt.Fprintln(os.Stderr, "warning: remove host GOROOT holder record failed; record may be stale")
		}
		held := time.Since(lock.owner.AcquiredAt)
		lock.err = lock.lock.Close()
		lock.logf("host GOROOT test lock released: owner=%q pid=%d held=%s", lock.owner.Name, lock.owner.PID, held.Round(time.Millisecond))
	})
	return lock.err
}

// LockHostGOROOT serializes only the macOS tests that either keep a whole-tree
// GOROOT fingerprint across an operation or execute the real tools from that
// same host tree. It retains ownership through all child tests and cleanup.
func LockHostGOROOT(t testing.TB) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), hostGOROOTLockTimeout)
	t.Cleanup(cancel)
	lock, err := acquireHostGOROOT(ctx, filepath.Join(os.TempDir(), "curator-host-goroot-test-lock-v1"), t.Name(), t.Logf)
	if err != nil {
		t.Fatalf("acquire host GOROOT test lock: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := lock.Close(); closeErr != nil {
			t.Errorf("release host GOROOT test lock: %v", closeErr)
		}
	})
}

func readHostGOROOTRecord(file io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(file, 4097))
}
