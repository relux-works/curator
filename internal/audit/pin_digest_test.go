package audit

// N8 regression at the library boundary: Pin parses a supported content
// identity before any filesystem access and refuses everything else, so
// pin state stays inside the audit namespace.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPinRefusesNonDigest refuses path-shaped and malformed values without
// touching the filesystem, and still pins a supported digest through the
// same entry. A mutant that drops the Pin validation pins the traversal
// input and fails the refusal rows.
func TestPinRefusesNonDigest(t *testing.T) {
	for _, allow := range []string{
		"",
		"../escape",
		"sub/dir",
		strings.Repeat("a", 63),
		strings.Repeat("z", 64),
	} {
		home := t.TempDir()
		if _, err := Pin(home, allow, "synthetic approval", "fixture"); err == nil {
			t.Fatalf("Pin(%q) succeeded, want a refusal", allow)
		}
		if _, err := os.Stat(filepath.Join(home, "audit")); !os.IsNotExist(err) {
			t.Fatalf("refused Pin(%q) touched the audit namespace", allow)
		}
		if _, err := os.Stat(filepath.Join(home, "escape")); !os.IsNotExist(err) {
			t.Fatalf("refused Pin(%q) wrote outside the audit namespace", allow)
		}
	}
	digest := "sha256:" + strings.Repeat("c", 64)
	home := t.TempDir()
	path, err := Pin(home, digest, "synthetic approval", "fixture")
	if err != nil {
		t.Fatalf("Pin(%q): %v", digest, err)
	}
	want := filepath.Join(home, "audit", strings.Repeat("c", 64), "trust.json")
	if path != want {
		t.Fatalf("Pin(%q) = %s, want %s", digest, path, want)
	}
}
