//go:build unix

package closureexec

import (
	"errors"
	"io/fs"
	"os"
)

// publishByLink makes the written temporary file immutable and then installs
// it under target with a no-replace hard link. On Unix the mode travels with
// the inode, so the member is never observable in a mutable state, and the
// caller's later removal of the temporary name only drops a directory entry.
func publishByLink(tmpName, target string, mode fs.FileMode) error {
	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}
	if err := os.Link(tmpName, target); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	return nil
}
