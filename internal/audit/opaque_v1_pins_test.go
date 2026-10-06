// Rework-2 regressions for revision-3 finding F2: trust pins are a
// versioned carrier. A pin authorizes only the (version, digest)
// identity it was issued for (Spec §8): legacy unversioned pins keep
// authorizing v1 reads only, and equal digest text across framings
// never shares trust. A mutant that ignores the recorded pin version
// authorizes the mismatched subjects below instead of requiring a pin.
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

// A new v2 pin records schema 2 with hash_version 2; v1 pins keep the
// byte-identical frozen schema-1 shape with no hash_version member,
// whether written by the legacy writer or the versioned writer.
func TestPinAtVersionWritesVersionedCarrier(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)

	homeV2 := t.TempDir()
	path, err := PinAtVersion(homeV2, digest, hashing.VersionV2, "v2 approval", "fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw := readPinCarrier(t, path)
	if raw["schema_version"] != float64(2) || raw["hash_version"] != float64(2) {
		t.Fatalf("v2 pin carrier = schema_version %v hash_version %v, want 2/2",
			raw["schema_version"], raw["hash_version"])
	}

	homeV1 := t.TempDir()
	legacyPath, err := Pin(homeV1, digest, "legacy v1 approval", "fixture")
	if err != nil {
		t.Fatal(err)
	}
	assertFrozenV1PinCarrier(t, legacyPath)

	explicitPath, err := PinAtVersion(t.TempDir(), digest, hashing.VersionV1, "explicit v1 approval", "fixture")
	if err != nil {
		t.Fatal(err)
	}
	assertFrozenV1PinCarrier(t, explicitPath)

	if _, err := PinAtVersion(t.TempDir(), digest, hashing.Version(9), "bad", "fixture"); err == nil ||
		!strings.Contains(err.Error(), "unsupported content hash version") {
		t.Fatalf("unknown-version pin = %v, want an unsupported-version refusal", err)
	}
}

func readPinCarrier(t *testing.T, path string) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func assertFrozenV1PinCarrier(t *testing.T, path string) {
	t.Helper()
	raw := readPinCarrier(t, path)
	if raw["schema_version"] != float64(1) {
		t.Fatalf("v1 pin carrier schema_version = %v, want 1", raw["schema_version"])
	}
	if _, present := raw["hash_version"]; present {
		t.Fatalf("v1 pin carrier gained hash_version %v, want the frozen shape", raw["hash_version"])
	}
}

// A legacy v1 pin never authorizes a v2 subject, even at equal digest
// text: the strict schema-2 v2 subject over a NUL tree still requires
// a v2 pin. The equal digest text is a synthetic version-pair fixture,
// not a hash collision.
func TestV2RejectsLegacyV1Pin(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{
		"scripts/tool": []byte("echo ok\n"),
		"assets/a.bin": []byte("x\x00y"),
	})
	digest, err := hashing.ContentSHA256WithVersion(root, nil, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "strict", "high")
	if _, err := Pin(cfg.Home(), digest, "legacy v1 approval", "fixture"); err != nil {
		t.Fatal(err)
	}
	_, errs := Gate(cfg, []Subject{{
		Name: "modern", Snapshot: root, HashVersion: hashing.VersionV2,
		SchemaVersion: 2, Capabilities: capabilities.ImplicitNone(),
	}})
	if len(errs) != 1 || !strings.Contains(errs[0], "requires pin") {
		t.Fatalf("v2 subject with a legacy v1 pin: errs = %v, want exactly the require-pin refusal", errs)
	}
}

// Symmetrically, a v2 pin never authorizes a v1 subject: the strict
// schema-2 v1 subject over a clean tree still requires a v1 pin. The
// tree is clean so the v1 opaque refusal cannot mask the pin decision.
func TestV1RejectsV2Pin(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("echo ok\n")})
	digest, err := hashing.ContentSHA256WithVersion(root, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "strict", "high")
	if _, err := PinAtVersion(cfg.Home(), digest, hashing.VersionV2, "v2 approval", "fixture"); err != nil {
		t.Fatal(err)
	}
	_, errs := Gate(cfg, []Subject{{
		Name: "legacy", Snapshot: root, HashVersion: hashing.VersionV1,
		SchemaVersion: 2, Capabilities: capabilities.ImplicitNone(),
	}})
	if len(errs) != 1 || !strings.Contains(errs[0], "requires pin") {
		t.Fatalf("v1 subject with a v2 pin: errs = %v, want exactly the require-pin refusal", errs)
	}
}

// Matching versions authorize: a v2 pin admits the v2 subject and a v1
// pin admits the v1 subject, both silently under strict audit.
func TestPinVersionMatchAuthorizes(t *testing.T) {
	for _, version := range []hashing.Version{hashing.VersionV1, hashing.VersionV2} {
		root := t.TempDir()
		writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("echo ok\n")})
		digest, err := hashing.ContentSHA256WithVersion(root, nil, version)
		if err != nil {
			t.Fatal(err)
		}
		cfg := newCfg(t, "strict", "high")
		var path string
		if version == hashing.VersionV1 {
			path, err = Pin(cfg.Home(), digest, "v1 approval", "fixture")
		} else {
			path, err = PinAtVersion(cfg.Home(), digest, version, "v2 approval", "fixture")
		}
		if err != nil {
			t.Fatal(err)
		}
		if version == hashing.VersionV1 {
			assertFrozenV1PinCarrier(t, path)
		} else if raw := readPinCarrier(t, path); raw["hash_version"] != float64(2) {
			t.Fatalf("v2 pin carrier hash_version = %v, want 2", raw["hash_version"])
		}
		warnings, errs := Gate(cfg, []Subject{{
			Name: "skill", Snapshot: root, HashVersion: version,
			SchemaVersion: 2, Capabilities: capabilities.ImplicitNone(),
		}})
		if len(warnings) != 0 || len(errs) != 0 {
			t.Fatalf("version %d: warnings=%v errs=%v, want silent admission", version, warnings, errs)
		}
	}
}

// A schemaless legacy record (pinned with no version at all) reads as
// v1 only: it authorizes the v1 subject and never the v2 subject.
func TestLegacySchemalessPinReadsAsV1(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("echo ok\n")})
	v1Digest, err := hashing.ContentSHA256WithVersion(root, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	v2Digest, err := hashing.ContentSHA256WithVersion(root, nil, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "strict", "high")
	for _, digest := range []string{v1Digest, v2Digest} {
		dir := filepath.Join(cfg.Home(), "audit", hashing.Normalize(digest))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		payload, err := json.Marshal(map[string]any{
			"content_sha256": strings.ToLower(digest),
			"pinned":         true,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "trust.json"), append(payload, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if !isPinned(cfg, v1Digest, hashing.VersionV1) || !isPinned(cfg, v1Digest, 0) {
		t.Fatal("schemaless pin must authorize v1 reads, including the zero-version reader")
	}
	if isPinned(cfg, v2Digest, hashing.VersionV2) {
		t.Fatal("schemaless pin must never authorize a v2 read")
	}
	if isPinned(cfg, v1Digest, hashing.Version(9)) {
		t.Fatal("unknown-version read must never report pinned")
	}
}
