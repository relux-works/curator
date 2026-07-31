//go:build !windows

package install

import (
	"os"
	"path/filepath"
	"testing"
)

// renameSchedule is when an atomic replacement of an open document can land.
type renameSchedule struct {
	// duringRead runs inside the read window.
	duringRead func()
	// settle runs once the read has returned and leaves the path replaced on
	// platforms whose kernel refused the replacement while the file was open.
	settle func(t *testing.T)
}

// renameReplacement stages payload beside path and replaces path with it inside
// the read window. POSIX renames over an open file freely: the reader keeps the
// old inode and the directory entry already names the new one.
func renameReplacement(t *testing.T, dir, path, payload string) renameSchedule {
	t.Helper()
	return renameSchedule{
		duringRead: func() {
			staged := filepath.Join(dir, "staged.json")
			writeDocument(t, staged, payload)
			if err := os.Rename(staged, path); err != nil {
				t.Fatal(err)
			}
		},
		settle: func(*testing.T) {},
	}
}
