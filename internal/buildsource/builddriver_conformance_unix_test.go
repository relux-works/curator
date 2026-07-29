//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package buildsource

import (
	"testing"

	"golang.org/x/sys/unix"
)

// makeBuildSourceSpecialFile creates a real non-regular snapshot member.
func makeBuildSourceSpecialFile(t *testing.T, path string) bool {
	t.Helper()
	if err := unix.Mkfifo(path, 0o600); err != nil {
		t.Logf("mkfifo: %v", err)
		return false
	}
	return true
}
