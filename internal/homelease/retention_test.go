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
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/scopes"
)

type retentionAuthority struct{ home, runtime, key string }

func (a retentionAuthority) Admit(_ context.Context, held managerlock.Held, environment, profile, digest string) (homelease.Admission, error) {
	if err := held.Validate(a.home); err != nil {
		return homelease.Admission{}, err
	}
	home := filepath.Join(a.home, "environments", profile, environment)
	return homelease.Admission{Home: home, Protected: []string{home}, RuntimeEntries: []string{a.runtime}, BuildKeys: []string{a.key}}, nil
}

// HL08/16/W18: prepared and persisted unknown holders retain real runtime
// entries absent from every ordinary scope, while the real collector removes
// an unreferenced neighbor. Build-key union is asserted at the real mark
// boundary; this test does not publish or execute a compiler artifact.
func TestHomeLeaseCollectorRetention(t *testing.T) {
	for _, phase := range []string{"prepared", "unknown"} {
		t.Run(phase, func(t *testing.T) {
			identity, err := managerlock.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			home := identity.Home()
			if err := privatedir.Protect(home); err != nil {
				t.Fatal(err)
			}
			retained, neighbor := "leased/"+strings.Repeat("a", 40), "neighbor/"+strings.Repeat("b", 40)
			for _, entry := range []string{retained, neighbor} {
				if err := os.MkdirAll(filepath.Join(home, "runtime", filepath.FromSlash(entry)), 0700); err != nil {
					t.Fatal(err)
				}
			}
			key := "sha256:" + strings.Repeat("c", 64)
			m, err := homelease.New(home, retentionAuthority{home, retained, key}, nil)
			if err != nil {
				t.Fatal(err)
			}
			r := homelease.Request{Environment: "claude", Profile: "sample", LockSHA256: strings.Repeat("d", 64), RequestID: "retention", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "dead-launcher-manager", ManagerRegisteredAt: "2020-01-01T00:00:00Z", RequestID: "retention"}}
			out, _, err := m.Acquire(context.Background(), r)
			if err != nil {
				t.Fatal(err)
			}
			if phase == "unknown" {
				// Persisted uncertainty is a valid cold-start projection. Its
				// initial immutable binding and protected references stay intact.
				l, err := state.Load(home)
				if err != nil {
					t.Fatal(err)
				}
				e := l.Leases[out.Lease.LeaseID]
				e.Record.Phase = "unknown"
				at := "2020-01-01T00:00:01Z"
				e.Record.Observation = state.Observation{State: "unknown", ObservedAt: &at}
				l.Leases[out.Lease.LeaseID] = e
				raw, err := state.Canonical(l)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(state.LedgerPath(home), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := scopes.CollectRuntime(home); err == nil {
				t.Fatal("historical unlocked collector accepted lease state")
			}
			if _, err := os.Stat(filepath.Join(home, "runtime", filepath.FromSlash(neighbor))); err != nil {
				t.Fatalf("refusal swept neighbor: %v", err)
			}
			locks, _ := managerlock.New(home)
			lock, err := locks.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			roots, err := homelease.ReferencesUnderLock(home, lock)
			if err != nil || !reflect.DeepEqual(roots.RuntimeEntries, []string{retained}) || !reflect.DeepEqual(roots.BuildKeys, []string{key}) {
				t.Fatalf("retained roots: %+v %v", roots, err)
			}
			result, err := scopes.Collect(scopes.MaintenanceRequest{Home: home, Lock: lock})
			_ = lock.Close()
			if err != nil || !reflect.DeepEqual(result.RemovedRuntime, []string{neighbor}) {
				t.Fatalf("collector: %+v %v", result, err)
			}
			if _, err := os.Stat(filepath.Join(home, "runtime", filepath.FromSlash(retained))); err != nil {
				t.Fatalf("live root lost: %v", err)
			}
			if _, err := os.Stat(filepath.Join(home, "runtime", filepath.FromSlash(neighbor))); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("unreferenced control retained: %v", err)
			}
			rows, err := m.List(context.Background(), "", "")
			if err != nil || len(rows) != 1 || rows[0].Phase != phase {
				t.Fatalf("collector expired lease: %+v %v", rows, err)
			}
		})
	}
}

// HL03/16: the same physical manager has the same retained roots through a
// directory alias. A foreign or closed acquisition cannot authorize that read;
// both refusals and successful reference reads preserve the complete preimage.
func TestHomeLeaseCollectorManagerAlias(t *testing.T) {
	for _, spelling := range []string{"canonical", "alias"} {
		t.Run(spelling, func(t *testing.T) {
			base := t.TempDir()
			configured := filepath.Join(base, "manager")
			if err := privatedir.MakeAll(configured); err != nil {
				t.Fatal(err)
			}
			locks, err := managerlock.New(configured)
			if err != nil {
				t.Fatal(err)
			}
			home := locks.Home()
			query := home
			if spelling == "alias" {
				query = filepath.Join(base, "manager-alias")
				if err := os.Symlink(home, query); err != nil {
					t.Fatal(err)
				}
			}
			retained := "leased/" + strings.Repeat("a", 40)
			key := "sha256:" + strings.Repeat("c", 64)
			m, err := homelease.New(query, retentionAuthority{home, retained, key}, nil)
			if err != nil {
				t.Fatal(err)
			}
			req := homelease.Request{Environment: "claude", Profile: "sample", LockSHA256: strings.Repeat("d", 64), RequestID: "alias-retention", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "dead-launcher-manager", ManagerRegisteredAt: "2020-01-01T00:00:00Z", RequestID: "alias-retention"}}
			if _, _, err := m.Acquire(context.Background(), req); err != nil {
				t.Fatal(err)
			}
			if _, err := state.Load(home); err != nil {
				t.Fatalf("canonical ledger control: %v", err)
			}
			lock, err := locks.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			before := boundarySnapshot(t, home)
			roots, err := homelease.ReferencesUnderLock(query, lock)
			if err != nil || !reflect.DeepEqual(roots.RuntimeEntries, []string{retained}) || !reflect.DeepEqual(roots.BuildKeys, []string{key}) {
				t.Errorf("same-root retained references: %+v %v", roots, err)
			}
			if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
				t.Fatal("reference read changed preimage")
			}
			foreignManager, err := managerlock.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			foreign, err := foreignManager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			roots, err = homelease.ReferencesUnderLock(query, foreign)
			if closeErr := foreign.Close(); closeErr != nil {
				t.Fatal(closeErr)
			}
			if !errors.Is(err, managerlock.ErrNotHeld) || !reflect.DeepEqual(roots, homelease.Roots{}) {
				t.Errorf("foreign lock: %+v %v", roots, err)
			}
			if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
				t.Fatal("foreign-lock refusal changed preimage")
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			roots, err = homelease.ReferencesUnderLock(query, lock)
			if !errors.Is(err, managerlock.ErrNotHeld) || !reflect.DeepEqual(roots, homelease.Roots{}) {
				t.Errorf("closed lock: %+v %v", roots, err)
			}
			if !reflect.DeepEqual(before, boundarySnapshot(t, home)) {
				t.Fatal("closed-lock refusal changed preimage")
			}
		})
	}
}
