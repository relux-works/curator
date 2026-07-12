// Package interop is the conformance gate: Curator must reproduce the
// golden fixtures byte for byte. The expectations under
// testdata/golden/expected were produced once by an independent conforming
// implementation of the protocol; regenerating them is a deliberate act
// (see testdata/golden/README.md).
package interop

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// internal/interop -> repo root
	return filepath.Dir(filepath.Dir(wd))
}

func golden(t *testing.T, rel string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "golden", filepath.FromSlash(rel))
}

func readGolden(t *testing.T, rel string) string {
	t.Helper()
	payload, err := os.ReadFile(golden(t, rel))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(payload))
}

func TestGoldenSnapshotHash(t *testing.T) {
	got, err := hashing.ContentSHA256(golden(t, "skill-fixture"), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := readGolden(t, "expected/snapshot_sha256.txt")
	if got != want {
		t.Fatalf("snapshot hash diverges from golden:\n got %s\nwant %s", got, want)
	}
}

func TestGoldenContextCopy(t *testing.T) {
	fixture := golden(t, "skill-fixture")
	spec, err := skillspec.Load(fixture)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "ctx")
	includeScripts := len(spec.Commands) == 0
	files, err := whitelist.CopyContext(fixture, dest, includeScripts, spec.RuntimeRoots)
	if err != nil {
		t.Fatal(err)
	}

	var wantFiles []string
	if err := json.Unmarshal([]byte(readGolden(t, "expected/context_files.json")), &wantFiles); err != nil {
		t.Fatal(err)
	}
	gotJSON, _ := json.Marshal(files)
	wantJSON, _ := json.Marshal(wantFiles)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("context file list diverges:\n got %s\nwant %s", gotJSON, wantJSON)
	}

	gotHash, err := hashing.ContentSHA256(dest, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantHash := readGolden(t, "expected/context_sha256.txt")
	if gotHash != wantHash {
		t.Fatalf("context hash diverges from golden:\n got %s\nwant %s", gotHash, wantHash)
	}
}

func readSigned(t *testing.T, rel string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(readGolden(t, rel)), &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestGoldenRegistryObjects(t *testing.T) {
	pinned := readGolden(t, "expected/registry/pinned_key.txt")
	keys := []string{pinned}

	audited := readSigned(t, "expected/registry/record_audited.json")
	if !registry.VerifySigned(audited, keys) {
		t.Fatal("golden audited record must verify against the pinned key")
	}
	revoked := readSigned(t, "expected/registry/record_revoked.json")
	if !registry.VerifySigned(revoked, keys) {
		t.Fatal("golden revoked record must verify")
	}
	forged := readSigned(t, "expected/registry/record_forged.json")
	if registry.VerifySigned(forged, keys) {
		t.Fatal("forged record must not verify")
	}
	snapshot := readSigned(t, "expected/registry/snapshot.json")
	if !registry.VerifySigned(snapshot, keys) {
		t.Fatal("golden snapshot must verify")
	}

	// wrong key rejects everything
	other := "ed25519:" + strings.Repeat("A", 43) + "="
	if registry.VerifySigned(audited, []string{other}) {
		t.Fatal("wrong key must not verify")
	}
}

func TestGoldenFederationSemantics(t *testing.T) {
	pinned := readGolden(t, "expected/registry/pinned_key.txt")
	audited := readSigned(t, "expected/registry/record_audited.json")
	revoked := readSigned(t, "expected/registry/record_revoked.json")
	snapHash := readGolden(t, "expected/snapshot_sha256.txt")

	registries := []registry.Registry{
		{Name: "golden-one", URL: "https://one", PublicKeys: []string{pinned}},
		{Name: "golden-two", URL: "https://two", PublicKeys: []string{pinned}},
	}
	fetch := func(url, _, _, _ string) ([]map[string]any, error) {
		if url == "https://one" {
			return []map[string]any{audited}, nil
		}
		return []map[string]any{revoked}, nil
	}
	resolution := registry.Resolve(registries,
		"git.example.com/skills/golden-skill",
		"0123456789abcdef0123456789abcdef01234567",
		snapHash, fetch)
	if resolution.Result != registry.ResultRevoked {
		t.Fatalf("deny-wins over golden records: %+v", resolution)
	}
}
