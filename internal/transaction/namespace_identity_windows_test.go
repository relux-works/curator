//go:build windows

package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWindowsNamespaceIdentityIsCompletedWhenItIsTaken pins the Windows premise
// the pass snapshot rests on, and the behaviour that answers it. os.Stat there
// returns a record that has still to read the volume serial and file index from
// the live filesystem, so a sweep that stopped at the stat would hold an
// identity bound to no object yet — and would receive a read it could not
// perform as "a different object", the answer that lets an aliased pair
// through.
//
// The identity the sweep hands back therefore has to be complete: it must still
// compare as itself after its object is gone. The record the sweep would have
// held without that completion is taken here as well, and the one outcome it
// may never produce is an accepted identity that then compares as something
// else.
func TestWindowsNamespaceIdentityIsCompletedWhenItIsTaken(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "live", "target")
	mustWrite(t, path, "present")
	candidate := targetNamespacePath{owner: "target 0", kind: "live", path: path, key: path}

	completed, err := namespaceIdentity(candidate)
	if err != nil {
		t.Fatalf("identity read failed: %v", err)
	}
	stated, err := namespaceStat(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if !os.SameFile(completed, completed) {
		t.Fatal("the identity the sweep took stopped comparing as itself once its object was gone")
	}

	// Either the host refuses the stalled record, which is the not-exist an
	// eager host would have reported at the stat, or it answers it — but it may
	// not answer with an identity that no longer compares as itself.
	stalled, err := completeNamespaceIdentity(candidate, stated)
	switch {
	case err == nil && !os.SameFile(stalled, stalled):
		t.Fatal("an incomplete identity was accepted and then compared as a different object")
	case err != nil && !os.IsNotExist(err):
		t.Fatalf("completing the identity of a removed path = %v, want a not-exist", err)
	}
}
