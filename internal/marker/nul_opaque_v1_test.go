// Rework-1 regressions for the Spec §8 version rule at the installed-tree
// currentness reader: a recorded v1 identity is never recomputed or
// trusted over NUL bytes, while a recorded v2 identity treats NUL as
// ordinary data. The collision case adapts the revision-1 reviewer's
// attack probe into maintained coverage.
package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
)

func writeMarkerTestFiles(t *testing.T, root string, files map[string][]byte) {
	t.Helper()
	for rel, payload := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// A legacy marker with no hash_version over a clean two-record tree reads
// current; after the files are replaced with their NUL-bearing one-record
// v1 collision, Current refuses with the opaque finding instead of
// trusting the colliding v1 digest. These are real equal-v1-digest trees,
// not assumed collisions.
func TestCurrentLegacyMarkerNULCollisionRefuses(t *testing.T) {
	enableV1Writers(t)
	root := t.TempDir()
	writeMarkerTestFiles(t, root, map[string][]byte{
		"assets/a.bin": []byte("x"),
		"docs/b.md":    []byte("y"),
	})
	digest, err := hashing.ContentSHA256(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := &Marker{Name: "test-skill", Source: "test-skill", RefKind: "tag", Ref: "v1",
		Commit: strings.Repeat("1", 40), ContentSHA256: digest,
		InstalledAt: "2026-10-06T00:00:00Z", SkillSchemaVersion: 3}
	if err := Write(root, m); err != nil {
		t.Fatal(err)
	}
	if current, err := Current(root, m); err != nil || !current {
		t.Fatalf("clean control current=%v err=%v, want true/nil", current, err)
	}
	writeMarkerTestFiles(t, root, map[string][]byte{
		"assets/a.bin": []byte("x\x00docs/b.md\x00y"),
	})
	if err := os.Remove(filepath.Join(root, "docs", "b.md")); err != nil {
		t.Fatal(err)
	}
	current, err := Current(root, m)
	if current || err == nil || !strings.Contains(err.Error(), opaquescan.FindingNUL) {
		t.Fatalf("v1 installed NUL collision: current=%v err=%v, want a pre-hash opaque refusal", current, err)
	}
}

// A recorded v2 identity over a NUL-bearing tree stays verifiable: the
// currentness reader recomputes in v2 and reports current.
func TestCurrentV2MarkerAdmitsNULBearingTree(t *testing.T) {
	enableV2Writers(t)
	root := t.TempDir()
	writeMarkerTestFiles(t, root, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	m := validMarkerV2()
	m.SkillSchemaVersion = 6
	if err := Write(root, m); err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != SchemaV5 || m.HashVersion != hashing.VersionV2 {
		t.Fatalf("marker = schema:%d hash_version:%d, want a recorded v2 identity", m.SchemaVersion, m.HashVersion)
	}
	if current, err := Current(root, m); err != nil || !current {
		t.Fatalf("v2 NUL current=%v err=%v, want true/nil", current, err)
	}
}

// A v2 marker file edited down to hash_version 1 is refused by the state
// reader: a v2-to-v1 downgrade never becomes a trusted v1 identity.
func TestReadStateRefusesV2ToV1Downgrade(t *testing.T) {
	enableV2Writers(t)
	root := t.TempDir()
	writeMarkerTestFiles(t, root, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	m := validMarkerV2()
	m.SkillSchemaVersion = 6
	if err := Write(root, m); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(root, Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	raw["hash_version"] = 1
	payload, err = json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, Name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadState(root); err == nil {
		t.Fatal("state reader admitted a hash_version 1 downgrade of a v2 marker")
	}
}

// Rework-2 regression for revision-3 finding F3 at the installed-tree
// currentness reader: the v1 opaque refusal is observed to precede
// hashing. The clean control hashes, proving the seam is wired; a
// discarded hash-before-refuse mutant computes one v1 identity and
// fails the zero assertion.
func TestCurrentV1NULRefusalComputesNoV1Identity(t *testing.T) {
	enableV1Writers(t)
	root := t.TempDir()
	writeMarkerTestFiles(t, root, map[string][]byte{
		"assets/a.bin": []byte("x"),
		"docs/b.md":    []byte("y"),
	})
	digest, err := hashing.ContentSHA256(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := &Marker{Name: "test-skill", Source: "test-skill", RefKind: "tag", Ref: "v1",
		Commit: strings.Repeat("1", 40), ContentSHA256: digest,
		InstalledAt: "2026-10-06T00:00:00Z", SkillSchemaVersion: 3}
	if err := Write(root, m); err != nil {
		t.Fatal(err)
	}
	if calls := hashing.CountV1Hashes(func() {
		if current, err := Current(root, m); err != nil || !current {
			t.Fatalf("clean control current=%v err=%v, want true/nil", current, err)
		}
	}); calls == 0 {
		t.Fatal("clean v1 currentness observed no v1 hash; the seam is not wired")
	}
	writeMarkerTestFiles(t, root, map[string][]byte{
		"assets/a.bin": []byte("x\x00docs/b.md\x00y"),
	})
	if err := os.Remove(filepath.Join(root, "docs", "b.md")); err != nil {
		t.Fatal(err)
	}
	var current bool
	var currentErr error
	if calls := hashing.CountV1Hashes(func() {
		current, currentErr = Current(root, m)
	}); calls != 0 {
		t.Fatalf("v1 NUL currentness computed %d v1 identities, want 0", calls)
	}
	if current || currentErr == nil || !strings.Contains(currentErr.Error(), opaquescan.FindingNUL) {
		t.Fatalf("v1 installed NUL collision: current=%v err=%v, want a pre-hash opaque refusal", current, currentErr)
	}
}
