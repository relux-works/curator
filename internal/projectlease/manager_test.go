package projectlease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

type authorities struct {
	home      string
	witnesses []managerlock.Held
	mu        sync.Mutex
	comps     map[string]Composition
	sessions  map[string]Evidence
	revoked   bool
	uncertain bool
}

func (a *authorities) Resolve(_ context.Context, held managerlock.Held, ref, sha string) (Composition, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.witnesses = append(a.witnesses, held)
	if err := held.Validate(a.home); err != nil {
		return Composition{}, err
	}
	c, ok := a.comps[ref]
	if !ok || c.SHA256 != sha {
		return c, errors.New("unknown protected composition")
	}
	return c, nil
}
func (a *authorities) Authorize(_ context.Context, held managerlock.Held, _ Composition, _ string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.witnesses = append(a.witnesses, held)
	if err := held.Validate(a.home); err != nil {
		return err
	}
	if a.uncertain {
		return errors.New("authority unreadable")
	}
	if a.revoked {
		return ErrRevoked
	}
	return nil
}
func (a *authorities) Observe(_ context.Context, ref string, _ int64) (Evidence, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, ok := a.sessions[ref]
	if !ok {
		return e, errors.New("session liveness unknown")
	}
	return e, nil
}

type fixture struct {
	m     *Manager
	a     *authorities
	files map[string][]byte
	c     Composition
	home  string
}

func setup(t *testing.T) fixture {
	t.Helper()
	home := t.TempDir()
	if err := privatedir.Protect(home); err != nil {
		t.Fatal(err)
	}
	return atHome(t, home)
}
func atHome(t *testing.T, home string) fixture {
	t.Helper()
	files := map[string][]byte{"AGENTS.md": []byte("approved\n")}
	c := Composition{Ref: "composition:first", SHA256: strings.Repeat("a", 64), CheckoutID: "checkout-a", Profile: "default", Environment: "codex_cli", HomeRef: "home:first", ConcurrentState: true, BuildKeys: []string{"sha256:" + strings.Repeat("b", 64)}, RuntimeEntries: []string{"tool/revision"}, Generated: inventory(files)}
	// These roots stand for artifacts already verified by the fixture authority;
	// receipt/hash qualification belongs to the composition/cache implementation.
	for _, path := range []string{filepath.Join(home, "runtime", "tool", "revision"), filepath.Join(home, "cache", "build", "go-v1", strings.Repeat("b", 64))} {
		if err := privatedir.MakeAll(path); err != nil {
			t.Fatal(err)
		}
	}
	a := &authorities{home: home, comps: map[string]Composition{c.Ref: c}, sessions: map[string]Evidence{}}
	m, err := New(home, a, a)
	if err != nil {
		t.Fatal(err)
	}
	return fixture{m, a, files, c, home}
}
func (f fixture) reserve(id string) Request {
	r := Request{LeaseID: id, ContextRef: f.c.Ref, ContextSHA256: f.c.SHA256, OwnerRef: "principal:operator", SessionRef: "session:" + id, Generation: 1}
	f.a.mu.Lock()
	f.a.sessions[r.SessionRef] = Evidence{SessionRef: r.SessionRef, Generation: 1, OwnerRef: r.OwnerRef, NoChildStarted: true}
	f.a.mu.Unlock()
	return r
}
func acquire(t *testing.T, f fixture, id string) Binding {
	t.Helper()
	b, err := f.m.Acquire(context.Background(), f.reserve(id))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func (f fixture) observe(id string, edit func(*Evidence)) {
	f.a.mu.Lock()
	defer f.a.mu.Unlock()
	e := f.a.sessions["session:"+id]
	edit(&e)
	f.a.sessions[e.SessionRef] = e
}
func activate(t *testing.T, f fixture, id string) Binding {
	t.Helper()
	b := acquire(t, f, id)
	if err := f.m.PublishPrepared(context.Background(), id, f.files, false); err != nil {
		t.Fatal(err)
	}
	f.observe(id, func(e *Evidence) { e.Running = true; e.ProcessGeneration = "boot-session-runner-1" })
	if err := f.m.Transfer(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	return b
}
func status(t *testing.T, f fixture, id string) Status {
	t.Helper()
	rows, err := f.m.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.Binding.LeaseID == id {
			return r
		}
	}
	t.Fatalf("missing status %s", id)
	return Status{}
}
func mustRefuse(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), code) {
		t.Fatalf("refusal %v, want %s", err, code)
	}
}

