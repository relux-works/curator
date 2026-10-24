package projectlease

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// The binding is the normative closed CCJ-1 carrier, excluding lifecycle state.
// Duplicate/extra/null/out-of-bound members and noncanonical bytes are refused.
func TestBindingGrammar(t *testing.T) {
	f := setup(t)
	b := acquire(t, f, "one")
	raw, _ := canonical(b)
	parsed, err := ParseBinding(raw)
	if err != nil || parsed != b {
		t.Fatalf("valid binding=%+v %v", parsed, err)
	}
	var v map[string]any
	if err = json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(map[string]any){func(v map[string]any) { v["extra"] = true }, func(v map[string]any) { v["generation"] = float64(0) }, func(v map[string]any) { v["session_ref"] = nil }, func(v map[string]any) { v["environment"] = "other" }, func(v map[string]any) { v["home_ref"] = "home:../escape" }, func(v map[string]any) { v["kind"] = "project-lease/2" }} {
		var x map[string]any
		_ = json.Unmarshal(raw, &x)
		edit(x)
		bad, _ := canonical(x)
		if _, err = ParseBinding(bad); err == nil {
			t.Fatalf("invalid binding accepted %s", bad)
		}
	}
	for _, bad := range [][]byte{append(raw, '\n'), []byte(strings.Replace(string(raw), `"generation":1`, `"generation":1,"generation":1`, 1)), []byte(strings.Replace(string(raw), `"generation":1`, `"generation":9007199254740992`, 1)), make([]byte, maxRecord+1)} {
		if _, err = ParseBinding(bad); err == nil {
			t.Fatal("bad transport accepted")
		}
	}
}

// Two independent managers contend on the existing lock. They cannot both
// reserve different compositions for one home.
func TestAcquisitionTOCTOU(t *testing.T) {
	f := setup(t)
	r1 := f.reserve("one")
	r2 := f.reserve("two")
	c := f.c
	c.Ref = "composition:next"
	c.SHA256 = strings.Repeat("c", 64)
	f.a.comps[c.Ref] = c
	r2.ContextRef = c.Ref
	r2.ContextSHA256 = c.SHA256
	m2, _ := New(f.home, f.a, f.a)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i, r := range []Request{r1, r2} {
		wg.Add(1)
		go func(i int, r Request) {
			defer wg.Done()
			<-start
			m := f.m
			if i == 1 {
				m = m2
			}
			_, err := m.Acquire(context.Background(), r)
			results <- err
		}(i, r)
	}
	close(start)
	wg.Wait()
	close(results)
	accepted, refused := 0, 0
	for err := range results {
		if err == nil {
			accepted++
		} else if strings.Contains(err.Error(), Busy) {
			refused++
		} else {
			t.Fatal(err)
		}
	}
	if accepted != 1 || refused != 1 {
		t.Fatalf("race accepted=%d refused=%d", accepted, refused)
	}
	if len(mustLoad(t, f).Leases) != 1 {
		t.Fatal("race published both leases")
	}
}

