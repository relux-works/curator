package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

// The 40 pinned §9.1/§9.4 vectors drive Gate and GateReadOnly, including
// fresh/cache decisions and exact host/file findings. Fixture files live in
// scripts/, the existing detector surface; no broader extraction is claimed.
func TestNetworkContactVectors(t *testing.T) {
	payload, err := os.ReadFile("testdata/network-contacts.json")
	if err != nil {
		t.Fatal(err)
	}
	type finding struct{ Class, Host, File string }
	var corpus struct {
		Cases []struct {
			Name     string
			Files    map[string]string
			Manifest map[string]any
			Expected struct {
				Findings []finding
				Passes   bool `json:"capability_check_passes"`
				Canary   bool `json:"static_canary_blocks"`
			}
		}
	}
	if err := json.Unmarshal(payload, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 40 {
		t.Fatalf("vectors = %d, want 40", len(corpus.Cases))
	}
	for _, tc := range corpus.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			caps, err := capabilities.Parse(tc.Manifest["capabilities"])
			if err != nil {
				t.Fatal(err)
			}
			subject := subjectWith(t, "", caps, 3)
			for file, text := range tc.Files {
				if err := os.WriteFile(filepath.Join(subject.Snapshot, "scripts", file), []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg := newCfg(t, "strict", "high")
			for _, wantHit := range []bool{false, true} {
				before, err := auditSubject(cfg, subject, false)
				if err != nil || before.CacheHit != wantHit {
					t.Fatalf("before Gate: hit=%v, want=%v, error=%v", before.CacheHit, wantHit, err)
				}
				warnings, errs := Gate(cfg, []Subject{subject})
				if (len(errs) == 0) != tc.Expected.Passes || len(warnings) != 0 {
					t.Fatalf("Gate pass=%v: warnings=%v errors=%v", tc.Expected.Passes, warnings, errs)
				}
				report, err := auditSubject(cfg, subject, false)
				if err != nil || !report.CacheHit {
					t.Fatalf("persisted report = %+v, error=%v", report, err)
				}
				var got []finding
				for _, f := range report.Findings {
					host := strings.TrimSuffix(strings.TrimPrefix(f.Evidence, "contacts "), ", which no capabilities.network entry covers")
					got = append(got, finding{f.ID, host, strings.TrimPrefix(f.File, "scripts/")})
				}
				want := tc.Expected.Findings
				if len(want) == 0 {
					want = nil
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("findings = %+v, want %+v (repeat=%v)", got, want, wantHit)
				}
				_, readErrs := GateReadOnly(cfg, []Subject{subject})
				if !reflect.DeepEqual(errs, readErrs) {
					t.Fatalf("read-only changed verdict: %v != %v", readErrs, errs)
				}
			}
			// Canary detection is independent of a subject's declaration: its
			// production fixture must still yield a blocking network finding.
			if tc.Expected.Canary {
				subject.Capabilities = capabilities.ImplicitNone()
				_, errs := Gate(newCfg(t, "strict", "high"), []Subject{subject})
				if len(errs) != 1 || !strings.Contains(errs[0], "contacts exfiltrate.example.com") {
					t.Fatalf("static canary did not block: %v", errs)
				}
			}
			if !CanaryPasses() {
				t.Fatal("production static canary failed")
			}
		})
	}
}

