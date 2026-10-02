package marker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/stateread"
)

// Drive the reader with foreign-writer bytes: Write's validation must not
// hide a missing reader check. Each refusal also has an accepted control.
func TestReadExternalBuildCrossFields(t *testing.T) {
	local := func(b *Build) {
		b.Substituted = true
		b.Substitution = &RepositorySubstitute{Type: "local-path"}
		b.EffectiveIdentity = &RepositoryIdentity{Kind: "operator-local-git", Value: "sha256:" + strings.Repeat("c", 64)}
		b.Commit = strings.Repeat("b", 40)
	}
	network := func(b *Build) {
		b.Substituted = true
		b.Substitution = &RepositorySubstitute{Type: "network-git", Ref: &RepositoryRef{Kind: "revision", Value: strings.Repeat("b", 40)}}
		b.EffectiveIdentity.Value = "git.example.com/forks/tools"
		b.Commit = strings.Repeat("b", 40)
	}
	sha256 := func(b *Build) {
		network(b)
		b.ObjectFormat = "sha256"
		b.DeclaredLockedCommit = &RepositoryCommit{ObjectFormat: "sha256", Hex: strings.Repeat("a", 64)}
		b.Commit = strings.Repeat("b", 64)
		b.Substitution.Ref.Value = strings.Repeat("b", 64)
	}
	rows := []struct {
		name   string
		mutate func(*Build)
		valid  bool
	}{
		{"unsubstituted-control", func(*Build) {}, true},
		{"local-control", local, true},
		{"network-sha1-control", network, true},
		{"network-sha256-control", sha256, true},
		{"network-tag-control", func(b *Build) { network(b); b.Substitution.Ref = &RepositoryRef{Kind: "tag", Value: "v1.2.0"} }, true},
		{"network-branch-control", func(b *Build) { network(b); b.Substitution.Ref = &RepositoryRef{Kind: "branch", Value: "release/v2"} }, true},
		{"declared-effective-value-mismatch", func(b *Build) { b.EffectiveIdentity.Value = "git.example.com/forks/tools" }, false},
		{"declared-effective-kind-mismatch", func(b *Build) { b.EffectiveIdentity.Kind = "operator-local-git" }, false},
		{"declared-effective-commit-mismatch", func(b *Build) { b.Commit = strings.Repeat("b", 40) }, false},
		{"local-identity-kind-mismatch", func(b *Build) { local(b); b.EffectiveIdentity.Kind = "network-git" }, false},
		{"network-identity-kind-mismatch", func(b *Build) { network(b); b.EffectiveIdentity.Kind = "operator-local-git" }, false},
		{"sha1-effective-revision-width", func(b *Build) { network(b); b.Substitution.Ref.Value = strings.Repeat("b", 64) }, false},
		{"sha256-effective-revision-width", func(b *Build) { sha256(b); b.Substitution.Ref.Value = strings.Repeat("b", 40) }, false},
		{"sha1-effective-revision-nonhex", func(b *Build) { network(b); b.Substitution.Ref.Value = strings.Repeat("z", 40) }, false},
		{"sha256-effective-revision-nonhex", func(b *Build) { sha256(b); b.Substitution.Ref.Value = strings.Repeat("z", 64) }, false},
		{"local-with-ref", func(b *Build) { local(b); b.Substitution.Ref = &RepositoryRef{Kind: "branch", Value: "main"} }, false},
		{"network-without-ref", func(b *Build) { network(b); b.Substitution.Ref = nil }, false},
		{"unknown-substitution-type", func(b *Build) { network(b); b.Substitution.Type = "unknown" }, false},
		{"unknown-ref-kind", func(b *Build) { network(b); b.Substitution.Ref.Kind = "unknown" }, false},
	}
	for _, version := range []int{ExternalSchemaVersion, PolicySchemaVersion, SchemaV5} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.name, func(t *testing.T) {
					m := v3Base()
					m.SchemaVersion = version
					if version != ExternalSchemaVersion {
						m.SkillSchemaVersion = 8
					}
					if version == SchemaV5 {
						m.HashVersion = 2
					}
					m.Commands, m.BuildRoots, m.BuildSource = []string{"external"}, []string{}, nil
					b := externalV3Build()
					row.mutate(&b)
					m.Builds["external"] = b
					payload, err := json.Marshal(m)
					if err != nil {
						t.Fatal(err)
					}
					dir := t.TempDir()
					if err := os.WriteFile(filepath.Join(dir, Name), payload, 0o600); err != nil {
						t.Fatal(err)
					}
					if got := Read(dir) != nil; got != row.valid {
						t.Errorf("Read accepted=%v, want %v", got, row.valid)
					}
					got, kind, err := ReadState(dir)
					if kind != stateread.KindPresent || (got != nil) != row.valid || (err == nil) != row.valid {
						t.Fatalf("ReadState marker=%v kind=%v err=%v, want present and valid=%v", got, kind, err, row.valid)
					}
				})
			}
		})
	}
}

func TestReadAuthoritativeValidMarkers(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	counts, _, err := conformancecoverage.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range []string{"install-marker-v1", "install-marker-v2", "install-marker-v3", "install-marker-v4", "install-marker-v5"} {
		wantCount, published := counts["marker/"+family+"/schema-cases"]
		if family == "install-marker-v1" {
			// The frozen v1 family predates the coverage ledger; all three
			// pinned suites publish its same two cases.
			wantCount, published = 2, true
		}
		if !published {
			continue
		}
		t.Run(family, func(t *testing.T) {
			cases := loadMarkerSchemaCases(t, family)
			if len(cases) != wantCount {
				t.Fatalf("published cases=%d, want %d", len(cases), wantCount)
			}
			for _, tc := range cases {
				if !tc.Valid {
					continue
				}
				t.Run(tc.Name, func(t *testing.T) {
					dir := t.TempDir()
					if err := os.WriteFile(filepath.Join(dir, Name), tc.Bytes, 0o600); err != nil {
						t.Fatal(err)
					}
					if Read(dir) == nil {
						t.Fatal("Read refused the published valid marker")
					}
				})
			}
		})
	}
}

// Assert the five published regressions independently of gap classification,
// so an outstanding ledger row cannot turn an admitted invalid case green.
func TestReadAuthoritativeMarkerExternalCrossFields(t *testing.T) {
	names := map[string]bool{
		"invalid-external-declared-effective-mismatch.json":   true,
		"invalid-marker-local-identity-kind-mismatch.json":    true,
		"invalid-marker-network-identity-kind-mismatch.json":  true,
		"invalid-marker-sha1-effective-revision-width.json":   true,
		"invalid-marker-sha256-effective-revision-width.json": true,
	}
	for _, family := range []string{"install-marker-v3", "install-marker-v4"} {
		t.Run(family, func(t *testing.T) {
			found := 0
			for _, tc := range loadMarkerSchemaCases(t, family) {
				if !names[tc.Name] {
					continue
				}
				found++
				t.Run(tc.Name, func(t *testing.T) {
					if tc.Valid {
						t.Fatal("published regression must be invalid")
					}
					dir := t.TempDir()
					if err := os.WriteFile(filepath.Join(dir, Name), tc.Bytes, 0o600); err != nil {
						t.Fatal(err)
					}
					if Read(dir) != nil {
						t.Fatal("Read admitted the published invalid marker")
					}
				})
			}
			if found != len(names) {
				t.Fatalf("published cross-field cases=%d, want %d", found, len(names))
			}
		})
	}
}
