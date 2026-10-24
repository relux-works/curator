//go:build windows

package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

// Real Git, bypassing Curator's admission gate, measures the adjacent usable
// and refused metadata roots. Both loose reads and file-transport pack creation
// must succeed below their measured transition; unrelated errors fail the test.
func TestWindowsGitMetadataLimit(t *testing.T) {
	src := makeRepo(t)
	root := windowsLongTestPath(t, t.TempDir())
	t.Setenv("GIT_TEMPLATE_DIR", t.TempDir())
	t.Logf("Git version: %s", gitRun(t, src, "version"))
	for _, packed := range []bool{false, true} {
		name := "loose"
		remote := src
		if packed {
			name = "packed"
			remote = "file:///" + filepath.ToSlash(src)
		}
		t.Run(name, func(t *testing.T) {
			probe := func(length int) bool {
				t.Helper()
				base := filepath.Join(root, name)
				if err := os.MkdirAll(base, 0o700); err != nil {
					t.Fatal(err)
				}
				padding := length - len(utf16.Encode([]rune(base))) - 1
				if padding < 1 || padding > 240 {
					t.Fatalf("probe has no valid padding: %d", padding)
				}
				dest := filepath.Join(base, strings.Repeat("c", padding))
				defer func() { _ = os.RemoveAll(dest) }()
				for _, args := range [][]string{
					{"clone", "--config", "core.longpaths=true", "--", remote, dest},
					{"-C", dest, "fetch", "--all", "--tags", "--prune"},
					{"-C", dest, "show", "v1:SKILL.md"},
				} {
					cmd := exec.Command("git", append([]string{"-c", "core.longpaths=true"}, args...)...)
					cmd.Env = gitEnv()
					output, err := cmd.CombinedOutput()
					if err != nil {
						if !strings.Contains(string(output), "Filename too long") {
							t.Fatalf("unrelated Git probe failure at root=%d: %v\n%s", length, err, output)
						}
						t.Logf("root=%d refused: %s", length, strings.TrimSpace(string(output)))
						return false
					}
					if args[len(args)-1] == "v1:SKILL.md" && strings.TrimSpace(string(output)) != "v1" {
						t.Fatalf("probe did not read exact payload: %s", output)
					}
				}
				return true
			}
			low, high := 160, 214
			if !probe(low) || probe(high) {
				t.Fatal("measurement needs a usable lower root and a path-length refusal above it")
			}
			for high-low > 1 {
				mid := low + (high-low)/2
				if probe(mid) {
					low = mid
				} else {
					high = mid
				}
			}
			if !probe(low) || probe(high) {
				t.Fatal("adjacent boundary controls were not reproducible")
			}
			suffix := filepath.Join(".git", "objects", "00", strings.Repeat("0", 38))
			if packed {
				suffix = filepath.Join(".git", "objects", "pack", "pack-"+strings.Repeat("0", 40)+".keep")
			}
			t.Logf("measured %s root bound=%d; first refused=%d; metadata bound=%d UTF-16 units; suffix=%s", name, low, high, low+1+len(suffix), suffix)
		})
	}
}

func windowsLongTestPath(t *testing.T, path string) string {
	t.Helper()
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetLongPathName(p, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || int(n) >= len(buf) {
		t.Fatal(fmt.Errorf("expand fixture path: length=%d: %w", n, err))
	}
	return windows.UTF16ToString(buf[:n])
}
