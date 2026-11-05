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
}
