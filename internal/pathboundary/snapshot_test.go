//go:build unix || windows

package pathboundary

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Snapshot validation accepts ordinary read-only/executable payloads and
// directory link counts, then rejects an outside-tree hardlink without changing
// either entry. The separate operator-source API continues to accept that form.
func TestSnapshotSingleLinkBoundary(t *testing.T) {
	for _, mode := range []os.FileMode{0600, 0444, 0755} {
		for _, targetKind := range []string{"directory", "file"} {
			t.Run(fmt.Sprintf("%o/%s", mode, targetKind), func(t *testing.T) {
				root := privateRoot(t)
				entry := filepath.Join(root, "snapshot")
				if err := os.Mkdir(entry, 0700); err != nil {
					t.Fatal(err)
				}
				payload := filepath.Join(entry, "payload")
				if err := os.WriteFile(payload, []byte("snapshot\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := ProtectTree(entry); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(payload, mode); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(payload, 0600) })
				target := entry
				if targetKind == "file" {
					target = payload
				}
				validate := func() error {
					return ValidateSnapshotWithinWithOwner(root, target, DefaultOwnerLookup())
				}
				if err := validate(); err != nil {
					t.Fatalf("intact single-link payload refused: %v", err)
				}
				plant := filepath.Join(t.TempDir(), "outside-link")
				if err := os.Link(payload, plant); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(payload, 0600)
					_ = os.Remove(plant)
				})
				before, err := os.Lstat(payload)
				if err != nil {
					t.Fatal(err)
				}
				err = validate()
				var failure *Failure
				if !errors.As(err, &failure) || failure.Check != CheckSingleLink || failure.Path != payload {
					t.Fatalf("want named single_link refusal, got %v", err)
				}
				for _, path := range []string{payload, plant} {
					after, err := os.Lstat(path)
					if err != nil || !os.SameFile(before, after) || after.Mode() != before.Mode() {
						t.Fatalf("refusal changed payload or plant: %s: %v", path, err)
					}
					data, err := os.ReadFile(path)
					if err != nil || string(data) != "snapshot\n" {
						t.Fatalf("refusal changed bytes: %s: %v", path, err)
					}
				}
				if err := ValidateWithinWithOwner(root, target, DefaultOwnerLookup()); err != nil {
					t.Fatalf("snapshot rule leaked into operator-source admission: %v", err)
				}
			})
		}
	}
}

// A discovered child that disappears during a metadata probe keeps the
// established absence rule; an unavailable proof for a still-present file is a
// named structural refusal. This does not relax named target absence.
func TestSnapshotMetadataProbeRefusalAndAbsence(t *testing.T) {
	for _, absent := range []bool{false, true} {
		t.Run(fmt.Sprint(absent), func(t *testing.T) {
			root := privateRoot(t)
			payload := filepath.Join(root, "payload")
			if err := os.WriteFile(payload, []byte("snapshot\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := ProtectTree(root); err != nil {
				t.Fatal(err)
			}
			if err := ValidateSnapshotWithinWithOwner(root, root, DefaultOwnerLookup()); err != nil {
				t.Fatal(err)
			}
			called := false
			err := validateWithinWithRegularCheck(root, root, DefaultOwnerLookup(), defaultEntryInfo, func(path string, _ os.FileInfo) error {
				called = true
				if absent {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					return os.ErrNotExist
				}
				return os.ErrPermission
			})
			if !called {
				t.Fatal("metadata probe was not exercised")
			}
			if absent {
				if err != nil {
					t.Fatalf("discovered absent child refused: %v", err)
				}
				return
			}
			var failure *Failure
			if !errors.As(err, &failure) || failure.Check != CheckSingleLink || failure.Path != payload || !errors.Is(err, os.ErrPermission) {
				t.Fatalf("unavailable metadata must fail closed: %v", err)
			}
			if data, err := os.ReadFile(payload); err != nil || string(data) != "snapshot\n" {
				t.Fatalf("refusal changed existing payload: %v", err)
			}
		})
	}
}
