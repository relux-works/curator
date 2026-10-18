package yarnmodernsource

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Nested sealed fixture trees are owner-writable before removal, preserving executability and external symlink targets.
func TestWritableTempDirRestoresNestedTrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission fixture is exercised on Linux and macOS")
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0o400); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(outside, 0o600) })
	t.Run("sealed", func(t *testing.T) {
		_ = t.TempDir() // Register TempDir removal before the audit.
		var root string
		t.Cleanup(func() {
			for _, path := range []string{root, filepath.Join(root, "nested"), filepath.Join(root, "nested", "executable")} {
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm()&0o200 == 0 {
					t.Errorf("tree not restored: %s: %v", path, err)
				}
			}
			info, err := os.Stat(filepath.Join(root, "nested", "executable"))
			if err != nil || info.Mode().Perm()&0o100 == 0 {
				t.Error("executable bit lost")
			}
		})
		root = writableTempDir(t)
		dir := filepath.Join(root, "nested")
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "executable"), []byte("x"), 0o500); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
	})
	if info, err := os.Stat(outside); err != nil || info.Mode().Perm()&0o200 != 0 {
		t.Errorf("outside target changed: %v", err)
	}
}
