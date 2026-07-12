package hashing

import (
	"os"
	"testing"
)

// TestCrossCheckTree hashes the tree given in CURATOR_CROSSCHECK_TREE and
// compares with CURATOR_CROSSCHECK_WANT. Skipped unless both are set; used
// by out-of-repo interoperability harnesses.
func TestCrossCheckTree(t *testing.T) {
	tree := os.Getenv("CURATOR_CROSSCHECK_TREE")
	want := os.Getenv("CURATOR_CROSSCHECK_WANT")
	if tree == "" || want == "" {
		t.Skip("cross-check env not set")
	}
	got, err := ContentSHA256(tree, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("hash = %s, want %s", got, want)
	}
}
