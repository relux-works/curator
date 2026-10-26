package projectadapter

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Claim: equal content, mode, size and mtime on distinct inodes cannot substitute
// for the walked entry; a second stat of the original inode is a valid neighbour.
func TestMuseEntryIdentity(t *testing.T) {
	home := offlineHome(t)
	parent := filepath.Dir(home)
	a, b := filepath.Join(parent, "a"), filepath.Join(parent, "b")
	stamp := time.Unix(1700000000, 0)
	for _, name := range []string{a, b} {
		if err := os.WriteFile(name, []byte("same"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(name, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	before, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	same, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	other, err := os.Stat(b)
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode() != other.Mode() || before.Size() != other.Size() || !before.ModTime().Equal(other.ModTime()) || os.SameFile(before, other) {
		t.Fatal("identity fixture is not isolated")
	}
	if !museEntryUnchanged(before, same) {
		t.Fatal("unchanged inode refused")
	}
	if museEntryUnchanged(before, other) {
		t.Fatal("distinct inode accepted")
	}
}
