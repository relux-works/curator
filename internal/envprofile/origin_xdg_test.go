package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
)

// Actual repair preserves credential-origin A while OpenCode reconciles the
// separate current operator XDG seed contract. Muse never seeds foreign XDG
// entries. Ambient data/state trees stay byte-equivalent. Account lookup is a
// raw fixture seam only; this does not qualify a native account/provider.
func TestProvisioningOriginXDGContracts(t *testing.T) {
	for _, adapter := range []string{"opencode", "muse"} {
		for _, state := range []string{"retarget", "unreadable", "shadow"} {
			t.Run(adapter+"/"+state, func(t *testing.T) {
				enableV1WritersForTest(t)
				fx := writeManagedFixture(t, "xdg-origin")
				a, b := t.TempDir(), t.TempDir()
				for _, root := range []string{a, b} {
					for _, name := range []string{"git", "gh", adapter} {
						if err := os.MkdirAll(filepath.Join(root, name), 0700); err != nil {
							t.Fatal(err)
						}
					}
					if err := os.WriteFile(filepath.Join(root, "git", "config"), []byte("synthetic config"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(filepath.Join(a, adapter, "auth.json"), nil, 0600); err != nil {
					t.Fatal(err)
				}
				// gh appears only after provisioning; a live managed gh entry is foreign.
				if err := os.Remove(filepath.Join(a, "gh")); err != nil {
					t.Fatal(err)
				}
				t.Setenv("XDG_CONFIG_HOME", a)
				fx.native[adapter] = filepath.Join(a, adapter)
				req := fx.request(adapter)
				req.OperatorXDG = a
				req.Repair = true
				if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("initial actual XDG owner: %v", err)
				}
				target := ManagedHomeDir(fx.home, fx.profile, adapter)
				parent := ManagedParent(fx.home, fx.profile, adapter)
				originPath := filepath.Join(target, envorigin.Name)
				origin, err := os.ReadFile(originPath)
				if err != nil {
					t.Fatal(err)
				}
				if adapter == "muse" {
					if err := os.WriteFile(filepath.Join(b, "muse", "settings.json"), []byte("foreign B settings"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if state == "shadow" && adapter == "opencode" {
					if err := os.WriteFile(filepath.Join(parent, "gh"), []byte("unrecorded managed shadow"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if state == "unreadable" {
					if err := os.RemoveAll(filepath.Join(b, "git")); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink("git", filepath.Join(b, "git")); err != nil {
						t.Fatal(err)
					}
				}
				data, st := t.TempDir(), t.TempDir()
				for _, root := range []string{data, st} {
					if err := os.WriteFile(filepath.Join(root, "ambient"), []byte("unrelated data/state"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				t.Setenv("XDG_CONFIG_HOME", b)
				t.Setenv("XDG_DATA_HOME", data)
				t.Setenv("XDG_STATE_HOME", st)
				beforeData, beforeState := nativeSnapshot(t, data), nativeSnapshot(t, st)
				beforeA, beforeB := nativeSnapshot(t, a), nativeSnapshot(t, b)
				beforeHome := nativeSnapshot(t, parent)
				req.OperatorXDG = b
				req.originAccountLookup = func() (envorigin.Account, error) {
					t.Error("maintenance rediscovered account")
					return envorigin.Account{}, fmt.Errorf("forbidden discovery")
				}
				result, err := Resolve(req)
				if adapter == "opencode" && state == "unreadable" {
					if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
						t.Fatalf("intended XDG seed read refusal: %v", err)
					}
					if result != nil && len(result.Document) != 0 {
						t.Fatal("seed refusal emitted fragment")
					}
					if nativeSnapshot(t, parent) != beforeHome {
						t.Fatal("seed refusal changed home preimages")
					}
				} else {
					if err != nil || result == nil || len(result.Document) == 0 {
						t.Fatalf("healthy XDG maintenance continuation: %v", err)
					}
					if adapter == "opencode" {
						if got, err := os.Readlink(filepath.Join(parent, "git")); err != nil || got != filepath.Join(b, "git") {
							t.Fatalf("current XDG seed failed to reconcile: %q %v", got, err)
						}
						if state == "shadow" {
							raw, err := os.ReadFile(filepath.Join(parent, "gh"))
							if err != nil || string(raw) != "unrecorded managed shadow" {
								t.Fatal("repair took over unrecorded XDG shadow")
							}
							warned := false
							for _, warning := range result.Warnings {
								warned = warned || strings.Contains(warning, envregistry.DiagSeedShadowed)
							}
							if !warned {
								t.Fatal("repair lost XDG shadow warning")
							}
						}
					} else {
						if got, err := os.Readlink(filepath.Join(target, "config/muse/auth.json")); err != nil || got != filepath.Join(a, "muse/auth.json") {
							t.Fatalf("Muse credential origin changed to B: %q %v", got, err)
						}
						for _, name := range []string{"git", "gh"} {
							if _, err := os.Lstat(filepath.Join(target, "config", name)); !os.IsNotExist(err) {
								t.Fatalf("Muse seeded foreign XDG %s: %v", name, err)
							}
						}
						if raw, err := os.ReadFile(filepath.Join(target, "config/muse/settings.json")); err == nil && strings.Contains(string(raw), "foreign B settings") {
							t.Fatal("Muse imported foreign B settings")
						}
					}
				}
				if raw, err := os.ReadFile(originPath); err != nil || string(raw) != string(origin) {
					t.Fatal("XDG reconciliation recaptured origin")
				}
				if nativeSnapshot(t, a) != beforeA || nativeSnapshot(t, b) != beforeB || nativeSnapshot(t, data) != beforeData || nativeSnapshot(t, st) != beforeState {
					t.Fatal("XDG maintenance changed native or ambient data/state trees")
				}
			})
		}
	}
}
