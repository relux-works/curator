// Package interop consumes the authoritative external Curator Protocol suite.
// It contains no implementation-owned expected values.
package interop

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
)

func suiteRoot(t *testing.T) string {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(absolute, "manifest.json")); err != nil {
		t.Fatalf("invalid CURATOR_CONFORMANCE_ROOT: %v", err)
	}
	return absolute
}

func TestGoldenMarkerObject(t *testing.T) {
	dir := t.TempDir()
	wantContextHash := readGolden(t, "expected/context_sha256.txt")
	wantFiles := []string{".skill_triggers/en.md", "SKILL.md", "references/notes.md"}
	generated := &marker.Marker{
		Name: "golden-skill", Source: "golden-skill",
		RefKind: "revision", Ref: "0123456789abcdef0123456789abcdef01234567",
		Commit: "0123456789abcdef0123456789abcdef01234567", ContentSHA256: wantContextHash,
		Agents: []string{"codex_cli"}, Commands: []string{"golden-tool"},
		SkillSchemaVersion: 5, RuntimeRoots: []string{"scripts"},
		InstalledAt: "2000-01-01T00:00:00Z", Files: wantFiles,
		Activation: &marker.Activation{Context: true, Commands: []string{"golden-tool"}},
		Requirers:  []string{"<project>"},
	}
	if err := marker.Write(dir, generated); err != nil {
		t.Fatal(err)
	}
	actualPayload, err := os.ReadFile(filepath.Join(dir, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	wantPayload, err := os.ReadFile(golden(t, "expected/marker.json"))
	if err != nil {
		t.Fatal(err)
	}
	var actual, want any
	if err := json.Unmarshal(actualPayload, &actual); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(wantPayload, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("marker object diverges from golden:\n got %s\nwant %s", actualPayload, wantPayload)
	}
}

func golden(t *testing.T, rel string) string {
	t.Helper()
	return filepath.Join(suiteRoot(t), filepath.FromSlash(rel))
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
	got, err := hashing.ContentSHA256(golden(t, "fixtures/skill"), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := readGolden(t, "expected/snapshot_sha256.txt")
	if got != want {
		t.Fatalf("snapshot hash diverges from golden:\n got %s\nwant %s", got, want)
	}
}

func TestGoldenContextCopy(t *testing.T) {
	fixture := golden(t, "fixtures/skill")
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
	decoder := json.NewDecoder(bytes.NewReader([]byte(readGolden(t, rel))))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
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
	wrongKeyID := readSigned(t, "expected/registry/record_wrong_key_id.json")
	if registry.VerifySigned(wrongKeyID, keys) {
		t.Fatal("record with a wrong key id must not verify")
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

func TestCanonicalJSONVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/canonical-valid.json"))
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var cases []struct {
		Name      string         `json:"name"`
		Input     map[string]any `json:"input"`
		Canonical string         `json:"canonical_utf8"`
	}
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		got, err := registry.CanonicalBytesChecked(testCase.Input)
		if err != nil {
			t.Fatalf("%s: %v", testCase.Name, err)
		}
		if string(got) != testCase.Canonical {
			t.Fatalf("%s canonical bytes:\n got %s\nwant %s", testCase.Name, got, testCase.Canonical)
		}
	}
}

func TestSourceIdentityVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/source-identities.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input    string  `json:"input"`
		Identity *string `json:"identity"`
		Error    string  `json:"error"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		got, err := identity.Parse(testCase.Input)
		if testCase.Error != "" {
			if err == nil {
				t.Errorf("Parse(%q) = %q, want error %s", testCase.Input, got, testCase.Error)
			}
			continue
		}
		if err != nil {
			t.Errorf("Parse(%q): %v", testCase.Input, err)
			continue
		}
		want := ""
		if testCase.Identity != nil {
			want = *testCase.Identity
		}
		if got != want {
			t.Errorf("Parse(%q) = %q, want %q", testCase.Input, got, want)
		}
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
