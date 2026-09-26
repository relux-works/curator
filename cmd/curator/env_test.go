package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
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
	writeNativeCredentials(t)
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

func loadEnvironmentIsolationConfig(t *testing.T, source stubConfigSource, userJSON, systemJSON string) stubConfigSource {
	t.Helper()
	if err := os.WriteFile(source.path, []byte(userJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	if systemJSON == "" {
		t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	} else {
		systemPath := filepath.Join(filepath.Dir(source.path), "system.json")
		if err := os.WriteFile(systemPath, []byte(systemJSON), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	}
	cfg, err := config.Load(source.path, nil)
	if err != nil {
		t.Fatalf("config.Load rejected isolation fixture: %v", err)
	}
	source.cfg = cfg
	return source
}

func installEnvTestProfile(t *testing.T, source stubConfigSource) {
	t.Helper()
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("profile install = %d\nstderr:\n%s", code, stderr)
	}
}

// TestEnvResolveIsolatedSystemLockUsesDirection drives system-config-v2
// environments.isolation through the CLI. With no user isolation entry, the
// locked isolated direction provisions a managed home without shared auth
// passthrough.
func TestEnvResolveIsolatedSystemLockUsesDirection(t *testing.T) {
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	user := `{"schema_version": 2, "skills_root": "x", "projects": {}}`
	system := `{"schema_version": 2, "locked": ["environments.isolation"], "environments": {"isolation": {"acme": {"codex_cli": "isolated"}}}}`
	source = loadEnvironmentIsolationConfig(t, source, user, system)
	if !source.cfg.Locked["environments.isolation"] || source.cfg.Env.Isolation["acme"]["codex_cli"] != "isolated" {
		t.Fatalf("loaded isolation lock = %+v, locked=%v", source.cfg.Env.Isolation, source.cfg.Locked)
	}
	installEnvTestProfile(t, source)
	code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair")
	if code != exitOK {
		t.Fatalf("resolve under isolated lock = %d\nstderr:\n%s", code, stderr)
	}
	link := filepath.Join(source.cfg.Home(), "environments", "acme", "codex_cli", "auth.json")
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("isolated locked home has no shared auth passthrough: %s (%v)", link, err)
	}
}

// TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock proves an
// explicit shared user value is rejected through env resolve, with the
// profile, environment, and system policy source in the diagnostic.
func TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock(t *testing.T) {
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	base := `{"schema_version": 2, "skills_root": "x", "projects": {}}`
	source = loadEnvironmentIsolationConfig(t, source, base, "")
	installEnvTestProfile(t, source)
	user := `{"schema_version": 2, "skills_root": "x", "projects": {}, "environments": {"isolation": {"acme": {"codex_cli": "shared"}}}}`
	system := `{"schema_version": 2, "locked": ["environments.isolation"], "environments": {"isolation": {"acme": {"codex_cli": "isolated"}}}}`
	source = loadEnvironmentIsolationConfig(t, source, user, system)
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair")
	if code != exitFail {
		t.Fatalf("resolve with explicit shared under isolated lock = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	for _, want := range []string{"environment_isolation_lock_conflict", "acme", "codex_cli", source.cfg.SystemConfigPath} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("lock refusal misses %q:\n%s", want, stderr)
		}
	}
	if stdout != "" {
		t.Fatalf("conflicting resolve emits no fragment: %q", stdout)
	}
}

// TestEnvResolveLockedIsolationRequiresMigration provisions a shared home,
// engages an isolated system lock with machine isolation silent, and proves
// resolve refuses without moving the recorded passthrough. The exact F-C2
// migration named by the refusal then plans and applies the unlink.
func TestEnvResolveLockedIsolationRequiresMigration(t *testing.T) {
	requireLinkCapability(t)
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	user := `{"schema_version": 2, "skills_root": "x", "projects": {}}`
	source = loadEnvironmentIsolationConfig(t, source, user, "")
	installEnvTestProfile(t, source)
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("shared provision = %d\nstderr:\n%s", code, stderr)
	}
	link := filepath.Join(source.cfg.Home(), "environments", "acme", "codex_cli", "auth.json")
	nativeAuth := filepath.Join(os.Getenv("CODEX_HOME"), "auth.json")
	target := nativeAuth
	if got, err := os.Readlink(link); err != nil || got != target {
		t.Fatalf("shared passthrough target = %q (%v), want %q", got, err, target)
	}
	nativeBytes, err := os.ReadFile(nativeAuth)
	if err != nil {
		t.Fatal(err)
	}
	system := `{"schema_version": 2, "locked": ["environments.isolation"], "environments": {"isolation": {"acme": {"codex_cli": "isolated"}}}}`
	source = loadEnvironmentIsolationConfig(t, source, user, system)
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair")
	if code != exitFail {
		t.Fatalf("resolve with existing shared passthrough = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	for _, want := range []string{"environment_credential_conflict", "migration needed", "env migrate --plan --profile acme --env codex_cli", "env migrate --apply --expect <plan-hash> --profile acme --env codex_cli"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("migration refusal misses %q:\n%s", want, stderr)
		}
	}
	if got, err := os.Readlink(link); err != nil || got != target {
		t.Fatalf("refused resolve preserves shared link: %q (%v)", got, err)
	}
	if got, err := os.ReadFile(nativeAuth); err != nil || string(got) != string(nativeBytes) {
		t.Fatalf("refused resolve preserves native credential bytes: %q (%v)", got, err)
	}
	code, plan, stderr := runProfile(t, source, "env", "migrate", "--plan", "--profile", "acme", "--env", "codex_cli")
	if code != exitOK || !strings.Contains(plan, "unlink") || !strings.Contains(plan, "auth.json") {
		t.Fatalf("F-C2 plan = %d\nplan:\n%s\nstderr:\n%s", code, plan, stderr)
	}
	code, applied, stderr := runProfile(t, source, "env", "migrate", "--apply", "--expect", planHash(t, plan), "--profile", "acme", "--env", "codex_cli")
	if code != exitOK {
		t.Fatalf("F-C2 apply = %d\nstdout:\n%s\nstderr:\n%s", code, applied, stderr)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("F-C2 migration removes the shared passthrough: %s (%v)", link, err)
	}
	if got, err := os.ReadFile(nativeAuth); err != nil || string(got) != string(nativeBytes) {
		t.Fatalf("F-C2 migration preserves native credential bytes: %q (%v)", got, err)
	}
}

