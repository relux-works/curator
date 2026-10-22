package scopes

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
)

type leaseAuthority struct{ c projectlease.Composition }

func (a leaseAuthority) Resolve(context.Context, managerlock.Held, string, string) (projectlease.Composition, error) {
	return a.c, nil
}
func (a leaseAuthority) Authorize(context.Context, managerlock.Held, projectlease.Composition, string) error {
	return nil
}
func (a leaseAuthority) Observe(context.Context, string, int64) (projectlease.Evidence, error) {
	return projectlease.Evidence{SessionRef: "session:first", Generation: 1, OwnerRef: "principal:operator", NoChildStarted: true}, nil
}

// Real leased runtime and build-cache artifacts survive GC with no ordinary
// marker. An unreferenced neighbor is swept. Unknown/corrupt lease evidence
// stops both sweeps before touching either cache or pruning consumers.
func TestCollectProjectLeaseReferencesAndUnknown(t *testing.T) {
	home := protectedTestHome(t)
	store := openProtectedStore(t, home)
	kept := publishRealEntry(t, store, "leased", "leased artifact")
	orphan := publishRealEntry(t, store, "orphan", "orphan artifact")
	backdateEntry(t, home, kept, 30*24*time.Hour)
	backdateEntry(t, home, orphan, 30*24*time.Hour)
	for _, entry := range []string{"tool/kept", "tool/orphan"} {
		if err := os.MkdirAll(filepath.Join(home, "runtime", entry), 0700); err != nil {
			t.Fatal(err)
		}
	}
	a := leaseAuthority{projectlease.Composition{Ref: "composition:first", SHA256: strings.Repeat("a", 64), CheckoutID: "checkout", Profile: "default", Environment: "codex_cli", HomeRef: "home:first", ConcurrentState: true, BuildKeys: []string{string(kept)}, RuntimeEntries: []string{"tool/kept"}, Generated: map[string]projectlease.Surface{}}}
	m, err := projectlease.New(home, a, a)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Acquire(context.Background(), projectlease.Request{LeaseID: "first", ContextRef: a.c.Ref, ContextSHA256: a.c.SHA256, OwnerRef: "principal:operator", SessionRef: "session:first", Generation: 1}); err != nil {
		t.Fatal(err)
	}
	if err = m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	lm, _ := managerlock.New(home)
	lock, err := lm.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	result, err := Collect(MaintenanceRequest{Home: home, Lock: lock, Cache: store})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedBuilds) != 1 || result.RemovedBuilds[0] != string(orphan) || len(result.RemovedRuntime) != 1 || result.RemovedRuntime[0] != "tool/orphan" {
		t.Fatalf("GC=%+v", result)
	}
	if _, err = os.Stat(realEntryPath(home, kept)); err != nil {
		t.Fatal("leased build removed", err)
	}
	if _, err = os.Stat(filepath.Join(home, "runtime/tool/kept")); err != nil {
		t.Fatal("leased runtime removed", err)
	}
	if _, err = CollectRuntime(home); err == nil || !strings.Contains(err.Error(), "serialized Collect") {
		t.Fatalf("unlocked collector=%v", err)
	}
	if err = os.Remove(filepath.Join(home, "state/project-leases/v1/ledger.json")); err != nil {
		t.Fatal(err)
	}
	if _, err = Collect(MaintenanceRequest{Home: home, Lock: lock, Cache: store}); err == nil || !strings.Contains(err.Error(), "sweeps skipped") {
		t.Fatalf("missing ledger=%v", err)
	}
	if _, err = os.Stat(realEntryPath(home, kept)); errors.Is(err, os.ErrNotExist) {
		t.Fatal("unknown build removed")
	}
	if _, err = os.Stat(filepath.Join(home, "runtime/tool/kept")); err != nil {
		t.Fatal("unknown runtime removed", err)
	}
}
