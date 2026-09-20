package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/sourcelock"
)

func testGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	gitFixture(t, dir, args, gitArgs, []string{
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	})
	return dir
}

func writeDraftPackage(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := map[string]any{"schema_version": 4, "capabilities": map[string]any{}, "commands": map[string]any{}, "dependencies": map[string]any{"skills": map[string]any{}}}
	payload, _ := json.Marshal(spec)
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func draftProject(t *testing.T, payload string, skills map[string]string) (string, string, []byte) {
	t.Helper()
	project := t.TempDir()
	home := t.TempDir()
	for dir, name := range skills {
		writeDraftPackage(t, filepath.Join(project, filepath.FromSlash(dir)), name)
	}
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	// The install .gitignore gate runs before closure: init a repository
	// and ignore every managed output so the draft lane is reached.
	testGit(t, project, "init", "-q")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.claude/skills/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return project, home, []byte(payload)
}

func resolveDraftForInstall(t *testing.T, project, home, payload string) *closure.DraftPlan {
	t.Helper()
	m, err := manifest.ParseBytesWithOptions([]byte(payload), filepath.Join(project, "Skillfile.json"), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload)})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), plan.Lock); err != nil {
		t.Fatal(err)
	}
	return plan
}

func draftTestConfig(home, skillsRoot string) *config.Config {
	return &config.Config{Path: filepath.Join(home, "config.json"), SkillsRoot: skillsRoot, DefaultAgents: []string{"claude_code"}, AdapterMode: "auto"}
}

func TestDraftInstallMissingLockFails(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	cfg := draftTestConfig(home, t.TempDir())
	result := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_unavailable") {
		t.Fatalf("result = %+v, want source_snapshot_unavailable", result)
	}
}

func TestDraftInstallStaleLockFails(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftForInstall(t, project, home, payload)
	// Change the manifest without refresh: frozen install must fail stale
	// and leave the lock file untouched.
	changed := `{"schema_version":2,"sources":{"s":{"path":"./skills"}},"skills":[{"from":"s","directory":".","include":["review"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := draftTestConfig(home, t.TempDir())
	result := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_lock_stale") {
		t.Fatalf("result = %+v, want source_lock_stale", result)
	}
}

func TestDraftInstallMissingSnapshotFails(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	plan := resolveDraftForInstall(t, project, home, payload)
	for _, tree := range plan.Frozen {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(tree))); err != nil {
			t.Fatal(err)
		}
	}
	cfg := draftTestConfig(home, t.TempDir())
	result := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_unavailable") {
		t.Fatalf("result = %+v, want source_snapshot_unavailable", result)
	}
}

func TestDraftInstallUsesPinnedBytes(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	plan := resolveDraftForInstall(t, project, home, payload)
	cfg := draftTestConfig(home, t.TempDir())
	first := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
	if first.Status != "ok" {
		t.Fatalf("first install = %+v", first)
	}
	// Live bytes gain an extra file, but the lock still pins the old
	// snapshot: a frozen install without refresh must stay green and
	// must not adopt the live extra.
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("live"), 0o644); err != nil {
		t.Fatal(err)
	}
	second := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
	if second.Status != "ok" {
		t.Fatalf("pinned reinstall = %+v, want ok on locked bytes", second)
	}
	locked, _ := plan.Lock.Find("review")
	frozen, err := closure.OpenDraftFrozen(home, plan.Lock, closure.FrozenOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(frozen["review"], "references", "extra.md")); !os.IsNotExist(err) {
		t.Fatalf("frozen tree adopted live extra.md")
	}
	_ = locked
}

func TestLegacyInstallUntouchedWhenDraftOff(t *testing.T) {
	// Golden for the frozen v1 path: a schema-2 manifest without opt-in
	// fails at parse time with the upgrade diagnostic, and a schema-1
	// install behaves exactly as before (dry-run ok).
	e := newEnv(t)
	e.skill("legacy")
	e.declare("legacy")
	got := e.install(Options{DryRun: true})
	if got.Status != "ok" {
		t.Fatalf("legacy dry-run = %+v, want ok", got)
	}
	project := t.TempDir()
	home := t.TempDir()
	writeDraftPackage(t, filepath.Join(project, "skills", "review"), "review")
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(`{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := draftTestConfig(home, t.TempDir())
	result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
	if result.Status != "failed" {
		t.Fatalf("draft-off install = %+v, want failed", result)
	}
	joined := strings.Join(result.Errors, ";")
	if !strings.Contains(joined, "source_selection_invalid") && !strings.Contains(joined, "schema_version") && !strings.Contains(joined, "newer tool") {
		t.Fatalf("draft-off errors = %q, want schema-2 refusal", joined)
	}
}

// setupGitInstall resolves one Git-selected review package through the
// production closure and writes its lock, returning the project, home,
// and cached repository root for production-entry install checks.
func setupGitInstall(t *testing.T) (project, home, repo, repoRoot string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	project, home, _ = draftProject(t, payload, nil)
	repo = t.TempDir()
	writeDraftPackage(t, filepath.Join(repo, "skills", "review"), "review")
	testGit(t, repo, "init", "-q", "-b", "main")
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "fixture")
	testGit(t, repo, "tag", "v1")
	m, err := manifest.ParseBytesWithOptions([]byte(payload), filepath.Join(project, "Skillfile.json"), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload),
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), plan.Lock); err != nil {
		t.Fatal(err)
	}
	// Production resolve publishes the machine bindings beside the lock in
	// the same attempt: frozen consumption derives the proving repository
	// and the declared identity from them, never from live state.
	bindings, err := sourcelock.NewBindings(plan.Lock.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: repo}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(DraftBindingsPath(home, project), bindings); err != nil {
		t.Fatal(err)
	}
	resolved, err := gitops.Resolve(repo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	repoRoot = filepath.Join(home, "cache", "example.org", "kit", resolved.Commit, "snapshot")
	return project, home, repo, repoRoot
}

