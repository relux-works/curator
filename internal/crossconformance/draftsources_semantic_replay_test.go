package crossconformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
)

func init() {
	registerDraftSemantic("path-missing-snapshot-identical-bytes", drivePathReplayIdentical)
	registerDraftSemantic("path-missing-snapshot-drifted-bytes", drivePathReplayDrifted)
	registerDraftSemantic("git-missing-snapshot-fetches-locked-commit", driveGitMissingSnapshotFetchesLockedCommit)
	registerDraftSemantic("git-moved-tag-replays-locked-commit", driveGitMovedTagReplay)
	registerDraftSemantic("git-moved-tag-replays-locked-commit-through-mirror", driveGitMovedTagReplayThroughMirror)
	registerDraftSemantic("missing-snapshot-unreachable-source", driveGitReplayUnreachable)
}

func drivePathReplayIdentical(t *testing.T, c draftSemanticCase) {
	project, home, lockBefore, _ := prepareLocalReplay(t)
	result := draftInstall(t, project, home, install.Options{Platform: draftPlatform()})
	if result.Status != "ok" {
		t.Fatalf("install = %+v, want replay of identical path bytes", result)
	}
	lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(lockBefore) != string(lockAfter) {
		t.Fatal("path replay rewrote the existing lock")
	}
	if got := strings.Join([]string{"install", "lock-byte-identical"}, ";"); got != c.Expected {
		t.Fatalf("observed outcome %q differs from corpus expectation %q", got, c.Expected)
	}
}

