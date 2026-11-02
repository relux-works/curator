package envunit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Production metadata wrappers prove private directories/files in the target's
// own protection model. This test run qualifies only its actual OS target.
func TestPrivateMetadataImages(t *testing.T) {
	root := filepath.Join(t.TempDir(), "private")
	if err := privatedir.Make(root); err != nil {
		t.Fatal(err)
	}
	dir, err := ObserveDirectory(root)
	if err != nil || dir.Kind != "directory" || dir.Mode != 0700 || !ValidProtection(dir.Protection) {
		t.Fatalf("private directory metadata: %+v %v", dir, err)
	}
	path := filepath.Join(root, "record")
	f, err := privatedir.Create(path, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString("noncredential record"); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	file, err := ObserveEntry(path, false)
	if err != nil || file.Kind != "file" || file.Mode != 0600 || !ValidProtection(file.Protection) {
		t.Fatalf("private file metadata: %+v %v", file, err)
	}
	before, err := os.Lstat(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = ApplyDirectoryProtection(root, dir); err != nil {
		t.Fatal(err)
	}
	after, err := ObserveDirectory(root)
	if err != nil || after != dir {
		t.Fatalf("metadata no-op changed pinned directory: %+v %v", after, err)
	}
	current, err := os.Lstat(root)
	if err != nil || !os.SameFile(before, current) {
		t.Fatal("metadata no-op replaced directory")
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "noncredential record" {
		t.Fatalf("metadata touched child payload: %q %v", got, err)
	}
}
