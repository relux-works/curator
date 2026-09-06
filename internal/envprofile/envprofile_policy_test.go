package envprofile

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextmaterialize"
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
			"mcp_package_allowlist":  []any{"https://example.com/m"},
			"overlays_allowed":       false,
			"overlay_default_weight": float64(500),
			"precedence":             map[string]any{"winner": "lower-weight", "placement": "winner-first"},
			"overlays": map[string]any{
				"a": []any{map[string]any{"source": "/srv/p", "revision": strings.Repeat("ab", 20), "weight": float64(7)}},
			},
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
	if policy.OverlayDefaultWeight != 500 {
		t.Fatalf("overlay default weight = %d, want 500", policy.OverlayDefaultWeight)
	}
	if policy.Precedence() != (contextmaterialize.Precedence{Winner: "lower-weight", Placement: "winner-first"}) {
		t.Fatalf("precedence = %+v", policy.Precedence())
	}
	// The declarations are carried even while the forbidding policy
	// empties them at resolution time.
	if len(policy.Overlays["a"]) != 1 || policy.Overlays["a"][0].Source != "/srv/p" {
		t.Fatalf("overlays = %+v", policy.Overlays)
	}
	if policy.Overlays["a"][0].Weight == nil || *policy.Overlays["a"][0].Weight != 7 {
		t.Fatalf("overlay weight = %+v", policy.Overlays["a"][0].Weight)
	}
	if len(policy.EffectiveOverlays("a")) != 0 {
		t.Fatalf("forbidding policy must empty the carried list")
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
