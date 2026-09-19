package main

// Draft source closure at the CLI production entry.
//
// These tests drive `project resolve` / `project refresh` and `install
// --dry-run` through run() with the draft switch set in-process. The lock
// is always created through the CLI itself, never by calling the closure
// helper under test. Git selections additionally run behind a stand-in
// git on PATH that rewrites one fixture https URL to a local bare
// repository; everything else — manifest and closure evaluation,
// endpoint planning, extraction, audit — is the production path. None of
// these tests run in parallel: they mutate process environment per run.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/sourcelock"
)

// withDraftSourcesSwitch pins the draft lane for one test and restores the
// process environment afterwards.
func withDraftSourcesSwitch(t *testing.T) {
	t.Helper()
	if value, ok := os.LookupEnv(install.EnvDraftSourcesV1); ok {
		t.Cleanup(func() { _ = os.Setenv(install.EnvDraftSourcesV1, value) })
	} else {
		t.Cleanup(func() { _ = os.Unsetenv(install.EnvDraftSourcesV1) })
	}
	if err := os.Setenv(install.EnvDraftSourcesV1, "1"); err != nil {
		t.Fatal(err)
	}
}

func writeCLISkill(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), []byte(`{"schema_version":4,"capabilities":{},"commands":{},"dependencies":{"skills":{}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeCLIScriptSkill writes one script package with an exported command
// and declared runtime roots, mirroring the reviewer's fixture shape: the
// runtime script bytes are excluded from the projected context hash, so
// only complete snapshot authentication can refuse their tampering.
func writeCLIScriptSkill(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := `{"schema_version":4,"capabilities":{},"commands":{"tool":{"type":"script","unix_path":"scripts/tool.sh"}},"dependencies":{"skills":{}},"runtime_roots":["scripts"]}`
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "tool.sh"), []byte("#!/bin/sh\necho original\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// setupCLIProject bootstraps a config, adds one project, and returns the
// config path and project root. Callers overwrite Skillfile.json with
// their own payload afterwards.
func setupCLIProject(t *testing.T, root string) (configPath, project string) {
	t.Helper()
	configPath = filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "skills-root")
	project = filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if code := runCode(t, configPath, []string{"bootstrap", "--non-interactive", "--skills-root", skillsRoot}); code != exitOK {
		t.Fatalf("bootstrap = %d", code)
	}
	if code := runCode(t, configPath, []string{"project", "add", "app", project, "--agents", "codex_cli"}); code != exitOK {
		t.Fatalf("project add = %d", code)
	}
	runGit(t, project, "init", "-q", "-b", "main")
	return configPath, project
}

func TestProjectResolveLocalCreatesLockThroughCLI(t *testing.T) {
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(project, "skills", "review"), "review")

	if code := runCode(t, configPath, []string{"project", "resolve", "app"}); code != exitOK {
		t.Fatalf("project resolve = %d", code)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatalf("lock unreadable after CLI resolve: %v", err)
	}
	if len(lock.Members) != 1 || lock.Members[0].Name != "review" {
		t.Fatalf("members = %+v, want [review]", lock.Members)
	}
	firstSHA := lock.LockSHA256

	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	// Live growth without refresh stays pinned: the frozen install must
	// stay green on the locked bytes.
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("live"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("pinned reinstall = %d\nstderr:\n%s", code, stderr)
	}
	// Explicit refresh reselects: the lock generation moves.
	if code := runCode(t, configPath, []string{"project", "refresh", "app"}); code != exitOK {
		t.Fatalf("project refresh = %d", code)
	}
	refreshed, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.LockSHA256 == firstSHA {
		t.Fatalf("refresh left the lock unchanged after a runtime-only edit")
	}
	// A missing locked snapshot fails instead of rescanning.
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install after refresh = %d\nstderr:\n%s", code, stderr)
	}
	if err := os.Remove(filepath.Join(project, "Skillfile.lock.json")); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitFail || !strings.Contains(stderr, "source_snapshot_unavailable") {
		t.Fatalf("missing-lock install = %d, stderr %q, want source_snapshot_unavailable", code, stderr)
	}
	// A changed manifest without refresh fails stale.
	if code := runCode(t, configPath, []string{"project", "resolve", "app"}); code != exitOK {
		t.Fatalf("re-resolve = %d", code)
	}
	changed := `{"schema_version":2,"sources":{"s":{"path":"./skills"}},"skills":[{"from":"s","directory":".","include":["review"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitFail || !strings.Contains(stderr, "source_lock_stale") {
		t.Fatalf("stale install = %d, stderr %q, want source_lock_stale", code, stderr)
	}
}

func TestProjectResolveLegacyUntouched(t *testing.T) {
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(`{"schema_version":1,"agents":["codex_cli"],"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, _ := capture(t, configPath, "project", "resolve", "app")
	if code != exitOK {
		t.Fatalf("legacy project resolve = %d", code)
	}
	if !strings.Contains(stdout, "alias: app\n") || !strings.Contains(stdout, "skillfile: ") {
		t.Fatalf("legacy output changed:\n%s", stdout)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatalf("legacy resolve wrote a lock: %v", err)
	}
}

func TestProjectResolveDraftOffRefuses(t *testing.T) {
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(project, "skills", "review"), "review")
	if value, ok := os.LookupEnv(install.EnvDraftSourcesV1); ok && value == "1" {
		t.Skip("draft switch is set in the ambient environment")
	}
	if code, _, _ := capture(t, configPath, "project", "resolve", "app"); code != exitFail {
		t.Fatalf("draft-off resolve = %d, want %d", code, exitFail)
	}
}

// installFakeGitForDraftCLI writes a stand-in git that rewrites exactly
// one fixture https URL to a local bare repository and delegates
// everything else to the real git. It returns a PATH directory to prepend.
func installFakeGitForDraftCLI(t *testing.T, declaredURL, bare, realGit string) string {
	t.Helper()
	fakeDir := t.TempDir()
	script := "#!/bin/sh\nargs=\"\"; for arg in \"$@\"; do\n" +
		"if [ \"$arg\" = '" + declaredURL + "' ]; then arg='file://" + bare + "'; fi\n" +
		"args=\"$args\n$arg\"; done\noldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return fakeDir
}

func TestProjectResolveGitSelectionThroughCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	// Fixture repository holding one package under skills/review.
	work := filepath.Join(root, "kit-work")
	bare := filepath.Join(root, "kit.git")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(work, "skills", "review"), "review")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "fixture")
	runGit(t, work, "tag", "v1")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	out, err := exec.Command("git", "--git-dir", bare, "rev-parse", "v1^{commit}").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	commit := strings.TrimSpace(string(out))

	configPath, project := setupCLIProject(t, root)
	const declaredURL = "https://fixture.test/kit.git"
	payload := `{"schema_version":2,"sources":{"s":{"git":"` + declaredURL + `","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	fakeDir := installFakeGitForDraftCLI(t, declaredURL, bare, realGit)
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}

	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatalf("lock misses review: %+v", lock.Members)
	}
	if member.Package.Kind != "network-git" || member.Package.Directory != "skills/review" {
		t.Fatalf("package = %+v", member.Package)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	// A tampered cached file fails the frozen install at the CLI entry.
	home := filepath.Dir(configPath)
	contextFile := filepath.Join(home, "cache", "fixture.test", "kit", commit, "snapshot", "skills", "review", "references", "info.md")
	if err := os.WriteFile(contextFile, []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitFail || !strings.Contains(stderr, "source_snapshot_changed") {
		t.Fatalf("tampered install = %d, stderr %q, want source_snapshot_changed", code, stderr)
	}
}

// setupScriptGitCLI builds the reviewer's fixture shape through the
// production CLI: a local Git repository holding one script package with
// runtime roots, a bare clone behind a stand-in git that rewrites the
// declared fixture URL, and a lock created only through `project resolve`.
// It returns the config path, project root, locked commit, cached runtime
// script, and bare repository for tamper and pinned-consumption checks.
func setupScriptGitCLI(t *testing.T) (configPath, project, commit, cachedScript, bare string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	work := filepath.Join(root, "kit-work")
	bare = filepath.Join(root, "kit.git")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCLIScriptSkill(t, filepath.Join(work, "skills", "review"), "review")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "fixture")
	runGit(t, work, "tag", "v1")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	out, err := exec.Command("git", "--git-dir", bare, "rev-parse", "v1^{commit}").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	commit = strings.TrimSpace(string(out))

	configPath, project = setupCLIProject(t, root)
	const declaredURL = "https://fixture.test/kit.git"
	payload := `{"schema_version":2,"sources":{"s":{"git":"` + declaredURL + `","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	fakeDir := installFakeGitForDraftCLI(t, declaredURL, bare, realGit)
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}

	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	home := filepath.Dir(configPath)
	cachedScript = filepath.Join(home, "cache", "fixture.test", "kit", commit, "snapshot", "skills", "review", "scripts", "tool.sh")
	return configPath, project, commit, cachedScript, bare
}

// TestProjectResolveGitRuntimeTamperRefusedThroughCLI proves a tampered
// runtime script under the cache fails the frozen install at the CLI
// production entry in both dry-run and real modes, even though the locked
// content hash excludes runtime roots.
func TestProjectResolveGitRuntimeTamperRefusedThroughCLI(t *testing.T) {
	configPath, _, _, cachedScript, _ := setupScriptGitCLI(t)
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("baseline install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	if err := os.WriteFile(cachedScript, []byte("#!/bin/sh\necho TAMPERED\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitFail || !strings.Contains(stderr, "source_snapshot_changed") {
		t.Fatalf("tampered dry-run install = %d, stderr %q, want source_snapshot_changed", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitFail || !strings.Contains(stderr, "source_snapshot_changed") {
		t.Fatalf("tampered real install = %d, stderr %q, want source_snapshot_changed", code, stderr)
	}
}

// TestProjectResolveGitMissingMemberRefusedThroughCLI proves a removed
// inventory member under the cache fails the frozen install at the CLI
// production entry instead of installing a partial package.
func TestProjectResolveGitMissingMemberRefusedThroughCLI(t *testing.T) {
	configPath, _, _, cachedScript, _ := setupScriptGitCLI(t)
	if err := os.Remove(cachedScript); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitFail || !strings.Contains(stderr, "source_snapshot_changed") {
		t.Fatalf("missing-member install = %d, stderr %q, want source_snapshot_changed", code, stderr)
	}
}

// writeCLIConsumerSkill writes one local package whose skill requirement
// names a transitive provider resolved through the legacy lane. The
// provider URL is intentionally unroutable: any clone or fetch attempt
// fails the resolve, so a green resolve proves zero network acquisitions.
func writeCLIConsumerSkill(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := `{"schema_version":4,"capabilities":{},"commands":{},"dependencies":{"skills":{"provider":{"git":"https://invalid.invalid/provider.git","ref":{"kind":"tag","value":"v1"}}}}}`
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// initCLIProviderRepo turns dir into a tagged provider checkout.
func initCLIProviderRepo(t *testing.T, dir string) {
	t.Helper()
	writeCLISkill(t, dir, "provider")
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "provider")
	runGit(t, dir, "tag", "v1")
}

const cliConsumerPayload = `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"review","from":"s","directory":"pkgs/review"}]}`

// TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI proves
// the production resolve serves transitive providers from the configured
// dependency repository root: the provider checkout lives only under the
// temp config's skills root, so the explicit attempt must resolve with
// zero acquisitions and the frozen install must stay green.
func TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI(t *testing.T) {
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	initCLIProviderRepo(t, filepath.Join(root, "skills-root", "provider"))
	writeCLIConsumerSkill(t, filepath.Join(project, "pkgs", "review"), "review")
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(cliConsumerPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	provider, ok := lock.Find("provider")
	if !ok {
		t.Fatalf("lock misses transitive provider: %+v", lock.Members)
	}
	if provider.Selection != nil {
		t.Fatalf("transitive provider carries a root selection index: %+v", provider)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI proves
// the production resolve never consults the process working directory for
// transitive providers: the only provider checkout sits under the process
// CWD while the configured root has none, so the explicit attempt must
// fall through to acquisition and fail instead of finding the CWD copy.
func TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI(t *testing.T) {
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	writeCLIConsumerSkill(t, filepath.Join(project, "pkgs", "review"), "review")
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(cliConsumerPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd := t.TempDir()
	initCLIProviderRepo(t, filepath.Join(cwd, "provider"))
	t.Chdir(cwd)
	// The acquisition fallback must fail closed, never prompt: the URL is
	// unroutable and there is no terminal in tests.
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitFail || !strings.Contains(stderr, "failed to clone") {
		t.Fatalf("CWD-only resolve = %d, stderr %q, want a failed acquisition, never the CWD checkout", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatalf("failed resolve published a lock: %v", err)
	}
}

// assertV5GitMarker binds an installed v5 Git marker to its lock
// generation at the CLI production entry: schema 5, the locked package
// kind and commit, and the binding lock. The replaced legacy identity
// is absent from the decoded marker and the raw bytes alike; the
// declared ref (tag vN) is asserted from the manifest the lock binds,
// which is the v5 shape of the "tag vN at the locked commit" identity.
func assertV5GitMarker(t *testing.T, project, skill string, lock *sourcelock.Lock, wantKind, wantTag string) *marker.Marker {
	t.Helper()
	member, ok := lock.Find(skill)
	if !ok {
		t.Fatalf("lock misses %s", skill)
	}
	installed := filepath.Join(project, ".agents", "skills", skill)
	recorded := marker.Read(installed)
	if recorded == nil {
		t.Fatalf("installed marker unreadable in %s", installed)
	}
	if recorded.SchemaVersion != marker.SchemaV5 {
		t.Fatalf("schema = %d, want %d", recorded.SchemaVersion, marker.SchemaV5)
	}
	if recorded.Package == nil || recorded.Package.Kind != wantKind ||
		recorded.Package.Repository != member.Package.Repository ||
		recorded.Package.Source != member.Package.Source ||
		recorded.Package.Directory != member.Package.Directory ||
		recorded.Package.Commit == nil ||
		recorded.Package.Commit.ObjectFormat != member.Package.Commit.ObjectFormat ||
		recorded.Package.Commit.Hex != member.Package.Commit.Hex ||
		recorded.LockSHA256 != lock.LockSHA256 {
		t.Fatalf("marker binds %+v %s, want the locked %s %+v %s",
			recorded.Package, recorded.LockSHA256, wantKind, member.Package, lock.LockSHA256)
	}
	if recorded.Source != "" || recorded.Git != "" ||
		recorded.RefKind != "" || recorded.Ref != "" || recorded.Commit != "" {
		t.Fatalf("marker identity = %q %q %s %s %s, want empty",
			recorded.Source, recorded.Git, recorded.RefKind, recorded.Ref, recorded.Commit)
	}
	payload, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"source", "git", "ref_kind", "ref", "commit"} {
		if _, present := raw[field]; present {
			t.Fatalf("v5 Git marker carries replaced field %q", field)
		}
	}
	manifestPayload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifestPayload), `"tag":"`+wantTag+`"`) {
		t.Fatalf("manifest does not declare tag %s:\n%s", wantTag, manifestPayload)
	}
	if err := lock.CheckStale(manifestPayload); err != nil {
		t.Fatalf("lock does not bind the declaring manifest: %v", err)
	}
	return recorded
}

// TestProjectResolveGitAliasSelectionThroughCLI proves frozen consumption
// recovers the declaring alias from the lock member's selection index when
// two aliases share one canonical repository: aliases a (tag v1) and s
// (tag v2) point at the same commit, the selection declares `from: s`,
// so the dry-run report and the bound lock must name v2, never the
// first alias.
func TestProjectResolveGitAliasSelectionThroughCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	work := filepath.Join(root, "kit-work")
	bare := filepath.Join(root, "kit.git")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(work, "skills", "review"), "review")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "fixture")
	runGit(t, work, "tag", "v1")
	runGit(t, work, "tag", "v2")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	out, err := exec.Command("git", "--git-dir", bare, "rev-parse", "v2^{commit}").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	commit := strings.TrimSpace(string(out))

	configPath, project := setupCLIProject(t, root)
	const declaredURL = "https://fixture.test/kit.git"
	payload := `{"schema_version":2,"sources":{"a":{"git":"` + declaredURL + `","tag":"v1"},"s":{"git":"` + declaredURL + `","tag":"v2"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	fakeDir := installFakeGitForDraftCLI(t, declaredURL, bare, realGit)
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}

	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatalf("lock misses review: %+v", lock.Members)
	}
	if member.Selection == nil || *member.Selection != 0 {
		t.Fatalf("review selection = %+v, want the declaring index 0", member.Selection)
	}
	if code, stdout, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	} else if !strings.Contains(stdout, "review tag v2 ") {
		t.Fatalf("dry-run report names the wrong ref:\n%s", stdout)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
	}
	assertV5GitMarker(t, project, "review", lock, "network-git", "v2")
	if member.Package.Commit.Hex != commit {
		t.Fatalf("locked commit = %s, want the tagged %s", member.Package.Commit.Hex, commit)
	}
}

// legacyGitCommit resolves one tag in a fixture repository to its commit.
func legacyGitCommit(t *testing.T, repo, tag string) string {
	t.Helper()
	out, err := exec.Command("git", "--git-dir", filepath.Join(repo, ".git"), "rev-parse", tag+"^{commit}").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// initCLILegacyRepo turns dir into a tagged legacy fixture checkout.
func initCLILegacyRepo(t *testing.T, dir, name string) {
	t.Helper()
	writeCLISkill(t, dir, name)
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "legacy")
	runGit(t, dir, "tag", "v1")
}

// TestProjectResolveLegacyConfiguredGitThroughCLI proves unchanged legacy
// configured-git roots (§1, {name, tag} under the configured skills root)
// resolve and install through the production CLI: resolve → dry-run → real
// install → pinned reinstall. It covers the default source location, a
// custom source location, and the declared ref recovery (dry-run names tag
// v1, the marker binds the locked package and lock while the manifest
// declares the ref).
func TestProjectResolveLegacyConfiguredGitThroughCLI(t *testing.T) {
	for _, mode := range []string{"default", "custom-source"} {
		t.Run(mode, func(t *testing.T) {
			withDraftSourcesSwitch(t)
			root := t.TempDir()
			configPath, project := setupCLIProject(t, root)
			skillsRoot := filepath.Join(root, "skills-root")
			source := "review"
			if mode == "custom-source" {
				source = "custom/review"
			}
			repo := filepath.Join(skillsRoot, filepath.FromSlash(source))
			initCLILegacyRepo(t, repo, "review")
			commit := legacyGitCommit(t, repo, "v1")
			var payload string
			if mode == "default" {
				payload = `{"schema_version":2,"skills":[{"name":"review","tag":"v1"}]}`
			} else {
				payload = `{"schema_version":2,"skills":[{"name":"review","source":"custom/review","tag":"v1"}]}`
			}
			if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
				t.Fatal(err)
			}
			if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
				t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
			}
			lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
			if err != nil {
				t.Fatal(err)
			}
			member, ok := lock.Find("review")
			if !ok {
				t.Fatalf("lock misses review: %+v", lock.Members)
			}
			if member.Package.Kind != "configured-git" {
				t.Fatalf("package = %+v, want configured-git", member.Package)
			}
			if member.Selection == nil || *member.Selection != 0 {
				t.Fatalf("review selection = %+v, want the declaring index 0", member.Selection)
			}
			if code, stdout, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
				t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
			} else if !strings.Contains(stdout, "review tag v1 ") {
				t.Fatalf("dry-run report names the wrong ref:\n%s", stdout)
			}
			if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
				t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
			}
			recorded := assertV5GitMarker(t, project, "review", lock, "configured-git", "v1")
			if recorded.Package.Source != source {
				t.Fatalf("marker package source = %q, want %q", recorded.Package.Source, source)
			}
			if member.Package.Commit.Hex != commit {
				t.Fatalf("locked commit = %s, want the tagged %s", member.Package.Commit.Hex, commit)
			}
			if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
				t.Fatalf("pinned reinstall = %d\nstderr:\n%s", code, stderr)
			}
		})
	}
}

