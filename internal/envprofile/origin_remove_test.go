package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
)

// Reaches actual Remove, complete journal cleanup, then a new Resolve request
// for disjoint Q. P remains byte-identical and nonlaunchable without native
// account/backend lookup. This is a library/process-state witness, not an OS
// power-loss, CLI/GC/lease owner or explicit later-purge qualification.
func prepareRemovalOriginFixture(t *testing.T) (*managedFixture, ResolveRequest, []byte, string, string) {
	t.Helper()
	fx, req, origin := recordedOriginFixture(t, "a")
	p := fx.profile
	q := "other-fixture"
	if err := os.Mkdir(ProfileDir(fx.home, q), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"source.json", "lock.json"} {
		raw, err := os.ReadFile(filepath.Join(ProfileDir(fx.home, p), name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ProfileDir(fx.home, q), name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	qreq := req
	qreq.Profile = q
	qreq.Repair = true
	if result, err := Resolve(qreq); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("disjoint Q setup owner: %v", err)
	}
	if err := SetCurrent(fx.home, q); err != nil {
		t.Fatal(err)
	}
	target := ManagedHomeDir(fx.home, p, "codex_cli")
	if err := os.WriteFile(filepath.Join(target, "session-fixture"), []byte("opaque retained fixture state"), 0600); err != nil {
		t.Fatal(err)
	}
	before := nativeSnapshot(t, target)
	return fx, req, origin, q, before
}

func removedOriginFixture(t *testing.T) (*managedFixture, ResolveRequest, []byte, string, string) {
	t.Helper()
	fx, req, origin, q, before := prepareRemovalOriginFixture(t)
	p := fx.profile
	target := ManagedHomeDir(fx.home, p, "codex_cli")
	if err := Remove(fx.home, p, false); err != nil {
		t.Fatalf("actual remove-retain: %v", err)
	}
	if _, err := os.Lstat(ProfileDir(fx.home, p)); !os.IsNotExist(err) {
		t.Fatalf("installed P still present: %v", err)
	}
	if got := nativeSnapshot(t, target); got != before {
		t.Fatal("remove changed retained home/origin/session fixture bytes")
	}
	return fx, req, origin, q, before
}

func TestProvisioningOriginRemoveRetainOwner(t *testing.T) {
	fx, req, origin, q, before := removedOriginFixture(t)
	p := fx.profile
	target := ManagedHomeDir(fx.home, p, "codex_cli")
	manager, err := envunit.ObserveManager(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	ledger, _, err := envlifecycle.Read(manager)
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger.Entries) != 1 || ledger.Entries[0].Removed == nil {
		t.Fatalf("missing permanent receipt: %+v", ledger)
	}
	entries, err := os.ReadDir(filepath.Join(fx.home, "state/transactions/v1"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("remove cleanup incomplete: %v %v", entries, err)
	}
	if err := envauthority.Validate(fx.home); err != nil {
		t.Fatalf("new common scan rejected intentional retention: %v", err)
	}
	calls := 0
	qreq := req
	qreq.Profile = q
	qreq.Repair = false
	qreq.originAccountLookup = func() (envorigin.Account, error) {
		calls++
		return envorigin.Account{}, fmt.Errorf("forbidden rediscovery")
	}
	// Independent native-skills custody can legitimately be consulted for Q;
	// preserve that fixture callback while still forbidding account discovery.
	qreq.NativeHomeOf = req.NativeHomeOf
	if result, err := Resolve(qreq); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("disjoint actual Q continuation: %v", err)
	}
	req.Repair = true
	req.originAccountLookup = qreq.originAccountLookup
	req.NativeHomeOf = func(id string) (string, error) { calls++; return "", fmt.Errorf("forbidden retained-P lookup") }
	result, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), "retained_profile_removed") || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
		t.Fatalf("intended retained-P refusal: %v", err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatal("removed P emitted fragment")
	}
	if calls != 0 {
		t.Fatalf("retained/disjoint continuation rediscovered authority %d times", calls)
	}
	raw, err := os.ReadFile(filepath.Join(target, envorigin.Name))
	if err != nil || !bytes.Equal(raw, origin) || nativeSnapshot(t, target) != before {
		t.Fatal("later Q/P operations changed retained P")
	}
}
