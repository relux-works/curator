//go:build !darwin && !linux && !windows

package closureexec

import (
	"io/fs"
	"os"
)

// Supported release targets provide an atomic no-replace primitive above.
// Keep other ports fail-closed when a target is already observable.
func renameTreeNoReplace(source, target string) error {
	if _, err := os.Lstat(target); err == nil {
		return fs.ErrExist
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Rename(source, target)
}
