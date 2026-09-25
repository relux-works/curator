package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// rewriteV5Build writes a recorded v5 marker back with one builds entry
// replaced by the given JSON text, bypassing Write so the document carries
// exactly the bytes a foreign writer could produce.
func rewriteV5Build(t *testing.T, dir, command, buildJSON string) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(dir, Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	var builds map[string]json.RawMessage
	if err := json.Unmarshal(raw["builds"], &builds); err != nil {
		t.Fatal(err)
	}
	builds[command] = json.RawMessage(buildJSON)
	rewrittenBuilds, err := json.Marshal(builds)
	if err != nil {
		t.Fatal(err)
	}
	raw["builds"] = rewrittenBuilds
	rewritten, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, Name), rewritten, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustExtJSON(t *testing.T, build Build) string {
	t.Helper()
	payload, err := json.Marshal(build)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func mutateExtJSON(t *testing.T, base Build, mutate func(map[string]json.RawMessage)) string {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(mustExtJSON(t, base)), &object); err != nil {
		t.Fatal(err)
	}
	mutate(object)
	payload, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func mutateNestedJSON(t *testing.T, outer map[string]json.RawMessage, field string, mutate func(map[string]json.RawMessage)) {
	t.Helper()
	var nested map[string]json.RawMessage
	if err := json.Unmarshal(outer[field], &nested); err != nil {
		t.Fatal(err)
	}
	mutate(nested)
	payload, err := json.Marshal(nested)
	if err != nil {
		t.Fatal(err)
	}
	outer[field] = payload
}

// TestMarkerV5ExternalBuildClosedShape is the reader regression for
// BUG-260920-2eg8nv: an external go-repository-v1 record must match the
// closed raw shape of its arm. An absent `substituted` decodes to the same
// false as an explicit one, and a null member decodes to the same zero
// value as an absent one, so the raw object is checked before the lossy
// decode. Every row drives marker.Read, the production entry.
func TestMarkerV5ExternalBuildClosedShape(t *testing.T) {
	substitutedLocal := v5ExternalBuild()
	substitutedLocal.Substituted = true
	substitutedLocal.Substitution = &RepositorySubstitute{Type: "local-path"}
	substitutedNetwork := v5ExternalBuild()
	substitutedNetwork.Substituted = true
	substitutedNetwork.Substitution = &RepositorySubstitute{
		Type: "network-git",
		Ref:  &RepositoryRef{Kind: "tag", Value: "v1.4.0"},
	}
	withTag := v5ExternalBuild()
	withTag.DeclaredTag = "v1.4.0"

	type row struct {
		build  func(t *testing.T) string
		readOK bool
	}
	rows := map[string]row{
		"control-unsubstituted":       {func(t *testing.T) string { return mustExtJSON(t, v5ExternalBuild()) }, true},
		"control-substituted-local":   {func(t *testing.T) string { return mustExtJSON(t, substitutedLocal) }, true},
		"control-substituted-network": {func(t *testing.T) string { return mustExtJSON(t, substitutedNetwork) }, true},
		"control-declared-tag":        {func(t *testing.T) string { return mustExtJSON(t, withTag) }, true},

		"missing-substituted": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { delete(o, "substituted") })
		}, false},
		"substituted-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["substituted"] = json.RawMessage(`null`) })
		}, false},
		"substituted-string": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["substituted"] = json.RawMessage(`"false"`) })
		}, false},
		"substituted-number": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["substituted"] = json.RawMessage(`0`) })
		}, false},
		"substitution-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["substitution"] = json.RawMessage(`null`) })
		}, false},
		"substituted-true-missing-substitution": {func(t *testing.T) string {
			missing := v5ExternalBuild()
			missing.Substituted = true
			return mustExtJSON(t, missing)
		}, false},
		"substituted-false-with-substitution": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				o["substitution"] = json.RawMessage(`{"type":"local-path"}`)
			})
		}, false},
		"foreign-field": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["extra"] = json.RawMessage(`1`) })
		}, false},
		"foreign-field-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["extra"] = json.RawMessage(`null`) })
		}, false},
		"declared-tag-null": {func(t *testing.T) string {
			return mutateExtJSON(t, withTag, func(o map[string]json.RawMessage) { o["declared_tag"] = json.RawMessage(`null`) })
		}, false},
		"declared-tag-number": {func(t *testing.T) string {
			return mutateExtJSON(t, withTag, func(o map[string]json.RawMessage) { o["declared_tag"] = json.RawMessage(`123`) })
		}, false},
		"commit-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["commit"] = json.RawMessage(`null`) })
		}, false},
		"repository-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["repository"] = json.RawMessage(`null`) })
		}, false},
		"receipt-version-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["receipt_schema_version"] = json.RawMessage(`null`) })
		}, false},
		"receipt-version-string": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["receipt_schema_version"] = json.RawMessage(`"3"`) })
		}, false},
		"missing-repository": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { delete(o, "repository") })
		}, false},
		"missing-commit": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { delete(o, "commit") })
		}, false},
		"missing-build-source": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { delete(o, "build_source") })
		}, false},
		"declared-identity-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) { o["declared_identity"] = json.RawMessage(`null`) })
		}, false},
		"declared-identity-extra": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "declared_identity", func(n map[string]json.RawMessage) { n["extra"] = json.RawMessage(`1`) })
			})
		}, false},
		"declared-identity-kind-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "declared_identity", func(n map[string]json.RawMessage) { n["kind"] = json.RawMessage(`null`) })
			})
		}, false},
		"locked-commit-extra": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "declared_locked_commit", func(n map[string]json.RawMessage) { n["ref"] = json.RawMessage(`"main"`) })
			})
		}, false},
		"locked-commit-hex-null": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "declared_locked_commit", func(n map[string]json.RawMessage) { n["hex"] = json.RawMessage(`null`) })
			})
		}, false},
		"effective-identity-extra": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "effective_identity", func(n map[string]json.RawMessage) { n["extra"] = json.RawMessage(`1`) })
			})
		}, false},
		"build-source-extra": {func(t *testing.T) string {
			return mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "build_source", func(n map[string]json.RawMessage) { n["extra"] = json.RawMessage(`1`) })
			})
		}, false},
		"local-path-with-ref": {func(t *testing.T) string {
			return mutateExtJSON(t, substitutedLocal, func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "substitution", func(n map[string]json.RawMessage) {
					n["ref"] = json.RawMessage(`{"kind":"tag","value":"v1.4.0"}`)
				})
			})
		}, false},
		"network-git-missing-ref": {func(t *testing.T) string {
			return mutateExtJSON(t, substitutedNetwork, func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "substitution", func(n map[string]json.RawMessage) { delete(n, "ref") })
			})
		}, false},
		"network-git-ref-null": {func(t *testing.T) string {
			return mutateExtJSON(t, substitutedNetwork, func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "substitution", func(n map[string]json.RawMessage) { n["ref"] = json.RawMessage(`null`) })
			})
		}, false},
		"substitution-unknown-type": {func(t *testing.T) string {
			return mutateExtJSON(t, substitutedLocal, func(o map[string]json.RawMessage) {
				mutateNestedJSON(t, o, "substitution", func(n map[string]json.RawMessage) { n["type"] = json.RawMessage(`"bad"`) })
			})
		}, false},
		"substitution-missing-type": {func(t *testing.T) string {
			return mutateExtJSON(t, substitutedLocal, func(o map[string]json.RawMessage) { o["substitution"] = json.RawMessage(`{}`) })
		}, false},
	}
	for name, r := range rows {
		t.Run(name, func(t *testing.T) {
			dir, _ := writeV5(t, v5BuildMarker(true, true))
			rewriteV5Build(t, dir, "ext", r.build(t))
			got := Read(dir)
			if (got != nil) != r.readOK {
				t.Fatalf("Read = %v, want readable=%v", got, r.readOK)
			}
			if got == nil {
				return
			}
			if got.Builds["ext"].Driver != "go-repository-v1" {
				t.Fatalf("recorded ext driver = %q", got.Builds["ext"].Driver)
			}
		})
	}
}

