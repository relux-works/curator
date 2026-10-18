//go:build darwin

package envsnapshot

import "golang.org/x/sys/unix"

func publishNoReplace(parent int, stage, target string) error {
	return unix.RenameatxNp(parent, stage, parent, target, unix.RENAME_EXCL)
}
