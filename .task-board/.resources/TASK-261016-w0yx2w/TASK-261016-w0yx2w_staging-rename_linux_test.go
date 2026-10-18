package stagingprobe

import "golang.org/x/sys/unix"

func noReplace(parent int, from, to string) error {
	return unix.Renameat2(parent, from, parent, to, unix.RENAME_NOREPLACE)
}
