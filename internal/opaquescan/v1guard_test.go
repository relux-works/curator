package opaquescan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

func writeTestFile(t *testing.T, root, rel string, payload []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

// The pre-hash guard refuses v1 roots with NUL bytes, carrying the shared
// finding id and the offending file; v2 roots skip the scan entirely and
// unknown versions refuse without scanning.
func TestRefuseNULV1(t *testing.T) {
	nulTree := t.TempDir()
	writeTestFile(t, nulTree, "assets/a.bin", []byte("x\x00y"))
	cleanTree := t.TempDir()
	writeTestFile(t, cleanTree, "assets/a.bin", []byte("x"))

	if err := RefuseNULV1(nulTree, hashing.VersionV1); err == nil ||
		!strings.Contains(err.Error(), FindingNUL) ||
		!strings.Contains(err.Error(), "assets/a.bin") {
		t.Fatalf("v1 NUL refusal = %v, want the finding id naming assets/a.bin", err)
	}
	if err := RefuseNULV1(nulTree, 0); err == nil ||
		!strings.Contains(err.Error(), "unsupported content hash version") {
		t.Fatalf("zero-version guard = %v, want an unsupported-version refusal", err)
	}
	if err := RefuseNULV1(nulTree, hashing.Version(3)); err == nil ||
		!strings.Contains(err.Error(), "unsupported content hash version 3") {
		t.Fatalf("unknown-version guard = %v, want an unsupported-version refusal", err)
	}
	if err := RefuseNULV1(cleanTree, hashing.VersionV1); err != nil {
		t.Fatalf("v1 clean guard = %v, want nil", err)
	}
	if err := RefuseNULV1(nulTree, hashing.VersionV2); err != nil {
		t.Fatalf("v2 NUL guard = %v, want nil (NUL is ordinary v2 data)", err)
	}
}

// The v1 guard scans nested files, not just the snapshot root: a mutant
// that narrows the scan to top-level entries must still trip on this tree.
func TestRefuseNULV1ScansNestedFiles(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/deep/opaque.unsupported", []byte("prefix\x00suffix"))
	if err := RefuseNULV1(root, hashing.VersionV1); err == nil ||
		!strings.Contains(err.Error(), "docs/deep/opaque.unsupported") {
		t.Fatalf("v1 deep-NUL refusal = %v, want the finding id naming the nested file", err)
	}
}
