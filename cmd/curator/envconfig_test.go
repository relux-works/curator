package main

import (
	"encoding/json"
	"os"
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
