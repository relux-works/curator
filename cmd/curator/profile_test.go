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

// TestProfileUpdatePinnedTagIsUnchanged checks a tag-pinned git profile
// reports unchanged through the CLI.
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
	if code, _, stderr := runProfile(t, source, "profile", "install", "file://"+repo, "--tag", "v1.0.0"); code != exitOK {
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
