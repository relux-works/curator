package publicationprobe

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) (string, *os.File) {
	t.Helper()
	root, err := os.MkdirTemp("..", "publication-probe-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	f, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, name := range []string{"target", "stage"} {
		if err := os.Mkdir(filepath.Join(root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	return root, f
}
func retain(t *testing.T, root, name string) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}
func same(t *testing.T, a, b *os.File) bool {
	t.Helper()
	ai, e := a.Stat()
	if e != nil {
		t.Fatal(e)
	}
	bi, e := b.Stat()
	if e != nil {
		t.Fatal(e)
	}
	return os.SameFile(ai, bi)
}
func replace(t *testing.T, root string) *os.File {
	t.Helper()
	if err := os.Rename(filepath.Join(root, "target"), filepath.Join(root, "original")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "target"), 0700); err != nil {
		t.Fatal(err)
	}
	return retain(t, root, "target")
}

// Platform capability probe, not a production publication test. A no-replace
// rename succeeds for absence but refuses even a verified empty target.
func TestNoReplaceCannotReplaceRetainedEmptyTarget(t *testing.T) {
	root, parent := fixture(t)
	fd := int(parent.Fd())
	original := retain(t, root, "target")
	if err := unix.RenameatxNp(fd, "stage", fd, "target", unix.RENAME_EXCL); !errors.Is(err, unix.EEXIST) {
		t.Fatalf("got %v, want EEXIST", err)
	}
	current := retain(t, root, "target")
	if !same(t, original, current) {
		t.Fatal("no-replace changed target")
	}
	if err := unix.RenameatxNp(fd, "stage", fd, "absent", unix.RENAME_EXCL); err != nil {
		t.Fatal("absent positive control", err)
	}
	t.Log("existing empty target: EEXIST; absent target: success")
}

// Darwin explicitly refuses the combination that would need to exchange only
// the expected target while refusing an independently replaced entry.
func TestNoReplaceAndExchangeAreIncompatible(t *testing.T) {
	_, parent := fixture(t)
	fd := int(parent.Fd())
	if err := unix.RenameatxNp(fd, "stage", fd, "target", unix.RENAME_EXCL|unix.RENAME_SWAP); !errors.Is(err, unix.EINVAL) {
		t.Fatalf("got %v, want EINVAL", err)
	}
	t.Log("RENAME_EXCL | RENAME_SWAP: EINVAL")
}

// Retained handles and an advisory flock do not stop a same-user replacement.
// Name-based rmdir after the last identity check deletes the competing empty
// entry, then no-replace publication succeeds over that lost entry.
func TestRecheckThenRmdirLosesCompetingEmptyEntry(t *testing.T) {
	root, parent := fixture(t)
	fd := int(parent.Fd())
	original := retain(t, root, "target")
	if err := unix.Flock(int(original.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	defer unix.Flock(int(original.Fd()), unix.LOCK_UN)
	checked := retain(t, root, "target")
	if !same(t, original, checked) {
		t.Fatal("positive identity check failed")
	}
	competitor := replace(t, root)
	if same(t, original, competitor) {
		t.Fatal("race was not planted")
	}
	if err := unix.Unlinkat(fd, "target", unix.AT_REMOVEDIR); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(root, "target")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("competing entry not removed", err)
	}
	if err := unix.RenameatxNp(fd, "stage", fd, "target", unix.RENAME_EXCL); err != nil {
		t.Fatal(err)
	}
	current := retain(t, root, "target")
	if same(t, current, competitor) {
		t.Fatal("witness did not replace competitor")
	}
	t.Log("retained identity checked; directory flock held; replacement allowed; rmdir removed competitor; no-replace then succeeded")
}

// An ordinary retained-parent rename replaces whichever empty entry currently
// occupies the target name; no expected-target inode is supplied to the kernel.
func TestOrdinaryRenameOverwritesCompetingEmptyEntry(t *testing.T) {
	root, parent := fixture(t)
	fd := int(parent.Fd())
	original := retain(t, root, "target")
	competitor := replace(t, root)
	if same(t, original, competitor) {
		t.Fatal("race was not planted")
	}
	if err := unix.Renameat(fd, "stage", fd, "target"); err != nil {
		t.Fatal(err)
	}
	current := retain(t, root, "target")
	if same(t, current, competitor) {
		t.Fatal("witness did not replace competitor")
	}
	t.Log("renameat succeeded over a different empty target inode")
}

// Exchange is atomic but unconditional: a nonempty substituted user directory
// is moved into the staging name before a post-exchange identity check can run.
func TestExchangeMovesCompetingNonemptyUserEntry(t *testing.T) {
	root, parent := fixture(t)
	fd := int(parent.Fd())
	original := retain(t, root, "target")
	competitor := replace(t, root)
	if err := os.WriteFile(filepath.Join(root, "target", "sentinel"), []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if same(t, original, competitor) {
		t.Fatal("race was not planted")
	}
	if err := unix.RenameatxNp(fd, "stage", fd, "target", unix.RENAME_SWAP); err != nil {
		t.Fatal(err)
	}
	moved := retain(t, root, "stage")
	if !same(t, moved, competitor) {
		t.Fatal("competitor was not exchanged")
	}
	b, err := os.ReadFile(filepath.Join(root, "stage", "sentinel"))
	if err != nil || string(b) != "preserve" {
		t.Fatal("sentinel witness invalid", err)
	}
	t.Log("RENAME_SWAP moved the competing nonempty user entry to stage; this is already a forbidden namespace mutation")
}

// unlinkat cannot remove a directory by the retained fd with an empty or dot
// name. Falling back to a parent/name removal reintroduces the planted race.
func TestNoHandleConditionalDirectoryRemoval(t *testing.T) {
	root, _ := fixture(t)
	target := retain(t, root, "target")
	fd := int(target.Fd())
	if err := unix.Unlinkat(fd, "", unix.AT_REMOVEDIR); !errors.Is(err, unix.ENOENT) {
		t.Fatalf("empty-name got %v, want ENOENT", err)
	}
	if err := unix.Unlinkat(fd, ".", unix.AT_REMOVEDIR); !errors.Is(err, unix.EINVAL) {
		t.Fatalf("dot-name got %v, want EINVAL", err)
	}
	t.Log("unlinkat(retained target, empty): ENOENT; dot: EINVAL")
}
