package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
)

func repinCommit(t *testing.T, repo string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// Revision, tag and range repins move an active Git root and its rendered
// context. Plain update and same-source install keep the original exact pin.
func TestProfileUpdateRepin(t *testing.T) {
	for _, form := range []string{"revision", "tag", "range"} {
		t.Run(form, func(t *testing.T) {
			requireGit(t)
			source, home := profileHome(t)
			repo := t.TempDir()
			writeContextPackage(t, repo, "repin", "1.0.0", "revision A\n")
			runGitRepo(t, repo, "init")
			commitGitRepo(t, repo, "v1.0.0")
			a := repinCommit(t, repo)
			const operand = "https://example.com/repin"
			serveGitRepos(t, map[string]string{operand: repo})
			if code, _, stderr := runProfile(t, source, "profile", "install", operand, "--revision", a); code != exitOK {
				t.Fatalf("install: %d %s", code, stderr)
			}
			before, err := envprofile.List(home)
			if err != nil {
				t.Fatal(err)
			}
			var oldHash string
			for _, info := range before {
				if info.Name == "repin" {
					oldHash = info.LockHash
				}
			}
			if oldHash == "" {
				t.Fatal("installed root absent")
			}
			writeContextPackage(t, repo, "repin", "1.1.0", "revision B\n")
			commitGitRepo(t, repo, "v1.1.0")
			b := repinCommit(t, repo)
			if code, stdout, stderr := runProfile(t, source, "profile", "update", "repin"); code != exitOK || !strings.Contains(stdout, "unchanged") {
				t.Fatalf("plain update: %d %s %s", code, stdout, stderr)
			}
			if code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--revision", a); code != exitOK || !strings.Contains(stdout, "updated profile repin") {
				t.Fatalf("reinstall exact source: %d %s %s", code, stdout, stderr)
			}
			for _, info := range mustRepinList(t, home) {
				if info.Name == "repin" && (info.LockHash != oldHash || info.Source.Req.Revision != a) {
					t.Fatal("plain update/reinstall moved the exact pin")
				}
			}
			ref := map[string]string{"revision": b, "tag": "v1.1.0", "range": "^1.1.0"}[form]
			code, stdout, stderr := runProfile(t, source, "profile", "update", "repin", "--"+form, ref)
			if code != exitOK || !strings.Contains(stdout, "repin: updated (lock ") {
				t.Fatalf("repin: %d %s %s", code, stdout, stderr)
			}
			profiles, err := envprofile.List(home)
			if err != nil {
				t.Fatal(err)
			}
			var lockHash string
			for _, info := range profiles {
				if info.Name == "repin" {
					root, ok := info.Lock.RootMember()
					if !ok || root.Commit != b || root.Version != "1.1.0" || info.LockHash == oldHash {
						t.Fatalf("repinned root: %+v", info)
					}
					lockHash = info.LockHash
				}
			}
			if lockHash == "" {
				t.Fatal("repinned profile missing")
			}
			code, stdout, stderr = runProfile(t, source, "profile", "list")
			if code != exitOK || !strings.Contains(stdout, "1.1.0") || !strings.Contains(stdout, lockHash) || !strings.Contains(stdout, ref) {
				t.Fatalf("list: %d %s %s", code, stdout, stderr)
			}
			rendered, err := os.ReadFile(filepath.Join(os.Getenv("CODEX_HOME"), "AGENTS.md"))
			if err != nil || !strings.Contains(string(rendered), "revision B") || strings.Contains(string(rendered), "revision A") {
				t.Fatalf("scope not rematerialized: %s %v", rendered, err)
			}
		})
	}
}

func mustRepinList(t *testing.T, home string) []envprofile.Info {
	t.Helper()
	profiles, err := envprofile.List(home)
	if err != nil {
		t.Fatal(err)
	}
	return profiles
}

// Requirement flags need one named Git profile and exactly one nonempty form.
// Refused grammar and path roots leave their source, lock and rendered homes.
func TestProfileUpdateRepinUsage(t *testing.T) {
	source, home := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "path-profile", "1.0.0", "path context\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install: %d %s", code, stderr)
	}
	cases := [][]string{
		{"--all", "--revision", "new"}, {"--all", "--tag", "v2.0.0"}, {"--all", "--range", "^2"},
		{"path-profile", "--revision", "new", "--tag", "v2.0.0"},
		{"path-profile", "--tag", "v2.0.0", "--range", "^2"},
		{"path-profile", "--revision", "new", "--range", "^2"},
		{"path-profile", "--revision", "new"}, {"path-profile", "--tag", "v2.0.0"}, {"path-profile", "--range", "^2"},
		{"--revision", "new"}, {"path-profile", "--revision="}, {"path-profile", "--revision=", "--tag", "v2.0.0"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			paths := []string{filepath.Join(envprofile.ProfileDir(home, "path-profile"), "source.json"), filepath.Join(envprofile.ProfileDir(home, "path-profile"), "lock.json"), filepath.Join(os.Getenv("CODEX_HOME"), "AGENTS.md")}
			before := make([]string, len(paths))
			for i, path := range paths {
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				before[i] = string(raw)
			}
			code, _, stderr := runProfile(t, source, append([]string{"profile", "update"}, args...)...)
			if code != exitUsage || !strings.Contains(stderr, envprofile.DiagRefConflict) {
				t.Fatalf("usage: %d %s", code, stderr)
			}
			for i, path := range paths {
				raw, err := os.ReadFile(path)
				if err != nil || string(raw) != before[i] {
					t.Fatalf("usage refusal changed %s: %v", filepath.Base(path), err)
				}
			}
		})
	}
}

// Dedicated help advertises repin forms and their named-Git-only boundary.
func TestProfileUpdateRepinHelp(t *testing.T) {
	source, _ := profileHome(t)
	code, stdout, stderr := runProfile(t, source, "profile", "update", "--help")
	if code != exitOK {
		t.Fatalf("help: %d %s", code, stderr)
	}
	for _, text := range []string{"--revision", "--tag", "--range", "named Git root", "directory", "--all"} {
		if !strings.Contains(stdout, text) {
			t.Fatalf("help missing %q: %s", text, stdout)
		}
	}
}
