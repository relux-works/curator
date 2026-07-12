package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
)

func TestDecideTable(t *testing.T) {
	high := []Finding{{ID: "x", Severity: SeverityHigh, Verifiable: true}}
	low := []Finding{{ID: "x", Severity: SeverityLow, Verifiable: true}}
	unverifiable := []Finding{{ID: "x", Severity: SeverityCritical, Verifiable: false}}
	cases := []struct {
		name     string
		findings []Finding
		mode     string
		failOn   string
		want     string
	}{
		{"clean allow", nil, "strict", "high", DecisionAllow},
		{"advisory warns", high, "advisory", "high", DecisionWarn},
		{"fail_on off warns", high, "strict", "off", DecisionWarn},
		{"strict blocks at threshold", high, "strict", "high", DecisionBlock},
		{"strict below threshold warns", low, "strict", "high", DecisionWarn},
		{"strict low threshold blocks low", low, "strict", "low", DecisionBlock},
		{"unverifiable never blocks", unverifiable, "strict", "low", DecisionWarn},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Decide(tc.findings, tc.mode, tc.failOn); got != tc.want {
				t.Fatalf("Decide = %s, want %s", got, tc.want)
			}
		})
	}
}

func newCfg(t *testing.T, mode, failOn string) *config.Config {
	home := t.TempDir()
	return &config.Config{
		Path: filepath.Join(home, "config.json"),
		Audit: config.Audit{
			Enabled: true, Mode: mode, FailOn: failOn, Backend: "null",
		},
	}
}

func subjectWith(t *testing.T, script string, caps capabilities.Manifest, schema int) Subject {
	snapshot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(snapshot, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshot, "scripts", "tool"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	return Subject{
		Name: "skill-a", Source: "skill-a", Git: "git@git.example.com:skills/skill-a.git",
		Commit: "abc", Snapshot: snapshot, SchemaVersion: schema, Capabilities: caps,
	}
}

func TestDetectorsAgainstCapabilities(t *testing.T) {
	// undeclared host and binary produce findings
	subject := subjectWith(t, "curl https://api.example.com/x\nsubprocess(\"jq\")\n", capabilities.ImplicitNone(), 3)
	findings := detect(subject.Snapshot, subject.Capabilities)
	if len(findings) != 2 {
		t.Fatalf("findings: %+v", findings)
	}
	// declared capabilities silence them
	declared := capabilities.Manifest{Network: []string{"api.example.com"}, Exec: []string{"jq"}}
	findings = detect(subject.Snapshot, declared)
	if len(findings) != 0 {
		t.Fatalf("declared capabilities must silence findings: %+v", findings)
	}
	// glob hosts work
	globbed := capabilities.Manifest{Network: []string{"*.example.com"}, Exec: []string{"jq"}}
	if findings = detect(subject.Snapshot, globbed); len(findings) != 0 {
		t.Fatalf("glob host must match: %+v", findings)
	}
}

func TestGateModes(t *testing.T) {
	subject := subjectWith(t, "curl https://exfil.example.net/x\n", capabilities.ImplicitNone(), 3)

	// advisory: warnings, no errors
	warnings, errs := Gate(newCfg(t, "advisory", "high"), []Subject{subject})
	if len(errs) != 0 || len(warnings) == 0 {
		t.Fatalf("advisory: warnings=%v errs=%v", warnings, errs)
	}
	// strict: blocks
	warnings, errs = Gate(newCfg(t, "strict", "high"), []Subject{subject})
	if len(errs) == 0 {
		t.Fatalf("strict must block: warnings=%v", warnings)
	}
	// disabled: nothing
	disabled := newCfg(t, "strict", "high")
	disabled.Audit.Enabled = false
	warnings, errs = Gate(disabled, []Subject{subject})
	if warnings != nil || errs != nil {
		t.Fatalf("disabled audit must be silent: %v %v", warnings, errs)
	}
}

func TestRequirePinForOldSchemas(t *testing.T) {
	subject := subjectWith(t, "echo ok\n", capabilities.ImplicitNone(), 2)
	cfg := newCfg(t, "strict", "high")
	_, errs := Gate(cfg, []Subject{subject})
	if len(errs) != 1 || !strings.Contains(errs[0], "requires pin") {
		t.Fatalf("errs: %v", errs)
	}
	// pin clears it
	contentHash, _ := hashing.ContentSHA256(subject.Snapshot, nil)
	if _, err := Pin(cfg.Home(), contentHash, "reviewed manually", "tester"); err != nil {
		t.Fatal(err)
	}
	_, errs = Gate(cfg, []Subject{subject})
	if len(errs) != 0 {
		t.Fatalf("pinned subject must pass: %v", errs)
	}
}

func TestLocalRevocations(t *testing.T) {
	subject := subjectWith(t, "echo ok\n", capabilities.ImplicitNone(), 3)
	contentHash, _ := hashing.ContentSHA256(subject.Snapshot, nil)

	cfg := newCfg(t, "advisory", "high")
	cfg.Audit.Revocations = []string{contentHash}
	_, errs := Gate(cfg, []Subject{subject})
	if len(errs) != 1 || !strings.Contains(errs[0], "revoked") {
		t.Fatalf("hash revocation must block even in advisory mode: %v", errs)
	}

	cfg = newCfg(t, "advisory", "high")
	cfg.Audit.Revocations = []string{"source:git@git.example.com:skills/*"}
	_, errs = Gate(cfg, []Subject{subject})
	if len(errs) != 1 || !strings.Contains(errs[0], "revoked") {
		t.Fatalf("source glob revocation must block: %v", errs)
	}
}

func TestVerdictCacheRedecidesUnderCurrentPolicy(t *testing.T) {
	subject := subjectWith(t, "curl https://exfil.example.net/x\n", capabilities.ImplicitNone(), 3)
	home := t.TempDir()
	advisory := &config.Config{Path: filepath.Join(home, "config.json"),
		Audit: config.Audit{Enabled: true, Mode: "advisory", FailOn: "high", Backend: "null"}}
	warnings, errs := Gate(advisory, []Subject{subject})
	if len(errs) != 0 || len(warnings) == 0 {
		t.Fatalf("first advisory run: %v %v", warnings, errs)
	}
	// same home, strict policy: the cached findings must now block
	strict := &config.Config{Path: filepath.Join(home, "config.json"),
		Audit: config.Audit{Enabled: true, Mode: "strict", FailOn: "high", Backend: "null"}}
	_, errs = Gate(strict, []Subject{subject})
	if len(errs) == 0 {
		t.Fatal("cache hit must re-decide under the current policy")
	}
}

func TestCanaryFires(t *testing.T) {
	if !runStaticCanary() {
		t.Fatal("static canary must pass with working detectors")
	}
}
