//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris)

package buildsource

import "testing"

// makeBuildSourceSpecialFile reports that this host has no portable FIFO, so
// the special-file vector is not reachable here.
func makeBuildSourceSpecialFile(t *testing.T, _ string) bool {
	t.Helper()
	return false
}
