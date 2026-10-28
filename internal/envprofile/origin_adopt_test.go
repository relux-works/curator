package envprofile

import (
	"bytes"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Proves the real adoption owner writes only missing explicit authority,
// preserves the legacy home, and retries byte-identically without discovery.
// Wrong attested bindings and wrong credential targets preserve all preimages.
func TestProvisioningOriginAdoptionOwner(t *testing.T) {
	for _, state := range []string{"healthy", "wrong-lock", "wrong-target", "missing-link"} {
		t.Run(state, func(t *testing.T) {
			fx, _, raw := recordedOriginFixture(t, "a")
			var record envorigin.Record
			recordPtr, err := envorigin.Parse(raw, envorigin.Binding{Profile: fx.profile, LockSHA256: "sha256:" + strings.TrimPrefix(fx.lockHash, "sha256:"), ManagedTarget: ManagedHomeDir(fx.home, fx.profile, "codex_cli")}, "codex_cli", fx.home)
			if err != nil {
				t.Fatal(err)
			}
			record = *recordPtr
			target := record.Binding.ManagedTarget
			path := filepath.Join(target, envorigin.Name)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if state == "wrong-target" || state == "missing-link" {
				if err := os.Remove(filepath.Join(target, "auth.json")); err != nil {
					t.Fatal(err)
				}
				if state == "wrong-target" {
					if err := os.Symlink(filepath.Join(t.TempDir(), "auth.json"), filepath.Join(target, "auth.json")); err != nil {
						t.Fatal(err)
					}
				}
			}
			before := nativeSnapshot(t, target)
			request := AdoptOriginRequest{Home: fx.home, Machine: envregistry.DefaultMachineConfig(), Attestation: envorigin.AdoptRequest{Adapter: "codex_cli", Binding: record.Binding, OperatorHome: record.OperatorHome, SelectorState: record.Selector.State, SelectorValue: record.Selector.Value, NativeRoot: record.NativeRoot}}
			if state == "wrong-lock" {
				request.Attestation.Binding.LockSHA256 = "sha256:" + strings.Repeat("f", 64)
			}
			err = AdoptOrigin(request)
			if state != "healthy" {
				diagnostic := envorigin.DiagInvalid
				if state != "wrong-lock" {
					diagnostic = envregistry.DiagCredentialConflict
				}
				if err == nil || !strings.Contains(err.Error(), diagnostic) {
					t.Fatalf("intended adoption refusal: %v", err)
				}
				if nativeSnapshot(t, target) != before {
					t.Fatal("adoption refusal changed home")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			adopted, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := envorigin.Parse(adopted, record.Binding, "codex_cli", fx.home)
			if err != nil || parsed.Capture.Method != "adoption" {
				t.Fatalf("adopted carrier: %v", err)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if nativeSnapshot(t, target) != before {
				t.Fatal("adoption changed another entry")
			}
			if err := os.WriteFile(path, adopted, 0600); err != nil {
				t.Fatal(err)
			}
			if err := AdoptOrigin(request); err != nil {
				t.Fatal(err)
			}
			again, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(again, adopted) {
				t.Fatal("identical attestation rewrote capture")
			}
		})
	}
}
