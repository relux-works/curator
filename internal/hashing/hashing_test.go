package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestByteLayout locks the exact byte layout of Spec 8.5:
// record_i = relpath NUL content, records joined with NUL.
func TestByteLayout(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "b.txt", "beta")
	write(t, dir, "a/x.txt", "alpha")

	// expected: "a/x.txt\0alpha" + "\0" + "b.txt\0beta"
	payload := []byte("a/x.txt\x00alpha\x00b.txt\x00beta")
	sum := sha256.Sum256(payload)
	want := "sha256:" + hex.EncodeToString(sum[:])

	got, err := ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("hash = %s, want %s", got, want)
	}
}

func TestMarkerExcluded(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.txt", "content")
	before, _ := ContentSHA256(dir, nil)
	write(t, dir, MarkerName, `{"anything": true}`)
	after, err := ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("marker file must not affect the hash")
	}
}

func TestEmptyTree(t *testing.T) {
	got, err := ContentSHA256(t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(nil)
	if got != "sha256:"+hex.EncodeToString(sum[:]) {
		t.Fatalf("empty tree hash = %s", got)
	}
}

func TestOrderIndependence(t *testing.T) {
	// Two trees with identical content written in different order hash equal.
	left := t.TempDir()
	write(t, left, "z.txt", "1")
	write(t, left, "a.txt", "2")
	right := t.TempDir()
	write(t, right, "a.txt", "2")
	write(t, right, "z.txt", "1")
	lh, _ := ContentSHA256(left, nil)
	rh, _ := ContentSHA256(right, nil)
	if lh != rh {
		t.Fatalf("order dependence: %s vs %s", lh, rh)
	}
}

func TestNormalize(t *testing.T) {
	if Normalize("SHA256:ABCdef") == Normalize("sha256:abcdee") {
		t.Fatal("distinct hashes must not normalize equal")
	}
	if Normalize("sha256:AbC") != "abc" || Normalize(" aBc ") != "abc" {
		t.Fatal("normalize failed")
	}
}
