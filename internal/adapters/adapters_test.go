package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeSkill(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".csk-install.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshProjectMirrorsAndPrunes(t *testing.T) {
	project := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	makeSkill(t, canonical, "skill-b")

	agents := []string{"claude_code", "cursor"}
	groups := []Group{{Root: canonical, Skills: []string{"skill-a", "skill-b"}}}
	if err := RefreshProject(project, agents, groups, "copy"); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".claude/skills/skill-a/SKILL.md", ".cursor/rules/skill-b/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("mirror missing: %s", rel)
		}
	}

	// drop skill-b: managed entry must be pruned
	groups = []Group{{Root: canonical, Skills: []string{"skill-a"}}}
	if err := RefreshProject(project, agents, groups, "copy"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(project, ".claude", "skills", "skill-b")); err == nil {
		t.Fatal("stale managed entry survived")
	}
	if _, err := os.Stat(filepath.Join(project, ".claude", "skills", "skill-a")); err != nil {
		t.Fatal("expected entry pruned")
	}
}

func TestUnmanagedConflictRefused(t *testing.T) {
	project := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")

	// a hand-placed directory without a marker in the adapter root
	foreign := filepath.Join(project, ".claude", "skills", "skill-a")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreign, "mine.md"), []byte("hands off"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := RefreshProject(project, []string{"claude_code"}, []Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy")
	if err == nil || !strings.Contains(err.Error(), "not managed") {
		t.Fatalf("err = %v, want unmanaged conflict", err)
	}
	if _, statErr := os.Stat(filepath.Join(foreign, "mine.md")); statErr != nil {
		t.Fatal("foreign content was destroyed")
	}
}

func TestMarkerDirectoryIsAdoptable(t *testing.T) {
	project := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	// a copied directory carrying an install marker counts as ours
	adopted := filepath.Join(project, ".claude", "skills", "skill-a")
	if err := os.MkdirAll(adopted, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adopted, ".csk-install.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RefreshProject(project, []string{"claude_code"}, []Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy"); err != nil {
		t.Fatalf("marker directory must be adoptable: %v", err)
	}
}

func TestNativeDiscoveryAgentsGetNoProjectMirror(t *testing.T) {
	project := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	if err := RefreshProject(project, []string{"opencode", "windsurf"}, []Group{{Root: canonical, Skills: []string{"skill-a"}}}, "copy"); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(project)
	for _, entry := range entries {
		if entry.Name() != ".agents" {
			t.Fatalf("unexpected project entry for native-discovery agent: %s", entry.Name())
		}
	}
}

func TestRefreshGlobalNativeDiscoveryMirror(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	canonical := filepath.Join(home, "global", "skills")
	makeSkill(t, canonical, "skill-g")
	if err := RefreshGlobal(home, userHome, []string{"claude_code", "opencode"}, []string{"skill-g"}, "copy"); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".claude/skills/skill-g/SKILL.md", ".agents/skills/skill-g/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(userHome, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("global mirror missing: %s", rel)
		}
	}
}

func TestSymlinkModeUsesRelativeLinks(t *testing.T) {
	project := t.TempDir()
	canonical := filepath.Join(project, ".agents", "skills")
	makeSkill(t, canonical, "skill-a")
	err := RefreshProject(project, []string{"claude_code"}, []Group{{Root: canonical, Skills: []string{"skill-a"}}}, "symlink")
	if err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	link := filepath.Join(project, ".claude", "skills", "skill-a")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("expected a symlink: %v", err)
	}
	if filepath.IsAbs(target) {
		t.Fatalf("symlink must be relative: %s", target)
	}
}

func TestGitignoreEntriesAndUnknownAgents(t *testing.T) {
	entries := RequiredGitignoreEntries([]string{"claude_code", "opencode", "bogus"})
	joined := strings.Join(entries, " ")
	if !strings.Contains(joined, ".agents/") || !strings.Contains(joined, ".claude/skills/") {
		t.Fatalf("entries: %v", entries)
	}
	if strings.Contains(joined, "bogus") {
		t.Fatalf("unknown agent produced an entry: %v", entries)
	}
	unknown := UnknownAgents([]string{"claude_code", "bogus", "bogus"})
	if len(unknown) != 1 || unknown[0] != "bogus" {
		t.Fatalf("unknown: %v", unknown)
	}
}
