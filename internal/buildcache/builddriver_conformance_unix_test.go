//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package buildcache

import (
	"testing"

	"golang.org/x/sys/unix"
)

// makeCacheSpecialFile creates a real non-regular member so the authoritative
// "artifact is not regular" vector runs against a genuine device-like entry.
func makeCacheSpecialFile(t *testing.T, path string) {
	t.Helper()
	if err := unix.Mkfifo(path, 0o600); err != nil {
		t.Skipf("this host cannot create the FIFO the vector needs: %v", err)
	}
}
