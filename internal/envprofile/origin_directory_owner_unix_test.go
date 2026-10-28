//go:build unix

package envprofile

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
)

// Even an ambient Claude backend cannot turn an unopenable recorded root into
// launch authority. Healthy recorded maintenance reaches a fragment; refusal
// precedes release/account/current-root lookup and preserves managed state.
// Only fixture directory handles are probed, never an actual native account.
func TestProvisioningOriginUnreadableDirectoryOwner(t *testing.T) {
	fx, _, _ := recordedOriginFixture(t, "a")
	req := fx.request(envregistry.ClaudeCode)
	req.Repair = true
	if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("healthy Claude provisioning: %v", err)
	}
	req.Repair = false
	if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("healthy recorded Claude continuation: %v", err)
	}
	root := fx.native[envregistry.ClaudeCode]
	if err := os.Chmod(root, 0); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chmod(root, 0700); err != nil {
			t.Error(err)
		}
	}()
	handle, probe := os.Open(root)
	if handle != nil {
		if err := handle.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if probe == nil {
		t.Log("permission capability: fixture directory remains readable; corresponding policy positive executes in envorigin")
		return
	}
	if !os.IsPermission(probe) {
		t.Fatalf("unclassified fixture permission probe: %v", probe)
	}
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	lookups := 0
	req.Detect = func(envregistry.Adapter) string { lookups++; return "unknown" }
	req.NativeHomeOf = func(string) (string, error) { lookups++; return "", fmt.Errorf("forbidden current caller lookup") }
	req.originAccountLookup = func() (envorigin.Account, error) {
		lookups++
		return envorigin.Account{}, fmt.Errorf("forbidden account lookup")
	}
	before := nativeSnapshot(t, EnvRoot(fx.home))
	for _, repair := range []bool{false, true} {
		req.Repair = repair
		result, err := Resolve(req)
		if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || !strings.Contains(err.Error(), "directory metadata") {
			t.Fatalf("intended unreadable origin refusal repair=%t: %v", repair, err)
		}
		if result != nil && len(result.Document) != 0 {
			t.Fatal("unreadable root emitted fragment")
		}
		if lookups != 0 {
			t.Fatalf("refusal reached backend/current/account lookup: %d", lookups)
		}
		if nativeSnapshot(t, EnvRoot(fx.home)) != before {
			t.Fatal("unreadable origin refusal changed managed preimages")
		}
	}
}
