package audit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

// Identical effective capabilities hit; changing any capability input misses without reusing the allowance.
func TestVerdictCacheBindsEffectiveCapabilities(t *testing.T) {
	cfg := newCfg(t, "strict", "high")
	caps := capabilities.Manifest{
		Network: []string{"github.com"}, Exec: []string{"git"},
		Filesystem: capabilities.Filesystem{Keyword: "repo"},
		Secrets:    []string{"credential-name"}, EnvRead: []string{"HOME"}, PromptScope: "repo",
	}
	subject := subjectWith(t, "curl https://github.com/data\n", caps, 3)
	digest, err := hashing.ContentSHA256WithVersion(subject.Snapshot, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	storeCachedFindings(cfg, digest, subject, nil, nil)
	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, subject); !hit {
		t.Fatal("same effective capabilities missed")
	}
	for _, change := range []struct {
		name  string
		apply func(*capabilities.Manifest)
	}{
		{"network", func(c *capabilities.Manifest) { c.Network = nil }},
		{"exec", func(c *capabilities.Manifest) { c.Exec = nil }},
		{"filesystem-keyword", func(c *capabilities.Manifest) { c.Filesystem.Keyword = "home-config" }},
		{"filesystem-paths", func(c *capabilities.Manifest) { c.Filesystem.Paths = []string{"src"} }},
		{"secrets", func(c *capabilities.Manifest) { c.Secrets = nil }},
		{"env-read", func(c *capabilities.Manifest) { c.EnvRead = nil }},
		{"prompt-scope", func(c *capabilities.Manifest) { c.PromptScope = "other" }},
	} {
		t.Run(change.name, func(t *testing.T) {
			probe := subject
			change.apply(&probe.Capabilities)
			if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, probe); hit {
				t.Fatal("changed capabilities reused the allowance")
			}
		})
	}
}

// Ruleset-2 allowances cannot authorize current audits; fresh refusals cache normally and old records remain byte-identical.
func TestCapabilityCacheRulesetRetainsOldRecords(t *testing.T) {
	cfg := newCfg(t, "strict", "high")
	subject := subjectWith(t, "curl https://github.com/data\n", capabilities.ImplicitNone(), 3)
	digest, err := hashing.ContentSHA256WithVersion(subject.Snapshot, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(trustDir(cfg.Home(), digest), "verdict-"+strings.Join([]string{cfg.Audit.Backend, cfg.Audit.Model, PromptVersion, "2"}, "-")+".json")
	old := []byte(`{"schema_version":1,"commit":"abc","directory":".","findings":[]}`)
	if err := os.MkdirAll(filepath.Dir(oldPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPath, old, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, wantHit := range []bool{false, true} {
		report, err := auditSubject(cfg, subject, true)
		if err != nil || report.CacheHit != wantHit || report.Decision != DecisionBlock || len(report.Findings) != 1 || report.Findings[0].ID != "audit.capability.network-undeclared" {
			t.Fatalf("audit = %+v, error=%v; want block, cache hit=%v", report, err, wantHit)
		}
	}
	if retained, err := os.ReadFile(oldPath); err != nil || !bytes.Equal(retained, old) {
		t.Fatalf("old record changed: error=%v", err)
	}
	declared := subject
	declared.Capabilities = capabilities.Manifest{Network: []string{"github.com"}}
	for _, wantHit := range []bool{false, true} {
		report, err := auditSubject(cfg, declared, true)
		if err != nil || report.CacheHit != wantHit || report.Decision != DecisionAllow || len(report.Findings) != 0 {
			t.Fatalf("declared control = %+v, error=%v; want allow, cache hit=%v", report, err, wantHit)
		}
	}
}
