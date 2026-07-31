//go:build windows

package buildsource

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newFrozenRootCase reaches the snapshot through a directory reparse point,
// because a Windows root replacement cannot be a rename.
//
// os.OpenRoot opens the validated directory with FILE_SHARE_READ and
// FILE_SHARE_WRITE and never FILE_SHARE_DELETE, so for as long as the token is
// frozen the kernel refuses to move or delete that directory — and refuses it
// for every enclosing directory too, because a rename fails while any handle in
// the subtree is open. The name itself can still be moved onto a different
// instance by repointing a reparse point in the path prefix, which is the
// replacement this fixture performs. The refused rename is asserted first: it
// is the platform's own half of the same rule and it is the reason the POSIX
// form cannot be used here.
func newFrozenRootCase(t *testing.T, parent string) frozenRootCase {
	t.Helper()
	first := filepath.Join(parent, "instance-1")
	second := filepath.Join(parent, "instance-2")
	for _, instance := range []string{first, second} {
		if err := os.MkdirAll(filepath.Join(instance, "snapshot"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(parent, "link")
	linkDirectory(t, first, link)
	return frozenRootCase{root: filepath.Join(link, "snapshot"), replace: func(t *testing.T) {
		t.Helper()
		writeTestFile(t, filepath.Join(second, "snapshot"), "file", []byte("same"))
		if err := os.Rename(filepath.Join(link, "snapshot"), filepath.Join(parent, "old")); err == nil {
			t.Fatal("a frozen root was renamed away while its validation handle was open")
		}
		if err := os.Remove(link); err != nil {
			t.Fatalf("remove the reparse point standing in for the root's parent: %v", err)
		}
		linkDirectory(t, second, link)
	}}
}

// linkDirectory points link at target through a directory reparse point,
// preferring a symbolic link and falling back to a junction. A junction needs
// no privilege, so this works on a plain local account. This case is required
// on Windows and tolerates no skip, so a host that can create neither fails it
// by name rather than quietly dropping the coverage.
func linkDirectory(t *testing.T, target, link string) {
	t.Helper()
	symlinkErr := os.Symlink(target, link)
	if symlinkErr == nil {
		return
	}
	t.Logf("symbolic links are unavailable to this account (%v); using a junction", symlinkErr)
	if output, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Fatalf("this host can create no directory reparse point: %v: %s", err, output)
	}
}

// writeInvalidProtocolPathEntry creates an entry whose name the protocol path
// rule refuses.
//
// The POSIX vector, `bad:name`, is unusable on NTFS: the colon opens an
// alternate data stream, so the entry that actually lands is the perfectly
// portable name `bad` and the case asserts nothing. A trailing dot is the
// Windows equivalent and a sharper one: PortableComponent refuses it precisely
// because Win32 path normalisation strips it, which would alias two distinct
// protocol paths onto one file. Creating it needs the extended-length prefix,
// which bypasses that normalisation — and is why the validator, not the path
// layer, has to be the thing that says no.
func writeInvalidProtocolPathEntry(t *testing.T, root string) string {
	t.Helper()
	const name = "bad."
	extended := `\\?\` + filepath.Join(root, name)
	if err := os.WriteFile(extended, []byte("x"), 0o644); err != nil {
		t.Fatalf("this host cannot create the name %q: %v", name, err)
	}
	// Only the extended-length form names this entry, so os.RemoveAll of the
	// temporary directory cannot delete it.
	t.Cleanup(func() { _ = os.Remove(extended) })
	return name
}
