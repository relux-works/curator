package envprofile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
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
	if err := op.recordHomePostimages(root); err != nil {
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

// Exact published images roll back to the preimage; replacing a published
// link, file, tree, or removal with unknown content refuses before any other
// target changes. This covers the shared guard on both removal and restoration.
func TestGlobalProfileCompensationPostimages(t *testing.T) {
	for _, kind := range []string{"link", "file", "tree", "removed"} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/replacement=%v", kind, replace), func(t *testing.T) {
				root := t.TempDir()
				path, other := filepath.Join(root, "entry"), filepath.Join(root, "other")
				if err := os.WriteFile(path, []byte("preimage"), 0640); err != nil {
					t.Fatal(err)
				}
				op := &operation{undo: &globalUndo{entries: map[string]*globalImage{}}}
				if err := op.capture(path, other); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				switch kind {
				case "link":
					if err := os.Symlink("missing-published-target", path); err != nil {
						t.Fatal(err)
					}
				case "file":
					if err := os.WriteFile(path, []byte("published"), 0600); err != nil {
						t.Fatal(err)
					}
				case "tree":
					if err := os.Mkdir(path, 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(path, "leaf"), []byte("published"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(other, []byte("other publication"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := op.recordPostimages(path, other); err != nil {
					t.Fatal(err)
				}
				if replace {
					if err := os.RemoveAll(path); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte("operator replacement"), 0644); err != nil {
						t.Fatal(err)
					}
				}
				err := op.undo.restore()
				if replace {
					if !errors.Is(err, transaction.ErrImplementationCorruption) {
						t.Fatalf("foreign postimage admitted: %v", err)
					}
					if data, err := os.ReadFile(path); err != nil || string(data) != "operator replacement" {
						t.Fatalf("foreign postimage changed: %q %v", data, err)
					}
					if data, err := os.ReadFile(other); err != nil || string(data) != "other publication" {
						t.Fatalf("conflict partially restored another target: %q %v", data, err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if data, err := os.ReadFile(path); err != nil || string(data) != "preimage" {
						t.Fatalf("exact postimage did not restore preimage: %q %v", data, err)
					}
					if _, err := os.Lstat(other); !os.IsNotExist(err) {
						t.Fatalf("exact postimage did not restore absence: %v", err)
					}
				}
			})
		}
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

// Replacing a recorded parent with a symlink is a corruption refusal before
// reading or restoring its leaf, even when the foreign leaf matches publication.
func TestGlobalProfileCompensationParentReplacement(t *testing.T) {
	root, foreign := t.TempDir(), t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "entry")
	op := &operation{undo: &globalUndo{entries: map[string]*globalImage{}}}
	if err := op.capture(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("published", path); err != nil {
		t.Fatal(err)
	}
	if err := op.recordPostimages(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("published", filepath.Join(foreign, "entry")); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(parent, filepath.Join(root, "saved")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, parent); err != nil {
		t.Fatal(err)
	}
	if err := op.undo.restore(); !errors.Is(err, transaction.ErrImplementationCorruption) {
		t.Fatalf("replacement parent admitted: %v", err)
	}
	if link, err := os.Readlink(filepath.Join(foreign, "entry")); err != nil || link != "published" {
		t.Fatalf("compensation changed foreign parent contents: %q %v", link, err)
	}
}

// The actual paired transaction returns a typed compensation error, retaining
// the original failure and published install record when its native marker is
// replaced by an operator. This tests the exported error chain, not its text.
func TestGlobalProfileCompensationFailureType(t *testing.T) {
	pinOperatorHome(t, t.TempDir())
	homes := pinHomes(t)
	home := t.TempDir()
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	planted := errors.New("planted continuation failure")
	path := filepath.Join(homes[envregistry.CodexCLI], envmarker.Name)
	called := false
	_, err = GlobalSkillTransaction(home, Policy{}, envregistry.MachineConfig{}, nil, nil, lock, func() error {
		called = true
		if err := os.WriteFile(path, []byte("operator marker"), 0600); err != nil {
			t.Fatal(err)
		}
		return planted
	})
	var compensation *GlobalCompensationError
	if !called || !errors.As(err, &compensation) || !errors.Is(err, planted) || !errors.Is(compensation, transaction.ErrImplementationCorruption) {
		t.Fatalf("paired compensation lost typed error chain: called=%v error=%v", called, err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "operator marker" {
		t.Fatalf("foreign marker changed: %q %v", data, err)
	}
	if _, err := os.Stat(sourcePath(home, DefaultProfile)); err != nil {
		t.Fatalf("compensation lost pending install record: %v", err)
	}
}

// A native home nested inside the manager home still rolls back before profile
// recovery records. An unwritable native parent leaves the published lock;
// the adjacent writable control restores both targets. Unix permission checks
// require an unprivileged process.
func TestGlobalProfileCompensationNestedNativeHome(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("requires unprivileged Unix directory permissions")
	}
	for _, writable := range []bool{false, true} {
		t.Run(fmt.Sprintf("writable=%v", writable), func(t *testing.T) {
			home := t.TempDir()
			parent := filepath.Join(home, "native")
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			native := filepath.Join(parent, "entry")
			record := filepath.Join(ProfileDir(home, "long-profile-name"), "lock.json")
			if err := os.MkdirAll(filepath.Dir(record), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(record, []byte("old lock"), 0600); err != nil {
				t.Fatal(err)
			}
			op := &operation{undo: &globalUndo{managerHome: home, entries: map[string]*globalImage{}}}
			if err := op.capture(native, record); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(native, []byte("published native"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(record, []byte("pending lock"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := op.recordPostimages(native, record); err != nil {
				t.Fatal(err)
			}
			if !writable {
				if err := os.Chmod(parent, 0500); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(parent, 0700) })
			}
			err := op.undo.restore()
			if writable {
				if err != nil {
					t.Fatal(err)
				}
				if data, err := os.ReadFile(record); err != nil || string(data) != "old lock" {
					t.Fatalf("writable nested rollback did not restore lock: %q %v", data, err)
				}
				if _, err := os.Stat(native); !os.IsNotExist(err) {
					t.Fatalf("writable nested rollback retained native entry: %v", err)
				}
			} else {
				if !errors.Is(err, os.ErrPermission) {
					t.Fatalf("nested native permission fault not exercised: %v", err)
				}
				if data, err := os.ReadFile(record); err != nil || string(data) != "pending lock" {
					t.Fatalf("nested native compensation lost pending lock: %q %v", data, err)
				}
			}
		})
	}
}
