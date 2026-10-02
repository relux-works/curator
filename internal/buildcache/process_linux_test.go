package buildcache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxExecutablePaths(t *testing.T) {
	for _, row := range []struct {
		name, image, stat string
		fail              bool
	}{
		{name: "live", image: "/cache/build/bin/runner"},
		{name: "deleted", image: "/cache/build/bin/runner (deleted)"},
		{name: "exited"},
		{name: "kernel", stat: "12 (kernel worker) S 0 0 0 0 0 2097152"},
		{name: "zombie", stat: "12 (runner) Z 0 0 0 0 0 0"},
		{name: "missing-live-image", stat: "12 (runner) S 0 0 0 0 0 0", fail: true},
		{name: "malformed-stat", stat: "malformed", fail: true},
	} {
		t.Run(row.name, func(t *testing.T) {
			proc := t.TempDir()
			dir := filepath.Join(proc, "12")
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			if row.image != "" {
				if err := os.Symlink(row.image, filepath.Join(dir, "exe")); err != nil {
					t.Fatal(err)
				}
			}
			if row.stat != "" {
				if err := os.WriteFile(filepath.Join(dir, "stat"), []byte(row.stat), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			paths, err := linuxExecutablePaths(proc)
			if (err != nil) != row.fail {
				t.Fatalf("paths=%v error=%v", paths, err)
			}
			if row.image != "" && (len(paths) != 1 || paths[0] != "/cache/build/bin/runner") {
				t.Fatalf("paths=%v", paths)
			}
		})
	}
	if _, err := linuxExecutablePaths(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("unreadable process table accepted")
	}
}
