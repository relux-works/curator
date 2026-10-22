package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The optional check sees the current caller's borrowed lock and an independent
// journal snapshot. Allowed commit/restart publishes; refusal before commit or
// after backup restores the old bytes and removes recoverable intent. Rollback
// must not consult a check which is now refusing forward publication.
func TestCommitCheck(t *testing.T) {
	for _, mode := range []string{"commit", "recovery", "after-backup"} {
		for _, refuse := range []bool{false, true} {
			t.Run(mode+map[bool]string{false: "/allowed", true: "/refused"}[refuse], func(t *testing.T) {
				root := t.TempDir()
				home := filepath.Join(root, "home")
				live := filepath.Join(root, "live")
				stage := filepath.Join(root, "stage")
				mustMkdirAll(t, stage)
				target := fileTarget(t, "owned", "file", live, filepath.Join(stage, "file"), "old", "new")
				lock := &pointerTestLock{}
				injected := errors.New("injected commit admission refusal")
				calls := 0
				backedUp := false
				check := func(got HomeLock, j Journal) error {
					calls++
					if got != lock || j.ProjectIdentity != "owner" {
						t.Fatal("check lost caller lock or journal identity")
					}
					j.Targets[0].DesiredDigest = "mutated snapshot"
					if refuse && (mode != "after-backup" || backedUp) {
						return injected
					}
					return nil
				}
				engine := mustEngine(t, home, WithCommitCheck(check), WithHooks(Hooks{Observe: func(e Event) {
					if e.Point == PointAfterBackup {
						backedUp = true
					}
				}}))
				journal, err := engine.Prepare(lock, Plan{TransactionID: "checked", ProjectIdentity: "owner", Targets: []Target{target}})
				if err != nil {
					t.Fatal(err)
				}
				if mode == "recovery" {
					engine = mustEngine(t, home, WithCommitCheck(check))
					err = engine.Recover(lock)
				} else {
					err = engine.Commit(lock, journal.TransactionID)
				}
				if refuse {
					if !errors.Is(err, injected) {
						t.Fatalf("commit check refusal = %v, want injected refusal", err)
					}
					if got := mustRead(t, live); got != "old" {
						t.Fatal("commit check refusal lost preimage", got)
					}
				} else if err != nil || mustRead(t, live) != "new" {
					t.Fatal("allowed commit check did not publish", err)
				}
				if calls == 0 {
					t.Fatal("commit check was never called")
				}
				if mode == "after-backup" && !backedUp {
					t.Fatal("backup refusal plant missed")
				}
				if _, err = os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
					t.Fatal("settled checked transaction remains", err)
				}
			})
		}
	}
}

// Owner-filtered observation recognizes the exact owner, does not match a
// prefix or another owner, and leaves its journal and live preimage untouched.
func TestPendingRecoveryForOwner(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	stage := filepath.Join(root, "stage")
	mustMkdirAll(t, stage)
	live := filepath.Join(root, "live")
	target := fileTarget(t, "owned", "file", live, filepath.Join(stage, "file"), "old", "new")
	engine := mustEngine(t, home)
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "owner-pending", ProjectIdentity: "project-leases", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(engine.journalPath(journal.TransactionID))
	if err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{"project-leases", "project", "other"} {
		pending, err := engine.PendingRecoveryFor(testLock{}, owner)
		if err != nil || pending != (owner == "project-leases") {
			t.Fatalf("owner-filtered pending = %v %v for %s", pending, err, owner)
		}
	}
	if _, err = engine.PendingRecoveryFor(testLock{}, ""); err == nil {
		t.Fatal("empty pending owner accepted")
	}
	if mustRead(t, live) != "old" {
		t.Fatal("observation changed live bytes")
	}
	after, err := os.ReadFile(engine.journalPath(journal.TransactionID))
	if err != nil || string(before) != string(after) {
		t.Fatal("observation changed intent", err)
	}
}
