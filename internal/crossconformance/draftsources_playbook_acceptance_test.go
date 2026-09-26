package crossconformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
)

const (
	playbookFixtureURL = "https://fixture.test/playbook.git"
	rolesFixtureURL    = "https://fixture.test/roles.git"
)

// TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI binds the
// two schema-2 selectors in one user-visible lifecycle: a collection from a
// tagged playbook repository and a transitive manifest dependency selected
// from a folder in another repository. It also proves update, committed-lock
// replay on a new home, tamper and unavailable-source refusals, and the
// negative selector rows through the compiled CLI.
func TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; native admission remains covered on Windows")
	}
	realGit := requireGit(t)
	root := t.TempDir()
	repoRoot := filepath.Join(root, "repositories")
	playbookWork, playbookBare, playbookV1 := buildPlaybookFixture(t, repoRoot)
	rolesBare, rolesCommit := buildRolesFixture(t, repoRoot)
	gitWrapper := writePlaybookFixtureGitWrapper(t, root, realGit, map[string]string{
		playbookFixtureURL: playbookBare,
		rolesFixtureURL:    rolesBare,
	})

	configPath, project, home := setupCLIProject(t, filepath.Join(root, "lifecycle"))
	enablePlaybookAdvisoryAudit(t, configPath)
	writePlaybookSourcePolicy(t, home)
	writePlaybookSkillfile(t, project, "v1.0.0", []string{"*"}, "skills")

	code, stdout, stderr := runPlaybookFixtureCLI(t, gitWrapper, home, configPath, "project", "resolve", "app")
	logPlaybookCommand(t, []string{"project", "resolve", "app"}, code)
	if code != 0 {
		t.Fatalf("initial resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	code, stdout, stderr = runPlaybookFixtureCLI(t, gitWrapper, home, configPath, "install", "app", "--audit", "advisory")
	logPlaybookCommand(t, []string{"install", "app", "--audit", "advisory"}, code)
	if code != 0 {
		t.Fatalf("initial install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	lockV1, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatalf("read initial lock: %v", err)
	}
	assertPlaybookLock(t, lockV1, playbookV1, rolesCommit, false)
	assertPlaybookInstalled(t, project, []string{"orchestrator", "developer", "reviewer", "qa"})
	assertPlaybookStatus(t, gitWrapper, home, configPath, []string{"orchestrator", "developer", "reviewer", "qa"})
	assertPlaybookAuditCoversLock(t, home, lockV1)
	lockV1Bytes, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}

	// Publish one more playbook skill as v1.1.0, then make the declared
	// source ref explicit and refresh through the documented project API.
	writeManifestDependencySkill(t, filepath.Join(playbookWork, "skills", "writer"), "writer", 4, map[string]map[string]any{})
	runDraftGit(t, playbookWork, "add", ".")
	runDraftGit(t, playbookWork, "commit", "-qm", "add writer")
	runDraftGit(t, playbookWork, "tag", "v1.1.0")
	playbookV11 := strings.TrimSpace(draftGitOutput(t, playbookWork, "rev-parse", "v1.1.0^{commit}"))
	runDraftGit(t, playbookWork, "push", "origin", "main", "refs/tags/v1.1.0:refs/tags/v1.1.0")
	writePlaybookSkillfile(t, project, "v1.1.0", []string{"*"}, "skills")

	code, stdout, stderr = runPlaybookFixtureCLI(t, gitWrapper, home, configPath, "project", "refresh", "app")
	logPlaybookCommand(t, []string{"project", "refresh", "app"}, code)
	if code != 0 {
		t.Fatalf("project refresh = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	code, stdout, stderr = runPlaybookFixtureCLI(t, gitWrapper, home, configPath, "install", "app", "--audit", "advisory")
	logPlaybookCommand(t, []string{"install", "app", "--audit", "advisory"}, code)
	if code != 0 {
		t.Fatalf("install after refresh = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lockV11, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatalf("read refreshed lock: %v", err)
	}
	assertPlaybookLock(t, lockV11, playbookV11, rolesCommit, true)
	lockV11Bytes, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(lockV11Bytes) == string(lockV1Bytes) {
		t.Fatal("refresh left the lock unchanged after adding skills/writer at v1.1.0")
	}
	assertPlaybookInstalled(t, project, []string{"orchestrator", "developer", "reviewer", "writer", "qa"})
	assertPlaybookStatus(t, gitWrapper, home, configPath, []string{"orchestrator", "developer", "reviewer", "writer", "qa"})
	assertPlaybookAuditCoversLock(t, home, lockV11)

	// Commit only the project declaration and its lock. The fresh home below
	// receives a clone of this committed project and starts without snapshots.
	runDraftGit(t, project, "add", "Skillfile.json", "Skillfile.lock.json")
	runDraftGit(t, project, "commit", "-qm", "lock playbook v1.1.0")
	freshProject := filepath.Join(root, "fresh-project")
	runDraftGit(t, "", "clone", "--quiet", "--", project, freshProject)
	committedLock, err := os.ReadFile(filepath.Join(freshProject, "Skillfile.lock.json"))
	if err != nil {
		t.Fatalf("fresh clone has no committed lock: %v", err)
	}
	if string(committedLock) != string(lockV11Bytes) {
		t.Fatal("fresh clone does not carry the committed lock bytes")
	}

	// Move the declared tag after the lock was committed. Replay must use the
	// locked commit, identity, and content hash rather than resolving the tag.
	writeManifestDependencySkill(t, filepath.Join(playbookWork, "skills", "unlocked"), "unlocked", 4, map[string]map[string]any{})
	runDraftGit(t, playbookWork, "add", ".")
	runDraftGit(t, playbookWork, "commit", "-qm", "move the v1.1.0 tag")
	runDraftGit(t, playbookWork, "tag", "-f", "v1.1.0")
	runDraftGit(t, playbookWork, "push", "--force", "origin", "main", "refs/tags/v1.1.0:refs/tags/v1.1.0")
	unlockedCommit := strings.TrimSpace(draftGitOutput(t, playbookWork, "rev-parse", "v1.1.0^{commit}"))
	if unlockedCommit == playbookV11 {
		t.Fatal("fixture did not move v1.1.0 away from the committed lock")
	}
	remoteTag := draftGitOutput(t, playbookWork, "ls-remote", "origin", "refs/tags/v1.1.0")
	if got := strings.Fields(remoteTag); len(got) != 2 || got[0] != unlockedCommit {
		t.Fatalf("remote v1.1.0 = %q, want moved commit %s", remoteTag, unlockedCommit)
	}

	freshConfig, freshHome := setupPlaybookHomeForProject(t, filepath.Join(root, "fresh-machine"), freshProject)
	writePlaybookSourcePolicy(t, freshHome)
	code, stdout, stderr = runPlaybookFixtureCLI(t, gitWrapper, freshHome, freshConfig, "install", "app", "--audit", "advisory")
	logPlaybookCommand(t, []string{"fresh-home install", "app", "--audit", "advisory"}, code)
	if code != 0 {
		t.Fatalf("fresh-home install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	freshLockPath := filepath.Join(freshProject, "Skillfile.lock.json")
	freshLockBytes, err := os.ReadFile(freshLockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(freshLockBytes) != string(committedLock) {
		t.Fatal("fresh-home replay changed the committed lock bytes")
	}
	freshLock, err := sourcelock.Read(freshLockPath)
	if err != nil {
		t.Fatal(err)
	}
	assertPlaybookLock(t, freshLock, playbookV11, rolesCommit, true)
	assertPlaybookInstalled(t, freshProject, []string{"orchestrator", "developer", "reviewer", "writer", "qa"})
	assertPlaybookStatus(t, gitWrapper, freshHome, freshConfig, []string{"orchestrator", "developer", "reviewer", "writer", "qa"})
	assertPlaybookAuditCoversLock(t, freshHome, freshLock)
	playbookSnapshot := snapshot.Dir(freshHome, "fixture.test/playbook", playbookV11)
	if _, err := os.Stat(filepath.Join(playbookSnapshot, "skills", "unlocked", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("fresh-home replay used the moved v1.1.0 tag: unlocked skill stat = %v", err)
	}
	if err := os.WriteFile(filepath.Join(playbookSnapshot, "skills", "developer", "SKILL.md"), []byte("tampered after fresh replay\n"), 0o644); err != nil {
		t.Fatalf("tamper fresh cached snapshot: %v", err)
	}
	code, _, stderr = runPlaybookFixtureCLI(t, gitWrapper, freshHome, freshConfig, "install", "app", "--audit", "advisory")
	logPlaybookCommand(t, []string{"tampered fresh-home install", "app", "--audit", "advisory"}, code)
	if code != 1 || !strings.Contains(stderr, "source_snapshot_changed") {
		t.Fatalf("tampered fresh-home install = %d, stderr %q, want source_snapshot_changed", code, stderr)
	}
	lockAfterTamper, err := os.ReadFile(freshLockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfterTamper) != string(committedLock) {
		t.Fatal("tampered replay changed the committed lock")
	}

	// An empty third home cannot recover a lock whose declared repositories
	// are unavailable. It must report the missing snapshot instead of treating
	// a failed read as an absent package or silently re-resolving.
	unreachableGit := writePlaybookFixtureGitWrapper(t, root, realGit, map[string]string{
		playbookFixtureURL: filepath.Join(root, "unreachable-playbook.git"),
		rolesFixtureURL:    filepath.Join(root, "unreachable-roles.git"),
	})
	unreachableConfig, unreachableHome := setupPlaybookHomeForProject(t, filepath.Join(root, "unreachable-machine"), freshProject)
	writePlaybookSourcePolicy(t, unreachableHome)
	code, _, stderr = runPlaybookFixtureCLI(t, unreachableGit, unreachableHome, unreachableConfig, "install", "app", "--audit", "advisory")
	logPlaybookCommand(t, []string{"unreachable-home install", "app", "--audit", "advisory"}, code)
	if code != 1 || !strings.Contains(stderr, "source_snapshot_unavailable") {
		t.Fatalf("unreachable-source install = %d, stderr %q, want source_snapshot_unavailable", code, stderr)
	}

	runPlaybookNegativeRow(t, filepath.Join(root, "negative-include"), gitWrapper, "include matches no skill", func(project string) {
		writePlaybookSkillfile(t, project, "v1.0.0", []string{"missing"}, "skills")
	}, "source_member_missing")
	runPlaybookNegativeRow(t, filepath.Join(root, "negative-directory"), gitWrapper, "directory escapes repository", func(project string) {
		writePlaybookSkillfile(t, project, "v1.0.0", []string{"*"}, "../outside")
	}, "source_selection_invalid")
	runPlaybookNegativeRow(t, filepath.Join(root, "negative-dependency"), gitWrapper, "dependency directory has no SKILL.md", func(project string) {
		writeMissingSkillDependency(t, project)
	}, "no SKILL.md")
}

func buildPlaybookFixture(t *testing.T, repoRoot string) (work, bare, v1Commit string) {
	t.Helper()
	work = filepath.Join(repoRoot, "playbook-work")
	bare = filepath.Join(repoRoot, "playbook.git")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifestDependencySkill(t, filepath.Join(work, "skills", "orchestrator"), "orchestrator", 4, map[string]map[string]any{})
	writeManifestDependencySkill(t, filepath.Join(work, "skills", "reviewer"), "reviewer", 4, map[string]map[string]any{})
	writeManifestDependencySkill(t, filepath.Join(work, "skills", "developer"), "developer", 9, map[string]map[string]any{
		"qa": {
			"git":       rolesFixtureURL,
			"ref":       map[string]any{"kind": "tag", "value": "v1.0.0"},
			"directory": "roles/qa",
		},
	})
	runDraftGit(t, work, "init", "-q", "-b", "main")
	runDraftGit(t, work, "add", ".")
	runDraftGit(t, work, "commit", "-qm", "playbook v1.0.0")
	runDraftGit(t, work, "tag", "v1.0.0")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	runDraftGit(t, work, "remote", "add", "origin", bare)
	v1Commit = strings.TrimSpace(draftGitOutput(t, work, "rev-parse", "v1.0.0^{commit}"))
	return work, bare, v1Commit
}

func buildRolesFixture(t *testing.T, repoRoot string) (bare, commit string) {
	t.Helper()
	work := filepath.Join(repoRoot, "roles-work")
	bare = filepath.Join(repoRoot, "roles.git")
	if err := os.MkdirAll(filepath.Join(work, "roles", "no-skill"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifestDependencySkill(t, filepath.Join(work, "roles", "qa"), "qa", 4, map[string]map[string]any{})
	if err := os.WriteFile(filepath.Join(work, "roles", "no-skill", "README.md"), []byte("not a skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, work, "init", "-q", "-b", "main")
	runDraftGit(t, work, "add", ".")
	runDraftGit(t, work, "commit", "-qm", "roles v1.0.0")
	runDraftGit(t, work, "tag", "v1.0.0")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	commit = strings.TrimSpace(draftGitOutput(t, work, "rev-parse", "v1.0.0^{commit}"))
	return bare, commit
}

func writePlaybookFixtureGitWrapper(t *testing.T, root, realGit string, remotes map[string]string) string {
	t.Helper()
	dir, err := os.MkdirTemp(root, "git-wrapper-")
	if err != nil {
		t.Fatal(err)
	}
	urls := make([]string, 0, len(remotes))
	for url := range remotes {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	var script strings.Builder
	script.WriteString("#!/bin/sh\nargs=\"\"\nfor arg in \"$@\"; do\n  case \"$arg\" in\n")
	for _, url := range urls {
		script.WriteString("    " + shellQuotePlaybook(url) + ") arg=" + shellQuotePlaybook("file://"+remotes[url]) + ";;\n")
	}
	script.WriteString("  esac\n  args=\"$args\n$arg\"\ndone\noldifs=$IFS\nIFS='\n'\nset -- $args\nIFS=$oldifs\nexec ")
	script.WriteString(shellQuotePlaybook(realGit))
	script.WriteString(" \"$@\"\n")
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func shellQuotePlaybook(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func runPlaybookFixtureCLI(t *testing.T, gitWrapper, home, configPath string, args ...string) (int, string, string) {
	t.Helper()
	env := []string{
		"PATH=" + gitWrapper + string(os.PathListSeparator) + os.Getenv("PATH"),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_NOSYSTEM=1",
	}
	return runCurator(t, home, configPath, env, args...)
}

func logPlaybookCommand(t *testing.T, args []string, code int) {
	t.Helper()
	t.Logf("curator %s: exit %d", strings.Join(args, " "), code)
}

func setupPlaybookHomeForProject(t *testing.T, root, project string) (configPath, home string) {
	t.Helper()
	home = filepath.Join(root, "home")
	configPath = filepath.Join(home, "config.json")
	skillsRoot := filepath.Join(root, "skills-root")
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "bootstrap", "--non-interactive", "--skills-root", skillsRoot); code != 0 {
		t.Fatalf("bootstrap fresh manager = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "project", "add", "app", project, "--agents", "codex_cli"); code != 0 {
		t.Fatalf("register fresh project = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	return configPath, home
}

func enablePlaybookAdvisoryAudit(t *testing.T, configPath string) {
	t.Helper()
	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(payload, &config); err != nil {
		t.Fatal(err)
	}
	config["audit"] = map[string]any{
		"enabled": true, "mode": "advisory", "backend": "null", "registry_policy": "advisory",
	}
	updated, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, updated, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writePlaybookSourcePolicy(t *testing.T, home string) {
	t.Helper()
	endpoint := func(url string) map[string]any {
		return map[string]any{"endpoints": []any{map[string]any{"url": url, "authentication": "team-https"}}, "fallback": "none"}
	}
	policy := map[string]any{
		"schema_version": 1,
		"repositories": map[string]any{
			"fixture.test/playbook": endpoint(playbookFixtureURL),
			"fixture.test/roles":    endpoint(rolesFixtureURL),
		},
	}
	payload, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "source-policy.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writePlaybookSkillfile(t *testing.T, project, tag string, include []string, directory string) {
	t.Helper()
	selection := map[string]any{"from": "playbook", "directory": directory}
	if include != nil {
		selection["include"] = include
	}
	document := map[string]any{
		"schema_version": 2,
		"sources":        map[string]any{"playbook": map[string]any{"git": playbookFixtureURL, "tag": tag}},
		"skills":         []any{selection},
	}
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertPlaybookLock(t *testing.T, lock *sourcelock.Lock, playbookCommit, rolesCommit string, withWriter bool) {
	t.Helper()
	rootSkills := []string{"orchestrator", "developer", "reviewer"}
	if withWriter {
		rootSkills = append(rootSkills, "writer")
	}
	if len(lock.Members) != len(rootSkills)+1 {
		t.Fatalf("lock members = %d (%+v), want %d", len(lock.Members), lock.Members, len(rootSkills)+1)
	}
	for _, name := range rootSkills {
		assertPlaybookLockMember(t, lock, name, "fixture.test/playbook", playbookCommit, "skills/"+name, true)
	}
	assertPlaybookLockMember(t, lock, "qa", "fixture.test/roles", rolesCommit, "roles/qa", false)
}

func assertPlaybookLockMember(t *testing.T, lock *sourcelock.Lock, name, repository, commit, directory string, selected bool) {
	t.Helper()
	member, ok := lock.Find(name)
	if !ok {
		t.Fatalf("lock misses %s", name)
	}
	if member.Package.Kind != sourcelock.KindNetworkGit || member.Package.Repository != repository ||
		member.Package.Commit.Hex != commit || member.Package.Directory != directory || member.Directory != directory {
		t.Fatalf("lock member %s identity = %+v, want %s@%s:%s", name, member.Package, repository, commit, directory)
	}
	if !strings.HasPrefix(member.ContentSHA256, "sha256:") || len(member.ContentSHA256) != len("sha256:")+64 {
		t.Fatalf("lock member %s content_sha256 = %q", name, member.ContentSHA256)
	}
	if selected != (member.Selection != nil) {
		t.Fatalf("lock member %s selection = %v, selected root = %t", name, member.Selection, selected)
	}
	t.Logf("lock member %s: repository=%s commit=%s directory=%s content_sha256=%s selected=%t", name, member.Package.Repository, member.Package.Commit.Hex, member.Package.Directory, member.ContentSHA256, selected)
}

func assertPlaybookInstalled(t *testing.T, project string, names []string) {
	t.Helper()
	for _, name := range names {
		payload, err := os.ReadFile(filepath.Join(project, ".agents", "skills", name, "SKILL.md"))
		if err != nil {
			t.Fatalf("installed skill %s is missing: %v", name, err)
		}
		if !strings.Contains(string(payload), "name: "+name) {
			t.Fatalf("installed skill %s has unexpected SKILL.md: %s", name, payload)
		}
	}
}

func assertPlaybookStatus(t *testing.T, gitWrapper, home, configPath string, names []string) {
	t.Helper()
	code, stdout, stderr := runPlaybookFixtureCLI(t, gitWrapper, home, configPath, "status", "app")
	logPlaybookCommand(t, []string{"status", "app"}, code)
	if code != 0 {
		t.Fatalf("status = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	for _, name := range names {
		if !strings.Contains(stdout, "app: "+name+" up-to-date") {
			t.Fatalf("status omits %s up-to-date:\n%s", name, stdout)
		}
	}
}

func assertPlaybookAuditCoversLock(t *testing.T, home string, lock *sourcelock.Lock) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, "source-audit", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	var records []audit.SourceAudit
	for _, path := range paths {
		if strings.HasSuffix(path, ".report.json") {
			continue
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		record, err := audit.ParseSourceAudit(payload)
		if err != nil {
			t.Fatalf("parse source audit %s: %v", path, err)
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		t.Fatalf("no source-audit records under %s", home)
	}
	for _, member := range lock.Members {
		found := false
		for _, record := range records {
			if record.Package.Kind == member.Package.Kind && record.Package.Repository == member.Package.Repository &&
				record.Package.Commit != nil && record.Package.Commit.Hex == member.Package.Commit.Hex &&
				record.Package.Directory == member.Package.Directory && record.ContentSHA256 == member.ContentSHA256 {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("audit records do not bind lock member %s at %s: %s", member.Name, member.Package.Directory, member.ContentSHA256)
		}
	}
	t.Logf("verified %d parsed source-audit records for %d locked members", len(records), len(lock.Members))
}

func runPlaybookNegativeRow(t *testing.T, root, gitWrapper, name string, writeManifest func(string), wantDiagnostic string) {
	t.Helper()
	configPath, project, home := setupCLIProject(t, root)
	enablePlaybookAdvisoryAudit(t, configPath)
	writePlaybookSourcePolicy(t, home)
	writeManifest(project)
	code, stdout, stderr := runPlaybookFixtureCLI(t, gitWrapper, home, configPath, "project", "resolve", "app")
	logPlaybookCommand(t, []string{"negative", name, "project", "resolve", "app"}, code)
	if code != 1 {
		t.Fatalf("negative row %q returned exit %d, want refusal exit 1\nstdout:\n%s", name, code, stdout)
	}
	if !strings.Contains(stderr, wantDiagnostic) {
		t.Fatalf("negative row %q stderr misses %q:\n%s", name, wantDiagnostic, stderr)
	}
}

func writeMissingSkillDependency(t *testing.T, project string) {
	t.Helper()
	writeManifestDependencySkill(t, filepath.Join(project, "pkgs", "consumer"), "consumer", 9, map[string]map[string]any{
		"missing-skill": {
			"git":       rolesFixtureURL,
			"ref":       map[string]any{"kind": "tag", "value": "v1.0.0"},
			"directory": "roles/no-skill",
		},
	})
	document := map[string]any{
		"schema_version": 2,
		"sources":        map[string]any{"project": map[string]any{"path": "./pkgs"}},
		"skills":         []any{map[string]any{"name": "consumer", "from": "project", "directory": "consumer"}},
	}
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}
