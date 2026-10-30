package homelease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
)

type fixtureAuthority struct{ home, digest string }

func (f *fixtureAuthority) Admit(_ context.Context, held managerlock.Held, environment, profile, digest string) (Admission, error) {
	if err := held.Validate(f.home); err != nil {
		return Admission{}, err
	}
	if digest != f.digest {
		return Admission{}, state.Refuse("environment_lease_lock_mismatch", "profile_lock_changed")
	}
	actual := filepath.Join(f.home, "environments", profile, environment)
	return Admission{Home: actual, Protected: []string{actual, filepath.Join(f.home, "profiles", profile)}, BuildKeys: []string{}, RuntimeEntries: []string{}}, nil
}

type fixtureObserver struct {
	proof Evidence
	err   error
}

func (f *fixtureObserver) Observe(context.Context, Holder) (Evidence, error) { return f.proof, f.err }

type fixture struct {
	m   *Manager
	a   *fixtureAuthority
	o   *fixtureObserver
	req Request
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	home := filepath.Join(t.TempDir(), "manager")
	if err := privatedir.MakeAll(filepath.Join(home, "environments", "sample", "claude_code")); err != nil {
		t.Fatal(err)
	}
	a := &fixtureAuthority{home: home, digest: strings.Repeat("a", 64)}
	o := &fixtureObserver{}
	m, err := New(home, a, o)
	if err != nil {
		t.Fatal(err)
	}
	return fixture{m: m, a: a, o: o, req: Request{Environment: "claude_code", Profile: "sample", LockSHA256: a.digest, RequestID: "request-1", Holder: Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "manager-1", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "request-1"}}}
}
func acquire(t *testing.T, f fixture) *Record {
	t.Helper()
	out, reason, err := f.m.Acquire(context.Background(), f.req)
	if err != nil || reason != "acquired" || !out.Changed {
		t.Fatalf("acquire: %+v %s %v", out, reason, err)
	}
	return out.Lease
}
func requireRefusal(t *testing.T, err error, code, reason string) {
	t.Helper()
	var refusal *Refusal
	if !errors.As(err, &refusal) || refusal.Code != code || refusal.Reason != reason {
		t.Fatalf("wanted %s/%s, got %v", code, reason, err)
	}
}
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = i.Mode().String()
		if d.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(p)
			out[rel] += target
			return err
		}
		if !d.IsDir() {
			b, err := os.ReadFile(p)
			out[rel] += string(b)
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func unchanged(t *testing.T, before map[string]string, root string) {
	t.Helper()
	if after := snapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("preimage changed: before=%v after=%v", before, after)
	}
}
func guard(t *testing.T, f fixture, targets []string) error {
	t.Helper()
	locks, err := managerlock.New(f.m.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := locks.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	return state.Guard(f.m.home, lock.Held(), targets)
}

// HL03/06/10: drift and changed replay refuse without state writes; identical
// replay preserves the current reservation, and an absent release is a no-op.
func TestAcquireReplayAndDigest(t *testing.T) {
	f := newFixture(t)
	first := acquire(t, f)
	before := snapshot(t, f.m.home)
	out, reason, err := f.m.Acquire(context.Background(), f.req)
	if err != nil || reason != "acquired" || out.Changed || !reflect.DeepEqual(first, out.Lease) {
		t.Fatalf("lost reply: %+v %s %v", out, reason, err)
	}
	unchanged(t, before, f.m.home)
	changed := f.req
	changed.LockSHA256 = strings.Repeat("b", 64)
	_, _, err = f.m.Acquire(context.Background(), changed)
	requireRefusal(t, err, "environment_lease_holder_mismatch", "request_conflict")
	unchanged(t, before, f.m.home)
	changed.RequestID = "request-2"
	changed.Holder.RequestID = changed.RequestID
	_, _, err = f.m.Acquire(context.Background(), changed)
	requireRefusal(t, err, "environment_lease_lock_mismatch", "profile_lock_changed")
	unchanged(t, before, f.m.home)
	out, reason, err = f.m.Release(context.Background(), "absent")
	if err != nil || reason != "unknown_lease" || out.Changed || out.Lease != nil {
		t.Fatalf("unknown release: %+v %s %v", out, reason, err)
	}
	unchanged(t, before, f.m.home)
	f.req.RequestID = "request-2"
	f.req.Holder.RequestID = f.req.RequestID
	second := acquire(t, f)
	if second.LeaseID == first.LeaseID {
		t.Fatal("new request reused id")
	}
}

// HL06/07/12/13: terminal acquire replay has zero effects and cannot revive a
// holder. Break requires invocation authority and CAS, with one permanent event.
func TestTerminalReplayAndAuditedBreak(t *testing.T) {
	for _, operation := range []string{"release", "break"} {
		t.Run(operation, func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			before := snapshot(t, f.m.home)
			_, _, err := f.m.Break(context.Background(), r.LeaseID, 1, false, "protection deliberately abandoned")
			requireRefusal(t, err, "usage", "operator_confirmation_required")
			unchanged(t, before, f.m.home)
			_, _, err = f.m.Break(context.Background(), r.LeaseID, 2, true, "protection deliberately abandoned")
			requireRefusal(t, err, "environment_lease_revision_conflict", "revision_changed")
			unchanged(t, before, f.m.home)
			var out Outcome
			var reason string
			if operation == "break" {
				out, reason, err = f.m.Break(context.Background(), r.LeaseID, 1, true, "protection deliberately abandoned")
			} else {
				f.o.proof = Evidence{Holder: r.Holder, NoStart: true, DescendantsTerminated: true, OwnershipReconciled: true}
				out, reason, err = f.m.Release(context.Background(), r.LeaseID)
			}
			if err != nil || !out.Changed {
				t.Fatalf("terminal: %+v %s %v", out, reason, err)
			}
			terminalRecord := out.Lease
			before = snapshot(t, f.m.home)
			out, reason, err = f.m.Acquire(context.Background(), f.req)
			want := "already_released"
			if operation == "break" {
				want = "already_broken"
			}
			if err != nil || reason != want || out.Changed || !reflect.DeepEqual(out.Lease, terminalRecord) {
				t.Fatalf("terminal replay: %+v %s %v", out, reason, err)
			}
			unchanged(t, before, f.m.home)
			rows, err := f.m.List(context.Background(), "", "")
			if err != nil || len(rows) != 1 || !reflect.DeepEqual(rows[0], *terminalRecord) {
				t.Fatalf("list: %+v %v", rows, err)
			}
			unchanged(t, before, f.m.home)
			if err := guard(t, f, nil); err != nil {
				t.Fatal(err)
			}
			events, err := os.ReadDir(filepath.Join(state.Namespace(f.m.home), "events"))
			if err != nil || len(events) != 1 {
				t.Fatalf("audit: %v %v", events, err)
			}
			if operation == "break" {
				out, reason, err = f.m.Break(context.Background(), r.LeaseID, 1, true, "protection deliberately abandoned")
				if err != nil || reason != "already_broken" || out.Changed {
					t.Fatalf("break replay: %+v %s %v", out, reason, err)
				}
				unchanged(t, before, f.m.home)
			}
		})
	}
}

