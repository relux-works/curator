package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// Load admits only canonical machine v5 selectors; failed validation never rewrites the file.
func TestCredentialStoreConfigGrammar(t *testing.T) {
	for _, tc := range []struct {
		name, member string
		valid        bool
	}{
		{"empty", `{}`, true},
		{"all selectors", `{"codex":{"codex_cli":"file","claude_code":"native","muse":"keychain","pi":"file","opencode":"native"}}`, true},
		{"null map", `null`, false},
		{"array map", `[]`, false},
		{"null profile", `{"p":null}`, false},
		{"invalid profile", `{"../p":{}}`, false},
		{"unknown id", `{"p":{"unknown":"file"}}`, false},
		{"alias id", `{"p":{"codex":"file"}}`, false},
		{"null selector", `{"p":{"codex_cli":null}}`, false},
		{"unknown selector", `{"p":{"codex_cli":"auto"}}`, false},
		{"wrong type", `{"p":{"codex_cli":true}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := fmt.Sprintf(`{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":%s}}`, tc.member)
			path := writeConfig(t, t.TempDir(), "config.json", text)
			cfg, err := Load(path, nil)
			if tc.valid {
				if err != nil {
					t.Fatal(err)
				}
				if cfg.Schema != 5 {
					t.Fatalf("schema = %d", cfg.Schema)
				}
			} else if err == nil || !strings.Contains(err.Error(), "environments.credential_store") {
				t.Fatalf("Load = %v, want credential_store refusal", err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != text {
				t.Fatalf("Load changed machine config: %v", err)
			}
		})
	}
}

// Frozen envelopes refuse even an empty selector map; v5 retains v3 waiver semantics.
func TestCredentialStoreConfigVersionBoundary(t *testing.T) {
	for _, version := range []int{1, 2, 3, 4} {
		want := "credential_store"
		if version == 1 || version == 4 {
			want = "schema_version"
		}
		loadFails(t, fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{},"environments":{"credential_store":{}}}`, version), want)
	}
	loadText(t, `{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":{},"secret_material_waivers":[{"pin":"`+strings.Repeat("a", 64)+`","hash_version":2,"file":"a.md","span":[1,2],"reason":"test"}]}}`)
	for _, version := range []int{1, 2, 3} {
		cfg := loadText(t, fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{}}`, version))
		env := cfg.EffectiveJSON()["environments"].(map[string]any)
		if _, exists := env["credential_store"]; exists {
			t.Fatal("old effective envelope acquired credential_store")
		}
	}
}

// Neither locked nor unlocked system config may select a store; isolation locks still bind v5.
func TestCredentialStoreConfigMachineOwnership(t *testing.T) {
	user := `{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":{"p":{"codex_cli":"file"}}}}`
	for _, locked := range []string{`[]`, `["environments.credential_store"]`} {
		_, _, err := loadWithSystem(t, user, `{"schema_version":2,"locked":`+locked+`,"environments":{"credential_store":{"p":{"codex_cli":"keychain"}}}}`)
		if err == nil || !strings.Contains(err.Error(), "credential_store") {
			t.Fatalf("system store accepted: %v", err)
		}
	}
	cfg, _, err := loadWithSystem(t, user, `{"schema_version":2,"locked":["environments.isolation"],"environments":{"isolation":{"p":{"codex_cli":"isolated"}}}}`)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Locked["environments.isolation"] || cfg.Env.Isolation["p"]["codex_cli"] != "isolated" {
		t.Fatal("isolation lock lost")
	}
	if EnvLockKey("credential_store.p.codex_cli") != "" {
		t.Fatal("store became lockable")
	}
}
