package config

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Old machine schemas preserve default3; the new carrier accepts exactly the
// closed level object. Lexically non-integer JSON and unknown members refuse.
func TestStartupVerificationConfiguration(t *testing.T) {
	for _, schema := range []int{1, 2, 3, 5, 7, StartupVerificationSchemaVersion} {
		cfg, err := Parse(map[string]any{"schema_version": schema, "skills_root": "fixture", "projects": map[string]any{}}, "fixture.json")
		if err != nil || cfg.StartupVerification.Level != 3 || cfg.StartupProjection().Origin != "default" {
			t.Fatalf("schema %d: %+v,%v", schema, cfg, err)
		}
	}
	for _, raw := range []any{nil, "3", true, json.Number("3.0"), json.Number("3e0"), json.Number("0"), json.Number("5"), map[string]any{}, map[string]any{"level": json.Number("3"), "extra": true}} {
		value := raw
		if _, ok := raw.(map[string]any); !ok && raw != nil {
			value = map[string]any{"level": raw}
		}
		_, err := Parse(map[string]any{"schema_version": StartupVerificationSchemaVersion, "skills_root": "fixture", "projects": map[string]any{}, "startup_verification": value}, "fixture.json")
		if err == nil || !strings.Contains(err.Error(), "startup_verification_invalid") {
			t.Fatalf("accepted invalid level object %#v: %v", value, err)
		}
	}
	for _, level := range []string{"1", "2", "3", "4"} {
		cfg, err := Parse(map[string]any{"schema_version": StartupVerificationSchemaVersion, "skills_root": "fixture", "projects": map[string]any{}, "startup_verification": map[string]any{"level": json.Number(level)}}, "fixture.json")
		if err != nil || cfg.StartupProjection().Origin != "machine" {
			t.Fatalf("level %s: %+v,%v", level, cfg, err)
		}
	}
}

// The centre locator is metadata only and has a positive HTTPS authority/path
// grammar. Legacy launcher projections never acquire a new grant-service key.
func TestGrantServiceEndpoint(t *testing.T) {
	for _, endpoint := range []string{"https://example.org", "https://example.org:443/verify/v1", "https://127.0.0.1/verify", "https://[::1]:8443/verify", "https://example.org/%E2%9C%93"} {
		cfg, err := Parse(grantServiceConfig(endpoint), "fixture.json")
		if err != nil || cfg.StartupProjection().GrantService.Endpoint != endpoint {
			t.Fatalf("endpoint %q: %+v,%v", endpoint, cfg, err)
		}
		payload, _ := json.Marshal(cfg.LauncherProjection())
		if strings.Contains(string(payload), "grant_service") {
			t.Fatal("changed closed launcher-config/1")
		}
	}
	for _, endpoint := range []string{"http://example.org", "/verify", "https://user@example.org", "https://example.org?", "https://example.org?q=1", "https://example.org#", "https://example.org/#fragment", "https://example.org/../verify", "https://example.org/%2e%2e/verify", "https://example.org/\\verify", "https://example.org/%5cverify", "https://example.org/%00", "https://example.org:0", "https://example.org:65536", "https://example.org:", "https://::1:443/verify", "https://bad_host/", "https://-bad.example/", "https://example.org/%zz", "https://example.org/ verify"} {
		if cfg, err := Parse(grantServiceConfig(endpoint), "fixture.json"); err == nil || cfg != nil {
			t.Fatalf("accepted forbidden endpoint %q", endpoint)
		}
	}
}

func grantServiceConfig(endpoint string) map[string]any {
	return map[string]any{"schema_version": StartupVerificationSchemaVersion, "skills_root": "fixture", "projects": map[string]any{}, "extensions": map[string]any{"grant_service": map[string]any{"endpoint": endpoint}}}
}

// A locked system baseline overrides a lower machine level without rewriting
// either file; older carrier files cannot carry the new setting themselves.
func TestStartupSystemBaseline(t *testing.T) {
	cfg, _, err := loadWithSystem(t, `{"schema_version":7,"skills_root":"fixture","projects":{}}`, `{"schema_version":11,"startup_verification":{"level":4},"locked":["startup_verification"]}`)
	if err != nil || cfg.StartupProjection().Level != 4 || cfg.StartupProjection().Origin != "system" {
		t.Fatalf("system baseline: %+v,%v", cfg, err)
	}
	before, _ := os.ReadFile(cfg.Path)
	policyBefore, _ := os.ReadFile(cfg.SystemConfigPath)
	if err := SetStartupVerification(context.Background(), cfg.Path, 1); err == nil {
		t.Fatal("machine edit weakened locked system baseline")
	}
	after, _ := os.ReadFile(cfg.Path)
	policyAfter, _ := os.ReadFile(cfg.SystemConfigPath)
	if !bytes.Equal(before, after) || !bytes.Equal(policyBefore, policyAfter) {
		t.Fatal("locked edit changed machine or system policy")
	}
	for _, schema := range []int{1, 2, 7} {
		if _, err := Parse(map[string]any{"schema_version": schema, "skills_root": "fixture", "projects": map[string]any{}, "startup_verification": map[string]any{"level": 3}}, "fixture.json"); err == nil {
			t.Fatalf("old carrier %d accepts new field", schema)
		}
	}
}
