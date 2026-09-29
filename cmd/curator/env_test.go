package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envprofile"
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
	if fragment["fragment"] != "launch-env-fragment-v2" || fragment["environment"] != "codex_cli" {
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

// TestEnvResolveRejectsSwappedStoreEntry drives the shipped CLI entry and
// proves a store entry whose bytes no longer match the lock pin emits no
// launch fragment.
func TestEnvResolveRejectsSwappedStoreEntry(t *testing.T) {
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "trusted\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("profile install = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("env resolve --repair = %d\nstderr:\n%s", code, stderr)
	}
	profiles, err := envprofile.List(source.cfg.Home())
	if err != nil {
		t.Fatal(err)
	}
	var member contextlock.Member
	for _, profile := range profiles {
		if profile.Name != "acme" || profile.Lock == nil {
			continue
		}
		member, _ = profile.Lock.Find(contextlock.KindContext, "acme")
	}
	if member.StateHash == "" {
		t.Fatal("installed path context has no state pin")
	}
	entry := contextstore.EntryDir(source.cfg.Home(), member.Kind, member.Name, member.StateHash)
	swapped := filepath.Join(entry, "context", "a.md")
	if err := os.WriteFile(swapped, []byte("swapped\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair", "--dry-run")
	if code != exitFail || stdout != "" || !strings.Contains(stderr, "environment_repair_failed") {
		t.Fatalf("dry-run repair = %d, stdout %q\nstderr:\n%s", code, stdout, stderr)
	}
	if contents, err := os.ReadFile(swapped); err != nil || string(contents) != "swapped\n" {
		t.Fatalf("dry-run repair changed swapped store bytes: %q (%v)", contents, err)
	}
	code, stdout, stderr = runProfile(t, source, "env", "resolve", "codex_cli")
	if code != exitFail || stdout != "" {
		t.Fatalf("swapped store resolve = %d, stdout %q\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "environment_store_untrusted") || !strings.Contains(stderr, "pin_hash") {
		t.Fatalf("resolve did not name the untrusted pin hash:\n%s", stderr)
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

func TestEnvStatusRegistryBoundaryPostureAndCheck(t *testing.T) {
	source, home := profileHome(t)
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("profile install stderr:\n%s", stderr)
	}
	for _, profile := range []string{"acme", "default"} {
		for _, environment := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
			if code, _, stderr := runProfile(t, source, "env", "resolve", environment, "--profile", profile, "--repair"); code != exitOK {
				t.Fatalf("repair %s %s stderr:\n%s", profile, environment, stderr)
			}
		}
	}
	const registryURL = "https://registry.example.test"
	source.cfg.DisableBuiltinRegistries = true
	source.cfg.AuditRegistries = []config.Registry{{Name: "trusted", URL: registryURL, PublicKeys: []string{"ed25519:test"}, Enabled: true}}
	bin := t.TempDir()
	for _, name := range []string{"curator-run", "curator-session"} {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		if err := os.WriteFile(filepath.Join(bin, name), []byte("stub"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	code, stdout, stderr := runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitOK {
		t.Fatalf("first-use registry status --check = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	var report struct {
		NonCurrent      bool `json:"non_current"`
		RegistryPosture []struct {
			Name                 string `json:"name"`
			URL                  string `json:"url"`
			HighWaterVersion     *int   `json:"high_water_version"`
			HighWaterLogSize     *int   `json:"high_water_log_size"`
			LastBoundaryVerified bool   `json:"last_boundary_verified"`
			BootstrapSource      string `json:"bootstrap_source"`
			BootstrapDiagnostic  string `json:"bootstrap_diagnostic"`
			BootstrapSeverity    string `json:"bootstrap_severity"`
			MirrorGroup          string `json:"mirror_group"`
			MirrorComparison     string `json:"last_mirror_comparison"`
			MirrorDiagnostic     string `json:"mirror_diagnostic"`
			MirrorSeverity       string `json:"mirror_severity"`
			Diagnostic           string `json:"diagnostic"`
		} `json:"registry_posture"`
	}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("status JSON: %v\n%s", err, stdout)
	}
	if len(report.RegistryPosture) != 1 {
		t.Fatalf("registry posture rows = %+v, want one configured trusted registry", report.RegistryPosture)
	}
	row := report.RegistryPosture[0]
	if row.Name != "trusted" || row.URL != registryURL || row.HighWaterVersion != nil || row.HighWaterLogSize != nil || row.LastBoundaryVerified || row.Diagnostic != "" || report.NonCurrent {
		t.Fatalf("first-use posture = %+v non_current=%v", row, report.NonCurrent)
	}

	stateDir := filepath.Join(home, "state", "registry")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(registryURL))
	stateName := "snapshot-" + hex.EncodeToString(sum[:])[:16] + ".json"
	statePath := filepath.Join(stateDir, stateName)
	statePayload, err := json.Marshal(map[string]any{
		"highest_version":   17,
		"head":              strings.Repeat("b", 64),
		"merkle_root":       strings.Repeat("a", 64),
		"log_size":          12,
		"boundary_verified": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, statePayload, 0o600); err != nil {
		t.Fatal(err)
	}
	catalog := `{"schema_version":1,"states":["` + stateName + `"]}`
	if err := os.WriteFile(filepath.Join(stateDir, "known-registries.json"), []byte(catalog), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitOK {
		t.Fatalf("persisted registry status --check = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("status JSON with persisted high-water: %v\n%s", err, stdout)
	}
	row = report.RegistryPosture[0]
	if row.HighWaterVersion == nil || *row.HighWaterVersion != 17 || row.HighWaterLogSize == nil || *row.HighWaterLogSize != 12 || !row.LastBoundaryVerified ||
		row.BootstrapSource != "first-use" || row.BootstrapDiagnostic != "registry_bootstrap_tofu" || row.BootstrapSeverity != "warning" || row.Diagnostic != "" {
		t.Fatalf("persisted registry posture = %+v", row)
	}

	statePayload, err = json.Marshal(map[string]any{
		"highest_version": 17, "head": strings.Repeat("b", 64), "merkle_root": strings.Repeat("a", 64),
		"log_size": 12, "boundary_verified": true, "bootstrap_source": "first-use", "mirror_group": "prod",
		"last_mirror_comparison": "diverged", "last_mirror_comparison_log_size": 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, statePayload, 0o600); err != nil {
		t.Fatal(err)
	}
	source.cfg.AuditRegistries[0].MirrorGroup = "prod"
	source.cfg.Audit.RegistryPolicy = "strict"
	code, stdout, stderr = runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitFail {
		t.Fatalf("strict divergent registry status --check = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitFail, stdout, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("status JSON with divergence: %v\n%s", err, stdout)
	}
	row = report.RegistryPosture[0]
	if !report.NonCurrent || row.MirrorGroup != "prod" || row.MirrorComparison != "diverged" || row.MirrorDiagnostic != "registry_view_divergence" || row.MirrorSeverity != "error" || !strings.Contains(row.Diagnostic, "registry_view_divergence") {
		t.Fatalf("strict divergence posture = %+v non_current=%v", row, report.NonCurrent)
	}
	if err := os.WriteFile(statePath, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runProfile(t, source, "env", "status", "--check", "--json")
	if code != exitFail {
		t.Fatalf("unreadable registry rollback state --check = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitFail, stdout, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("status JSON with unreadable state: %v\n%s", err, stdout)
	}
	if !report.NonCurrent || len(report.RegistryPosture) != 1 || !strings.Contains(report.RegistryPosture[0].Diagnostic, "unreadable") {
		t.Fatalf("unreadable registry state posture = %+v non_current=%v", report.RegistryPosture, report.NonCurrent)
	}
	entries, err := os.ReadDir(stateDir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("read-only status changed rollback state directory: entries=%v error=%v", entries, err)
	}
}

// TestEnvStatusReportsDroppedSystemModuleThroughCLI proves env status prints
// the effective admission policy and each dropped system module's package
// and path through the production command entry point.
func TestEnvStatusReportsDroppedSystemModuleThroughCLI(t *testing.T) {
	requireGit(t)
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	leaf := t.TempDir()
	writeGitRepoFile(t, leaf, "agent-context.json", `{"schema_version": 1, "name": "sysleaf", "version": "1.0.0",`+
		`"context": {"modules": [{"path": "90-system.md", "class": "system"}]}}`+"\n")
	writeGitRepoFile(t, leaf, "context/90-system.md", "Leaf system prompt.\n")
	runGitRepo(t, leaf, "init")
	commitGitRepo(t, leaf, "v1.0.0")
	mid := t.TempDir()
	writeGitRepoFile(t, mid, "agent-context.json", `{"schema_version": 1, "name": "sysmid", "version": "1.0.0",`+
		`"requires": {"contexts": {"sysleaf": {"git": "https://example.com/sysleaf", "range": "*"}}},`+
		`"context": {"modules": [{"path": "00-mid.md"}]}}`+"\n")
	writeGitRepoFile(t, mid, "context/00-mid.md", "Mid context.\n")
	runGitRepo(t, mid, "init")
	commitGitRepo(t, mid, "v1.0.0")
	root := t.TempDir()
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "sysroot", "version": "1.0.0",`+
		`"requires": {"contexts": {"sysmid": {"git": "https://example.com/sysmid", "range": "*"}}},`+
		`"context": {"modules": [{"path": "00-root.md"}]}}`+"\n")
	writeGitRepoFile(t, root, "context/00-root.md", "Root context.\n")
	runGitRepo(t, root, "init")
	commitGitRepo(t, root, "v1.0.0")
	serveGitRepos(t, map[string]string{
		"https://example.com/sysleaf": leaf,
		"https://example.com/sysmid":  mid,
		"https://example.com/sysroot": root,
	})
	if code, _, stderr := runProfile(t, source, "profile", "install", "https://example.com/sysroot"); code != exitOK {
		t.Fatalf("profile install = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "claude_code", "--profile", "sysroot", "--repair"); code != exitOK {
		t.Fatalf("env resolve --repair = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("env status = %d\nstderr:\n%s", code, stderr)
	}
	for _, want := range []string{
		"profile sysroot:",
		"transitive_system_modules=drop",
		"dropped system module sysleaf 90-system.md",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("env status omits %q:\n%s", want, stdout)
		}
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
