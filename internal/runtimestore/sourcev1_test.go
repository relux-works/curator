package runtimestore

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceV1KeyShape(t *testing.T) {
	digest := "sha256:" + strings.Repeat("ab", 32)
	key, err := SourceV1Key(digest)
	if err != nil {
		t.Fatalf("SourceV1Key = %v", err)
	}
	if key != "source-v1-"+strings.Repeat("ab", 32) {
		t.Fatalf("key = %q", key)
	}
	// The namespace prefix keeps draft leaves disjoint from bare-commit
	// legacy leaves: no 40/64-hex commit equals a namespaced key.
	if key == strings.Repeat("ab", 32) {
		t.Fatalf("namespaced key collides with a bare commit leaf")
	}
}

func TestSourceV1KeyRefusesMalformedDigests(t *testing.T) {
	for _, digest := range []string{
		"",
		"sha256:xyz",
		strings.Repeat("ab", 32),
		"SHA256:" + strings.Repeat("ab", 32),
		"sha256:" + strings.Repeat("AB", 32),
		"sha256:" + strings.Repeat("ab", 20),
	} {
		if _, err := SourceV1Key(digest); err == nil {
			t.Fatalf("SourceV1Key(%q) admitted a malformed digest", digest)
		}
	}
}

func TestSourceV1DirLayout(t *testing.T) {
	digest := "sha256:" + strings.Repeat("cd", 32)
	dir, err := SourceV1Dir(filepath.Join(string(filepath.Separator), "home"), "review", digest)
	if err != nil {
		t.Fatalf("SourceV1Dir = %v", err)
	}
	want := filepath.Join(string(filepath.Separator), "home", "runtime", "review", "source-v1-"+strings.Repeat("cd", 32))
	if dir != want {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
	if _, err := SourceV1Dir(filepath.Join(string(filepath.Separator), "home"), "", digest); err == nil {
		t.Fatalf("SourceV1Dir admitted an empty skill name")
	}
	if _, err := SourceV1Dir(filepath.Join(string(filepath.Separator), "home"), "review", "bad"); err == nil {
		t.Fatalf("SourceV1Dir admitted a malformed digest")
	}
}
