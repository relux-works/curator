//go:build unix

package closureexec

import (
	"io/fs"
	"os"
)

// markTreeDirImmutable removes the write bits from one directory of an
// admitted or replay tree. On Unix that is the whole mechanism: a 0o500
// directory blocks accidental member creation and the validation below can
// read the property back from the permission bits.
func markTreeDirImmutable(path string) error {
	return os.Chmod(path, 0o500) // #nosec G302 -- immutable directories require owner execute permission for read-only traversal.
}

// treeDirIsImmutable reports whether a directory of an admitted tree still
// carries the immutable shape markTreeDirImmutable gave it.
func treeDirIsImmutable(info fs.FileInfo) bool {
	return info.Mode().Perm()&0o222 == 0
}

// restoreTreeDirMutable undoes markTreeDirImmutable for cleanup.
func restoreTreeDirMutable(path string) error {
	return os.Chmod(path, 0o700) // #nosec G302 -- cleanup must restore owner traversal before removal.
}
