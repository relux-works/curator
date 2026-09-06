package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envmarker"
)

// These tests drive the production run() entry point for the profile rows.

func profileHome(t *testing.T) (stubConfigSource, string) {
	t.Helper()
	home := t.TempDir()
	source := stubConfigSource{
		path: filepath.Join(home, "config.json"),
		cfg:  &config.Config{Path: filepath.Join(home, "config.json")},
	}
	base := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(base, "claude"))
	t.Setenv("CODEX_HOME", filepath.Join(base, "codex"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "xdg"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(base, "pi"))
	return source, home
}

func writeContextPackage(t *testing.T, root, name, version, module string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version": 1, "name": "` + name + `", "version": "` + version + `",` +
		`"context": {"modules": [{"path": "a.md"}]}}`
	if err := os.WriteFile(filepath.Join(root, "agent-context.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "context", "a.md"), []byte(module), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runProfile(t *testing.T, source stubConfigSource, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	code := run(args, source, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestProfileInstallListUse drives install, list, and use through run():
// install activates the first profile, list reports it, and use writes the
// linked homes with markers.
func TestProfileInstallListUse(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, stdout, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	} else if !strings.Contains(stdout, "installed and activated profile acme") {
		t.Fatalf("stdout:\n%s", stdout)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "list")
	if code != exitOK {
		t.Fatalf("list = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "acme\tacme\tpath") || !strings.Contains(stdout, "current") {
		t.Fatalf("list stdout:\n%s", stdout)
	}
	code, stdout, stderr = runProfile(t, source, "profile", "use", "acme")
	if code != exitOK {
		t.Fatalf("use = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "claude_code: switched") {
		t.Fatalf("use stdout:\n%s", stdout)
	}
	claude := filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md")
	payload, err := os.ReadFile(claude) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "## Context: acme 1.0.0") {
		t.Fatalf("document:\n%s", payload)
	}
	marker, err := envmarker.Read(os.Getenv("CLAUDE_CONFIG_DIR"))
	if err != nil || marker == nil || marker.Profile.Name != "acme" {
		t.Fatalf("marker %+v %v", marker, err)
	}
}

// TestProfileUseUnknownIsAFailure narrows the lookup gate: switching to a
// profile that is not installed fails with exit 1, never success.
func TestProfileUseUnknownIsAFailure(t *testing.T) {
	source, _ := profileHome(t)
	code, _, stderr := runProfile(t, source, "profile", "use", "ghost")
	if code != exitFail {
		t.Fatalf("use ghost = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
	}
	if !strings.Contains(stderr, "profile_not_found") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

// TestProfileRemoveInUseIsAFailure narrows the removal gate through the
// CLI: removing the current profile fails with profile_in_use.
func TestProfileRemoveInUseIsAFailure(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, _, stderr := runProfile(t, source, "profile", "remove", "acme")
	if code != exitFail || !strings.Contains(stderr, "profile_in_use") {
		t.Fatalf("remove = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileInstallRefConflictIsAFailure checks a path operand with a
// requirement flag fails instead of silently ignoring the flag.
func TestProfileInstallRefConflictIsAFailure(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	code, _, stderr := runProfile(t, source, "profile", "install", pkg, "--range", "^1.0.0")
	if code == exitOK || !strings.Contains(stderr, "profile_install_ref_conflict") {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileSyncCoversTheMachineScope checks sync re-materializes the
// current profile across the adapters.
func TestProfileSyncCoversTheMachineScope(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "sync")
	if code != exitOK {
		t.Fatalf("sync = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "codex_cli: synced") {
		t.Fatalf("sync stdout:\n%s", stdout)
	}
}

// TestProfileComposeIsRefused checks compose names its bound instead of
// pretending to edit machine configuration.
func TestProfileComposeIsRefused(t *testing.T) {
	source, _ := profileHome(t)
	code, _, stderr := runProfile(t, source, "profile", "compose", "acme", "list")
	if code != exitFail || !strings.Contains(stderr, "schema 2") {
		t.Fatalf("compose = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileUseUnknownEnvironmentIsAFailure narrows the registry gate:
// an explicit operand naming an unregistered environment fails with
// environment_unknown.
func TestProfileUseUnknownEnvironmentIsAFailure(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, _, stderr := runProfile(t, source, "profile", "use", "acme", "--env", "ghost")
	if code != exitFail || !strings.Contains(stderr, "environment_unknown") {
		t.Fatalf("use = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileUseTargetIsStageB checks scoped target switches name their
// bound instead of touching fixed homes.
func TestProfileUseTargetIsStageB(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, _, stderr := runProfile(t, source, "profile", "use", "acme", "--target", "xcode")
	if code != exitFail || !strings.Contains(stderr, "environment_target_unknown") {
		t.Fatalf("use = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileInstallWarnsOnSystemModule checks the always-warn class
// reaches stderr on install with exit 0.
func TestProfileInstallWarnsOnSystemModule(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pkg, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version": 1, "name": "sys", "version": "1.0.0",` +
		`"context": {"modules": [{"path": "s.md", "class": "system"}]}}`
	if err := os.WriteFile(filepath.Join(pkg, "agent-context.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg, "context", "s.md"), []byte("system\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runProfile(t, source, "profile", "install", pkg)
	if code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "context-system-module-present") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

// TestProfileScopedUseAndClear drives a narrowed switch and its clearing
// through run(): the scope marker appears in list output and vanishes after
// --clear.
func TestProfileScopedUseAndClear(t *testing.T) {
	source, _ := profileHome(t)
	first, second := t.TempDir(), t.TempDir()
	writeContextPackage(t, first, "one", "1.0.0", "one\n")
	writeContextPackage(t, second, "two", "1.0.0", "two\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", first); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "install", second); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "two", "--env", "codex_cli"); code != exitOK {
		t.Fatalf("scoped use stderr:\n%s", stderr)
	}
	_, stdout, _ := runProfile(t, source, "profile", "list")
	if !strings.Contains(stdout, "env:codex_cli=two") {
		t.Fatalf("list stdout:\n%s", stdout)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "--clear", "--env", "codex_cli"); code != exitOK {
		t.Fatalf("clear stderr:\n%s", stderr)
	}
	_, stdout, _ = runProfile(t, source, "profile", "list")
	if strings.Contains(stdout, "env:codex_cli") {
		t.Fatalf("list after clear:\n%s", stdout)
	}
}

// TestProfileUpdatePinnedTagIsUnchanged checks a tag-pinned git profile
// reports unchanged through the CLI. The repository is served under a fake
// network identity (git insteadOf): a file:// operand is rejected (see
// TestProfileInstallFileOperandIsRefused) and must never reach this path.
func TestProfileUpdatePinnedTagIsUnchanged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
	source, _ := profileHome(t)
	repo := t.TempDir()
	writeContextPackage(t, repo, "groot", "1.0.0", "git module\n")
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init")
	git("add", ".")
	git("commit", "-m", "one")
	git("tag", "v1.0.0")
	gitconfig := filepath.Join(t.TempDir(), "gitconfig")
	rewrite := "[url \"file://" + repo + "\"]\n\tinsteadOf = https://example.com/groot\n"
	if err := os.WriteFile(gitconfig, []byte(rewrite), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	if code, _, stderr := runProfile(t, source, "profile", "install", "https://example.com/groot", "--tag", "v1.0.0"); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "update", "groot")
	if code != exitOK {
		t.Fatalf("update = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "unchanged") {
		t.Fatalf("update stdout:\n%s", stdout)
	}
}

// TestProfileInstallFileOperandIsRefused checks F12 through run(): a file://
// git operand is rejected with profile_source_invalid, never installed.
func TestProfileInstallFileOperandIsRefused(t *testing.T) {
	source, _ := profileHome(t)
	code, _, stderr := runProfile(t, source, "profile", "install", "file:///tmp/never-there-pkg")
	if code != exitFail {
		t.Fatalf("install file:// = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
	}
	if !strings.Contains(stderr, "profile_source_invalid") || !strings.Contains(stderr, "no network identity") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

// TestProfileListMigrationHonoursSystemPolicy checks F10 through run(): the
// CLI threads its already-loaded machine configuration into the builtin
// default migration, so a global skill outside the system-locked
// allowed_sources refuses `profile list` instead of migrating. The refusal
// precedes any clone, so no git fixture is needed.
func TestProfileListMigrationHonoursSystemPolicy(t *testing.T) {
	home := t.TempDir()
	writeTestProfileConfig(t, filepath.Join(home, "config.json"),
		`{"schema_version":1,"skills_root":"skills","projects":{}}`)
	if err := os.MkdirAll(filepath.Join(home, "global"), 0o755); err != nil {
		t.Fatal(err)
	}
	skillfile := `{"schema_version": 1, "skills": [{"name": "hello", ` +
		`"git": "https://example.com/skills/hello", "tag": "v1.0.0"}]}` + "\n"
	if err := os.WriteFile(filepath.Join(home, "global", "Skillfile.json"), []byte(skillfile), 0o644); err != nil {
		t.Fatal(err)
	}
	system := filepath.Join(t.TempDir(), "system.json")
	writeTestProfileConfig(t, system,
		`{"schema_version":1,"locked":["allowed_sources"],"allowed_sources":["github.com/relux-works"]}`)
	t.Setenv("CURATOR_SYSTEM_CONFIG", system)
	base := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(base, "claude"))
	t.Setenv("CODEX_HOME", filepath.Join(base, "codex"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "xdg"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(base, "pi"))
	var stdout, stderr strings.Builder
	code := run([]string{"profile", "list"}, fileConfigSource(filepath.Join(home, "config.json")), &stdout, &stderr)
	if code != exitFail {
		t.Fatalf("list = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitFail, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "profile_source_invalid") ||
		!strings.Contains(stderr.String(), "allowed sources") {
		t.Fatalf("stderr:\n%s", stderr.String())
	}
}

func writeTestProfileConfig(t *testing.T, path, payload string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
}
