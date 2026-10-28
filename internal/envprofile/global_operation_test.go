package envprofile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// Historical test-only model: this does not qualify production rollback.
// Compensation restores bytes, modes, link text, removed backup trees and
// absence without tightening an unchanged ancestor. It never follows a captured
// symlink into its unrelated target.
func TestGlobalProfileCompensationImages(t *testing.T) {
	root := t.TempDir()
	if runtime.GOOS != "windows" {
		if err := os.Chmod(root, 0755); err != nil {
			t.Fatal(err)
		}
	}
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
	op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}, managerHome: root}}
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
	if runtime.GOOS != "windows" {
		if info, err := os.Lstat(root); err != nil || info.Mode().Perm() != 0755 {
			t.Fatalf("tree restoration tightened its ancestor: %v", err)
		}
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
				op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}}}
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
	op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}, managerHome: root}}
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
	op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}}}
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

// A later failure and an unknown replacement of a committed marker preserve
// the original error and report transaction corruption. Durable recovery must
// also refuse the unknown bytes; it cannot bless them or pretend rollback was
// complete. The healthy/ordinary rollback pair lives in the joined owner test.
func TestProvisioningOriginGlobalRollbackConflict(t *testing.T) {
	fx, resolver := recordedGlobalOriginFixture(t)
	manager, err := managerlock.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	}()
	planted := errors.New("planted joined publication failure")
	path := filepath.Join(fx.native["codex_cli"], envmarker.Name)
	saved := newOperationEngine
	t.Cleanup(func() { newOperationEngine = saved })
	called := false
	newOperationEngine = func(home string, opts ...transaction.Option) (*transaction.Engine, error) {
		return saved(home, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
			if e.Point == transaction.PointAfterInstall && e.LivePath == path && !called {
				called = true
				if err := os.WriteFile(path, []byte("operator marker"), 0600); err != nil {
					t.Fatal(err)
				}
				return planted
			}
			return nil
		}}))...)
	}
	calls := 0
	_, err = publishGlobalContinuationFixture(t, fx.home, resolver, lock, &calls)
	newOperationEngine = saved
	if !called || !errors.Is(err, planted) || !errors.Is(err, transaction.ErrImplementationCorruption) {
		t.Fatalf("joined rollback lost failure/corruption: called=%v error=%v", called, err)
	}
	if raw, e := os.ReadFile(path); e != nil || string(raw) != "operator marker" {
		t.Fatalf("foreign marker changed: %q %v", raw, e)
	}
	engine, err := saved(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	before := map[string]string{}
	for _, root := range []string{fx.native["codex_cli"], EnvRoot(fx.home), filepath.Join(fx.home, "state", "transactions")} {
		before[root] = nativeSnapshot(t, root)
	}
	if err = engine.Recover(lock); err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || !strings.Contains(err.Error(), "unreachable live/backup state for native-marker") {
		t.Fatalf("fresh recovery blessed unknown marker: %v", err)
	}
	for root, snapshot := range before {
		if nativeSnapshot(t, root) != snapshot {
			t.Fatal("fresh refusal changed pending unit or surface preimages")
		}
	}
	if raw, e := os.ReadFile(path); e != nil || string(raw) != "operator marker" {
		t.Fatalf("fresh recovery changed foreign marker: %q %v", raw, e)
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
			op := &legacyCompensationModel{undo: &globalUndo{managerHome: home, entries: map[string]*globalImage{}}}
			if err := op.capture(native, record); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(native, []byte("published native"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(record, []byte("pending lock"), 0600); err != nil {
				t.Fatal(err)
			}
			if !writable {
				if err := os.Chmod(parent, 0500); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(parent, 0700) })
			}
			if err := op.recordPostimages(native, record); err != nil {
				t.Fatal(err)
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

// Existing parent modes restore only from the published mode; operator modes
// refuse before a captured leaf changes. Created parents use the same guard before
// removal; an unchanged created parent disappears on successful compensation.
func TestGlobalProfileCompensationParentModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are the tested metadata")
	}
	for _, created := range []bool{false, true} {
		for _, changed := range []bool{false, true} {
			t.Run(fmt.Sprintf("created=%v/changed=%v", created, changed), func(t *testing.T) {
				root := t.TempDir()
				parent := filepath.Join(root, "parent")
				leaf := filepath.Join(parent, "entry")
				if !created {
					if err := os.Mkdir(parent, 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.Chmod(parent, 0755); err != nil {
						t.Fatal(err)
					}
				}
				op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}}}
				if err := op.capture(leaf); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(parent, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(parent, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(leaf, []byte("publication"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := op.recordPostimages(leaf); err != nil {
					t.Fatal(err)
				}
				if changed {
					if err := os.Chmod(parent, 0750); err != nil {
						t.Fatal(err)
					}
				}
				err := op.undo.restore()
				if changed {
					if !errors.Is(err, transaction.ErrImplementationCorruption) {
						t.Fatalf("operator parent mode admitted: %v", err)
					}
					if info, err := os.Lstat(parent); err != nil || info.Mode().Perm() != 0750 {
						t.Fatalf("operator parent mode changed: %v %v", info, err)
					}
					if data, err := os.ReadFile(leaf); err != nil || string(data) != "publication" {
						t.Fatalf("parent conflict partially restored leaf: %q %v", data, err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if _, err := os.Lstat(leaf); !os.IsNotExist(err) {
					t.Fatalf("exact rollback retained leaf: %v", err)
				}
				if created {
					if _, err := os.Lstat(parent); !os.IsNotExist(err) {
						t.Fatalf("exact rollback retained created parent: %v", err)
					}
				} else if info, err := os.Lstat(parent); err != nil || info.Mode().Perm() != 0755 {
					t.Fatalf("exact rollback lost original parent mode: %v %v", info, err)
				}
			})
		}
	}
}

// Mode-only and symlink-target-only changes are foreign postimages even when
// file bytes are unchanged. Owner mismatch cases exercise the recorded identity
// comparison with synthetic prior owners; they do not require a privileged
// chown. The capture control checks the real UID on Unix and nonempty SID on
// Windows. Unchanged neighbors restore the original bytes/mode or absence.
func TestGlobalProfileCompensationMetadata(t *testing.T) {
	for _, kind := range []string{"existing-file-mode", "created-file-mode", "symlink-target", "owner", "parent-owner", "owner-capture"} {
		for _, changed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/changed=%v", kind, changed), func(t *testing.T) {
				if runtime.GOOS == "windows" && strings.Contains(kind, "file-mode") {
					t.Skip("Unix permission bits are the tested metadata")
				}
				root := t.TempDir()
				path := filepath.Join(root, "entry")
				op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}}}
				if kind == "existing-file-mode" {
					if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
						t.Fatal(err)
					}
				}
				if err := op.capture(path); err != nil {
					t.Fatal(err)
				}
				if kind == "symlink-target" {
					if err := os.Symlink("published-target", path); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(path, []byte("published"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := op.recordPostimages(path); err != nil {
					t.Fatal(err)
				}
				if kind == "owner-capture" {
					image := op.undo.postimages[path]
					if image.owner == "" || (runtime.GOOS != "windows" && string(image.owner) != fmt.Sprintf("uid:%d", os.Geteuid())) {
						t.Fatal("image owner not captured")
					}
				}
				if changed {
					switch kind {
					case "existing-file-mode", "created-file-mode":
						if err := os.Chmod(path, 0640); err != nil {
							t.Fatal(err)
						}
					case "symlink-target":
						if err := os.Remove(path); err != nil {
							t.Fatal(err)
						}
						if err := os.Symlink("operator-target", path); err != nil {
							t.Fatal(err)
						}
					case "owner":
						op.undo.postimages[path].owner += "-different-recorded-owner"
					case "parent-owner":
						op.undo.parentPostimages[root].owner += "-different-recorded-owner"
					}
				}
				before, err := readGlobalImage(path)
				if err != nil {
					t.Fatal(err)
				}
				err = op.undo.restore()
				if changed && kind != "owner-capture" {
					if !errors.Is(err, transaction.ErrImplementationCorruption) {
						t.Fatalf("foreign metadata admitted: %v", err)
					}
					after, err := readGlobalImage(path)
					if err != nil || !reflect.DeepEqual(before, after) {
						t.Fatalf("foreign metadata changed: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if kind == "existing-file-mode" {
					data, err := os.ReadFile(path)
					if err != nil || string(data) != "original" {
						t.Fatalf("original bytes lost: %q %v", data, err)
					}
					if info, err := os.Lstat(path); err != nil || info.Mode().Perm() != 0644 {
						t.Fatalf("original mode lost: %v", err)
					}
				} else if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Fatalf("unchanged metadata blocked removal: %v", err)
				}
			})
		}
	}
}

// An empty created parent is removed only at its published mode. A mode-only
// operator change is retained even when every captured leaf is already absent.
func TestGlobalProfileCompensationCreatedEmptyParentMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are the tested metadata")
	}
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprintf("changed=%v", changed), func(t *testing.T) {
			parent := filepath.Join(t.TempDir(), "created")
			leaf := filepath.Join(parent, "absent")
			op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}}}
			if err := op.capture(leaf); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			if err := op.recordPostimages(leaf); err != nil {
				t.Fatal(err)
			}
			if changed {
				if err := os.Chmod(parent, 0750); err != nil {
					t.Fatal(err)
				}
			}
			err := op.undo.restore()
			if changed {
				if !errors.Is(err, transaction.ErrImplementationCorruption) {
					t.Fatalf("changed empty parent admitted: %v", err)
				}
				if info, err := os.Lstat(parent); err != nil || info.Mode().Perm() != 0750 {
					t.Fatalf("changed empty parent removed or chmodded: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if _, err := os.Lstat(parent); !os.IsNotExist(err) {
					t.Fatalf("unchanged empty parent retained: %v", err)
				}
			}
		})
	}
}

