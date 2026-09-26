package scopes

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/stateread"
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
	consumers := mustLoadConsumers(t, home)
	if len(consumers) != 2 {
		t.Fatalf("consumers: %v", consumers)
	}
}

func mustLoadConsumers(t *testing.T, home string) []string {
	t.Helper()
	consumers, err := LoadConsumers(home)
	if err != nil {
		t.Fatal(err)
	}
	return consumers
}

func makeScopesStateUnreadable(t *testing.T, path string, directory bool) {
	t.Helper()
	requirePOSIXModeBitUnreadability(t)
	mode := os.FileMode(0o644)
	if directory {
		mode = 0o755
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Skipf("this host cannot create mode-000 manager state: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, mode) })
	if directory {
		if _, err := os.ReadDir(path); err == nil {
			t.Skip("this environment can read a mode-000 directory; unreadability is untestable here")
		}
	} else if _, err := os.ReadFile(path); err == nil {
		t.Skip("this environment can read a mode-000 file; unreadability is untestable here")
	}
}

func requirePOSIXModeBitUnreadability(t testing.TB) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
}

func TestConsumerRegistryAbsenceAndUnreadabilityDiffer(t *testing.T) {
	home := t.TempDir()
	if consumers, err := LoadConsumers(home); err != nil || len(consumers) != 0 {
		t.Fatalf("absent consumer registry = (%v, %v), want empty and nil", consumers, err)
	}
	project := t.TempDir()
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ConsumersName)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	makeScopesStateUnreadable(t, path, false)
	if consumers, err := LoadConsumers(home); err == nil || len(consumers) != 0 || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("unreadable consumer registry = (%v, %v), want typed unreadable state", consumers, err)
	}
	if err := RecordConsumer(home, t.TempDir()); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("record over unreadable consumer registry = %v, want typed unreadable refusal", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(original) {
		t.Fatalf("unreadable registry changed = (%q, %v), want original bytes", got, err)
	}
}

func TestSweepRuntimeAbsenceAndUnreadabilityDiffer(t *testing.T) {
	home := t.TempDir()
	removed, err := CollectRuntime(home)
	if err != nil || len(removed) != 0 {
		t.Fatalf("absent runtime root = (%v, %v), want empty and nil", removed, err)
	}
	runtimeRoot := filepath.Join(home, "runtime")
	if err := os.MkdirAll(runtimeRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	makeScopesStateUnreadable(t, runtimeRoot, true)
	if _, err := CollectRuntime(home); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), runtimeRoot) {
		t.Fatalf("unreadable runtime root error = %v, want typed unreadable state for %s", err, runtimeRoot)
	}
}

func TestSweepRuntimeUnreadableSkillInventoryIsNotIgnored(t *testing.T) {
	home := t.TempDir()
	runtimeRoot := filepath.Join(home, "runtime")
	skillRoot := filepath.Join(runtimeRoot, "skill-a")
	if err := os.MkdirAll(filepath.Join(skillRoot, strings.Repeat("1", 40)), 0o755); err != nil {
		t.Fatal(err)
	}
	makeScopesStateUnreadable(t, skillRoot, true)
	if _, err := CollectRuntime(home); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), skillRoot) {
		t.Fatalf("unreadable runtime skill inventory error = %v, want typed unreadable state for %s", err, skillRoot)
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
		InstalledAt: "2026-07-13T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGcKeepsReferencedRuntime(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	commit1 := strings.Repeat("1", 40)
	commit2 := strings.Repeat("2", 40)
	commit3 := strings.Repeat("3", 40)
	commit9 := strings.Repeat("9", 40)
	installMarker(t, filepath.Join(project, ".agents", "skills"), "skill-a", commit1)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	// hybrid store reference
	installMarker(t, HybridSkillsRoot(home), "skill-h", commit9)

	for _, entry := range []string{"skill-a/" + commit1, "skill-a/" + commit2, "skill-h/" + commit9, "skill-x/" + commit3} {
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
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-a", commit1)); err != nil {
		t.Fatal("referenced runtime removed")
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-h", commit9)); err != nil {
		t.Fatal("hybrid-referenced runtime removed")
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-a", commit2)); err == nil {
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
	installMarker(t, filepath.Join(live, ".agents", "skills"), "skill-a", strings.Repeat("1", 40))
	if err := RecordConsumer(home, live); err != nil {
		t.Fatal(err)
	}
	if _, err := CollectRuntime(home); err != nil {
		t.Fatal(err)
	}
	consumers := mustLoadConsumers(t, home)
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

func TestHybridManifestAbsenceAndUnreadabilityDiffer(t *testing.T) {
	home := t.TempDir()
	if decls, err := LoadHybridDecls(home); err != nil || len(decls) != 0 {
		t.Fatalf("absent hybrid manifest = (%v, %v), want empty and nil", decls, err)
	}
	if err := AddHybridDecl(home, "skill-h", "tag", "v1", "", []string{"alias-a"}); err != nil {
		t.Fatalf("add with absent manifest: %v", err)
	}
	path := HybridManifestPath(home)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	makeScopesStateUnreadable(t, path, false)
	if _, err := LoadHybridDecls(home); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("load unreadable hybrid manifest error = %v, want typed unreadable state", err)
	}
	if err := AddHybridDecl(home, "skill-b", "tag", "v2", "", []string{"alias-b"}); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("add over unreadable hybrid manifest error = %v, want typed unreadable refusal", err)
	}
	if err := RemoveHybridDecl(home, "skill-h"); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("remove from unreadable hybrid manifest error = %v, want typed unreadable refusal", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(original) {
		t.Fatalf("unreadable hybrid manifest changed = (%q, %v), want original bytes", got, err)
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
