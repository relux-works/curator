package homelease

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
)

// Corpus 1/3/44: absent materializer authority and stale digests cannot allocate
// another reservation. Canonical alias replay is identical; after an explicit
// break, newly admitted authority can acquire the new exact digest. This proves
// the real store boundary, not the launcher's persisted-operation ordering.
func TestSemanticAcquireAuthorityAndAlias(t *testing.T) {
	f := newFixture(t)
	r := acquire(t, f)
	before := snapshot(t, f.m.home)
	f.m.authority = nil
	f.req.RequestID, f.req.Holder.RequestID = "next-request", "next-request"
	_, _, err := f.m.Acquire(context.Background(), f.req)
	requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
	unchanged(t, before, f.m.home)
	f.m.authority = f.a
	f.req.RequestID, f.req.Holder.RequestID = "request-1", "request-1"
	f.req.Environment = "claude"
	out, reason, err := f.m.Acquire(context.Background(), f.req)
	if err != nil || reason != "acquired" || out.Changed || !reflect.DeepEqual(out.Lease, r) {
		t.Fatalf("canonical alias replay: %+v %s %v", out, reason, err)
	}
	unchanged(t, before, f.m.home)
	if _, _, err := f.m.Break(context.Background(), r.LeaseID, 1, true, "fixture protection deliberately abandoned"); err != nil {
		t.Fatal(err)
	}
	f.a.digest = strings.Repeat("b", 64)
	f.req.RequestID, f.req.Holder.RequestID = "next-request", "next-request"
	before = snapshot(t, f.m.home)
	_, _, err = f.m.Acquire(context.Background(), f.req)
	requireRefusal(t, err, "environment_lease_lock_mismatch", "profile_lock_changed")
	unchanged(t, before, f.m.home)
	f.req.LockSHA256 = f.a.digest
	got := acquire(t, f)
	if got.Environment != "claude_code" || got.LockSHA256 != f.a.digest {
		t.Fatal("fresh admission substituted identity or digest")
	}
}

// Corpus 49: an absent ID has the exact bind refusal and no side effects;
// the same independently observed tuple binds the real prepared ID once.
func TestSemanticUnknownBind(t *testing.T) {
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	f.o.proof.Holder = h
	before := snapshot(t, f.m.home)
	_, _, err := f.m.Bind(context.Background(), "absent", 1, h)
	requireRefusal(t, err, "environment_lease_unknown", "unknown_lease")
	unchanged(t, before, f.m.home)
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed {
		t.Fatalf("existing-ID control: %+v %s %v", out, reason, err)
	}
}

// Corpus 63/66, Curator half only: terminal replay is read-only, then a new
// request acquires a different lease without changing prior binding/audit or
// terminal record. Zero manager calls/op.json settlement belong to curator-run
// and are deliberately not simulated by a made-up consumer in this test.
func TestSemanticFreshRequestAfterTerminalReplay(t *testing.T) {
	for _, operation := range []string{"release", "break"} {
		t.Run(operation, func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			if operation == "release" {
				f.o.proof = Evidence{Holder: r.Holder, NoStart: true, DescendantsTerminated: true, OwnershipReconciled: true}
				if _, _, err := f.m.Release(context.Background(), r.LeaseID); err != nil {
					t.Fatal(err)
				}
			} else if _, _, err := f.m.Break(context.Background(), r.LeaseID, 1, true, "fixture protection deliberately abandoned"); err != nil {
				t.Fatal(err)
			}
			before := snapshot(t, f.m.home)
			out, reason, err := f.m.Acquire(context.Background(), f.req)
			if err != nil || reason != "already_"+map[string]string{"release": "released", "break": "broken"}[operation] || out.Changed {
				t.Fatalf("terminal replay: %+v %s %v", out, reason, err)
			}
			unchanged(t, before, f.m.home)
			oldRecord := *out.Lease
			binding, err := os.ReadFile(state.BindingPath(f.m.home, r.LeaseID))
			if err != nil {
				t.Fatal(err)
			}
			events := snapshot(t, filepath.Join(state.Namespace(f.m.home), "events"))
			f.req.RequestID, f.req.Holder.RequestID = "fresh-request", "fresh-request"
			fresh := acquire(t, f)
			if fresh.LeaseID == r.LeaseID || fresh.Holder.RequestID == r.Holder.RequestID {
				t.Fatal("fresh admission revived old identity")
			}
			got, err := os.ReadFile(state.BindingPath(f.m.home, r.LeaseID))
			if err != nil || !reflect.DeepEqual(got, binding) {
				t.Fatal("immutable binding changed")
			}
			unchanged(t, events, filepath.Join(state.Namespace(f.m.home), "events"))
			rows, err := f.m.List(context.Background(), "", "")
			if err != nil || len(rows) != 2 {
				t.Fatalf("two independent dispositions: %v %v", rows, err)
			}
			found := false
			for _, row := range rows {
				if row.LeaseID == r.LeaseID {
					found = reflect.DeepEqual(row, oldRecord)
				}
			}
			if !found {
				t.Fatal("terminal record changed")
			}
		})
	}
}

