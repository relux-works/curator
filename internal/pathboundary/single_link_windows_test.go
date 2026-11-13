//go:build windows

package pathboundary

import (
	"os"
	"path/filepath"
	"testing"
)

// Handle metadata must identify the same file observed by lstat. A replacement
// that is itself single-link cannot satisfy the earlier file's proof.
func TestSingleLinkWindowsBindsObservedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload")
	if err := os.WriteFile(path, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := singleLink(path, info); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, path+"-previous"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := singleLink(path, info); err == nil || err.Error() != "snapshot file identity changed during link metadata inspection" {
		t.Fatalf("replacement accepted or misclassified: %v", err)
	}
}
