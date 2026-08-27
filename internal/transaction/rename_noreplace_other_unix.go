//go:build unix && !darwin && !linux

package transaction

import (
	"fmt"
	"os"
)

// Supported release platforms use a native no-replace rename. Other Unix
// ports retain a fail-before-overwrite fallback until their exclusive rename
// primitive is wired explicitly.
func durableRenameNoReplace(from, to string) error {
	if _, err := os.Lstat(to); err == nil {
		return fmt.Errorf("rename destination already exists: %s", to)
	} else if !os.IsNotExist(err) {
		return err
	}
	return durableRename(from, to)
}
