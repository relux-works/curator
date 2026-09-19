package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// v5GitAttestedMarker returns a network-git draft marker carrying every
// retained migration-table field: the frozen package, the binding lock,
// the registry attestation the effective plan selected and a legacy
// development-substitution identifier. The package replaces the legacy
// source identity on every arm, so none of the five replaced fields is
// set; the declared ref selection lives in the manifest and bound lock.
func v5GitAttestedMarker() *Marker {
	m := v5GitMarker()
	m.Attestation = &Attestation{Registry: "example.org/kit", Status: "audited", KeyID: "0123456789abcdef"}
	m.Substituted = "dev:review"
	m.Requirements = []string{"dep"}
	m.McpServers = map[string][]string{"srv": {"tool"}}
	return m
}

// v5GitAttestedConfiguredMarker is the configured-git arm: the package
// names the configured source path, and the top level carries the same
// attestation and substitution summaries without any legacy identity.
func v5GitAttestedConfiguredMarker() *Marker {
	m := v5GitAttestedMarker()
	m.Package = &Package{
		Kind:      "configured-git",
		Source:    "vendor/kit",
		Commit:    &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("55", 20)},
		Directory: ".",
	}
	return m
}

// TestMarkerV5GitRoundTrip pins the 25-field v4 migration for the Git arms:
// the frozen package binds the selection, the lock binds the generation,
// and the replaced legacy identity is absent — never empty — while every
// retained field keeps its meaning, including attestation and substituted
// for eligible Git packages.
func TestMarkerV5GitRoundTrip(t *testing.T) {
	for name, marker := range map[string]*Marker{
		"network-git":    v5GitAttestedMarker(),
		"configured-git": v5GitAttestedConfiguredMarker(),
	} {
		t.Run(name, func(t *testing.T) {
			dir, recorded := writeV5(t, marker)
			if recorded.SchemaVersion != SchemaV5 {
				t.Fatalf("schema = %d, want %d", recorded.SchemaVersion, SchemaV5)
			}
			if recorded.Package == nil || !packageEqual(recorded.Package, marker.Package) {
				t.Fatalf("package = %+v, want %+v", recorded.Package, marker.Package)
			}
			if recorded.Source != "" || recorded.Git != "" ||
				recorded.RefKind != "" || recorded.Ref != "" || recorded.Commit != "" {
				t.Fatalf("identity = %q %q %s %s %s, want empty",
					recorded.Source, recorded.Git, recorded.RefKind, recorded.Ref, recorded.Commit)
			}
			if recorded.Attestation == nil || *recorded.Attestation != *marker.Attestation {
				t.Fatalf("attestation = %+v, want %+v", recorded.Attestation, marker.Attestation)
			}
			if recorded.Substituted != marker.Substituted {
				t.Fatalf("substituted = %q, want %q", recorded.Substituted, marker.Substituted)
			}
			payload, err := os.ReadFile(filepath.Join(dir, Name))
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(payload, &raw); err != nil {
				t.Fatal(err)
			}
			// The replaced legacy identity is absent, never empty.
			for _, field := range []string{"source", "git", "ref_kind", "ref", "commit"} {
				if _, present := raw[field]; present {
					t.Fatalf("v5 Git marker carries replaced field %q", field)
				}
			}
			// Every retained migration-table field is present.
			for _, field := range []string{
				"schema_version", "name", "package", "lock_sha256", "content_sha256", "locale",
				"agents", "commands", "dependencies", "skill_schema_version", "runtime_roots",
				"build_roots", "builds", "installed_at", "files",
				"requirements", "mcp_servers", "attestation", "activation", "requirers", "substituted",
			} {
				if _, present := raw[field]; !present {
					t.Fatalf("v5 Git marker misses retained field %q", field)
				}
			}
			// An identical restaging is current: the attested Git package
			// with a matching summary satisfies currentness when every
			// other comparison passes (attested-network-current).
			fresh := *marker
			fresh.ContentSHA256 = recorded.ContentSHA256
			fresh.Files = append([]string(nil), recorded.Files...)
			if current, err := Current(dir, &fresh); err != nil || !current {
				t.Fatalf("Current = %v, %v, want true", current, err)
			}
		})
	}
}