// TestEnvStatusMatrix drives status before and after provisioning, with
// --check and --json.
func TestEnvStatusMatrix(t *testing.T) {
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	// The §12 provider rows join the matrix: run and session are always
	// reported, and a missing row is non-current — so the post-repair
	// --check plants stub providers, which warn outside the trust
	// roots under revision A but stay current.
	bin := t.TempDir()
	for _, name := range []string{"curator-run", "curator-session"} {
		full := filepath.Join(bin, name)
		if runtime.GOOS == "windows" {
			full += ".exe"
		}
		if err := os.WriteFile(full, []byte(""), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
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
	for _, want := range []string{"provider run:", "provider session:", "subcommand_provider_outside_trust_roots"} {
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
	for _, key := range []string{"homes", "scopes", "adapters", "targets", "profiles", "providers", "unregistered_environments", "orphans", "notes", "non_current"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("status --json misses snake_case member %q: %v", key, decoded)
		}
	}
	if _, ok := decoded["Homes"]; ok {
		t.Fatalf("status --json publishes Go field names: %v", decoded)
	}
	providers, _ := decoded["providers"].([]any)
	names := map[string]bool{}
	for _, entry := range providers {
		row, _ := entry.(map[string]any)
		if row == nil {
			t.Fatalf("provider row is not an object: %v", entry)
		}
		for _, key := range []string{"name", "executable", "resolved", "verdict", "current", "trust_roots_consulted"} {
			if _, ok := row[key]; !ok {
				t.Fatalf("provider row misses snake_case member %q: %v", key, row)
			}
		}
		name, _ := row["name"].(string)
		names[name] = true
	}
	if !names["run"] || !names["session"] {
		t.Fatalf("providers miss the always-reported rows: %v", names)
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
	// The temp bin is outside the trust roots, so revision A dispatches
	// with the migration warning.
	if !strings.Contains(stderr, "subcommand_provider_outside_trust_roots") {
		t.Fatalf("stderr carries no outside-roots warning:\n%s", stderr)
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

// TestEnvStatusS4Posture drives the §12 posture through run(): with an
// absent knob the text names s4-warn with an unbounded effective list,
// warns the empty allowlist, and prints no rows for the empty MCP set;
// the JSON carries the same posture members.
func TestEnvStatusS4Posture(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, stdout, _ := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "s4_profile: s4-warn, passable_env_names: unbounded") {
		t.Fatalf("status postures no S4 profile:\n%s", stdout)
	}
	if !strings.Contains(stdout, "warning: mcp_package_allowlist_empty") {
		t.Fatalf("status warns no empty allowlist:\n%s", stdout)
	}
	if strings.Contains(stdout, "mcp-declaration") {
		t.Fatalf("an empty MCP set must print no rows:\n%s", stdout)
	}
	code, jsonOut, _ := runProfile(t, source, "env", "status", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d", code)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Fatalf("status --json is not JSON: %v", err)
	}
	if decoded["s4_profile"] != "s4-warn" {
		t.Fatalf("s4_profile = %v", decoded["s4_profile"])
	}
	if decoded["passable_env_names"] != nil {
		t.Fatalf("passable_env_names = %v, want null for unbounded", decoded["passable_env_names"])
	}
	warnings, _ := decoded["warnings"].([]any)
	found := false
	for _, raw := range warnings {
		found = found || strings.HasPrefix(raw.(string), "mcp_package_allowlist_empty:")
	}
	if !found {
		t.Fatalf("warnings = %v, want the allowlist-empty row", warnings)
	}
}

// TestEnvStatusEffectivePassableList drives the posture with a
// configured knob: the text renders the effective list.
func TestEnvStatusEffectivePassableList(t *testing.T) {
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {},`+
		`"environments": {"passable_env_names": ["A_B"]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, stdout, _ := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, `s4_profile: s4-warn, passable_env_names: ["A_B"]`) {
		t.Fatalf("status postures no effective list:\n%s", stdout)
	}
}
