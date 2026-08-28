// Package testtoolchain coordinates tests that hold a whole-tree fingerprint
// of the host Go installation with tests that execute tools from that tree.
package testtoolchain

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
)

const hostGOROOTLockTimeout = 5 * time.Minute

// AcquireHostGOROOT acquires the cross-process host-toolchain lock for a whole
// test process. A package whose tests may safely share the host toolchain with
// each other can hold this once in TestMain, preserving intra-package
// parallelism while remaining isolated from other package processes.
func AcquireHostGOROOT(ctx context.Context) (*managerlock.HomeLock, error) {
	if runtime.GOOS != "darwin" {
		return nil, nil
	}
	manager, err := managerlock.New(filepath.Join(os.TempDir(), "curator-host-goroot-test-lock-v1"))
	if err != nil {
		return nil, err
	}
	return manager.AcquireHomeOnly(ctx, false)
}

// LockHostGOROOT serializes only the macOS tests that either keep a whole-tree
// GOROOT fingerprint across an operation or execute the real tools from that
// same host tree. Two hosted macOS runs observed the tool tree change while
// those exact test classes overlapped; the lock gives that shared external test
// fixture one owner without changing production toolchain verification or
// serializing unrelated tests.
func LockHostGOROOT(t testing.TB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), hostGOROOTLockTimeout)
	t.Cleanup(cancel)
	lock, err := AcquireHostGOROOT(ctx)
	if err != nil {
		t.Fatalf("acquire host GOROOT test lock: %v", err)
	}
	if lock == nil {
		return
	}
	t.Cleanup(func() {
		if closeErr := lock.Close(); closeErr != nil {
			t.Errorf("release host GOROOT test lock: %v", closeErr)
		}
	})
}
