package homelease

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

// HL13/15: faults around audit/tombstone publication cannot acknowledge a
// partial break. A retry after real journal recovery emits exactly one event;
// no test substitutes a fabricated successful observation.
func TestBreakAuditFaultRecovery(t *testing.T) {
	for _, point := range []transaction.Point{transaction.PointBeforeBackup, transaction.PointAfterBackup, transaction.PointBeforeInstall, transaction.PointAfterInstall, transaction.PointBeforeCleanup} {
		t.Run(string(point), func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			fired := false
			f.m.hooks = transaction.Hooks{Fault: func(e transaction.Event) error {
				if !fired && e.Point == point {
					fired = true
					return errors.New("fixture publication fault")
				}
				return nil
			}}
			out, _, err := f.m.Break(context.Background(), r.LeaseID, 1, true, "protection deliberately abandoned")
			requireRefusal(t, err, "environment_lease_write_failed", "publication_failed")
			if out.Changed || !fired {
				t.Fatalf("fault not exercised or acknowledged: %+v %t", out, fired)
			}
			f.m.hooks = transaction.Hooks{}
			lm, _ := managerlock.New(f.m.home)
			lock, err := lm.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := transaction.New(f.m.home)
			if err != nil {
				t.Fatal(err)
			}
			if err := engine.Recover(lock); err != nil {
				t.Fatal("recovery control:", err)
			}
			_ = lock.Close()
			out, reason, err := f.m.Break(context.Background(), r.LeaseID, 1, true, "protection deliberately abandoned")
			if err != nil || reason != "broken" && reason != "already_broken" || out.Lease.Phase != "broken" {
				t.Fatalf("retry: %+v %s %v", out, reason, err)
			}
			events, err := os.ReadDir(filepath.Join(state.Namespace(f.m.home), "events"))
			if err != nil || len(events) != 1 {
				t.Fatalf("audit: %v %v", events, err)
			}
			before := snapshot(t, f.m.home)
			out, reason, err = f.m.Break(context.Background(), r.LeaseID, 1, true, "protection deliberately abandoned")
			if err != nil || reason != "already_broken" || out.Changed {
				t.Fatalf("terminal replay: %+v %s %v", out, reason, err)
			}
			unchanged(t, before, f.m.home)
		})
	}
}

// HL15: both forward and rollback entry refuse before changing a pending
// journal, live bytes or sidecars. Restoring the exact terminal fixture ledger
// permits the original pending intent, proving the guard actually caused refusal.
func TestHomeLeaseRecoveryPreflight(t *testing.T) {
	for _, phase := range []transaction.Phase{transaction.PhasePrepared, transaction.PhaseRollingBack} {
		t.Run(string(phase), func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			blocking, err := os.ReadFile(state.LedgerPath(f.m.home))
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := f.m.Break(context.Background(), r.LeaseID, 1, true, "protection deliberately abandoned"); err != nil {
				t.Fatal(err)
			}
			terminal, err := os.ReadFile(state.LedgerPath(f.m.home))
			if err != nil {
				t.Fatal(err)
			}
			terminalLedger, err := state.Load(f.m.home)
			if err != nil {
				t.Fatal(err)
			}
			eventPath := state.EventPath(f.m.home, *terminalLedger.Leases[r.LeaseID].TerminalEvent)
			eventBytes, err := os.ReadFile(eventPath)
			if err != nil {
				t.Fatal(err)
			}
			lm, _ := managerlock.New(f.m.home)
			lock, err := lm.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = lock.Close() }()
			engine, err := transaction.New(f.m.home)
			if err != nil {
				t.Fatal(err)
			}
			live := filepath.Join(f.a.home, "environments", "sample", "claude_code", "generated")
			source := filepath.Join(f.a.home, "desired-fixture")
			if err := os.WriteFile(source, []byte("desired"), 0600); err != nil {
				t.Fatal(err)
			}
			plan := transaction.Plan{TransactionID: "fixture-pending", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "generated", Identifier: "generated", LivePath: live, StagedSource: source, PreimageDigest: transaction.DigestAbsent}}}
			journal, err := engine.Prepare(lock, plan)
			if err != nil {
				t.Fatal(err)
			}
			// This isolated preimage plant models a previously pending legacy
			// writer plus durable protection. A compliant acquire refuses pending.
			journal.Phase = phase
			journalPath := filepath.Join(f.m.home, "state", "transactions", "v1", plan.TransactionID+".json")
			raw, err := json.Marshal(journal)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(journalPath, append(raw, '\n'), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(state.LedgerPath(f.m.home), blocking, 0600); err != nil {
				t.Fatal(err)
			}
			// Restore a complete pre-terminal fixture, not a corrupt ledger
			// with an orphan terminal event (which has its own refusal test).
			if err := os.Remove(eventPath); err != nil {
				t.Fatal(err)
			}
			before := snapshot(t, f.m.home)
			err = engine.Recover(lock)
			requireRefusal(t, err, "environment_home_leased", "blocking_holders")
			unchanged(t, before, f.m.home)
			if err := os.WriteFile(state.LedgerPath(f.m.home), terminal, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(eventPath, eventBytes, 0600); err != nil {
				t.Fatal(err)
			}
			if err := engine.Recover(lock); err != nil {
				t.Fatalf("unleased recovery: %v", err)
			}
			_, err = os.Stat(live)
			if phase == transaction.PhasePrepared && err != nil {
				t.Fatal("forward control did not publish", err)
			}
			if phase == transaction.PhaseRollingBack && !errors.Is(err, os.ErrNotExist) {
				t.Fatal("rollback control did not preserve absence", err)
			}
		})
	}
}

