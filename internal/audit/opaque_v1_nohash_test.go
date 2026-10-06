// Rework-2 regressions for revision-3 finding F3: the v1 opaque
// refusal is observed to precede hashing, not merely to return an empty
// digest. Each test below drives a guarded production entry over a
// NUL-bearing tree inside hashing.CountV1Hashes and requires zero v1
// computations; the clean control in the same test requires at least
// one, proving the seam is wired. The narrowing mutant — a discarded
// ContentSHA256WithVersion(snapshot, nil, VersionV1) inside the refusal
// branch, or the live hash moved above the guard — computes one v1
// identity and fails every zero assertion here.
package audit

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

// The gate refuses a v1 NUL tree with the opaque finding whether audit
// is enabled or not, and neither path computes a v1 identity. The
// enabled clean control hashes, proving the observation is live.
func TestGateV1NULRefusalComputesNoV1Identity(t *testing.T) {
	nulTree := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{
		"scripts/tool": []byte("echo ok\n"),
		"assets/a.bin": []byte("x\x00y"),
	})
	subject := Subject{Name: "nul-tree", Snapshot: nulTree, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone()}

	disabled := newCfg(t, "advisory", "off")
	disabled.Audit.Enabled = false
	var disabledErrs []string
	if calls := hashing.CountV1Hashes(func() {
		_, disabledErrs = Gate(disabled, []Subject{subject})
	}); calls != 0 {
		t.Fatalf("disabled-audit v1 NUL gate computed %d v1 identities, want 0", calls)
	}
	if len(disabledErrs) != 1 || !strings.Contains(disabledErrs[0], "audit.opaque.nul-byte") {
		t.Fatalf("disabled-audit v1 NUL gate errs = %v, want the opaque refusal", disabledErrs)
	}

	enabled := newCfg(t, "advisory", "off")
	var enabledErrs []string
	if calls := hashing.CountV1Hashes(func() {
		_, enabledErrs = Gate(enabled, []Subject{subject})
	}); calls != 0 {
		t.Fatalf("enabled-audit v1 NUL gate computed %d v1 identities, want 0", calls)
	}
	if len(enabledErrs) != 1 || !strings.Contains(enabledErrs[0], "audit.opaque.nul-byte") {
		t.Fatalf("enabled-audit v1 NUL gate errs = %v, want the opaque refusal", enabledErrs)
	}

	cleanTree := t.TempDir()
	writeFiles(t, cleanTree, map[string][]byte{"scripts/tool": []byte("echo ok\n")})
	clean := Subject{Name: "clean-tree", Snapshot: cleanTree, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone()}
	if calls := hashing.CountV1Hashes(func() {
		if _, errs := Gate(enabled, []Subject{clean}); len(errs) != 0 {
			t.Fatalf("clean v1 gate errs = %v, want admission", errs)
		}
	}); calls == 0 {
		t.Fatal("clean v1 gate observed no v1 hash; the seam is not wired")
	}
}

// The audit pipeline refuses a v1 NUL tree before hashing for both the
// legacy zero version and the explicit v1 version; the refusal carries
// no digest because none was computed. The clean control hashes.
func TestAuditSubjectV1NULRefusalComputesNoV1Identity(t *testing.T) {
	nulTree := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{"assets/a.bin": []byte("x\x00y")})
	cfg := newCfg(t, "advisory", "off")
	for _, version := range []hashing.Version{0, hashing.VersionV1} {
		subject := Subject{Name: "nul-tree", Snapshot: nulTree, SchemaVersion: 3,
			Capabilities: capabilities.ImplicitNone(), HashVersion: version}
		var report Report
		var err error
		if calls := hashing.CountV1Hashes(func() {
			report, err = auditSubject(cfg, subject, true)
		}); calls != 0 {
			t.Fatalf("version %d: v1 NUL audit computed %d v1 identities, want 0", version, calls)
		}
		if err != nil {
			t.Fatal(err)
		}
		if report.Decision != DecisionBlock || report.ContentSHA256 != "" {
			t.Fatalf("version %d: report %+v, want a digest-free block", version, report)
		}
	}

	cleanTree := t.TempDir()
	writeFiles(t, cleanTree, map[string][]byte{"scripts/tool": []byte("echo ok\n")})
	clean := Subject{Name: "clean-tree", Snapshot: cleanTree, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone()}
	if calls := hashing.CountV1Hashes(func() {
		if _, err := auditSubject(cfg, clean, false); err != nil {
			t.Fatal(err)
		}
	}); calls == 0 {
		t.Fatal("clean v1 audit observed no v1 hash; the seam is not wired")
	}
}

// The draft source-audit lane refuses a NUL tree with the opaque finding
// before its live v1 pipeline hashes anything. The clean control
// establishes a binding and hashes.
func TestCheckSourceAuditV1NULRefusalComputesNoV1Identity(t *testing.T) {
	nulTree := t.TempDir()
	writeFiles(t, nulTree, map[string][]byte{"assets/a.bin": []byte("x\x00y")})
	cfg := newCfg(t, "advisory", "off")
	var auditErr error
	if calls := hashing.CountV1Hashes(func() {
		_, auditErr = CheckSourceAudit(cfg, sourceTestSubject(nulTree, 3), true, sourceAuditTestNow)
	}); calls != 0 {
		t.Fatalf("v1 NUL source audit computed %d v1 identities, want 0", calls)
	}
	if auditErr == nil || !strings.Contains(auditErr.Error(), "audit.opaque.nul-byte") {
		t.Fatalf("v1 NUL source audit err = %v, want the opaque refusal", auditErr)
	}

	if calls := hashing.CountV1Hashes(func() {
		if _, err := CheckSourceAudit(cfg, sourceTestSubject(cleanSnapshot(t, "echo ok\n"), 3), true, sourceAuditTestNow); err != nil {
			t.Fatal(err)
		}
	}); calls == 0 {
		t.Fatal("clean v1 source audit observed no v1 hash; the seam is not wired")
	}
}
