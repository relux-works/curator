package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/install"
)

// Drive main -> cli.install -> install.Project -> gitignore.Ensure with the
// compiled production binary, including real process exits and published bytes.
func TestInstallGitignoreEntryPoint(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "curator"+productionBinarySuffix())
	build := exec.Command("go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build curator: %v\n%s", err, out)
	}
	rows := []struct {
		name      string
		git       bool
		parentGit bool
		fix       bool
		ignore    string
		broken    bool
		gitExit   int
		installed bool
		code      int
	}{
		{name: "non-git", installed: true},
		{name: "non-git-fix", fix: true, installed: true},
		{name: "git-parent-ignored", parentGit: true, ignore: ".agents/\n.codex/skills/\n", installed: true},
		{name: "git-ignored", git: true, ignore: ".agents/\n.codex/skills/\n", installed: true},
		{name: "git-missing-ignore", git: true},
		{name: "git-fix-ignore", git: true, fix: true, installed: true},
		{name: "git-ignore-negation", git: true, fix: true, ignore: ".agents/\n!.agents/\n.codex/skills/\n"},
		{name: "broken-git", broken: true, fix: true, code: exitFail},
		{name: "unexpected-exit-2", gitExit: 2, fix: true, code: exitFail},
		{name: "unexpected-exit-128", gitExit: 128, fix: true, code: exitFail},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			if row.gitExit != 0 && runtime.GOOS == "windows" {
				t.Skip("PATH git shell shim requires Unix; corrupt .git row runs on all platforms")
			}
			root := t.TempDir()
			userHome := filepath.Join(root, "user")
			project := filepath.Join(root, "product")
			configPath := filepath.Join(root, "manager", "config.json")
			for _, dir := range []string{userHome, project} {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			environment := append(os.Environ(), "HOME="+userHome, "USERPROFILE="+userHome, "CURATOR_CONFIG="+configPath)
			invoke := func(args ...string) (int, string) {
				t.Helper()
				cmd := exec.Command(binary, args...)
				cmd.Env = environment
				cmd.Dir = project
				out, err := cmd.CombinedOutput()
				code := 0
				if err != nil {
					var exitErr *exec.ExitError
					if !errors.As(err, &exitErr) {
						t.Fatalf("launch curator: %v", err)
					}
					code = exitErr.ExitCode()
				}
				t.Logf("curator %v: exit %d\n%s", args, code, out)
				return code, string(out)
			}
			for _, args := range [][]string{
				{"bootstrap", "--non-interactive", "--skills-root", filepath.Join(root, "skills-root")},
				{"project", "add", "app", project, "--agents", "codex_cli"},
			} {
				if code, out := invoke(args...); code != 0 {
					t.Fatalf("setup exit %d: %s", code, out)
				}
			}
			source := filepath.Join(project, "repos", "review")
			writeCLISkill(t, source, "review")
			// The product folder holds a real nested repository; it is never
			// initialized as a repository in either positive non-git row.
			runGit(t, source, "init", "-q")
			manifest := `{"schema_version":2,"agents":["codex_cli"],"sources":{"s":{"path":"./repos"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`
			if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(manifest), 0o644); err != nil {
				t.Fatal(err)
			}
			if code, out := invoke("project", "resolve", "app"); code != 0 {
				t.Fatalf("resolve exit %d: %s", code, out)
			}
			for _, path := range []string{filepath.Join(project, "Skillfile.lock.json"), install.DraftBindingsPath(filepath.Dir(configPath), project)} {
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("lock/bindings missing: %v", err)
				}
			}
			if err := os.Remove(filepath.Join(project, ".gitignore")); err != nil {
				t.Fatal(err)
			}
			if row.parentGit {
				runGit(t, root, "init", "-q")
			}
			if row.git {
				runGit(t, project, "init", "-q")
			}
			if row.ignore != "" {
				if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(row.ignore), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if row.broken {
				if err := os.WriteFile(filepath.Join(project, ".git"), []byte("gitdir: missing\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if row.gitExit != 0 {
				shim := filepath.Join(root, "shim")
				if err := os.Mkdir(shim, 0o755); err != nil {
					t.Fatal(err)
				}
				// Exit 2 even carries the absence diagnostic: both code and
				// message must agree before the non-repository fallback applies.
				message := "fatal: unexpected git failure"
				if row.gitExit == 2 {
					message = "fatal: not a git repository (or any of the parent directories): .git"
				}
				script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' '%s' >&2\nexit %d\n", message, row.gitExit)
				if err := os.WriteFile(filepath.Join(shim, "git"), []byte(script), 0o755); err != nil {
					t.Fatal(err)
				}
				environment = append(environment, "PATH="+shim+string(os.PathListSeparator)+os.Getenv("PATH"))
			}
			beforeSource := digestTree(t, source)
			args := []string{"install", "app"}
			if row.fix {
				args = append(args, "--fix-gitignore")
			}
			code, out := invoke(args...)
			if code != row.code {
				t.Fatalf("install exit %d, want %d:\n%s", code, row.code, out)
			}
			installed := filepath.Join(project, ".agents", "skills", "review")
			if row.installed {
				for _, name := range []string{"SKILL.md", filepath.Join("references", "info.md")} {
					want, err := os.ReadFile(filepath.Join(source, name))
					if err != nil {
						t.Fatal(err)
					}
					got, err := os.ReadFile(filepath.Join(installed, name))
					if err != nil || !bytes.Equal(got, want) {
						t.Fatalf("installed %s bytes missing/differ: %v\n%s", name, err, out)
					}
				}
			} else if _, err := os.Stat(installed); !os.IsNotExist(err) {
				t.Fatalf("refused/skipped row published skill: %v", err)
			}
			nonGit := !row.git && !row.parentGit && !row.broken && row.gitExit == 0
			if strings.Contains(out, "gitignore hygiene check does not apply") != nonGit {
				t.Fatalf("incorrect hygiene notice:\n%s", out)
			}
			if nonGit {
				for _, name := range []string{".git", ".gitignore"} {
					if _, err := os.Lstat(filepath.Join(project, name)); !os.IsNotExist(err) {
						t.Fatalf("non-git install wrote %s: %v", name, err)
					}
				}
			}
			if !row.installed && row.code == 0 && !strings.Contains(out, "skipped") {
				t.Fatalf("git policy behavior changed:\n%s", out)
			}
			if row.code != 0 && !strings.Contains(out, "git check-ignore failed") {
				t.Fatalf("git failure diagnostic missing:\n%s", out)
			}
			if row.ignore != "" {
				got, err := os.ReadFile(filepath.Join(project, ".gitignore"))
				if err != nil || string(got) != row.ignore {
					t.Fatalf("existing ignore file changed: %v\n%s", err, got)
				}
			}
			if !row.installed && row.ignore == "" {
				if _, err := os.Lstat(filepath.Join(project, ".gitignore")); !os.IsNotExist(err) {
					t.Fatalf("refused/skipped row wrote .gitignore: %v", err)
				}
			}
			entries, err := os.ReadDir(project)
			if err != nil {
				t.Fatal(err)
			}
			allowed := map[string]bool{"Skillfile.json": true, "Skillfile.lock.json": true, "repos": true, ".agents": true, ".codex": true, ".git": true, ".gitignore": true}
			for _, entry := range entries {
				if !allowed[entry.Name()] {
					t.Fatalf("install wrote undeclared project surface: %s", entry.Name())
				}
			}
			assertTreesEqual(t, beforeSource, digestTree(t, source))
		})
	}
}
