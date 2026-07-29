package snapshot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildsource"
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

func repositoryAtCommit(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	for name, content := range files {
		full := filepath.Join(repo, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "snapshot")
	head, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return repo, head.Commit
}

func TestGetCachesValidatedCommitTree(t *testing.T) {
	repo, commit := repositoryAtCommit(t, map[string]string{
		"SKILL.md":         "x",
		"nested/value.txt": "value",
	})
	home := t.TempDir()
	first, err := Get(home, "internal/skill-a", repo, commit)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(first), "cache/internal/skill-a/"+commit+"/snapshot") {
		t.Fatalf("layout: %s", first)
	}
	firstInfo, err := os.Stat(first)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Get(home, "internal/skill-a", repo, commit)
	if err != nil {
		t.Fatal(err)
	}
	secondInfo, err := os.Stat(second)
	if err != nil {
		t.Fatal(err)
	}
	if second != first || !os.SameFile(firstInfo, secondInfo) {
		t.Fatalf("valid cache entry was replaced: %s vs %s", first, second)
	}
	validated, err := buildsource.Validate(second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = validated.Close() }()
}

func TestGetRepairsTamperedOrIncompleteCache(t *testing.T) {
	repo, commit := repositoryAtCommit(t, map[string]string{
		".csk-install.json": "{\"source\":true}\n",
		"SKILL.md":          "original",
		"nested/value.txt":  "value",
	})
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{
			name: "changed file",
			mutate: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("tampered"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "changed root marker",
			mutate: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, ".csk-install.json"), []byte("{\"source\":false}\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "missing file",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "nested", "value.txt")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "extra file",
			mutate: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "extra"), []byte("tampered"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "link",
			mutate: func(t *testing.T, root string) {
				if err := os.Symlink("SKILL.md", filepath.Join(root, "link")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			cached, err := Get(home, "internal/skill-a", repo, commit)
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(t, cached)

			repaired, err := Get(home, "internal/skill-a", repo, commit)
			if err != nil {
				t.Fatal(err)
			}
			payload, err := os.ReadFile(filepath.Join(repaired, "SKILL.md"))
			if err != nil || string(payload) != "original" {
				t.Fatalf("repaired SKILL.md = %q, error = %v", payload, err)
			}
			payload, err = os.ReadFile(filepath.Join(repaired, "nested", "value.txt"))
			if err != nil || string(payload) != "value" {
				t.Fatalf("repaired nested file = %q, error = %v", payload, err)
			}
			if _, err := os.Lstat(filepath.Join(repaired, "extra")); !os.IsNotExist(err) {
				t.Fatalf("extra cache entry survived repair: %v", err)
			}
			if _, err := os.Lstat(filepath.Join(repaired, "link")); !os.IsNotExist(err) {
				t.Fatalf("link cache entry survived repair: %v", err)
			}
			matches, err := filepath.Glob(filepath.Join(filepath.Dir(repaired), ".snapshot-*.tmp"))
			if err != nil || len(matches) != 0 {
				t.Fatalf("staging remnants = %v, error = %v", matches, err)
			}
			backups, err := filepath.Glob(filepath.Join(filepath.Dir(repaired), ".snapshot-replaced-*"))
			if err != nil || len(backups) != 0 {
				t.Fatalf("replacement remnants = %v, error = %v", backups, err)
			}
		})
	}
}

func TestGetValidatesRepositoryBeforeReusingCache(t *testing.T) {
	repo, commit := repositoryAtCommit(t, map[string]string{"SKILL.md": "valid"})
	home := t.TempDir()
	target, err := Get(home, "internal/skill-a", repo, commit)
	if err != nil {
		t.Fatal(err)
	}

	// Make the immutable source invalid at a different commit while placing a
	// valid-looking directory at that commit's cache path. Directory presence
	// must not permit reuse before source validation.
	if err := os.Symlink("SKILL.md", filepath.Join(repo, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	gitRun(t, repo, "add", "link")
	gitRun(t, repo, "commit", "-qm", "invalid link")
	invalidHead, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	invalidTarget := Dir(home, "internal/skill-a", invalidHead.Commit)
	if err := os.MkdirAll(invalidTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(invalidTarget, "SKILL.md"), []byte("looks valid"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GetValidated(home, "internal/skill-a", repo, invalidHead.Commit); err == nil {
		t.Fatal("invalid immutable repository source reused a present cache directory")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("unrelated valid cache entry was disturbed: %v", err)
	}
}

func TestConcurrentGetPublishesOneValidSnapshot(t *testing.T) {
	repo, commit := repositoryAtCommit(t, map[string]string{
		"SKILL.md": "content",
		"a/b/c":    "nested",
	})
	home := t.TempDir()
	const workers = 6
	paths := make([]string, workers)
	errs := make([]error, workers)
	var group sync.WaitGroup
	for index := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			paths[index], errs[index] = Get(home, "internal/skill-a", repo, commit)
		}()
	}
	group.Wait()
	for index, err := range errs {
		if err != nil {
			t.Fatalf("worker %d: %v", index, err)
		}
		if paths[index] != paths[0] {
			t.Fatalf("worker paths differ: %q vs %q", paths[index], paths[0])
		}
	}
	token, err := buildsource.Validate(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = token.Close() }()
}

func TestConcurrentGetRepairsPresentTamperedSnapshotAtomically(t *testing.T) {
	repo, commit := repositoryAtCommit(t, map[string]string{
		".csk-install.json": "{\"source\":true}\n",
		"SKILL.md":          "original",
		"a/b/c":             "nested",
	})
	home := t.TempDir()
	target, err := Get(home, "internal/skill-a", repo, commit)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := buildsource.Validate(target)
	if err != nil {
		t.Fatal(err)
	}
	expectedIdentity := expected.Identity()
	if err := expected.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}

	const workers = 24
	start := make(chan struct{})
	stopObserver := make(chan struct{})
	observerDone := make(chan struct{})
	var missingTarget atomic.Int64
	go func() {
		defer close(observerDone)
		for {
			select {
			case <-stopObserver:
				return
			default:
				if _, err := os.Lstat(target); os.IsNotExist(err) {
					missingTarget.Add(1)
				}
			}
		}
	}()

	errs := make([]error, workers)
	var group sync.WaitGroup
	for index := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if index%2 == 0 {
				path, err := Get(home, "internal/skill-a", repo, commit)
				if err != nil {
					errs[index] = err
					return
				}
				token, err := buildsource.Validate(path)
				if err != nil {
					errs[index] = err
					return
				}
				defer func() { _ = token.Close() }()
				if token.Identity() != expectedIdentity {
					errs[index] = fmt.Errorf("repaired identity = %#v, want %#v", token.Identity(), expectedIdentity)
				}
				return
			}
			token, err := GetValidated(home, "internal/skill-a", repo, commit)
			if err != nil {
				errs[index] = err
				return
			}
			defer func() { _ = token.Close() }()
			if token.Identity() != expectedIdentity {
				errs[index] = fmt.Errorf("repaired identity = %#v, want %#v", token.Identity(), expectedIdentity)
				return
			}
			errs[index] = token.Recheck()
		}()
	}
	close(start)
	group.Wait()
	close(stopObserver)
	<-observerDone

	for index, err := range errs {
		if err != nil {
			t.Errorf("worker %d: %v", index, err)
		}
	}
	if observed := missingTarget.Load(); observed != 0 {
		t.Errorf("live snapshot path was missing %d times during repair", observed)
	}
	for _, pattern := range []string{".snapshot-*.tmp", ".snapshot-replaced-*"} {
		matches, err := filepath.Glob(filepath.Join(filepath.Dir(target), pattern))
		if err != nil || len(matches) != 0 {
			t.Errorf("publication remnants for %q = %v, error = %v", pattern, matches, err)
		}
	}
}

