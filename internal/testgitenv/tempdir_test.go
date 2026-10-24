package testgitenv

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Fixtures remain writable and cleanup removes a read-only object after the
// subtest. Windows additionally proves the protected private DACL and preserves
// ordinary temporary-fixture relative links and same-volume renames; Unix
// retains the ordinary testing.T.TempDir mode and lifecycle.
func TestShortTempDirIsPrivateAndCleaned(t *testing.T) {
	var dir string
	t.Run("fixture", func(t *testing.T) {
		dir = ShortTempDir(t)
		if runtime.GOOS == "windows" {
			if err := privatedir.Validate(dir); err != nil {
				t.Fatal(err)
			}
			ordinary := t.TempDir()
			if _, err := filepath.Rel(ordinary, dir); err != nil {
				t.Fatalf("short fixture crossed the ordinary temporary volume: %v", err)
			}
			from := filepath.Join(ordinary, "rename-control")
			to := filepath.Join(dir, "rename-control")
			if err := os.WriteFile(from, []byte("same-volume"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(from, to); err != nil {
				t.Fatalf("ordinary-to-short fixture rename: %v", err)
			}
			if body, err := os.ReadFile(to); err != nil || string(body) != "same-volume" {
				t.Fatalf("renamed fixture=%q: %v", body, err)
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
