package marker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

// Proven absence is allowed; an untraversable ancestor or dangling ownership
// link is a refusal, even if the underlying OS also reports a missing leaf.
func TestCheckOwnershipDistinguishesAbsentFromUnaddressable(t *testing.T) {
	root := t.TempDir()
	if err := CheckOwnership(filepath.Join(root, "absent", "skill")); err != nil {
		t.Fatalf("absent: %v", err)
	}
	path := filepath.Join(root, "file")
	if err := os.WriteFile(path, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckOwnership(filepath.Join(path, "skill")); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("unaddressable: %v", err)
	}
	dir := filepath.Join(root, "skill")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing", filepath.Join(dir, Name)); err != nil {
		t.Fatal(err)
	}
	if err := CheckOwnership(dir); err == nil || !strings.Contains(err.Error(), dir) {
		t.Fatalf("dangling ownership: %v", err)
	}
}
