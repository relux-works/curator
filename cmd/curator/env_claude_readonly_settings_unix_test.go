//go:build unix

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
)

// The real CLI opt-in creates the mandatory regular 0400 settings file;
// resolve and status report writable drift without mutating it, and repair
// restores the invariant. This same positive test is the main-branch plant.
func TestEnvClaudeReadOnlySettings(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{}}`)
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatal(code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "claude_settings_mode.acme", "read-only"); code != exitOK {
		t.Fatal(code, stderr)
	}
	source = reloadSource(t, source)
	if code, _, stderr := runProfile(t, source, "env", "resolve", "claude_code", "--repair"); code != exitOK {
		t.Fatal(code, stderr)
	}
	path := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", "claude_code"), "settings.json")
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o400 {
		t.Fatalf("settings mode %v, want regular 0400", info.Mode())
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	diagnostic := "Claude settings settings.json has write bits (0600); claude_settings_mode=read-only requires no write bits"
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "claude_code")
	if code != exitFail || stdout != "" || !strings.Contains(stderr, diagnostic) {
		t.Fatalf("resolve: %d %s %s", code, stdout, stderr)
	}
	code, stdout, stderr = runProfile(t, source, "env", "status")
	if code != exitOK || !strings.Contains(stdout, "acme claude_code: non-current") || !strings.Contains(stdout+stderr, diagnostic) {
		t.Fatalf("status: %d %s %s", code, stdout, stderr)
	}
	info, err = os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatal("read-only commands changed mode", err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "claude_code", "--repair"); code != exitOK {
		t.Fatal(code, stderr)
	}
	info, err = os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o400 {
		t.Fatal("repair did not restore 0400", err)
	}
}

// Invalid values refuse with the exact field diagnostic and preserve the
// machine file; valid dotted profile names round-trip through show/unset.
func TestEnvConfigClaudeSettingsMode(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{}}`)
	before, err := os.ReadFile(source.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"writable", "true", "null", "17", "{}"} {
		code, _, stderr := runProfile(t, source, "env", "config", "set", "claude_settings_mode.team.prod", value)
		if code != exitFail || strings.TrimPrefix(stderr, "warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults\n") != "curator: environments.claude_settings_mode.team.prod: must be one of native, read-only\n" {
			t.Fatalf("%s: %d %q", value, code, stderr)
		}
		after, err := os.ReadFile(source.path)
		if err != nil || string(after) != string(before) {
			t.Fatal("invalid edit changed config", err)
		}
	}
	for _, value := range []string{"read-only", "native"} {
		if code, _, stderr := runProfile(t, source, "env", "config", "set", "claude_settings_mode.team.prod", value); code != exitOK {
			t.Fatal(code, stderr)
		}
		source = reloadSource(t, source)
		code, stdout, stderr := runProfile(t, source, "env", "config", "show", "claude_settings_mode.team.prod")
		if code != exitOK || strings.TrimSpace(stdout) != `"`+value+`"` {
			t.Fatal(code, stdout, stderr)
		}
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "unset", "claude_settings_mode.team.prod"); code != exitOK {
		t.Fatal(code, stderr)
	}
}
