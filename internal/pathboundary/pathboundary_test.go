package pathboundary

import (
	"errors"
	"io/fs"
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

func TestValidateSkipsEntryRemovedBetweenReadDirAndLstat(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	removed := filepath.Join(root, "removed.md")
	if err := os.WriteFile(removed, []byte("module\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	removedDirectory := filepath.Join(root, "removed-directory")
	if err := os.Mkdir(removedDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}

	lookedUp := map[string]bool{}
	err := validateWithOwnerAndEntryInfo(root, DefaultOwnerLookup(), func(path string, entry fs.DirEntry) (os.FileInfo, error) {
		if filepath.Clean(path) == filepath.Clean(removed) || filepath.Clean(path) == filepath.Clean(removedDirectory) {
			lookedUp[filepath.Clean(path)] = true
			if err := os.Remove(path); err != nil {
				t.Fatalf("remove entry after readdir: %v", err)
			}
		}
		return entry.Info()
	})
	for _, path := range []string{removed, removedDirectory} {
		if !lookedUp[filepath.Clean(path)] {
			t.Errorf("entry info lookup hook did not run for removed entry %q", path)
		}
	}
	if err != nil {
		t.Fatalf("Validate rejected an entry removed before lstat: %v", err)
	}
}

func TestValidateRejectsEntryReplacedBySymlinkBetweenReadDirAndLstat(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "target.md"), []byte("module\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	replaced := filepath.Join(root, "replaced.md")
	if err := os.WriteFile(replaced, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}

	lookedUp := false
	err := validateWithOwnerAndEntryInfo(root, DefaultOwnerLookup(), func(path string, entry fs.DirEntry) (os.FileInfo, error) {
		if filepath.Clean(path) == filepath.Clean(replaced) {
			lookedUp = true
			if err := os.Remove(path); err != nil {
				t.Fatalf("remove entry after readdir: %v", err)
			}
			if err := os.Symlink("target.md", path); err != nil {
				t.Skipf("host cannot create symbolic links: %v", err)
			}
		}
		return entry.Info()
	})
	if !lookedUp {
		t.Fatal("entry info lookup hook did not run for replaced entry")
	}
	var failure *Failure
	if !errors.As(err, &failure) || failure.Check != CheckLinkSafety {
		t.Fatalf("Validate error = %v, want link-safety failure for replacement symlink", err)
	}
}

func TestValidateFailsClosedForOtherEntryLstatErrors(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(root, "unreadable.md")
	if err := os.WriteFile(other, []byte("module\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected lstat failure")
	err := validateWithOwnerAndEntryInfo(root, DefaultOwnerLookup(), func(path string, entry fs.DirEntry) (os.FileInfo, error) {
		if filepath.Clean(path) == filepath.Clean(other) {
			return nil, injected
		}
		return entry.Info()
	})
	var failure *Failure
	if !errors.As(err, &failure) || failure.Check != CheckRegular || !errors.Is(err, injected) {
		t.Fatalf("Validate error = %v, want regular-types failure retaining the lstat error", err)
	}
}

// The same vanished-entry rule applies to probes after lstat (on Windows the
// readdir metadata is cached, so the owner/DACL/link handle probes are the
// first to observe a removed entry).
func TestValidateSkipsEntryVanishedDuringOwnerProbe(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	vanished := filepath.Join(root, "vanished.lock")
	if err := os.WriteFile(vanished, []byte("lock\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	lookup := DefaultOwnerLookup()
	probed := false
	err := ValidateWithOwner(root, func(path string, info os.FileInfo) (OwnerIdentity, error) {
		if filepath.Clean(path) == filepath.Clean(vanished) {
			probed = true
			return "", &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
		}
		return lookup(path, info)
	})
	if !probed {
		t.Fatal("owner probe did not run for the vanished entry")
	}
	if err != nil {
		t.Fatalf("Validate rejected an entry that vanished during the owner probe: %v", err)
	}
}

func TestValidateFailsClosedForOtherOwnerProbeErrors(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(root, "unreadable.md")
	if err := os.WriteFile(other, []byte("module\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	lookup := DefaultOwnerLookup()
	injected := errors.New("injected owner probe failure")
	err := ValidateWithOwner(root, func(path string, info os.FileInfo) (OwnerIdentity, error) {
		if filepath.Clean(path) == filepath.Clean(other) {
			return "", injected
		}
		return lookup(path, info)
	})
	var failure *Failure
	if !errors.As(err, &failure) || failure.Check != CheckOwnership || !errors.Is(err, injected) {
		t.Fatalf("Validate error = %v, want ownership failure retaining the probe error", err)
	}
}
