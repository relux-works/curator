package homelease_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/homelease"
	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
)

type projectAuthority struct {
	home        string
	composition projectlease.Composition
}

func (a projectAuthority) Resolve(_ context.Context, held managerlock.Held, _, _ string) (projectlease.Composition, error) {
	return a.composition, held.Validate(a.home)
}
func (a projectAuthority) Authorize(_ context.Context, held managerlock.Held, _ projectlease.Composition, _ string) error {
	return held.Validate(a.home)
}
func (a projectAuthority) Observe(_ context.Context, session string, generation int64) (projectlease.Evidence, error) {
	return projectlease.Evidence{SessionRef: session, Generation: generation, OwnerRef: "principal:operator", NoChildStarted: true}, nil
}

// HL10/12/W16: actual project allocations remain project-lease/1 and busy
// publication remains project_view_busy. Profile-home lifecycle calls cannot
// bind, reconcile or operator-break those IDs and preserve both ledgers.
func TestProjectKindBoundary(t *testing.T) {
	home := t.TempDir()
	a := projectAuthority{home: home, composition: projectlease.Composition{Ref: "composition:fixture", SHA256: strings.Repeat("a", 64), CheckoutID: "checkout", Profile: "sample", Environment: "codex_cli", HomeRef: "home:fixture", ConcurrentState: true, BuildKeys: []string{}, RuntimeEntries: []string{}, Generated: map[string]projectlease.Surface{}}}
	p, err := projectlease.New(home, a, a)
	if err != nil {
		t.Fatal(err)
	}
	b, err := p.Acquire(context.Background(), projectlease.Request{LeaseID: "project-one", ContextRef: a.composition.Ref, ContextSHA256: a.composition.SHA256, OwnerRef: "principal:operator", SessionRef: "session:fixture", Generation: 1})
	if err != nil || b.Kind != "project-lease/1" {
		t.Fatalf("real project allocation: %+v %v", b, err)
	}
	m, err := homelease.New(home, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	before := boundarySnapshot(t, home)
	session, generation := "fixture-session", int64(1)
	h := homelease.Holder{Kind: "bound", Launcher: "curator-run", ManagerID: "manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "request", Session: &session, Generation: &generation, Process: &state.Process{PID: 123, StartedAt: "2026-10-27T00:00:01Z"}}
	for _, action := range []func() error{
		func() error { _, _, err := m.Bind(context.Background(), b.LeaseID, 1, h); return err },
		func() error { _, _, err := m.Release(context.Background(), b.LeaseID); return err },
		func() error {
			_, _, err := m.Break(context.Background(), b.LeaseID, 1, true, "protection deliberately abandoned")
			return err
		},
	} {
		var refusal *state.Refusal
		if err := action(); !errors.As(err, &refusal) || refusal.Code != "environment_lease_authority_refused" || refusal.Reason != "unsupported_lease_kind" {
			t.Fatalf("cross-kind lifecycle accepted: %v", err)
		}
		if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
			t.Fatal("cross-kind refusal changed state")
		}
	}
	err = p.Update(context.Background(), a.composition.Ref, a.composition.SHA256, "principal:operator", map[string][]byte{}, false)
	if err == nil || !strings.Contains(err.Error(), projectlease.Busy) {
		t.Fatalf("project busy code changed: %v", err)
	}
	if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
		t.Fatal("busy refusal changed state")
	}
	out, reason, err := m.Release(context.Background(), "absent")
	if err != nil || reason != "unknown_lease" || out.Changed {
		t.Fatalf("absent profile control: %+v %s %v", out, reason, err)
	}
	if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
		t.Fatal("absent control changed state")
	}
}

func boundarySnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		out[rel] = i.Mode().String()
		if d.Type()&os.ModeSymlink != 0 {
			s, err := os.Readlink(path)
			out[rel] += s
			return err
		}
		if !d.IsDir() {
			b, err := os.ReadFile(path)
			out[rel] += string(b)
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return out
}