// TestMarkerV5LocalBuildClosedShape proves the local go-v1 arm is validated
// from raw JSON at marker.Read: external-only values cannot disappear into
// the decoded struct's zero values. Each row rewrites bytes after a valid
// marker has been written, bypassing Write like a foreign marker producer.
func TestMarkerV5LocalBuildClosedShape(t *testing.T) {
	for field := range map[string]struct{}{
		"repository":   {},
		"substituted":  {},
		"substitution": {},
	} {
		t.Run(field+"-null", func(t *testing.T) {
			dir, _ := writeV5(t, v5BuildMarker(true, false))
			build := mutateExtJSON(t, v5LocalBuild(), func(o map[string]json.RawMessage) {
				o[field] = json.RawMessage(`null`)
			})
			rewriteV5Build(t, dir, "tool", build)
			if got := Read(dir); got != nil {
				t.Fatalf("Read admitted local go-v1 build with %s:null: %+v", field, got.Builds["tool"])
			}
		})
	}
}

// TestMarkerV5DeclaredTagRequiresValidGitRef drives empty and malformed tag
// names through marker.Read. The same validator is used for draft Skillfile
// source refs, so marker summaries cannot retain a value the locked source
// contract would reject.
func TestMarkerV5DeclaredTagRequiresValidGitRef(t *testing.T) {
	for name, tag := range map[string]string{
		"empty":     `""`,
		"malformed": `"release..next"`,
	} {
		t.Run(name, func(t *testing.T) {
			dir, _ := writeV5(t, v5BuildMarker(false, true))
			build := mutateExtJSON(t, v5ExternalBuild(), func(o map[string]json.RawMessage) {
				o["declared_tag"] = json.RawMessage(tag)
			})
			rewriteV5Build(t, dir, "ext", build)
			if got := Read(dir); got != nil {
				t.Fatalf("Read admitted declared_tag %s: %+v", tag, got.Builds["ext"])
			}
		})
	}

	withValidTag := v5ExternalBuild()
	withValidTag.DeclaredTag = "release/v1.4.0"
	dir, _ := writeV5(t, v5BuildMarker(false, true))
	rewriteV5Build(t, dir, "ext", mustExtJSON(t, withValidTag))
	if got := Read(dir); got == nil || got.Builds["ext"].DeclaredTag != withValidTag.DeclaredTag {
		t.Fatalf("Read did not preserve a valid declared_tag: got=%+v", got)
	}
}
