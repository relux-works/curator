//go:build unix

package envprofile

import (
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// TestResolveRejectsGroupWritableOnlyPathOverlay narrows the §4 permissions
// boundary through the production resolve entry. The fixture adds only the
// group-write bit, so a check that rejects other-writable paths but admits
// group-writable paths cannot pass this row.
func TestResolveRejectsGroupWritableOnlyPathOverlay(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root, overlay, policy := pathOverlayFixture(t, "")
	if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy}); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(overlay)
	if err != nil {
		t.Fatal(err)
	}
	originalMode := info.Mode().Perm()
	if err := os.Chmod(overlay, originalMode|0o020); err != nil {
		t.Fatalf("add only group-write permission to path overlay: %v", err)
	}
	defer func() {
		if err := os.Chmod(overlay, originalMode); err != nil {
			t.Errorf("restore path overlay mode: %v", err)
		}
	}()
	mutated, err := os.Lstat(overlay)
	if err != nil {
		t.Fatal(err)
	}
	if mutated.Mode().Perm()&0o022 != 0o020 {
		t.Fatalf("fixture permissions = %#o, want exactly group-write among group/other mutation bits", mutated.Mode().Perm())
	}

	result, err := Resolve(ResolveRequest{
		Home: home, Profile: "team-context", EnvID: "claude_code", LaunchDir: t.TempDir(),
		Machine: envregistry.DefaultMachineConfig(), Policy: policy,
		Detect: func(envregistry.Adapter) string { return "unknown" },
	})
	if err == nil || result != nil || !strings.Contains(err.Error(), DiagPathSourceUntrusted) || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("Resolve = (%v, %v), want no fragment and %s naming permissions", result, err, DiagPathSourceUntrusted)
	}
}
