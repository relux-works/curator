// Tests for the v1 hash-observation seam: CountV1Hashes reports every v1
// content-hash computation and no v2 computation, so guarded production
// entries can prove refusal-before-hashing by observation rather than by
// the shape of their returned digest.
package hashing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountV1HashesObservesOnlyV1Computations(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "a.bin"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"assets/a.bin": []byte("x")}

	if calls := CountV1Hashes(func() {
		if _, err := ContentSHA256(root, nil); err != nil {
			t.Fatal(err)
		}
	}); calls != 1 {
		t.Fatalf("v1 tree hash observations = %d, want 1", calls)
	}
	if calls := CountV1Hashes(func() {
		if _, err := ContentSHA256WithVersion(root, nil, VersionV1); err != nil {
			t.Fatal(err)
		}
	}); calls != 1 {
		t.Fatalf("dispatched v1 tree hash observations = %d, want exactly 1 (no double count)", calls)
	}
	if calls := CountV1Hashes(func() {
		if _, err := ContentIdentity(root, nil, VersionV1); err != nil {
			t.Fatal(err)
		}
	}); calls != 1 {
		t.Fatalf("v1 identity observations = %d, want exactly 1 (no double count)", calls)
	}
	if calls := CountV1Hashes(func() {
		if _, err := ContentSHA256Files(files, VersionV1); err != nil {
			t.Fatal(err)
		}
	}); calls != 1 {
		t.Fatalf("v1 file-set hash observations = %d, want 1", calls)
	}
	if calls := CountV1Hashes(func() {
		if _, err := ContentSHA256WithVersion(root, nil, VersionV2); err != nil {
			t.Fatal(err)
		}
		if _, err := ContentSHA256Files(files, VersionV2); err != nil {
			t.Fatal(err)
		}
	}); calls != 0 {
		t.Fatalf("v2 hash observations = %d, want 0 (v2 hashes NUL as ordinary data)", calls)
	}
	if calls := CountV1Hashes(func() {
		if _, err := ContentSHA256WithVersion(root, nil, Version(9)); err == nil {
			t.Fatal("unknown version must fail without hashing")
		}
	}); calls != 0 {
		t.Fatalf("failed-dispatch observations = %d, want 0", calls)
	}
	// Observation stops when the call returns: later hashes are silent.
	if _, err := ContentSHA256(root, nil); err != nil {
		t.Fatal(err)
	}
	if calls := CountV1Hashes(func() {}); calls != 0 {
		t.Fatalf("empty observation = %d, want 0", calls)
	}
}
