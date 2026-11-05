//go:build !windows

package install

import (
	"os"
	"testing"
)

func denyProjectLedgerRead(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0600) })
	if _, err := os.ReadFile(path); !os.IsPermission(err) {
		t.Fatalf("unreadable-state fixture did not deny a real read: %v", err)
	}
}
