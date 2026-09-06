package envprofile

import (
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// TestPolicyFromConfigCarriesEnvGates drives the production PolicyFromConfig
// entry point: the MCP package allowlist and the overlays_allowed composition
// policy come from the effective schema-2 configuration, and the absent
// configuration keeps the §12.1 defaults (every identity permitted, overlays
// joined).
func TestPolicyFromConfigCarriesEnvGates(t *testing.T) {
	cfg, err := config.Parse(map[string]any{
		"schema_version": float64(2), "skills_root": "x", "projects": map[string]any{},
		"environments": map[string]any{
			"mcp_package_allowlist": []any{"https://example.com/m"},
			"overlays_allowed":      false,
		},
	}, "config.json")
	if err != nil {
		t.Fatal(err)
	}
	policy := PolicyFromConfig(cfg)
	if len(policy.MCPAllowlist) != 1 || policy.MCPAllowlist[0] != "https://example.com/m" {
		t.Fatalf("MCP allowlist = %v", policy.MCPAllowlist)
	}
	if !policy.ForbidsOverlays() {
		t.Fatalf("overlays_allowed=false must forbid overlays")
	}

	absent := PolicyFromConfig(nil)
	if len(absent.MCPAllowlist) != 0 || absent.ForbidsOverlays() {
		t.Fatalf("absent configuration must keep the defaults: %+v", absent)
	}

	schema1, err := config.Parse(map[string]any{
		"schema_version": float64(1), "skills_root": "x", "projects": map[string]any{},
	}, "config.json")
	if err != nil {
		t.Fatal(err)
	}
	defaults := PolicyFromConfig(schema1)
	if len(defaults.MCPAllowlist) != 0 || defaults.ForbidsOverlays() {
		t.Fatalf("schema-1 file must keep the defaults: %+v", defaults)
	}
}
