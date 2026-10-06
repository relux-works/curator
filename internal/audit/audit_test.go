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

func TestOpaqueNULBlocksEvenWhenAV1TwinHasCachedAllow(t *testing.T) {
	nulTree := t.TempDir()
	cleanTwin := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{
		"assets/a.bin": []byte("x\x00docs/b.md\x00y"),
	})
	writeFiles(t, cleanTwin, map[string][]byte{
		"assets/a.bin": []byte("x"),
		"docs/b.md":    []byte("y"),
	})
	leftHash, err := hashing.ContentSHA256(nulTree, nil)
	if err != nil {
		t.Fatal(err)
	}
	rightHash, err := hashing.ContentSHA256(cleanTwin, nil)
	if err != nil {
		t.Fatal(err)
	}
	if leftHash != rightHash {
		t.Fatalf("v1 framing did not collide: %s != %s", leftHash, rightHash)
	}

	cfg := newCfg(t, "advisory", "off")
	clean := Subject{Name: "clean-twin", Snapshot: cleanTwin, SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
	if warnings, errs := Gate(cfg, []Subject{clean}); len(errs) != 0 || len(warnings) != 0 {
		t.Fatalf("NUL-free twin: warnings=%v errs=%v", warnings, errs)
	}
	if findings, hit := loadCachedFindings(cfg, leftHash, hashing.VersionV1); !hit || len(findings) != 0 {
		t.Fatalf("clean twin cache = (%+v, hit=%v), want cached allow", findings, hit)
	}

	opaque := Subject{Name: "nul-tree", Snapshot: nulTree, SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
	if _, errs := Gate(cfg, []Subject{opaque}); len(errs) == 0 {
		t.Fatal("NUL-bearing tree was admitted from its clean twin's cached allow")
	}
	report, err := auditSubject(cfg, opaque, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionBlock || report.CacheHit || len(report.Findings) != 1 || report.Findings[0].ID != findingOpaqueNUL || report.Findings[0].File != "assets/a.bin" {
		t.Fatalf("opaque report %+v, want an uncached blocking finding naming assets/a.bin", report)
	}
}

func TestGateBlocksDeepOpaqueNULWhenAuditIsDisabledAndNamesFile(t *testing.T) {
	snapshot := t.TempDir()
	writeFiles(t, snapshot, map[string][]byte{
		"docs/deep/opaque.unsupported": []byte("prefix\x00suffix"),
	})
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json")}
	if cfg.Audit.Enabled {
		t.Fatal("test requires the default disabled audit configuration")
	}

	warnings, errs := Gate(cfg, []Subject{{
		Name: "opaque-skill", Snapshot: snapshot, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(),
	}})
	if len(warnings) != 0 || len(errs) != 1 ||
		!strings.Contains(errs[0], findingOpaqueNUL) ||
		!strings.Contains(errs[0], "docs/deep/opaque.unsupported") {
		t.Fatalf("Gate with disabled audit = warnings %v, errors %v; want a blocking file-naming opaque finding", warnings, errs)
	}
}

func writeFiles(t *testing.T, root string, files map[string][]byte) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
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

func TestGateReadOnlyDoesNotWriteVerdictCache(t *testing.T) {
	subject := subjectWith(t, "curl https://exfil.example.net/x\n", capabilities.ImplicitNone(), 3)
	cfg := newCfg(t, "advisory", "high")
	warnings, errs := GateReadOnly(cfg, []Subject{subject})
	if len(errs) != 0 || len(warnings) == 0 {
		t.Fatalf("read-only gate: warnings=%v errs=%v", warnings, errs)
	}
	if _, err := os.Stat(filepath.Join(cfg.Home(), "audit")); !os.IsNotExist(err) {
		t.Fatalf("read-only gate wrote audit state: %v", err)
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

// Gate is the production entry used by install (including global install) and
// the audit CLI; GateReadOnly is the corresponding dry-run entry.
func TestGatePinPolicyFreshAndCached(t *testing.T) {
	cases := []struct {
		name, mode, failOn, script, want string
		schema                           int
		pinned, revoked                  bool
	}{
		{"strict schema3 at threshold", "strict", "high", "curl https://exfil.example.net/x\n", DecisionBlock, 3, true, false},
		{"strict schema4 above threshold", "strict", "medium", "curl https://exfil.example.net/x\n", DecisionBlock, 4, true, false},
		{"strict schema2 with finding", "strict", "high", "curl https://exfil.example.net/x\n", DecisionBlock, 2, true, false},
		{"strict schema1 clean pinned", "strict", "high", "echo ok\n", DecisionAllow, 1, true, false},
		{"strict schema2 clean pinned", "strict", "high", "echo ok\n", DecisionAllow, 2, true, false},
		{"strict schema2 clean unpinned", "strict", "high", "echo ok\n", DecisionRequirePin, 2, false, false},
		{"strict revoked pin", "strict", "high", "echo ok\n", DecisionBlock, 3, true, true},
		{"advisory revoked pin", "advisory", "high", "echo ok\n", DecisionBlock, 2, true, true},
		{"advisory schema3 pinned finding", "advisory", "high", "curl https://exfil.example.net/x\n", DecisionWarn, 3, true, false},
		{"advisory schema2 pinned finding", "advisory", "high", "curl https://exfil.example.net/x\n", DecisionWarn, 2, true, false},
		{"advisory unpinned finding", "advisory", "high", "curl https://exfil.example.net/x\n", DecisionWarn, 3, false, false},
		{"strict pinned below threshold", "strict", "critical", "curl https://exfil.example.net/x\n", DecisionWarn, 3, true, false},
		{"strict pinned fail_on off", "strict", "off", "curl https://exfil.example.net/x\n", DecisionWarn, 3, true, false},
	}
	for _, entry := range []struct {
		name string
		gate func(*config.Config, []Subject) ([]string, []string)
	}{
		{"Gate", Gate},
		{"GateReadOnly", GateReadOnly},
	} {
		for _, cached := range []bool{false, true} {
			path := "fresh"
			if cached {
				path = "cached"
			}
			for _, tc := range cases {
				t.Run(entry.name+"/"+path+"/"+tc.name, func(t *testing.T) {
					cfg := newCfg(t, tc.mode, tc.failOn)
					subject := subjectWith(t, tc.script, capabilities.ImplicitNone(), tc.schema)
					contentHash, err := hashing.ContentSHA256(subject.Snapshot, nil)
					if err != nil {
						t.Fatal(err)
					}
					if cached {
						// Populate through production, before pinning/revoking. The
						// target policy must be recomputed from cached findings.
						warm := *cfg
						warm.Audit.Mode = "advisory"
						_, errs := Gate(&warm, []Subject{subject})
						if len(errs) != 0 {
							t.Fatalf("warm cache: %v", errs)
						}
					}
					if _, hit := loadCachedFindings(cfg, contentHash, hashing.VersionV1); hit != cached {
						t.Fatalf("cache hit before gate = %v, want %v", hit, cached)
					}
					if tc.pinned {
						if _, err := Pin(cfg.Home(), contentHash, "reviewed manually", "tester"); err != nil {
							t.Fatal(err)
						}
					}
					if tc.revoked {
						cfg.Audit.Revocations = []string{contentHash}
					}
					warnings, errs := entry.gate(cfg, []Subject{subject})
					switch tc.want {
					case DecisionBlock:
						evidence := "audit.capability.network-undeclared"
						if tc.revoked {
							evidence = "revoked"
						}
						if len(warnings) != 0 || len(errs) != 1 || !strings.Contains(errs[0], "audit blocked:") || !strings.Contains(errs[0], evidence) {
							t.Fatalf("want block (%s): warnings=%v errs=%v", evidence, warnings, errs)
						}
					case DecisionRequirePin:
						if len(warnings) != 0 || len(errs) != 1 || !strings.Contains(errs[0], "audit requires pin:") {
							t.Fatalf("want require_pin: warnings=%v errs=%v", warnings, errs)
						}
					case DecisionWarn:
						if len(errs) != 0 || len(warnings) != 1 || !strings.Contains(warnings[0], "audit warning:") || !strings.Contains(warnings[0], "audit.capability.network-undeclared") {
							t.Fatalf("want warn: warnings=%v errs=%v", warnings, errs)
						}
					case DecisionAllow:
						if len(warnings) != 0 || len(errs) != 0 {
							t.Fatalf("want allow: warnings=%v errs=%v", warnings, errs)
						}
					}
				})
			}
		}
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
