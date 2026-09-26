package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// writeNativeCredentials seeds live native credential targets for the
// homes profileHome redirects: codex auth.json, the pi agent-root
// auth.json, and the Linux claude_code credentials file. The seeds keep
// the detached-pending finding out of tests about other behaviors, and
// the writes are harmless where an adapter links nothing.
func writeNativeCredentials(t *testing.T) {
	t.Helper()
	if dir := os.Getenv("CODEX_HOME"); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if dir := os.Getenv("PI_CODING_AGENT_DIR"); dir != "" {
		if err := os.MkdirAll(filepath.Join(dir, "agent"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "agent", "auth.json"), []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".credentials.json"), []byte("{\"t\":\"operator-claude\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// gitFileURL renders a local fixture repository path as the file:// remote
// a git [url ...] section matches. The value is written into git
// configuration, where a backslash is an escape: a raw Windows temporary
// path collapses and the clone fails with "does not appear to be a git
// repository". Forward slashes parse identically on every platform, and the
// three-slash form keeps a Windows volume out of the host position —
// never interpolate a raw t.TempDir() into a git config value.
func gitFileURL(path string) string {
	slashed := filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" {
		return "file:///" + slashed
	}
	return "file://" + slashed
}

// TestGitFileURLHasNoBackslash pins the fixture wiring: the insteadOf
// remote must reach git configuration without a backslash, which git
// parses as an escape. A mutant that restores the raw interpolation fails
// this test on Windows, where the temporary directory carries separators
// git would consume; the volume-branch grammar itself is proven by the
// envprofile mapping test on every platform.
func TestGitFileURLHasNoBackslash(t *testing.T) {
	repo := t.TempDir()
	url := gitFileURL(repo)
	if url != "file://"+filepath.ToSlash(repo) && url != "file:///"+filepath.ToSlash(repo) {
		t.Fatalf("remote %q, want the file URL of %q", url, repo)
	}
	if strings.Contains(url, "\\") {
		t.Fatalf("remote %q carries a backslash git parses as an escape", url)
	}
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

// TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable drives
// four path forms through the production CLI entry point: an absent operand,
// an unreadable root, an unreadable module directory, and a regular file.
// Absence and failed reads must never report each other's diagnostic.
func TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable(t *testing.T) {
	tests := []struct {
		name          string
		want          string
		forbidden     []string
		skipOnWindows bool
		preparePath   func(*testing.T) (string, func())
	}{
		{
			name:      "missing_operand",
			want:      "profile_source_path_missing",
			forbidden: []string{"profile_source_path_unreadable"},
			preparePath: func(t *testing.T) (string, func()) {
				return filepath.Join(t.TempDir(), "absent"), func() {}
			},
		},
		{
			name:          "unreadable_root",
			want:          "profile_source_path_unreadable",
			forbidden:     []string{"profile_source_path_missing", "profile_source_invalid"},
			skipOnWindows: true,
			preparePath: func(t *testing.T) (string, func()) {
				pkg := t.TempDir()
				writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
				return pkg, makeProfileTestPathUnreadable(t, pkg, true)
			},
		},
		{
			name:          "unreadable_module_directory",
			want:          "profile_source_path_unreadable",
			forbidden:     []string{"profile_source_path_missing", "profile_source_invalid"},
			skipOnWindows: true,
			preparePath: func(t *testing.T) (string, func()) {
				pkg := t.TempDir()
				writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
				modules := filepath.Join(pkg, "context")
				cleanup := makeProfileTestPathUnreadable(t, modules, true)
				return pkg, cleanup
			},
		},
		{
			name:      "regular_file",
			want:      "profile_source_invalid",
			forbidden: []string{"profile_source_path_missing", "profile_source_path_unreadable"},
			preparePath: func(t *testing.T) (string, func()) {
				file := filepath.Join(t.TempDir(), "package")
				if err := os.WriteFile(file, []byte("not a context package\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				return file, func() {}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skipOnWindows && runtime.GOOS == "windows" {
				t.Skip("the unreadable case uses POSIX mode bits; Windows ACL unreadability is not reproducible for the runner account")
			}
			source, _ := profileHome(t)
			path, cleanup := test.preparePath(t)
			defer cleanup()
			code, _, stderr := runProfile(t, source, "profile", "install", path)
			if code != exitFail {
				t.Fatalf("install %q = %d, want failure; stderr:\n%s", path, code, stderr)
			}
			if !strings.Contains(stderr, test.want) {
				t.Fatalf("stderr %q does not contain %q", stderr, test.want)
			}
			for _, forbidden := range test.forbidden {
				if strings.Contains(stderr, forbidden) {
					t.Fatalf("stderr %q must not contain %q", stderr, forbidden)
				}
			}
		})
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

// TestProfileComposeListsDeclaredOverlays checks compose list reads the
// machine overlays.<profile> list. (Stage (c) implements the compose row of
// cli/curator.md, replacing the earlier refusal stub; the add/remove/list
// paths carry their own tests in envconfig_test.go.)
func TestProfileComposeListsDeclaredOverlays(t *testing.T) {
	source, _ := profileHome(t)
	code, stdout, stderr := runProfile(t, source, "profile", "compose", "acme", "list")
	if code != exitOK {
		t.Fatalf("compose = %d\nstderr:\n%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("empty overlay list printed %q", stdout)
	}
	code, _, _ = runProfile(t, source, "profile", "compose", "acme", "frobnicate")
	if code != exitUsage {
		t.Fatalf("unknown compose action = %d, want usage", code)
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
// bound instead of touching fixed homes: an undeclared target is
// environment_target_unknown, while a declared target names the deferred
// writes instead of misreporting unknown.
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
	code, _, stderr = runProfile(t, source, "profile", "use", "acme", "--target", "xcode-coding-assistant")
	if code != exitFail || strings.Contains(stderr, "environment_target_unknown") {
		t.Fatalf("a declared target must not report unknown: %d\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "writes are deferred") {
		t.Fatalf("a declared target names the deferred writes: %d\n%s", code, stderr)
	}
}

// TestProfileUseTargetBoundToAdapter narrows the §7.6 per-adapter
// resolution gate through the production run() entry point: the same
// target resolves for the adapters that declare it and is
// environment_target_unknown for the adapters that do not. Weakening
// TargetFor to ignore its adapter argument (delegating to TargetByID)
// admits pi/opencode here and fails this test.
func TestProfileUseTargetBoundToAdapter(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	for _, env := range []string{"pi", "opencode"} {
		code, _, stderr := runProfile(t, source, "profile", "use", "acme", "--env", env, "--target", "xcode-coding-assistant")
		if code != exitFail || !strings.Contains(stderr, "environment_target_unknown") {
			t.Fatalf("use --env %s --target xcode-coding-assistant = %d, want environment_target_unknown\nstderr:\n%s", env, code, stderr)
		}
	}
	for _, env := range []string{"claude_code", "codex_cli"} {
		code, _, stderr := runProfile(t, source, "profile", "use", "acme", "--env", env, "--target", "xcode-coding-assistant")
		if code != exitFail || strings.Contains(stderr, "environment_target_unknown") {
			t.Fatalf("use --env %s --target xcode-coding-assistant must not report unknown: %d\nstderr:\n%s", env, code, stderr)
		}
		if !strings.Contains(stderr, "writes are deferred") {
			t.Fatalf("use --env %s --target xcode-coding-assistant names the deferred writes: %d\nstderr:\n%s", env, code, stderr)
		}
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
	rewrite := "[url \"" + gitFileURL(repo) + "\"]\n\tinsteadOf = https://example.com/groot\n"
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

// TestProfileMachineUseSkipsScopedAdapter drives F16 through run(): with
// env:codex_cli scoped to two, a machine-scope use of three leaves the codex
// home on two (bytes and marker) while the other homes move to three, the
// listing keeps reporting env:codex_cli=two, and the same holds through
// install --use. A mutant that restores the unconditional machine pass
// overwrites the codex home and must fail this test.
func TestProfileMachineUseSkipsScopedAdapter(t *testing.T) {
	source, _ := profileHome(t)
	one, two, three, four := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	writeContextPackage(t, one, "one", "1.0.0", "one\n")
	writeContextPackage(t, two, "two", "1.0.0", "two\n")
	writeContextPackage(t, three, "three", "1.0.0", "three\n")
	writeContextPackage(t, four, "four", "1.0.0", "four\n")
	for _, pkg := range []string{one, two} {
		if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
			t.Fatalf("install stderr:\n%s", stderr)
		}
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "two", "--env", "codex_cli"); code != exitOK {
		t.Fatalf("scoped use stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "two", "--env", "opencode"); code != exitOK {
		t.Fatalf("opencode scoped use stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "install", three); code != exitOK {
		t.Fatalf("install three stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "use", "three")
	if code != exitOK {
		t.Fatalf("use three = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if strings.Contains(stdout, "codex_cli") || strings.Contains(stdout, "opencode: switched") {
		t.Fatalf("machine use must skip both scoped adapters, stdout:\n%s", stdout)
	}
	codexDoc, err := os.ReadFile(filepath.Join(os.Getenv("CODEX_HOME"), "AGENTS.md")) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(codexDoc), "## Context: two 1.0.0") {
		t.Fatalf("codex home does not carry two:\n%s", codexDoc)
	}
	codexMarker, err := envmarker.Read(os.Getenv("CODEX_HOME"))
	if err != nil || codexMarker == nil || codexMarker.Profile.Name != "two" {
		t.Fatalf("codex marker %+v %v", codexMarker, err)
	}
	opencodeHome := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "opencode")
	opencodeDoc, err := os.ReadFile(filepath.Join(opencodeHome, "AGENTS.md")) // #nosec G304 -- test home
	if err != nil || !strings.Contains(string(opencodeDoc), "## Context: two 1.0.0") {
		t.Fatalf("opencode AGENTS.md err=%v, content:\n%s", err, opencodeDoc)
	}
	opencodeMarker, err := envmarker.Read(opencodeHome)
	if err != nil || opencodeMarker == nil || opencodeMarker.Profile.Name != "two" {
		t.Fatalf("opencode marker %+v %v", opencodeMarker, err)
	}
	claudeDoc, err := os.ReadFile(filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md")) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(claudeDoc), "## Context: three 1.0.0") {
		t.Fatalf("claude home does not carry three:\n%s", claudeDoc)
	}
	_, stdout, _ = runProfile(t, source, "profile", "list")
	if !strings.Contains(stdout, "env:codex_cli=two") || !strings.Contains(stdout, "env:opencode=two") {
		t.Fatalf("list stdout:\n%s", stdout)
	}
	code, stdout, stderr = runProfile(t, source, "profile", "install", four, "--use")
	if code != exitOK {
		t.Fatalf("install four --use = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if strings.Contains(stdout, "codex_cli") || strings.Contains(stdout, "opencode: switched") {
		t.Fatalf("install --use must skip both scoped adapters, stdout:\n%s", stdout)
	}
	codexDoc, err = os.ReadFile(filepath.Join(os.Getenv("CODEX_HOME"), "AGENTS.md")) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(codexDoc), "## Context: two 1.0.0") {
		t.Fatalf("codex home does not carry two after install --use:\n%s", codexDoc)
	}
	claudeDoc, err = os.ReadFile(filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md")) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(claudeDoc), "## Context: four 1.0.0") {
		t.Fatalf("claude home does not carry four:\n%s", claudeDoc)
	}
	opencodeDoc, err = os.ReadFile(filepath.Join(opencodeHome, "AGENTS.md")) // #nosec G304 -- test home
	if err != nil || !strings.Contains(string(opencodeDoc), "## Context: two 1.0.0") {
		t.Fatalf("opencode home changed after install --use: err=%v\n%s", err, opencodeDoc)
	}
}

// TestProfileUseReportsAllAdaptersAlreadyScoped drives the machine-scope
// use command through run() when every registered adapter has a scoped
// current. The empty materialization set must produce an operator-visible
// line instead of a silent success.
func TestProfileUseReportsAllAdaptersAlreadyScoped(t *testing.T) {
	source, _ := profileHome(t)
	one, two, three := t.TempDir(), t.TempDir(), t.TempDir()
	writeContextPackage(t, one, "one", "1.0.0", "one\n")
	writeContextPackage(t, two, "two", "1.0.0", "two\n")
	writeContextPackage(t, three, "three", "1.0.0", "three\n")
	for _, pkg := range []string{one, two} {
		if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
			t.Fatalf("install stderr:\n%s", stderr)
		}
	}
	for _, adapter := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
		if code, _, stderr := runProfile(t, source, "profile", "use", "two", "--env", adapter); code != exitOK {
			t.Fatalf("scope %s stderr:\n%s", adapter, stderr)
		}
	}
	if code, _, stderr := runProfile(t, source, "profile", "install", three); code != exitOK {
		t.Fatalf("install three stderr:\n%s", stderr)
	}

	code, stdout, stderr := runProfile(t, source, "profile", "use", "three")
	if code != exitOK {
		t.Fatalf("machine use three = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "No adapter homes were switched; every registered adapter had a scoped current.") {
		t.Fatalf("empty machine switch was silent or unclear; stdout:\n%s", stdout)
	}
	if strings.Contains(stdout, ": switched (") {
		t.Fatalf("a machine switch with all adapters scoped reported a surface write:\n%s", stdout)
	}
	listCode, listOut, listErr := runProfile(t, source, "profile", "list")
	if listCode != exitOK || strings.Contains(listErr, "curator:") {
		t.Fatalf("profile list = %d\nstderr:\n%s", listCode, listErr)
	}
	for _, adapter := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
		if !strings.Contains(listOut, "env:"+adapter+"=two") {
			t.Fatalf("scope %s did not remain visible:\n%s", adapter, listOut)
		}
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

// TestProfileListMigrationHonoursSystemPolicy checks F10/F15 through run():
// the CLI threads its already-loaded machine configuration into the builtin
// default migration on every command that reaches it, so a global skill
// outside the system-locked allowed_sources refuses `profile list`, `use`,
// `sync` and `update` instead of migrating. The refusal precedes any clone,
// so no git fixture is needed. A mutant that drops PolicyFromConfig from
// the use or sync call sites (M-D) admits the migration there and must fail
// the corresponding subtest.
func TestProfileListMigrationHonoursSystemPolicy(t *testing.T) {
	for _, command := range [][]string{
		{"profile", "list"},
		{"profile", "use", "default"},
		{"profile", "sync"},
		{"profile", "update", "default"},
	} {
		command := command
		t.Run(strings.Join(command[1:], "-"), func(t *testing.T) {
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
			code := run(command, fileConfigSource(filepath.Join(home, "config.json")), &stdout, &stderr)
			if code != exitFail {
				t.Fatalf("%v = %d, want %d\nstdout:\n%s\nstderr:\n%s", command, code, exitFail, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), "profile_source_invalid") ||
				!strings.Contains(stderr.String(), "allowed sources") {
				t.Fatalf("%v stderr:\n%s", command, stderr.String())
			}
		})
	}
}

// TestProfileInstallUseActivatesThroughSwitch checks F13 through run():
// installing a second profile with --use on a machine current on another
// profile leaves the marker and the materialized bytes in agreement with
// the recorded current. A pointer-only activation (the pre-fix shape) or a
// mutant that drops materialization for one adapter leaves them on the
// previous profile and must fail this test.
func TestProfileInstallUseActivatesThroughSwitch(t *testing.T) {
	source, _ := profileHome(t)
	first, second := t.TempDir(), t.TempDir()
	writeContextPackage(t, first, "alpha", "1.0.0", "alpha\n")
	writeContextPackage(t, second, "beta", "1.0.0", "beta\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", first); code != exitOK {
		t.Fatalf("install alpha stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "install", second, "--use")
	if code != exitOK {
		t.Fatalf("install beta --use = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "installed and activated profile beta") {
		t.Fatalf("stdout:\n%s", stdout)
	}
	if !strings.Contains(stdout, "claude_code: switched") {
		t.Fatalf("install --use must report the switch, stdout:\n%s", stdout)
	}
	code, stdout, _ = runProfile(t, source, "profile", "list")
	if code != exitOK || !strings.Contains(stdout, "beta") || !strings.Contains(stdout, "current") {
		t.Fatalf("list stdout:\n%s", stdout)
	}
	claude := filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md")
	payload, err := os.ReadFile(claude) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "## Context: beta 1.0.0") {
		t.Fatalf("CLAUDE.md does not carry beta:\n%s", payload)
	}
	marker, err := envmarker.Read(os.Getenv("CLAUDE_CONFIG_DIR"))
	if err != nil || marker == nil || marker.Profile.Name != "beta" {
		t.Fatalf("marker %+v %v", marker, err)
	}
}

// TestProfileInstallUsePartialLeavesCurrent checks F13 through run(): a
// forced per-adapter failure during install --use leaves the previous
// current and reports the partial result. A pointer-only activation moves
// current and reports success, and must fail this test.
func TestProfileInstallUsePartialLeavesCurrent(t *testing.T) {
	source, _ := profileHome(t)
	first, second := t.TempDir(), t.TempDir()
	writeContextPackage(t, first, "alpha", "1.0.0", "alpha\n")
	writeContextPackage(t, second, "beta", "1.0.0", "beta\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", first); code != exitOK {
		t.Fatalf("install alpha stderr:\n%s", stderr)
	}
	claude := os.Getenv("CLAUDE_CONFIG_DIR")
	if err := os.RemoveAll(claude); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claude, []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "install", second, "--use")
	if code != exitFail {
		t.Fatalf("install beta --use = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitFail, stdout, stderr)
	}
	if !strings.Contains(stderr, "profile_use_partial") {
		t.Fatalf("stderr:\n%s", stderr)
	}
	if !strings.Contains(stdout, "installed profile beta (lock ") {
		t.Fatalf("partial install must report its persisted profile, stdout:\n%s", stdout)
	}
	_, stdout, _ = runProfile(t, source, "profile", "list")
	alphaCurrent, betaCurrent := false, false
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(line, "alpha\t") && strings.Contains(line, "current") {
			alphaCurrent = true
		}
		if strings.HasPrefix(line, "beta\t") && strings.Contains(line, "current") {
			betaCurrent = true
		}
	}
	if !alphaCurrent {
		t.Fatalf("current must stay alpha, list stdout:\n%s", stdout)
	}
	if betaCurrent {
		t.Fatalf("beta must not be current, list stdout:\n%s", stdout)
	}
}

// TestProfileInstallFirstActivationFailureReportsPersistedProfile drives
// the early activation-refusal path through run(): the new profile lock is
// published, then the machine policy refuses the activation before
// materializeScope can return adapter rows. The CLI must still name the
// installed profile. Requiring an Activation row to print it is the
// narrowing mutant this test is intended to kill.
func TestProfileInstallFirstActivationFailureReportsPersistedProfile(t *testing.T) {
	source, home := profileHome(t)
	required := "required"
	source.cfg.Env.RequireCurrent = &required
	source.cfg.Locked = map[string]bool{"environments.require_current_profile": true}
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "beta", "1.0.0", "beta\n")

	code, stdout, stderr := runProfile(t, source, "profile", "install", pkg, "--use")
	if code != exitFail {
		t.Fatalf("first install --use = %d, want policy refusal\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "requires current profile \"required\"") {
		t.Fatalf("stderr does not show the pre-materialization refusal:\n%s", stderr)
	}
	if !strings.Contains(stdout, "installed profile beta (lock ") {
		t.Fatalf("persisted first install was not reported:\n%s", stdout)
	}
	if strings.Contains(stdout, "switched (") {
		t.Fatalf("activation was refused before materialization but reported an adapter switch:\n%s", stdout)
	}
	if _, err := os.Stat(filepath.Join(home, "profiles", "beta", "lock.json")); err != nil {
		t.Fatalf("profile lock was not persisted before activation refusal: %v", err)
	}
	listCode, listOut, listErr := runProfile(t, source, "profile", "list")
	if listCode != exitOK || !strings.Contains(listOut, "beta\tbeta\tpath") {
		t.Fatalf("persisted profile list = %d\nstdout:\n%s\nstderr:\n%s", listCode, listOut, listErr)
	}
}

// TestProfileReinstallActivationFailureKeepsUpdatedWording pins the
// reinstall wording while using the same pre-materialization refusal path:
// a changed path snapshot is published, activation is refused, and the CLI
// reports updated profile rather than relabeling it as a fresh install.
func TestProfileReinstallActivationFailureKeepsUpdatedWording(t *testing.T) {
	source, _ := profileHome(t)
	alpha, beta := t.TempDir(), t.TempDir()
	writeContextPackage(t, alpha, "alpha", "1.0.0", "alpha\n")
	writeContextPackage(t, beta, "beta", "1.0.0", "beta\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", alpha); code != exitOK {
		t.Fatalf("install alpha stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "install", beta); code != exitOK {
		t.Fatalf("install beta stderr:\n%s", stderr)
	}
	if err := os.WriteFile(filepath.Join(beta, "context", "a.md"), []byte("beta changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	required := "alpha"
	source.cfg.Env.RequireCurrent = &required
	source.cfg.Locked = map[string]bool{"environments.require_current_profile": true}

	code, stdout, stderr := runProfile(t, source, "profile", "install", beta, "--use")
	if code != exitFail {
		t.Fatalf("reinstall --use = %d, want policy refusal\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "updated profile beta (lock ") {
		t.Fatalf("partial reinstall lost its updated wording:\n%s", stdout)
	}
	if strings.Contains(stdout, "installed profile beta (lock ") {
		t.Fatalf("partial reinstall was labeled as a fresh install:\n%s", stdout)
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
