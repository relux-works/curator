package projectlease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/transaction"
)

func releaseForRepair(t *testing.T, f fixture) {
	t.Helper()
	f.observe("one", func(e *Evidence) {
		e.Running = false
		e.Terminated = true
		e.DescendantsTerminated = true
		e.OwnershipReconciled = true
	})
	if err := f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Released {
		t.Fatal("repair fixture is not released")
	}
}

func driftForRepair(t *testing.T, path, kind string) {
	t.Helper()
	var err error
	switch kind {
	case "changed":
		err = os.WriteFile(path, []byte("drift\n"), 0600)
	case "missing":
		err = os.Remove(path)
	case "mode":
		if runtime.GOOS == "windows" {
			t.Skip("exact mode bits are Unix-specific")
		}
		err = os.Chmod(path, 0700)
	}
	if err != nil {
		t.Fatal(err)
	}
}

// Idle Update restores receipt-approved bytes/modes after full release,
// including changed, missing and mode-drifted files, without replacing the
// native home or synthetic login. The unchanged case is a positive control.
func TestIdleRepairGeneratedDrift(t *testing.T) {
	for _, kind := range []string{"unchanged", "changed", "missing", "mode"} {
		t.Run(kind, func(t *testing.T) {
			f := setup(t)
			activate(t, f, "one")
			releaseForRepair(t, f)
			home := filepath.Join(f.home, homeRelative(f.c))
			target := filepath.Join(home, "AGENTS.md")
			before, err := os.Stat(home)
			if err != nil {
				t.Fatal(err)
			}
			login := filepath.Join(home, "native.json")
			if err = os.WriteFile(login, []byte("fixture login"), 0600); err != nil {
				t.Fatal(err)
			}
			driftForRepair(t, target, kind)
			if err = f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, false); err != nil {
				t.Fatalf("idle repair refused %s generated state: %v", kind, err)
			}
			raw, err := os.ReadFile(target)
			info, statErr := os.Stat(target)
			if err != nil || statErr != nil || string(raw) != string(f.files["AGENTS.md"]) {
				t.Fatalf("repair did not restore approved bytes: %q %v %v", raw, err, statErr)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
				t.Fatalf("repair mode = %o, want 600", info.Mode().Perm())
			}
			after, err := os.Stat(home)
			if err != nil || !os.SameFile(before, after) {
				t.Fatal("repair replaced native home", err)
			}
			raw, err = os.ReadFile(login)
			if err != nil || string(raw) != "fixture login" {
				t.Fatal("repair changed native login", err)
			}
		})
	}
}

// The same drift cannot be repaired under a live lease: Update returns Busy,
// shared PublishPrepared/Resume refuse drift, and none of them rewrites it.
func TestLiveLeaseDriftRefusesRepair(t *testing.T) {
	for _, kind := range []string{"changed", "missing", "mode"} {
		t.Run(kind, func(t *testing.T) {
			f := setup(t)
			activate(t, f, "one")
			acquire(t, f, "two")
			target := filepath.Join(f.home, homeRelative(f.c), "AGENTS.md")
			driftForRepair(t, target, kind)
			before, err := transaction.DigestTarget(transaction.KindBytes, target)
			if err != nil {
				t.Fatal(err)
			}
			mustRefuse(t, f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, false), Busy)
			if err = f.m.PublishPrepared(context.Background(), "two", f.files, false); err == nil {
				t.Fatal("live shared publication repaired drift")
			}
			if _, err = f.m.Resume(context.Background(), "one"); err == nil {
				t.Fatal("live resume accepted drift")
			}
			after, err := transaction.DigestTarget(transaction.KindBytes, target)
			if err != nil || before != after || status(t, f, "one").State != Active {
				t.Fatal("refused live repair changed state", err)
			}
		})
	}
}

// Fault after a generated-file install rolls the repair back to its actual
// drifted preimage and original ledger. A subsequent valid repair succeeds.
func TestIdleRepairRollback(t *testing.T) {
	f := setup(t)
	activate(t, f, "one")
	releaseForRepair(t, f)
	target := filepath.Join(f.home, homeRelative(f.c), "AGENTS.md")
	driftForRepair(t, target, "changed")
	before, _ := os.ReadFile(statePath(f.home))
	injected := errors.New("injected idle repair install fault")
	installed := false
	f.m.hooks = transaction.Hooks{Fault: func(e transaction.Event) error {
		if e.Point == transaction.PointAfterInstall {
			installed = true
			return injected
		}
		return nil
	}}
	err := f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, false)
	if !installed || !errors.Is(err, injected) {
		t.Fatalf("repair fault did not reach intended install: %v", err)
	}
	f.m.hooks = transaction.Hooks{}
	if err = f.m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(target)
	after, ledgerErr := os.ReadFile(statePath(f.home))
	if err != nil || ledgerErr != nil || string(raw) != "drift\n" || string(before) != string(after) {
		t.Fatal("failed repair lost its preimage or changed ledger", err, ledgerErr)
	}
	if err = f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, false); err != nil {
		t.Fatal("retry repair failed", err)
	}
}

// A directory replacing an inventoried regular file is not repairable drift:
// its unrecorded child must survive, with no generated or ledger publication.
// Normal regular-file repair is covered by TestIdleRepairGeneratedDrift.
func TestIdleRepairRefusesDirectoryTakeover(t *testing.T) {
	f := setup(t)
	activate(t, f, "one")
	releaseForRepair(t, f)
	target := filepath.Join(f.home, homeRelative(f.c), "AGENTS.md")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(target, "native.json")
	if err := os.WriteFile(child, []byte("unrecorded native"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := transaction.DigestTarget(transaction.KindBytes, target)
	if err != nil {
		t.Fatal(err)
	}
	ledgerBefore, err := os.ReadFile(statePath(f.home))
	if err != nil {
		t.Fatal(err)
	}
	mustRefuse(t, f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, false), "not a regular file")
	after, err := transaction.DigestTarget(transaction.KindBytes, target)
	ledgerAfter, ledgerErr := os.ReadFile(statePath(f.home))
	if err != nil || ledgerErr != nil || before != after || string(ledgerBefore) != string(ledgerAfter) {
		t.Fatal("directory refusal changed native state or ledger", err, ledgerErr)
	}
}
