package transaction

import (
	"errors"
	"path/filepath"
	"testing"
)

// Filtering cannot mask corrupt nonmatching intent or mutate stored journals.
// The same valid intent is selectable and remains byte-identical afterward.
func TestPendingRecoveryMatching(t *testing.T) {
	root := t.TempDir()
	engine := mustEngine(t, filepath.Join(root, "home"))
	target := fileTarget(t, "profile", "one", filepath.Join(root, "live"), filepath.Join(root, "stage"), "old", "new")
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "pending", ProjectIdentity: "capture", Targets: []Target{target}}); err != nil {
		t.Fatal(err)
	}
	path := engine.journalPath("pending")
	before := mustRead(t, path)
	unchanged := func() {
		t.Helper()
		if mustRead(t, path) != before || mustRead(t, target.LivePath) != "old" {
			t.Fatal("selector changed the journal or live target")
		}
	}
	if pending, err := engine.PendingRecoveryMatching(testLock{}, nil); err == nil || pending {
		t.Fatalf("nil selector = %v, %v", pending, err)
	}
	unchanged()
	want := errors.New("selector unavailable")
	if pending, err := engine.PendingRecoveryMatching(testLock{}, func(Journal) (bool, error) { return false, want }); !errors.Is(err, want) || pending {
		t.Fatalf("selector failure = %v, %v", pending, err)
	}
	unchanged()
	if pending, err := engine.PendingRecoveryMatching(testLock{}, func(j Journal) (bool, error) {
		j.Targets[0].LivePath = filepath.Join(root, "detached-only")
		return false, nil
	}); err != nil || pending {
		t.Fatalf("nonmatching control = %v, %v", pending, err)
	}
	unchanged()
	if pending, err := engine.PendingRecoveryMatching(testLock{}, func(j Journal) (bool, error) { return j.Targets[0].LivePath == target.LivePath, nil }); err != nil || !pending {
		t.Fatalf("matching control = %v, %v", pending, err)
	}
	unchanged()
	corrupt := engine.journalPath("z-corrupt")
	mustWrite(t, corrupt, "{")
	if pending, err := engine.PendingRecoveryMatching(testLock{}, func(Journal) (bool, error) { return false, nil }); err == nil || pending {
		t.Fatalf("filter masked corrupt inventory = %v, %v", pending, err)
	}
	unchanged()
	if mustRead(t, corrupt) != "{" {
		t.Fatal("observation repaired corrupt intent")
	}
}
