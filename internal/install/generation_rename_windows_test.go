//go:build windows

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

// renameReplacement stages payload beside path and replaces path with it as
// early as Windows permits.
//
// os.Open takes FILE_SHARE_READ and FILE_SHARE_WRITE and never
// FILE_SHARE_DELETE, so while readDocument holds the document open the kernel
// refuses to rename anything over it. The window still opens and the attempt is
// still made — the refusal is asserted, because it is a real property of the
// platform and the reason the POSIX schedule cannot be used here — and the
// replacement settles the moment the handle closes.
//
// The three properties the case asserts are unchanged by that ordering: the
// read returns the bytes of the instance it opened, the recorded generation
// belongs to those bytes, and a later path read reports a different generation
// and therefore restarts the run.
func renameReplacement(t *testing.T, dir, path, payload string) renameSchedule {
	t.Helper()
	staged := filepath.Join(dir, "staged.json")
	replace := func() error { return os.Rename(staged, path) }
	return renameSchedule{
		duringRead: func() {
			writeDocument(t, staged, payload)
			if err := replace(); err == nil {
				t.Fatal("an open declaration document was replaced by rename; " +
					"this platform is expected to hold it until the read closes")
			}
		},
		settle: func(t *testing.T) {
			t.Helper()
			if err := replace(); err != nil {
				t.Fatalf("the replacement did not settle once the read closed: %v", err)
			}
		},
	}
}
