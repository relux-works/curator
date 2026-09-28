package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/envfragment"
)

type securityPostureVector struct {
	Name             string                     `json:"name"`
	RolloutRevision  string                     `json:"rollout_revision"`
	Machine          map[string]any             `json:"machine"`
	System           map[string]any             `json:"system"`
	ShippedRevisions SecurityPostureRevisions   `json:"shipped_revisions"`
	Operation        securityPostureOperation   `json:"operation"`
	Expected         securityPostureExpectation `json:"expected"`
}

type securityPostureOperation struct {
	Kind                   string `json:"kind"`
	MCPDeclarationsPresent bool   `json:"mcp_declarations_present"`
}

type securityPostureExpectation struct {
	Profile                 string                      `json:"profile"`
	ProfileSource           string                      `json:"profile_source"`
	Effective               map[string]any              `json:"effective"`
	Sources                 map[string]string           `json:"sources"`
	PassableEnvNullExplicit bool                        `json:"passable_env_null_explicit"`
	Diagnostics             []SecurityPostureDiagnostic `json:"diagnostics"`
	Outcome                 string                      `json:"outcome"`
	CuratorStatusRows       []SecurityPostureRow        `json:"curator_status_rows"`
	EnvStatusRows           json.RawMessage             `json:"env_status_rows"`
}

// TestSecurityPostureVectors drives every revision-A-owned vector through
// config.Load and the same diagnostics, outcome, and status-row APIs consumed
// by the manager. Registry reachability belongs to 1sapuy; revision-B default
// and flipped shipped-row cases belong to the flip leaf.
func TestSecurityPostureVectors(t *testing.T) {
	root := conformanceRoot(t)
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "security-posture.json")) // #nosec G304 -- pinned conformance input
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []securityPostureVector `json:"cases"`
	}
	if err := json.Unmarshal(payload, &corpus); err != nil {
		t.Fatal(err)
	}
	cases := corpus.Cases
	if len(cases) == 0 {
		t.Fatal("security-posture vectors publish no cases")
	}
	conformancecoverage.RunOutcomes(t, "security-posture/vectors", cases,
		func(tc securityPostureVector) string { return tc.Name }, func(caseT *testing.T, tc securityPostureVector) conformancecoverage.Observation {
			if reason := securityPostureBound(tc.Name); reason != "" {
				return conformancecoverage.Observation{BoundReason: reason}
			}
			cfg := loadSecurityPostureVector(t, tc)
			checkJSONEqual(t, "effective posture", tc.Expected.Effective, securityPostureEffective(cfg, tc.ShippedRevisions.EnvPassthrough))
			if got := cfg.EffectiveSecurityPosture(); got != tc.Expected.Profile {
				caseT.Fatalf("profile = %q, want %q", got, tc.Expected.Profile)
			}
			if cfg.SecurityPostureSource != tc.Expected.ProfileSource {
				caseT.Fatalf("profile source = %q, want %q", cfg.SecurityPostureSource, tc.Expected.ProfileSource)
			}
			checkJSONEqual(caseT, "posture sources", tc.Expected.Sources, securityPostureVectorSources(cfg))
			gotNullExplicit := cfg.Env.PassableEnvNamesSet && cfg.Env.PassableEnvNames == nil
			if gotNullExplicit != tc.Expected.PassableEnvNullExplicit {
				caseT.Fatalf("passable_env_names null explicit = %v, want %v", gotNullExplicit, tc.Expected.PassableEnvNullExplicit)
			}
			checkJSONEqual(caseT, "diagnostics", tc.Expected.Diagnostics, cfg.SecurityPostureDiagnostics(tc.Operation.Kind, tc.Operation.MCPDeclarationsPresent))
			if got := securityPostureVectorOutcome(cfg, tc.Operation); got != tc.Expected.Outcome {
				caseT.Fatalf("outcome = %q, want %q", got, tc.Expected.Outcome)
			}
			checkJSONEqual(caseT, "curator status rows", tc.Expected.CuratorStatusRows, cfg.SecurityPostureStatusRows(tc.ShippedRevisions))
			var wantEnvRows any
			if err := json.Unmarshal(tc.Expected.EnvStatusRows, &wantEnvRows); err != nil {
				caseT.Fatalf("decode expected env_status_rows: %v", err)
			}
			var gotEnvRows any
			if cfg.Schema != SchemaVersion {
				gotEnvRows = cfg.SecurityPostureStatusRows(tc.ShippedRevisions)
			}
			checkJSONEqual(caseT, "env status rows", wantEnvRows, gotEnvRows)
			return conformancecoverage.Observation{}
		})
}

func TestSecurityPostureSchema1MachineRemainsPermissiveWithSystemPosture(t *testing.T) {
	dir := t.TempDir()
	machine := map[string]any{
		"schema_version": SchemaVersion,
		"skills_root":    filepath.Join(dir, "skills"),
		"projects":       map[string]any{},
	}
	system := map[string]any{
		"schema_version":   SchemaVersion2,
		"security_posture": SecurityPostureHardened,
		"locked":           []any{"security_posture"},
	}
	machinePath := writeSecurityPostureJSON(t, dir, "machine.json", machine)
	systemPath := writeSecurityPostureJSON(t, dir, "system.json", system)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := Load(machinePath, nil)
	if err != nil {
		t.Fatalf("Load schema-1 machine with a system posture: %v", err)
	}
	if cfg.EffectiveSecurityPosture() != SecurityPosturePermissive || cfg.SecurityPostureSource != "profile" {
		t.Fatalf("schema-1 security posture = %q from %q, want permissive/profile", cfg.EffectiveSecurityPosture(), cfg.SecurityPostureSource)
	}
}

