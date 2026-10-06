// Rework-1 regression for the Spec §8 version rule at the frozen package
// context hash: the frozen v1 identity is never computed over NUL bytes.
// Adapts the revision-1 reviewer's attack probe into maintained coverage.
package closure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/skillspec"
)

func writeFrozenTestFiles(t *testing.T, root string, files map[string][]byte) {
	t.Helper()
	for rel, payload := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// A frozen tree whose projected context carries NUL bytes is refused with
// the opaque finding before any v1 digest is computed: no digest is
// returned. The NUL-free control still hashes.
func TestContentHashForRefusesNULBeforeHashing(t *testing.T) {
	nulTree := t.TempDir()
	writeFrozenTestFiles(t, nulTree, map[string][]byte{
		"SKILL.md":     []byte("# Test\n"),
		"assets/a.bin": []byte("x\x00y"),
	})
	digest, err := ContentHashFor(nulTree, &skillspec.Spec{})
	if err == nil || !strings.Contains(err.Error(), opaquescan.FindingNUL) {
		t.Fatalf("frozen NUL context = (%q, %v), want the opaque refusal", digest, err)
	}
	if digest != "" {
		t.Fatalf("frozen NUL context returned digest %s alongside the refusal", digest)
	}

	cleanTree := t.TempDir()
	writeFrozenTestFiles(t, cleanTree, map[string][]byte{
		"SKILL.md":     []byte("# Test\n"),
		"assets/a.bin": []byte("x"),
	})
	if digest, err := ContentHashFor(cleanTree, &skillspec.Spec{}); err != nil || digest == "" {
		t.Fatalf("clean frozen context = (%q, %v), want a digest", digest, err)
	}
}
