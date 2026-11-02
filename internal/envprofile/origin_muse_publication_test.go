package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

// The actual Muse owner publishes empty runtime directories durably, preserves
// strict native-file liveness, and restores all home preimages on metadata fault.
// The account seam supplies raw fixture facts, not an authority verdict or a
// claim about production account lookup/Windows rename semantics.
func TestProvisioningOriginMusePublication(t *testing.T) {
	for _, state := range []string{"healthy", "missing-native", "directory-fault"} {
		t.Run(state, func(t *testing.T) {
			enableV1WritersForTest(t)
			fx := writeManagedFixture(t, "origin-muse")
			anchor, xdg := t.TempDir(), t.TempDir()
			native := filepath.Join(xdg, "muse")
			if err := os.Mkdir(native, 0700); err != nil {
				t.Fatal(err)
			}
			if state != "missing-native" {
				if err := os.WriteFile(filepath.Join(native, "auth.json"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("XDG_CONFIG_HOME", xdg)
			fx.native[envregistry.Muse] = native
			req := fx.request(envregistry.Muse)
			req.Repair = true
			req.originAccountLookup = func() (envorigin.Account, error) {
				identity, err := envorigin.LookupIdentity()
				identity.Home = anchor
				return identity, err
			}
			fault := errors.New("fixture Muse metadata publication fault")
			reached := false
			if state == "directory-fault" {
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == transaction.PointAfterDirectoryInstall {
						reached = true
						return fault
					}
					return nil
				}})}
			}
			result, err := Resolve(req)
			home := ManagedHomeDir(fx.home, fx.profile, envregistry.Muse)
			if state != "healthy" {
				if state == "directory-fault" {
					if !reached || !errors.Is(err, fault) {
						t.Fatalf("intended metadata fault not reached: %v", err)
					}
				} else if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
					t.Fatalf("strict Muse liveness refusal: %v", err)
				}
				if result != nil && len(result.Document) != 0 {
					t.Fatal("failure emitted fragment")
				}
				if _, err := os.Lstat(home); !os.IsNotExist(err) {
					t.Fatalf("failed publication left a home: %v", err)
				}
				return
			}
			if err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("healthy Muse continuation: %v %v", result, err)
			}
			for _, rel := range []string{"config/muse", "data/muse", "state", "cache"} {
				info, err := os.Lstat(filepath.Join(home, rel))
				if err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
					t.Fatalf("runtime directory %s: %v %v", rel, info, err)
				}
			}
			if text, err := os.Readlink(filepath.Join(home, "config/muse/auth.json")); err != nil || text != filepath.Join(native, "auth.json") {
				t.Fatalf("recorded Muse source: %q %v", text, err)
			}
		})
	}
}
