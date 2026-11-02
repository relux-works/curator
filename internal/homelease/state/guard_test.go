package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// HL03/14: physical directory aliases protect even absent children, while an
// actual sibling remains disjoint. Case behavior follows this filesystem's
// SameFile result, rather than unconditional lowercasing of user identifiers.
func TestPhysicalOverlap(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "ManagedHome")
	if err := os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	upper := filepath.Join(root, strings.ToUpper(filepath.Base(home)))
	info, err := os.Stat(home)
	if err != nil {
		t.Fatal(err)
	}
	aliasInfo, aliasErr := os.Stat(upper)
	caseAlias := aliasErr == nil && os.SameFile(info, aliasInfo)
	for _, tc := range []struct {
		path string
		want bool
	}{
		{filepath.Join(home, "new-child"), true},
		{filepath.Join(upper, "new-child"), caseAlias},
		{filepath.Join(root, "sibling", "new-child"), false},
	} {
		got, err := Overlap(tc.path, home)
		if err != nil || got != tc.want {
			t.Fatalf("physical overlap %s = %v %v; want %v", filepath.Base(filepath.Dir(tc.path)), got, err, tc.want)
		}
	}
	link := filepath.Join(root, "alias")
	if err := os.Symlink(home, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if got, err := Overlap(filepath.Join(link, "new-child"), home); err != nil || !got {
		t.Fatalf("symlink child escaped protection: %v %v", got, err)
	}
	if _, err := Overlap(link+string(filepath.Separator)+".."+string(filepath.Separator)+"ManagedHome", home); err == nil {
		t.Fatal("unclean traversal accepted as identity")
	}
}
