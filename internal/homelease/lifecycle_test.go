package homelease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
)

func boundHolder(h Holder) Holder {
	session, generation := "fixture-session", int64(7)
	h.Kind, h.Session, h.Generation = "bound", &session, &generation
	h.Process = &state.Process{PID: 123, StartedAt: "2026-10-27T00:00:01Z"}
	return h
}

// HL07/09: every mismatching identity and incomplete termination proof keeps
// the real reservation and audit unchanged. Matching independent evidence is
// the nearby positive control; this fixture does not qualify a live observer.
func TestIndependentEvidenceRefusals(t *testing.T) {
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	before := snapshot(t, f.m.home)
	wrong := h
	wrong.ManagerRegisteredAt = "2026-10-27T00:00:02Z"
	_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, wrong)
	requireRefusal(t, err, "environment_lease_holder_mismatch", "manager_incarnation_changed")
	unchanged(t, before, f.m.home)
	f.o.proof.Holder = wrong
	_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_holder_mismatch", "session_generation_changed")
	unchanged(t, before, f.m.home)
	f.o.proof.Holder = h
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed || out.Lease.Revision != 2 {
		t.Fatalf("bind control: %+v %s %v", out, reason, err)
	}
	before = snapshot(t, f.m.home)
	wrong = h
	generation := int64(8)
	wrong.Generation = &generation
	_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, wrong)
	requireRefusal(t, err, "environment_lease_revision_conflict", "revision_changed")
	unchanged(t, before, f.m.home)
	for _, proof := range []Evidence{
		{Holder: wrong, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: true},
		{Holder: h, Terminal: true, DescendantsTerminated: false, OwnershipReconciled: true},
		{Holder: h, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: false},
		{Holder: h, NoStart: true, DescendantsTerminated: true, OwnershipReconciled: true},
	} {
		f.o.proof = proof
		_, _, err = f.m.Release(context.Background(), r.LeaseID)
		want := "termination_unconfirmed"
		if !reflect.DeepEqual(proof.Holder, h) {
			want = "session_state_unknown"
		}
		requireRefusal(t, err, "environment_lease_release_unconfirmed", want)
		unchanged(t, before, f.m.home)
	}
	f.o.proof = Evidence{Holder: h, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: true}
	out, reason, err = f.m.Release(context.Background(), r.LeaseID)
	if err != nil || reason != "released" || !out.Changed || out.Lease.Revision != 3 {
		t.Fatalf("release control: %+v %s %v", out, reason, err)
	}
}

// HL07 (corpus 65/68): historical successful bind replay never outranks either
// terminal disposition. Its nonterminal exact retry is the no-write control.
func TestTerminalDispositionBeforeHistoricalBind(t *testing.T) {
	for _, operation := range []string{"release", "break"} {
		t.Run(operation, func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			h := boundHolder(r.Holder)
			f.o.proof.Holder = h
			if _, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h); err != nil {
				t.Fatal(err)
			}
			before := snapshot(t, f.m.home)
			out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
			if err != nil || reason != "bound" || out.Changed {
				t.Fatalf("nonterminal retry: %+v %s %v", out, reason, err)
			}
			unchanged(t, before, f.m.home)
			if operation == "break" {
				_, _, err = f.m.Break(context.Background(), r.LeaseID, 2, true, "protection deliberately abandoned")
			} else {
				f.o.proof = Evidence{Holder: h, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: true}
				_, _, err = f.m.Release(context.Background(), r.LeaseID)
			}
			if err != nil {
				t.Fatal(err)
			}
			before = snapshot(t, f.m.home)
			_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
			requireRefusal(t, err, "environment_lease_revision_conflict", "revision_changed")
			unchanged(t, before, f.m.home)
		})
	}
}

// HL04: malformed, linked, case-aliased and partially published state cannot
// authorize absent-ID release. The complete namespace's absent release is the
// positive no-write control before each isolated corruption plant.
func TestUncertainNamespaceRefusesAbsentRelease(t *testing.T) {
	plants := map[string]func(*testing.T, fixture){
		"malformed": func(t *testing.T, f fixture) { writeFixtureState(t, state.LedgerPath(f.m.home), []byte(`{}`)) },
		"orphan-binding": func(t *testing.T, f fixture) {
			writeFixtureState(t, state.BindingPath(f.m.home, "orphan"), []byte(`{}`))
		},
		"unknown-file": func(t *testing.T, f fixture) {
			writeFixtureState(t, filepath.Join(state.Namespace(f.m.home), "unrecognized"), []byte(`{}`))
		},
		"case-alias": func(t *testing.T, f fixture) {
			raw, err := os.ReadFile(state.LedgerPath(f.m.home))
			if err != nil {
				t.Fatal(err)
			}
			writeFixtureState(t, state.LedgerPath(f.m.home), []byte(strings.Replace(string(raw), `"kind"`, `"Kind"`, 1)))
		},
		"symlink": func(t *testing.T, f fixture) {
			path := state.LedgerPath(f.m.home)
			if err := os.Rename(path, path+".original"); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(path+".original", path); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
		},
	}
	for name, plant := range plants {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			acquire(t, f)
			before := snapshot(t, f.m.home)
			out, reason, err := f.m.Release(context.Background(), "missing")
			if err != nil || reason != "unknown_lease" || out.Changed {
				t.Fatalf("complete control: %+v %s %v", out, reason, err)
			}
			unchanged(t, before, f.m.home)
			plant(t, f)
			before = snapshot(t, f.m.home)
			_, _, err = f.m.Release(context.Background(), "missing")
			var refusal *Refusal
			if !errors.As(err, &refusal) || refusal.Code != "environment_lease_unreadable" {
				t.Fatalf("uncertain authority admitted absent release: %v", err)
			}
			unchanged(t, before, f.m.home)
		})
	}
}

