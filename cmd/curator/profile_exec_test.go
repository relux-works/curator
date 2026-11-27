package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/pathboundary"
)

func initLocalSkillGitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(rel, ".sh") {
			mode = 0o755
		}
		if err := os.WriteFile(full, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	runGit := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	runGit("init", "-b", "main")
	runGit("add", ".")
	runGit("commit", "-m", "init")
	runGit("tag", "v1.0.0")
	return dir
}

func TestProfileExecWithAndWithoutEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell script fixture")
	}
	source, home := profileHome(t)
	writeNativeCredentials(t)

	skillDir := initLocalSkillGitRepo(t, map[string]string{
		"SKILL.md":         "---\nname: demo\ndescription: demo skill\n---\n# Demo\n",
		"scripts/demo.sh":  "#!/bin/sh\nprintf 'TOOL_OK CODEX_HOME=%s CLAUDE_CONFIG_DIR=%s\\n' \"${CODEX_HOME:-}\" \"${CLAUDE_CONFIG_DIR:-}\"\n",
		"agent-skill.json": `{"schema_version":10,"runtime_roots":["scripts"],"capabilities":{},"commands":{"demo-tool":{"type":"script","unix_path":"scripts/demo.sh"}}}`,
	})
	const skillURL = "https://example.com/skills/demo"
	serveGitRepos(t, map[string]string{
		skillURL: skillDir,
	})

	pkgDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pkgDir, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	ctxManifest := `{"schema_version":2,"name":"demo-root","version":"1.0.0","context":{"modules":[{"path":"a.md"}]},"requires":{"skills":{"demo":{"git":"` + skillURL + `","tag":"v1.0.0"}}}}`
	if err := os.WriteFile(filepath.Join(pkgDir, "agent-context.json"), []byte(ctxManifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "context", "a.md"), []byte("# demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(pkgDir); err != nil {
		t.Fatalf("protect pkgDir: %v", err)
	}

	code, stdout, stderr := runProfile(t, source, "profile", "install", pkgDir, "--as", "demo-prof")
	if code != exitOK {
		t.Fatalf("profile install failed (%d): stdout=%s stderr=%s", code, stdout, stderr)
	}

	// 1. profile exec without --env prepends command root and runs exported command.
	code, stdout, stderr = runProfile(t, source, "profile", "exec", "demo-prof", "--", "demo-tool")
	if code != exitOK {
		t.Fatalf("profile exec failed (%d): stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.HasPrefix(stdout, "TOOL_OK ") {
		t.Fatalf("profile exec stdout = %q, want TOOL_OK prefix", stdout)
	}

	// 2. profile exec --env codex applies managed CODEX_HOME and prepends command root.
	code, stdout, stderr = runProfile(t, source, "profile", "exec", "demo-prof", "--env", "codex", "--", "sh", "-c", "demo-tool")
	if code != exitOK {
		t.Fatalf("profile exec --env codex failed (%d): stdout=%s stderr=%s", code, stdout, stderr)
	}
	wantCodexHome := filepath.Join(home, "environments", "demo-prof", "codex_cli")
	if !strings.Contains(stdout, "CODEX_HOME="+wantCodexHome) {
		t.Fatalf("profile exec --env codex stdout = %q, want CODEX_HOME=%s", stdout, wantCodexHome)
	}

	// 3. Flag before profile name also works (--env claude demo-prof -- ...).
	code, stdout, stderr = runProfile(t, source, "profile", "exec", "--env", "claude", "demo-prof", "--", "sh", "-c", "demo-tool")
	if code != exitOK {
		t.Fatalf("profile exec --env claude failed (%d): stdout=%s stderr=%s", code, stdout, stderr)
	}
	wantClaudeDir := filepath.Join(home, "environments", "demo-prof", "claude_code")
	if !strings.Contains(stdout, "CLAUDE_CONFIG_DIR="+wantClaudeDir) {
		t.Fatalf("profile exec --env claude stdout = %q, want CLAUDE_CONFIG_DIR=%s", stdout, wantClaudeDir)
	}

	// 4. env resolve --format shell emits both managed env var and v9 PATH prepend.
	code, stdout, stderr = runProfile(t, source, "env", "resolve", "codex", "--profile", "demo-prof", "--repair", "--format", "shell")
	if code != exitOK {
		t.Fatalf("env resolve --format shell failed (%d): stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "export CODEX_HOME='"+wantCodexHome+"'\n") {
		t.Fatalf("shell output missing CODEX_HOME: %q", stdout)
	}
	if !strings.Contains(stdout, `export PATH='`+filepath.Join(home, "environments", "demo-prof", "commands")+`/`) || !strings.Contains(stdout, `':"$PATH"`+"\n") {
		t.Fatalf("shell output missing v9 PATH export: %q", stdout)
	}

	// 5. Usage errors exit 2.
	for _, bad := range [][]string{
		{"profile", "exec", "demo-prof"},
		{"profile", "exec", "demo-prof", "--"},
		{"profile", "exec", "--bogus", "demo-prof", "--", "true"},
	} {
		code, _, _ = runProfile(t, source, bad...)
		if code != exitUsage {
			t.Fatalf("run(%v) = %d, want exitUsage (%d)", bad, code, exitUsage)
		}
	}
}
