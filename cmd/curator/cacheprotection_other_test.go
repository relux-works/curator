//go:build !windows

package main

import (
	"os"
	"testing"
)

// restoreCacheProtection adds nothing outside Windows: the snapshot restore
// already puts back the mode bits, which are the whole protected boundary here.
func restoreCacheProtection(*testing.T, string, bool) {}

// breakCacheProtection makes a cache entry writable by group and other, which
// is what "the protected boundary stops being provable" means on a unix host.
func breakCacheProtection(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o777); err != nil {
		t.Fatal(err)
	}
}
