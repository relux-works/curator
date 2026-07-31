//go:build !windows

package main

import (
	"os"
	"testing"
)

// restoreCacheProtection puts the protected boundary of one cache node back.
// Outside Windows the boundary is the mode bits, so owner-only is what it
// restores; a caller that also replays a recorded mode may overwrite this
// afterwards with the same intent.
func restoreCacheProtection(t *testing.T, path string, directory bool) {
	t.Helper()
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// breakCacheProtection makes a cache entry writable by group and other, which
// is what "the protected boundary stops being provable" means on a unix host.
func breakCacheProtection(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o777); err != nil {
		t.Fatal(err)
	}
}

// foreignArtifactPath names an artifact this target does not derive, so a
// marker recording it is genuine target drift. A unix target derives
// bin/build-tool, so the foreign spelling here is the Windows one.
func foreignArtifactPath() string { return "bin/build-tool.exe" }
