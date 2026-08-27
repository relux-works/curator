//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris)

package skillspec

import (
	"os"
	"testing"
)

// makeConformanceSpecialFile falls back to a non-directory member on hosts
// without a portable FIFO. Curator rejects the vector at the same seam: the
// build root or source directory is not a directory.
func makeConformanceSpecialFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("not a directory\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
