package transaction

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envunit"
)

func directoryPlan(t *testing.T, home string) (Plan, string) {
	t.Helper()
	dir := filepath.Join(home, "new-home")
	live := filepath.Join(dir, "owned-output")
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("owned-output"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := envunit.ObserveDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	return Plan{TransactionID: "directory-unit", ProjectIdentity: "profiles", Directories: []DirectoryAction{{LivePath: dir, Before: before, Mode: 0700}}, Targets: []Target{{PrivateFile: true, Class: "profile", Identifier: live, LivePath: live, StagedSource: source, PreimageDigest: DigestAbsent}}}, live
}

// Detects an unallocated in-memory planning placeholder planted as a durable
// preparing journal. A real prepared-directory restart is the nearby positive;
// this tests the journal reader, not a resolver-only admission substitute.
func TestProvisioningOriginMetadataUnboundPreparing(t *testing.T) {
	home := t.TempDir()
	plan, live := directoryPlan(t, home)
	engine, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	journal, _, err := engine.buildJournal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if journal.Directories[0].After.Identity != "" {
		t.Fatal("plant did not reach unallocated directory placeholder")
	}
	if err := engine.ensureJournalRoot(); err != nil {
		t.Fatal(err)
	}
	raw, err := marshalJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(engine.journalPath(plan.TransactionID), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.loadJournal(plan.TransactionID); err == nil || !strings.Contains(err.Error(), "unbound durable directory") {
		t.Fatalf("wanted specific unbound durable directory refusal, got %v", err)
	}
	if _, err := os.Lstat(filepath.Dir(live)); !os.IsNotExist(err) {
		t.Fatalf("reader changed absent live directory: %v", err)
	}
	if got, err := os.ReadFile(engine.journalPath(plan.TransactionID)); err != nil || string(got) != string(raw) {
		t.Fatalf("reader modified planted journal: %v", err)
	}
}

// Detects a missing metadata action before journal creation, with its intended
// diagnostic. The healthy publication above prevents reject-all qualification.
func TestProvisioningOriginMetadataUnenlisted(t *testing.T) {
	home := t.TempDir()
	plan, live := directoryPlan(t, home)
	plan.Directories = nil
	engine, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Prepare(testLock{}, plan); err == nil || !strings.Contains(err.Error(), "transaction_parent_not_enlisted") {
		t.Fatalf("wanted unenlisted parent refusal, got %v", err)
	}
	if _, err := os.Lstat(filepath.Dir(live)); !os.IsNotExist(err) {
		t.Fatalf("refusal created live parent: %v", err)
	}
	if _, err := os.Lstat(engine.journalPath(plan.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("refusal created journal: %v", err)
	}
}

// Proves absent-parent creation is journaled and Prepare leaves live parents
// absent. The payload is a synthetic owned output, never a tool home snapshot.
func TestProvisioningOriginMetadataPublication(t *testing.T) {
	home := t.TempDir()
	engine, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	plan, live := directoryPlan(t, home)
	lock := testLock{}
	if _, err := engine.Prepare(lock, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Dir(live)); !os.IsNotExist(err) {
		t.Fatalf("Prepare changed live parent: %v", err)
	}
	if err := engine.Commit(lock, plan.TransactionID); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(live)
	if err != nil || string(got) != "owned-output" {
		t.Fatalf("publication: %q %v", got, err)
	}
	if state, err := envunit.ObserveDirectory(filepath.Dir(live)); err != nil || state.Mode != 0700 {
		t.Fatalf("parent protection %+v %v", state, err)
	}
}

// Proves restart recognizes the exact rename-before-journal-state window.
// In-memory callbacks/undo are discarded; a new Engine resumes durable proof.
func TestProvisioningOriginMetadataRestart(t *testing.T) {
	home := t.TempDir()
	plan, live := directoryPlan(t, home)
	reached := false
	engine, err := New(home, WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointAfterDirectoryInstall {
			reached = true
			panic("owned fixture stop")
		}
		return nil
	}}))
	if err != nil {
		t.Fatal(err)
	}
	lock := testLock{}
	if _, err := engine.Prepare(lock, plan); err != nil {
		t.Fatal(err)
	}
	func() { defer func() { _ = recover() }(); _ = engine.Commit(lock, plan.TransactionID) }()
	if !reached {
		t.Fatal("directory install boundary was not reached")
	}
	fresh, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(lock); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(live)
	if err != nil || string(got) != "owned-output" {
		t.Fatalf("restart output %q %v", got, err)
	}
}