// Two checkouts on one profile have distinct stable homes; identical sessions
// share bytes only with the protected adapter's concurrent-state capability.
func TestStableHomesAndConcurrentSessions(t *testing.T) {
	f := setup(t)
	activate(t, f, "one")
	acquire(t, f, "two")
	before, err := os.ReadFile(filepath.Join(f.home, homeRelative(f.c), "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.m.PublishPrepared(context.Background(), "two", f.files, false); err != nil {
		t.Fatal(err)
	}
	f.observe("two", func(e *Evidence) { e.Running = true; e.ProcessGeneration = "another-runner" })
	if err = f.m.Transfer(context.Background(), "two"); err != nil {
		t.Fatal(err)
	}
	c := f.c
	c.Ref = "composition:second"
	c.SHA256 = strings.Repeat("c", 64)
	c.CheckoutID = "checkout-b"
	c.HomeRef = "home:second"
	f.a.comps[c.Ref] = c
	other := f
	other.c = c
	activate(t, other, "three")
	if homeRelative(c) == homeRelative(f.c) {
		t.Fatal("two projects alias")
	}
	after, _ := os.ReadFile(filepath.Join(f.home, homeRelative(f.c), "AGENTS.md"))
	if string(before) != string(after) {
		t.Fatal("changed active home")
	}
	c = f.c
	c.Ref = "composition:exclusive"
	c.SHA256 = strings.Repeat("e", 64)
	c.CheckoutID = "checkout-c"
	c.HomeRef = "home:third"
	c.ConcurrentState = false
	f.a.comps[c.Ref] = c
	exclusive := f
	exclusive.c = c
	activate(t, exclusive, "four")
	_, err = f.m.Acquire(context.Background(), exclusive.reserve("five"))
	mustRefuse(t, err, Busy)
	if status(t, f, "one").State != Active || status(t, f, "two").State != Active {
		t.Fatal("shared leases lost")
	}
}

// Changed compositions and all repair/update attempts refuse while leased,
// preserving generated bytes and native login. An idle update reuses the home.
func TestBusyUpdateAndIdleReuse(t *testing.T) {
	f := setup(t)
	activate(t, f, "one")
	home := filepath.Join(f.home, homeRelative(f.c))
	login := filepath.Join(home, "auth.json")
	if err := os.WriteFile(login, []byte("fixture-native-state"), 0600); err != nil {
		t.Fatal(err)
	}
	c := f.c
	c.Ref = "composition:next"
	c.SHA256 = strings.Repeat("d", 64)
	next := map[string][]byte{"AGENTS.md": []byte("next\n")}
	c.Generated = inventory(next)
	f.a.comps[c.Ref] = c
	_, err := f.m.Acquire(context.Background(), func() Request { r := f.reserve("two"); r.ContextRef = c.Ref; r.ContextSHA256 = c.SHA256; return r }())
	mustRefuse(t, err, Busy)
	mustRefuse(t, f.m.Update(context.Background(), c.Ref, c.SHA256, "principal:operator", next, false), Busy)
	mustRefuse(t, f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, false), Busy)
	raw, _ := os.ReadFile(filepath.Join(home, "AGENTS.md"))
	if string(raw) != "approved\n" {
		t.Fatal("busy home changed")
	}
	f.observe("one", func(e *Evidence) {
		e.Running = false
		e.Terminated = true
		e.OwnershipReconciled = true
		e.DescendantsTerminated = true
	})
	if err = f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if err = f.m.Update(context.Background(), c.Ref, c.SHA256, "principal:operator", next, false); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(login)
	if string(raw) != "fixture-native-state" {
		t.Fatal("native login changed")
	}
	raw, _ = os.ReadFile(filepath.Join(home, "AGENTS.md"))
	if string(raw) != "next\n" {
		t.Fatal("idle update absent")
	}
}

// Wait releases the home lock between checks: a terminated/reconciled session
// can unblock an update. Cancellation preserves bytes and names the session.
func TestBusyWaitAndCancellation(t *testing.T) {
	f := setup(t)
	activate(t, f, "one")
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	err := f.m.Update(ctx, f.c.Ref, f.c.SHA256, "principal:operator", f.files, true)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wait=%v", err)
	}
	mustRefuse(t, err, "session:one")
	done := make(chan error, 1)
	go func() {
		done <- f.m.Update(context.Background(), f.c.Ref, f.c.SHA256, "principal:operator", f.files, true)
	}()
	f.observe("one", func(e *Evidence) {
		e.Running = false
		e.Terminated = true
		e.OwnershipReconciled = true
		e.DescendantsTerminated = true
	})
	if err = f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("wait never completed")
	}
}

