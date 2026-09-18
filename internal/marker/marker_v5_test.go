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

// v5GitMarker returns a minimal valid network-git draft marker.
func v5GitMarker() *Marker {
	m := v5LocalMarker()
	m.Package = &Package{
		Kind:       "network-git",
		Repository: "example.org/kit",
		Commit:     &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("44", 20)},
		Directory:  "skills/review",
	}
	return m
}

// rewriteV5Package writes a recorded v5 marker back with its raw package
// object replaced by the given JSON text, bypassing Write so the document
// carries exactly the bytes a foreign or hostile writer could produce.
func rewriteV5Package(t *testing.T, dir string, packageJSON string) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(dir, Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	raw["package"] = json.RawMessage(packageJSON)
	rewritten, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	// json.Marshal re-encodes RawMessage members verbatim only when they
	// are valid JSON; a duplicate-key row is valid JSON to the encoder and
	// survives as written.
	if err := os.WriteFile(filepath.Join(dir, Name), rewritten, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestMarkerV5PackageClosedShape is the reader/currentness regression for
// review finding F1 (revision 3): a package object must match the closed
// raw shape of its arm before decoding. Foreign-arm members are refused
// whether they carry null or an empty value, because the decoded Package
// value collapses both to "absent".
func TestMarkerV5PackageClosedShape(t *testing.T) {
	const (
		snapshot = "sha256:abababababababababababababababababababababababababababababababab"
		hex40    = "4444444444444444444444444444444444444444"
	)
	localValid := `{"kind":"local-snapshot","snapshot":"` + snapshot + `"}`
	gitValid := `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":"skills/review"}`
	configuredValid := `{"kind":"configured-git","source":"vendor/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":"."}`

	type row struct {
		arm     string // "local" or "git": which recorded marker to start from
		pkg     string
		current bool // whether Current against the original expectation must stay true
		readOK  bool
	}
	rows := map[string]row{
		// Valid controls per arm: the closed-shape check admits exactly the
		// arm's members and Current stays true.
		"local-control":               {"local", localValid, true, true},
		"git-control":                 {"git", gitValid, true, true},
		"configured-control-readable": {"git", configuredValid, false, true},

		// local-snapshot: every foreign-arm member with null and with an
		// empty value is refused.
		"local-source-null":      {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","source":null}`, false, false},
		"local-source-empty":     {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","source":""}`, false, false},
		"local-repository-null":  {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","repository":null}`, false, false},
		"local-repository-empty": {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","repository":""}`, false, false},
		"local-directory-null":   {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","directory":null}`, false, false},
		"local-directory-empty":  {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","directory":""}`, false, false},
		"local-commit-null":      {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","commit":null}`, false, false},
		"local-commit-empty":     {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","commit":{}}`, false, false},
		"local-snapshot-null":    {"local", `{"kind":"local-snapshot","snapshot":null}`, false, false},
		"local-snapshot-missing": {"local", `{"kind":"local-snapshot"}`, false, false},
		"local-unknown-member":   {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","extra":1}`, false, false},
		"local-duplicate-key":    {"local", `{"kind":"local-snapshot","snapshot":"` + snapshot + `","snapshot":"` + snapshot + `"}`, false, false},
		"local-kind-null":        {"local", `{"kind":null,"snapshot":"` + snapshot + `"}`, false, false},
		"local-kind-number":      {"local", `{"kind":5,"snapshot":"` + snapshot + `"}`, false, false},
		"local-package-null":     {"local", `null`, false, false},
		"local-package-array":    {"local", `[]`, false, false},

		// network-git: snapshot/source foreign members, nulls in required
		// members, and an open commit object are refused.
		"git-snapshot-null":         {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":"skills/review","snapshot":null}`, false, false},
		"git-snapshot-empty":        {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":"skills/review","snapshot":""}`, false, false},
		"git-source-null":           {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":"skills/review","source":null}`, false, false},
		"git-source-empty":          {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":"skills/review","source":""}`, false, false},
		"git-directory-null":        {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":null}`, false, false},
		"git-directory-missing":     {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"}}`, false, false},
		"git-commit-null":           {"git", `{"kind":"network-git","repository":"example.org/kit","commit":null,"directory":"skills/review"}`, false, false},
		"git-commit-string":         {"git", `{"kind":"network-git","repository":"example.org/kit","commit":"` + hex40 + `","directory":"skills/review"}`, false, false},
		"git-commit-hex-null":       {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":null},"directory":"skills/review"}`, false, false},
		"git-commit-extra-member":   {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `","ref":"main"},"directory":"skills/review"}`, false, false},
		"git-commit-missing-format": {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"hex":"` + hex40 + `"},"directory":"skills/review"}`, false, false},
		"git-commit-duplicate-key":  {"git", `{"kind":"network-git","repository":"example.org/kit","commit":{"object_format":"sha1","object_format":"sha1","hex":"` + hex40 + `"},"directory":"skills/review"}`, false, false},

		// configured-git: repository/snapshot foreign members refused.
		"configured-repository-null":  {"git", `{"kind":"configured-git","source":"vendor/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":".","repository":null}`, false, false},
		"configured-repository-empty": {"git", `{"kind":"configured-git","source":"vendor/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":".","repository":""}`, false, false},
		"configured-snapshot-null":    {"git", `{"kind":"configured-git","source":"vendor/kit","commit":{"object_format":"sha1","hex":"` + hex40 + `"},"directory":".","snapshot":null}`, false, false},
	}
	for name, r := range rows {
		t.Run(name, func(t *testing.T) {
			var m *Marker
			if r.arm == "local" {
				m = v5LocalMarker()
			} else {
				m = v5GitMarker()
			}
			dir, recorded := writeV5(t, m)
			rewriteV5Package(t, dir, r.pkg)
			got := Read(dir)
			if (got != nil) != r.readOK {
				t.Fatalf("Read = %v, want readable=%v for package %s", got, r.readOK, r.pkg)
			}
			expected := *recorded
			current, _ := Current(dir, &expected)
			if current != r.current {
				t.Fatalf("Current = %v, want %v for package %s", current, r.current, r.pkg)
			}
			if got == nil {
				// A refused document must never satisfy currentness for
				// any expectation, including one that mirrors its own
				// malformed value.
				var malformed Package
				_ = json.Unmarshal([]byte(r.pkg), &malformed)
				expected.Package = &malformed
				if current, _ := Current(dir, &expected); current {
					t.Fatalf("Current = true for a refused package %s", r.pkg)
				}
			}
		})
	}
}