// A genuine ruleset-3 refusal (or stale allow for a real host) cannot mask
// ruleset-4 classification in either hash framing; old bytes are retained.
func TestNetworkContactCacheInvalidatesRulesetThree(t *testing.T) {
	for _, version := range []hashing.Version{hashing.VersionV1, hashing.VersionV2} {
		for _, host := range []string{"api.test", "api.example.com"} {
			t.Run(fmt.Sprintf("v%d/%s", version, host), func(t *testing.T) {
				cfg := newCfg(t, "strict", "high")
				subject := subjectWith(t, "https://"+host+"/fixture\n", capabilities.ImplicitNone(), 3)
				subject.HashVersion = version
				digest, err := hashing.ContentSHA256WithVersion(subject.Snapshot, nil, version)
				if err != nil {
					t.Fatal(err)
				}
				oldFindings := []Finding{{ID: "audit.capability.network-undeclared", Severity: SeverityHigh, File: "scripts/tool", Evidence: "contacts api.test, which no capabilities.network entry covers", Verifiable: true}}
				if host != "api.test" {
					oldFindings = nil
				}
				storeCachedFindings(cfg, digest, subject, oldFindings, nil)
				currentPath := verdictPath(cfg, digest)
				oldPath := filepath.Join(filepath.Dir(currentPath), "verdict-"+strings.Join([]string{cfg.Audit.Backend, cfg.Audit.Model, PromptVersion, "3"}, "-")+".json")
				old, err := os.ReadFile(currentPath)
				if err != nil {
					t.Fatal(err)
				}
				var record map[string]any
				if err := json.Unmarshal(old, &record); err != nil {
					t.Fatal(err)
				}
				record["ruleset_version"] = "3"
				old, err = json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(currentPath); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(oldPath, old, 0600); err != nil {
					t.Fatal(err)
				}
				for _, wantHit := range []bool{false, true} {
					report, err := auditSubject(cfg, subject, true)
					wantDecision := DecisionBlock
					if host == "api.test" {
						wantDecision = DecisionAllow
					}
					if err != nil || report.CacheHit != wantHit || report.Decision != wantDecision {
						t.Fatalf("stale ruleset reused: report=%+v error=%v, want hit=%v decision=%s", report, err, wantHit, wantDecision)
					}
					_, errs := Gate(cfg, []Subject{subject})
					if (len(errs) == 0) != (host == "api.test") {
						t.Fatalf("Gate stale verdict: %v", errs)
					}
				}
				retained, err := os.ReadFile(oldPath)
				if err != nil || !bytes.Equal(retained, old) {
					t.Fatalf("old record changed: %v", err)
				}
			})
		}
	}
}

// URL authority parsing excludes credentials, brackets, ports and paths.
// A malformed path escape cannot hide a real host. A second trailing dot
// and a non-loopback IPv6 host cannot gain an exemption;
// exempt network hosts do not erase an independent exec finding.
func TestNetworkContactNormalizationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, raw, want string
		passes          bool
	}{
		{"reserved-port", "https://api.test:8443/x", "", true},
		{"percent-path-real", "https://api.example.com/%zz", "api.example.com", false},
		{"percent-path-reserved", "https://api.test/%zz", "", true},
		{"loopback-port", "http://[::1]:8080/x", "", true},
		{"real-ipv6", "http://[2001:db8::2]:8080/x", "2001:db8::2", false},
		{"credentials-are-not-host", "https://api.test@evil.com:443/x", "evil.com", false},
		{"port-is-not-host", "https://EXAMPLE.COM.:443/x", "example.com", false},
		{"one-dot-only", "https://api.test../x", "api.test.", false},
		{"invalid-idna", "https://bad_label.test/x", "bad_label.test", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			subject := subjectWith(t, tc.raw+"\n", capabilities.ImplicitNone(), 3)
			_, errs := Gate(newCfg(t, "strict", "high"), []Subject{subject})
			if (len(errs) == 0) != tc.passes {
				t.Fatalf("URL classification: %v", errs)
			}
			if !tc.passes && (len(errs) != 1 || !strings.Contains(errs[0], "contacts "+tc.want+",")) {
				t.Fatalf("host evidence: %v", errs)
			}
		})
	}
	subject := subjectWith(t, "https://api.test/x\nsubprocess(\"nc\")\n", capabilities.ImplicitNone(), 3)
	_, errs := Gate(newCfg(t, "strict", "medium"), []Subject{subject})
	if len(errs) != 1 || !strings.Contains(errs[0], "audit.capability.exec-undeclared") {
		t.Fatalf("reserved host cleared exec gate: %v", errs)
	}
}