// HL06/15/16: completed rollback of a first acquisition restores authoritative
// absence and permits a fresh acquisition. An interrupted committed cleanup
// instead blocks read, replay and GC-mark admission until explicit recovery;
// settled replay creates no second binding. No compiled-artifact GC is tested.
func TestAcquirePublicationFaultRecovery(t *testing.T) {
	for _, point := range []transaction.Point{transaction.PointBeforeBackup, transaction.PointAfterBackup, transaction.PointBeforeInstall, transaction.PointAfterInstall, transaction.PointBeforeCleanup} {
		t.Run(string(point), func(t *testing.T) {
			f := newFixture(t)
			fired := false
			f.m.hooks = transaction.Hooks{Fault: func(e transaction.Event) error {
				if !fired && e.Point == point {
					fired = true
					return errors.New("fixture acquisition fault")
				}
				return nil
			}}
			out, _, err := f.m.Acquire(context.Background(), f.req)
			requireRefusal(t, err, "environment_lease_write_failed", "publication_failed")
			if out.Changed || !fired {
				t.Fatalf("fault not exercised or acknowledged: %+v %t", out, fired)
			}
			f.m.hooks = transaction.Hooks{}
			before := snapshot(t, f.m.home)
			if point != transaction.PointBeforeCleanup {
				rows, err := f.m.List(context.Background(), "", "")
				if err != nil || len(rows) != 0 {
					t.Fatalf("completed rollback did not restore absence: %+v %v", rows, err)
				}
				unchanged(t, before, f.m.home)
				out, reason, err := f.m.Acquire(context.Background(), f.req)
				if err != nil || !out.Changed || out.Lease == nil || reason != "acquired" {
					t.Fatalf("fresh acquisition after rollback: %+v %s %v", out, reason, err)
				}
				before = snapshot(t, f.m.home)
				out, reason, err = f.m.Acquire(context.Background(), f.req)
				if err != nil || out.Changed || out.Lease == nil || reason != "acquired" {
					t.Fatalf("fresh acquisition replay: %+v %s %v", out, reason, err)
				}
				unchanged(t, before, f.m.home)
				return
			}
			if rows, err := f.m.List(context.Background(), "", ""); err == nil || len(rows) != 0 {
				t.Fatalf("pending read exposed an authoritative result: %+v %v", rows, err)
			} else {
				requireRefusal(t, err, "environment_lease_recovery_required", "pending_mutation")
			}
			unchanged(t, before, f.m.home)
			out, _, err = f.m.Acquire(context.Background(), f.req)
			requireRefusal(t, err, "environment_lease_recovery_required", "pending_mutation")
			if out.Changed {
				t.Fatal("pending replay acknowledged a write")
			}
			unchanged(t, before, f.m.home)
			lm, err := managerlock.New(f.m.home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := lm.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			roots, err := ReferencesUnderLock(f.m.home, lock)
			requireRefusal(t, err, "environment_lease_recovery_required", "pending_mutation")
			if len(roots.Homes)+len(roots.Protected)+len(roots.BuildKeys)+len(roots.RuntimeEntries) != 0 {
				t.Fatalf("pending roots returned: %+v", roots)
			}
			unchanged(t, before, f.m.home)
			engine, err := transaction.New(f.m.home)
			if err != nil {
				t.Fatal(err)
			}
			if err := engine.Recover(lock); err != nil {
				t.Fatalf("explicit recovery: %v", err)
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			rows, err := f.m.List(context.Background(), "", "")
			if err != nil || len(rows) != 1 || !rows[0].Blocking() || rows[0].Holder.Kind != "prepared" || rows[0].Holder.RequestID != f.req.RequestID || rows[0].Revision != 1 {
				t.Fatalf("recovered reservation: %+v %v", rows, err)
			}
			before = snapshot(t, f.m.home)
			out, reason, err := f.m.Acquire(context.Background(), f.req)
			if err != nil || out.Changed || reason != "acquired" || out.Lease == nil || out.Lease.LeaseID != rows[0].LeaseID {
				t.Fatalf("settled replay: %+v %s %v", out, reason, err)
			}
			unchanged(t, before, f.m.home)
			bindings, err := os.ReadDir(filepath.Join(state.Namespace(f.m.home), "bindings"))
			if err != nil || len(bindings) != 1 {
				t.Fatalf("reservation inventory: %v %v", bindings, err)
			}
		})
	}
}

// A failed first publication may remove only its own empty new namespace.
// A pre-existing valid empty ledger retains its bytes and directory identity;
// foreign content planted during rollback remains evidence and fails closed.
func TestAcquireRollbackPreservesNamespaceEvidence(t *testing.T) {
	for _, prior := range []string{"existing-ledger", "foreign-file"} {
		t.Run(prior, func(t *testing.T) {
			f := newFixture(t)
			namespace := state.Namespace(f.m.home)
			var original os.FileInfo
			var ledger []byte
			if prior == "existing-ledger" {
				if err := privatedir.MakeAll(namespace); err != nil {
					t.Fatal(err)
				}
				var err error
				original, err = os.Lstat(namespace)
				if err != nil || !os.SameFile(original, original) {
					t.Fatal("original directory identity unavailable", err)
				}
				ledger, err = state.Canonical(state.Empty())
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(state.LedgerPath(f.m.home), ledger, 0600); err != nil {
					t.Fatal(err)
				}
			}
			foreign := filepath.Join(namespace, "foreign-evidence")
			fired := false
			f.m.hooks = transaction.Hooks{Fault: func(e transaction.Event) error {
				if !fired && e.Point == transaction.PointBeforeBackup {
					fired = true
					if prior == "foreign-file" {
						if err := os.WriteFile(foreign, []byte("preserve fixture evidence"), 0600); err != nil {
							return err
						}
					}
					return errors.New("fixture rollback fault")
				}
				return nil
			}}
			out, _, err := f.m.Acquire(context.Background(), f.req)
			requireRefusal(t, err, "environment_lease_write_failed", "publication_failed")
			if out.Changed || !fired {
				t.Fatal("rollback fault not exercised")
			}
			f.m.hooks = transaction.Hooks{}
			before := snapshot(t, f.m.home)
			rows, err := f.m.List(context.Background(), "", "")
			if prior == "foreign-file" {
				requireRefusal(t, err, "environment_lease_unreadable", "ledger_missing")
				raw, err := os.ReadFile(foreign)
				if err != nil || string(raw) != "preserve fixture evidence" {
					t.Fatal("foreign evidence removed or changed", err)
				}
				out, _, err = f.m.Acquire(context.Background(), f.req)
				requireRefusal(t, err, "environment_lease_unreadable", "ledger_missing")
				if out.Changed {
					t.Fatal("unknown namespace acknowledged acquisition")
				}
			} else {
				if err != nil || len(rows) != 0 {
					t.Fatalf("existing empty ledger: %+v %v", rows, err)
				}
				current, err := os.Lstat(namespace)
				if err != nil || !os.SameFile(original, current) {
					t.Fatal("existing namespace replaced", err)
				}
				raw, err := os.ReadFile(state.LedgerPath(f.m.home))
				if err != nil || string(raw) != string(ledger) {
					t.Fatal("existing ledger changed", err)
				}
			}
			unchanged(t, before, f.m.home)
			if prior == "existing-ledger" {
				acquire(t, f)
			}
		})
	}
}
