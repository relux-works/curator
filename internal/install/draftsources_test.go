package install

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/snapshot"
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
	m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
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
	result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
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
	result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_lock_stale") {
		t.Fatalf("result = %+v, want source_lock_stale", result)
	}
}

func TestDraftInstallReplaysMissingLocalSnapshot(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	plan := resolveDraftForInstall(t, project, home, payload)
	for _, tree := range plan.Frozen {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(tree))); err != nil {
			t.Fatal(err)
		}
	}
	cfg := draftTestConfig(home, t.TempDir())
	lockBefore, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("result = %+v, want replayed install", result)
	}
	lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfter) != string(lockBefore) {
		t.Fatal("install replay modified Skillfile.lock.json")
	}
	member, _ := plan.Lock.Find("review")
	if _, err := snapshot.OpenLocal(home, member.Package.Snapshot); err != nil {
		t.Fatalf("replay did not restore the locked snapshot: %v", err)
	}
}

func TestDraftInstallUsesPinnedBytes(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	plan := resolveDraftForInstall(t, project, home, payload)
	cfg := draftTestConfig(home, t.TempDir())
	first := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
	if first.Status != "ok" {
		t.Fatalf("first install = %+v", first)
	}
	// Live bytes gain an extra file, but the lock still pins the old
	// snapshot: a frozen install without refresh must stay green and
	// must not adopt the live extra.
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("live"), 0o644); err != nil {
		t.Fatal(err)
	}
	second := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
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

func TestSchema2InstallRequiresLockAndLegacyStillWorks(t *testing.T) {
	// Schema 1 keeps its existing install path.
	e := newEnv(t)
	e.skill("legacy")
	e.declare("legacy")
	got := e.install(Options{DryRun: true})
	if got.Status != "ok" {
		t.Fatalf("legacy dry-run = %+v, want ok", got)
	}
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	cfg := draftTestConfig(home, t.TempDir())
	result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
	if result.Status != "failed" {
		t.Fatalf("schema-2 install without a lock = %+v, want failed", result)
	}
	joined := strings.Join(result.Errors, ";")
	if !strings.Contains(joined, "source_snapshot_unavailable") {
		t.Fatalf("schema-2 errors = %q, want the missing-lock diagnostic", joined)
	}
	resolveDraftForInstall(t, project, home, payload)
	result = Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("schema-2 install after resolve = %+v, want ok", result)
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
	m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
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

func setupDraftGitSource(t *testing.T, sourceDeclaration string) (project, home, repo, remote string, plan *closure.DraftPlan) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":` + sourceDeclaration + `},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	project, home, _ = draftProject(t, payload, nil)
	repo = t.TempDir()
	writeDraftPackage(t, filepath.Join(repo, "skills", "review"), "review")
	testGit(t, repo, "init", "-q", "-b", "main")
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "locked source")
	testGit(t, repo, "tag", "v1")
	remote = filepath.Join(t.TempDir(), "source.git")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	testGit(t, remote, "init", "--bare", "-q")
	testGit(t, repo, "remote", "add", "origin", remote)
	testGit(t, repo, "push", "-q", "origin", "main", "--tags")
	m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err = closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload),
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), plan.Lock); err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(DraftBindingsPath(home, project), plan.Bindings); err != nil {
		t.Fatal(err)
	}
	return project, home, repo, remote, plan
}

func freshDraftGitHome(t *testing.T, project string, lock *sourcelock.Lock, remote string) string {
	t.Helper()
	home := t.TempDir()
	repo := t.TempDir()
	testGit(t, repo, "init", "-q", "-b", "main")
	testGit(t, repo, "remote", "add", "origin", remote)
	bindings, err := sourcelock.NewBindings(lock.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: repo}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(DraftBindingsPath(home, project), bindings); err != nil {
		t.Fatal(err)
	}
	return home
}