// Link/unreadable authority state cannot escape into another home or erase
// liveness evidence. Generated drift refuses transfer rather than rebaselining.
func TestLinksDriftAndMissingState(t *testing.T) {
	f := setup(t)
	acquire(t, f, "one")
	if err := f.m.PublishPrepared(context.Background(), "one", f.files, false); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(f.home, homeRelative(f.c), "AGENTS.md")
	if err := os.WriteFile(target, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	f.observe("one", func(e *Evidence) { e.Running = true; e.ProcessGeneration = "runner" })
	mustRefuse(t, f.m.Transfer(context.Background(), "one"), "drift")
	if err := os.WriteFile(target, f.files["AGENTS.md"], 0600); err != nil {
		t.Fatal(err)
	}
	if err := f.m.Transfer(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(statePath(f.home)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.m.Status(context.Background()); err == nil {
		t.Fatal("missing ledger became no leases")
	}
	lm, _ := managerlock.New(f.home)
	lock, err := lm.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ReferencesUnderLock(f.home, lock); err == nil {
		t.Fatal("missing state erased GC roots")
	}
	_ = lock.Close()
	g := setup(t)
	outside := t.TempDir()
	if err = os.MkdirAll(filepath.Join(g.home, "state"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(outside, filepath.Join(g.home, "state/project-leases")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err = g.m.Acquire(context.Background(), g.reserve("one")); err == nil {
		t.Fatal("linked authority accepted")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal("wrote through linked authority", err)
	}
}

// Interrupted transactions retain a journal, status does not recover it, and
// explicit recovery reconciles durable intent while keeping liveness unknown.
func TestPendingRecoveryStatus(t *testing.T) {
	f := setup(t)
	acquire(t, f, "one")
	f.m.hooks = transaction.Hooks{Fault: func(e transaction.Event) error {
		if e.Point == transaction.PointBeforeCleanup {
			return errors.New("cleanup interrupted")
		}
		return nil
	}}
	mustRefuse(t, f.m.PublishPrepared(context.Background(), "one", f.files, false), "cleanup interrupted")
	before, _ := os.ReadFile(statePath(f.home))
	row := status(t, f, "one")
	if row.State != Unknown || !row.Busy {
		t.Fatalf("pending status=%+v", row)
	}
	after, _ := os.ReadFile(statePath(f.home))
	if string(before) != string(after) {
		t.Fatal("status recovered or rewrote")
	}
	f.m.hooks = transaction.Hooks{}
	if err := f.m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "one").State != Unknown {
		t.Fatal("recovery released lease")
	}
}

// Portable output inventories reject path traversal, case/normalization aliases
// and file/ancestor collisions before publishing; a neighboring nested file is valid.
func TestReceiptPortablePaths(t *testing.T) {
	for _, paths := range [][]string{{"../outside"}, {"a", "a-b", "a/child"}, {"A", "a"}, {"caf\u00e9", "cafe\u0301"}} {
		f := setup(t)
		files := map[string][]byte{}
		for _, p := range paths {
			files[p] = []byte("fixture")
		}
		c := f.c
		c.Generated = inventory(files)
		f.a.comps[c.Ref] = c
		if _, err := f.m.Acquire(context.Background(), f.reserve("one")); err == nil {
			t.Fatal("invalid receipt paths accepted", paths)
		}
	}
	f := setup(t)
	f.files = map[string][]byte{"nested/file": []byte("fixture")}
	f.c.Generated = inventory(f.files)
	f.a.comps[f.c.Ref] = f.c
	acquire(t, f, "one")
	if err := f.m.PublishPrepared(context.Background(), "one", f.files, false); err != nil {
		t.Fatal(err)
	}
}

// A profile spelling alias cannot allocate another home over an already leased
// physical view, even on filesystems that could keep both case spellings.
func TestHomeAllocationAlias(t *testing.T) {
	f := setup(t)
	activate(t, f, "one")
	c := f.c
	c.Ref = "composition:alias"
	c.SHA256 = strings.Repeat("d", 64)
	c.Profile = "Default"
	c.HomeRef = "home:alias"
	f.a.comps[c.Ref] = c
	r := f.reserve("two")
	r.ContextRef = c.Ref
	r.ContextSHA256 = c.SHA256
	_, err := f.m.Acquire(context.Background(), r)
	mustRefuse(t, err, "allocation identity mismatch")
}

// A lease acquisition cannot pass an idle publication's held manager lock.
// After publication it reserves the original composition without losing either
// durable record; transfer still refuses until that composition is republished.
func TestPublicationSerializesLeaseAcquisition(t *testing.T) {
	f := setup(t)
	acquire(t, f, "one")
	if err := f.m.PublishPrepared(context.Background(), "one", f.files, false); err != nil {
		t.Fatal(err)
	}
	f.observe("one", func(e *Evidence) { e.OwnershipReconciled = true })
	if err := f.m.Reconcile(context.Background(), "one"); err != nil {
		t.Fatal(err)
	}
	c := f.c
	c.Ref = "composition:next"
	c.SHA256 = strings.Repeat("c", 64)
	files := map[string][]byte{"AGENTS.md": []byte("next")}
	c.Generated = inventory(files)
	f.a.comps[c.Ref] = c
	staged := make(chan struct{})
	release := make(chan struct{})
	var unblock sync.Once
	var firstPublication sync.Once
	defer func() { unblock.Do(func() { close(release) }) }()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	f.m.hooks = transaction.Hooks{Observe: func(e transaction.Event) {
		if e.Point == transaction.PointPrepared {
			firstPublication.Do(func() {
				close(staged)
				select {
				case <-release:
				case <-ctx.Done():
				}
			})
		}
	}}
	update := make(chan error, 1)
	go func() { update <- f.m.Update(ctx, c.Ref, c.SHA256, "principal:operator", files, false) }()
	select {
	case <-staged:
	case err := <-update:
		t.Fatalf("update missed staging point: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	acquired := make(chan error, 1)
	r := f.reserve("two")
	waiting := &waitObservedContext{Context: ctx, waiting: make(chan struct{})}
	go func() { _, err := f.m.Acquire(waiting, r); acquired <- err }()
	select {
	case err := <-acquired:
		t.Fatalf("acquisition passed an in-flight update: %v", err)
	case <-waiting.waiting:
	}
	// Done is requested only after the real OS lock refuses acquisition.
	// Publication is still held at PointPrepared, so no scheduling delay is
	// needed to establish contention before the test releases the barrier.
	select {
	case err := <-acquired:
		t.Fatalf("acquisition passed an in-flight update: %v", err)
	default:
	}
	unblock.Do(func() { close(release) })
	if err := <-update; err != nil {
		t.Fatal(err)
	}
	if err := <-acquired; err != nil {
		t.Fatal(err)
	}
	f.m.hooks = transaction.Hooks{}
	l := mustLoad(t, f)
	if _, ok := l.Leases["two"]; !ok || l.Views[homeRelative(c)].Composition.Ref != c.Ref {
		t.Fatal("serialized operation lost durable state")
	}
	f.observe("two", func(e *Evidence) { e.Running = true; e.ProcessGeneration = "runner-two" })
	mustRefuse(t, f.m.Transfer(context.Background(), "two"), "materialize")
	if err := f.m.PublishPrepared(context.Background(), "two", f.files, false); err != nil {
		t.Fatal(err)
	}
	if err := f.m.Transfer(context.Background(), "two"); err != nil {
		t.Fatal(err)
	}
}

// Invalid witness authority is refused before callbacks or allocations. A real
// same-home acquisition reaches both protected callbacks and becomes stale on
// owner release; a lock on another home cannot authorize this manager's reads.
func TestAuthorityRequiresHeldWitness(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	resolve := func(held managerlock.Held) error {
		_, err := f.m.resolve(ctx, held, f.c.Ref, f.c.SHA256, "principal:operator")
		return err
	}
	if !errors.Is(resolve(managerlock.Held{}), managerlock.ErrNotHeld) {
		t.Fatal("zero witness reached authority")
	}
	if len(f.a.witnesses) != 0 {
		t.Fatal("zero witness invoked callbacks")
	}
	lm, err := managerlock.New(f.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := lm.AcquireHomeOnly(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	held := lock.Held()
	defer func() { _ = lock.Close() }()
	if err := resolve(held); err != nil {
		t.Fatal(err)
	}
	if len(f.a.witnesses) != 2 {
		t.Fatalf("callbacks=%d", len(f.a.witnesses))
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(resolve(held), managerlock.ErrNotHeld) {
		t.Fatal("stale witness reached authority")
	}
	other, err := managerlock.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := other.AcquireHomeOnly(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = foreign.Close() }()
	if !errors.Is(resolve(foreign.Held()), managerlock.ErrNotHeld) {
		t.Fatal("wrong-home witness reached authority")
	}
	if len(f.a.witnesses) != 2 {
		t.Fatal("invalid witness invoked callbacks")
	}
	if _, err := os.Stat(statePath(f.home)); !os.IsNotExist(err) {
		t.Fatalf("invalid witness allocated state: %v", err)
	}
	for _, observed := range f.a.witnesses {
		if !errors.Is(observed.Validate(f.home), managerlock.ErrNotHeld) {
			t.Fatal("callback witness extended owner lifetime")
		}
	}
	// Existing fixture callbacks validate every witness across the full lifecycle.
	activate(t, f, "witness")
	if _, err := f.m.Resume(ctx, "witness"); err != nil {
		t.Fatal(err)
	}
	if status(t, f, "witness").AuthorityUnknown {
		t.Fatal("status omitted valid witness")
	}
}
