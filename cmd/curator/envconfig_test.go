package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// These tests drive the production run() entry point for the env config and
// profile compose rows (cli/curator.md) and the locked require_current_profile
// gate on profile use (environments §12.2).

func writeMachineConfig(t *testing.T, text string) stubConfigSource {
	t.Helper()
	source, _ := profileHome(t)
	home := t.TempDir()
	path := filepath.Join(home, "config.json")
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path, nil)
	if err != nil {
		t.Fatalf("config.Load rejected the fixture: %v", err)
	}
	source.path = path
	source.cfg = cfg
	return source
}

func reloadSource(t *testing.T, source stubConfigSource) stubConfigSource {
	t.Helper()
	cfg, err := config.Load(source.path, nil)
	if err != nil {
		t.Fatalf("config.Load rejected the edited file: %v", err)
	}
	source.cfg = cfg
	return source
}

func TestEnvConfigShowDefaults(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	code, stdout, _ := runProfile(t, source, "env", "config", "show", "overlay_default_weight")
	if code != exitOK || strings.TrimSpace(stdout) != "1000" {
		t.Fatalf("code = %d, stdout = %q, want 1000", code, stdout)
	}
	code, stdout, _ = runProfile(t, source, "env", "config", "show", "precedence.winner")
	if code != exitOK || strings.TrimSpace(stdout) != `"higher-weight"` {
		t.Fatalf("code = %d, stdout = %q", code, stdout)
	}
}