func TestSecurityPostureSystemValueMustBeLocked(t *testing.T) {
	dir := t.TempDir()
	machine := map[string]any{
		"schema_version":   SchemaVersion2,
		"skills_root":      filepath.Join(dir, "skills"),
		"projects":         map[string]any{},
		"security_posture": SecurityPosturePermissive,
	}
	system := map[string]any{
		"schema_version":   SchemaVersion2,
		"security_posture": SecurityPostureHardened,
	}
	machinePath := writeSecurityPostureJSON(t, dir, "machine.json", machine)
	systemPath := writeSecurityPostureJSON(t, dir, "system.json", system)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	if _, err := Load(machinePath, nil); err == nil || !strings.Contains(err.Error(), "must lock security_posture") {
		t.Fatalf("Load system posture without its required lock error = %v", err)
	}
}

func securityPostureBound(name string) string {
	switch name {
	case "unreachable-registry-permissive-warns", "unreachable-registry-hardened-refuses":
		return "install-time unreachable-registry behavior is owned by TASK-260910-1sapuy"
	case "revision-B-default-hardened-flip-install", "refusal-mcp-allowlist-empty-with-declarations", "hardened-contradiction-status-check-non-current", "posture-rows-flipped-revisions":
		return "revision-B posture default or shipped-row flip is owned by TASK-260927-25hk87"
	default:
		return ""
	}
}

func loadSecurityPostureVector(t *testing.T, tc securityPostureVector) *Config {
	t.Helper()
	dir := t.TempDir()
	machine := cloneStringAnyMap(tc.Machine)
	if _, ok := machine["skills_root"]; !ok {
		machine["skills_root"] = filepath.Join(dir, "skills")
	}
	if _, ok := machine["projects"]; !ok {
		machine["projects"] = map[string]any{}
	}
	system := cloneStringAnyMap(tc.System)
	if _, ok := system["schema_version"]; !ok {
		system["schema_version"] = SchemaVersion2
	}
	machinePath := writeSecurityPostureJSON(t, dir, "machine.json", machine)
	systemPath := writeSecurityPostureJSON(t, dir, "system.json", system)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := Load(machinePath, nil)
	if err != nil {
		t.Fatalf("Load vector config: %v", err)
	}
	return cfg
}

func cloneStringAnyMap(input map[string]any) map[string]any {
	cloned := make(map[string]any, len(input)+2)
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func writeSecurityPostureJSON(t *testing.T, dir, name string, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func securityPostureEffective(cfg *Config, s4 string) map[string]any {
	effective := map[string]any{
		"audit_mode":            cfg.Audit.Mode,
		"audit_registry_policy": cfg.Audit.RegistryPolicy,
		"allowed_sources":       append([]string{}, cfg.AllowedSources...),
	}
	if cfg.Schema == SchemaVersion {
		return effective
	}
	profile := envfragment.S4Warn
	if s4 == string(envfragment.S4Enforce) {
		profile = envfragment.S4Enforce
	}
	effective["mcp_package_allowlist"] = append([]string{}, cfg.Env.MCPPackageAllowlist...)
	effective["passable_env_names"] = envfragment.EffectivePassable(cfg.Env.PassableEnvNames, cfg.Env.PassableEnvNamesSet, profile)
	effective["transitive_system_modules"] = cfg.Env.TransitiveSystemModules
	effective["require_source_signers"] = cfg.Env.RequireSourceSigners
	return effective
}

func securityPostureVectorSources(cfg *Config) map[string]string {
	sources := map[string]string{
		"audit_mode":            cfg.PostureSourceFor("audit.mode"),
		"audit_registry_policy": cfg.PostureSourceFor("audit.registry_policy"),
		"allowed_sources":       cfg.PostureSourceFor("allowed_sources"),
	}
	if cfg.Schema != SchemaVersion {
		sources["mcp_package_allowlist"] = cfg.PostureSourceFor("environments.mcp_package_allowlist")
		sources["passable_env_names"] = cfg.PostureSourceFor("environments.passable_env_names")
		sources["transitive_system_modules"] = cfg.PostureSourceFor("environments.transitive_system_modules")
		sources["require_source_signers"] = cfg.PostureSourceFor("environments.require_source_signers")
	}
	return sources
}

func securityPostureVectorOutcome(cfg *Config, operation securityPostureOperation) string {
	switch operation.Kind {
	case "status":
		return "current"
	case "status-check":
		if cfg.SecurityPostureContradiction(operation.MCPDeclarationsPresent) {
			return "non-current"
		}
		return "current"
	case "install", "update", "profile-install", "profile-update", "launch-composition":
		if err := cfg.CheckSecurityPosture(operation.Kind, operation.MCPDeclarationsPresent); err != nil {
			return "refused"
		}
		return "proceeds"
	default:
		return fmt.Sprintf("unsupported operation %q", operation.Kind)
	}
}

func checkJSONEqual(t testing.TB, label string, want, got any) {
	t.Helper()
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal expected %s: %v", label, err)
	}
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal actual %s: %v", label, err)
	}
	if !bytes.Equal(wantJSON, gotJSON) {
		var wantPretty, gotPretty any
		_ = json.Unmarshal(wantJSON, &wantPretty)
		_ = json.Unmarshal(gotJSON, &gotPretty)
		if reflect.DeepEqual(wantPretty, gotPretty) {
			return
		}
		t.Fatalf("%s mismatch:\n got: %s\nwant: %s", label, gotJSON, wantJSON)
	}
}
