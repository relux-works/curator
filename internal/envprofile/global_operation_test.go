package envprofile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Compensation restores bytes, modes, link text, removed backup trees and
// absence. It never follows a captured symlink into its unrelated target.
func TestGlobalProfileCompensationImages(t *testing.T) {
	root := t.TempDir()
	foreign := filepath.Join(t.TempDir(), "sentinel")
	if err := os.WriteFile(foreign, []byte("foreign bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "tree")
	if err := os.Mkdir(dir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("preimage"), 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	before, err := readGlobalImage(dir)
	if err != nil {
		t.Fatal(err)
	}
	op := &operation{undo: &globalUndo{entries: map[string]*globalImage{}, managerHome: root}}
	if err := op.capture(dir, filepath.Join(root, "new", "leaf")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "new"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "new", "leaf"), []byte("new bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := op.undo.restore(); err != nil {
		t.Fatal(err)
	}
	after, err := readGlobalImage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("compensation did not restore exact surface image")
	}
	if _, err := os.Lstat(filepath.Join(root, "new")); !os.IsNotExist(err) {
		t.Fatalf("compensation retained new parent: %v", err)
	}
	if data, err := os.ReadFile(foreign); err != nil || string(data) != "foreign bytes" {
		t.Fatalf("compensation followed a link: %q %v", data, err)
	}
	// Capturing an unchanged image is a no-op on restore; inode identity remains.
	info, err := os.Stat(filepath.Join(dir, "file"))
	if err != nil {
		t.Fatal(err)
	}
	if err := op.undo.restore(); err != nil {
		t.Fatal(err)
	}
	current, err := os.Stat(filepath.Join(dir, "file"))
	if err != nil || !os.SameFile(info, current) {
		t.Fatal("unchanged compensation rewrote its preimage")
	}
}

// A manager-owned parent symlink is refused before snapshotting descendants;
// an ordinary directory next to it remains a valid capture control.
func TestGlobalProfileCompensationParentLinkRefuses(t *testing.T) {
	root, foreign := t.TempDir(), t.TempDir()
	if err := os.Symlink(foreign, filepath.Join(root, "parent")); err != nil {
		t.Fatal(err)
	}
	op := &operation{undo: &globalUndo{entries: map[string]*globalImage{}, managerHome: root}}
	if err := op.capture(filepath.Join(root, "parent", "child")); err == nil {
		t.Fatal("snapshot followed a manager parent link")
	}
	if err := os.Mkdir(filepath.Join(root, "valid"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := op.capture(filepath.Join(root, "valid", "child")); err != nil {
		t.Fatal(err)
	}
}
