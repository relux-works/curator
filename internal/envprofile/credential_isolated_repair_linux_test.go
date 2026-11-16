//go:build linux

package envprofile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

// After a profile update changes the lock, the managed home is stale and a
// repair reprojects it. On Linux an isolated claude_code home owns its own
// .credentials.json. Repair must keep those bytes and that file identity
// exactly, so a profile rollout never touches an isolated refresh family.
// The synthetic credential is never a real token. The refusal controls for
// mis-targeted, regular-file and unrecorded credential links live in
// credential_link_test.go (TestStaleCredentialLinkRefusals and its rows).
func TestRepairKeepsIsolatedClaudeCredentialAfterLockChange(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.request("claude_code")
	req.Machine.Isolation = map[string]map[string]string{"acme": {"claude_code": "isolated"}}
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatalf("provision isolated claude_code home: %v", err)
	}
	credential := filepath.Join(ManagedHomeDir(fx.home, "acme", "claude_code"), ".credentials.json")
	synthetic := []byte("{\"synthetic\":\"isolated-family-fixture\"}\n")
	if err := os.WriteFile(credential, synthetic, 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Lstat(credential)
	if err != nil {
		t.Fatal(err)
	}

	lock, _, err := contextlock.Read(lockPath(fx.home, "acme"))
	if err != nil {
		t.Fatal(err)
	}
	kept := lock.Members[:0]
	for _, member := range lock.Members {
		if member.Kind != "skill" {
			kept = append(kept, member)
		}
	}
	lock.Members = kept
	lock.Sort()
	if _, err := contextlock.Write(lockPath(fx.home, "acme"), lock); err != nil {
		t.Fatal(err)
	}

	stale := req
	stale.Repair = false
	if _, err := Resolve(stale); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("resolve after the lock change = %v; want %s", err, DiagHomeStale)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatalf("repair after the lock change: %v", err)
	}
	after, err := os.Lstat(credential)
	if err != nil {
		t.Fatalf("isolated credential missing after repair: %v", err)
	}
	if !after.Mode().IsRegular() || !os.SameFile(before, after) {
		t.Fatalf("repair replaced the isolated credential: mode %v, same file %v", after.Mode(), os.SameFile(before, after))
	}
	got, err := os.ReadFile(credential)
	if err != nil || !bytes.Equal(got, synthetic) {
		t.Fatalf("repair changed the isolated credential bytes: %v", err)
	}
	if _, err := Resolve(stale); err != nil {
		t.Fatalf("resolve after repair: %v", err)
	}
}
