package pathboundary

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateTrustedDirectoryTree(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.MkdirAll(filepath.Join(root, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "context", "a.md"), []byte("module\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	if err := Validate(root); err != nil {
		t.Fatalf("trusted directory rejected: %v", err)
	}
}

func TestValidateRootIgnoresIndependentChildRoots(t *testing.T) {
	root := filepath.Join(t.TempDir(), "imports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "untrusted-sibling"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRoot(root); err != nil {
		t.Fatalf("trusted parent rejected because of an independent child: %v", err)
	}
}

func TestValidateRejectsGroupOrWorldWritableComponents(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	restore, err := makeWorldWritableDirectoryForTest(root)
	if err != nil {
		t.Fatalf("create world-writable mutation boundary: %v", err)
	}
	defer func() {
		if err := restore(); err != nil {
			t.Errorf("restore permission fixture: %v", err)
		}
	}()
	err = Validate(root)
	if err == nil {
		t.Fatal("Validate admitted a world-writable mutation boundary")
	}
	failure, ok := err.(*Failure)
	if !ok || failure.Check != CheckPermissions {
		t.Fatalf("Validate error = %v, want permissions failure", err)
	}
}

func TestValidateRejectsInternalSymlink(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.MkdirAll(filepath.Join(root, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "inside.md"), []byte("module\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "inside.md"), filepath.Join(root, "context", "link.md")); err != nil {
		t.Skipf("host cannot create symbolic links: %v", err)
	}
	err := Validate(root)
	failure, ok := err.(*Failure)
	if !ok || failure.Check != CheckLinkSafety {
		t.Fatalf("Validate error = %v, want link-safety failure", err)
	}
}

func TestValidateRejectsEscapingSymlinkAsContainmentFailure(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape.md")); err != nil {
		t.Skipf("host cannot create symbolic links: %v", err)
	}
	err := Validate(root)
	failure, ok := err.(*Failure)
	if !ok || failure.Check != CheckContainment {
		t.Fatalf("Validate error = %v, want containment failure", err)
	}
}

func TestValidateRejectsMissingDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "absent")
	err := Validate(root)
	if !IsAbsent(err) {
		t.Fatalf("Validate error = %v, want an absent path failure", err)
	}
}

func TestValidateRejectsInjectedForeignOwner(t *testing.T) {
	root := filepath.Join(t.TempDir(), "foreign-owner")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	lookup := DefaultOwnerLookup()
	injected := func(path string, info os.FileInfo) (OwnerIdentity, error) {
		if filepath.Clean(path) == filepath.Clean(root) {
			return "injected:foreign-owner", nil
		}
		return lookup(path, info)
	}
	err := ValidateWithOwner(root, injected)
	failure, ok := err.(*Failure)
	if !ok || failure.Check != CheckOwnership {
		t.Fatalf("ValidateWithOwner error = %v, want ownership failure", err)
	}
}