// A restarted launching manager and detach do not release the durable runner
// binding. Descendants, uncertain liveness, aged records and reused process
// identities retain it; only the exact reconciled session can release it.
func TestDetachDescendantsUnknownAndPIDReuse(t *testing.T) {
	f := setup(t)
	b := activate(t, f, "one")
	immutable, _ := os.ReadFile(bindingPath(f.home, "one"))
	m, err := New(f.home, f.a, f.a)
	if err != nil {
		t.Fatal(err)
	}
	f.m = m
	if status(t, f, "one").State != Active {
		t.Fatal("launcher exit released lease")
	}
	f.observe("one", func(e *Evidence) {
		e.Running = false
		e.Terminated = true
		e.OwnershipReconciled = true
		e.DescendantsTerminated = false
	})
	if err = m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Unknown {
		t.Fatal("live descendants released")
	}
	f.a.mu.Lock()
	delete(f.a.sessions, b.SessionRef)
	f.a.mu.Unlock()
	ancient := time.Now().Add(-30 * 24 * time.Hour)
	if err = os.Chtimes(statePath(f.home), ancient, ancient); err != nil {
		t.Fatal(err)
	}
	if err = m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if !status(t, f, "one").Busy {
		t.Fatal("TTL or missing liveness released lease")
	}
	f.a.mu.Lock()
	f.a.sessions[b.SessionRef] = Evidence{SessionRef: b.SessionRef, Generation: 1, OwnerRef: b.OwnerRef, ProcessGeneration: "reused-pid-new-process", Terminated: true, OwnershipReconciled: true, DescendantsTerminated: true}
	f.a.mu.Unlock()
	if err = m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Unknown {
		t.Fatal("process incarnation mismatch released")
	}
	f.observe("one", func(e *Evidence) { e.ProcessGeneration = "boot-session-runner-1" })
	if err = m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Released {
		t.Fatal("authoritative termination not released")
	}
	after, _ := os.ReadFile(bindingPath(f.home, "one"))
	if !reflect.DeepEqual(immutable, after) {
		t.Fatal("binding mutated with state")
	}
}

// Recovery is conservative; resume binds the original composition and current
// authority. Revocation refuses new admissions without stopping an active lease.
func TestRecoveryResumeAndRevocation(t *testing.T) {
	f := setup(t)
	b := activate(t, f, "one")
	if err := f.m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Unknown {
		t.Fatal("recovery inferred liveness")
	}
	_, err := f.m.Resume(context.Background(), "one")
	mustRefuse(t, err, "reconciled")
	if err = f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	latest := f.c
	latest.Ref = "composition:latest"
	latest.SHA256 = strings.Repeat("f", 64)
	f.a.comps[latest.Ref] = latest
	got, err := f.m.Resume(context.Background(), "one")
	if err != nil || got != b {
		t.Fatalf("resume=%+v %v", got, err)
	}
	f.a.revoked = true
	_, err = f.m.Acquire(context.Background(), f.reserve("two"))
	if !errors.Is(err, ErrRevoked) {
		t.Fatalf("revoked acquire=%v", err)
	}
	_, err = f.m.Resume(context.Background(), "one")
	if !errors.Is(err, ErrRevoked) {
		t.Fatalf("revoked resume=%v", err)
	}
	row := status(t, f, "one")
	if !row.Revoked || row.State != Active || !row.Busy {
		t.Fatalf("revoked running status=%+v", row)
	}
	f.a.revoked = false
	f.a.uncertain = true
	row = status(t, f, "one")
	if !row.AuthorityUnknown || row.Revoked {
		t.Fatalf("unreadable authority=%+v", row)
	}
}

// Aborted pre-start work releases only with reserved-session no-child and
// ownership proof. A transfer without published bytes cannot start a provider.
func TestPreparedReservationAndTransferGates(t *testing.T) {
	f := setup(t)
	acquire(t, f, "one")
	mustRefuse(t, f.m.Transfer(context.Background(), "one"), "materialize")
	if err := f.m.PublishPrepared(context.Background(), "one", f.files, false); err != nil {
		t.Fatal(err)
	}
	mustRefuse(t, f.m.Transfer(context.Background(), "one"), "runner identity")
	f.observe("one", func(e *Evidence) { e.NoChildStarted = false; e.OwnershipReconciled = true })
	if err := f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Unknown {
		t.Fatal("missing no-child proof released")
	}
	f.observe("one", func(e *Evidence) { e.NoChildStarted = true })
	if err := f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Released {
		t.Fatal("pre-start abort not released")
	}
}

