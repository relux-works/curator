package envprofile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
)

// Real generic publication and activation owners cannot overwrite retained P,
// its installed namespace or the private ledger. Disjoint Q publication still
// reaches its write. A nil unit/ambient current profile grants no exception.
func TestProvisioningOriginLifecycleEffectAdmission(t *testing.T) {
	for _, target := range []string{"installed", "home", "ledger", "activation"} {
		t.Run(target, func(t *testing.T) {
			fx, _, _, q, before := removedOriginFixture(t)
			manager, err := envunit.ObserveManager(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			_, receipt, err := envlifecycle.Read(manager)
			if err != nil {
				t.Fatal(err)
			}
			var path string
			switch target {
			case "installed":
				path = sourcePath(fx.home, fx.profile)
			case "home":
				path = filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envorigin.Name)
			case "ledger":
				path = envlifecycle.Path(fx.home)
			}
			if target == "activation" {
				err = SetCurrent(fx.home, fx.profile)
			} else {
				op, e := beginOperation(fx.home)
				if e != nil {
					t.Fatal(e)
				}
				err = op.publish(map[string][]byte{path: []byte("forbidden fixture replacement")})
				_ = op.close()
			}
			if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
				t.Fatalf("intended lifecycle effect refusal: %v", err)
			}
			_, after, e := envlifecycle.Read(manager)
			if e != nil || !bytes.Equal(after, receipt) || nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, "codex_cli")) != before {
				t.Fatal("refused owner changed retained state")
			}
			if _, e := os.Lstat(ProfileDir(fx.home, fx.profile)); !os.IsNotExist(e) {
				t.Fatalf("refused owner installed P: %v", e)
			}
			op, e := beginOperation(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			healthy := filepath.Join(ProfileDir(fx.home, q), "effect-control")
			e = op.publish(map[string][]byte{healthy: []byte("healthy Q")})
			_ = op.close()
			if e != nil {
				t.Fatalf("disjoint owner rejected: %v", e)
			}
			raw, e := os.ReadFile(healthy)
			if e != nil || string(raw) != "healthy Q" {
				t.Fatal("disjoint owner did not reach actual write")
			}
		})
	}
}

// A generic manager publisher cannot replace the shared lease ledger under a
// disjoint physical name and bypass its semantic profile-reference policy.
func TestProvisioningOriginLeaseLedgerOwnerRequired(t *testing.T) {
	fx, _, _, q, before := removedOriginFixture(t)
	op, e := beginOperation(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := op.close(); err != nil {
			t.Error(err)
		}
	}()
	path := filepath.Join(fx.home, "state", "project-leases", "v1", "ledger.json")
	e = op.publish(map[string][]byte{path: []byte("forbidden generic ledger")})
	if e == nil || !strings.Contains(e.Error(), envorigin.DiagInvalid) {
		t.Fatalf("intended lease owner refusal: %v", e)
	}
	if _, e = os.Lstat(path); !os.IsNotExist(e) {
		t.Fatal("generic publisher changed lease ledger")
	}
	if nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, "codex_cli")) != before {
		t.Fatal("refused publisher changed retained home")
	}
	if e = op.publish(map[string][]byte{filepath.Join(ProfileDir(fx.home, q), "ledger-control"): []byte("reached")}); e != nil {
		t.Fatal(e)
	}
}