func packageEqual(a, b *Package) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind || a.Snapshot != b.Snapshot || a.Repository != b.Repository ||
		a.Source != b.Source || a.Directory != b.Directory {
		return false
	}
	if a.Commit == nil || b.Commit == nil {
		return a.Commit == b.Commit
	}
	return *a.Commit == *b.Commit
}

// TestMarkerV5GitRefusesLegacyIdentityBytes proves the closed v5 shape on
// the Git arms: any of the five replaced legacy fields is refused at the
// writer, and the reader refuses the same shapes from foreign bytes that
// bypassed Write — as a string, as null, or as an empty value — and never
// reports such a document current.
func TestMarkerV5GitRefusesLegacyIdentityBytes(t *testing.T) {
	commit := strings.Repeat("44", 20)
	writeCases := map[string]func(*Marker){
		"source":   func(m *Marker) { m.Source = "review" },
		"git":      func(m *Marker) { m.Git = "https://example.org/kit.git" },
		"ref_kind": func(m *Marker) { m.RefKind = "tag" },
		"ref":      func(m *Marker) { m.Ref = "v1" },
		"commit":   func(m *Marker) { m.Commit = commit },
	}
	for name, mutate := range writeCases {
		t.Run("write/"+name, func(t *testing.T) {
			for _, base := range []*Marker{v5GitAttestedMarker(), v5GitAttestedConfiguredMarker()} {
				m := *base
				mutate(&m)
				if err := Write(t.TempDir(), &m); err == nil {
					t.Fatalf("Write admitted a v5 Git marker carrying %s", name)
				}
			}
		})
	}
	stringValue := map[string]string{
		"source":   `"review"`,
		"git":      `"https://example.org/kit.git"`,
		"ref_kind": `"tag"`,
		"ref":      `"v1"`,
		"commit":   `"` + commit + `"`,
	}
	for _, arm := range []string{"network-git", "configured-git"} {
		for field, value := range stringValue {
			for _, form := range []struct {
				name  string
				value string
			}{
				{"string", value},
				{"null", "null"},
				{"empty", `""`},
			} {
				t.Run("read/"+arm+"/"+field+"/"+form.name, func(t *testing.T) {
					base := v5GitAttestedMarker()
					if arm == "configured-git" {
						base = v5GitAttestedConfiguredMarker()
					}
					dir, recorded := writeV5(t, base)
					spliceV5Member(t, dir, `"`+field+`":`+form.value)
					if got := Read(dir); got != nil {
						t.Fatalf("Read admitted a v5 %s marker carrying %s=%s: %+v",
							arm, field, form.value, got)
					}
					expected := *recorded
					if current, _ := Current(dir, &expected); current {
						t.Fatalf("Current stayed true for a v5 %s marker carrying %s=%s",
							arm, field, form.value)
					}
				})
			}
		}
	}
}

