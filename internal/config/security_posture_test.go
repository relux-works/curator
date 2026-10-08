package config

import (
	"strings"
	"testing"
)

func TestSecurityPostureRevisionBDefaultsAndPermissiveCompatibility(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	for _, tc := range []struct {
		name, text, posture, source, auditMode, registryPolicy, transitive string
		requireSigners                                                     bool
	}{
		{"schema2-default", `{"schema_version":2,"skills_root":"x","projects":{}}`, "hardened", "profile", "strict", "strict", "error", true},
		{"schema2-empty-environments", `{"schema_version":2,"skills_root":"x","projects":{},"environments":{}}`, "hardened", "profile", "strict", "strict", "error", true},
		{"explicit-permissive", `{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{}}`, "permissive", "explicit", "advisory", "advisory", "drop", false},
		{"schema1-frozen", `{"schema_version":1,"skills_root":"x","projects":{}}`, "permissive", "profile", "advisory", "advisory", "drop", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadText(t, tc.text)
			if cfg.EffectiveSecurityPosture() != tc.posture || cfg.SecurityPostureSource != tc.source {
				t.Fatalf("posture = %s/%s, want %s/%s", cfg.EffectiveSecurityPosture(), cfg.SecurityPostureSource, tc.posture, tc.source)
			}
			if cfg.Audit.Mode != tc.auditMode || cfg.Audit.RegistryPolicy != tc.registryPolicy || cfg.Env.TransitiveSystemModules != tc.transitive || cfg.Env.RequireSourceSigners != tc.requireSigners {
				t.Fatalf("defaults = %s/%s/%s/%t, want %s/%s/%s/%t", cfg.Audit.Mode, cfg.Audit.RegistryPolicy, cfg.Env.TransitiveSystemModules, cfg.Env.RequireSourceSigners, tc.auditMode, tc.registryPolicy, tc.transitive, tc.requireSigners)
			}
			if len(cfg.AllowedSources) != 0 || len(cfg.Env.MCPPackageAllowlist) != 0 || cfg.Env.PassableEnvNamesSet || cfg.Env.PassableEnvNames != nil {
				t.Fatalf("allowlist or passthrough defaults changed: %+v", cfg.Env)
			}
			warning := cfg.SecurityPostureWarning()
			if (tc.posture == "permissive") != strings.Contains(warning, DiagSecurityPosturePermissive) {
				t.Fatalf("warning = %q for %s", warning, tc.posture)
			}
		})
	}
}

func TestSecurityPostureRevisionBExplicitKnobsStillWin(t *testing.T) {
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	cfg := loadText(t, `{"schema_version":2,"skills_root":"x","projects":{},"audit":{"mode":"advisory","registry_policy":"advisory"},"environments":{"transitive_system_modules":"drop","require_source_signers":false}}`)
	if !cfg.SecurityPostureHardened() || cfg.Audit.Mode != "advisory" || cfg.Audit.RegistryPolicy != "advisory" || cfg.Env.TransitiveSystemModules != "drop" || cfg.Env.RequireSourceSigners {
		t.Fatalf("explicit knobs lost to hardened defaults: %+v", cfg)
	}
	for _, knob := range []string{"audit.mode", "audit.registry_policy", "environments.transitive_system_modules", "environments.require_source_signers"} {
		if cfg.PostureSourceFor(knob) != "explicit" {
			t.Errorf("%s source = %s", knob, cfg.PostureSourceFor(knob))
		}
	}
}
