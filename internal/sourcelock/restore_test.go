package sourcelock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRestoreReturnsExactPriorBytes pins the rollback half of lock
// publication: the staged bytes land verbatim with no temp files left
// behind.
func TestRestoreReturnsExactPriorBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Skillfile.lock.json")
	if err := os.WriteFile(path, []byte("published"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Restore(path, []byte("prior")); err != nil {
		t.Fatalf("Restore = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "prior" {
		t.Fatalf("restored = %q, want %q", got, "prior")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".lock-") {
			t.Fatalf("Restore left a temp file behind: %s", entry.Name())
		}
	}
}
