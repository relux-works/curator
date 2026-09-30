package snapshot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/gitops"
)

// TestServingAnEntryAdvancesItsLastUseTime pins the last-use signal that
// snapshot-cache retention (manager profile section 10.1, `curator cache
// prune`) reads: the modification time of cache/<source>/<commit>. Get and
// AuthenticateGit each stage privately inside that directory, so every
// publication and every serving advances it.
func TestServingAnEntryAdvancesItsLastUseTime(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "SKILL.md"), []byte("last use"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "one")
	head, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	if _, err := Get(home, "example/skill", repo, head.Commit); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Dir(Dir(home, "example/skill", head.Commit))
	past := time.Now().Add(-30 * 24 * time.Hour)
	for name, serve := range map[string]func() error{
		"Get": func() error {
			_, err := Get(home, "example/skill", repo, head.Commit)
			return err
		},
		"AuthenticateGit": func() error {
			_, err := AuthenticateGit(home, "example/skill", repo, head.Commit)
			return err
		},
	} {
		if err := os.Chtimes(entry, past, past); err != nil {
			t.Fatal(err)
		}
		if err := serve(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		info, err := os.Stat(entry)
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().After(past.Add(24 * time.Hour)) {
			t.Fatalf("%s left the entry's last-use time at %s", name, info.ModTime())
		}
	}
}