// TestProjectResolveLegacyNetworkGitThroughCLI proves unchanged legacy
// network-git roots (§1, {name, tag, git} with no from: selector) resolve
// and install through the production CLI from the SkillsRoot checkout the
// legacy lane snapshotted them from: resolve → dry-run → real install →
// pinned reinstall. It covers the default and custom source locations.
func TestProjectResolveLegacyNetworkGitThroughCLI(t *testing.T) {
	for _, mode := range []string{"default", "custom-source"} {
		t.Run(mode, func(t *testing.T) {
			withDraftSourcesSwitch(t)
			root := t.TempDir()
			configPath, project := setupCLIProject(t, root)
			skillsRoot := filepath.Join(root, "skills-root")
			source := "review"
			if mode == "custom-source" {
				source = "custom/review"
			}
			repo := filepath.Join(skillsRoot, filepath.FromSlash(source))
			const declaredURL = "https://fixture.test/review.git"
			// The checkout's own origin is a local bare repository so the
			// explicit resolve/refresh fetch stays offline; the manifest's
			// declared git URL keeps the network-git identity under test.
			work := filepath.Join(root, "network-work")
			bare := filepath.Join(root, "network.git")
			initCLILegacyRepo(t, work, "review")
			runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
			runGit(t, "", "clone", "--quiet", "--", bare, repo)
			commit := legacyGitCommit(t, repo, "v1")
			var payload string
			if mode == "default" {
				payload = `{"schema_version":2,"skills":[{"name":"review","tag":"v1","git":"` + declaredURL + `"}]}`
			} else {
				payload = `{"schema_version":2,"skills":[{"name":"review","source":"custom/review","tag":"v1","git":"` + declaredURL + `"}]}`
			}
			if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
				t.Fatal(err)
			}
			if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
				t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
			}
			lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
			if err != nil {
				t.Fatal(err)
			}
			member, ok := lock.Find("review")
			if !ok {
				t.Fatalf("lock misses review: %+v", lock.Members)
			}
			if member.Package.Kind != "network-git" {
				t.Fatalf("package = %+v, want network-git", member.Package)
			}
			if member.Selection == nil || *member.Selection != 0 {
				t.Fatalf("review selection = %+v, want the declaring index 0", member.Selection)
			}
			if code, stdout, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
				t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
			} else if !strings.Contains(stdout, "review tag v1 ") {
				t.Fatalf("dry-run report names the wrong ref:\n%s", stdout)
			}
			if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
				t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
			}
			assertV5GitMarker(t, project, "review", lock, "network-git", "v1")
			if member.Package.Commit.Hex != commit {
				t.Fatalf("locked commit = %s, want the tagged %s", member.Package.Commit.Hex, commit)
			}
			if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
				t.Fatalf("pinned reinstall = %d\nstderr:\n%s", code, stderr)
			}
		})
	}
}

