package scopes

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
)

// managerHome is carried separately from the opaque composition projection.
type raceLeaseAuthority struct {
	c                projectlease.Composition
	managerHome      string
	entered, release chan struct{}
}

func (a *raceLeaseAuthority) Resolve(_ context.Context, held managerlock.Held, _, _ string) (projectlease.Composition, error) {
	if err := held.Validate(a.managerHome); err != nil {
		return projectlease.Composition{}, err
	}
	if a.entered != nil {
		close(a.entered)
		<-a.release
	}
	return a.c, nil
}

func (a *raceLeaseAuthority) Authorize(_ context.Context, held managerlock.Held, _ projectlease.Composition, _ string) error {
	return held.Validate(a.managerHome)
}

func (a *raceLeaseAuthority) Observe(_ context.Context, ref string, generation int64) (projectlease.Evidence, error) {
	return projectlease.Evidence{SessionRef: ref, Generation: generation, OwnerRef: "principal:operator", NoChildStarted: true}, nil
}

// Regression from the independent review: a real foreign-home lock cannot
// sweep runtime while Acquire pauses after same-home witness validation.
// Resuming Acquire succeeds with its artifact intact; correct-home GC then
// retains both leased roots and removes the unreferenced positive control.
func TestCollectProjectLeaseForeignLockInterleaving(t *testing.T) {
	ctx := context.Background()
	home := protectedTestHome(t)
	for _, entry := range []string{"first", "next", "orphan"} {
		if err := os.MkdirAll(filepath.Join(home, "runtime", "tool", entry), 0700); err != nil {
			t.Fatal(err)
		}
	}
	a := &raceLeaseAuthority{managerHome: home}
	a.c = projectlease.Composition{Ref: "composition:first", SHA256: strings.Repeat("a", 64), CheckoutID: "first", Profile: "default", Environment: "codex_cli", HomeRef: "home:first", ConcurrentState: true, BuildKeys: []string{}, RuntimeEntries: []string{"tool/first"}, Generated: map[string]projectlease.Surface{}}
	m, err := projectlease.New(home, a, a)
	if err != nil {
		t.Fatal(err)
	}
	request := func(id string) projectlease.Request {
		return projectlease.Request{LeaseID: id, ContextRef: a.c.Ref, ContextSHA256: a.c.SHA256, OwnerRef: "principal:operator", SessionRef: "session:" + id, Generation: 1}
	}
	if _, err = m.Acquire(ctx, request("first")); err != nil {
		t.Fatal(err)
	}
	a.c.Ref, a.c.SHA256, a.c.CheckoutID, a.c.HomeRef = "composition:next", strings.Repeat("c", 64), "next", "home:next"
	a.c.RuntimeEntries = []string{"tool/next"}
	r := request("next")
	a.entered, a.release = make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() { _, err := m.Acquire(ctx, r); done <- err }()
	<-a.entered
	foreign, err := managerlock.New(t.TempDir())
	if err != nil {
		close(a.release)
		<-done
		t.Fatal(err)
	}
	lock, err := foreign.AcquireHomeOnly(ctx, false)
	if err != nil {
		close(a.release)
		<-done
		t.Fatal(err)
	}
	result, collectErr := Collect(MaintenanceRequest{Home: home, Lock: lock})
	_ = lock.Close()
	_, artifactErr := os.Stat(filepath.Join(home, "runtime", "tool", "next"))
	close(a.release)
	acquireErr := <-done
	if !errors.Is(collectErr, managerlock.ErrNotHeld) || len(result.RemovedRuntime) != 0 || artifactErr != nil || acquireErr != nil {
		t.Fatalf("foreign GC interleaving: result=%+v gc=%v artifact=%v acquire=%v", result, collectErr, artifactErr, acquireErr)
	}
	lm, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err = lm.AcquireHomeOnly(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	result, err = Collect(MaintenanceRequest{Home: home, Lock: lock, Cache: &recordingCache{}})
	if err != nil || len(result.RemovedRuntime) != 1 || result.RemovedRuntime[0] != "tool/orphan" {
		t.Fatalf("same-home GC control: %+v %v", result, err)
	}
	for _, entry := range []string{"first", "next"} {
		if _, err = os.Stat(filepath.Join(home, "runtime", "tool", entry)); err != nil {
			t.Fatal("same-home GC removed leased runtime", err)
		}
	}
}
