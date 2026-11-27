package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadOrDefaultZeroConfigAndSystemEnforcement(t *testing.T) {
	home := t.TempDir()
	cfgPath := filepath.Join(home, "config.json")

	// Absent user config yields in-memory defaults without writing to disk.
	cfg, err := LoadOrDefault(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadOrDefault absent: %v", err)
	}
	if cfg.SkillsRoot != filepath.Join(home, "skills") {
		t.Fatalf("SkillsRoot = %q, want %q", cfg.SkillsRoot, filepath.Join(home, "skills"))
	}
	if cfg.EffectiveSecurityPosture() != SecurityPosturePermissive {
		t.Fatalf("posture = %q, want permissive", cfg.EffectiveSecurityPosture())
	}
	if _, statErr := os.Stat(cfgPath); !os.IsNotExist(statErr) {
		t.Fatalf("LoadOrDefault wrote to disk: %v", statErr)
	}

	// Present system config is still enforced over absent user config.
	sysPath := filepath.Join(t.TempDir(), "system-config.json")
	if err := os.WriteFile(sysPath, []byte(`{"schema_version":2,"security_posture":"hardened","locked":["security_posture"]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", sysPath)
	cfg, err = LoadOrDefault(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadOrDefault with system lock: %v", err)
	}
	if cfg.EffectiveSecurityPosture() != SecurityPostureHardened || !cfg.Locked["security_posture"] {
		t.Fatalf("system hardened lock not enforced: posture=%q locked=%v", cfg.EffectiveSecurityPosture(), cfg.Locked)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")

	// Malformed user config still fails closed (never falls back to defaults).
	if err := os.WriteFile(cfgPath, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadOrDefault(cfgPath, nil); err == nil {
		t.Fatal("malformed config unexpectedly fell back to defaults")
	}
}

func TestSchema2OptionalSkillsRootAndSchema1Strict(t *testing.T) {
	home := t.TempDir()
	cfgPath := filepath.Join(home, "config.json")

	// Schema 2 without skills_root defaults to <home>/skills.
	cfg, err := Parse(map[string]any{"schema_version": float64(2), "projects": map[string]any{}}, cfgPath)
	if err != nil {
		t.Fatalf("schema 2 without skills_root failed: %v", err)
	}
	if cfg.SkillsRoot != filepath.Join(home, "skills") {
		t.Fatalf("schema 2 default SkillsRoot = %q", cfg.SkillsRoot)
	}

	// Schema 2 with explicit empty skills_root still fails closed.
	if _, err := Parse(map[string]any{"schema_version": float64(2), "skills_root": "", "projects": map[string]any{}}, cfgPath); err == nil {
		t.Fatal("schema 2 with explicit empty skills_root should fail")
	}

	// Frozen schema 1 still requires skills_root.
	if _, err := Parse(map[string]any{"schema_version": float64(1), "projects": map[string]any{}}, cfgPath); err == nil || !strings.Contains(err.Error(), "skills_root") {
		t.Fatalf("schema 1 without skills_root = %v, want skills_root error", err)
	}
}

func TestBootstrapDefaultsSkillsRoot(t *testing.T) {
	home := t.TempDir()
	cfgPath := filepath.Join(home, "config.json")
	if err := Bootstrap(cfgPath, "", "", nil, false); err != nil {
		t.Fatalf("Bootstrap with empty skillsRoot: %v", err)
	}
	cfg, err := Load(cfgPath, nil)
	if err != nil {
		t.Fatalf("Load bootstrapped config: %v", err)
	}
	if cfg.SkillsRoot != filepath.Join(home, "skills") {
		t.Fatalf("bootstrapped SkillsRoot = %q, want %q", cfg.SkillsRoot, filepath.Join(home, "skills"))
	}
}
