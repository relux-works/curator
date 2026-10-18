//go:build linux

package envsnapshot

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

// Real access/default POSIX ACLs refuse even when Unix mode bits alone grant
// no group/other write. Tests require an ACL-capable filesystem, never skip a
// failed ACL operation as if it were evidence of absence.
func TestDestinationKernelACL(t *testing.T) {
	for _, attr := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		t.Run(attr, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			parent := filepath.Join(root, "acl-parent")
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			d := admitDestination(t, filepath.Join(parent, "child"), policy, DestinationAbsent)
			setDestinationKernelACL(t, parent, attr)
			_, err := OpenDestination(filepath.Join(parent, "child"), policy)
			requireDestinationError(t, err, DestinationInvalid, "directory-acl")
			requireDestinationError(t, d.Revalidate(), DestinationInvalid, "directory-acl")
			assertAbsent(t, filepath.Join(parent, "child"))
		})
	}
}

func setDestinationKernelACL(t *testing.T, parent, attr string) {
	t.Helper()
	// Linux posix_acl_xattr_header(version=2) + five tagged entries.
	acl := make([]byte, 4+5*8)
	binary.LittleEndian.PutUint32(acl, 2)
	entries := []struct {
		tag, perm uint16
		id        uint32
	}{
		{1, 7, 0xffffffff}, {2, 4, 65534}, {4, 0, 0xffffffff},
		{16, 4, 0xffffffff}, {32, 0, 0xffffffff},
	}
	for i, e := range entries {
		binary.LittleEndian.PutUint16(acl[4+i*8:], e.tag)
		binary.LittleEndian.PutUint16(acl[6+i*8:], e.perm)
		binary.LittleEndian.PutUint32(acl[8+i*8:], e.id)
	}
	if err := unix.Setxattr(parent, attr, acl, 0); err != nil {
		t.Fatal("kernel ACL setup failed", err)
	}
}

// A default ACL on a prior home directory can affect future children without
// changing its 0700 mode; actual prior admission must still reject it.
func TestDestinationPriorKernelACL(t *testing.T) {
	root, policy := destinationSandbox(t)
	path := filepath.Join(root, "snapshot")
	snapshotAtDestination(t, path)
	d := admitDestination(t, path, policy, DestinationSnapshot)
	setDestinationKernelACL(t, filepath.Join(path, "home"), "system.posix_acl_default")
	_, err := d.MatchPrior(read(t, "golden/manifest.json"))
	requireDestinationError(t, err, Mismatch, "snapshot-owner-acl")
	_, err = OpenDestination(path, policy)
	requireDestinationError(t, err, Mismatch, "snapshot-owner-acl")
	b, err := os.ReadFile(filepath.Join(path, "manifest.json"))
	if err != nil || Digest(b) != Digest(read(t, "golden/manifest.json")) {
		t.Fatal("prior ACL refusal changed metadata", err)
	}
}