// TestProjectResolveLegacyMixedThroughCLI proves a mixed manifest with one
// legacy configured-git root and one draft local selection resolves and
// consumes its lock through the production CLI. The frozen install is
// exercised dry-run (local draft real installs carry no Git ref for the
// schema-2 marker); real-install and pinned-reinstall coverage for the
// legacy member lives in the dedicated legacy tests above.
func TestProjectResolveLegacyMixedThroughCLI(t *testing.T) {
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	skillsRoot := filepath.Join(root, "skills-root")
	initCLILegacyRepo(t, filepath.Join(skillsRoot, "legacy-review"), "legacy-review")
	writeCLISkill(t, filepath.Join(project, "pkgs", "local-review"), "local-review")
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"legacy-review","tag":"v1"},{"name":"local-review","from":"s","directory":"pkgs/local-review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := lock.Find("legacy-review"); !ok {
		t.Fatalf("lock misses legacy-review: %+v", lock.Members)
	}
	if _, ok := lock.Find("local-review"); !ok {
		t.Fatalf("lock misses local-review: %+v", lock.Members)
	}
	if code, stdout, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	} else if !strings.Contains(stdout, "legacy-review tag v1 ") || !strings.Contains(stdout, "local-review") {
		t.Fatalf("dry-run report misses a member:\n%s", stdout)
	}
	// A second dry-run consumes the same pin without reselecting.
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("pinned reinstall = %d\nstderr:\n%s", code, stderr)
	}
}

