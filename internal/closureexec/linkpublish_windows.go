//go:build windows

package closureexec

import (
	"errors"
	"io/fs"
	"os"
)

// publishByLink installs the written temporary file under target with a
// no-replace hard link and only then makes the member immutable.
//
// The order is deliberately different from Unix. Windows hard links share the
// read-only attribute, and os.Remove retries an access-denied delete by
// clearing that attribute first — so removing a temporary name whose inode was
// already made read-only silently strips the protection from the published
// member it still shares. Deferring the attribute until the temporary link is
// the caller's only remaining concern makes its removal an ordinary delete,
// and the member becomes immutable before publication is acknowledged. The
// no-replace property is unaffected: an existing target still returns
// fs.ErrExist untouched.
func publishByLink(tmpName, target string, mode fs.FileMode) error {
	if err := os.Link(tmpName, target); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}
		return err
	}
	if err := os.Remove(tmpName); err != nil {
		return err
	}
	return os.Chmod(target, mode)
}
