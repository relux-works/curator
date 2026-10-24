package testgitenv

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Fixtures remain writable and cleanup removes a read-only object after the
// subtest. Windows additionally proves the protected private DACL; Unix retains
// the ordinary testing.T.TempDir mode and lifecycle.
func TestShortTempDirIsPrivateAndCleaned(t *testing.T) {
	var dir string
	t.Run("fixture", func(t *testing.T) {
		dir = ShortTempDir(t)
		if runtime.GOOS == "windows" {
			if err := privatedir.Validate(dir); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "object"), []byte("fixture"), 0o400); err != nil {
			t.Fatal(err)
		}
		if body, err := os.ReadFile(filepath.Join(dir, "object")); err != nil || string(body) != "fixture" {
			t.Fatalf("fixture=%q: %v", body, err)
		}
	})
	if _, err := os.Lstat(dir); !os.IsNotExist(err) {
		t.Fatalf("short fixture was not removed: %v", err)
	}
}
