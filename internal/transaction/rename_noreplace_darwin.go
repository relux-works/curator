//go:build darwin

package transaction

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

func durableRenameNoReplace(from, to string) error {
	if err := unix.RenamexNp(from, to, unix.RENAME_EXCL); err != nil {
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
