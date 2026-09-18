package marker

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
)

func v5LocalBuild() Build {
	return Build{Driver: buildmeta.DriverGoV1, ReceiptSchemaVersion: 3, ExecutionPolicy: buildmeta.ExecutionPolicy,
		CacheKey: buildmeta.CacheKey("sha256:" + strings.Repeat("1", 64)), ReceiptSHA256: buildmeta.ReceiptHash("sha256:" + strings.Repeat("2", 64)),
		ArtifactSHA256: "sha256:" + strings.Repeat("3", 64), ArtifactPath: "bin/tool"}
}

func v5ExternalBuild() Build {
	return Build{Driver: "go-repository-v1", ReceiptSchemaVersion: 3, ExecutionPolicy: buildmeta.ExecutionPolicy,
		Repository: "tools", DeclaredIdentity: &RepositoryIdentity{Kind: "network-git", Value: "git.example.com/skills/tools"},
		DeclaredLockedCommit: &RepositoryCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)},
		EffectiveIdentity:    &RepositoryIdentity{Kind: "network-git", Value: "git.example.com/skills/tools"},
		ObjectFormat:         "sha1", Commit: strings.Repeat("a", 40),
		BuildSource:      &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("4", 64)},
		DescriptorTarget: "tool",
		CacheKey:         buildmeta.CacheKey("sha256:" + strings.Repeat("5", 64)), ReceiptSHA256: buildmeta.ReceiptHash("sha256:" + strings.Repeat("6", 64)),
		ArtifactSHA256: "sha256:" + strings.Repeat("7", 64), ArtifactPath: "bin/ext"}
}

func v5BuildMarker(local, external bool) *Marker {
	m := v5LocalMarker()
	m.SkillSchemaVersion = 7
	m.Commands = []string{"tool", "ext"}
	m.Activation = &Activation{Context: true, Commands: []string{"tool", "ext"}}
	m.Builds = map[string]Build{}
	if local {
		m.BuildRoots = []string{"build"}
		m.BuildSource = &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("8", 64)}
		m.Builds["tool"] = v5LocalBuild()
	}
	if external {
		m.Builds["ext"] = v5ExternalBuild()
	}
	return m
}

// TestMarkerV5BuildsBindReceiptVersion3OnBothArms: a schema-5 marker records
// every retained driver field with receipt_schema_version 3 for local go-v1
// and external go-repository-v1 builds, and refuses every other version.
func TestMarkerV5BuildsBindReceiptVersion3OnBothArms(t *testing.T) {
	for name, m := range map[string]*Marker{"local": v5BuildMarker(true, false), "external": v5BuildMarker(false, true), "both": v5BuildMarker(true, true)} {
		t.Run(name, func(t *testing.T) {
			_, recorded := writeV5(t, m)
			if recorded.SchemaVersion != SchemaV5 || len(recorded.Builds) != len(m.Builds) {
				t.Fatalf("recorded = %+v", recorded)
			}
			for command, build := range recorded.Builds {
				if build.ReceiptSchemaVersion != 3 || build.ExecutionPolicy != buildmeta.ExecutionPolicy {
					t.Fatalf("%s: build = %+v", command, build)
				}
			}
			if ext, ok := recorded.Builds["ext"]; ok && (ext.DeclaredLockedCommit == nil || ext.EffectiveIdentity == nil || ext.BuildSource == nil || ext.DescriptorTarget != "tool" || ext.Repository != "tools") {
				t.Fatalf("external record lost retained fields: %+v", ext)
			}
		})
	}
	refusals := map[string]func(*Marker){
		"local receipt 1": func(m *Marker) { b := m.Builds["tool"]; b.ReceiptSchemaVersion = 1; m.Builds["tool"] = b },
		"local receipt absent": func(m *Marker) {
			b := m.Builds["tool"]
			b.ReceiptSchemaVersion = 0
			b.ExecutionPolicy = ""
			m.Builds["tool"] = b
		},
		"local receipt 2":            func(m *Marker) { b := m.Builds["tool"]; b.ReceiptSchemaVersion = 2; m.Builds["tool"] = b },
		"local policy absent":        func(m *Marker) { b := m.Builds["tool"]; b.ExecutionPolicy = ""; m.Builds["tool"] = b },
		"local with repository":      func(m *Marker) { b := m.Builds["tool"]; b.Repository = "tools"; m.Builds["tool"] = b },
		"external receipt 2":         func(m *Marker) { b := m.Builds["ext"]; b.ReceiptSchemaVersion = 2; m.Builds["ext"] = b },
		"external receipt 1":         func(m *Marker) { b := m.Builds["ext"]; b.ReceiptSchemaVersion = 1; m.Builds["ext"] = b },
		"external hash-only record":  func(m *Marker) { b := m.Builds["ext"]; b.DeclaredLockedCommit = nil; m.Builds["ext"] = b },
		"external missing source":    func(m *Marker) { b := m.Builds["ext"]; b.BuildSource = nil; m.Builds["ext"] = b },
		"external missing effective": func(m *Marker) { b := m.Builds["ext"]; b.EffectiveIdentity = nil; m.Builds["ext"] = b },
		"external commit length":     func(m *Marker) { b := m.Builds["ext"]; b.Commit = strings.Repeat("a", 39); m.Builds["ext"] = b },
		"external substitution bool": func(m *Marker) { b := m.Builds["ext"]; b.Substituted = true; m.Builds["ext"] = b },
	}
	for name, mutate := range refusals {
		m := v5BuildMarker(true, true)
		mutate(m)
		dir := t.TempDir()
		if err := Write(dir, m); err != nil {
			continue // refused at write: also a refusal
		}
		if Read(dir) != nil {
			t.Fatalf("%s: v5 marker read back as valid", name)
		}
	}
	// Legacy schema-4 markers keep their receipt versions: 3 is refused there.
	legacy := v5BuildMarker(true, true)
	legacy.SchemaVersion = PolicySchemaVersion
	legacy.Package, legacy.LockSHA256 = nil, ""
	legacy.SkillSchemaVersion = 8
	legacy.Source, legacy.RefKind, legacy.Ref, legacy.Commit = "review", "tag", "v1", strings.Repeat("9", 40)
	rawLegacy := *legacy
	dir := t.TempDir()
	if err := Write(dir, &rawLegacy); err == nil && Read(dir) != nil {
		recorded := Read(dir)
		if recorded.SchemaVersion == PolicySchemaVersion {
			t.Fatal("legacy marker accepted receipt version 3")
		}
	}
}