// Corpus 11/44: a current authority digest cannot supersede a different
// generation retained by either of two real store records. Finishing only one
// holder still blocks; finishing both permits a fresh exact-digest reservation.
func TestSemanticIndependentGenerationHolders(t *testing.T) {
	f := newFixture(t)
	a := acquire(t, f)
	f.req.RequestID, f.req.Holder.RequestID = "request-b", "request-b"
	b := acquire(t, f)
	f.o.proof = Evidence{Holder: a.Holder, NoStart: true, DescendantsTerminated: true, OwnershipReconciled: true}
	if _, _, err := f.m.Release(context.Background(), a.LeaseID); err != nil {
		t.Fatal(err)
	}
	f.a.digest = strings.Repeat("b", 64)
	f.req.LockSHA256 = f.a.digest
	f.req.RequestID, f.req.Holder.RequestID = "generation-b", "generation-b"
	before := snapshot(t, f.m.home)
	_, _, err := f.m.Acquire(context.Background(), f.req)
	requireRefusal(t, err, "environment_lease_lock_mismatch", "profile_lock_changed")
	unchanged(t, before, f.m.home)
	err = guard(t, f, nil)
	requireRefusal(t, err, "environment_home_leased", "blocking_holders")
	if refusal := err.(*state.Refusal); refusal.Details.Count != 1 || refusal.Details.Holders[0].LeaseID != b.LeaseID {
		t.Fatal("remaining holder lost")
	}
	unchanged(t, before, f.m.home)
	f.o.proof.Holder = b.Holder
	if _, _, err := f.m.Release(context.Background(), b.LeaseID); err != nil {
		t.Fatal(err)
	}
	if err := guard(t, f, nil); err != nil {
		t.Fatalf("all-terminal writer control: %v", err)
	}
	c := acquire(t, f)
	if c.LockSHA256 != f.a.digest || c.LeaseID == a.LeaseID || c.LeaseID == b.LeaseID {
		t.Fatal("new generation not independently acquired")
	}
}

// Corpus 7/22/47: no field of the caller's proposed tuple supplies independent
// bind authority. Each mismatching observer tuple preserves the store before
// a matching synthetic observer proves the store's positive transition. This
// exercises lifecycle decisions, not a production manager observation adapter.
func TestSemanticEveryObservedHolderField(t *testing.T) {
	for name, change := range map[string]func(*Holder){
		"manager":       func(h *Holder) { h.ManagerID = "different-manager" },
		"registration":  func(h *Holder) { h.ManagerRegisteredAt = "2026-10-27T00:01:00Z" },
		"request":       func(h *Holder) { h.RequestID = "different-request" },
		"werkstatt":     func(h *Holder) { s := "different-werkstatt"; h.Werkstatt = &s },
		"session":       func(h *Holder) { s := "different-session"; h.Session = &s },
		"generation":    func(h *Holder) { n := int64(8); h.Generation = &n },
		"pid":           func(h *Holder) { h.Process = &state.Process{PID: 124, StartedAt: h.Process.StartedAt} },
		"process-start": func(h *Holder) { h.Process = &state.Process{PID: h.Process.PID, StartedAt: "2026-10-27T00:00:02Z"} },
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			h := boundHolder(r.Holder)
			observed := h
			change(&observed)
			f.o.proof.Holder = observed
			before := snapshot(t, f.m.home)
			_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
			requireRefusal(t, err, "environment_lease_holder_mismatch", "session_generation_changed")
			unchanged(t, before, f.m.home)
			f.o.proof.Holder = h
			out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
			if err != nil || reason != "bound" || !out.Changed || out.Lease.Revision != 2 {
				t.Fatalf("exact tuple control: %+v %s %v", out, reason, err)
			}
		})
	}
}

// HL07/09: unavailable production integration keeps both prepared and bound
// leases intact; availability is never inferred from a well-formed holder.
// A separately supplied synthetic observer provides the paired lifecycle
// control without claiming that the external observe protocol carries proof.
func TestSemanticUnavailableObserver(t *testing.T) {
	for _, bound := range []bool{false, true} {
		t.Run(map[bool]string{false: "prepared", true: "bound"}[bound], func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			h := boundHolder(r.Holder)
			if bound {
				f.o.proof.Holder = h
				out, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
				if err != nil {
					t.Fatal(err)
				}
				r = out.Lease
			}
			f.m.observer = nil
			before := snapshot(t, f.m.home)
			if !bound {
				_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
				requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
				unchanged(t, before, f.m.home)
			}
			_, _, err := f.m.Release(context.Background(), r.LeaseID)
			requireRefusal(t, err, "environment_lease_release_unconfirmed", "session_state_unknown")
			unchanged(t, before, f.m.home)
			f.m.observer = f.o
			f.o.proof = Evidence{Holder: r.Holder, NoStart: !bound, Terminal: bound, DescendantsTerminated: true, OwnershipReconciled: true}
			out, reason, err := f.m.Release(context.Background(), r.LeaseID)
			if err != nil || reason != "released" || !out.Changed {
				t.Fatalf("independent evidence control: %+v %s %v", out, reason, err)
			}
		})
	}
}
