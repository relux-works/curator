package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
)

type globalLockCLIFixture struct {
	source  stubConfigSource
	home    string
	gitURL  string
	codex   string
	managed string
}

func newGlobalLockCLIFixture(t *testing.T) globalLockCLIFixture {
	t.Helper()
	source, home := profileHome(t)
	source.cfg.SkillsRoot = filepath.Join(home, "skills-root")
	userHome := t.TempDir()
	t.Setenv("HOME", userHome)
	t.Setenv("USERPROFILE", userHome)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(userHome, ".claude"))
	codex := filepath.Join(userHome, ".codex")
	t.Setenv("CODEX_HOME", codex)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(userHome, ".config"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(userHome, ".pi"))
	writeNativeCredentials(t)

	repo := t.TempDir()
	writeCLISkill(t, repo, "new-skill")
	runGit(t, repo, "init", "-q", "-b", "main")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-qm", "publish new-skill")
	runGit(t, repo, "tag", "v1")
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	rewrite := "[url \"" + gitFileURL(repo) + "\"]\n\tinsteadOf = https://example.com/new-skill\n"
	if err := os.WriteFile(gitConfig, []byte(rewrite), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	fixture := globalLockCLIFixture{
		source: source, home: home, gitURL: "https://example.com/new-skill", codex: codex,
		managed: filepath.Join(envprofile.ManagedHomeDir(home, "default", "codex_cli"), "skills", "new-skill"),
	}
	requireGlobalCommand(t, fixture, exitOK, "global", "init")
	return fixture
}

func (fixture globalLockCLIFixture) run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	return runProfile(t, fixture.source, args...)
}

func (fixture globalLockCLIFixture) writeLegacyGlobalDeclaration(t *testing.T) {
	t.Helper()
	writeFile(t, filepath.Join(fixture.home, "global", "Skillfile.json"),
		`{"schema_version":1,"skills":[{"name":"new-skill","git":"`+fixture.gitURL+`","tag":"v1"}]}`)
}

func (fixture globalLockCLIFixture) lock(t *testing.T) (*contextlock.Lock, string) {
	t.Helper()
	lock, hash, err := contextlock.Read(filepath.Join(fixture.home, "profiles", "default", "lock.json"))
	if err != nil {
		t.Fatalf("read default profile lock: %v", err)
	}
	return lock, hash
}

func (fixture globalLockCLIFixture) requireSkillLockAndStore(t *testing.T) (*contextlock.Lock, string) {
	t.Helper()
	lock, hash := fixture.lock(t)
	for _, member := range lock.Members {
		if member.Kind == contextlock.KindSkill && member.Name == "new-skill" {
			if member.Commit == "" || member.Source != "example.com/new-skill" {
				t.Fatalf("new-skill lock member = %+v, want exact Git pin", member)
			}
			entry := contextstore.EntryDir(fixture.home, contextlock.KindSkill, member.Name, member.Commit)
			if info, err := os.Stat(entry); err != nil || !info.IsDir() {
				t.Fatalf("referenced store entry %s is unavailable: %v", entry, err)
			}
			return lock, hash
		}
	}
	t.Fatalf("default lock has no new-skill member: %+v", lock.Members)
	return nil, ""
}

func (fixture globalLockCLIFixture) requireNativeSkill(t *testing.T) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(fixture.codex, ".agent-context", "skills", "new-skill", "SKILL.md"))
	if err != nil {
		t.Fatalf("read materialized native skill: %v", err)
	}
	if !strings.Contains(string(payload), "# new-skill") {
		t.Fatalf("native skill bytes = %q", payload)
	}
	marker, err := envmarker.Read(fixture.codex)
	if err != nil {
		t.Fatalf("read native environment marker: %v", err)
	}
	if marker == nil || marker.Profile.Name != "default" || marker.Surfaces[envmarker.SurfaceSkills].Paths == nil {
		t.Fatalf("native marker does not record the profile skill surface: %+v", marker)
	}
	_, lockHash := fixture.lock(t)
	if marker.Profile.LockSHA256 != strings.TrimPrefix(lockHash, "sha256:") {
		t.Fatalf("native marker lock = %s, profile lock = %s", marker.Profile.LockSHA256, lockHash)
	}
}

func requireGlobalCommand(t *testing.T, fixture globalLockCLIFixture, want int, args ...string) (string, string) {
	t.Helper()
	code, stdout, stderr := fixture.run(t, args...)
	if code != want {
		t.Fatalf("curator %v = %d, want %d\nstdout:\n%s\nstderr:\n%s", args, code, want, stdout, stderr)
	}
	return stdout, stderr
}

func TestGlobalAddPublishesProfileLockBeforeNativeMaterialization(t *testing.T) {
	fixture := newGlobalLockCLIFixture(t)
	_, stderr := requireGlobalCommand(t, fixture, exitOK, "global", "add", "new-skill", "--git", fixture.gitURL, "--tag", "v1")
	requireOnePermissivePostureWarning(t, stderr)
	fixture.requireSkillLockAndStore(t)
	fixture.requireNativeSkill(t)
	if _, err := os.Lstat(fixture.managed); err != nil {
		t.Fatalf("managed-home skill was not materialized: %v", err)
	}
}

func TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers(t *testing.T) {
	fixture := newGlobalLockCLIFixture(t)
	requireGlobalCommand(t, fixture, exitOK, "profile", "sync")
	_, beforeHash := fixture.lock(t)
	conflictDir := filepath.Join(fixture.codex, ".agent-context", "skills", "new-skill")
	conflictFile := filepath.Join(conflictDir, "SKILL.md")
	writeFile(t, conflictFile, "operator-owned preimage\n")
	manifestPath := filepath.Join(fixture.home, "global", "Skillfile.json")
	manifestBefore, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	markerPath := filepath.Join(fixture.codex, envmarker.Name)
	markerBefore, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}

	_, stderr := requireGlobalCommand(t, fixture, exitFail, "global", "add", "new-skill", "--git", fixture.gitURL, "--tag", "v1")
	if !strings.Contains(stderr, envprofile.DiagUnmanagedConflict) || !strings.Contains(stderr, ".agent-context/skills/new-skill/SKILL.md") {
		t.Fatalf("global add conflict stderr = %q", stderr)
	}
	_, afterHash := fixture.requireSkillLockAndStore(t)
	if afterHash == beforeHash {
		t.Fatalf("profile lock hash stayed %s after the blocked add", afterHash)
	}
	preimage, err := os.ReadFile(conflictFile)
	if err != nil || string(preimage) != "operator-owned preimage\n" {
		t.Fatalf("conflicting surface changed: %q, %v", preimage, err)
	}
	markerAfter, err := os.ReadFile(markerPath)
	if err != nil || string(markerAfter) != string(markerBefore) {
		t.Fatalf("native marker changed on conflict: err=%v", err)
	}
	manifestAfter, err := os.ReadFile(manifestPath)
	if err != nil || string(manifestAfter) != string(manifestBefore) {
		t.Fatalf("legacy declaration changed on conflict: err=%v", err)
	}
	if _, err := os.Lstat(fixture.managed); !os.IsNotExist(err) {
		t.Fatalf("managed surface changed before native conflict was reported: %v", err)
	}

	requireGlobalCommand(t, fixture, exitOK, "profile", "sync", "--takeover")
	backup := filepath.Join(fixture.codex, ".agent-environment-backup", "1", ".agent-context", "skills", "new-skill", "SKILL.md")
	backupBytes, err := os.ReadFile(backup)
	if err != nil || string(backupBytes) != "operator-owned preimage\n" {
		t.Fatalf("takeover backup = %q, %v", backupBytes, err)
	}
	fixture.requireNativeSkill(t)
	requireGlobalCommand(t, fixture, exitOK, "global", "add", "new-skill", "--git", fixture.gitURL, "--tag", "v1")
	fixture.requireNativeSkill(t)
}

func TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization(t *testing.T) {
	fixture := newGlobalLockCLIFixture(t)
	fixture.writeLegacyGlobalDeclaration(t)
	requireGlobalCommand(t, fixture, exitOK, "global", "install")
	fixture.requireSkillLockAndStore(t)
	fixture.requireNativeSkill(t)
}

func TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers(t *testing.T) {
	fixture := newGlobalLockCLIFixture(t)
	fixture.writeLegacyGlobalDeclaration(t)
	conflictDir := filepath.Join(fixture.codex, ".agent-context", "skills", "new-skill")
	conflictFile := filepath.Join(conflictDir, "SKILL.md")
	writeFile(t, conflictFile, "operator-owned preimage\n")
	manifestPath := filepath.Join(fixture.home, "global", "Skillfile.json")
	manifestBefore, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	_, stderr := requireGlobalCommand(t, fixture, exitFail, "global", "install")
	if !strings.Contains(stderr, envprofile.DiagUnmanagedConflict) || !strings.Contains(stderr, ".agent-context/skills/new-skill/SKILL.md") {
		t.Fatalf("global install conflict stderr = %q", stderr)
	}
	fixture.requireSkillLockAndStore(t)
	preimage, err := os.ReadFile(conflictFile)
	if err != nil || string(preimage) != "operator-owned preimage\n" {
		t.Fatalf("conflicting surface changed: %q, %v", preimage, err)
	}
	manifestAfter, err := os.ReadFile(manifestPath)
	if err != nil || string(manifestAfter) != string(manifestBefore) {
		t.Fatalf("global install changed its declaration source on conflict: err=%v", err)
	}

	requireGlobalCommand(t, fixture, exitOK, "profile", "sync", "--takeover")
	backup := filepath.Join(fixture.codex, ".agent-environment-backup", "1", ".agent-context", "skills", "new-skill", "SKILL.md")
	backupBytes, err := os.ReadFile(backup)
	if err != nil || string(backupBytes) != "operator-owned preimage\n" {
		t.Fatalf("takeover backup = %q, %v", backupBytes, err)
	}
	fixture.requireNativeSkill(t)
	requireGlobalCommand(t, fixture, exitOK, "global", "install")
	fixture.requireNativeSkill(t)
}
