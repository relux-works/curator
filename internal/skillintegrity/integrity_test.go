package skillintegrity

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/skillspec"
)

func fixture(t *testing.T) (string, Identity) {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, root, "SKILL.md", []byte("---\nname: fixture-skill\ndescription: Fixture skill.\n---\nFixture context.\n"), 0o644)
	writeFixture(t, root, "references/guide.md", []byte("Fixture reference."), 0o644)
	return root, Identity{Name: "fixture-skill", Source: "https://example.org/skill.git", Commit: strings.Repeat("a", 40)}
}
func writeFixture(t *testing.T, root, path string, data []byte, mode os.FileMode) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(full, mode); err != nil {
		t.Fatal(err)
	}
}

// Native canonical/legacy/runtime fallback and manifestless skills all receive
// complete evidence, without writing a synthesized declaration into the source.
// Canonical encoding binds the returned evidence and survives the strict reader.
func TestBuildAndVerifyEvidence(t *testing.T) {
	for _, path := range []string{"", "agent-skill.json", "csk-skill.json", "agents/runtime.json"} {
		t.Run(path, func(t *testing.T) {
			root, id := fixture(t)
			if path != "" {
				data := []byte(syntheticDeclaration)
				if path == "agents/runtime.json" {
					data = []byte(`{"commands":{}}`)
				}
				writeFixture(t, root, path, data, 0o644)
			}
			e, err := Build(context.Background(), root, id, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(e.Manifest.Files) != 2+btoi(path != "") || e.Manifest.DeclarationPath != path {
				t.Fatalf("unexpected roster/declaration: %+v", e.Manifest)
			}
			if path == "" {
				if e.Manifest.ManifestKind != "synthesized" || !bytes.Equal(e.Manifest.Declaration, []byte(syntheticDeclaration)) {
					t.Fatal("manifestless declaration gained authority")
				}
				if _, err := os.Lstat(filepath.Join(root, "agent-skill.json")); !os.IsNotExist(err) {
					t.Fatal("build wrote a native manifest")
				}
			}
			raw, err := protocoljson.MarshalCanonical(e.Object())
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeEntry(raw, id)
			if err != nil {
				t.Fatal(err)
			}
			if err := Verify(context.Background(), root, id, decoded, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func btoi(value bool) int {
	if value {
		return 1
	}
	return 0
}

// Native runtime commands remain native and their exact declaration bytes are
// bound beside reference assets and executable metadata. Matching dual native
// manifests retain the canonical precedence rather than becoming synthetic.
func TestNativeDeclarationsPreserveAuthority(t *testing.T) {
	root, id := fixture(t)
	declaration := []byte(`{"commands":{"hello":"scripts/hello.sh"}}`)
	writeFixture(t, root, "agents/runtime.json", declaration, 0o644)
	writeFixture(t, root, "scripts/hello.sh", []byte("#!/bin/sh\nexit 0\n"), 0o755)
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.Manifest.ManifestKind != "native" || e.Manifest.DeclarationSHA256 != digest(declaration) || len(e.Manifest.Files) != 4 {
		t.Fatal("lost native declaration or complete roster")
	}
	spec, err := skillspec.Load(root)
	if err != nil || spec.Commands["hello"].UnixPath != "scripts/hello.sh" {
		t.Fatalf("native commands changed: %+v,%v", spec, err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "agents/runtime.json"))
	if !bytes.Equal(after, declaration) {
		t.Fatal("build changed native declaration")
	}
	root, id = fixture(t)
	writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o644)
	writeFixture(t, root, "csk-skill.json", []byte(syntheticDeclaration), 0o644)
	e, err = Build(context.Background(), root, id, nil)
	if err != nil || e.Manifest.DeclarationPath != "agent-skill.json" {
		t.Fatalf("matching native manifests: %+v,%v", e, err)
	}
}

// A present unreadable manifest is unavailable evidence, never a synthesized
// declaration. Root execution cannot establish the Unix mode-denial witness.
func TestUnreadableNativeIsNotAbsence(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("fixture requires Unix permission denial without root bypass")
	}
	root, id := fixture(t)
	path := filepath.Join(root, "agent-skill.json")
	writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o000)
	defer func() { _ = os.Chmod(path, 0o644) }()
	e, err := Build(context.Background(), root, id, nil)
	if !IsCode(err, "skill_integrity_unavailable") || e.Manifest.ManifestKind != "" {
		t.Fatalf("unreadable native became evidence: %+v,%v", e, err)
	}
}

// Roster additions/removals, content changes, execute-bit drift and an added
// native declaration each refuse against previously valid evidence.
func TestEvidenceRejectsEveryDriftDirection(t *testing.T) {
	for _, change := range []string{"add", "remove", "edit", "execute", "native-substitution"} {
		t.Run(change, func(t *testing.T) {
			root, id := fixture(t)
			e, err := Build(context.Background(), root, id, nil)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "add":
				writeFixture(t, root, "references/extra.md", []byte("extra"), 0o644)
			case "remove":
				if err := os.Remove(filepath.Join(root, "references/guide.md")); err != nil {
					t.Fatal(err)
				}
			case "edit":
				writeFixture(t, root, "references/guide.md", []byte("changed"), 0o644)
			case "execute":
				if err := os.Chmod(filepath.Join(root, "references/guide.md"), 0o755); err != nil {
					t.Fatal(err)
				}
				observed, err := os.Stat(filepath.Join(root, "references/guide.md"))
				if err != nil {
					t.Fatal(err)
				}
				if observed.Mode().Perm()&0o111 == 0 {
					t.Skip("filesystem cannot materialize the executable-bit mutation")
				}
			case "native-substitution":
				writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o644)
			}
			if err := Verify(context.Background(), root, id, e, nil); !IsCode(err, "skill_integrity_mismatch") {
				t.Fatalf("%s drift: %v", change, err)
			}
		})
	}
}

// Malformed/present metadata cannot become an absent synthetic declaration;
// canonical/legacy disagreement and every payload link boundary refuse.
func TestEvidenceStructuralRefusals(t *testing.T) {
	for _, kind := range []string{"bad-manifest", "dual-conflict", "symlink", "hardlink", "nested-git", "case-collision"} {
		t.Run(kind, func(t *testing.T) {
			root, id := fixture(t)
			want := "skill_integrity_boundary_invalid"
			switch kind {
			case "bad-manifest":
				writeFixture(t, root, "agent-skill.json", []byte(`{"schema_version":999}`), 0o644)
				want = "skill_integrity_manifest_invalid"
			case "dual-conflict":
				writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o644)
				writeFixture(t, root, "csk-skill.json", []byte(`{"schema_version":1}`), 0o644)
				want = "skill_integrity_manifest_invalid"
			case "symlink":
				if err := os.Symlink("references/guide.md", filepath.Join(root, "linked.md")); err != nil {
					t.Fatal(err)
				}
			case "hardlink":
				if err := os.Link(filepath.Join(root, "references/guide.md"), filepath.Join(root, "linked.md")); err != nil {
					t.Fatal(err)
				}
			case "nested-git":
				writeFixture(t, root, "nested/.git/config", []byte("fixture"), 0o644)
			case "case-collision":
				writeFixture(t, root, "references/GUIDE.md", []byte("other"), 0o644)
				// A case-insensitive filesystem cannot materialize the forbidden
				// pair. Its independent evidence-parser control still runs below.
				entries, _ := os.ReadDir(filepath.Join(root, "references"))
				if len(entries) == 1 {
					t.Skip("filesystem cannot represent the colliding pair")
				}
			}
			if _, err := Build(context.Background(), root, id, nil); !IsCode(err, want) {
				t.Fatalf("%s: %v", kind, err)
			}
		})
	}
}

