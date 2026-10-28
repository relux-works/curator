//go:build unix

package envorigin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Native-root metadata is not evidence of readable directory access. Probe only
// an owned directory handle (no names or payloads); a privileged process may
// retain access despite mode 000, which is a real outside-positive branch, not
// a skip or an invented permission failure. This is not an account witness.
func TestProvisioningOriginUnreadableDirectory(t *testing.T) {
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	native := filepath.Join(root, "native")
	for _, path := range []string{manager, native} {
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateExternal(manager, native); err != nil {
		t.Fatalf("readable outside positive: %v", err)
	}
	if err := os.Chmod(native, 0); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chmod(native, 0700); err != nil {
			t.Error(err)
		}
	}()
	handle, probe := os.Open(native)
	if handle != nil {
		if err := handle.Close(); err != nil {
			t.Fatal(err)
		}
	}
	err := ValidateExternal(manager, native)
	if probe == nil {
		t.Log("permission capability: this process can open the mode-000 owned directory; outside positive executes")
		if err != nil {
			t.Fatalf("readable privileged outside refused: %v", err)
		}
	} else {
		if !os.IsPermission(probe) {
			t.Fatalf("unclassified fixture directory probe: %v", probe)
		}
		if err == nil || !strings.Contains(err.Error(), DiagInvalid) {
			t.Fatalf("unreadable native root admitted: %v", err)
		}
	}
}