func writeFixtureState(t *testing.T, path string, raw []byte) {
	t.Helper()
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

// HL02: bounded real OS-lock contention returns the exact refusal without
// lease or journal side effects; releasing the lock admits the same request.
func TestAcquireContention(t *testing.T) {
	f := newFixture(t)
	locks, err := managerlock.New(f.m.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := locks.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	before := snapshot(t, f.m.home)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, _, err = f.m.Acquire(ctx, f.req)
	requireRefusal(t, err, "environment_lock_unavailable", "mutation_lock_busy")
	unchanged(t, before, f.m.home)
	_ = lock.Close()
	acquire(t, f)
}

// HL08/11/19: prepared, active and persisted unknown holders are all named,
// ordered and counted. Unavailable liveness never changes or releases any of
// them; explicit independent terminal disposition of all three admits writes.
func TestEveryHolderNamedIncludingUnknown(t *testing.T) {
	f := newFixture(t)
	first := acquire(t, f)
	f.req.RequestID, f.req.Holder.RequestID = "request-2", "request-2"
	second := acquire(t, f)
	h := boundHolder(second.Holder)
	f.o.proof.Holder = h
	bound, _, err := f.m.Bind(context.Background(), second.LeaseID, 1, h)
	if err != nil {
		t.Fatal(err)
	}
	f.req.RequestID, f.req.Holder.RequestID = "request-3", "request-3"
	third := acquire(t, f)
	l, err := state.Load(f.m.home)
	if err != nil {
		t.Fatal(err)
	}
	e := l.Leases[third.LeaseID]
	e.Record.Phase = "unknown"
	at := "2026-10-27T00:00:03Z"
	e.Record.Observation = state.Observation{State: "unknown", ObservedAt: &at}
	l.Leases[third.LeaseID] = e
	raw, err := state.Canonical(l)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureState(t, state.LedgerPath(f.m.home), raw)
	f.o.err = errors.New("fixture launcher is dead and manager is unavailable")
	before := snapshot(t, f.m.home)
	_, _, err = f.m.Release(context.Background(), third.LeaseID)
	requireRefusal(t, err, "environment_lease_release_unconfirmed", "session_state_unknown")
	unchanged(t, before, f.m.home)
	err = guard(t, f, nil)
	var refusal *Refusal
	if !errors.As(err, &refusal) || refusal.Details.Count != 3 || len(refusal.Details.Holders) != 3 {
		t.Fatalf("lost holder: %v", err)
	}
	rows, err := f.m.List(context.Background(), "", "")
	if err != nil || !reflect.DeepEqual(rows, refusal.Details.Holders) {
		t.Fatalf("unstable holder order: %+v %v", rows, err)
	}
	for _, r := range rows {
		if !strings.Contains(refusal.Error(), r.LeaseID) || !r.Blocking() {
			t.Fatalf("omitted holder %s", r.LeaseID)
		}
	}
	unchanged(t, before, f.m.home)
	for _, r := range []*Record{first, bound.Lease, third} {
		if _, _, err := f.m.Break(context.Background(), r.LeaseID, r.Revision, true, "protection deliberately abandoned"); err != nil {
			t.Fatal(err)
		}
	}
	if err := guard(t, f, nil); err != nil {
		t.Fatalf("terminal controls still blocked: %v", err)
	}
}

// HL15: the legacy migration owner's pending intent also blocks fresh
// allocation and absent-ID reconciliation. Only removing the isolated pending
// fixture after owner reconciliation permits a real new acquire.
func TestPendingLegacyMigration(t *testing.T) {
	f := newFixture(t)
	acquire(t, f)
	path := filepath.Join(f.m.home, "state", "env-migration.json")
	writeFixtureState(t, path, []byte(`{"version":1,"plan":"fixture-pending","ops":[]}`))
	before := snapshot(t, f.m.home)
	f.req.RequestID, f.req.Holder.RequestID = "after-pending", "after-pending"
	_, _, err := f.m.Acquire(context.Background(), f.req)
	requireRefusal(t, err, "environment_lease_recovery_required", "pending_mutation")
	unchanged(t, before, f.m.home)
	_, _, err = f.m.Release(context.Background(), "absent")
	requireRefusal(t, err, "environment_lease_recovery_required", "pending_mutation")
	unchanged(t, before, f.m.home)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	acquire(t, f)
}