// Evidence JSON is closed, required and bound to the original identity/hash.
// Missing/nullable booleans and identity fields cannot acquire zero defaults.
func TestEvidenceJSONRefusals(t *testing.T) {
	root, id := fixture(t)
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(e)
	for _, kind := range []string{"unknown", "missing-executable", "null-executable", "null-source", "changed-identity", "changed-hash", "case-collision", "nonportable-path"} {
		t.Run(kind, func(t *testing.T) {
			var object map[string]any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&object); err != nil {
				t.Fatal(err)
			}
			m := object["manifest"].(map[string]any)
			files := m["files"].([]any)
			first := files[0].(map[string]any)
			switch kind {
			case "nonportable-path":
				first["path"] = "notes:local.txt"
			case "unknown":
				object["extra"] = true
			case "missing-executable":
				delete(first, "executable")
			case "null-executable":
				first["executable"] = nil
			case "null-source":
				m["source"] = nil
			case "changed-identity":
				m["commit"] = strings.Repeat("b", 40)
			case "changed-hash":
				object["manifest_sha256"] = "sha256:" + strings.Repeat("0", 64)
			case "case-collision":
				second := files[1].(map[string]any)
				second["path"] = "skill.md"
			}
			if kind != "changed-hash" {
				canonical, err := protocoljson.MarshalCanonical(m)
				if err != nil {
					t.Fatal(err)
				}
				object["manifest_sha256"] = digest(canonical)
			}
			if kind == "null-executable" {
				object["manifest_sha256"] = e.ManifestSHA256
			}
			changed, _ := json.Marshal(object)
			if _, err := DecodeEntry(changed, id); !IsCode(err, "skill_integrity_manifest_invalid") {
				t.Fatalf("%s: %v", kind, err)
			}
		})
	}
}

