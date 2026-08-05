package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestNativeBlackBoxProjectGlobalLifecycle exercises only the built command's
// public surface. It is the native project/global activation and uninstall
// spine used by rc.5 qualification; external-repository source, cache, failure,
// and rollback cases are bound separately by internal/rc5interop.
func TestNativeBlackBoxProjectGlobalLifecycle(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "bin", "curator")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if err := os.MkdirAll(filepath.Dir(binary), 0o755); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build black-box curator: %v\n%s", err, output)
	}

	home := filepath.Join(root, "manager-home")
	configPath := filepath.Join(home, "config.json")
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")
	fixture := filepath.Join(skillsRoot, "native-fixture")
	writeNativeFixture(t, fixture)
	runNativeGit(t, project, "init", "-q", "-b", "main")

	env := append(os.Environ(), "CURATOR_CONFIG="+configPath, "HOME="+root, "USERPROFILE="+root)
	runNativeCurator(t, binary, env, "bootstrap", "--non-interactive", "--skills-root", skillsRoot, "--default-agents", "codex_cli")
	runNativeCurator(t, binary, env, "init", project)
	runNativeCurator(t, binary, env, "add", "native-fixture", "--source", "native-fixture", "--tag", "v1", "--project", project)
	runNativeCurator(t, binary, env, "status", project, "--json", "--check")
	runNativeCurator(t, binary, env, "shell-init", "bash", "--install", "--no-global")
	if _, err := os.Stat(filepath.Join(project, ".agents", "skills", "native-fixture", "SKILL.md")); err != nil {
		t.Fatalf("project activation missing: %v", err)
	}

	runNativeCurator(t, binary, env, "global", "init")
	runNativeCurator(t, binary, env, "global", "add", "native-fixture", "--source", "native-fixture", "--tag", "v1")
	runNativeCurator(t, binary, env, "global", "status", "--json", "--check")
	if _, err := os.Stat(filepath.Join(home, "global", "skills", "native-fixture", "SKILL.md")); err != nil {
		t.Fatalf("global activation missing: %v", err)
	}

	runNativeCurator(t, binary, env, "remove", "native-fixture", "--project", project)
	runNativeCurator(t, binary, env, "install", project)
	if _, err := os.Stat(filepath.Join(project, ".agents", "skills", "native-fixture")); !os.IsNotExist(err) {
		t.Fatalf("project uninstall left installed state: %v", err)
	}
	runNativeCurator(t, binary, env, "global", "remove", "native-fixture")
	runNativeCurator(t, binary, env, "global", "install")
	if _, err := os.Stat(filepath.Join(home, "global", "skills", "native-fixture")); !os.IsNotExist(err) {
		t.Fatalf("global uninstall left installed state: %v", err)
	}
}

func writeNativeFixture(t *testing.T, root string) {
	t.Helper()
	write := func(relative, content string, mode os.FileMode) {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	write("SKILL.md", "---\nname: native-fixture\ndescription: native qualification fixture\n---\n# Native fixture\n", 0o644)
	write("scripts/native-fixture", "#!/bin/sh\necho native-fixture\n", 0o755)
	manifest, err := json.Marshal(map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{},
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{
			"native-fixture": map[string]any{"type": "script", "unix_path": "scripts/native-fixture", "win_path": "scripts/native-fixture"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	write("agent-skill.json", string(manifest), 0o644)
	runNativeGit(t, root, "init", "-q", "-b", "main")
	runNativeGit(t, root, "add", ".")
	runNativeGit(t, root, "commit", "-qm", "native fixture")
	runNativeGit(t, root, "tag", "v1")
}

func runNativeGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	command := exec.Command("git", gitArgs...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Curator Test", "GIT_AUTHOR_EMAIL=curator@example.test",
		"GIT_COMMITTER_NAME=Curator Test", "GIT_COMMITTER_EMAIL=curator@example.test",
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func runNativeCurator(t *testing.T, binary string, env []string, args ...string) {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Env = env
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("curator %v: %v\n%s", args, err, output)
	}
}
