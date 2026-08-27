//go:build !windows

package scopes

import (
	"os"
	"testing"
)

// protectedTestHome builds a manager home the protected build cache accepts as
// its trust anchor: owner-only, writable by nobody else.
func protectedTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.Chmod(home, 0o700); err != nil {
		t.Fatal(err)
	}
	return home
}