// HL07/08/09/11/19: only independently matching evidence binds/releases. A
// dead observer cannot clear protection; every holder appears in refusal.
func TestEvidenceAndAllHolders(t *testing.T) {
	f := newFixture(t)
	first := acquire(t, f)
	f.req.RequestID = "request-2"
	f.req.Holder.RequestID = f.req.RequestID
	second := acquire(t, f)
	session := "SES-one"
	generation := int64(1)
	bound := first.Holder
	bound.Kind = "bound"
	bound.Session = &session
	bound.Generation = &generation
	bound.Process = &state.Process{PID: 42, StartedAt: "2026-10-27T00:00:01Z"}
	before := snapshot(t, f.m.home)
	f.o.err = errors.New("launcher dead")
	_, _, err := f.m.Bind(context.Background(), first.LeaseID, 1, bound)
	requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
	unchanged(t, before, f.m.home)
	f.o.err = nil
	f.o.proof = Evidence{Holder: bound}
	out, reason, err := f.m.Bind(context.Background(), first.LeaseID, 1, bound)
	if err != nil || reason != "bound" || !out.Changed {
		t.Fatalf("bind: %+v %s %v", out, reason, err)
	}
	before = snapshot(t, f.m.home)
	out, reason, err = f.m.Bind(context.Background(), first.LeaseID, 1, bound)
	if err != nil || reason != "bound" || out.Changed {
		t.Fatalf("bind replay: %+v %s %v", out, reason, err)
	}
	unchanged(t, before, f.m.home)
	_, _, err = f.m.Release(context.Background(), first.LeaseID)
	requireRefusal(t, err, "environment_lease_release_unconfirmed", "termination_unconfirmed")
	unchanged(t, before, f.m.home)
	err = guard(t, f, []string{filepath.Join(f.m.home, "environments", "sample", "claude_code", "file")})
	requireRefusal(t, err, "environment_home_leased", "blocking_holders")
	var refusal *Refusal
	_ = errors.As(err, &refusal)
	if refusal.Details.Count != 2 {
		t.Fatalf("lost holder: %+v", refusal)
	}
	unchanged(t, before, f.m.home)
	if err := guard(t, f, []string{filepath.Join(f.m.home, "environments", "other", "claude_code")}); err != nil {
		t.Fatal(err)
	}
	f.o.proof = Evidence{Holder: bound, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: true}
	if _, _, err := f.m.Release(context.Background(), first.LeaseID); err != nil {
		t.Fatal(err)
	}
	err = guard(t, f, nil)
	requireRefusal(t, err, "environment_home_leased", "blocking_holders")
	f.o.proof = Evidence{Holder: second.Holder, NoStart: true, DescendantsTerminated: true, OwnershipReconciled: true}
	if _, _, err := f.m.Release(context.Background(), second.LeaseID); err != nil {
		t.Fatal(err)
	}
	if err := guard(t, f, nil); err != nil {
		t.Fatal(err)
	}
}

// HL04/17: absent list creates no home, while a partially initialized namespace
// is uncertain and cannot authorize a successful absent-ID release.
func TestReadOnlyAbsenceAndMissingLedger(t *testing.T) {
	home := filepath.Join(t.TempDir(), "absent")
	m, err := New(home, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := m.List(context.Background(), "", "")
	if err != nil || len(rows) != 0 {
		t.Fatalf("absent list: %v %v", rows, err)
	}
	if _, err := os.Lstat(home); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("list created state: %v", err)
	}
	f := newFixture(t)
	acquire(t, f)
	if err := os.Remove(state.LedgerPath(f.m.home)); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, f.m.home)
	_, _, err = f.m.Release(context.Background(), "absent")
	requireRefusal(t, err, "environment_lease_unreadable", "ledger_missing")
	unchanged(t, before, f.m.home)
}
