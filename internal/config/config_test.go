package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, dir, name, text string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const minimal = `{"schema_version": 1, "skills_root": "/tmp/skills", "projects": {}}`

func TestParseMinimalDefaults(t *testing.T) {
	cfg, err := Load(writeConfig(t, t.TempDir(), "config.json", minimal), nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AdapterMode != "auto" || cfg.WorktreeAliasPattern != DefaultWorktreeAliasPattern {
		t.Fatalf("defaults: %+v", cfg)
	}
	if len(cfg.DefaultAgents) != 1 || cfg.DefaultAgents[0] != "codex_cli" {
		t.Fatalf("default agents: %v", cfg.DefaultAgents)
	}
	if cfg.Audit.Enabled || cfg.Audit.Mode != "advisory" || cfg.Audit.RegistryPolicy != "advisory" {
		t.Fatalf("audit defaults: %+v", cfg.Audit)
	}
	if cfg.Audit.MaxRequestBytes != DefaultMaxRequestBytes ||
		cfg.Audit.SnapshotMaxAgeSeconds != DefaultSnapshotMaxAgeSeconds ||
		cfg.Audit.SnapshotClockSkewSeconds != DefaultSnapshotClockSkewSeconds ||
		cfg.Audit.CacheTTLSeconds != DefaultCacheTTLSeconds ||
		cfg.Audit.OfflineGraceSeconds != DefaultOfflineGraceSeconds {
		t.Fatalf("audit timing defaults: %+v", cfg.Audit)
	}
}

func TestParseProjectsAndRegistries(t *testing.T) {
	text := `{
		"schema_version": 1, "skills_root": "~/skills",
		"adapter_mode": "copy",
		"projects": {"app": {"path": "~/dev/app", "agents": ["claude_code"], "project_alias": "application"}},
		"allowed_sources": ["git.example.com/skills"],
		"audit_registries": [
			{"name": "corp", "url": "https://registry.example.com", "public_keys": ["ed25519:` + strings.Repeat("A", 43) + `="]},
			{"name": "off", "url": "https://off.example.com", "enabled": false}
		],
		"audit": {"enabled": true, "mode": "strict", "registry_policy": "strict",
			"max_request_bytes": 2048, "snapshot_max_age_seconds": 86400,
			"snapshot_clock_skew_seconds": 0, "cache_ttl_seconds": 0, "offline_grace_seconds": 0,
			"revocations": ["sha256:` + strings.Repeat("a", 64) + `", "source:git.example.com/bad/*"]}
	}`
	cfg, err := Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
	if err != nil {
		t.Fatal(err)
	}
	project := cfg.Projects["app"]
	if project.ProjectAlias != "application" || project.CheckoutAlias != "app" {
		t.Fatalf("project aliases: %+v", project)
	}
	trusted := cfg.TrustedRegistries()
	if len(trusted) != 1 || trusted[0].Name != "corp" {
		t.Fatalf("trusted registries: %+v", trusted)
	}
	if !cfg.Audit.Enabled || cfg.Audit.Mode != "strict" {
		t.Fatalf("audit: %+v", cfg.Audit)
	}
	if cfg.Audit.MaxRequestBytes != 2048 || cfg.Audit.SnapshotMaxAgeSeconds != 86400 ||
		cfg.Audit.SnapshotClockSkewSeconds != 0 || cfg.Audit.CacheTTLSeconds != 0 ||
		cfg.Audit.OfflineGraceSeconds != 0 {
		t.Fatalf("audit policy values: %+v", cfg.Audit)
	}
}

func TestParseRejections(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{"schema", `{"schema_version": 2, "skills_root": "x", "projects": {}}`, "schema_version"},
		{"skills_root", `{"schema_version": 1, "projects": {}}`, "skills_root"},
		{"projects", `{"schema_version": 1, "skills_root": "x"}`, "projects"},
		{"adapter", `{"schema_version": 1, "skills_root": "x", "projects": {}, "adapter_mode": "hardlink"}`, "adapter_mode"},
		{"unknown top", `{"schema_version": 1, "skills_root": "x", "projects": {}, "typo": true}`, "unsupported"},
		{"project key", `{"schema_version": 1, "skills_root": "x", "projects": {"-app": {"path": "x"}}}`, "portable identifier"},
		{"project unknown", `{"schema_version": 1, "skills_root": "x", "projects": {"app": {"path": "x", "typo": true}}}`, "unsupported"},
		{"project alias", `{"schema_version": 1, "skills_root": "x", "projects": {"app": {"path": "x", "project_alias": "App Label"}}}`, "portable identifier"},
		{"duplicate agents", `{"schema_version": 1, "skills_root": "x", "projects": {}, "default_agents": ["codex_cli", "codex_cli"]}`, "unique"},
		{"empty locale", `{"schema_version": 1, "skills_root": "x", "projects": {}, "preferred_locale": ""}`, "preferred_locale"},
		{"regex", `{"schema_version": 1, "skills_root": "x", "projects": {}, "worktree_alias_pattern": "["}`, "worktree_alias_pattern"},
		{"reg url", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "ftp://x"}]}`, "audit_registries[0]"},
		{"reg insecure remote", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "http://registry.example.com"}]}`, "loopback"},
		{"reg credentials", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://user:secret@registry.example.com"}]}`, "credentials"},
		{"reg percent path", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://registry.example.com/a%2Fb"}]}`, "percent"},
		{"reg dot path", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://registry.example.com/a/../b"}]}`, "dot segments"},
		{"reg doubled path", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://registry.example.com/a//b"}]}`, "doubled"},
		{"reg unicode host", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://réregistry.example/a"}]}`, "ASCII"},
		{"reg unicode path", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://registry.example/é"}]}`, "ASCII"},
		{"reg bad host", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://bad..example"}]}`, "ASCII"},
		{"reg unknown", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://x", "required": true}]}`, "unsupported"},
		{"reg key", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "n", "url": "https://x", "public_keys": ["ed25519:bad"]}]}`, "Ed25519"},
		{"reg dup", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "a", "url": "https://x"}, {"name": "b", "url": "https://x"}]}`, "audit_registries[1]"},
		{"reg canonical dup", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit_registries": [{"name": "a", "url": "https://EXAMPLE.com:443/"}, {"name": "b", "url": "https://example.com"}]}`, "duplicates url"},
		{"audit mode", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"mode": "paranoid"}}`, "audit.mode"},
		{"audit unknown", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"level": 1}}`, "audit"},
		{"revocation", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"revocations": ["nope"]}}`, "audit.revocations[0]"},
		{"policy class", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"source_policy": {"default_class": "secret"}}}`, "audit.source_policy.default_class"},
		{"registry_policy", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"registry_policy": "always"}}`, "audit.registry_policy"},
		{"cache ttl", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"cache_ttl_seconds": -1}}`, "audit.cache_ttl_seconds"},
		{"max request", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"max_request_bytes": 10485761}}`, "audit.max_request_bytes"},
		{"source policy unknown", `{"schema_version": 1, "skills_root": "x", "projects": {}, "audit": {"source_policy": {"classification": "public"}}}`, "unsupported"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, t.TempDir(), "config.json", tc.text), nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want mention of %q", err, tc.want)
			}
		})
	}
}

func TestRegistryURLCanonicalization(t *testing.T) {
	for input, want := range map[string]string{
		"https://EXAMPLE.com:443/v1/": "https://example.com/v1",
		"HTTPS://EXAMPLE.com/v1/":     "https://example.com/v1",
		"http://localhost:80/":        "http://localhost",
		"http://127.0.0.1:8080/":      "http://127.0.0.1:8080",
		"http://[::1]:8080/":          "http://[::1]:8080",
		"http://[0:0:0:0:0:0:0:1]/":   "http://[::1]",
	} {
		got, err := canonicalRegistryURL(input)
		if err != nil || got != want {
			t.Errorf("canonicalRegistryURL(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}

func TestSystemConfigLockedKeys(t *testing.T) {
	dir := t.TempDir()
	userPath := writeConfig(t, dir, "config.json", `{
		"schema_version": 1, "skills_root": "/s", "projects": {},
		"allowed_sources": ["user.example.com/x"],
		"preferred_locale": "en"
	}`)
	systemPath := writeConfig(t, dir, "system.json", `{
		"schema_version": 1, "locked": ["allowed_sources"],
		"allowed_sources": ["org.example.com/skills"],
		"preferred_locale": "ru"
	}`)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)

	var warnings []string
	cfg, err := Load(userPath, func(message string) { warnings = append(warnings, message) })
	if err != nil {
		t.Fatal(err)
	}
	// locked key wins with a warning
	if len(cfg.AllowedSources) != 1 || cfg.AllowedSources[0] != "org.example.com/skills" {
		t.Fatalf("locked key did not win: %v", cfg.AllowedSources)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "allowed_sources") {
		t.Fatalf("warnings: %v", warnings)
	}
	// unlocked system key is only a default: user value stays
	if cfg.PreferredLocale != "en" {
		t.Fatalf("unlocked default overrode user value: %q", cfg.PreferredLocale)
	}
}

func TestSystemConfigUnlockedDefaultApplies(t *testing.T) {
	dir := t.TempDir()
	userPath := writeConfig(t, dir, "config.json", minimal)
	systemPath := writeConfig(t, dir, "system.json", `{"schema_version": 1, "locked": [], "preferred_locale": "ru"}`)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PreferredLocale != "ru" {
		t.Fatalf("system default not applied: %q", cfg.PreferredLocale)
	}
}

func TestSystemConfigLockedButUnsetFails(t *testing.T) {
	dir := t.TempDir()
	userPath := writeConfig(t, dir, "config.json", minimal)
	systemPath := writeConfig(t, dir, "system.json", `{"schema_version": 1, "locked": ["audit"]}`)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	if _, err := Load(userPath, nil); err == nil || !strings.Contains(err.Error(), "locks") {
		t.Fatalf("err = %v, want locked-but-unset error", err)
	}
}

func TestMalformedSystemConfigFailsClosed(t *testing.T) {
	cases := []string{
		`{"locked": []}`,
		`{"schema_version": 2, "locked": []}`,
		`{"schema_version": 1, "locked": ["projects"]}`,
		`{"schema_version": 1, "locked": ["audit", "audit"], "audit": {}}`,
		`{"schema_version": 1, "typo": true}`,
	}
	for _, body := range cases {
		dir := t.TempDir()
		userPath := writeConfig(t, dir, "config.json", minimal)
		systemPath := writeConfig(t, dir, "system.json", body)
		t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
		if _, err := Load(userPath, nil); err == nil {
			t.Errorf("system config %s was accepted", body)
		}
	}
}

func TestMissingConfigError(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "absent.json"), nil)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err = %v", err)
	}
}
