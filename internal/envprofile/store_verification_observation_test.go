package envprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Observations bind identity before replacement, even with identical bytes,
// mode and mtime. An untouched file and the retained original are neighbors.
func TestStoreVerificationObservationBindsReplacement(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "entry")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	before := storeVerificationNodes(t, path)
	if !sameStoreVerificationNodes(before, storeVerificationNodes(t, path)) {
		t.Fatal("intact observation neighbor refused")
	}
	info := before[path].info
	saved := filepath.Join(root, "saved")
	if err := os.Rename(path, saved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, info.Mode().Perm()); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	after := storeVerificationNodes(t, path)
	if sameStoreVerificationNodes(before, after) {
		t.Fatal("replacement borrowed earlier store identity")
	}
	retained, err := storeVerificationFileInfo(saved)
	if err != nil || !os.SameFile(info, retained) {
		t.Fatalf("retained original identity refused: %v", err)
	}
	t.Logf("store identity observation on %s: replacement distinct, retained original equal", runtime.GOOS)
}