// Checking immediately before parent chmod rejects a change after preflight,
// including an operator's return to the preimage. An unchanged publication
// restores sticky and ordinary bits exactly.
func TestGlobalProfileCompensationLateParentMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix special permission bits are the tested metadata")
	}
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprintf("changed=%v", changed), func(t *testing.T) {
			root := t.TempDir()
			parent := filepath.Join(root, "parent")
			original := os.FileMode(0755) | os.ModeSticky
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(parent, original); err != nil {
				t.Fatal(err)
			}
			op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}}}
			leaf := filepath.Join(parent, "absent")
			if err := op.capture(leaf); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(parent, 0700); err != nil {
				t.Fatal(err)
			}
			if err := op.recordPostimages(leaf); err != nil {
				t.Fatal(err)
			}
			if err := op.undo.checkParentPostimage(parent); err != nil {
				t.Fatal(err)
			}
			if changed {
				if err := os.Chmod(parent, original); err != nil {
					t.Fatal(err)
				}
			}
			err := op.undo.restoreParents()
			if changed {
				if !errors.Is(err, transaction.ErrImplementationCorruption) {
					t.Fatalf("late parent mode admitted: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if info, err := os.Lstat(parent); err != nil || (info.Mode().Perm() != 0755 || info.Mode()&os.ModeSticky == 0) {
				t.Fatalf("parent special mode lost: %v", err)
			}
		})
	}
}

