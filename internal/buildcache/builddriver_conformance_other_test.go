//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris)

package buildcache

import (
	"os"
	"testing"
)

// makeCacheSpecialFile falls back to a directory on hosts without a portable
// FIFO. Curator rejects it at the same seam: the artifact member is not the
// manager-derived regular file.
func makeCacheSpecialFile(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
}