func TestGetRepairsTamperedSelectionAndGenerationTypes(t *testing.T) {
	repo, commit := repositoryAtCommit(t, map[string]string{"SKILL.md": "original"})
	home := t.TempDir()
	legacy, err := Get(home, "internal/skill-a", repo, commit)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "SKILL.md"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	pointer := Dir(home, "internal/skill-a", commit) + ".current"
	if err := os.Mkdir(pointer, 0o755); err != nil {
		t.Fatal(err)
	}

	firstRepair, err := Get(home, "internal/skill-a", repo, commit)
	if err != nil {
		t.Fatal(err)
	}
	if firstRepair == legacy {
		t.Fatal("tampered legacy snapshot was reused")
	}
	if info, err := os.Lstat(pointer); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("repaired pointer is not a regular file: info=%v error=%v", info, err)
	}
	if err := os.RemoveAll(firstRepair); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(firstRepair, []byte("wrong type"), 0o644); err != nil {
		t.Fatal(err)
	}

	secondRepair, err := Get(home, "internal/skill-a", repo, commit)
	if err != nil {
		t.Fatal(err)
	}
	if secondRepair == firstRepair {
		t.Fatal("generation with changed root type was reused")
	}
	validated, err := buildsource.Validate(secondRepair)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = validated.Close() }()
	if payload, err := os.ReadFile(filepath.Join(secondRepair, "SKILL.md")); err != nil || string(payload) != "original" {
		t.Fatalf("repaired content = %q, error = %v", payload, err)
	}
	if _, err := os.Lstat(legacy); err != nil {
		t.Fatalf("legacy path disappeared during generation repair: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(legacy), "snapshot-generation-*"))
	if err != nil || len(matches) > 2 {
		t.Fatalf("retained generations = %v, error = %v", matches, err)
	}
}

func TestSnapshotLockSerializesCallers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.lock")
	first, err := acquireSnapshotLock(path)
	if err != nil {
		t.Fatal(err)
	}
	acquired := make(chan *snapshotLock, 1)
	attempting := make(chan struct{})
	errs := make(chan error, 1)
	go func() {
		close(attempting)
		second, err := acquireSnapshotLock(path)
		if err != nil {
			errs <- err
			return
		}
		acquired <- second
	}()
	<-attempting
	select {
	case second := <-acquired:
		_ = second.Close()
		t.Fatal("second caller acquired the snapshot lock before release")
	case err := <-errs:
		t.Fatal(err)
	case <-time.After(50 * time.Millisecond):
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case second := <-acquired:
		if err := second.Close(); err != nil {
			t.Fatal(err)
		}
	case err := <-errs:
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("second caller did not acquire the released snapshot lock")
	}
}