// Publication is receipt-bound, inventories generated outputs separately,
// rejects unmanaged/login collisions, and rolls back exact bytes on a fault.
func TestPublicationReceiptAndRollback(t *testing.T) {
	f := setup(t)
	acquire(t, f, "one")
	mustRefuse(t, f.m.PublishPrepared(context.Background(), "one", map[string][]byte{"AGENTS.md": []byte("forged")}, false), "receipt")
	if err := f.m.PublishPrepared(context.Background(), "one", f.files, false); err != nil {
		t.Fatal(err)
	}
	if err := f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	f.observe("one", func(e *Evidence) { e.OwnershipReconciled = true })
	if err := f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	next := map[string][]byte{"AGENTS.md": []byte("second")}
	c := f.c
	c.Ref = "composition:next"
	c.SHA256 = strings.Repeat("d", 64)
	c.Generated = inventory(next)
	f.a.comps[c.Ref] = c
	f.m.hooks = transaction.Hooks{Fault: func(e transaction.Event) error {
		if e.Point == transaction.PointAfterInstall {
			return errors.New("injected publication failure")
		}
		return nil
	}}
	mustRefuse(t, f.m.Update(context.Background(), c.Ref, c.SHA256, "principal:operator", next, false), "injected publication failure")
	f.m.hooks = transaction.Hooks{}
	if err := f.m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(f.home, homeRelative(f.c), "AGENTS.md"))
	if string(raw) != "approved\n" {
		t.Fatal("rollback lost old generated bytes")
	}
	if err := f.m.Update(context.Background(), c.Ref, c.SHA256, "principal:operator", next, false); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(f.home, homeRelative(c))
	if err := os.WriteFile(filepath.Join(home, "auth.json"), []byte("native"), 0600); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"auth.json": []byte("replace")}
	d := c
	d.Ref = "composition:collision"
	d.SHA256 = strings.Repeat("f", 64)
	d.Generated = inventory(files)
	f.a.comps[d.Ref] = d
	mustRefuse(t, f.m.Update(context.Background(), d.Ref, d.SHA256, "principal:operator", files, false), "unmanaged native")
	raw, _ = os.ReadFile(filepath.Join(home, "auth.json"))
	if string(raw) != "native" {
		t.Fatal("native state replaced")
	}
}

// Invalid identity/authority inputs never allocate a lease, and private record
// corruption is reported as unknown evidence rather than an empty GC root set.
func TestIdentityAndUnreadableEvidence(t *testing.T) {
	f := setup(t)
	for _, edit := range []func(*Request){func(r *Request) { r.Generation = 0 }, func(r *Request) { r.OwnerRef = "principal:../x" }, func(r *Request) { r.ContextSHA256 = strings.Repeat("0", 64) }, func(r *Request) { r.LeaseID = "../escape" }, func(r *Request) { r.Generation = 2 }} {
		r := f.reserve("one")
		edit(&r)
		if _, err := f.m.Acquire(context.Background(), r); err == nil {
			t.Fatal("invalid request accepted")
		}
	}
	if len(mustLoad(t, f).Leases) != 0 {
		t.Fatal("invalid input allocated")
	}
	acquire(t, f, "one")
	lockManager, _ := managerlock.New(f.home)
	lock, err := lockManager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	roots, err := ReferencesUnderLock(f.home, lock)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots.Homes) != 1 || !reflect.DeepEqual(roots.BuildKeys, f.c.BuildKeys) || !reflect.DeepEqual(roots.RuntimeEntries, f.c.RuntimeEntries) {
		t.Fatalf("roots=%+v", roots)
	}
	if err = os.WriteFile(bindingPath(f.home, "one"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ReferencesUnderLock(f.home, lock); err == nil {
		t.Fatal("corrupt binding became absent")
	}
	if err = lock.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = f.m.Status(context.Background()); err == nil {
		t.Fatal("corruption hidden")
	}
}
func mustLoad(t *testing.T, f fixture) *ledger {
	t.Helper()
	l, err := load(f.home)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// Status of an absent manager performs no provisioning. Existing status reads
// neither recover a pending journal nor rewrite protected state.
func TestStatusIsReadOnly(t *testing.T) {
	f := setup(t)
	missing := filepath.Join(f.home, "missing")
	m, _ := New(missing, f.a, f.a)
	if rows, err := m.Status(context.Background()); err != nil || len(rows) != 0 {
		t.Fatalf("absent status=%v %v", rows, err)
	}
	if _, err := os.Lstat(missing); !os.IsNotExist(err) {
		t.Fatal("status provisioned manager")
	}
	acquire(t, f, "one")
	before, _ := os.ReadFile(statePath(f.home))
	if _, err := f.m.Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(statePath(f.home))
	if string(before) != string(after) {
		t.Fatal("status wrote state")
	}
}
