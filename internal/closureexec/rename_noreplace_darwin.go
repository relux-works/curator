//go:build darwin

package closureexec

import "golang.org/x/sys/unix"

func renameTreeNoReplace(source, target string) error {
	return unix.RenameatxNp(unix.AT_FDCWD, source, unix.AT_FDCWD, target, unix.RENAME_EXCL)
}