func drivePathReplayDrifted(t *testing.T, c draftSemanticCase) {
	project, home, lockBefore, sourceRoot := prepareLocalReplay(t)
	if err := os.WriteFile(filepath.Join(sourceRoot, "skills", "review", "SKILL.md"), []byte("---\nname: review\ndescription: changed\n---\n# review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := draftInstall(t, project, home, install.Options{Platform: draftPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
		t.Fatalf("install = %+v, want source_snapshot_changed", result)
	}
	lockAfter, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(lockBefore) != string(lockAfter) {
		t.Fatal("refused path replay rewrote the existing lock")
	}
	if got := "source_snapshot_changed"; got != c.Expected {
		t.Fatalf("observed outcome %q differs from corpus expectation %q", got, c.Expected)
	}
}

func prepareLocalReplay(t *testing.T) (project, home string, lockBefore []byte, sourceRoot string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"path":"src"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	project, home = draftProject(t, string(payload), map[string]string{"src/skills/review": "review"})
	sourceRoot = filepath.Join(project, "src")
	plan := resolveDraftPlan(t, project, home, string(payload))
	member, ok := plan.Lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	lockBefore, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if member.Package.Kind != sourcelock.KindLocalSnapshot {
		t.Fatalf("package kind = %q, want local snapshot", member.Package.Kind)
	}
	if err := os.RemoveAll(snapshot.LocalStoreDir(home)); err != nil {
		t.Fatal(err)
	}
	return project, home, lockBefore, sourceRoot
}

type draftGitReplayCLI struct {
	configPath string
	project    string
	home       string
	identity   string
	work       string
	emptyRepo  string
	lockBefore []byte
	commit     string
	member     sourcelock.Member
}

func prepareGitReplayCLI(t *testing.T, identityValue, sourceKind, sourceURL, tag string) draftGitReplayCLI {
	t.Helper()
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	var payload string
	// Build the source declaration explicitly to keep the schema's source
	// arm closed and ensure both URL and logical-repository replay use the
	// same production lock path.
	if sourceKind == "repository" {
		payload = `{"schema_version":2,"sources":{"s":{"repository":"` + identityValue + `","tag":"` + tag + `"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	} else {
		payload = `{"schema_version":2,"sources":{"s":{"git":"` + sourceURL + `","tag":"` + tag + `"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	}
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(root, "source-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(work, "skills", "review")
	writeDraftSkill(t, skillDir, "review")
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: review\ndescription: locked bytes\n---\n# review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, work, "init", "-q", "-b", "main")
	runDraftGit(t, work, "add", ".")
	runDraftGit(t, work, "commit", "-qm", "locked commit")
	runDraftGit(t, work, "tag", tag)
	commit := draftGitOutput(t, work, "rev-parse", "HEAD")
	m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload),
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": work}},
	})
	if err != nil {
		t.Fatalf("ResolveDraft: %v", err)
	}
	lockPath := sourcelock.PathIn(project)
	if err := sourcelock.Write(lockPath, plan.Lock); err != nil {
		t.Fatal(err)
	}
	member, ok := plan.Lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	lockBefore, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	emptyRepo := filepath.Join(home, "replay-empty")
	if err := os.MkdirAll(filepath.Dir(emptyRepo), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(emptyRepo, 0o755); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, emptyRepo, "init", "-q")
	bindings, err := sourcelock.NewBindings(plan.Lock.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: emptyRepo}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(install.DraftBindingsPath(home, project), bindings); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(snapshot.Dir(home, member.Package.Repository, member.Package.Commit.Hex)); err != nil {
		t.Fatal(err)
	}
	return draftGitReplayCLI{configPath: configPath, project: project, home: home, identity: identityValue,
		work: work, emptyRepo: emptyRepo, lockBefore: lockBefore, commit: commit, member: member}
}

func replayGitReplayCLI(t *testing.T, fixture draftGitReplayCLI, endpoint, bare string, unavailable bool) (int, string, string, string) {
	t.Helper()
	env, logPath := draftEndpointGitShim(t, endpoint, bare, unavailable)
	code, stdout, stderr := runCurator(t, fixture.home, fixture.configPath, env, "install", "app")
	return code, stdout, stderr, readDraftFetchLog(t, logPath)
}

func driveGitMissingSnapshotFetchesLockedCommit(t *testing.T, c draftSemanticCase) {
	const identityValue = "example.org/kit"
	fixture := prepareGitReplayCLI(t, identityValue, "git", "https://example.org/kit.git", "v1.0.0")
	bare := filepath.Join(t.TempDir(), "source.git")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", fixture.work, bare)
	endpoint := "https://example.org/kit.git"
	code, stdout, stderr, logText := replayGitReplayCLI(t, fixture, endpoint, bare, false)
	if code != 0 {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(logText, endpoint) || !strings.Contains(logText, fixture.commit) {
		t.Fatalf("fetch log = %q, want declared endpoint and locked object id", logText)
	}
	after, err := os.ReadFile(sourcelock.PathIn(fixture.project))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(fixture.lockBefore) {
		t.Fatal("Git replay rewrote the existing lock")
	}
	got := strings.Join([]string{"fetch-locked-commit", "install", "lock-byte-identical"}, ";")
	if got != c.Expected {
		t.Fatalf("observed outcome %q differs from corpus expectation %q", got, c.Expected)
	}
}

func driveGitMovedTagReplay(t *testing.T, c draftSemanticCase) {
	driveMovedTagReplay(t, c, false)
}

func driveGitMovedTagReplayThroughMirror(t *testing.T, c draftSemanticCase) {
	driveMovedTagReplay(t, c, true)
}

func driveMovedTagReplay(t *testing.T, c draftSemanticCase, mirror bool) {
	identityValue := "example.org/kit"
	sourceKind := "git"
	url := "https://example.org/kit.git"
	endpoint := url
	if mirror {
		sourceKind = "repository"
		url = ""
		endpoint = "https://mirror.example.net/kit.git"
	}
	fixture := prepareGitReplayCLI(t, identityValue, sourceKind, url, "v1.0.0")
	if fixture.member.Package.Commit.Hex != fixture.commit {
		t.Fatalf("locked member commit = %s, want %s", fixture.member.Package.Commit.Hex, fixture.commit)
	}
	changed := []byte("---\nname: review\ndescription: moved tag\n---\n# review\n")
	if err := os.WriteFile(filepath.Join(fixture.work, "skills", "review", "SKILL.md"), changed, 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, fixture.work, "add", ".")
	runDraftGit(t, fixture.work, "commit", "-qm", "moved tag")
	runDraftGit(t, fixture.work, "tag", "-f", "v1.0.0")
	bare := filepath.Join(t.TempDir(), "source.git")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", fixture.work, bare)
	if mirror {
		policyDoc := `{"schema_version":2,"repositories":{"` + identityValue + `":{"endpoints":[{"url":"` + endpoint + `","authentication":"team-https","mirror_of":"` + identityValue + `"}],"fallback":"none"}}}`
		if err := os.WriteFile(filepath.Join(filepath.Dir(fixture.configPath), "source-policy.json"), []byte(policyDoc), 0o644); err != nil {
			t.Fatal(err)
		}
		policy := v2Parse(t, policyDoc)
		resolution := v2ResolveFor(t, policy, identityValue)
		if len(resolution.Attempts) != 1 || resolution.Attempts[0].URL != endpoint {
			t.Fatalf("mirror plan = %+v", resolution.Attempts)
		}
	}
	code, stdout, stderr, logText := replayGitReplayCLI(t, fixture, endpoint, bare, false)
	if code != 0 {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(logText, endpoint) || !strings.Contains(logText, fixture.commit) || strings.Contains(logText, "v1.0.0") || strings.Contains(logText, "refs/tags") {
		t.Fatalf("fetch log = %q, want only the locked object id, without tag resolution", logText)
	}
	recorded := marker.Read(filepath.Join(fixture.project, ".agents", "skills", "review"))
	if recorded == nil || recorded.Package.Repository != identityValue || recorded.Package.Commit.Hex != fixture.commit || recorded.ContentSHA256 != fixture.member.ContentSHA256 {
		t.Fatalf("installed marker = %+v, want locked repository, commit and content", recorded)
	}
	after, err := os.ReadFile(sourcelock.PathIn(fixture.project))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(fixture.lockBefore) {
		t.Fatal("moved-tag replay rewrote the existing lock")
	}
	fetchStep := "fetch-commit-A-by-object-id"
	if mirror {
		fetchStep = "fetch-commit-A-by-object-id-from-listed-mirror"
	}
	got := strings.Join([]string{fetchStep, "verify-object-format-and-commit", "verify-package-identity-and-content-sha256", "install", "no-tag-resolution", "lock-byte-identical"}, ";")
	if got != c.Expected {
		t.Fatalf("observed outcome %q differs from corpus expectation %q", got, c.Expected)
	}
}

func driveGitReplayUnreachable(t *testing.T, c draftSemanticCase) {
	const identityValue = "unreachable.fixture.test/kit"
	endpoint := "https://unreachable.fixture.test/kit.git"
	fixture := prepareGitReplayCLI(t, identityValue, "git", endpoint, "v1.0.0")
	code, stdout, stderr, logText := replayGitReplayCLI(t, fixture, endpoint, "", true)
	if code == 0 || !strings.Contains(stderr, "source_snapshot_unavailable") {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s; want source_snapshot_unavailable", code, stdout, stderr)
	}
	after, err := os.ReadFile(sourcelock.PathIn(fixture.project))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(fixture.lockBefore) {
		t.Fatal("unreachable replay rewrote the existing lock")
	}
	if !strings.Contains(logText, endpoint) {
		t.Fatalf("fetch log = %q, want declared endpoint attempt", logText)
	}
	if got := "source_snapshot_unavailable"; got != c.Expected {
		t.Fatalf("observed outcome %q differs from corpus expectation %q", got, c.Expected)
	}
}

// TestDraftSourcesReplayRejectsLockedObjectFormatMismatch is C1's killing row:
// the 40-character SHA-1 prefix resolves to a real object in a SHA-256 Git
// repository, so a length-only commit check would accept and install it.
func TestDraftSourcesReplayRejectsLockedObjectFormatMismatch(t *testing.T) {
	const payload = `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1.0.0"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	project, home := draftProject(t, payload, nil)
	work := filepath.Join(t.TempDir(), "sha256-source")
	if err := os.MkdirAll(filepath.Join(work, "skills", "review"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, filepath.Join(work, "skills", "review"), "review")
	if err := os.WriteFile(filepath.Join(work, "skills", "review", "SKILL.md"), []byte("---\nname: review\ndescription: sha256 source\n---\n# review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, work, "init", "-q", "--object-format=sha256")
	runDraftGit(t, work, "add", ".")
	runDraftGit(t, work, "commit", "-qm", "sha256 source")
	runDraftGit(t, work, "tag", "v1.0.0")
	m := draftManifest(t, project, payload)
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot:     project,
		Home:            home,
		Manifest:        m,
		ManifestPayload: []byte(payload),
		Expansion:       manifest.ExpansionOptions{GitRoots: map[string]string{"s": work}},
	})
	if err != nil {
		t.Fatalf("ResolveDraft: %v", err)
	}
	member, ok := plan.Lock.Find("review")
	if !ok || member.Package.Commit.ObjectFormat != "sha256" || len(member.Package.Commit.Hex) != 64 {
		t.Fatalf("resolved member = %+v; want SHA-256 commit", member)
	}
	members := append([]sourcelock.Member(nil), plan.Lock.Members...)
	for index := range members {
		if members[index].Name == "review" {
			members[index].Package.Commit = sourcelock.Commit{ObjectFormat: "sha1", Hex: member.Package.Commit.Hex[:40]}
		}
	}
	forged, err := sourcelock.New(plan.Lock.ManifestSHA256, members)
	if err != nil {
		t.Fatalf("forge valid-width lock with mismatched repository format: %v", err)
	}
	lockPath := sourcelock.PathIn(project)
	if err := sourcelock.Write(lockPath, forged); err != nil {
		t.Fatal(err)
	}
	lockBefore, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := sourcelock.NewBindings(forged.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: work}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(install.DraftBindingsPath(home, project), bindings); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(snapshot.Dir(home, member.Package.Repository, member.Package.Commit.Hex)); err != nil {
		t.Fatal(err)
	}
	result := draftInstall(t, project, home, install.Options{Platform: draftPlatform()})
	joined := strings.Join(result.Errors, "; ")
	if result.Status != "failed" || !strings.Contains(joined, "source_snapshot_changed") || !strings.Contains(joined, "object format") {
		t.Fatalf("install = %+v, want locked object-format refusal", result)
	}
	lockAfter, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(lockBefore) != string(lockAfter) {
		t.Fatal("object-format refusal rewrote the lock")
	}
}
