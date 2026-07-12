package snapshot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestGetCachesByCommit(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "one")
	head, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	first, err := Get(home, "internal/skill-a", repo, head.Commit)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(first), "cache/internal/skill-a/"+head.Commit+"/snapshot") {
		t.Fatalf("layout: %s", first)
	}
	// cache hit returns the same directory without touching the repo
	marker := filepath.Join(first, "cache-hit-marker")
	if err := os.WriteFile(marker, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := Get(home, "internal/skill-a", repo, head.Commit)
	if err != nil {
		t.Fatal(err)
	}
	if second != first {
		t.Fatalf("cache miss on second call: %s vs %s", second, first)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("cached snapshot was rebuilt")
	}
}