// Directory-only metadata, as recorded for empty Muse runtime directories,
// checkpoints at home publication even without a captured child. Unchanged
// existing/created parents restore mode/absence; operator chmod preserves both
// the parent and the published marker and returns the specific conflict.
func TestGlobalProfileCompensationDirectoryOnlyParents(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are the tested metadata")
	}
	for _, created := range []bool{false, true} {
		for _, changed := range []bool{false, true} {
			t.Run(fmt.Sprintf("created=%v/changed=%v", created, changed), func(t *testing.T) {
				root := t.TempDir()
				parent := filepath.Join(root, "cache")
				marker := filepath.Join(root, envmarker.Name)
				if !created {
					if err := os.Mkdir(parent, 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.Chmod(parent, 0755); err != nil {
						t.Fatal(err)
					}
				}
				op := &legacyCompensationModel{undo: &globalUndo{entries: map[string]*globalImage{}}}
				if err := op.capture(marker); err != nil {
					t.Fatal(err)
				}
				if err := op.noteParents(filepath.Join(parent, ".directory-mode")); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(parent, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(parent, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(marker, []byte("published marker"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := op.recordHomePostimages(root); err != nil {
					t.Fatal(err)
				}
				if changed {
					if err := os.Chmod(parent, 0750); err != nil {
						t.Fatal(err)
					}
				}
				err := op.undo.restore()
				if changed {
					if !errors.Is(err, transaction.ErrImplementationCorruption) {
						t.Fatalf("directory-only operator mode admitted: %v", err)
					}
					if info, err := os.Lstat(parent); err != nil || info.Mode().Perm() != 0750 {
						t.Fatalf("directory-only operator mode changed: %v", err)
					}
					if data, err := os.ReadFile(marker); err != nil || string(data) != "published marker" {
						t.Fatalf("directory-only conflict changed marker: %q %v", data, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("directory-only rollback failed: %v", err)
				}
				if _, err := os.Lstat(marker); !os.IsNotExist(err) {
					t.Fatalf("directory-only rollback retained marker: %v", err)
				}
				if created {
					if _, err := os.Lstat(parent); !os.IsNotExist(err) {
						t.Fatalf("directory-only rollback retained parent: %v", err)
					}
				} else if info, err := os.Lstat(parent); err != nil || info.Mode().Perm() != 0755 {
					t.Fatalf("directory-only original mode lost: %v", err)
				}
			})
		}
	}
}