// requireRealGit skips POSIX-only draft Git CLI tests without git and
// returns the real git binary for the stand-in transport wrapper.
func requireRealGit(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	return realGit
}

// withFixtureGitPATH prepends the stand-in git rewriting one fixture
// https URL to a local bare repository.
func withFixtureGitPATH(t *testing.T, declaredURL, bare, realGit string) {
	t.Helper()
	fakeDir := installFakeGitForDraftCLI(t, declaredURL, bare, realGit)
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}
}

// writeTagFixture commits packages and tags the result v1 in work.
func writeTagFixture(t *testing.T, work string, packages map[string]string) {
	t.Helper()
	dirs := make([]string, 0, len(packages))
	for dir := range packages {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		writeCLISkill(t, filepath.Join(work, filepath.FromSlash(dir)), packages[dir])
	}
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "tagged")
	runGit(t, work, "tag", "v1")
}

// cloneBareResolveCommit clones work to a bare repository, bootstraps the
// CLI project, installs the fixture transport, and returns the config
// path, project root, bare repository, and the tagged commit.
func cloneBareResolveCommit(t *testing.T, root, work, declaredURL, realGit string) (configPath, project, bare, commit string) {
	t.Helper()
	bare = filepath.Join(root, "kit.git")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	out, err := exec.Command("git", "--git-dir", bare, "rev-parse", "v1^{commit}").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	commit = strings.TrimSpace(string(out))
	configPath, project = setupCLIProject(t, root)
	withFixtureGitPATH(t, declaredURL, bare, realGit)
	return configPath, project, bare, commit
}

