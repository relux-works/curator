//go:build !windows

package snapshot

import (
	"fmt"
	"os"
	"syscall"
)

// captureFileToken pins the physical identity of one admitted file at
// inventory time: device and inode from a single Lstat, the pair
// os.SameFile compares on unix. The token is a plain string so the
// post-copy recheck compares the recorded pin against a fresh
// inspection instead of two lazily-resolved FileInfo values. Lstat
// (never Stat) keeps link semantics: a file replaced by a link carries
// the link's own identity and fails the recheck.
func captureFileToken(path string) (string, error) {
	info, err := os.Lstat(path) // #nosec G304 -- admitted member under capture
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return "", fmt.Errorf("cannot read file identity for %s", path)
	}
	return fmt.Sprintf("%d:%d", stat.Dev, stat.Ino), nil
}
