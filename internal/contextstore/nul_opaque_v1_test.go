// Rework-1 regression for the Spec §8 version rule at the store content
// hash: under the v1 writer a NUL-bearing entry is refused before any v1
// identity is computed; under v2, NUL is ordinary data.
package contextstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
)

func TestContentHashFollowsWriterVersionNULRule(t *testing.T) {
	nulTree := t.TempDir()
	path := filepath.Join(nulTree, "assets", "a.bin")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x\x00y"), 0o644); err != nil {
		t.Fatal(err)
	}

	prior := hashing.EnableV2Writers
	t.Cleanup(func() { hashing.EnableV2Writers = prior })

	hashing.EnableV2Writers = false
	if digest, err := ContentHash(nulTree); err == nil || !strings.Contains(err.Error(), opaquescan.FindingNUL) {
		t.Fatalf("v1 store hash over NUL = (%q, %v), want the opaque refusal", digest, err)
	}

	hashing.EnableV2Writers = true
	want, err := hashing.ContentSHA256WithVersion(nulTree, map[string]bool{}, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if digest, err := ContentHash(nulTree); err != nil || digest != want {
		t.Fatalf("v2 store hash over NUL = (%q, %v), want %s", digest, err, want)
	}
}

// Rework-2 regression for revision-3 finding F3 at the store content
// hash: under the v1 writer the opaque refusal is observed to precede
// hashing. The clean control hashes, proving the seam is wired.
func TestContentHashV1NULRefusalComputesNoV1Identity(t *testing.T) {
	nulTree := t.TempDir()
	nulPath := filepath.Join(nulTree, "assets", "a.bin")
	if err := os.MkdirAll(filepath.Dir(nulPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nulPath, []byte("x\x00y"), 0o644); err != nil {
		t.Fatal(err)
	}
	cleanTree := t.TempDir()
	cleanPath := filepath.Join(cleanTree, "assets", "a.bin")
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cleanPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	prior := hashing.EnableV2Writers
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
	hashing.EnableV2Writers = false

	var digest string
	var err error
	if calls := hashing.CountV1Hashes(func() {
		digest, err = ContentHash(nulTree)
	}); calls != 0 {
		t.Fatalf("v1 store hash over NUL computed %d v1 identities, want 0", calls)
	}
	if err == nil || !strings.Contains(err.Error(), opaquescan.FindingNUL) {
		t.Fatalf("v1 store hash over NUL = (%q, %v), want the opaque refusal", digest, err)
	}
	if calls := hashing.CountV1Hashes(func() {
		if _, err := ContentHash(cleanTree); err != nil {
			t.Fatal(err)
		}
	}); calls == 0 {
		t.Fatal("clean v1 store hash observed no v1 hash; the seam is not wired")
	}
}
