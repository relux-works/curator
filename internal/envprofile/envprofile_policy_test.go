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
			"mcp_package_allowlist":   []any{"https://example.com/m"},
			"overlays_allowed":        false,
			"overlay_default_weight":  float64(500),
			"precedence":              map[string]any{"winner": "lower-weight", "placement": "winner-first"},
			"require_current_profile": "acme",
			"overlays": map[string]any{
				"a": []any{map[string]any{"source": "https://example.com/p", "revision": strings.Repeat("ab", 20), "weight": float64(7)}},
			},
		},
	}, "config.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Locked["environments.require_current_profile"] = true
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
	if len(policy.Overlays["a"]) != 1 || policy.Overlays["a"][0].Source != "https://example.com/p" {
		t.Fatalf("overlays = %+v", policy.Overlays)
	}
	if policy.Overlays["a"][0].Weight == nil || *policy.Overlays["a"][0].Weight != 7 {
		t.Fatalf("overlay weight = %+v", policy.Overlays["a"][0].Weight)
	}
	if len(policy.EffectiveOverlays("a")) != 0 {
		t.Fatalf("forbidding policy must empty the carried list")
	}
	if policy.RequireCurrent == nil || *policy.RequireCurrent != "acme" || !policy.RequireCurrentLocked {
		t.Fatalf("require_current not carried: %+v locked=%v", policy.RequireCurrent, policy.RequireCurrentLocked)
	}
	if err := policy.CheckMachineUse("other"); err == nil || !strings.Contains(err.Error(), "environments.require_current_profile") {
		t.Fatalf("locked require must refuse another profile: %v", err)
	}
	// The builtin default profile is the first thing an operator under a
	// lock tries; it is refused like any other non-required name. A mutant
	// admitting exactly DefaultProfile must fail this row.
	if err := policy.CheckMachineUse(DefaultProfile); err == nil || !strings.Contains(err.Error(), "environments.require_current_profile") {
		t.Fatalf("locked require must refuse the builtin default profile: %v", err)
	}
	if err := policy.CheckMachineUse("acme"); err != nil {
		t.Fatalf("required profile refused: %v", err)
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

// TestMachineClearReachesSeamGate drives the machine-clear combination
// (clearScope with no scope, which the CLI rejects as usage) through the
// production UseWithPolicy seam: the single structural gate refuses when
// the recorded current is not the locked requirement. An operand beside
// --clear is refused without reaching the gate, a scoped clear is
// unaffected, and the required profile passes the gate (failing later only
// on the installed check). A mutant that exempts the clear path from the
// gate (`scope == "" && !clearScope`) succeeds on the machine clear and
// must fail this test.
func TestMachineClearReachesSeamGate(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	first := t.TempDir()
	writePackage(t, first, "one", "1.0.0", "one\n")
	second := t.TempDir()
	writePackage(t, second, "two", "1.0.0", "two\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: first}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: second}); err != nil {
		t.Fatal(err)
	}
	required := "acme"
	locked := Policy{OverlaysAllowed: true, RequireCurrent: &required, RequireCurrentLocked: true}
	if _, err := UseWithPolicy(home, "bogus", "", "", true, locked); err == nil || !strings.Contains(err.Error(), "takes no profile operand") {
		t.Fatalf("operand beside --clear = %v, want the operand refusal", err)
	}
	if _, err := UseWithPolicy(home, "", "", "", true, locked); err == nil || !strings.Contains(err.Error(), "environments.require_current_profile") {
		t.Fatalf("machine clear under the lock = %v, want the §12.2 refusal", err)
	}
	if current, _ := Current(home); current != "one" {
		t.Fatalf("refused switch moved the machine current to %q", current)
	}
	if _, err := UseWithPolicy(home, "", "codex_cli", "", true, locked); err != nil {
		t.Fatalf("scoped clear under the lock: %v", err)
	}
	if _, err := UseWithPolicy(home, "acme", "", "", false, locked); err == nil || strings.Contains(err.Error(), "environments.require_current_profile") {
		t.Fatalf("required profile must pass the gate: %v", err)
	} else if !strings.Contains(err.Error(), DiagNotFound) {
		t.Fatalf("required profile must fail only on the installed check: %v", err)
	}
}
