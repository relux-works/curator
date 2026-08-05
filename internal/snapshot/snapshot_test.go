package snapshot

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

func TestConcurrentGetAcceptsOneImmutablePublisher(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "SKILL.md"), []byte("concurrent"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "one")
	head, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	start := make(chan struct{})
	const workers = 16
	results := make([]string, workers)
	errors := make([]error, workers)
	var group sync.WaitGroup
	for index := 0; index < workers; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			results[index], errors[index] = Get(home, "internal/skill-a", repo, head.Commit)
		}(index)
	}
	close(start)
	group.Wait()

	want := Dir(home, "internal/skill-a", head.Commit)
	for index := range results {
		if errors[index] != nil || results[index] != want {
			t.Fatalf("worker %d = %q, %v; want %q", index, results[index], errors[index], want)
		}
	}
	payload, err := os.ReadFile(filepath.Join(want, "SKILL.md"))
	if err != nil || string(payload) != "concurrent" {
		t.Fatalf("published snapshot = %q, %v", payload, err)
	}
}

func TestPublishRejectsSubstitutedDestination(t *testing.T) {
	root := t.TempDir()
	expected := filepath.Join(root, "expected")
	target := filepath.Join(root, "snapshot")
	if err := os.Mkdir(expected, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(expected, "SKILL.md"), []byte("expected"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	wrong := []byte("substituted")
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), wrong, 0o644); err != nil {
		t.Fatal(err)
	}

	err := publishPreparedSnapshot(expected, target)
	if !errors.Is(err, ErrDestinationConflict) {
		t.Fatalf("publishPreparedSnapshot() error = %v; want destination conflict", err)
	}
	payload, readErr := os.ReadFile(filepath.Join(target, "SKILL.md"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(payload) != string(wrong) {
		t.Fatalf("substituted destination was overwritten: %q", payload)
	}
}

func TestPublishRejectsUnsafeDestination(t *testing.T) {
	root := t.TempDir()
	expected := filepath.Join(root, "expected")
	target := filepath.Join(root, "snapshot")
	if err := os.Mkdir(expected, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(expected, "SKILL.md"), []byte("expected"), 0o644); err != nil {
		t.Fatal(err)
	}
	unsafe := []byte("not a directory")
	if err := os.WriteFile(target, unsafe, 0o644); err != nil {
		t.Fatal(err)
	}

	err := publishPreparedSnapshot(expected, target)
	if !errors.Is(err, ErrDestinationConflict) {
		t.Fatalf("publishPreparedSnapshot() error = %v; want destination conflict", err)
	}
	payload, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(payload) != string(unsafe) {
		t.Fatalf("unsafe destination was overwritten: %q", payload)
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
	// A cache hit is authenticated against the immutable commit archive.
	marker := filepath.Join(first, "cache-hit-marker")
	if err := os.WriteFile(marker, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Get(home, "internal/skill-a", repo, head.Commit); !errors.Is(err, ErrDestinationConflict) {
		t.Fatalf("tampered cache hit error = %v, want ErrDestinationConflict", err)
	}
}
