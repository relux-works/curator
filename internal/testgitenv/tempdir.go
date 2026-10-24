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
// budget without adding the test/subtest name to their root. Hosted runners
// provide a short RUNNER_TEMP; other Windows runs can set a short TMP. Unix
// keeps testing.T.TempDir, and every Windows directory has the same private
// DACL required of other managed-home fixtures.
func ShortTempDir(t testing.TB) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		return t.TempDir()
	}
	base := os.Getenv("RUNNER_TEMP")
	if base == "" {
		base = os.TempDir()
	}
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
