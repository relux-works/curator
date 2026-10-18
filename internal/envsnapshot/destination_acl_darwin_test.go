//go:build darwin

package envsnapshot

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A real kernel deny-only ancestor ACL cannot broaden mode permissions; a
// kernel permit ACE does broaden them and must refuse even with mode 0700.
func TestDestinationKernelACL(t *testing.T) {
	root, policy := destinationSandbox(t)
	parent := filepath.Join(root, "acl-parent")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if out, err := exec.Command("/bin/chmod", "-N", parent).CombinedOutput(); err != nil {
			t.Errorf("ACL fixture cleanup failed: %v %s", err, out)
		}
	})
	chmod := func(acl string) {
		t.Helper()
		if out, err := exec.Command("/bin/chmod", "+a", acl, parent).CombinedOutput(); err != nil {
			t.Fatalf("ACL setup failed: %v %s", err, out)
		}
	}
	chmod("everyone deny delete")
	d := admitDestination(t, filepath.Join(parent, "child"), policy, DestinationAbsent)
	chmod("everyone allow add_file,add_subdirectory")
	_, err := OpenDestination(filepath.Join(parent, "child"), policy)
	requireDestinationError(t, err, DestinationInvalid, "directory-acl")
	requireDestinationError(t, d.Revalidate(), DestinationInvalid, "directory-acl")
	assertAbsent(t, filepath.Join(parent, "child"))
}

// SNAP-ID's mode/byte checks alone cannot detect a file ACL granting another
// principal read access. Prior admission and later MatchPrior both refuse it.
func TestDestinationPriorKernelACL(t *testing.T) {
	root, policy := destinationSandbox(t)
	path := filepath.Join(root, "snapshot")
	snapshotAtDestination(t, path)
	d := admitDestination(t, path, policy, DestinationSnapshot)
	member := filepath.Join(path, "home/AGENTS.md")
	t.Cleanup(func() {
		if out, err := exec.Command("/bin/chmod", "-N", member).CombinedOutput(); err != nil {
			t.Errorf("ACL fixture cleanup failed: %v %s", err, out)
		}
	})
	if out, err := exec.Command("/bin/chmod", "+a", "everyone allow read", member).CombinedOutput(); err != nil {
		t.Fatalf("ACL setup failed: %v %s", err, out)
	}
	_, err := d.MatchPrior(read(t, "golden/manifest.json"))
	requireDestinationError(t, err, Mismatch, "snapshot-owner-acl")
	_, err = OpenDestination(path, policy)
	requireDestinationError(t, err, Mismatch, "snapshot-owner-acl")
	b, err := os.ReadFile(member)
	if err != nil || Digest(b) != Digest(read(t, "golden/home/AGENTS.md")) {
		t.Fatal("prior ACL refusal changed file bytes", err)
	}
}
