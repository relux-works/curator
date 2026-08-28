//go:build unix

package privatedir

import (
	"fmt"
	"io/fs"
	"os"
)

// On Unix the private shape is the 0o700 permission bits on a real directory.
// 0o700 grants nothing to group or other, so no common umask weakens it.

func makePrivate(path string) error { return os.Mkdir(path, 0o700) }

func makeAllPrivate(path string) error { return os.MkdirAll(path, 0o700) }

func validatePrivate(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return fmt.Errorf("%s is not a private owner-only directory", path)
	}
	return nil
}

func protectPrivate(path string) error {
	return os.Chmod(path, 0o700) // #nosec G302 -- a private directory, not a regular file.
}