func draftInstallResult(t *testing.T, project, home string) string {
	t.Helper()
	cfg := draftTestConfig(home, t.TempDir())
	result := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	return strings.Join(result.Errors, ";")
}

func TestDraftInstallGitPinnedSubtree(t *testing.T) {
	project, home, repo, repoRoot := setupGitInstall(t)
	draftInstallResult(t, project, home)
	// Live repository growth without a tag move must not enter the frozen
	// install: the lock still pins the v1 commit subtree.
	writeDraftPackage(t, filepath.Join(repo, "skills", "docs"), "docs")
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "unselected")
	draftInstallResult(t, project, home)
	subtree := filepath.Join(repoRoot, "skills", "review")
	if _, err := os.Stat(filepath.Join(subtree, "SKILL.md")); err != nil {
		t.Fatalf("locked subtree missing: %v", err)
	}
}

func TestDraftInstallGitTamperRefused(t *testing.T) {
	t.Run("tampered-file", func(t *testing.T) {
		project, home, _, repoRoot := setupGitInstall(t)
		if err := os.WriteFile(filepath.Join(repoRoot, "skills", "review", "references", "info.md"), []byte("tampered"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg := draftTestConfig(home, t.TempDir())
		result := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
			t.Fatalf("result = %+v, want source_snapshot_changed", result)
		}
	})
	t.Run("symlink", func(t *testing.T) {
		project, home, _, repoRoot := setupGitInstall(t)
		if err := os.Symlink(
			filepath.Join(repoRoot, "skills", "review", "SKILL.md"),
			filepath.Join(repoRoot, "skills", "review", "references", "evil.md")); err != nil {
			t.Fatal(err)
		}
		cfg := draftTestConfig(home, t.TempDir())
		result := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
			t.Fatalf("result = %+v, want source_snapshot_changed", result)
		}
	})
	t.Run("partial-cache", func(t *testing.T) {
		project, home, _, repoRoot := setupGitInstall(t)
		if err := os.Remove(filepath.Join(repoRoot, "skills", "review", "references", "info.md")); err != nil {
			t.Fatal(err)
		}
		cfg := draftTestConfig(home, t.TempDir())
		result := Project(cfg, project, "test", Options{DryRun: true, DraftSourcesV1: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
			t.Fatalf("result = %+v, want source_snapshot_changed", result)
		}
	})
}

func TestDraftSourcesSwitch(t *testing.T) {
	if !DraftSourcesEnabled(func(string) string { return "1" }) {
		t.Fatalf("enabled with 1")
	}
	if DraftSourcesEnabled(func(string) string { return "" }) {
		t.Fatalf("enabled with empty")
	}
	if DraftSourcesEnabled(nil) {
		t.Fatalf("enabled with nil")
	}
}
