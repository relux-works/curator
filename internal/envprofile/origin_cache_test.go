package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// Actual repair reloads backend evidence at the lock boundary. A file→keyring
// change at recorded A cannot reuse the pre-lock successful file observation,
// remove its old auth link, rerender surfaces or emit a fragment. The nearby
// unchanged backend repair reaches publication instead of rejecting all cases.
func TestProvisioningOriginBackendCacheBoundary(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "changed"}[changed], func(t *testing.T) {
			fx, req, _ := recordedOriginFixture(t, "a")
			req.Detect = func(envregistry.Adapter) string { return "unknown" }
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if e := os.Remove(filepath.Join(target, "AGENTS.md")); e != nil {
				t.Fatal(e)
			}
			before := nativeSnapshot(t, target)
			reached := false
			req.Repair = true
			req.beforeRepairLock = func() {
				reached = true
				if changed {
					if e := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte("cli_auth_credentials_store = \"keyring\"\n"), 0600); e != nil {
						t.Fatal(e)
					}
				}
			}
			result, e := Resolve(req)
			if !reached {
				t.Fatal("actual repair lock boundary not reached")
			}
			if changed {
				if e == nil || !strings.Contains(e.Error(), envregistry.DiagCredentialConflict) {
					t.Fatalf("intended backend transition refusal: %v", e)
				}
				if result != nil && len(result.Document) != 0 {
					t.Fatal("backend refusal emitted fragment")
				}
				if nativeSnapshot(t, target) != before {
					t.Fatal("backend refusal mutated managed preimages")
				}
				return
			}
			if e != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("healthy actual repair: %v", e)
			}
			if _, e := os.Lstat(filepath.Join(target, "AGENTS.md")); e != nil {
				t.Fatal("healthy repair failed to publish")
			}
		})
	}
}
