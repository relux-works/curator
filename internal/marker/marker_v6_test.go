package marker

import (
	"strings"
	"testing"
)

// v6NetworkMarker returns a schema-9 network-git draft marker selecting a
// repository subdirectory: the draft-sources-v2 carrier for a schema-9
// installation (skillfile-sources §4). The caller stamps content fields
// through writeV5 like production staging does.
func v6NetworkMarker() *Marker {
	m := v5GitAttestedMarker()
	m.SkillSchemaVersion = 9
	m.Package = &Package{
		Kind:       "network-git",
		Repository: "github.com/example/role-skills",
		Commit:     &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("44", 20)},
		Directory:  "skills/backend",
	}
	return m
}

// TestMarkerV6CarrierSelection pins the writer rule: v6 exactly for
// schema-9 installations with a package, v5 for every other packaged
// installation (skillfile-sources §4).
func TestMarkerV6CarrierSelection(t *testing.T) {
	for name, tc := range map[string]struct {
		skill int
		want  int
	}{
		"skill-9-network-git": {9, SchemaV6},
		"skill-8-network-git": {8, SchemaV5},
		"skill-4-network-git": {4, SchemaV5},
	} {
		t.Run(name, func(t *testing.T) {
			m := v6NetworkMarker()
			m.SkillSchemaVersion = tc.skill
			_, recorded := writeV5(t, m)
			if recorded.SchemaVersion != tc.want {
				t.Fatalf("recorded schema = %d, want %d", recorded.SchemaVersion, tc.want)
			}
		})
	}
	t.Run("skill-9-local-snapshot", func(t *testing.T) {
		m := v5LocalMarker()
		m.SkillSchemaVersion = 9
		_, recorded := writeV5(t, m)
		if recorded.SchemaVersion != SchemaV6 {
			t.Fatalf("recorded schema = %d, want %d", recorded.SchemaVersion, SchemaV6)
		}
	})
	t.Run("skill-9-configured-git", func(t *testing.T) {
		m := v5GitAttestedConfiguredMarker()
		m.SkillSchemaVersion = 9
		_, recorded := writeV5(t, m)
		if recorded.SchemaVersion != SchemaV6 {
			t.Fatalf("recorded schema = %d, want %d", recorded.SchemaVersion, SchemaV6)
		}
	})
}

// TestMarkerV6RecordsSelectedDirectory proves the v6 round trip keeps the
// normalized dependency directory on the package: the legacy-lane record
// duty of core §4.4.
func TestMarkerV6RecordsSelectedDirectory(t *testing.T) {
	_, recorded := writeV5(t, v6NetworkMarker())
	if recorded.SchemaVersion != SchemaV6 || recorded.SkillSchemaVersion != 9 {
		t.Fatalf("recorded marker = %+v, want marker schema 6 carrying skill schema 9", recorded)
	}
	if recorded.Package == nil || recorded.Package.Directory != "skills/backend" {
		t.Fatalf("recorded marker lost the selected directory: %+v", recorded)
	}
}

func TestMarkerV6RefusesMalformedIdentity(t *testing.T) {
	cases := map[string]func(*Marker){
		"missing-package": func(m *Marker) { m.Package = nil },
		"missing-lock":    func(m *Marker) { m.LockSHA256 = "" },
		"malformed-lock":  func(m *Marker) { m.LockSHA256 = "sha256:xyz" },
		"legacy-source":   func(m *Marker) { m.Source = "review" },
		"legacy-git":      func(m *Marker) { m.Git = "https://example.org/kit.git" },
		"legacy-ref":      func(m *Marker) { m.RefKind = "tag"; m.Ref = "v1"; m.Commit = strings.Repeat("22", 20) },
		"unknown-kind":    func(m *Marker) { m.Package.Kind = "bad-kind" },
		"skill-schema-0":  func(m *Marker) { m.SkillSchemaVersion = 0 },
		"skill-schema-10": func(m *Marker) { m.SkillSchemaVersion = 10 },
		"glob-directory":  func(m *Marker) { m.Package.Directory = "skills/*" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := v6NetworkMarker()
			mutate(m)
			if err := Write(t.TempDir(), m); err == nil {
				t.Fatalf("Write admitted a v6 marker with %s", name)
			}
		})
	}
}

// TestMarkerV6Currentness pins the v6 currentness comparison: a v6
// installation is current only for the exact frozen package and lock
// generation it was installed from, and neither a core marker nor a
// source-extension v5 marker can establish currentness for a schema-9
// installation.
func TestMarkerV6Currentness(t *testing.T) {
	base := v6NetworkMarker()
	dir, _ := writeV5(t, base)
	control := v6NetworkMarker()
	control.ContentSHA256 = base.ContentSHA256
	control.Files = append([]string(nil), base.Files...)
	if current, err := Current(dir, control); err != nil || !current {
		t.Fatalf("control Current = %v, %v, want true", current, err)
	}
	t.Run("directory-mismatch", func(t *testing.T) {
		fresh := v6NetworkMarker()
		fresh.ContentSHA256 = base.ContentSHA256
		fresh.Files = append([]string(nil), base.Files...)
		fresh.Package = &Package{
			Kind:       "network-git",
			Repository: "github.com/example/role-skills",
			Commit:     &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("44", 20)},
			Directory:  "skills/frontend",
		}
		if current, err := Current(dir, fresh); err != nil || current {
			t.Fatalf("Current = %v, %v across directory mismatch, want false", current, err)
		}
	})
	t.Run("v5-recorded-never-current-for-schema-9", func(t *testing.T) {
		legacy := v5GitAttestedMarker()
		legacy.Package = &Package{
			Kind:       "network-git",
			Repository: "github.com/example/role-skills",
			Commit:     &Commit{ObjectFormat: "sha1", Hex: strings.Repeat("44", 20)},
			Directory:  "skills/backend",
		}
		// Both fixtures stamp the same content bytes, so the recorded
		// v5 content identity already matches the expectation and the
		// schema mismatch is the only possible verdict.
		v5Dir, recorded := writeV5(t, legacy)
		fresh := v6NetworkMarker()
		fresh.ContentSHA256 = recorded.ContentSHA256
		fresh.Files = append([]string(nil), recorded.Files...)
		if current, err := Current(v5Dir, fresh); err != nil || current {
			t.Fatalf("Current = %v, %v for a v5 record against a schema-9 expectation, want false", current, err)
		}
	})
}
