package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
)

// This boundary witness exercises packageOf -> real Identity, not wire security.
func TestRequirementDeclarationSurvivesIdentity(t *testing.T) {
	const canonical = "example.com/team/dep"
	for _, kind := range []string{contextlock.KindContext, contextlock.KindSkill, contextlock.KindMCP} {
		for _, declaration := range []string{"https://example.com/team/dep.git", "ssh://git@example.com/team/dep.git", "git@example.com:team/dep.git", "example.com:team/dep.git", canonical} {
			t.Run(kind+"/"+declaration, func(t *testing.T) {
				weight := int64(7)
				req := contextpkg.Requirement{Git: declaration, Range: "^1.0.0", Tag: "v1.2.0", Revision: "0123456789012345678901234567890123456789", Directory: "nested", Weight: &weight, Mode: "manual", Commands: []string{"check"}}
				manifest := &contextpkg.Manifest{}
				switch kind {
				case contextlock.KindContext:
					manifest.Contexts = map[string]contextpkg.Requirement{"dep": req}
				case contextlock.KindSkill:
					manifest.Skills = map[string]contextpkg.Requirement{"dep": req}
				case contextlock.KindMCP:
					manifest.MCP = map[string]contextpkg.Requirement{"dep": req}
				}
				pkg := packageOf(manifest)
				if len(pkg.Requires) != 1 {
					t.Fatalf("requirements: %+v", pkg.Requires)
				}
				got := pkg.Requires[0]
				if got.Source != declaration || got.Kind != kind || got.Name != "dep" || got.Range != req.Range || got.Tag != req.Tag || got.Revision != req.Revision || got.Directory != req.Directory {
					t.Fatalf("declaration/metadata changed: %+v", got)
				}
				if kind == contextlock.KindContext && (got.Weight == nil || *got.Weight != weight) {
					t.Fatalf("weight: %+v", got)
				}
				if kind == contextlock.KindSkill && (got.Mode != req.Mode || !reflect.DeepEqual(got.Commands, req.Commands)) {
					t.Fatalf("skill metadata: %+v", got)
				}
				manager := newGitManager(t.TempDir())
				key, err := manager.Identity(got.Kind, got.Name, got.Source)
				if err != nil || key != canonical {
					t.Fatalf("identity = %q, %v", key, err)
				}
				raw, exists := manager.fetchRaw[key]
				if declaration == canonical {
					if exists {
						t.Fatalf("canonical input invented raw mapping: %q", raw)
					}
				} else if !exists || raw != declaration {
					t.Fatalf("endpoint = %q, present=%t; want %q", raw, exists, declaration)
				}
			})
		}
	}
}

// Public cold Install uses independently authored bytes/pins. Offline insteadOf
// routing deliberately does not certify SSH authentication or rewrite isolation.
func TestInstallDependencyUsesDeclaredEndpoint(t *testing.T) {
	for _, tc := range []struct{ name, kind, collection, declaration string }{
		{"context-ssh", contextlock.KindContext, "contexts", "ssh://git@example.com/team/dep"},
		{"context-https", contextlock.KindContext, "contexts", "https://example.com/team/dep.git"},
		{"skill-ssh", contextlock.KindSkill, "skills", "ssh://git@example.com/team/dep"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pinHomes(t)
			pinOperatorHome(t, t.TempDir())
			ids := newGitIdentities(t)
			const expected = "# Independent correct dependency bytes\n"
			files := func(payload string) map[string]string {
				if tc.kind == contextlock.KindSkill {
					return map[string]string{"SKILL.md": payload}
				}
				return map[string]string{"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`, "context/a.md": payload}
			}
			correct := gitRepo(t, files(expected), "v1.0.0")
			decoy := gitRepo(t, files("# Independent WRONG decoy\n"), "v1.0.0")
			pin := fixtureGitOutput(t, correct, "rev-parse", "HEAD")
			if pin == fixtureGitOutput(t, decoy, "rev-parse", "HEAD") {
				t.Fatal("fixture pins must differ")
			}
			ids.serve(correct, tc.declaration)
			ids.serve(decoy, "https://example.com/team/dep")
			root, home := gitFixtureHome(t), gitFixtureHome(t)
			writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"root","version":"1.0.0","requires":{%q:{"dep":{"git":%q,"tag":"v1.0.0"}}}}`, tc.collection, tc.declaration), nil)
			info, _, _, err := Install(home, InstallOptions{Operand: root})
			if err != nil {
				t.Fatal(err)
			}
			member, ok := info.Lock.Find(tc.kind, "dep")
			if !ok || member.Commit != pin {
				t.Fatalf("independent declared-endpoint pin: got %+v; want %s", member, pin)
			}
			if member.Source != "example.com/team/dep" {
				t.Fatalf("canonical lock source: %q", member.Source)
			}
			path := "context/a.md"
			if tc.kind == contextlock.KindSkill {
				path = "SKILL.md"
			}
			payload, err := os.ReadFile(filepath.Join(contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey()), filepath.FromSlash(path)))
			if err != nil || string(payload) != expected {
				t.Fatalf("independent declared-endpoint bytes: %q, %v", payload, err)
			}
			entries, err := os.ReadDir(profileReposDir(home))
			if err != nil || len(entries) != 1 {
				t.Fatalf("canonical cache entries: %v, %v", entries, err)
			}
			if _, err := os.Stat(newGitManager(home).repoDir("example.com/team/dep")); err != nil {
				t.Fatal(err)
			}
		})
	}
}
