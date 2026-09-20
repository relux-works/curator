//go:build unix

package pnpmsource

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCopyContainedTreeDereferencesLinks pins the POSIX contract the Windows
// junction normalization must preserve: links inside the tree are copied by
// content, while escapes and cycles still fail closed. os.Symlink needs a
// privilege Windows runners may not grant, so the junction spelling of the
// same contract lives in linkentry_windows_test.go (mklink /J).
func TestCopyContainedTreeDereferencesLinks(t *testing.T) {
	t.Run("nested link content is copied without links", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "pkg", "index.js"), []byte("module.exports = 1\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, "node_modules"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(root, "pkg"), filepath.Join(root, "node_modules", "pkg")); err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(t.TempDir(), "copied")
		if err := copyContainedTree(root, destination); err != nil {
			t.Fatal(err)
		}
		payload, err := os.ReadFile(filepath.Join(destination, "node_modules", "pkg", "index.js"))
		if err != nil || string(payload) != "module.exports = 1\n" {
			t.Fatalf("through-link content was not copied: %q %v", payload, err)
		}
		info, err := os.Lstat(filepath.Join(destination, "node_modules", "pkg"))
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("through-link copy left a link: %+v %v", info, err)
		}
	})
	t.Run("escape through a link is refused", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("outside\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
			t.Fatal(err)
		}
		assertCode(t, copyContainedTree(root, filepath.Join(t.TempDir(), "copied")), CodeLocalPathEscape)
	})
	t.Run("link cycle is refused", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "dir"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(root, filepath.Join(root, "dir", "cycle")); err != nil {
			t.Fatal(err)
		}
		assertCode(t, copyContainedTree(root, filepath.Join(t.TempDir(), "copied")), CodeInputUndeclared)
	})
}