// Proves a later file failure restores directory absence; a nearby healthy
// publication prevents a reject-all implementation from satisfying this claim.
func TestProvisioningOriginMetadataRollback(t *testing.T) {
	home := t.TempDir()
	plan, live := directoryPlan(t, home)
	reached := false
	failure := errors.New("owned file boundary failure")
	engine, err := New(home, WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointBeforeInstall {
			reached = true
			return failure
		}
		return nil
	}}))
	if err != nil {
		t.Fatal(err)
	}
	lock := testLock{}
	if _, err := engine.Prepare(lock, plan); err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(lock, plan.TransactionID); !errors.Is(err, failure) {
		t.Fatalf("wrong failure: %v", err)
	}
	if !reached {
		t.Fatal("file install failure did not reach owner")
	}
	if _, err := os.Lstat(filepath.Dir(live)); !os.IsNotExist(err) {
		t.Fatalf("rollback left new live directory: %v", err)
	}
}

// Proves restarted forward publication refuses a substituted prepared parent
// specifically, preserving both that entry and unrelated fixture bytes.
func TestProvisioningOriginMetadataSubstitution(t *testing.T) {
	home := t.TempDir()
	plan, live := directoryPlan(t, home)
	engine, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Prepare(testLock{}, plan); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Dir(live), 0700); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(filepath.Dir(live), "unrelated")
	if err := os.WriteFile(foreign, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	fresh, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.Recover(testLock{}); !errors.Is(err, ErrImplementationCorruption) {
		t.Fatalf("wanted identity refusal, got %v", err)
	}
	if data, err := os.ReadFile(foreign); err != nil || string(data) != "preserve" {
		t.Fatalf("substituted directory changed: %q %v", data, err)
	}
	if _, err := os.Lstat(live); !os.IsNotExist(err) {
		t.Fatalf("refusal published payload: %v", err)
	}
}

// Proves nested missing parents and existing protection changes are distinct
// durable actions, and a later file failure restores all three preimages.
func TestProvisioningOriginMetadataNested(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			home := t.TempDir()
			parent := filepath.Join(home, "existing")
			if err := os.Mkdir(parent, 0755); err != nil {
				t.Fatal(err)
			}
			plan, oldLive := directoryPlan(t, home)
			_ = oldLive
			outer := filepath.Join(parent, "outer")
			inner := filepath.Join(outer, "inner")
			live := filepath.Join(inner, "owned-output")
			plan.Targets[0].LivePath = live
			plan.Targets[0].Identifier = live
			plan.Directories = nil
			for _, path := range []string{parent, outer, inner} {
				before, err := envunit.ObserveDirectory(path)
				if err != nil {
					t.Fatal(err)
				}
				plan.Directories = append(plan.Directories, DirectoryAction{LivePath: path, Before: before, Mode: 0700})
			}
			failure := errors.New("nested fixture file failure")
			reached := false
			engine, err := New(home, WithHooks(Hooks{Fault: func(e Event) error {
				if fail && e.Point == PointBeforeInstall {
					reached = true
					return failure
				}
				return nil
			}}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Prepare(testLock{}, plan); err != nil {
				t.Fatal(err)
			}
			err = engine.Commit(testLock{}, plan.TransactionID)
			if fail {
				if !errors.Is(err, failure) || !reached {
					t.Fatalf("intended owner failure not reached: %v", err)
				}
				if _, err := os.Lstat(outer); !os.IsNotExist(err) {
					t.Fatalf("new parent survived rollback %v", err)
				}
				if info, err := os.Lstat(parent); err != nil || info.Mode().Perm() != 0755 {
					t.Fatalf("protection preimage not restored: %v %v", info, err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if data, err := os.ReadFile(live); err != nil || string(data) != "owned-output" {
					t.Fatalf("nested publication: %q %v", data, err)
				}
			}
		})
	}
}
