package scopes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
)

func TestConsumersRoundTrip(t *testing.T) {
	home := t.TempDir()
	if err := RecordConsumer(home, "/a/b"); err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, "/a/b"); err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, "/c"); err != nil {
		t.Fatal(err)
	}
	consumers := LoadConsumers(home)
	if len(consumers) != 2 {
		t.Fatalf("consumers: %v", consumers)
	}
}

func installMarker(t *testing.T, skillsDir, name, commit string) {
	t.Helper()
	dir := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	hash, _ := hashing.ContentSHA256(dir, nil)
	if err := marker.Write(dir, &marker.Marker{
		Name: name, Source: name, RefKind: "tag", Ref: "v1",
		Commit: commit, ContentSHA256: hash,
		Agents: []string{}, Commands: []string{}, Dependencies: []string{},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGcKeepsReferencedRuntime(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	installMarker(t, filepath.Join(project, ".agents", "skills"), "skill-a", "c1")
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	// hybrid store reference
	installMarker(t, HybridSkillsRoot(home), "skill-h", "c9")

	for _, entry := range []string{"skill-a/c1", "skill-a/c2", "skill-h/c9", "skill-x/c3"} {
		if err := os.MkdirAll(filepath.Join(home, "runtime", filepath.FromSlash(entry)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := CollectRuntime(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 2 {
		t.Fatalf("removed: %v", removed)
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-a", "c1")); err != nil {
		t.Fatal("referenced runtime removed")
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-h", "c9")); err != nil {
		t.Fatal("hybrid-referenced runtime removed")
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-a", "c2")); err == nil {
		t.Fatal("unreferenced commit survived")
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-x")); err == nil {
		t.Fatal("empty skill dir survived")
	}
}

func TestGcPrunesDeadConsumers(t *testing.T) {
	home := t.TempDir()
	dead := filepath.Join(t.TempDir(), "gone")
	if err := RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}
	live := t.TempDir()
	installMarker(t, filepath.Join(live, ".agents", "skills"), "skill-a", "c1")
	if err := RecordConsumer(home, live); err != nil {
		t.Fatal(err)
	}
	if _, err := CollectRuntime(home); err != nil {
		t.Fatal(err)
	}
	consumers := LoadConsumers(home)
	if len(consumers) != 1 {
		t.Fatalf("dead consumer not pruned: %v", consumers)
	}
}

func writeHybrid(t *testing.T, home, text string) {
	t.Helper()
	path := HybridManifestPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHybridDeclsRequireTargets(t *testing.T) {
	home := t.TempDir()
	writeHybrid(t, home, `{"schema_version": 1, "skills": [{"name": "skill-h", "tag": "v1"}]}`)
	if _, err := LoadHybridDecls(home); err == nil {
		t.Fatal("hybrid decl without targets must fail")
	}
	writeHybrid(t, home, `{"schema_version": 1, "skills": [{"name": "skill-h", "tag": "v1", "targets": []}]}`)
	if _, err := LoadHybridDecls(home); err == nil {
		t.Fatal("empty targets must fail")
	}
	writeHybrid(t, home, `{"schema_version": 1, "skills": [{"name": "skill-h", "tag": "v1", "targets": ["my-alias"]}]}`)
	decls, err := LoadHybridDecls(home)
	if err != nil || len(decls) != 1 || decls[0].Targets[0] != "my-alias" {
		t.Fatalf("decls: %+v, %v", decls, err)
	}
}

func TestAppliesToProject(t *testing.T) {
	project := t.TempDir()
	abs, _ := filepath.Abs(project)
	posix := filepath.ToSlash(abs)

	byAlias := HybridDecl{Targets: []string{"my-alias"}}
	if !AppliesToProject(byAlias, []string{"my-alias", "other"}, project) {
		t.Fatal("alias target must match")
	}
	if AppliesToProject(byAlias, []string{"different"}, project) {
		t.Fatal("non-matching alias must not match")
	}

	byPath := HybridDecl{Targets: []string{posix}}
	if !AppliesToProject(byPath, nil, project) {
		t.Fatal("exact path must match")
	}

	byGlob := HybridDecl{Targets: []string{filepath.ToSlash(filepath.Dir(posix)) + "/*"}}
	if !AppliesToProject(byGlob, nil, project) {
		t.Fatal("glob must match")
	}
}

func TestHybridAddRemove(t *testing.T) {
	home := t.TempDir()
	if err := AddHybridDecl(home, "skill-h", "tag", "v1", "git@example.com:x/skill-h.git", []string{"alias-a"}); err != nil {
		t.Fatal(err)
	}
	// replace with new targets
	if err := AddHybridDecl(home, "skill-h", "tag", "v2", "", []string{"alias-b"}); err != nil {
		t.Fatal(err)
	}
	decls, err := LoadHybridDecls(home)
	if err != nil || len(decls) != 1 {
		t.Fatalf("decls: %+v, %v", decls, err)
	}
	if decls[0].Decl.Ref.Value != "v2" || decls[0].Targets[0] != "alias-b" {
		t.Fatalf("replace failed: %+v", decls[0])
	}
	if err := RemoveHybridDecl(home, "skill-h"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveHybridDecl(home, "skill-h"); err == nil {
		t.Fatal("removing absent decl must fail")
	}
	// add validates through the standard manifest rules
	if err := AddHybridDecl(home, "-bad", "tag", "v1", "", []string{"a"}); err == nil {
		t.Fatal("invalid name must fail")
	}
	_ = manifest.SchemaVersion
}
