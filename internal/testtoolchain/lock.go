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

// LockHostGOROOT serializes only the macOS tests that either keep a whole-tree
// GOROOT fingerprint across an operation or execute the real tools from that
// same host tree. Two hosted macOS runs observed the tool tree change while
// those exact test classes overlapped; the lock gives that shared external test
// fixture one owner without changing production toolchain verification or
// serializing unrelated tests.
func LockHostGOROOT(t testing.TB) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		return
	}

	manager, err := managerlock.New(filepath.Join(os.TempDir(), "curator-host-goroot-test-lock-v1"))
	if err != nil {
		t.Fatalf("create host GOROOT test lock: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), hostGOROOTLockTimeout)
	t.Cleanup(cancel)
	lock, err := manager.AcquireHomeOnly(ctx, false)
	if err != nil {
		t.Fatalf("acquire host GOROOT test lock: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := lock.Close(); closeErr != nil {
			t.Errorf("release host GOROOT test lock: %v", closeErr)
		}
	})
}