// TestProjectResolveGitTagDiffersFromHEADThroughCLI proves an individual
// Git selection resolves from its declared tag when default-branch HEAD
// no longer carries the package: tag v1 holds skills/review while HEAD
// removed its SKILL.md. A checkout-based expansion would refuse the
// attempt; the pinned resolve must lock tag v1 and install it for real.
func TestProjectResolveGitTagDiffersFromHEADThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	work := filepath.Join(root, "kit-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTagFixture(t, work, map[string]string{"skills/review": "review"})
	if err := os.Remove(filepath.Join(work, "skills", "review", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", "-A")
	runGit(t, work, "commit", "-qm", "head-removes-skill")
	const declaredURL = "https://fixture.test/kit.git"
	configPath, project, _, commit := cloneBareResolveCommit(t, root, work, declaredURL, realGit)
	payload := `{"schema_version":2,"sources":{"s":{"git":"` + declaredURL + `","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatalf("lock misses review: %+v", lock.Members)
	}
	if member.Package.Commit.Hex != commit {
		t.Fatalf("review commit = %s, want the tagged %s", member.Package.Commit.Hex, commit)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
	}
	assertV5GitMarker(t, project, "review", lock, "network-git", "v1")
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("pinned reinstall = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProjectResolveGitCollectionPinnedToTagThroughCLI proves collection
// membership comes from the declared tag's tree: tag v1 holds alpha and
// beta while HEAD removed beta. A checkout-based expansion would silently
// lock alpha alone; the pinned resolve must lock both and consume them.
func TestProjectResolveGitCollectionPinnedToTagThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	work := filepath.Join(root, "kit-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTagFixture(t, work, map[string]string{"skills/alpha": "alpha", "skills/beta": "beta"})
	if err := os.RemoveAll(filepath.Join(work, "skills", "beta")); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", "-A")
	runGit(t, work, "commit", "-qm", "head-removes-beta")
	const declaredURL = "https://fixture.test/kit.git"
	configPath, project, _, commit := cloneBareResolveCommit(t, root, work, declaredURL, realGit)
	payload := `{"schema_version":2,"sources":{"s":{"git":"` + declaredURL + `","tag":"v1"}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Members) != 2 {
		t.Fatalf("members = %+v, want [alpha beta]", lock.Members)
	}
	for _, name := range []string{"alpha", "beta"} {
		member, ok := lock.Find(name)
		if !ok {
			t.Fatalf("lock misses %s: %+v", name, lock.Members)
		}
		if member.Package.Commit.Hex != commit {
			t.Fatalf("%s commit = %s, want the tagged %s", name, member.Package.Commit.Hex, commit)
		}
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("pinned reinstall = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProjectRefreshGitBranchMembershipChangeThroughCLI proves a branch
// refresh observes the newly fetched commit's tree while the acquired
// checkout stays behind: the collection resolves alpha, the branch then
// gains beta, and refresh must lock both even though the checkout HEAD
// never moved. Deleting the bare repository afterwards must not fail the
// pinned reinstall.
func TestProjectRefreshGitBranchMembershipChangeThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	work := filepath.Join(root, "kit-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(work, "skills", "alpha"), "alpha")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "alpha-only")
	bare := filepath.Join(root, "kit.git")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	oldOut, err := exec.Command("git", "--git-dir", bare, "rev-parse", "main").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	oldCommit := strings.TrimSpace(string(oldOut))
	const declaredURL = "https://fixture.test/kit.git"
	configPath, project := setupCLIProject(t, root)
	withFixtureGitPATH(t, declaredURL, bare, realGit)
	payload := `{"schema_version":2,"sources":{"s":{"git":"` + declaredURL + `","branch":"main"}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Members) != 1 {
		t.Fatalf("members = %+v, want [alpha]", lock.Members)
	}
	// The branch advances upstream while the acquired checkout stays at
	// the old commit: a fetch moves origin/main, never the working tree.
	writeCLISkill(t, filepath.Join(work, "skills", "beta"), "beta")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "add-beta")
	runGit(t, work, "push", bare, "main")
	home := filepath.Dir(configPath)
	checkoutOut, err := exec.Command("git", "--git-dir", filepath.Join(draftRepoDir(home, project, "s"), ".git"), "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("checkout rev-parse: %v", err)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	refreshed, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(refreshed.Members) != 2 {
		t.Fatalf("refreshed members = %+v, want [alpha beta]", refreshed.Members)
	}
	if refreshed.LockSHA256 == lock.LockSHA256 {
		t.Fatalf("refresh left the lock unchanged after the branch gained a member")
	}
	staleHEAD := strings.TrimSpace(string(checkoutOut))
	if staleHEAD != oldCommit {
		t.Fatalf("checkout HEAD = %s, want the unchanged %s", staleHEAD, oldCommit)
	}
	for _, name := range []string{"alpha", "beta"} {
		member, ok := refreshed.Find(name)
		if !ok {
			t.Fatalf("refreshed lock misses %s: %+v", name, refreshed.Members)
		}
		if member.Package.Commit.Hex == oldCommit {
			t.Fatalf("%s commit = %s, want the refreshed branch commit", name, oldCommit)
		}
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	if err := os.RemoveAll(bare); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("pinned reinstall without the repository = %d\nstderr:\n%s", code, stderr)
	}
}

// setCLIAllowedSources rewrites the temp config's allowed_sources without
// disturbing other fields, so the production CLI loads the restrictive
// allowlist through its normal config path.
func setCLIAllowedSources(t *testing.T, configPath string, allowed []string) {
	t.Helper()
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	items := make([]any, 0, len(allowed))
	for _, item := range allowed {
		items = append(items, item)
	}
	obj["allowed_sources"] = items
	payload, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, append(payload, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

// installLoggingGitForDraftCLI prepends a POSIX git wrapper that records
// every invocation's arguments to logPath before delegating to the real
// git. A denied acquisition that reaches the tool leaves the declared URL
// in the log; a gate before acquisition leaves no trace of it.
func installLoggingGitForDraftCLI(t *testing.T, realGit, logPath string) {
	t.Helper()
	if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	fakeDir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + logPath + "'\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}
}

// writeCLIConsumerSkillWithGit writes one local package whose skill
// requirement names a transitive provider at the given git URL. The URL
// is unroutable, so any clone or fetch attempt fails; only a SkillsRoot
// checkout or an allowlist refusal avoids the network.
func writeCLIConsumerSkillWithGit(t *testing.T, dir, name, gitURL string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := `{"schema_version":4,"capabilities":{},"commands":{},"dependencies":{"skills":{"provider":{"git":"` + gitURL + `","ref":{"kind":"tag","value":"v1"}}}}}`
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestProjectResolveDraftAllowlistLegacyThroughCLI proves the production
// resolve gates legacy named roots with the machine allowlist before any
// acquisition: a denied root fails with source not allowed, leaves
// lock/bindings/installed state unchanged, and never reaches git, while
// the allowed control resolves from the SkillsRoot checkout.
func TestProjectResolveDraftAllowlistLegacyThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	withDraftSourcesSwitch(t)
	const deniedURL = "https://denied.test/review.git"
	t.Run("denied", func(t *testing.T) {
		root := t.TempDir()
		configPath, project := setupCLIProject(t, root)
		setCLIAllowedSources(t, configPath, []string{"allowed.test"})
		home := filepath.Dir(configPath)
		logPath := filepath.Join(root, "git.log")
		installLoggingGitForDraftCLI(t, realGit, logPath)
		t.Setenv("GIT_TERMINAL_PROMPT", "0")
		payload := `{"schema_version":2,"skills":[{"name":"review","tag":"v1","git":"` + deniedURL + `"}]}`
		if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
			t.Fatal(err)
		}
		lockPath := filepath.Join(project, "Skillfile.lock.json")
		bindingsPath := install.DraftBindingsPath(home, project)
		cloneDest := filepath.Join(root, "skills-root", "review")
		if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitFail || !strings.Contains(stderr, "source not allowed") {
			t.Fatalf("denied legacy resolve = %d, stderr %q, want source not allowed", code, stderr)
		}
		if _, err := os.Lstat(lockPath); !os.IsNotExist(err) {
			t.Fatalf("denied resolve published a lock: %v", err)
		}
		if _, err := os.Lstat(bindingsPath); !os.IsNotExist(err) {
			t.Fatalf("denied resolve published bindings: %v", err)
		}
		if _, err := os.Lstat(cloneDest); !os.IsNotExist(err) {
			t.Fatalf("denied resolve acquired %s", cloneDest)
		}
		if _, err := os.Lstat(filepath.Join(project, ".agents", "skills", "review")); !os.IsNotExist(err) {
			t.Fatalf("denied resolve installed state changed")
		}
		log, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(log), "denied.test") {
			t.Fatalf("denied resolve reached git before the gate:\n%s", log)
		}
	})
	t.Run("allowed", func(t *testing.T) {
		root := t.TempDir()
		configPath, project := setupCLIProject(t, root)
		setCLIAllowedSources(t, configPath, []string{"denied.test"})
		skillsRoot := filepath.Join(root, "skills-root")
		initCLILegacyRepo(t, filepath.Join(skillsRoot, "review"), "review")
		payload := `{"schema_version":2,"skills":[{"name":"review","tag":"v1","git":"` + deniedURL + `"}]}`
		if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
			t.Fatal(err)
		}
		if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
			t.Fatalf("allowed legacy resolve = %d\nstderr:\n%s", code, stderr)
		}
		lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := lock.Find("review"); !ok {
			t.Fatalf("allowed lock misses review: %+v", lock.Members)
		}
		if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
			t.Fatalf("allowed install --dry-run = %d\nstderr:\n%s", code, stderr)
		}
	})
}

// TestProjectResolveDraftAllowlistTransitiveThroughCLI proves the
// production resolve gates transitive dependencies with the machine
// allowlist before any acquisition: a denied provider fails with source
// not allowed, leaves lock/bindings/installed state unchanged, and never
// reaches git, while the allowed control resolves from the SkillsRoot
// checkout with zero acquisitions.
func TestProjectResolveDraftAllowlistTransitiveThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	withDraftSourcesSwitch(t)
	const providerURL = "https://denied.test/provider.git"
	t.Run("denied", func(t *testing.T) {
		root := t.TempDir()
		configPath, project := setupCLIProject(t, root)
		setCLIAllowedSources(t, configPath, []string{"allowed.test"})
		home := filepath.Dir(configPath)
		logPath := filepath.Join(root, "git.log")
		installLoggingGitForDraftCLI(t, realGit, logPath)
		t.Setenv("GIT_TERMINAL_PROMPT", "0")
		writeCLIConsumerSkillWithGit(t, filepath.Join(project, "pkgs", "review"), "review", providerURL)
		if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(cliConsumerPayload), 0o644); err != nil {
			t.Fatal(err)
		}
		lockPath := filepath.Join(project, "Skillfile.lock.json")
		bindingsPath := install.DraftBindingsPath(home, project)
		cloneDest := filepath.Join(root, "skills-root", "provider")
		if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitFail || !strings.Contains(stderr, "source not allowed") {
			t.Fatalf("denied transitive resolve = %d, stderr %q, want source not allowed", code, stderr)
		}
		if _, err := os.Lstat(lockPath); !os.IsNotExist(err) {
			t.Fatalf("denied resolve published a lock: %v", err)
		}
		if _, err := os.Lstat(bindingsPath); !os.IsNotExist(err) {
			t.Fatalf("denied resolve published bindings: %v", err)
		}
		if _, err := os.Lstat(cloneDest); !os.IsNotExist(err) {
			t.Fatalf("denied resolve acquired %s", cloneDest)
		}
		if _, err := os.Lstat(filepath.Join(project, ".agents", "skills", "review")); !os.IsNotExist(err) {
			t.Fatalf("denied resolve installed state changed")
		}
		log, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(log), "denied.test") {
			t.Fatalf("denied resolve reached git before the gate:\n%s", log)
		}
	})
	t.Run("allowed", func(t *testing.T) {
		root := t.TempDir()
		configPath, project := setupCLIProject(t, root)
		setCLIAllowedSources(t, configPath, []string{"denied.test"})
		initCLIProviderRepo(t, filepath.Join(root, "skills-root", "provider"))
		// Rewrite the consumer fixture to require the allowlisted
		// denied.test provider instead of the default invalid host.
		writeCLIConsumerSkillWithGit(t, filepath.Join(project, "pkgs", "review"), "review", providerURL)
		if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(cliConsumerPayload), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("GIT_TERMINAL_PROMPT", "0")
		if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
			t.Fatalf("allowed transitive resolve = %d\nstderr:\n%s", code, stderr)
		}
		lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := lock.Find("provider"); !ok {
			t.Fatalf("allowed lock misses transitive provider: %+v", lock.Members)
		}
		if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
			t.Fatalf("allowed install --dry-run = %d\nstderr:\n%s", code, stderr)
		}
	})
}

// TestProjectResolveGitRealInstallThroughCLI proves a successfully resolved
// Git selection installs for real through the CLI with the accepted v5
// package and lock binding in its marker, and that a second install
// consumes the pin without the network: deleting the bare repository
// must not fail it.
func TestProjectResolveGitRealInstallThroughCLI(t *testing.T) {
	configPath, project, commit, _, bare := setupScriptGitCLI(t)
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
		t.Fatalf("installed skill missing: %v", err)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	recorded := assertV5GitMarker(t, project, "review", lock, "network-git", "v1")
	if recorded.Package.Commit.Hex != commit {
		t.Fatalf("marker commit = %s, want the tagged %s", recorded.Package.Commit.Hex, commit)
	}
	if err := os.RemoveAll(bare); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("pinned reinstall without the repository = %d\nstderr:\n%s", code, stderr)
	}
}

// bareRev parses one ref in a bare fixture repository.
func bareRev(t *testing.T, bare, ref string) string {
	t.Helper()
	out, err := exec.Command("git", "--git-dir", bare, "rev-parse", ref).Output()
	if err != nil {
		t.Fatalf("rev-parse %s: %v", ref, err)
	}
	return strings.TrimSpace(string(out))
}

// checkoutHEAD parses HEAD of a non-bare checkout.
func checkoutHEAD(t *testing.T, checkout string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", checkout, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD in %s: %v", checkout, err)
	}
	return strings.TrimSpace(string(out))
}

// requireCheckoutStaleFor fails when the checkout already resolves ref,
// proving its working tree predates the refreshed upstream ref.
func requireCheckoutStaleFor(t *testing.T, checkout, ref string) {
	t.Helper()
	if err := exec.Command("git", "-C", checkout, "rev-parse", "--verify", ref+"^{commit}").Run(); err == nil {
		t.Fatalf("checkout already carries %s, want a stale tree", ref)
	}
}

// setupLegacyBranchRefresh builds an upstream script package with a local
// bare repository, clones it into the configured dependency root, and
// resolves the given legacy payload through the production CLI. The
// checkout keeps a real local origin while its working tree stays at the
// resolved commit, so the caller can advance the bare repository and prove
// the explicit refresh fetches it. It returns the config path, project,
// home, upstream work dir, bare repository, checkout, and locked commit.
func setupLegacyBranchRefresh(t *testing.T, payload string) (configPath, project, home, work, bare, checkout, oldCommit string) {
	t.Helper()
	return setupLegacyBranchRefreshWithSkill(t, payload, writeCLIScriptSkill)
}

// setupLegacyBranchRefreshWithSkill builds the same upstream bare
// repository, configured checkout and production-CLI lock as
// setupLegacyBranchRefresh, but writes the upstream package with writeSkill.
// The fetch-failure test uses a plain skill (the same portable shape as the
// TestProjectResolveLegacy* fixtures that real-install on Windows): the
// script form declares a unix-only command path that the real install
// refuses on Windows, while this test needs the pre-refresh real install
// to succeed there.
func setupLegacyBranchRefreshWithSkill(t *testing.T, payload string, writeSkill func(t *testing.T, dir, name string)) (configPath, project, home, work, bare, checkout, oldCommit string) {
	t.Helper()
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	work = filepath.Join(root, "upstream-work")
	writeSkill(t, work, "review")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "initial")
	bare = filepath.Join(root, "upstream.git")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	oldCommit = bareRev(t, bare, "main")
	configPath, project = setupCLIProject(t, root)
	home = filepath.Dir(configPath)
	checkout = filepath.Join(root, "skills-root", "review")
	runGit(t, "", "clone", "--quiet", "--", bare, checkout)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatalf("lock misses review: %+v", lock.Members)
	}
	if member.Package.Commit.Hex != oldCommit {
		t.Fatalf("locked commit = %s, want %s", member.Package.Commit.Hex, oldCommit)
	}
	// Pinned consumption before any refresh: the frozen install stays
	// green without touching the network.
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("pinned install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	return configPath, project, home, work, bare, checkout, oldCommit
}

// advanceUpstreamScript commits a runtime-only change upstream and pushes
// it to the bare repository, leaving every existing checkout stale. It
// returns the new upstream commit.
func advanceUpstreamScript(t *testing.T, work, bare, body string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(work, "scripts", "tool.sh"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "runtime-only change")
	runGit(t, work, "push", bare, "main")
	return bareRev(t, bare, "main")
}

// TestProjectRefreshLegacyConfiguredGitBranchThroughCLI proves the explicit
// refresh fetches a configured-git branch root ({name,branch} under the
// configured skills root) whose checkout stayed stale: a runtime-only
// upstream advance must publish a new lock with complete runtime bytes
// while the working tree HEAD never moves.
func TestProjectRefreshLegacyConfiguredGitBranchThroughCLI(t *testing.T) {
	configPath, project, home, work, bare, checkout, oldCommit := setupLegacyBranchRefresh(t,
		`{"schema_version":2,"skills":[{"name":"review","branch":"main"}]}`)
	_ = bare
	lockBefore, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	newBody := "#!/bin/sh\necho refreshed\n"
	newCommit := advanceUpstreamScript(t, work, bare, newBody)
	if newCommit == oldCommit {
		t.Fatalf("upstream advance kept commit %s", oldCommit)
	}
	if got := checkoutHEAD(t, checkout); got != oldCommit {
		t.Fatalf("checkout HEAD = %s, want the stale %s", got, oldCommit)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	refreshed, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(refreshed.Members) != 1 {
		t.Fatalf("refreshed members = %+v, want exactly [review]", refreshed.Members)
	}
	member, ok := refreshed.Find("review")
	if !ok {
		t.Fatalf("refreshed lock misses review: %+v", refreshed.Members)
	}
	if member.Package.Kind != "configured-git" {
		t.Fatalf("package = %+v, want configured-git", member.Package)
	}
	if member.Package.Commit.Hex != newCommit {
		t.Fatalf("refreshed commit = %s, want %s", member.Package.Commit.Hex, newCommit)
	}
	if refreshed.LockSHA256 == lockBefore.LockSHA256 {
		t.Fatalf("refresh left the lock unchanged after the runtime-only branch advance")
	}
	// The fetch moves origin/main only: the working tree stays stale while
	// the published lock and stored bytes carry the new commit.
	if got := checkoutHEAD(t, checkout); got != oldCommit {
		t.Fatalf("checkout HEAD = %s, want the unchanged %s", got, oldCommit)
	}
	cached := filepath.Join(home, "cache", "review", newCommit, "snapshot", "scripts", "tool.sh")
	content, err := os.ReadFile(cached)
	if err != nil {
		t.Fatalf("refreshed runtime bytes unreadable: %v", err)
	}
	if string(content) != newBody {
		t.Fatalf("refreshed runtime bytes = %q, want %q", content, newBody)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("post-refresh install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProjectRefreshLegacyNetworkGitBranchThroughCLI proves the explicit
// refresh fetches a network-git branch root ({name,branch,git}) whose
// checkout stayed stale: the declared fixture identity is never contacted
// (the checkout's own local origin serves the fetch) and the refreshed
// lock keeps the network-git kind with the new commit and runtime bytes.
func TestProjectRefreshLegacyNetworkGitBranchThroughCLI(t *testing.T) {
	const declaredURL = "https://fixture.test/review.git"
	configPath, project, home, work, bare, checkout, oldCommit := setupLegacyBranchRefresh(t,
		`{"schema_version":2,"skills":[{"name":"review","branch":"main","git":"`+declaredURL+`"}]}`)
	_ = bare
	lockBefore, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if member, ok := lockBefore.Find("review"); !ok || member.Package.Kind != "network-git" {
		t.Fatalf("members = %+v, want one network-git review", lockBefore.Members)
	}
	newBody := "#!/bin/sh\necho refreshed\n"
	newCommit := advanceUpstreamScript(t, work, bare, newBody)
	if newCommit == oldCommit {
		t.Fatalf("upstream advance kept commit %s", oldCommit)
	}
	if got := checkoutHEAD(t, checkout); got != oldCommit {
		t.Fatalf("checkout HEAD = %s, want the stale %s", got, oldCommit)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	refreshed, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := refreshed.Find("review")
	if !ok {
		t.Fatalf("refreshed lock misses review: %+v", refreshed.Members)
	}
	if member.Package.Kind != "network-git" {
		t.Fatalf("package = %+v, want network-git", member.Package)
	}
	if member.Package.Commit.Hex != newCommit {
		t.Fatalf("refreshed commit = %s, want %s", member.Package.Commit.Hex, newCommit)
	}
	if refreshed.LockSHA256 == lockBefore.LockSHA256 {
		t.Fatalf("refresh left the lock unchanged after the runtime-only branch advance")
	}
	if got := checkoutHEAD(t, checkout); got != oldCommit {
		t.Fatalf("checkout HEAD = %s, want the unchanged %s", got, oldCommit)
	}
	cached := filepath.Join(home, "cache", "review", newCommit, "snapshot", "scripts", "tool.sh")
	content, err := os.ReadFile(cached)
	if err != nil {
		t.Fatalf("refreshed runtime bytes unreadable: %v", err)
	}
	if string(content) != newBody {
		t.Fatalf("refreshed runtime bytes = %q, want %q", content, newBody)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("post-refresh install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProjectRefreshTransitiveTagAdvanceThroughCLI proves the explicit
// refresh fetches a transitive provider checkout for a newly required ref:
// the consumer moves from tag v1 to tag v2 while the provider checkout
// stays stale, so only a fetch of the transitive repository can resolve
// the refresh.
func TestProjectRefreshTransitiveTagAdvanceThroughCLI(t *testing.T) {
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	const providerURL = "https://fixture.test/provider.git"
	providerWork := filepath.Join(root, "provider-work")
	writeCLISkill(t, providerWork, "provider")
	runGit(t, providerWork, "init", "-q", "-b", "main")
	runGit(t, providerWork, "add", ".")
	runGit(t, providerWork, "commit", "-qm", "provider v1")
	runGit(t, providerWork, "tag", "v1")
	bare := filepath.Join(root, "provider.git")
	runGit(t, "", "clone", "--quiet", "--bare", "--", providerWork, bare)
	v1 := bareRev(t, bare, "v1")
	configPath, project := setupCLIProject(t, root)
	checkout := filepath.Join(root, "skills-root", "provider")
	runGit(t, "", "clone", "--quiet", "--", bare, checkout)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	consumerDir := filepath.Join(project, "pkgs", "review")
	writeCLIConsumerSkillWithGit(t, consumerDir, "review", providerURL)
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(cliConsumerPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	provider, ok := lock.Find("provider")
	if !ok {
		t.Fatalf("lock misses transitive provider: %+v", lock.Members)
	}
	if provider.Package.Commit.Hex != v1 {
		t.Fatalf("locked provider = %s, want %s", provider.Package.Commit.Hex, v1)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("pinned install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
	// Tag v2 exists only upstream: the checkout is stale for it, and the
	// consumer now requires it.
	if err := os.WriteFile(filepath.Join(providerWork, "references", "info.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, providerWork, "add", ".")
	runGit(t, providerWork, "commit", "-qm", "provider v2")
	runGit(t, providerWork, "tag", "v2")
	runGit(t, providerWork, "push", bare, "main")
	runGit(t, providerWork, "push", bare, "v2")
	v2 := bareRev(t, bare, "v2")
	requireCheckoutStaleFor(t, checkout, "refs/tags/v2")
	specV2 := `{"schema_version":4,"capabilities":{},"commands":{},"dependencies":{"skills":{"provider":{"git":"` + providerURL + `","ref":{"kind":"tag","value":"v2"}}}}}`
	if err := os.WriteFile(filepath.Join(consumerDir, "agent-skill.json"), []byte(specV2), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	refreshed, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(refreshed.Members) != 2 {
		t.Fatalf("refreshed members = %+v, want [provider review]", refreshed.Members)
	}
	provider, ok = refreshed.Find("provider")
	if !ok {
		t.Fatalf("refreshed lock misses transitive provider: %+v", refreshed.Members)
	}
	if provider.Package.Commit.Hex != v2 {
		t.Fatalf("refreshed provider = %s, want %s", provider.Package.Commit.Hex, v2)
	}
	if _, ok := refreshed.Find("review"); !ok {
		t.Fatalf("refreshed lock misses review: %+v", refreshed.Members)
	}
	if refreshed.LockSHA256 == lock.LockSHA256 {
		t.Fatalf("refresh left the lock unchanged after the transitive tag advance")
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("post-refresh install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProjectRefreshFetchFailurePreservesStateThroughCLI proves a failed
// acquisition during the explicit refresh leaves the prior lock, bindings
// and installed state unchanged: the bare repository is removed so the
// fetch of the branch checkout fails and the refresh must refuse.
func TestProjectRefreshFetchFailurePreservesStateThroughCLI(t *testing.T) {
	configPath, project, home, _, bare, _, _ := setupLegacyBranchRefreshWithSkill(t,
		`{"schema_version":2,"skills":[{"name":"review","branch":"main"}]}`, writeCLISkill)
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
	}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsPath := install.DraftBindingsPath(home, project)
	installedSkill := filepath.Join(project, ".agents", "skills", "review", "SKILL.md")
	lockBefore, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	bindingsBefore, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	skillBefore, err := os.ReadFile(installedSkill)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(bare); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitFail || !strings.Contains(stderr, "failed to fetch") {
		t.Fatalf("broken refresh = %d, stderr %q, want a failed fetch", code, stderr)
	}
	lockAfter, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfter) != string(lockBefore) {
		t.Fatalf("failed refresh rewrote the lock:\nbefore %s\nafter %s", lockBefore, lockAfter)
	}
	bindingsAfter, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(bindingsAfter) != string(bindingsBefore) {
		t.Fatalf("failed refresh rewrote the bindings")
	}
	skillAfter, err := os.ReadFile(installedSkill)
	if err != nil {
		t.Fatalf("failed refresh disturbed the installed skill: %v", err)
	}
	if string(skillAfter) != string(skillBefore) {
		t.Fatalf("failed refresh changed the installed skill")
	}
	// The preserved prior state stays consumable without the network.
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("post-failure install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
}

// installFailingRemoteGit prepends a POSIX git wrapper that fails only
// the bare `git remote` enumeration (exit 73) and delegates every other
// invocation to the real git. It models a failed remote enumeration, not
// an origin-less repository: Fetch must propagate the failure instead of
// treating it as a successful empty enumeration.
func installFailingRemoteGit(t *testing.T, realGit string) {
	t.Helper()
	fakeDir := t.TempDir()
	script := "#!/bin/sh\nif [ \"$1\" = \"remote\" ] && [ \"$#\" = \"1\" ]; then\n" +
		"echo \"fake remote enumeration failure\" >&2\nexit 73\nfi\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}
}

// TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI proves
// a failed `git remote` enumeration fails the explicit refresh instead of
// succeeding with a stale lock: the wrapper fails only the enumeration
// (exit 73) while fetch itself would succeed, so a Fetch that collapses
// the error into "no remote" would exit 0 and keep the stale lock. The
// refresh must exit nonzero and leave lock, bindings and installed state
// unchanged.
func TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	configPath, project, home, _, _, _, _ := setupLegacyBranchRefreshWithSkill(t,
		`{"schema_version":2,"skills":[{"name":"review","branch":"main"}]}`, writeCLISkill)
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("real install = %d\nstderr:\n%s", code, stderr)
	}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsPath := install.DraftBindingsPath(home, project)
	installedSkill := filepath.Join(project, ".agents", "skills", "review", "SKILL.md")
	lockBefore, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	bindingsBefore, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	skillBefore, err := os.ReadFile(installedSkill)
	if err != nil {
		t.Fatal(err)
	}
	installFailingRemoteGit(t, realGit)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitFail || !strings.Contains(stderr, "failed to fetch") {
		t.Fatalf("broken remote enumeration refresh = %d, stderr %q, want a failed fetch", code, stderr)
	}
	lockAfter, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfter) != string(lockBefore) {
		t.Fatalf("failed refresh rewrote the lock:\nbefore %s\nafter %s", lockBefore, lockAfter)
	}
	bindingsAfter, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(bindingsAfter) != string(bindingsBefore) {
		t.Fatalf("failed refresh rewrote the bindings")
	}
	skillAfter, err := os.ReadFile(installedSkill)
	if err != nil {
		t.Fatalf("failed refresh disturbed the installed skill: %v", err)
	}
	if string(skillAfter) != string(skillBefore) {
		t.Fatalf("failed refresh changed the installed skill")
	}
	// The preserved prior state stays consumable without the network.
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("post-failure install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
}

// TestProjectResolveOriginLessSkipsFetchThroughCLI proves the explicit
// resolve never fetches an origin-less configured checkout: the resolve
// succeeds from the local tag while the git call log carries no fetch.
func TestProjectResolveOriginLessSkipsFetchThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	initCLILegacyRepo(t, filepath.Join(root, "skills-root", "review"), "review")
	payload := `{"schema_version":2,"skills":[{"name":"review","tag":"v1"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(root, "git.log")
	installLoggingGitForDraftCLI(t, realGit, logPath)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(log), "\n") {
		if strings.Contains(line, "fetch") {
			t.Fatalf("origin-less resolve fetched:\n%s", log)
		}
	}
}

// installLoggingRewritingGit prepends a POSIX git wrapper that records
// every invocation's arguments to logPath, rewrites exactly one fixture
// https URL to a local bare repository, and delegates everything else to
// the real git.
func installLoggingRewritingGit(t *testing.T, declaredURL, bare, realGit, logPath string) {
	t.Helper()
	if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	fakeDir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + logPath + "'\nargs=\"\"; for arg in \"$@\"; do\n" +
		"if [ \"$arg\" = '" + declaredURL + "' ]; then arg='file://" + bare + "'; fi\n" +
		"args=\"$args\n$arg\"; done\noldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}
}

// countFetchLogLines counts logged git invocations that fetch.
func countFetchLogLines(t *testing.T, logPath string) int {
	t.Helper()
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(log), "\n") {
		if strings.Contains(line, "fetch") {
			count++
		}
	}
	return count
}

// TestProjectRefreshAliasFetchDedupedThroughCLI proves the explicit refresh
// fetches an alias tree exactly once: the production alias acquisition
// fetches it, and the pinned closure never fetches it twice. A fresh
// resolve clones, so it must fetch zero times; the refresh fetches once.
func TestProjectRefreshAliasFetchDedupedThroughCLI(t *testing.T) {
	realGit := requireRealGit(t)
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	work := filepath.Join(root, "kit-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(work, "skills", "review"), "review")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "fixture")
	bare := filepath.Join(root, "kit.git")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	configPath, project := setupCLIProject(t, root)
	const declaredURL = "https://fixture.test/kit.git"
	payload := `{"schema_version":2,"sources":{"s":{"git":"` + declaredURL + `","branch":"main"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(root, "git.log")
	installLoggingRewritingGit(t, declaredURL, bare, realGit, logPath)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	if got := countFetchLogLines(t, logPath); got != 0 {
		log, _ := os.ReadFile(logPath)
		t.Fatalf("fresh resolve fetched %d times, want 0:\n%s", got, log)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	if got := countFetchLogLines(t, logPath); got != 1 {
		log, _ := os.ReadFile(logPath)
		t.Fatalf("refresh fetched %d times, want exactly 1:\n%s", got, log)
	}
	refreshed, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.LockSHA256 != lock.LockSHA256 {
		t.Fatalf("refresh without upstream change moved the lock")
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
		t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
	}
}