// seedDraftReplaySource pre-creates the private replay repository production
// derives for one declared Git source and plants a repo-local url.insteadOf
// redirect from every planned endpoint to the fixture repository. Production
// resolves the declared endpoint and fetches exactly the locked object ID
// there (fetchLockedCommitFromDeclaredSource); a repo-local redirect is
// honoured by that isolated fetch on every platform, while a PATH git shim
// is not executable on Windows and GIT_CONFIG_* overrides are refused by the
// isolated lane. The derivation mirrors production (projectAbs + NUL + alias
// + NUL + identity, hex prefix under <home>/source-replay); if it drifts,
// the seed lands elsewhere and the test fails closed with
// source_snapshot_unavailable.
func seedDraftReplaySource(t *testing.T, cfg *config.Config, project, alias, gitURL, repository, localRepo string) {
	t.Helper()
	policy, err := loadReplaySourcePolicy(cfg)
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := config.ResolveRepositoryEndpoints(policy, gitURL, repository)
	if err != nil {
		t.Fatal(err)
	}
	projectAbs, err := filepath.Abs(project)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(projectAbs + "\x00" + alias + "\x00" + resolution.Identity))
	repo := filepath.Join(cfg.Home(), "source-replay", hex.EncodeToString(sum[:])[:20])
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	testGit(t, repo, "init", "-q")
	var redirect strings.Builder
	for _, attempt := range resolution.Attempts {
		target, ok := attempt.ConnectionURL()
		if !ok || target == "" {
			t.Fatal("replay endpoint has no connection target")
		}
		fmt.Fprintf(&redirect, "[url \"%s\"]\n\tinsteadOf = %s\n", draftReplayFileURL(localRepo), target)
	}
	configPath := filepath.Join(repo, ".git", "config")
	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	payload = append(payload, []byte(redirect.String())...)
	if err := os.WriteFile(configPath, payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func draftReplayFileURL(path string) string {
	slashPath := strings.ReplaceAll(path, `\`, "/")
	if len(slashPath) >= 3 && slashPath[1] == ':' && slashPath[2] == '/' {
		slashPath = "/" + slashPath
	}
	if strings.HasPrefix(slashPath, "//") {
		if host, rest, ok := strings.Cut(strings.TrimPrefix(slashPath, "//"), "/"); ok {
			return (&url.URL{Scheme: "file", Host: host, Path: "/" + rest}).String()
		}
	}
	return (&url.URL{Scheme: "file", Path: slashPath}).String()
}

func TestDraftReplayFileURLUsesPortableGitConfigSyntax(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "posix", path: "/tmp/curator/source.git", want: "file:///tmp/curator/source.git"},
		{name: "windows-drive", path: `C:\Users\runner\source.git`, want: "file:///C:/Users/runner/source.git"},
		{name: "windows-unc", path: `\\server\share\source.git`, want: "file://server/share/source.git"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := draftReplayFileURL(test.path); got != test.want {
				t.Fatalf("draftReplayFileURL(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func writeDraftReplayRepositoryPolicy(t *testing.T, home string) {
	t.Helper()
	payload := `{"schema_version":1,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none"}}}`
	if err := os.WriteFile(filepath.Join(home, config.SourcePolicyFileName), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade(t *testing.T) {
	gitSources := []struct {
		name        string
		declaration string
	}{
		{name: "git-tag", declaration: `{"git":"https://example.org/kit.git","tag":"v1"}`},
		{name: "repository", declaration: `{"repository":"example.org/kit","tag":"v1"}`},
	}
	for _, source := range gitSources {
		for _, operation := range []struct {
			name  string
			fetch bool
		}{{name: "install"}, {name: "upgrade", fetch: true}} {
			t.Run(source.name+"/"+operation.name, func(t *testing.T) {
				project, _, _, remote, plan := setupDraftGitSource(t, source.declaration)
				home := freshDraftGitHome(t, project, plan.Lock, remote)
				before, err := os.ReadFile(sourcelock.PathIn(project))
				if err != nil {
					t.Fatal(err)
				}
				result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Fetch: operation.fetch, Platform: installPlatform()})
				if result.Status != "ok" {
					t.Fatalf("%s = %+v", operation.name, result)
				}
				after, err := os.ReadFile(sourcelock.PathIn(project))
				if err != nil {
					t.Fatal(err)
				}
				if string(after) != string(before) {
					t.Fatalf("%s changed Skillfile.lock.json", operation.name)
				}
				member, _ := plan.Lock.Find("review")
				tree := snapshot.Dir(home, member.Package.Repository, member.Package.Commit.Hex)
				got, err := os.ReadFile(filepath.Join(tree, "skills", "review", "references", "info.md"))
				if err != nil || string(got) != "context" {
					t.Fatalf("replayed locked bytes = %q, %v; want context", got, err)
				}
			})
		}
	}
	for _, operation := range []struct {
		name  string
		fetch bool
	}{{name: "install"}, {name: "upgrade", fetch: true}} {
		t.Run("path/"+operation.name, func(t *testing.T) {
			payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
			project, oldHome, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
			plan := resolveDraftForInstall(t, project, oldHome, payload)
			home := t.TempDir()
			before, err := os.ReadFile(sourcelock.PathIn(project))
			if err != nil {
				t.Fatal(err)
			}
			result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Fetch: operation.fetch, Platform: installPlatform()})
			if result.Status != "ok" {
				t.Fatalf("%s = %+v", operation.name, result)
			}
			after, err := os.ReadFile(sourcelock.PathIn(project))
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(before) {
				t.Fatalf("%s changed Skillfile.lock.json", operation.name)
			}
			if _, err := snapshot.OpenLocal(home, plan.Lock.Members[0].Package.Snapshot); err != nil {
				t.Fatalf("path replay did not publish the locked snapshot: %v", err)
			}
		})
	}
}

func TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings(t *testing.T) {
	gitSources := []struct {
		name        string
		declaration string
		git         string
		repository  string
	}{
		{name: "git-tag", declaration: `{"git":"https://example.org/kit.git","tag":"v1"}`, git: "https://example.org/kit.git"},
		{name: "repository", declaration: `{"repository":"example.org/kit","tag":"v1"}`, repository: "example.org/kit"},
	}
	for _, source := range gitSources {
		for _, operation := range []struct {
			name  string
			fetch bool
		}{{name: "install"}, {name: "upgrade", fetch: true}} {
			t.Run(source.name+"/"+operation.name, func(t *testing.T) {
				project, _, _, remote, plan := setupDraftGitSource(t, source.declaration)
				home := t.TempDir()
				if source.name == "repository" {
					writeDraftReplayRepositoryPolicy(t, home)
				}
				cfg := draftTestConfig(home, t.TempDir())
				seedDraftReplaySource(t, cfg, project, "s", source.git, source.repository, remote)
				before, err := os.ReadFile(sourcelock.PathIn(project))
				if err != nil {
					t.Fatal(err)
				}
				result := Project(cfg, project, "test", Options{DryRun: true, Fetch: operation.fetch, Platform: installPlatform()})
				if result.Status != "ok" {
					t.Fatalf("%s without machine bindings = %+v", operation.name, result)
				}
				after, err := os.ReadFile(sourcelock.PathIn(project))
				if err != nil {
					t.Fatal(err)
				}
				if string(after) != string(before) {
					t.Fatalf("%s changed Skillfile.lock.json", operation.name)
				}
				member, _ := plan.Lock.Find("review")
				tree := snapshot.Dir(home, member.Package.Repository, member.Package.Commit.Hex)
				got, err := os.ReadFile(filepath.Join(tree, "skills", "review", "references", "info.md"))
				if err != nil || string(got) != "context" {
					t.Fatalf("declared-source replay bytes = %q, %v; want locked content", got, err)
				}
			})
		}
	}
}

// TestDraftReplaySeedSurvivesIsolatedFetch pins the redirect channel the
// no-bindings rows depend on: a repo-local url.insteadOf is honoured by the
// isolated exact-SHA fetch, which refuses user/system configuration and
// ambient GIT_CONFIG_* overrides. If the lane ever stops reading repo-local
// configuration, the fresh-machine rows lose their declared source on every
// platform; this test names that channel directly.
func TestDraftReplaySeedSurvivesIsolatedFetch(t *testing.T) {
	project, _, _, remote, plan := setupDraftGitSource(t, `{"git":"https://example.org/kit.git","tag":"v1"}`)
	home := t.TempDir()
	cfg := draftTestConfig(home, t.TempDir())
	seedDraftReplaySource(t, cfg, project, "s", "https://example.org/kit.git", "", remote)
	entries, err := os.ReadDir(filepath.Join(home, "source-replay"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("seeded replay repositories = %v, %v; want exactly one", entries, err)
	}
	member, ok := plan.Lock.Find("review")
	if !ok {
		t.Fatal("locked review member missing")
	}
	repo := filepath.Join(home, "source-replay", entries[0].Name())
	if err := gitops.FetchCommitFromURLIsolated(repo, "https://example.org/kit.git", member.Package.Commit.Hex); err != nil {
		t.Fatalf("isolated fetch through the seeded redirect = %v", err)
	}
}

func TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock(t *testing.T) {
	project, _, _, _, plan := setupDraftGitSource(t, `{"repository":"example.org/kit","tag":"v1"}`)
	home := t.TempDir()
	before, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Platform: installPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_unavailable") {
		t.Fatalf("install without a reachable Git endpoint = %+v, want source_snapshot_unavailable", result)
	}
	after, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("unavailable Git replay modified Skillfile.lock.json")
	}
	member, _ := plan.Lock.Find("review")
	if _, err := snapshot.AuthenticateGit(home, member.Package.Repository, "", member.Package.Commit.Hex); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
		t.Fatalf("unavailable Git replay left a snapshot published: %v", err)
	}
}

func TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution(t *testing.T) {
	project, _, repo, remote, plan := setupDraftGitSource(t, `{"git":"https://example.org/kit.git","tag":"v1"}`)
	member, _ := plan.Lock.Find("review")
	lockedCommit := member.Package.Commit.Hex
	info := filepath.Join(repo, "skills", "review", "references", "info.md")
	if err := os.WriteFile(info, []byte("moved tag content"), 0o644); err != nil {
		t.Fatal(err)
	}
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "move v1")
	testGit(t, repo, "tag", "-f", "v1")
	moved, err := gitops.Resolve(repo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	if moved.Commit == lockedCommit {
		t.Fatal("fixture tag did not move away from the locked commit")
	}
	testGit(t, repo, "push", "-q", "--force", "origin", "main", "refs/tags/v1:refs/tags/v1")

	home := freshDraftGitHome(t, project, plan.Lock, remote)
	lockBefore, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Fetch: true, Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("upgrade after moved tag = %+v", result)
	}
	lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfter) != string(lockBefore) {
		t.Fatal("upgrade after moved tag changed Skillfile.lock.json")
	}
	tree := snapshot.Dir(home, member.Package.Repository, lockedCommit)
	got, err := os.ReadFile(filepath.Join(tree, "skills", "review", "references", "info.md"))
	if err != nil || string(got) != "context" {
		t.Fatalf("replayed bytes = %q, %v; want bytes from locked commit %s", got, err, lockedCommit)
	}
}

func TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot(t *testing.T) {
	project, _, _, remote, plan := setupDraftGitSource(t, `{"git":"https://example.org/kit.git","tag":"v1"}`)
	members := append([]sourcelock.Member(nil), plan.Lock.Members...)
	members[0].ContentSHA256 = "sha256:" + strings.Repeat("0", 64)
	if members[0].ContentSHA256 == plan.Lock.Members[0].ContentSHA256 {
		members[0].ContentSHA256 = "sha256:" + strings.Repeat("1", 64)
	}
	forged, err := sourcelock.New(plan.Lock.ManifestSHA256, members)
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), forged); err != nil {
		t.Fatal(err)
	}
	home := freshDraftGitHome(t, project, plan.Lock, remote)
	bindings, err := sourcelock.ReadBindings(DraftBindingsPath(home, project))
	if err != nil {
		t.Fatal(err)
	}
	currentBindings, err := sourcelock.NewBindings(forged.LockSHA256, bindings.Sources)
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(DraftBindingsPath(home, project), currentBindings); err != nil {
		t.Fatal(err)
	}
	lockBefore, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Platform: installPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
		t.Fatalf("install = %+v, want source_snapshot_changed", result)
	}
	lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfter) != string(lockBefore) {
		t.Fatal("Git content-hash mismatch modified Skillfile.lock.json")
	}
	member, _ := forged.Find("review")
	if _, err := snapshot.AuthenticateGit(home, member.Package.Repository, bindings.Sources["s"].Location, member.Package.Commit.Hex); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
		t.Fatalf("mismatching Git package was published to the snapshot store: %v", err)
	}
}

func TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock(t *testing.T) {
	t.Run("changed-path-identity-even-when-context-hash-is-same", func(t *testing.T) {
		payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
		project, oldHome, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
		plan := resolveDraftForInstall(t, project, oldHome, payload)
		packageDir := filepath.Join(project, "skills", "review")
		spec, err := skillspec.Load(packageDir)
		if err != nil {
			t.Fatal(err)
		}
		beforeHash, err := closure.ContentHashFor(packageDir, spec)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(packageDir, "unprojected-extra.txt"), []byte("new package byte"), 0o644); err != nil {
			t.Fatal(err)
		}
		afterHash, err := closure.ContentHashFor(packageDir, spec)
		if err != nil {
			t.Fatal(err)
		}
		if beforeHash != afterHash {
			t.Fatalf("fixture content hash changed: %s -> %s", beforeHash, afterHash)
		}
		home := t.TempDir()
		lockBefore, err := os.ReadFile(sourcelock.PathIn(project))
		if err != nil {
			t.Fatal(err)
		}
		result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
			t.Fatalf("install = %+v, want source_snapshot_changed", result)
		}
		lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
		if err != nil {
			t.Fatal(err)
		}
		if string(lockAfter) != string(lockBefore) {
			t.Fatal("changed path replay modified Skillfile.lock.json")
		}
		member, _ := plan.Lock.Find("review")
		if member.ContentSHA256 != beforeHash {
			t.Fatalf("locked hash = %s, computed %s", member.ContentSHA256, beforeHash)
		}
		if _, err := snapshot.OpenLocal(home, member.Package.Snapshot); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
			t.Fatalf("mismatching package was published to the snapshot store: %v", err)
		}
	})

	t.Run("content-hash-mismatch-with-matching-package-identity", func(t *testing.T) {
		payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
		project, oldHome, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
		plan := resolveDraftForInstall(t, project, oldHome, payload)
		members := append([]sourcelock.Member(nil), plan.Lock.Members...)
		members[0].ContentSHA256 = "sha256:" + strings.Repeat("0", 64)
		if members[0].ContentSHA256 == plan.Lock.Members[0].ContentSHA256 {
			members[0].ContentSHA256 = "sha256:" + strings.Repeat("1", 64)
		}
		forged, err := sourcelock.New(plan.Lock.ManifestSHA256, members)
		if err != nil {
			t.Fatal(err)
		}
		if err := sourcelock.Write(sourcelock.PathIn(project), forged); err != nil {
			t.Fatal(err)
		}
		home := t.TempDir()
		lockBefore, err := os.ReadFile(sourcelock.PathIn(project))
		if err != nil {
			t.Fatal(err)
		}
		result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
			t.Fatalf("install = %+v, want source_snapshot_changed", result)
		}
		lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
		if err != nil {
			t.Fatal(err)
		}
		if string(lockAfter) != string(lockBefore) {
			t.Fatal("content-hash mismatch modified Skillfile.lock.json")
		}
		if _, err := snapshot.OpenLocal(home, plan.Lock.Members[0].Package.Snapshot); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
			t.Fatalf("content-hash mismatch was published to the snapshot store: %v", err)
		}
	})

	t.Run("unreachable-path", func(t *testing.T) {
		payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
		project, oldHome, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
		resolveDraftForInstall(t, project, oldHome, payload)
		if err := os.RemoveAll(filepath.Join(project, "skills", "review")); err != nil {
			t.Fatal(err)
		}
		home := t.TempDir()
		lockBefore, err := os.ReadFile(sourcelock.PathIn(project))
		if err != nil {
			t.Fatal(err)
		}
		result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{DryRun: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_unavailable") {
			t.Fatalf("install = %+v, want source_snapshot_unavailable", result)
		}
		lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
		if err != nil {
			t.Fatal(err)
		}
		if string(lockAfter) != string(lockBefore) {
			t.Fatal("unavailable path replay modified Skillfile.lock.json")
		}
	})
}

func TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftForInstall(t, project, home, payload)
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	result := Project(draftTestConfig(home, t.TempDir()), project, "test", Options{FixGitignore: true, Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("install with gitignore repair = %+v", result)
	}
	ignore, err := os.ReadFile(filepath.Join(project, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(ignore), sourcelock.FileName) {
		t.Fatalf("Curator added the lock to .gitignore: %s", ignore)
	}
	if _, err := os.Stat(sourcelock.PathIn(project)); err != nil {
		t.Fatalf("lock is not project-owned: %v", err)
	}
	if relative, err := filepath.Rel(project, DraftBindingsPath(home, project)); err != nil || relative == "." || !strings.HasPrefix(relative, "..") {
		t.Fatalf("machine bindings path %q is not outside the project: %v", relative, err)
	}
	cmd := exec.Command("git", "-C", project, "check-ignore", "-q", sourcelock.FileName)
	err = cmd.Run()
	if err == nil {
		t.Fatal("Skillfile.lock.json is gitignored after Curator repaired generated paths")
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("git check-ignore exit = %v, want 1 (not ignored)", err)
	}
}

func draftInstallResult(t *testing.T, project, home string) string {
	t.Helper()
	cfg := draftTestConfig(home, t.TempDir())
	result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
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
		result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
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
		result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
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
		result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
			t.Fatalf("result = %+v, want source_snapshot_changed", result)
		}
	})
}
