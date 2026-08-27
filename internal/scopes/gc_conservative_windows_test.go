//go:build windows

package scopes

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCollectStaysFailSafeOnRedirectedWindowsScopes is the Windows half of the
// redirected-scope rule: a junction or symbolic link standing in for a skills
// directory must be refused, reported, and still refused on the next pass.
//
// Junctions are the important case here — they need no privilege, so any local
// account can create one, and Go has reported them as different mode bits in
// different releases.
func TestCollectStaysFailSafeOnRedirectedWindowsScopes(t *testing.T) {
	tests := map[string]failSafeCase{
		"redirected project skill root": {
			arrange: func(t *testing.T, _, project string) {
				elsewhere := t.TempDir()
				installBuildMarker(t, elsewhere, "skill-hidden", strings.Repeat("a", 40), "tool", buildKey("1"))
				if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
					t.Fatal(err)
				}
				linkDirectoryForTest(t, elsewhere, filepath.Join(project, ".agents", "skills"))
			},
			warning: "symbolic link or reparse point",
			verify:  registryKeeps,
		},
		"redirected installed skill": {
			arrange: func(t *testing.T, _, project string) {
				elsewhere := t.TempDir()
				installBuildMarker(t, elsewhere, "skill-hidden", strings.Repeat("a", 40), "tool", buildKey("1"))
				skills := filepath.Join(project, ".agents", "skills")
				if err := os.MkdirAll(skills, 0o755); err != nil {
					t.Fatal(err)
				}
				linkDirectoryForTest(t, filepath.Join(elsewhere, "skill-hidden"), filepath.Join(skills, "skill-hidden"))
			},
			warning: "symbolic link or reparse point",
			verify:  registryKeeps,
		},
		"redirected global scope root": {
			arrange: func(t *testing.T, home, _ string) {
				elsewhere := t.TempDir()
				installBuildMarker(t, elsewhere, "skill-hidden", strings.Repeat("a", 40), "tool", buildKey("2"))
				if err := os.MkdirAll(filepath.Join(home, "global"), 0o755); err != nil {
					t.Fatal(err)
				}
				linkDirectoryForTest(t, elsewhere, filepath.Join(home, "global", "skills"))
			},
			warning: "symbolic link or reparse point",
			verify:  func(*testing.T, string, string) {},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) { runFailSafeAcrossTwoPasses(t, test) })
	}
}

// linkDirectoryForTest creates a directory reparse point, preferring a symbolic
// link and falling back to a junction. A junction needs no privilege, so this
// keeps the redirected-scope cases running on a plain local account instead of
// skipping the coverage the acceptance criteria ask for.
func linkDirectoryForTest(t *testing.T, target, link string) {
	t.Helper()
	symlinkErr := os.Symlink(target, link)
	if symlinkErr == nil {
		return
	}
	t.Logf("symbolic links are unavailable to this account (%v); using a junction", symlinkErr)
	output, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Skipf("this host cannot create a directory reparse point: %v: %s", err, output)
	}
}
