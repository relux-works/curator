package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitEnv returns an env with committer identity for test repositories.
func gitEnv() []string {
	return append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
	)
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// makeRepo builds a repository with one commit and a tag v1, then a second
// commit on main.
func makeRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-q", "-m", "one")
	gitRun(t, dir, "tag", "v1")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "commit", "-qam", "two")
	return dir
}

func TestCloneRefusesSuspiciousURLs(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out")
	if err := Clone("", dest); err == nil {
		t.Fatal("empty URL must be refused")
	}
	if err := Clone("--upload-pack=evil", dest); err == nil {
		t.Fatal("dash-prefixed URL must be refused")
	}
}

func TestCloneRefusesExistingDestination(t *testing.T) {
	dest := t.TempDir()
	if err := Clone("https://example.invalid/repo.git", dest); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloneAndResolve(t *testing.T) {
	src := makeRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := Clone(src, dest); err != nil {
		t.Fatal(err)
	}

	tag, err := Resolve(dest, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	head, err := Resolve(dest, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if tag.Commit == head.Commit {
		t.Fatal("tag v1 must not equal HEAD")
	}
	if len(tag.Commit) != 40 {
		t.Fatalf("commit: %q", tag.Commit)
	}

	// branch resolution prefers origin
	branch, err := Resolve(dest, "branch", "main")
	if err != nil {
		t.Fatal(err)
	}
	if branch.Commit != head.Commit {
		t.Fatalf("branch main = %s, want %s", branch.Commit, head.Commit)
	}

	if _, err := Resolve(dest, "tag", "v404"); err == nil {
		t.Fatal("missing tag must fail")
	}
	if _, err := Resolve(dest, "semver", "1"); err == nil {
		t.Fatal("unknown kind must fail")
	}
	if _, err := Resolve(t.TempDir(), "tag", "v1"); err == nil {
		t.Fatal("non-repo must fail")
	}
}

func TestArchiveExtractsExactTree(t *testing.T) {
	src := makeRepo(t)
	v1, err := Resolve(src, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "snap")
	if err := Archive(src, v1.Commit, dest); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "v1" {
		t.Fatalf("content = %q, want v1 (the tagged tree, not HEAD)", content)
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		t.Fatal("archive must not contain .git")
	}
}

func TestArchiveRejectsLinks(t *testing.T) {
	if _, err := exec.LookPath("ln"); err != nil {
		t.Skip("no ln on this platform")
	}
	src := makeRepo(t)
	if err := os.Symlink("SKILL.md", filepath.Join(src, "link.md")); err != nil {
		t.Skip("symlinks unavailable")
	}
	gitRun(t, src, "add", ".")
	gitRun(t, src, "commit", "-qm", "with link")
	head, _ := Resolve(src, "revision", "HEAD")
	err := Archive(src, head.Commit, filepath.Join(t.TempDir(), "snap"))
	if err == nil || !strings.Contains(err.Error(), "links") {
		t.Fatalf("err = %v, want link rejection", err)
	}
}

func TestFetchAndSubmodules(t *testing.T) {
	src := makeRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := Clone(src, dest); err != nil {
		t.Fatal(err)
	}
	if err := Fetch(dest); err != nil {
		t.Fatal(err)
	}
	if HasSubmodules(dest) {
		t.Fatal("no submodules expected")
	}
	if err := os.WriteFile(filepath.Join(dest, ".gitmodules"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if !HasSubmodules(dest) {
		t.Fatal("submodule marker not detected")
	}
}