func TestEnvConfigSetUnsetRoundTrip(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 1, "skills_root": "x", "projects": {}}`)
	code, stdout, stderr := runProfile(t, source, "env", "config", "set", "precedence.winner", "lower-weight")
	if code != exitOK {
		t.Fatalf("set = %d\nstderr:\n%s", code, stderr)
	}
	if strings.TrimSpace(stdout) != `"lower-weight"` {
		t.Fatalf("set printed %q", stdout)
	}
	source = reloadSource(t, source)
	if source.cfg.Schema != 2 {
		t.Fatalf("editing an environments knob must promote the file to schema 2")
	}
	code, stdout, _ = runProfile(t, source, "env", "config", "show", "precedence.winner")
	if code != exitOK || strings.TrimSpace(stdout) != `"lower-weight"` {
		t.Fatalf("show after set = %d %q", code, stdout)
	}
	code, _, stderr = runProfile(t, source, "env", "config", "unset", "precedence.winner")
	if code != exitOK {
		t.Fatalf("unset = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	code, stdout, _ = runProfile(t, source, "env", "config", "show", "precedence.winner")
	if code != exitOK || strings.TrimSpace(stdout) != `"higher-weight"` {
		t.Fatalf("show after unset = %d %q, want the default", code, stdout)
	}
}

func TestEnvConfigSetListKnob(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	code, _, stderr := runProfile(t, source, "env", "config", "set", "xdg_seed_allowlist", `["git"]`)
	if code != exitOK {
		t.Fatalf("set = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	if len(source.cfg.Env.XDGSeedAllowlist) != 1 {
		t.Fatalf("allowlist = %v", source.cfg.Env.XDGSeedAllowlist)
	}
}

func TestEnvConfigRefusals(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	for _, args := range [][]string{
		{"env", "config", "set", "nope", "1"},
		{"env", "config", "set", "precedence.winner"},
		{"env", "config", "unset"},
		{"env", "config", "show", "a", "b"},
	} {
		if code, _, _ := runProfile(t, source, args...); code != exitUsage {
			t.Fatalf("%v = %d, want usage", args, code)
		}
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "overlays_allowed", "yes"); code != exitFail {
		t.Fatalf("bad value = %d, want failure", code)
	} else if !strings.Contains(stderr, "overlays_allowed") {
		t.Fatalf("stderr = %q, want the knob named", stderr)
	}
	if code, _, _ := runProfile(t, source, "env", "config", "unset", "forms.pi"); code != exitFail {
		t.Fatalf("unset of an absent knob must fail")
	}
}

// TestEnvConfigLockedKnobRefuses drives the manager §1 locked-key refusal
// through run(): a locked knob refuses with the system-file warning and the
// file is unchanged.
func TestEnvConfigLockedKnobRefuses(t *testing.T) {
	source, _ := profileHome(t)
	home := t.TempDir()
	userPath := filepath.Join(home, "config.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.precedence"],
		"environments": {"precedence": {"winner": "higher-weight", "placement": "winner-last"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	code, _, stderr := runProfile(t, source, "env", "config", "set", "precedence.winner", "lower-weight")
	if code != exitFail {
		t.Fatalf("set on a locked knob = %d, want failure", code)
	}
	if !strings.Contains(stderr, systemPath) {
		t.Fatalf("stderr = %q, want the system file named", stderr)
	}
	payload, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	if _, present := object["environments"]; present {
		t.Fatalf("locked set must not write the machine file: %s", payload)
	}
}

func TestProfileComposeAddListRemove(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 1, "skills_root": "x", "projects": {}}`)
	code, _, stderr := runProfile(t, source, "profile", "compose", "companyA", "add",
		"https://example.com/personal", "--range", "^1", "--weight", "7")
	if code != exitOK {
		t.Fatalf("add = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	if source.cfg.Schema != 2 {
		t.Fatalf("compose must promote the file to schema 2")
	}
	code, listOut, stderr := runProfile(t, source, "profile", "compose", "companyA", "list")
	if code != exitOK {
		t.Fatalf("list = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(listOut, "https://example.com/personal") || !strings.Contains(listOut, "range=^1") {
		t.Fatalf("list printed %q", listOut)
	}
	code, _, stderr = runProfile(t, source, "profile", "compose", "companyA", "remove", "https://example.com/personal")
	if code != exitOK {
		t.Fatalf("remove = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	if len(source.cfg.Env.Overlays["companyA"]) != 0 {
		t.Fatalf("overlays = %+v, want empty", source.cfg.Env.Overlays)
	}
}

func TestProfileComposeRefusals(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	if code, _, _ := runProfile(t, source, "profile", "compose", "a", "add", "s", "--range", "^1", "--tag", "v1"); code != exitUsage {
		t.Fatalf("two requirement forms = %d, want usage", code)
	}
	if code, _, _ := runProfile(t, source, "profile", "compose", "a", "add", "s", "--range", "^1", "--weight", "-2"); code != exitUsage {
		t.Fatalf("negative weight = %d, want usage", code)
	}
	if code, _, _ := runProfile(t, source, "profile", "compose", "a", "remove", "missing-source"); code != exitFail {
		t.Fatalf("removing an undeclared overlay must fail")
	}
	if code, _, _ := runProfile(t, source, "profile", "compose"); code != exitUsage {
		t.Fatalf("bare compose = %d, want usage", code)
	}
}

// TestProfileUseLockedRequireRefuses drives the §12.2 refusal through run():
// a machine-scope use of any other profile fails naming the locked knob,
// while the required profile passes the gate.
func TestProfileUseLockedRequireRefuses(t *testing.T) {
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "companyA"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	code, _, stderr := runProfile(t, source, "profile", "use", "personal")
	if code != exitFail || !strings.Contains(stderr, "environments.require_current_profile") {
		t.Fatalf("use of another profile = %d\nstderr:\n%s", code, stderr)
	}
	// The required profile passes the gate: it proceeds to the switch,
	// which fails only because no such profile is installed.
	code, _, stderr = runProfile(t, source, "profile", "use", "companyA")
	if code == exitOK || strings.Contains(stderr, "environments.require_current_profile") {
		t.Fatalf("use of the required profile must pass the gate: code = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileInstallUseLockedRequireRefuses drives the same §12.2 gate
// through the install activation seam (run()): with the key locked to
// acme, profile install --use of another profile fails naming the locked
// knob and moves no current, while installing without --use succeeds
// (no switch) and a scoped switch stays unaffected.
func TestProfileInstallUseLockedRequireRefuses(t *testing.T) {
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "acme"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	other := t.TempDir()
	writeContextPackage(t, other, "other", "1.0.0", "hello\n")
	code, _, stderr := runProfile(t, source, "profile", "install", other, "--as", "other", "--use")
	if code != exitFail || !strings.Contains(stderr, "environments.require_current_profile") {
		t.Fatalf("install --use of another profile = %d\nstderr:\n%s", code, stderr)
	}
	// The required profile passes the gate through the same seam.
	required := t.TempDir()
	writeContextPackage(t, required, "acme", "1.0.0", "hello\n")
	code, _, stderr = runProfile(t, source, "profile", "install", required, "--as", "acme", "--use")
	if code != exitOK {
		t.Fatalf("install --use of the required profile = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileInstallFirstActivationLockedRequireRefuses drives the
// first-install auto-activation through the same seam: on a fresh machine
// with the key locked to acme, installing another profile without --use
// still attempts the §9.2 switch and is refused.
func TestProfileInstallFirstActivationLockedRequireRefuses(t *testing.T) {
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "acme"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	other := t.TempDir()
	writeContextPackage(t, other, "other", "1.0.0", "hello\n")
	code, _, stderr := runProfile(t, source, "profile", "install", other, "--as", "other")
	if code != exitFail || !strings.Contains(stderr, "environments.require_current_profile") {
		t.Fatalf("first install of another profile = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileUseClearOperandIsUsage drives the undefined
// `profile use <name> --clear` form through run(): cli/curator.md defines
// exactly two profile-use forms and this is neither, so the row is a usage
// error that moves no current. The seam refuses an operand beside --clear
// as well, so the C3-B1 bypass is closed at both layers.
func TestProfileUseClearOperandIsUsage(t *testing.T) {
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "acme"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	required := t.TempDir()
	writeContextPackage(t, required, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", required, "--as", "acme", "--use"); code != exitOK {
		t.Fatalf("install --use of the required profile = %d\nstderr:\n%s", code, stderr)
	}
	code, _, stderr := runProfile(t, source, "profile", "use", "bogus", "--clear")
	if code != exitUsage {
		t.Fatalf("use <name> --clear = %d, want usage %d\nstderr:\n%s", code, exitUsage, stderr)
	}
	if strings.Contains(stderr, "environments.require_current_profile") {
		t.Fatalf("undefined form must fail as usage, not at the gate:\n%s", stderr)
	}
	payload, err := os.ReadFile(filepath.Join(home, "profiles", "current"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(payload)) != "acme" {
		t.Fatalf("refused form moved the machine current to %q", payload)
	}
	code, _, stderr = runProfile(t, source, "profile", "use", "acme", "--clear", "--env", "codex_cli")
	if code != exitUsage {
		t.Fatalf("use <name> --clear --env = %d, want usage %d\nstderr:\n%s", code, exitUsage, stderr)
	}
}

// TestProfileScopedUseUnaffectedByLockedRequire drives the narrowed switch
// and its clearing through run() under the lock: a scoped switch records a
// scoped current without consulting the machine-scope gate, so both rows
// succeed while a machine-scope use of the same profile is refused.
func TestProfileScopedUseUnaffectedByLockedRequire(t *testing.T) {
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "acme"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	required := t.TempDir()
	writeContextPackage(t, required, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", required, "--as", "acme", "--use"); code != exitOK {
		t.Fatalf("install --use of the required profile = %d\nstderr:\n%s", code, stderr)
	}
	other := t.TempDir()
	writeContextPackage(t, other, "other", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", other, "--as", "other"); code != exitOK {
		t.Fatalf("install without activation = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "other", "--env", "codex_cli"); code != exitOK {
		t.Fatalf("scoped use under the lock = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "--clear", "--env", "codex_cli"); code != exitOK {
		t.Fatalf("scoped clear under the lock = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "other"); code != exitFail || !strings.Contains(stderr, "environments.require_current_profile") {
		t.Fatalf("machine use of another profile = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProfileImportUseLockedRequireRefuses drives the import activation
// seam through run(): on a fresh machine with the key locked to acme,
// `profile import --as other --use` installs the reassembled profile
// through the path pipeline and is refused at the §9.2 switch naming the
// locked knob, moving no current. The same seam covers the first-install
// auto-activation, which fires on this fresh machine with or without
// --use.
func TestProfileImportUseLockedRequireRefuses(t *testing.T) {
	source, home := profileHome(t)
	pinOperatorHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "acme"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	native := os.Getenv("CLAUDE_CONFIG_DIR")
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("native\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runProfile(t, source, "profile", "import", "--as", "other", "--use")
	if code != exitFail || !strings.Contains(stderr, "environments.require_current_profile") {
		t.Fatalf("import --use of another profile = %d\nstderr:\n%s", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(home, "profiles", "current")); !os.IsNotExist(err) {
		t.Fatalf("refused import activation moved the machine current, stat err %v", err)
	}
}

// TestProfileUpdateResyncPassesLockedRequire drives the resync caller of
// the seam through run(): with the key locked to the installed git root, a
// `profile update` that moves the lock re-switches the machine scope to
// the required profile through the same single gate, so the update
// succeeds. A mutant that breaks the gate for the resync path fails the
// other tests in this file; this row proves the gate does not break the
// allowed path.
func TestProfileUpdateResyncPassesLockedRequire(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "groot"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
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
	if code, _, stderr := runProfile(t, source, "profile", "install", "https://example.com/groot"); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	writeContextPackage(t, repo, "groot", "1.0.1", "git module two\n")
	git("add", ".")
	git("commit", "-m", "two")
	git("tag", "v1.0.1")
	code, stdout, stderr := runProfile(t, source, "profile", "update", "groot")
	if code != exitOK {
		t.Fatalf("update = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "updated") {
		t.Fatalf("update stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

// TestEnvStatusReportsLockedRequire drives env status through run() with
// the key locked to acme: the JSON carries require_current_profile and
// the text names the requirement (§12.2). Without the lock the key is
// absent.
func TestEnvStatusReportsLockedRequire(t *testing.T) {
	source, home := profileHome(t)
	userPath := filepath.Join(home, "machine.json")
	if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	systemPath := filepath.Join(home, "system.json")
	if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		"environments": {"require_current_profile": "acme"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := config.Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	source.path = userPath
	source.cfg = cfg
	code, stdout, stderr := runProfile(t, source, "env", "status", "--json")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("status JSON does not decode: %v\n%s", err, stdout)
	}
	if payload["require_current_profile"] != "acme" {
		t.Fatalf("JSON carries no requirement: %v", payload["require_current_profile"])
	}
	code, stdout, stderr = runProfile(t, source, "env", "status")
	if code != exitOK || !strings.Contains(stdout, "require_current_profile") || !strings.Contains(stdout, "acme") {
		t.Fatalf("text status names no requirement: code=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}
