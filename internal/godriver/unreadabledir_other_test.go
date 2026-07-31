//go:build !windows

package godriver

import (
	"os"
	"testing"
)

// denyDirectoryListing takes away the right to list one directory and gives it
// back when the test ends. Outside Windows that right is a mode bit, so mode
// 0o000 is the whole of it.
//
// The restore has to run before the temporary root is removed. t.Cleanup is
// last-in-first-out and t.TempDir registered its removal first, so registering
// here is what orders them.
func denyDirectoryListing(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
}
