//go:build unix

package privatedir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

func createPrivateTempFile(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func validatePrivateFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%s is not an owner-only regular file", path)
	}
	return nil
}

func protectPrivateFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}
	return os.Chmod(path, info.Mode().Perm()&0o700)
}

func protectPrivateTree(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if entry.IsDir() && info.IsDir() {
			return protectPrivate(path)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file or directory", path)
		}
		return protectPrivateFile(path)
	})
}
