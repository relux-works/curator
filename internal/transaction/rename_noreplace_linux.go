//go:build linux

package transaction

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

func durableRenameNoReplace(from, to string) error {
	if err := unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(to)); err != nil {
		return err
	}
	if filepath.Dir(from) != filepath.Dir(to) {
		return syncDirectory(filepath.Dir(from))
	}
	return nil
}
