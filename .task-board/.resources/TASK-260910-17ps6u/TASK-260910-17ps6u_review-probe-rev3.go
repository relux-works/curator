package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

// v5LocalMarker returns a minimal valid local-snapshot draft marker. The
// caller stamps content fields (content hash, files, installed_at) through
// Write like production staging does.
func v5LocalMarker() *Marker {
	return &Marker{
		Name:               "review",
		Package:            &Package{Kind: "local-snapshot", Snapshot: "sha256:" + strings.Repeat("ab", 32)},
		LockSHA256:         "sha256:" + strings.Repeat("cd", 32),
		ContentSHA256:      "sha256:" + strings.Repeat("ef", 32),
		Locale:             "en",
		Agents:             []string{"claude_code"},
		Commands:           []string{"tool"},
		Dependencies:       []string{},
		SkillSchemaVersion: 4,
		RuntimeRoots:       []string{"scripts"},
		BuildRoots:         []string{},
		InstalledAt:        "2026-09-18T00:00:00Z",
		Files:              []string{"SKILL.md"},
		Builds:             map[string]Build{},
		Activation:         &Activation{Context: true, Commands: []string{"tool"}},
		Requirers:          []string{"<project>"},
	}
}

func writeV5(t *testing.T, m *Marker) (string, *Marker) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	m.ContentSHA256 = hash
	m.Files = []string{"SKILL.md", "references/info.md"}
	if err := Write(dir, m); err != nil {
		t.Fatalf("Write = %v", err)
	}
	recorded := Read(dir)
	if recorded == nil {
		t.Fatalf("Read returned nil for a freshly written v5 marker")
	}
	return dir, recorded
}

func TestMarkerV5LocalRoundTrip(t *testing.T) {
	dir, recorded := writeV5(t, v5LocalMarker())
	_ = dir
	if recorded.SchemaVersion != SchemaV5 {
		t.Fatalf("schema = %d, want %d", recorded.SchemaVersion, SchemaV5)
	}
	if recorded.Package == nil || recorded.Package.Kind != "local-snapshot" ||
		recorded.Package.Snapshot != "sha256:"+strings.Repeat("ab", 32) {
		t.Fatalf("package = %+v", recorded.Package)
	}
	if recorded.LockSHA256 != "sha256:"+strings.Repeat("cd", 32) {
		t.Fatalf("lock = %q", recorded.LockSHA256)
	}
	// The replaced legacy identity is absent, never empty.
	payload, err := os.ReadFile(filepath.Join(dir, Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"source", "git", "ref_kind", "ref", "commit", "substituted", "attestation"} {
		if _, present := raw[field]; present {
			t.Fatalf("v5 marker carries replaced/forbidden field %q", field)
		}
	}
}

func TestMarkerV5Currentness(t *testing.T) {
	m := v5LocalMarker()
	dir, recorded := writeV5(t, m)
	// An identical restaging is current.
	fresh := v5LocalMarker()
	fresh.ContentSHA256 = recorded.ContentSHA256
	fresh.Files = append([]string(nil), recorded.Files...)
	current, err := Current(dir, fresh)
	if err != nil || !current {
		t.Fatalf("Current = %v, %v, want true", current, err)
	}
	changedLock := v5LocalMarker()
	changedLock.LockSHA256 = "sha256:" + strings.Repeat("00", 32)
	if current, _ := Current(dir, changedLock); current {
		t.Fatalf("Current stayed true across a lock generation change")
	}
	changedPackage := v5LocalMarker()
	changedPackage.Package = &Package{Kind: "local-snapshot", Snapshot: "sha256:" + strings.Repeat("11", 32)}
	if current, _ := Current(dir, changedPackage); current {
		t.Fatalf("Current stayed true across a package identity change")
	}
	legacy := v5LocalMarker()
	legacy.Package = nil
	legacy.LockSHA256 = ""
	legacy.Source = "review"
	legacy.RefKind = "tag"
	legacy.Ref = "v1"
	legacy.Commit = strings.Repeat("22", 20)
	if current, _ := Current(dir, legacy); current {
		t.Fatalf("Current matched a v5 marker against a legacy expectation")
	}
}

func TestMarkerV5RefusesMalformedIdentity(t *testing.T) {
	cases := map[string]func(*Marker){
		"missing-package": func(m *Marker) { m.Package = nil },
		"missing-lock":    func(m *Marker) { m.LockSHA256 = "" },
		"malformed-lock":  func(m *Marker) { m.LockSHA256 = "sha256:xyz" },
		"legacy-source":   func(m *Marker) { m.Source = "review" },
		"legacy-ref":      func(m *Marker) { m.RefKind = "tag"; m.Ref = "v1"; m.Commit = strings.Repeat("22", 20) },
		"attestation": func(m *Marker) {
			m.Attestation = &Attestation{Registry: "example.org/kit", Status: "audited"}
		},
		"substituted":      func(m *Marker) { m.Substituted = "dev:review" },
		"bad-snapshot":     func(m *Marker) { m.Package.Snapshot = "bad" },
		"snapshot-and-git": func(m *Marker) { m.Package.Commit = &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("33", 20)} },
		"unknown-kind":     func(m *Marker) { m.Package.Kind = "bad-kind" },
		"skill-schema-0":   func(m *Marker) { m.SkillSchemaVersion = 0 },
		"skill-schema-9":   func(m *Marker) { m.SkillSchemaVersion = 9 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := v5LocalMarker()
			mutate(m)
			if err := Write(t.TempDir(), m); err == nil {
				t.Fatalf("Write admitted a v5 marker with %s", name)
			}
		})
	}
}

func TestMarkerV5GitArmShape(t *testing.T) {
	m := v5LocalMarker()
	m.Package = &Package{
		Kind:       "network-git",
		Repository: "example.org/kit",
		Commit:     &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("44", 20)},
		Directory:  "skills/review",
	}
	if err := Write(t.TempDir(), m); err != nil {
		t.Fatalf("Write refused a structural network-git v5 package: %v", err)
	}
	// A snapshot digest never satisfies the commit arm: the bare hex
	// grammar rejects the prefixed form, so the arms cannot be confused.
	m.Package.Commit = &Commit{ObjectFormat: "sha1", Hex: "sha256:" + strings.Repeat("44", 20)}
	if err := Write(t.TempDir(), m); err == nil {
		t.Fatalf("Write admitted a snapshot digest in the commit arm")
	}
}

func TestReviewV5ForeignNullFields(t *testing.T) {
 for _, field := range []string{"source", "repository", "directory", "commit"} {
  t.Run(field, func(t *testing.T) {
   dir, expected := writeV5(t, v5LocalMarker())
   path := filepath.Join(dir, Name)
   payload, err := os.ReadFile(path); if err != nil { t.Fatal(err) }
   var raw map[string]any
   if err := json.Unmarshal(payload, &raw); err != nil { t.Fatal(err) }
   raw["package"].(map[string]any)[field] = nil
   payload, _ = json.Marshal(raw)
   if err := os.WriteFile(path, payload, 0600); err != nil { t.Fatal(err) }
   if Read(dir) != nil { t.Errorf("local-snapshot marker admitted forbidden package field %s:null", field) }; if current, err := Current(dir, expected); current || err != nil { t.Errorf("Current after foreign field = %v, %v; want false, nil", current, err) }
  })
 }
}
