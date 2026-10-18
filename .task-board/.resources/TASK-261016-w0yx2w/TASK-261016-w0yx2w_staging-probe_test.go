//go:build darwin || linux

package stagingprobe

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func sandbox(t *testing.T) (string, *os.File) {
	t.Helper()
	root := t.TempDir()
	f, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	return root, f
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func identity(t *testing.T, parent *os.File, name string, original *os.File) {
	t.Helper()
	var observed, retained unix.Stat_t
	must(t, unix.Fstatat(int(parent.Fd()), name, &observed, unix.AT_SYMLINK_NOFOLLOW))
	must(t, unix.Fstat(int(original.Fd()), &retained))
	if observed.Dev != retained.Dev || observed.Ino != retained.Ino {
		t.Fatal("precondition: last identity check failed")
	}
}

func swapBeforeUnlink(t *testing.T, root string, parent *os.File, kind string) {
	t.Helper()
	stage := filepath.Join(root, "stage")
	if kind == "directory" {
		must(t, os.Mkdir(stage, 0700))
	} else {
		must(t, os.WriteFile(stage, []byte("invocation-owned"), 0600))
	}
	original, err := os.Open(stage)
	must(t, err)
	defer func() { must(t, original.Close()) }()
	// The adversarial schedule is AFTER the last no-follow identity check.
	identity(t, parent, "stage", original)
	must(t, unix.Renameat(int(parent.Fd()), "stage", int(parent.Fd()), "original"))
	flags := 0
	switch kind {
	case "directory":
		must(t, os.Mkdir(stage, 0700))
		flags = unix.AT_REMOVEDIR
	case "file":
		must(t, os.WriteFile(stage, []byte("unrelated-user-file"), 0600))
	case "link":
		must(t, os.WriteFile(filepath.Join(root, "outside"), []byte("preserved"), 0600))
		must(t, os.Symlink("outside", stage))
	default:
		t.Fatal("unknown fixture kind")
	}
	must(t, unix.Unlinkat(int(parent.Fd()), "stage", flags))
	if _, err := original.Stat(); err != nil {
		t.Fatal("retained original handle lost", err)
	}
	if _, err := os.Lstat(filepath.Join(root, "original")); err != nil {
		t.Fatal("moved original lost", err)
	}
}

// A retained parent and a successful no-follow identity check do not bind the
// later unlink to the retained object. All three replacements are removed.
func TestProbeUnlinkUsesCurrentEntry(t *testing.T) {
	for _, kind := range []string{"file", "directory", "link"} {
		t.Run(kind, func(t *testing.T) {
			root, parent := sandbox(t)
			swapBeforeUnlink(t, root, parent, kind)
			if _, err := os.Lstat(filepath.Join(root, "stage")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("replacement was not removed", err)
			}
			if kind == "link" {
				b, err := os.ReadFile(filepath.Join(root, "outside"))
				if err != nil || string(b) != "preserved" {
					t.Fatal("symlink referent changed", err)
				}
			}
		})
	}
}

// EXPECTED RED: these assertions state the requested unconditional invariant
// against a candidate check-then-unlink algorithm, not against product code.
func TestRequestedCleanupInvariant(t *testing.T) {
	for _, kind := range []string{"file", "directory", "link"} {
		t.Run(kind, func(t *testing.T) {
			root, parent := sandbox(t)
			swapBeforeUnlink(t, root, parent, kind)
			if _, err := os.Lstat(filepath.Join(root, "stage")); err != nil {
				t.Fatal("VIOLATION: cleanup removed substituted staging entry")
			}
		})
	}
}

// No-replace protects an existing destination, including a competing empty
// directory, while absent-target publication succeeds on the actual syscall.
func TestProbeNoReplaceDestination(t *testing.T) {
	for _, competing := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "competitor"}[competing], func(t *testing.T) {
			root, parent := sandbox(t)
			must(t, os.Mkdir(filepath.Join(root, "stage"), 0700))
			must(t, os.WriteFile(filepath.Join(root, "stage", "payload"), []byte("captured"), 0600))
			var before os.FileInfo
			if competing {
				must(t, os.Mkdir(filepath.Join(root, "target"), 0700))
				var err error
				before, err = os.Stat(filepath.Join(root, "target"))
				must(t, err)
			}
			err := noReplace(int(parent.Fd()), "stage", "target")
			if competing {
				if !errors.Is(err, unix.EEXIST) {
					t.Fatal("expected EEXIST", err)
				}
				after, statErr := os.Stat(filepath.Join(root, "target"))
				must(t, statErr)
				if !os.SameFile(before, after) {
					t.Fatal("competitor replaced")
				}
				entries, readErr := os.ReadDir(filepath.Join(root, "target"))
				must(t, readErr)
				if len(entries) != 0 {
					t.Fatal("competitor changed")
				}
			} else {
				must(t, err)
				b, readErr := os.ReadFile(filepath.Join(root, "target", "payload"))
				must(t, readErr)
				if string(b) != "captured" {
					t.Fatal("wrong captured bytes")
				}
			}
		})
	}
}

// No-replace does not condition the source on its retained handle: a source
// replacement AFTER the last identity check is moved to the absent target.
func TestProbeNoReplaceSourceSubstitution(t *testing.T) {
	root, parent := sandbox(t)
	must(t, os.Mkdir(filepath.Join(root, "stage"), 0700))
	original, err := os.Open(filepath.Join(root, "stage"))
	must(t, err)
	defer func() { must(t, original.Close()) }()
	identity(t, parent, "stage", original)
	must(t, unix.Renameat(int(parent.Fd()), "stage", int(parent.Fd()), "original"))
	must(t, os.Mkdir(filepath.Join(root, "stage"), 0700))
	must(t, os.WriteFile(filepath.Join(root, "stage", "user-file"), []byte("unrelated"), 0600))
	must(t, noReplace(int(parent.Fd()), "stage", "target"))
	b, err := os.ReadFile(filepath.Join(root, "target", "user-file"))
	must(t, err)
	if string(b) != "unrelated" {
		t.Fatal("replacement was not moved")
	}
	retained, err := original.Stat()
	must(t, err)
	published, err := os.Stat(filepath.Join(root, "target"))
	must(t, err)
	if os.SameFile(retained, published) {
		t.Fatal("published original, expected source-substitution witness")
	}
}

// Neither an empty name nor '.' removes the retained directory through its own
// descriptor on the exercised platform; a parent-relative name is required.
func TestProbeUnlinkHandleOnly(t *testing.T) {
	root, parent := sandbox(t)
	for _, name := range []string{"", "."} {
		err := unix.Unlinkat(int(parent.Fd()), name, unix.AT_REMOVEDIR)
		if err == nil {
			t.Fatal("unexpected handle-only removal")
		}
		t.Logf("name=%q errno=%v", name, err)
		if _, statErr := os.Stat(root); statErr != nil {
			t.Fatal("retained directory removed", statErr)
		}
	}
}
