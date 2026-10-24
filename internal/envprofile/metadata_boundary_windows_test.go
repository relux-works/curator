//go:build windows

package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
)

// The manager propagates the typed refusal before creating its repository
// parent. A neighboring admitted source-bound cache still clones and binds
// its original network identity; neither path changes existing home bytes.
func TestWindowsMCPMetadataRootRefusal(t *testing.T) {
	ids := newGitIdentities(t)
	source := "packages.example/tool"
	ids.serve(gitRepo(t, map[string]string{"SKILL.md": "payload"}, "v1"), "https://"+source)
	for _, refused := range []bool{false, true} {
		name := "valid"
		if refused {
			name = "refused"
		}
		t.Run(name, func(t *testing.T) {
			home := gitFixtureHome(t)
			if refused {
				home = filepath.Join(home, strings.Repeat("long", 20))
			}
			if err := os.MkdirAll(home, 0o700); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(home, "sentinel")
			if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			manager := newProjectMCPGitManager(home)
			manager.recordRaw(source, "https://"+source)
			repo, err := manager.ensureRepo(source)
			if refused {
				var refusal *gitops.MetadataPathError
				if !errors.As(err, &refusal) || refusal.Actual <= refusal.Limit || !strings.Contains(err.Error(), "profile_source_invalid") {
					t.Fatalf("want typed manager metadata refusal: %v", err)
				}
				if _, err := os.Lstat(manager.reposDir()); !os.IsNotExist(err) {
					t.Fatalf("metadata refusal created repository parent: %v", err)
				}
			} else if err != nil || repo != manager.repoDir(source) {
				t.Fatalf("supported cache acquisition failed: %s %v", repo, err)
			}
			if body, err := os.ReadFile(marker); err != nil || string(body) != "keep" {
				t.Fatalf("home sentinel changed: %q %v", body, err)
			}
		})
	}
}
