//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package skillspec

import (
	"testing"

	"golang.org/x/sys/unix"
)

// makeConformanceSpecialFile creates a real non-regular, non-directory member so
// the authoritative "special file" vectors run against a genuine device-like
// snapshot entry instead of a plain file stand-in.
func makeConformanceSpecialFile(t *testing.T, path string) {
	t.Helper()
	if err := unix.Mkfifo(path, 0o600); err != nil {
		t.Skipf("this host cannot create the FIFO the vector needs: %v", err)
	}
}
