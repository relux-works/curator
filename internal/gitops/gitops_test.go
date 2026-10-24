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

func TestExtractProducesExactTree(t *testing.T) {
	src := makeRepo(t)
	v1, err := Resolve(src, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "snap")
	if err := Extract(src, v1.Commit, dest); err != nil {
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

// TestArchiveRejectsLinks proves extraction refuses a tree containing a
// symlink entry (mode 120000). The entry is written through mktree, so the
// case needs neither `ln` nor host symlink support and runs on every GOOS.
func TestArchiveRejectsLinks(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "x")
	commit := commitRawTree(t, repo, []string{
		"100644 blob " + oid + "\tSKILL.md",
		"120000 blob " + oid + "\tlink.md",
	})
	dest := filepath.Join(t.TempDir(), "snap")
	err := Extract(repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "links") {
		t.Fatalf("err = %v, want link rejection", err)
	}
	if _, statErr := os.Stat(filepath.Join(dest, "link.md")); statErr == nil {
		t.Fatal("refused extraction must not leave link.md behind")
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

// Clone enables long-path support before copying a template beyond MAX_PATH
// and persists it for a real packed transfer, independently of ambient defaults.
// The packed fixture stays below Git's independent absolute metadata-path limit;
// this test does not claim arbitrary-length repository roots are supported.
func TestCloneLongDestination(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(configPath, []byte("[core]\nlongpaths = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", configPath)
	src := makeRepo(t)
	for _, tc := range []struct {
		name   string
		length int
		packed bool
	}{
		{"initialization", 238, false},
		{"packed-transfer", 180, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if len(root)+1 >= tc.length {
				t.Fatalf("temporary root leaves no room for long-path witness: %d", len(root))
			}
			dest := filepath.Join(root, strings.Repeat("c", tc.length-len(root)-1))
			template := t.TempDir()
			name := strings.Repeat("template", 5)
			if err := os.WriteFile(filepath.Join(template, name), []byte("template"), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("GIT_TEMPLATE_DIR", template)
			packPath := filepath.Join(dest, ".git", "objects", "pack", "pack-"+strings.Repeat("0", 40)+".keep")
			if len(dest) >= 248 || !tc.packed && len(filepath.Join(dest, ".git", name)) <= 260 || tc.packed && len(packPath) >= 250 {
				t.Fatalf("invalid long-path witness: directory=%d pack=%d", len(dest), len(packPath))
			}
			if err := Clone("--upload-pack=forbidden", dest); err == nil || !strings.Contains(err.Error(), "suspicious git URL") {
				t.Fatalf("invalid source accepted: %v", err)
			}
			if _, err := os.Lstat(dest); !os.IsNotExist(err) {
				t.Fatalf("refused clone created destination: %v", err)
			}
			remote := src
			if tc.packed {
				// file:// exercises index-pack instead of the local hard-link path.
				remote = "file://" + filepath.ToSlash(src)
				if filepath.VolumeName(src) != "" {
					remote = "file:///" + filepath.ToSlash(src)
				}
			}
			if err := Clone(remote, dest); err != nil {
				t.Fatal(err)
			}
			// Inspect stdout separately: Git can warn on stderr while probing
			// an absent alternates file before loading this local setting.
			if got, err := run(dest, "config", "--local", "--get", "core.longpaths"); err != nil || got != "true" {
				t.Fatalf("cache long-path support = %q, want true: %v", got, err)
			}
			if tc.packed {
				packs, err := filepath.Glob(filepath.Join(dest, ".git", "objects", "pack", "pack-*.pack"))
				if err != nil || len(packs) == 0 {
					t.Fatalf("file transport did not install a pack: %v %v", packs, err)
				}
			}
			if body, err := os.ReadFile(filepath.Join(dest, ".git", name)); err != nil || string(body) != "template" {
				t.Fatalf("long-path template = %q: %v", body, err)
			}
			if err := Fetch(dest); err != nil {
				t.Fatal(err)
			}
			ref, err := Resolve(dest, "tag", "v1")
			if err != nil {
				t.Fatal(err)
			}
			snapshot := filepath.Join(t.TempDir(), "snapshot")
			if err := Extract(dest, ref.Commit, snapshot); err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(filepath.Join(snapshot, "SKILL.md"))
			if err != nil || string(body) != "v1" {
				t.Fatalf("long-path payload = %q: %v", body, err)
			}
		})
	}
}
