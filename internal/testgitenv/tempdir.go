package testgitenv

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// ShortTempDir keeps Windows Git cache fixtures inside the qualified metadata
// budget without adding the test/subtest name to their root. Windows uses a
// private directory at the ordinary temporary volume's root: RUNNER_TEMP can
// be on another drive, breaking fixture renames and relative links. Unix keeps
// testing.T.TempDir. Every Windows directory has the private DACL required of
// other managed-home fixtures before any fixture data is written.
func ShortTempDir(t testing.TB) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		return t.TempDir()
	}
	ordinary, err := filepath.Abs(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.VolumeName(ordinary) + string(os.PathSeparator)
	dir, err := os.MkdirTemp(base, "c-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Read-only Git objects may need their write attribute restored.
		// WalkDir does not follow symlinks, including fixture link targets.
		_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.Type()&(os.ModeSymlink|os.ModeIrregular) != 0 {
				return nil
			}
			info, err := entry.Info()
			if err == nil {
				_ = os.Chmod(path, info.Mode()|0o200)
			}
			return nil
		})
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove short Git fixture: %v", err)
		}
	})
	if err := privatedir.Protect(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}
