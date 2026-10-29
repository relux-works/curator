package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

// Missing-tail evidence permits a separately verified planned owner without
// creating the native parent. Existing case aliases and broken parent links
// stay unknown; this helper alone does not prove profile or package authority.
func TestPlannedOwnerMissingParentCaseGuard(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "native", "skills", "tool")
	if casingAliasUnmanaged(missing) {
		t.Fatal("proven missing parent discarded planned owner")
	}
	if _, err := os.Lstat(filepath.Dir(missing)); !os.IsNotExist(err) {
		t.Fatalf("read-only guard created parent: %v", err)
	}
	parent := filepath.Join(root, "existing")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "Tool"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if !casingAliasUnmanaged(filepath.Join(parent, "tool")) {
		t.Fatal("case alias blessed by planned owner")
	}
	broken := filepath.Join(root, "broken")
	if err := os.Symlink(filepath.Join(root, "absent"), broken); err != nil {
		t.Fatal(err)
	}
	if !casingAliasUnmanaged(filepath.Join(broken, "skills", "tool")) {
		t.Fatal("broken parent treated as absent")
	}
}