// spliceV5Member rewrites a recorded marker with one extra top-level
// member spliced in, bypassing Write so the document carries exactly the
// bytes a foreign writer could produce.
func spliceV5Member(t *testing.T, dir, member string) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(dir, Name))
	if err != nil {
		t.Fatal(err)
	}
	trimmed := strings.TrimSuffix(string(payload), "\n")
	rewritten := strings.TrimSuffix(trimmed, "}") + "," + member + "}"
	if err := os.WriteFile(filepath.Join(dir, Name), []byte(rewritten), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestMarkerV5PlanMismatch proves every marker-plan mismatch case is
// non-current: changed registry, status, key, substitution identifier,
// package or lock against the effective plan refuses currency. Declared
// ref currency flows through the lock: a moved ref re-resolves to a new
// lock generation, and the lock_sha256 comparison observes it.
func TestMarkerV5PlanMismatch(t *testing.T) {
	base := v5GitAttestedMarker()
	dir, _ := writeV5(t, base)
	control := v5GitAttestedMarker()
	control.ContentSHA256 = base.ContentSHA256
	control.Files = append([]string(nil), base.Files...)
	if current, err := Current(dir, control); err != nil || !current {
		t.Fatalf("control Current = %v, %v, want true", current, err)
	}
	cases := map[string]func(*Marker){
		"registry": func(m *Marker) { m.Attestation.Registry = "example.org/other" },
		"status":   func(m *Marker) { m.Attestation.Status = "deprecated" },
		"key_id":   func(m *Marker) { m.Attestation.KeyID = "fedcba9876543210" },
		"attestation-absent": func(m *Marker) {
			m.Attestation = nil
		},
		"substituted": func(m *Marker) { m.Substituted = "dev:other" },
		"substituted-absent": func(m *Marker) {
			m.Substituted = ""
		},
		"package": func(m *Marker) {
			m.Package = &Package{Kind: "network-git", Repository: "example.org/kit",
				Commit: &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("66", 20)}, Directory: "skills/review"}
		},
		"lock_sha256": func(m *Marker) { m.LockSHA256 = "sha256:" + strings.Repeat("00", 32) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			fresh := v5GitAttestedMarker()
			fresh.ContentSHA256 = base.ContentSHA256
			fresh.Files = append([]string(nil), base.Files...)
			attestation := *fresh.Attestation
			fresh.Attestation = &attestation
			mutate(fresh)
			if current, err := Current(dir, fresh); err != nil || current {
				t.Fatalf("Current = %v, %v across %s mismatch, want false", current, err, name)
			}
		})
	}
}

// TestMarkerV5LocalRejectsEvidenceSummaryBytes proves the reader refuses a
// local-snapshot marker carrying registry or substitution summaries even
// when the bytes come from a foreign writer that bypassed Write.
func TestMarkerV5LocalRejectsEvidenceSummaryBytes(t *testing.T) {
	for name, member := range map[string]string{
		"attestation": `"attestation":{"registry":"example.org/kit","status":"audited"}`,
		"substituted": `"substituted":"dev:review"`,
		"source":      `"source":"review"`,
		"ref_kind":    `"ref_kind":"tag"`,
		"ref":         `"ref":"v1"`,
		"commit":      `"commit":"` + strings.Repeat("22", 20) + `"`,
		"git":         `"git":"https://example.org/kit.git"`,
	} {
		t.Run(name, func(t *testing.T) {
			dir, recorded := writeV5(t, v5LocalMarker())
			payload, err := os.ReadFile(filepath.Join(dir, Name))
			if err != nil {
				t.Fatal(err)
			}
			trimmed := strings.TrimSuffix(string(payload), "\n")
			rewritten := strings.TrimSuffix(trimmed, "}") + "," + member + "}"
			if err := os.WriteFile(filepath.Join(dir, Name), []byte(rewritten), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := Read(dir); got != nil {
				t.Fatalf("Read admitted a local marker with %s: %+v", name, got)
			}
			expected := *recorded
			if current, _ := Current(dir, &expected); current {
				t.Fatalf("Current stayed true for a local marker with %s", name)
			}
		})
	}
}

// TestMarkerV5LegacyInterchange proves old markers never attest schema-2
// currency and schema-5 markers never attest legacy currency: either
// direction is non-current even when the legacy side names an exact ref
// triple, so migration and rollback both reinstall.
func TestMarkerV5LegacyInterchange(t *testing.T) {
	legacyDir, _ := install(t)
	// A legacy installation is not current for a schema-2 expectation.
	v5 := v5GitAttestedMarker()
	if current, err := Current(legacyDir, v5); err != nil || current {
		t.Fatalf("legacy-recorded Current = %v, %v, want false", current, err)
	}
	// A schema-5 installation is not current for a legacy expectation,
	// even one naming an exact ref triple.
	v5Dir, _ := writeV5(t, v5GitAttestedMarker())
	legacyExpectation := &Marker{
		Name: "review", Source: "review", RefKind: "tag", Ref: "v1",
		Commit: strings.Repeat("44", 20),
	}
	if current, err := Current(v5Dir, legacyExpectation); err != nil || current {
		t.Fatalf("v5-recorded Current = %v, %v, want false", current, err)
	}
}
