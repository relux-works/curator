package projectlease

import (
	"context"
	"errors"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The actual launcher subprocess returns/exits after durable runner transfer;
// the parent's next manager still observes an active lease, not caller lifetime.
func TestLauncherExitRetainsLease(t *testing.T) {
	f := setup(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestLeaseLauncherHelper$")
	cmd.Env = append(os.Environ(), "CURATOR_TEST_LEASE_HOME="+f.home)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("launcher helper failed: %v\n%s", err, out)
	}
	f.reserve("one")
	f.observe("one", func(e *Evidence) { e.Running = true; e.ProcessGeneration = "boot-session-runner-1" })
	m, err := New(f.home, f.a, f.a)
	if err != nil {
		t.Fatal(err)
	}
	f.m = m
	if row := status(t, f, "one"); row.State != Active || !row.Busy {
		t.Fatalf("launcher exit released transferred lease: %+v", row)
	}
}
func TestLeaseLauncherHelper(t *testing.T) {
	home := os.Getenv("CURATOR_TEST_LEASE_HOME")
	if home == "" {
		return
	}
	f := atHome(t, home)
	m, err := New(home, f.a, f.a)
	if err != nil {
		t.Fatal(err)
	}
	f.m = m
	activate(t, f, "one")
}

// Executable generated control surfaces retain their protected receipt mode;
// native bytes never enter the inventory. Mode changes refuse provider start.
func TestGeneratedModesAndInventory(t *testing.T) {
	f := setup(t)
	f.files = map[string][]byte{"bin/dispatch": []byte("fixture generated command")}
	f.c.Generated = inventory(f.files)
	file := f.c.Generated["bin/dispatch"]
	file.Mode = 0700
	f.c.Generated["bin/dispatch"] = file
	f.a.comps[f.c.Ref] = f.c
	acquire(t, f, "one")
	if err := f.m.PublishPrepared(context.Background(), "one", f.files, false); err != nil {
		t.Fatal(err)
	}
	f.observe("one", func(e *Evidence) { e.Running = true; e.ProcessGeneration = "runner" })
	if err := f.m.Transfer(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		return
	} // Windows uses private DACLs; native project adapters are not qualified there.
	info, err := os.Stat(filepath.Join(f.home, homeRelative(f.c), "bin/dispatch"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Fatal("receipt executable mode lost")
	}
	if err = os.Chmod(filepath.Join(f.home, homeRelative(f.c), "bin/dispatch"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = f.m.Resume(context.Background(), "one")
	mustRefuse(t, err, "mode drift")
}

// A terminated native session resumes under a new reserved runner generation
// with the original composition even after the idle home's current view changed.
// Retargeting, missing original state and revocation create no replacement lease.
func TestPreparedResumeUsesOriginalComposition(t *testing.T) {
	f := setup(t)
	old := activate(t, f, "one")
	f.observe("one", func(e *Evidence) {
		e.Running = false
		e.Terminated = true
		e.OwnershipReconciled = true
		e.DescendantsTerminated = true
	})
	if err := f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	next := f.c
	next.Ref = "composition:latest"
	next.SHA256 = strings.Repeat("c", 64)
	nextFiles := map[string][]byte{"AGENTS.md": []byte("latest")}
	next.Generated = inventory(nextFiles)
	f.a.comps[next.Ref] = next
	if err := f.m.Update(context.Background(), next.Ref, next.SHA256, old.OwnerRef, nextFiles, false); err != nil {
		t.Fatal(err)
	}
	r := f.reserve("resumed")
	r.ContextRef = next.Ref
	r.ContextSHA256 = next.SHA256
	_, err := f.m.PrepareResume(context.Background(), "one", r)
	mustRefuse(t, err, "cannot change")
	r.ContextRef = ""
	r.ContextSHA256 = ""
	f.a.revoked = true
	_, err = f.m.PrepareResume(context.Background(), "one", r)
	if !errors.Is(err, ErrRevoked) {
		t.Fatalf("revoked resume reservation=%v", err)
	}
	if len(mustLoad(t, f).Leases) != 1 {
		t.Fatal("refused resume created fresh lease")
	}
	f.a.revoked = false
	b, err := f.m.PrepareResume(context.Background(), "one", r)
	if err != nil {
		t.Fatal(err)
	}
	if b.ContextRef != old.ContextRef || b.ContextSHA256 != old.ContextSHA256 {
		t.Fatal("resume chose current composition")
	}
	if err = f.m.PublishPrepared(context.Background(), r.LeaseID, f.files, false); err != nil {
		t.Fatal(err)
	}
	f.observe("resumed", func(e *Evidence) { e.Running = true; e.ProcessGeneration = "resumed-runner" })
	if err = f.m.Transfer(context.Background(), r.LeaseID); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(f.home, homeRelative(f.c), "AGENTS.md"))
	if string(raw) != "approved\n" {
		t.Fatal("resumed wrong generated view")
	}
}

// An actual abrupt launcher exit leaves durable recovery intent. Read-only
// status retains the busy lease; recovery never infers session termination.
func TestLeaseCrashRecovery(t *testing.T) {
	f := setup(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestLeaseCrashHelper$")
	cmd.Env = append(os.Environ(), "CURATOR_TEST_LEASE_CRASH_HOME="+f.home)
	out, err := cmd.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 43 {
		t.Fatalf("crash helper=%v %s", err, out)
	}
	f.reserve("one")
	if row := status(t, f, "one"); row.State != Unknown || !row.Busy {
		t.Fatalf("crash status=%+v", row)
	}
	if err = f.m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if row := status(t, f, "one"); row.State != Unknown || !row.Busy {
		t.Fatalf("crash recovery released lease=%+v", row)
	}
	mustRefuse(t, f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, false), Busy)
}
func TestLeaseCrashHelper(t *testing.T) {
	home := os.Getenv("CURATOR_TEST_LEASE_CRASH_HOME")
	if home == "" {
		return
	}
	f := atHome(t, home)
	acquire(t, f, "one")
	f.m.hooks = transaction.Hooks{Observe: func(e transaction.Event) {
		if e.Point == transaction.PointAfterInstall {
			os.Exit(43)
		}
	}}
	if err := f.m.PublishPrepared(context.Background(), "one", f.files, false); err != nil {
		t.Fatal(err)
	}
	t.Fatal("crash plant did not execute")
}
