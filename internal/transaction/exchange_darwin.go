//go:build darwin

package transaction

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

const exchangeSupported = true

func durableExchange(a, b string) error {
	if err := unix.RenamexNp(a, b, unix.RENAME_SWAP); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(b)) // transaction sidecars share one parent
}
