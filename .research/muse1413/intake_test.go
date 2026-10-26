package muse1413

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The real vendor test entry skips without opt-in and rejects wrong bytes
// before creating its run root. The gated replay supplies the exact-byte control.
func TestVendorIntakeBoundary(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, shape := range []string{"unset", "wrong-digest"} {
		t.Run(shape, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), ".temp", "uncreated")
			t.Setenv("MUSE1413_RUN_ROOT", root)
			binary := ""
			if shape == "wrong-digest" {
				binary = filepath.Join(t.TempDir(), "foreign-build")
				write(t, binary, []byte("synthetic unqualified binary\n"))
			}
			t.Setenv("MUSE1413_BINARY", binary)
			out, code := command(t, ".", os.Environ(), self, "-test.run=^TestMuse1413OfflineBoundary$", "-test.v")
			want, message := 0, "--- SKIP: TestMuse1413OfflineBoundary"
			if shape == "wrong-digest" {
				want, message = 1, "unqualified binary digest"
			}
			if code != want || !strings.Contains(string(out), message) {
				t.Fatalf("intake exit=%d want=%d: %s", code, want, out)
			}
			if _, err := os.Lstat(root); !os.IsNotExist(err) {
				t.Fatalf("intake created fixture or unreadable root: %v", err)
			}
		})
	}
}

// Digest-gated version inspection runs under the same offline sandbox and
// disposable environment as the loader, never the updating launcher.
func TestMuse1413Identity(t *testing.T) {
	p := setup(t)
	for _, sub := range []string{"native", "tmp", "config", "data", "state", "cache"} {
		if err := os.MkdirAll(filepath.Join(p.root, sub), 0700); err != nil {
			t.Fatal(err)
		}
	}
	args := append([]string{"-f", p.sandbox, "/usr/bin/env", "-i"}, p.env(p.root)...)
	args = append(args, p.binary, "--provider", "echo", "--version")
	out, code := command(t, p.root, p.env(p.root), "/usr/bin/sandbox-exec", args...)
	write(t, filepath.Join(p.root, "version.log"), out)
	if code != 0 || strings.TrimSpace(string(out)) != "Muse Code 1.4.13 (1.4.13-R6996.1)" {
		t.Fatalf("version exit=%d: %s", code, out)
	}
	t.Log("version exit=0; Muse Code 1.4.13 (1.4.13-R6996.1); exact digest verified")
}
