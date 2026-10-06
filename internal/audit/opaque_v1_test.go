// Tests for the Spec §8 version rule of the opaque-NUL gate: the interim
// NUL block applies exactly when a v1 identity is computed or trusted,
// and never for a v2 computation or verification.
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

// writeNULCollisionPair builds the v1 framing collision: a NUL-bearing
// single-record tree and a NUL-free twin with the same v1 digest but
// different v2 digests.
func writeNULCollisionPair(t *testing.T, nulTree, cleanTwin string) (v1, v2NUL, v2Twin string) {
	t.Helper()
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
	v2Left, err := hashing.ContentSHA256WithVersion(nulTree, nil, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	v2Right, err := hashing.ContentSHA256WithVersion(cleanTwin, nil, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if v2Left == v2Right {
		t.Fatalf("v2 framing must distinguish the NUL tree from its twin: %s", v2Left)
	}
	return leftHash, v2Left, v2Right
}

// A v2 subject skips the opaque rule entirely, even when optional audit is
// disabled: NUL bytes are ordinary v2 data, so there is nothing to refuse
// and no scan failure to report.
func TestGateV2AdmitsNULBearingTreeWhenAuditIsDisabled(t *testing.T) {
	snapshot := t.TempDir()
	writeFiles(t, snapshot, map[string][]byte{
		"docs/deep/opaque.unsupported": []byte("prefix\x00suffix"),
	})
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json")}
	if cfg.Audit.Enabled {
		t.Fatal("test requires the default disabled audit configuration")
	}

	for _, entry := range []struct {
		name string
		gate func(*config.Config, []Subject) ([]string, []string)
	}{
		{"Gate", Gate},
		{"GateReadOnly", GateReadOnly},
	} {
		t.Run(entry.name, func(t *testing.T) {
			warnings, errs := entry.gate(cfg, []Subject{{
				Name: "opaque-skill", Snapshot: snapshot, SchemaVersion: 3,
				Capabilities: capabilities.ImplicitNone(), HashVersion: hashing.VersionV2,
			}})
			if warnings != nil || errs != nil {
				t.Fatalf("v2 gate over NUL tree = warnings %v, errors %v; want silent admission", warnings, errs)
			}
		})
	}
}

// A v2 audit of the NUL tree reports the v2 identity, never consults the
// v1 digest, and stores nothing under it: there is no path where a v1
// identity is computed or trusted over NUL bytes.
func TestGateV2ReportsV2IdentityAndStoresNothingUnderV1(t *testing.T) {
	nulTree, cleanTwin := t.TempDir(), t.TempDir()
	v1Hash, v2NUL, v2Twin := writeNULCollisionPair(t, nulTree, cleanTwin)
	cfg := newCfg(t, "advisory", "off")

	clean := Subject{Name: "clean-twin", Snapshot: cleanTwin, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(), HashVersion: hashing.VersionV2}
	if warnings, errs := Gate(cfg, []Subject{clean}); len(errs) != 0 || len(warnings) != 0 {
		t.Fatalf("NUL-free v2 twin: warnings=%v errs=%v", warnings, errs)
	}
	if findings, hit := loadCachedFindings(cfg, v2Twin, hashing.VersionV2); !hit || len(findings) != 0 {
		t.Fatalf("clean twin v2 cache = (%+v, hit=%v), want cached allow", findings, hit)
	}

	opaque := Subject{Name: "nul-tree", Snapshot: nulTree, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(), HashVersion: hashing.VersionV2}
	// The read-only audit runs before the persisting gate so the
	// freshness assertion below is meaningful.
	report, err := auditSubject(cfg, opaque, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionAllow || report.CacheHit || report.ContentSHA256 != v2NUL {
		t.Fatalf("v2 opaque report %+v, want a fresh allow carrying the v2 identity %s", report, v2NUL)
	}
	if warnings, errs := Gate(cfg, []Subject{opaque}); len(errs) != 0 || len(warnings) != 0 {
		t.Fatalf("NUL-bearing v2 tree: warnings=%v errs=%v, want admission", warnings, errs)
	}
	// The v1 digest of the NUL tree keys no trust state: no verdict was
	// stored under it, so a v1 reader can never inherit this admission.
	if _, err := os.Stat(filepath.Join(cfg.Home(), "audit", hashing.Normalize(v1Hash))); !os.IsNotExist(err) {
		t.Fatalf("v2 audit left trust state under the v1 digest: %v", err)
	}
	if _, hit := loadCachedFindings(cfg, v1Hash, hashing.VersionV1); hit {
		t.Fatal("v2 audit cached a verdict under the v1 digest")
	}
}

// Every v1 identity path keeps the blocking opaque finding with the
// unchanged id, severity, evidence, and message: explicit v1 subjects and
// legacy subjects with no recorded version alike.
func TestGateV1KeepsOpaqueBlockWithUnchangedFinding(t *testing.T) {
	nulTree := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{
		"assets/a.bin": []byte("x\x00docs/b.md\x00y"),
	})
	v1Hash, err := hashing.ContentSHA256(nulTree, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, version := range []hashing.Version{0, hashing.VersionV1} {
		t.Run(map[hashing.Version]string{0: "legacy-zero", hashing.VersionV1: "explicit-v1"}[version], func(t *testing.T) {
			subject := Subject{Name: "nul-tree", Snapshot: nulTree, SchemaVersion: 3,
				Capabilities: capabilities.ImplicitNone(), HashVersion: version}

			disabled := &config.Config{Path: filepath.Join(t.TempDir(), "config.json")}
			_, errs := Gate(disabled, []Subject{subject})
			want := "audit blocked: nul-tree: critical audit.opaque.nul-byte - " +
				"regular file contains a NUL byte and is treated as opaque (file: assets/a.bin)"
			if len(errs) != 1 || errs[0] != want {
				t.Fatalf("disabled-audit v1 gate errors = %v, want exactly %q", errs, want)
			}

			cfg := newCfg(t, "advisory", "off")
			if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 1 || !strings.Contains(errs[0], findingOpaqueNUL) {
				t.Fatalf("enabled-audit v1 gate errors = %v, want the blocking opaque finding", errs)
			}
			report, err := auditSubject(cfg, subject, true)
			if err != nil {
				t.Fatal(err)
			}
			if report.Decision != DecisionBlock || report.CacheHit || len(report.Findings) != 1 {
				t.Fatalf("v1 opaque report %+v, want an uncached block with one finding", report)
			}
			finding := report.Findings[0]
			if finding.ID != findingOpaqueNUL || finding.Severity != SeverityCritical ||
				finding.File != "assets/a.bin" ||
				finding.Evidence != "regular file contains a NUL byte and is treated as opaque" ||
				!finding.Verifiable {
				t.Fatalf("v1 opaque finding %+v, want the unchanged critical file-naming finding", finding)
			}
			// The blocked v1 result stays outside the verdict cache.
			if _, hit := loadCachedFindings(cfg, v1Hash, hashing.VersionV1); hit {
				t.Fatal("blocked v1 NUL result was stored in the verdict cache")
			}
		})
	}
}

// An unknown framing version refuses instead of guessing a rule: it must
// neither scan-and-block as v1 nor admit as v2, on clean trees either.
func TestGateRejectsUnknownHashVersion(t *testing.T) {
	nulTree := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	cleanTree := t.TempDir()
	writeFiles(t, cleanTree, map[string][]byte{
		"assets/a.bin": []byte("x"),
	})
	cfg := newCfg(t, "advisory", "off")

	for _, root := range []string{nulTree, cleanTree} {
		subject := Subject{Name: "skill-a", Snapshot: root, SchemaVersion: 3,
			Capabilities: capabilities.ImplicitNone(), HashVersion: hashing.Version(3)}
		if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 1 ||
			!strings.Contains(errs[0], "unsupported content hash version 3") {
			t.Fatalf("unknown-version gate over %s = %v, want an unsupported-version refusal", root, errs)
		}
		if _, errs := GateReadOnly(cfg, []Subject{subject}); len(errs) != 1 ||
			!strings.Contains(errs[0], "unsupported content hash version 3") {
			t.Fatalf("unknown-version read-only gate over %s = %v, want an unsupported-version refusal", root, errs)
		}
		if _, err := auditSubject(cfg, subject, false); err == nil ||
			!strings.Contains(err.Error(), "unsupported content hash version 3") {
			t.Fatalf("unknown-version auditSubject over %s = %v, want an unsupported-version error", root, err)
		}
	}
}

// The v2 rule is enforced at the decision point, not by caller convention:
// even handed NUL paths, a v2 subject never blocks on them.
func TestAuditSubjectWithOpaquePathsIgnoresNULPathsForV2(t *testing.T) {
	snapshot := t.TempDir()
	writeFiles(t, snapshot, map[string][]byte{
		"assets/a.bin": []byte("prefix\x00suffix"),
	})
	cfg := newCfg(t, "advisory", "off")
	subject := Subject{Name: "nul-tree", Snapshot: snapshot, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(), HashVersion: hashing.VersionV2}
	report, err := auditSubjectWithOpaquePaths(cfg, subject, false, []string{"assets/a.bin"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionAllow || len(report.Findings) != 0 {
		t.Fatalf("v2 report with handed NUL paths %+v, want an allow with no findings", report)
	}
	wantV2, err := hashing.ContentSHA256WithVersion(snapshot, nil, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if report.ContentSHA256 != wantV2 {
		t.Fatalf("v2 report identity = %s, want %s", report.ContentSHA256, wantV2)
	}
}
