//go:build unix

package envorigin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Proves raw lookup evidence cannot substitute the invoking process identity.
// Numeric process IDs are local scalars; no account database/home is consulted.
// This fixture seam does not qualify the production account transport.
func TestProvisioningOriginAccountLookupBinding(t *testing.T) {
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	anchor := filepath.Join(root, "operator")
	if err := os.Mkdir(anchor, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", filepath.Join(root, "native"))
	b := Binding{Profile: "p", LockSHA256: "sha256:" + strings.Repeat("a", 64), ManagedTarget: ManagedTarget(manager, "p", "codex_cli")}
	uid, euid := uint64(os.Getuid()), uint64(os.Geteuid())
	if uid != euid {
		t.Fatal("fixture requires an ordinary same-identity process")
	}
	valid := func() (Account, error) {
		return Account{UID: euid, RealUID: uid, EffectiveUID: euid, Home: anchor}, nil
	}
	if _, err := Select(manager, b, "codex_cli", valid); err != nil {
		t.Fatalf("valid raw account control: %v", err)
	}
	forged := func() (Account, error) {
		return Account{UID: euid + 1, RealUID: euid + 1, EffectiveUID: euid + 1, Home: anchor}, nil
	}
	if _, err := Select(manager, b, "codex_cli", forged); err == nil || !strings.Contains(err.Error(), DiagInvalid) {
		t.Fatalf("lookup replaced actual process identity: %v", err)
	}
}
