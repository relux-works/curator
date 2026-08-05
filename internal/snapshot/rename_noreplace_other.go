//go:build !unix && !windows

package snapshot

import (
	"fmt"
	"os"
)

func renameNoReplace(from, to string) error {
	if _, err := os.Lstat(to); err == nil {
		return fmt.Errorf("rename destination already exists: %s", to)
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Rename(from, to)
}
