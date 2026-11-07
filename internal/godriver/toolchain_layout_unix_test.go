//go:build unix

package godriver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Claim: a self-contained toolchain fingerprints, while Debian-shaped src
// links outside GOROOT keep their exact refusal code and explain the upstream
// tree remedy. Neither refusal rewrites the link or the external source file.
func TestDistroToolchainLayoutRemedy(t *testing.T) {
	for _, row := range []struct {
		name, code string
		absolute   bool
	}{
		{name: "clean-upstream"},
		{name: "escaping-src", code: "toolchain_link_escape"},
		{name: "absolute-src", code: "toolchain_link_absolute", absolute: true},
	} {
		t.Run(row.name, func(t *testing.T) {
			base := t.TempDir()
			root := filepath.Join(base, "go")
			outside := filepath.Join(base, "packaged-src")
			if err := os.MkdirAll(root, 0o755); err != nil {
				t.Fatal(err)
			}
			copyTestToolchain(t, testToolchain(t), root)
			if err := os.MkdirAll(outside, 0o755); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(outside, "runtime.go")
			if err := os.WriteFile(source, []byte("unchanged source\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			target := "../packaged-src"
			if row.absolute {
				target = outside
			}
			if row.code == "" {
				if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink(target, filepath.Join(root, "src")); err != nil {
				t.Fatal(err)
			}
			fingerprint, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 "+runtime.GOOS+"/"+runtime.GOARCH)
			if row.code == "" {
				if err != nil || fingerprint == "" {
					t.Fatalf("closed tree refused: %s %v", fingerprint, err)
				}
				return
			}
			var d *Diagnostic
			if !errors.As(err, &d) || d.Code != row.code || d.Remedy != "select a clean upstream Go tree; set CURATOR_GO to its real GOROOT/bin/go" {
				t.Fatalf("wrong refusal/remedy: %v", err)
			}
			got, readErr := os.Readlink(filepath.Join(root, "src"))
			if readErr != nil || got != target {
				t.Fatalf("refusal altered source link: %q %v", got, readErr)
			}
			content, readErr := os.ReadFile(source)
			if readErr != nil || string(content) != "unchanged source\n" {
				t.Fatalf("refusal altered external source: %q %v", content, readErr)
			}
		})
	}
}
