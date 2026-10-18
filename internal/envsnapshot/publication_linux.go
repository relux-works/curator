//go:build linux

package envsnapshot

import "golang.org/x/sys/unix"

func publishNoReplace(parent int, stage, target string) error {
	return unix.Renameat2(parent, stage, parent, target, unix.RENAME_NOREPLACE)
}
