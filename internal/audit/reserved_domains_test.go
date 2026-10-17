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

// Reserved descendants are silent for both schemes, ignoring case and one root dot; bare localhost keeps its exception.
func TestReservedTopLevelDomains(t *testing.T) {
	for _, host := range []string{
		"board.example.test", "BOARD.EXAMPLE.TEST", "Board.Example.TeSt.", "board.test:8443", "Board.TeSt.:8443",
		"docs.example", "Docs.Example.",
		"missing.invalid", "Missing.Invalid.",
		"localhost", "LOCALHOST", "127.0.0.1", "service.localhost", "Service.LocalHost.",
	} {
		for _, scheme := range []string{"http", "https"} {
			t.Run(scheme+"/"+host, func(t *testing.T) {
				subject := subjectWith(t, "curl "+scheme+"://"+host+"/fixture\n", capabilities.ImplicitNone(), 3)
				if findings := detect(subject.Snapshot, subject.Capabilities); len(findings) != 0 {
					t.Fatalf("reserved host produced findings: %+v", findings)
				}
			})
		}
	}
}

// Bare TLD labels, real hosts, suffix lookalikes and second-level example domains retain high, verifiable findings.
func TestReservedDomainLookalikesRemainAudited(t *testing.T) {
	for _, host := range []string{
		"github.com", "brew.sh", "GITHUB.COM.",
		"test", "TEST", "test.", "example", "EXAMPLE.", "invalid", "INVALID.", "localhost.",
		"evil.test.com", "example.com.attacker", "notest", "evillocalhost",
		"test.attacker", "invalid.attacker", "localhost.attacker",
		"example.com", "example.net", "example.org", "exfiltrate.example.com",
		"api.example.net", "api.example.org", "board.test..", "board..test", "-board.test",
	} {
		t.Run(host, func(t *testing.T) {
			subject := subjectWith(t, "curl https://"+host+"/data\n", capabilities.ImplicitNone(), 3)
			findings := detect(subject.Snapshot, subject.Capabilities)
			if len(findings) != 1 || findings[0].ID != "audit.capability.network-undeclared" ||
				findings[0].Severity != SeverityHigh || !findings[0].Verifiable ||
				!strings.Contains(findings[0].Evidence, "contacts "+strings.ToLower(host)+",") {
				t.Fatalf("host %q produced findings %+v", host, findings)
			}
		})
	}
}

// Reserved-looking userinfo cannot hide a real authority; invalid ports and DNS labels retain findings.
func TestReservedDomainAuthorityBoundary(t *testing.T) {
	for _, tc := range []struct {
		url, wantHost string
	}{
		{"https://board.test@github.com/data", "github.com"},
		{"https://board.test:443@brew.sh/data", "brew.sh"},
		{"https://board.test:bad/data", "board.test"},
		{"https://board_.test/data", "board"},
	} {
		t.Run(tc.url, func(t *testing.T) {
			subject := subjectWith(t, "curl "+tc.url+"\n", capabilities.ImplicitNone(), 3)
			findings := detect(subject.Snapshot, subject.Capabilities)
			if len(findings) != 1 || findings[0].ID != "audit.capability.network-undeclared" ||
				!strings.Contains(findings[0].Evidence, "contacts "+tc.wantHost+",") {
				t.Fatalf("URL produced findings %+v, want host %s", findings, tc.wantHost)
			}
		})
	}
}

// Declared real hosts retain the previous detector behavior at shell and prose punctuation boundaries.
func TestDeclaredHostURLPunctuation(t *testing.T) {
	for _, text := range []string{
		"$(curl https://github.com)",
		"See https://github.com, then continue",
		"curl https://github.com:443",
	} {
		t.Run(text, func(t *testing.T) {
			subject := subjectWith(t, text, capabilities.Manifest{Network: []string{"github.com"}}, 3)
			if findings := detect(subject.Snapshot, subject.Capabilities); len(findings) != 0 {
				t.Fatalf("declared host produced findings at a punctuation boundary: %+v", findings)
			}
		})
	}
}

// Terminal shell/prose punctuation permits reserved hosts, but userinfo still identifies the real host.
func TestReservedHostURLPunctuation(t *testing.T) {
	for _, text := range []string{"$(curl https://board.test)", "See https://board.test, then continue", "(https://Board.Test.:443)", "curl https://board.test;"} {
		t.Run(text, func(t *testing.T) {
			subject := subjectWith(t, text, capabilities.ImplicitNone(), 3)
			if findings := detect(subject.Snapshot, subject.Capabilities); len(findings) != 0 {
				t.Fatalf("reserved punctuation: %+v", findings)
			}
		})
	}
	for _, text := range []string{"$(curl https://board.test@evil.com)", "See https://board.test,@evil.com, then continue", "https://board.test;user@evil.com"} {
		subject := subjectWith(t, text, capabilities.ImplicitNone(), 3)
		if _, errs := Gate(newCfg(t, "strict", "high"), []Subject{subject}); len(errs) != 1 || !strings.Contains(errs[0], "contacts evil.com") {
			t.Fatalf("userinfo hid real host: %v", errs)
		}
	}
}

