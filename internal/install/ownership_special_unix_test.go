//go:build !windows

package install

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func ownershipSpecial(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
}
