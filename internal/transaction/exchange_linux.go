//go:build linux

package transaction

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

const exchangeSupported = true

func durableExchange(a, b string) error {
	if err := unix.Renameat2(unix.AT_FDCWD, a, unix.AT_FDCWD, b, unix.RENAME_EXCHANGE); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(b)) // transaction sidecars share one parent
}
