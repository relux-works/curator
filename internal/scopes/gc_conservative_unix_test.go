//go:build unix

package scopes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCollectStaysFailSafeOnRedirectedUnixScopes proves a symlinked scope root
// or installed skill is refused rather than followed, and stays refused on the
// next pass.
//
// A redirected scope is the sharpest version of the hazard: the markers that
// name live build keys are still on disk, just not where maintenance looked.
// Following the link would be unsafe; forgetting the consumer behind it would
// be worse, because the pass after that would sweep artifacts those markers
// still reference.
func TestCollectStaysFailSafeOnRedirectedUnixScopes(t *testing.T) {
	tests := map[string]failSafeCase{
		"redirected project skill root": {
			arrange: func(t *testing.T, _, project string) {
				elsewhere := t.TempDir()
				installBuildMarker(t, elsewhere, "skill-hidden", strings.Repeat("a", 40), "tool", buildKey("1"))
				if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(elsewhere, filepath.Join(project, ".agents", "skills")); err != nil {
					t.Fatal(err)
				}
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
				if err := os.Symlink(filepath.Join(elsewhere, "skill-hidden"), filepath.Join(skills, "skill-hidden")); err != nil {
					t.Fatal(err)
				}
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
				if err := os.Symlink(elsewhere, filepath.Join(home, "global", "skills")); err != nil {
					t.Fatal(err)
				}
			},
			warning: "symbolic link or reparse point",
			verify:  func(*testing.T, string, string) {},
		},
		"redirected hybrid scope root": {
			arrange: func(t *testing.T, home, _ string) {
				elsewhere := t.TempDir()
				installBuildMarker(t, elsewhere, "skill-hidden", strings.Repeat("a", 40), "tool", buildKey("3"))
				if err := os.MkdirAll(filepath.Dir(HybridSkillsRoot(home)), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(elsewhere, HybridSkillsRoot(home)); err != nil {
					t.Fatal(err)
				}
			},
			warning: "symbolic link or reparse point",
			verify:  func(*testing.T, string, string) {},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) { runFailSafeAcrossTwoPasses(t, test) })
	}
}
