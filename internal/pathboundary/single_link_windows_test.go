//go:build windows

package pathboundary

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Handle metadata must identify the same file observed by lstat. A replacement
// that is itself single-link cannot satisfy the earlier file's proof.
func TestSingleLinkWindowsBindsObservedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload")
	if err := os.WriteFile(path, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err = snapshotFileInfo(path, info)
	if err != nil {
		t.Fatal(err)
	}
	if err := singleLink(path, info); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, path+"-previous"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := singleLink(path, info); err == nil || err.Error() != "snapshot file identity changed during link metadata inspection" {
		t.Fatalf("replacement accepted or misclassified: %v", err)
	}
}

// The production strict entry point must freeze identity before the owner
// probe, for both named files and discovered children. A lazy Lstat identity
// would instead accept the replacement. The intact neighbor remains valid.
func TestSnapshotWindowsRejectsReplacementDuringOwnerProbe(t *testing.T) {
	for _, targetKind := range []string{"file", "directory"} {
		t.Run(targetKind, func(t *testing.T) {
			root := privateRoot(t)
			payload := filepath.Join(root, "payload")
			if err := os.WriteFile(payload, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := ProtectTree(root); err != nil {
				t.Fatal(err)
			}
			target := root
			if targetKind == "file" {
				target = payload
			}
			if err := ValidateSnapshotWithinWithOwner(root, target, DefaultOwnerLookup()); err != nil {
				t.Fatalf("intact positive refused: %v", err)
			}
			replaced := false
			err := ValidateSnapshotWithinWithOwner(root, target, func(path string, info os.FileInfo) (OwnerIdentity, error) {
				if path == payload && !replaced {
					replaced = true
					if err := os.Rename(payload, payload+"-previous"); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(payload, []byte("replacement"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				return DefaultOwnerLookup()(path, info)
			})
			var failure *Failure
			if !replaced || !errors.As(err, &failure) || failure.Check != CheckSingleLink || failure.Path != payload || failure.Cause == nil || failure.Cause.Error() != "snapshot file identity changed during link metadata inspection" {
				t.Fatalf("replacement accepted or misclassified: %v", err)
			}
			for path, want := range map[string]string{payload: "replacement", payload + "-previous": "original"} {
				if got, err := os.ReadFile(path); err != nil || string(got) != want {
					t.Fatalf("refusal changed fixture bytes: %v", err)
				}
			}
		})
	}
}
