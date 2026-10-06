// Rework-1 regressions for the Spec §8 version rule of the opaque-NUL
// gate: refusal happens before any v1 identity is computed, and v2 audit
// verdicts live in a versioned carrier that legacy v1 records never
// satisfy. These tests adapt the revision-1 reviewer's attack probes
// into maintained production-entry coverage.
package audit

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
)

// A mixed gate over one NUL tree refuses exactly the v1 subject; the v2
// subject over the same bytes is admitted silently.
func TestGateMixedVersionsRefuseOnlyTheV1Subject(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"deep/a.bin": []byte("x\x00y")})
	cfg := newCfg(t, "advisory", "off")
	_, errs := Gate(cfg, []Subject{
		{Name: "legacy", Snapshot: root},
		{Name: "modern", Snapshot: root, HashVersion: hashing.VersionV2},
	})
	if len(errs) != 1 || !strings.Contains(errs[0], "audit blocked: legacy: critical audit.opaque.nul-byte") {
		t.Fatalf("mixed gate errors = %v, want exactly the legacy opaque refusal", errs)
	}
}

// The v1 opaque refusal precedes hashing: a blocked v1 report carries no
// computed ContentSHA256 at all. A hashing-behind-refusal regression
// would populate this field.
func TestAuditSubjectV1NULBlockCarriesNoComputedIdentity(t *testing.T) {
	nulTree := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{
		"assets/a.bin": []byte("x\x00docs/b.md\x00y"),
	})
	cfg := newCfg(t, "advisory", "off")
	for _, version := range []hashing.Version{0, hashing.VersionV1} {
		subject := Subject{Name: "nul-tree", Snapshot: nulTree, SchemaVersion: 3,
			Capabilities: capabilities.ImplicitNone(), HashVersion: version}
		report, err := auditSubject(cfg, subject, true)
		if err != nil {
			t.Fatal(err)
		}
		if report.Decision != DecisionBlock || len(report.Findings) != 1 {
			t.Fatalf("version %d: report %+v, want a block with one finding", version, report)
		}
		if report.ContentSHA256 != "" {
			t.Fatalf("version %d: refusal carries computed v1 identity %s, want none", version, report.ContentSHA256)
		}
	}
}

// The draft source-audit lane refuses a NUL tree with the opaque finding
// and stores no verdict: the frozen v1 binding is never computed over
// those bytes and never cached.
func TestCheckSourceAuditV1NULRefusesWithoutVerdict(t *testing.T) {
	nulTree := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	v1Hash, err := hashing.ContentSHA256(nulTree, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "advisory", "off")
	if _, err := CheckSourceAudit(cfg, sourceTestSubject(nulTree, 3), true, sourceAuditTestNow); err == nil ||
		!strings.Contains(err.Error(), "audit.opaque.nul-byte") {
		t.Fatalf("draft source audit over NUL = %v, want the opaque refusal", err)
	}
	if _, hit := loadCachedFindings(cfg, v1Hash, hashing.VersionV1); hit {
		t.Fatal("draft source audit cached a verdict for a NUL tree")
	}
}

// A v2 verdict is stored in the versioned carrier: schema 2 with an
// explicit hash_version 2, never the frozen v1 shape.
func TestV2VerdictCarrierRecordsHashVersion(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"assets/a.bin": []byte("x\x00y")})
	cfg := newCfg(t, "advisory", "off")
	subject := Subject{Name: "nul", Snapshot: root, HashVersion: hashing.VersionV2,
		SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
	report, err := auditSubject(cfg, subject, true)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(verdictPath(cfg, report.ContentSHA256))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["schema_version"] != float64(2) || raw["hash_version"] != float64(2) {
		t.Fatalf("v2 verdict carrier = schema_version %v hash_version %v, want 2/2",
			raw["schema_version"], raw["hash_version"])
	}
}

// A v2 subject never consumes a legacy unversioned record, even at equal
// digest text: version and digest form one identity (Spec §8). The equal
// digest text is a synthetic version-pair fixture, not a hash collision.
func TestV2RejectsLegacyUnversionedVerdict(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("contact https://unlisted.example/")})
	digest, err := hashing.ContentSHA256WithVersion(root, nil, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "advisory", "off")
	legacy := Subject{Name: "legacy", Snapshot: root, HashVersion: hashing.VersionV1, SchemaVersion: 3}
	storeCachedFindings(cfg, digest, legacy, nil, nil)
	modern := Subject{Name: "modern", Snapshot: root, HashVersion: hashing.VersionV2,
		SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
	report, err := auditSubject(cfg, modern, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.CacheHit || len(report.Findings) == 0 {
		t.Fatalf("v2 report %+v, want a fresh detection, not the legacy-shaped cache entry", report)
	}
}

// Symmetrically, a v1 subject never consumes a v2 record: the seeded v2
// allow at the v1 digest must not authorize the v1 read.
func TestV1IgnoresV2VerdictCarrier(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("contact https://unlisted.example/")})
	digest, err := hashing.ContentSHA256WithVersion(root, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "advisory", "off")
	modern := Subject{Name: "modern", Snapshot: root, HashVersion: hashing.VersionV2, SchemaVersion: 3}
	storeCachedFindings(cfg, digest, modern, nil, nil)
	legacy := Subject{Name: "legacy", Snapshot: root, HashVersion: hashing.VersionV1,
		SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
	report, err := auditSubject(cfg, legacy, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.CacheHit || len(report.Findings) == 0 {
		t.Fatalf("v1 report %+v, want a fresh detection, not the v2 cache entry", report)
	}
}

// Legacy schema-1 records without hash_version keep authorizing v1
// reads, and the script-policy backfill preserves their frozen shape.
func TestV1HonorsLegacySchema1Verdict(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("contact https://unlisted.example/")})
	digest, err := hashing.ContentSHA256WithVersion(root, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "advisory", "off")
	legacy := Subject{Name: "legacy", Snapshot: root, HashVersion: hashing.VersionV1,
		SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
	storeCachedFindings(cfg, digest, legacy, nil, nil)
	assertFrozenV1Carrier(t, cfg, digest)
	report, err := auditSubject(cfg, legacy, true)
	if err != nil {
		t.Fatal(err)
	}
	if !report.CacheHit || len(report.Findings) != 0 {
		t.Fatalf("v1 report %+v, want the cached legacy allow", report)
	}
	assertFrozenV1Carrier(t, cfg, digest)
}

func assertFrozenV1Carrier(t *testing.T, cfg *config.Config, digest string) {
	t.Helper()
	payload, err := os.ReadFile(verdictPath(cfg, digest))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["schema_version"] != float64(1) {
		t.Fatalf("v1 verdict carrier schema_version = %v, want 1", raw["schema_version"])
	}
	if _, present := raw["hash_version"]; present {
		t.Fatalf("v1 verdict carrier gained hash_version %v, want the frozen shape", raw["hash_version"])
	}
}