// A real empty file supplies the valid zero-size control. Alternate numeric
// spellings must refuse even when coercion would reproduce its valid hash.
func TestEvidenceSizeTokenGrammar(t *testing.T) {
	root, id := fixture(t)
	writeFixture(t, root, "references/guide.md", nil, 0o644)
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeEntry(raw, id); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"-0", "0.0", "0e0"} {
		changed := bytes.Replace(raw, []byte(`"size":0`), []byte(`"size":`+token), 1)
		if bytes.Equal(raw, changed) {
			t.Fatal("control plant did not apply")
		}
		if _, err := DecodeEntry(changed, id); !IsCode(err, "skill_integrity_manifest_invalid") {
			t.Fatalf("size %s: %v", token, err)
		}
	}
}

// Shared closure budgets cannot reset at each skill. Cancellation does not
// synthesize evidence, and an observed mismatch survives later unavailability.
func TestEvidenceBudgetsAndInterruptedMismatch(t *testing.T) {
	root, id := fixture(t)
	budget := NewBudget()
	budget.files = MaxFiles
	if _, err := Build(context.Background(), root, id, budget); !IsCode(err, "skill_integrity_manifest_invalid") {
		t.Fatalf("object budget: %v", err)
	}
	budget = NewBudget()
	budget.bytes = MaxBytes
	if _, err := Build(context.Background(), root, id, budget); !IsCode(err, "skill_integrity_manifest_invalid") {
		t.Fatalf("byte budget: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Build(ctx, root, id, nil); !IsCode(err, "skill_integrity_unavailable") {
		t.Fatalf("cancellation: %v", err)
	}
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, root, "SKILL.md", []byte("changed first file"), 0o644)
	// This fixture context cancels at the fifth deadline observation: after
	// SKILL.md has been captured, before the next directory is inspected.
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	controlled := &cancelAfterChecks{Context: ctx2, cancel: cancel2, after: 5}
	err = Verify(controlled, root, id, e, nil)
	if !IsCode(err, "skill_integrity_mismatch") || !IsCode(err, "skill_integrity_unavailable") {
		t.Fatalf("lost prior mismatch/unavailability: %v", err)
	}
}

type cancelAfterChecks struct {
	context.Context
	cancel       context.CancelFunc
	calls, after int
}

func (c *cancelAfterChecks) Err() error {
	c.calls++
	if c.calls == c.after {
		c.cancel()
	}
	return c.Context.Err()
}

// The resolver's admitted legacy root form permits bare SKILL.md. Applying a
// directory selector invokes the existing stricter metadata contract instead.
func TestLegacyRootMetadataCompatibility(t *testing.T) {
	root, id := fixture(t)
	writeFixture(t, root, "SKILL.md", []byte("# Existing legacy context\n"), 0o644)
	if _, err := Build(context.Background(), root, id, nil); err != nil {
		t.Fatalf("legacy valid form rejected: %v", err)
	}
	id.Directory = "skills/fixture-skill"
	if _, err := Build(context.Background(), root, id, nil); !IsCode(err, "skill_integrity_manifest_invalid") {
		t.Fatalf("selected invalid metadata: %v", err)
	}
	writeFixture(t, root, "SKILL.md", []byte("---\nname: fixture-skill\ndescription: Valid selected context.\n---\nContext.\n"), 0o644)
	if _, err := Build(context.Background(), root, id, nil); err != nil {
		t.Fatalf("selected valid form rejected: %v", err)
	}
}
