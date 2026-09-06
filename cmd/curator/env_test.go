package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// These tests drive the production run() entry point for the env rows and
// the umbrella dispatch.

// TestEnvResolveBareStaleIsAFailure narrows the fail-closed gate through
// the CLI: a stale home reports environment_home_stale on stderr with
// exit 1 and emits no fragment on stdout.
func TestEnvResolveBareStaleIsAFailure(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex_cli")
	if code != exitFail {
		t.Fatalf("resolve = %d, want %d", code, exitFail)
	}
	if !strings.Contains(stderr, "environment_home_stale") {
		t.Fatalf("stderr:\n%s", stderr)
	}
	if stdout != "" {
		t.Fatalf("a stale resolve emits no fragment, got %q", stdout)
	}
}

// TestEnvResolveRepairEmitsFragment provisions through the CLI and proves
// the fragment names the managed home, then proves the bare resolve is
// current with identical bytes.
func TestEnvResolveRepairEmitsFragment(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair")
	if code != exitOK {
		t.Fatalf("resolve --repair = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "Managed home provisioned at") {
		t.Fatalf("the first repair prints the first-resolve notice:\n%s", stderr)
	}
	var fragment map[string]any
	if err := json.Unmarshal([]byte(stdout), &fragment); err != nil {
		t.Fatalf("fragment is not JSON: %v\nstdout:\n%s", err, stdout)
	}
	if fragment["fragment"] != "launch-env-fragment-v1" || fragment["environment"] != "codex_cli" {
		t.Fatalf("fragment header: %v", fragment)
	}
	code, again, _ := runProfile(t, source, "env", "resolve", "codex_cli")
	if code != exitOK || again != stdout {
		t.Fatalf("bare resolve = %d, identical bytes: %v", code, again == stdout)
	}
	code, envOut, _ := runProfile(t, source, "env", "resolve", "codex_cli", "--format", "env")
	if code != exitOK || !strings.HasPrefix(envOut, "CODEX_HOME=") {
		t.Fatalf("env format = %d %q", code, envOut)
	}
	code, shellOut, _ := runProfile(t, source, "env", "resolve", "codex_cli", "--format", "shell")
	if code != exitOK || !strings.HasPrefix(shellOut, "export CODEX_HOME='") {
		t.Fatalf("shell format = %d %q", code, shellOut)
	}
}

// TestEnvResolveUnknowns covers the operand gates through the CLI.
func TestEnvResolveUnknowns(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "cursor"); code != exitFail || !strings.Contains(stderr, "environment_unknown") {
		t.Fatalf("resolve cursor = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--profile", "ghost"); code != exitFail || !strings.Contains(stderr, "profile_unknown") {
		t.Fatalf("resolve ghost = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, _ := runProfile(t, source, "env", "resolve", "codex_cli", "--format", "pwsh"); code == exitOK {
		t.Fatal("a pwsh format is not offered in revision 1")
	}
}

// TestEnvStatusMatrix drives status before and after provisioning, with
// --check and --json.
func TestEnvStatusMatrix(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, stdout, _ := runProfile(t, source, "env", "status", "--check")
	if code != exitFail {
		t.Fatalf("status --check on unprovisioned homes = %d, want %d\n%s", code, exitFail, stdout)
	}
	if !strings.Contains(stdout, "acme codex_cli: non-current, unprovisioned") {
		t.Fatalf("status rows:\n%s", stdout)
	}
	for _, profile := range []string{"acme", "default"} {
		for _, env := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
			if code, _, stderr := runProfile(t, source, "env", "resolve", env, "--profile", profile, "--repair"); code != exitOK {
				t.Fatalf("repair %s %s stderr:\n%s", profile, env, stderr)
			}
		}
	}
	code, stdout, _ = runProfile(t, source, "env", "status", "--check")
	if code != exitOK {
		t.Fatalf("status --check after repair = %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "acme codex_cli: current, provisioned") {
		t.Fatalf("status rows:\n%s", stdout)
	}
	if !strings.Contains(stdout, "tool codex_cli: recorded") {
		t.Fatalf("tool rows:\n%s", stdout)
	}
	for _, want := range []string{"mode managed-home", "form monolithic", "seeded-projects:", "profile acme:", "member context acme weight"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("status rows miss %q:\n%s", want, stdout)
		}
	}
	code, jsonOut, _ := runProfile(t, source, "env", "status", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d", code)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Fatalf("status --json is not JSON: %v", err)
	}
	for _, key := range []string{"homes", "scopes", "adapters", "targets", "profiles", "unregistered_environments", "orphans", "notes", "non_current"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("status --json misses snake_case member %q: %v", key, decoded)
		}
	}
	if _, ok := decoded["Homes"]; ok {
		t.Fatalf("status --json publishes Go field names: %v", decoded)
	}
}

// TestUmbrellaMissingProvider narrows the discovery gate: an unknown
// subcommand with no curator-<name> on PATH names the executable and the
// installation guidance, and downloads nothing.
func TestUmbrellaMissingProvider(t *testing.T) {
	source, _ := profileHome(t)
	t.Setenv("PATH", t.TempDir())
	code, _, stderr := runProfile(t, source, "run")
	if code != exitFail {
		t.Fatalf("run = %d, want %d", code, exitFail)
	}
	if !strings.Contains(stderr, "subcommand_provider_missing") || !strings.Contains(stderr, "curator-run") {
		t.Fatalf("stderr:\n%s", stderr)
	}
}

// TestUmbrellaDispatchesToProvider proves argv and PATH alone drive the
// dispatch: the provider receives the remaining arguments verbatim.
func TestUmbrellaDispatchesToProvider(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no bash on this platform")
	}
	source, _ := profileHome(t)
	bin := t.TempDir()
	script := "#!/bin/sh\necho \"provider got: $@\"\n"
	path := filepath.Join(bin, "curator-run")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	code, stdout, stderr := runProfile(t, source, "run", "alpha", "--beta")
	if code != exitOK {
		t.Fatalf("run = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "provider got: alpha --beta") {
		t.Fatalf("provider did not receive argv verbatim:\n%s", stdout)
	}
}

// TestUmbrellaPropagatesExitCode proves the provider's exit code is the
// command's exit code.
func TestUmbrellaPropagatesExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no bash on this platform")
	}
	source, _ := profileHome(t)
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "curator-fail"), []byte("#!/bin/sh\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if code, _, _ := runProfile(t, source, "fail"); code != 3 {
		t.Fatalf("fail = %d, want 3", code)
	}
}

// TestUmbrellaInvalidNameIsUsage proves dispatch input validation: a name
// outside the identifier grammar is a usage error, not a lookup.
func TestUmbrellaInvalidNameIsUsage(t *testing.T) {
	source, _ := profileHome(t)
	if code, _, _ := runProfile(t, source, "not a command!"); code != exitUsage {
		t.Fatalf("invalid name = %d, want %d", code, exitUsage)
	}
}

// TestImplementedCommandWinsOverProvider proves discovery runs only for
// unknown names: profile never dispatches even with a curator-profile on
// PATH.
func TestImplementedCommandWinsOverProvider(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no bash on this platform")
	}
	source, _ := profileHome(t)
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "curator-profile"), []byte("#!/bin/sh\necho shadowed\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	code, stdout, _ := runProfile(t, source, "profile")
	if code == exitOK || strings.Contains(stdout, "shadowed") {
		t.Fatalf("the implemented subcommand must win: %d %q", code, stdout)
	}
}