// Each URL start is inspected even when a declared or reserved URL directly precedes an undeclared URL.
func TestReservedURLDiscoveryKeepsLaterHosts(t *testing.T) {
	for _, text := range []string{
		"sed 's|https://github.com|https://evil.com|g'",
		"sed 's|https://board.test|https://evil.com|g'",
		"https://github.com,https://evil.com",
		"https://board.test@github.com|https://evil.com",
	} {
		t.Run(text, func(t *testing.T) {
			subject := subjectWith(t, text, capabilities.Manifest{Network: []string{"github.com"}}, 3)
			if _, errs := Gate(newCfg(t, "strict", "high"), []Subject{subject}); len(errs) == 0 || !strings.Contains(strings.Join(errs, "\n"), "contacts evil.com") {
				t.Fatalf("later undeclared URL hidden: %v", errs)
			}
		})
	}
}

// A reserved-only strict audit passes; adding an undeclared real host or revocation still blocks.
func TestReservedDomainsKeepStrictGate(t *testing.T) {
	cfg := newCfg(t, "strict", "high")
	reserved := subjectWith(t, "curl https://board.example.test/data\n", capabilities.ImplicitNone(), 3)
	if warnings, errs := Gate(cfg, []Subject{reserved}); len(warnings) != 0 || len(errs) != 0 {
		t.Fatalf("reserved-only gate: warnings=%v errors=%v", warnings, errs)
	}
	mixed := subjectWith(t, "curl https://board.example.test/data\ncurl https://github.com/data\n", capabilities.ImplicitNone(), 3)
	if _, errs := Gate(cfg, []Subject{mixed}); len(errs) != 1 ||
		!strings.Contains(errs[0], "audit.capability.network-undeclared") ||
		!strings.Contains(errs[0], "contacts github.com") || strings.Contains(errs[0], "contacts board.example.test") {
		t.Fatalf("mixed gate: errors=%v", errs)
	}
	cfg.Audit.Revocations = []string{"source:skill-a"}
	if _, errs := Gate(cfg, []Subject{reserved}); len(errs) != 1 || !strings.Contains(errs[0], "revoked") {
		t.Fatalf("revoked reserved-only gate: errors=%v", errs)
	}
	if !CanaryPasses() {
		t.Fatal("the existing real-host and executable static canary stopped detecting its plants")
	}
}

// A previous-ruleset refusal is not reused or overwritten when the current detector exempts a reserved host.
func TestReservedDomainsInvalidateOldRulesetCache(t *testing.T) {
	cfg := newCfg(t, "strict", "high")
	subject := subjectWith(t, "curl https://board.example.test/data\n", capabilities.ImplicitNone(), 3)
	digest, err := hashing.ContentSHA256WithVersion(subject.Snapshot, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	old := []byte(`{"schema_version":1,"commit":"abc","directory":".","findings":[{"id":"audit.capability.network-undeclared","severity":"high","file":"scripts/tool","evidence":"contacts board.example.test","verifiable":true}]}`)
	var oldPaths []string
	for _, ruleset := range []string{"2", "3"} {
		oldPath := filepath.Join(trustDir(cfg.Home(), digest), "verdict-"+strings.Join([]string{cfg.Audit.Backend, cfg.Audit.Model, PromptVersion, ruleset}, "-")+".json")
		if err := os.MkdirAll(filepath.Dir(oldPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(oldPath, old, 0o644); err != nil {
			t.Fatal(err)
		}
		oldPaths = append(oldPaths, oldPath)
	}
	if warnings, errs := Gate(cfg, []Subject{subject}); len(warnings) != 0 || len(errs) != 0 {
		t.Fatalf("current gate reused the old refusal: warnings=%v errors=%v", warnings, errs)
	}
	if current, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, subject); !hit || len(current) != 0 {
		t.Fatalf("current cache: findings=%+v hit=%v", current, hit)
	}
	for _, oldPath := range oldPaths {
		if retained, err := os.ReadFile(oldPath); err != nil || !bytes.Equal(retained, old) {
			t.Fatalf("old-ruleset record was changed: error=%v", err)
		}
	}
}
