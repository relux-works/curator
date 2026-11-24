package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/gitops"
)

func TestProfileSourceCloneAndRefreshStayShallow(t *testing.T) {
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"README.md": "first\n"}, "v1.0.0")
	operand := ids.serve(repo, "https://example.com/team/shallow.git")
	manager := newGitManager(home)
	source, err := manager.Identity("skill", "shallow", operand)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := manager.ensureRepo(source)
	if err != nil {
		t.Fatal(err)
	}
	assertShallowProfileSource(t, dir)
	if count := fixtureGitOutput(t, dir, "rev-list", "--count", "HEAD"); count != "1" {
		t.Fatalf("initial history has %s commits, want 1", count)
	}
	writeGitFile(t, repo, "README.md", "second\n")
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", "second")
	gitRun(t, repo, "tag", "v1.1.0")
	if err := manager.fetch(source); err != nil {
		t.Fatal(err)
	}
	assertShallowProfileSource(t, dir)
	branch := fixtureGitOutput(t, repo, "symbolic-ref", "--short", "HEAD")
	if count := fixtureGitOutput(t, dir, "rev-list", "--count", "refs/remotes/origin/"+branch); count != "1" {
		t.Fatalf("refreshed history has %s commits, want 1", count)
	}
	if !cacheHasTag(dir, "v1.1.0") {
		t.Fatal("shallow refresh did not retain the newly published tag")
	}
}

func TestProfileSourceInitialCloneRetainsNonDefaultBranchTag(t *testing.T) {
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"README.md": "default branch\n"}, "v1.0.0")
	defaultBranch := fixtureGitOutput(t, repo, "symbolic-ref", "--short", "HEAD")
	defaultPin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	gitRun(t, repo, "checkout", "-b", "release")
	writeGitFile(t, repo, "README.md", "release branch\n")
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", "release")
	releasePin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	gitRun(t, repo, "tag", "v2.0.0")
	gitRun(t, repo, "checkout", defaultBranch)

	operand := ids.serve(repo, "https://example.com/team/side-tag.git")
	manager := newGitManager(home)
	source, err := manager.Identity("skill", "side-tag", operand)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := manager.Candidates("", "side-tag", source)
	if err != nil {
		t.Fatal(err)
	}
	foundDefault, foundRelease := false, false
	for _, candidate := range candidates {
		switch candidate.Tag {
		case "v1.0.0":
			foundDefault = candidate.Commit == defaultPin
		case "v2.0.0":
			foundRelease = candidate.Commit == releasePin
		}
	}
	if !foundDefault || !foundRelease {
		t.Fatalf("default and release tags were not both candidates (default=%t release=%t): %+v", foundDefault, foundRelease, candidates)
	}
	resolved, err := manager.ResolveTag("", "side-tag", source, "v2.0.0")
	if err != nil || resolved != releasePin {
		t.Fatalf("resolve side-branch tag = %q, %v; want %s", resolved, err, releasePin)
	}
	dir, err := manager.ensureRepo(source)
	if err != nil {
		t.Fatal(err)
	}
	assertShallowProfileSource(t, dir)
	if got := fixtureGitOutput(t, dir, "rev-parse", "--verify", "refs/remotes/origin/release"); got != releasePin {
		t.Fatalf("release head = %s, want %s", got, releasePin)
	}
	if count := fixtureGitOutput(t, dir, "rev-list", "--count", "refs/remotes/origin/release"); count != "1" {
		t.Fatalf("release history has %s commits, want one shallow tip", count)
	}
}

func TestProfileSourceCompactsFullCloneAndKeepsInstalledPin(t *testing.T) {
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"README.md": "pinned\n"}, "v1.0.0")
	oldPin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	gitRun(t, repo, "tag", "-d", "v1.0.0")
	writeGitFile(t, repo, "README.md", "current\n")
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", "current")
	operand := ids.serve(repo, "https://example.com/team/compact.git")
	manager := newGitManager(home)
	source, err := manager.Identity("skill", "compact", operand)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(manager.repoDir(source)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := gitops.Clone(operand, manager.repoDir(source)); err != nil {
		t.Fatal(err)
	}
	if count := fixtureGitOutput(t, manager.repoDir(source), "rev-list", "--count", "HEAD"); count != "2" {
		t.Fatalf("legacy clone history has %s commits, want 2", count)
	}
	lockPath := filepath.Join(home, "profiles", "p", "lock.json")
	_, err = contextlock.Write(lockPath, &contextlock.Lock{
		SchemaVersion: 2,
		HashVersion:   2,
		Root:          "profile",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "profile", Version: "1.0.0", StateHash: strings.Repeat("a", 64), RequiredBy: []string{}},
			{Kind: contextlock.KindSkill, Name: "compact", Source: source, Commit: oldPin, RequiredBy: []string{}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	dir, err := manager.ensureRepo(source)
	if err != nil {
		t.Fatalf("compact existing source: %v", err)
	}
	assertShallowProfileSource(t, dir)
	if got := fixtureGitOutput(t, dir, "rev-parse", "--verify", oldPin+"^{commit}"); got != oldPin {
		t.Fatalf("installed pin %s was not preserved: %s", oldPin, got)
	}
	gitRun(t, dir, "gc", "--prune=now")
	if _, err := gitops.Resolve(dir, "revision", oldPin); err != nil {
		t.Fatalf("git gc pruned installed pin %s: %v", oldPin, err)
	}
	if got := fixtureGitOutput(t, dir, "rev-parse", "--verify", profilePinRef(oldPin)+"^{commit}"); got != oldPin {
		t.Fatalf("git gc lost installed pin ref: %s", got)
	}
	if got := fixtureGitOutput(t, dir, "config", "--get", "remote.origin.url"); got != operand {
		t.Fatalf("compaction changed stored origin to %q", got)
	}
}

func TestProfileSourceCompactsCommitOnlyInOldCacheAndPinsIt(t *testing.T) {
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"README.md": "old pin\n"}, "v1.0.0")
	defaultBranch := fixtureGitOutput(t, repo, "symbolic-ref", "--short", "HEAD")
	oldPin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	operand := ids.serve(repo, "https://example.com/team/local-only.git")
	manager := newGitManager(home)
	source, err := manager.Identity("skill", "local-only", operand)
	if err != nil {
		t.Fatal(err)
	}
	oldCache := manager.repoDir(source)
	if err := os.MkdirAll(filepath.Dir(oldCache), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := gitops.Clone(operand, oldCache); err != nil {
		t.Fatal(err)
	}
	refs := strings.Fields(fixtureGitOutput(t, oldCache, "for-each-ref", "--format=%(refname)"))
	if len(refs) == 0 {
		t.Fatal("full source cache has no refs to remove for the local-object control")
	}
	for _, ref := range refs {
		gitRun(t, oldCache, "update-ref", "-d", ref)
	}
	if got := fixtureGitOutput(t, oldCache, "for-each-ref", "--format=%(refname)"); got != "" {
		t.Fatalf("old cache still has refs after test setup: %s", got)
	}
	if got := fixtureGitOutput(t, oldCache, "rev-parse", "--verify", oldPin+"^{commit}"); got != oldPin {
		t.Fatalf("old cache lost local-only pin during setup: %s", got)
	}

	gitRun(t, repo, "checkout", "--orphan", "replacement")
	gitRun(t, repo, "rm", "-rf", ".")
	writeGitFile(t, repo, "README.md", "replacement root\n")
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", "replacement root")
	gitRun(t, repo, "branch", "-D", defaultBranch)
	gitRun(t, repo, "branch", "-m", defaultBranch)
	gitRun(t, repo, "tag", "-d", "v1.0.0")
	remoteRefs := fixtureGitOutput(t, repo, "ls-remote", operand)
	if strings.Contains(remoteRefs, oldPin) {
		t.Fatalf("server still advertises old pin %s:\n%s", oldPin, remoteRefs)
	}
	if _, err := gitops.Resolve(repo, "revision", oldPin); err != nil {
		t.Fatalf("fixture server should retain the unadvertised object for this control: %v", err)
	}

	writeProfileSourcePinLock(t, home, source, oldPin)
	dir, err := manager.ensureRepo(source)
	if err != nil {
		t.Fatalf("compact using local-only old-cache pin: %v", err)
	}
	assertShallowProfileSource(t, dir)
	if got := fixtureGitOutput(t, dir, "rev-parse", "--verify", oldPin+"^{commit}"); got != oldPin {
		t.Fatalf("local-only pin %s was not copied: %s", oldPin, got)
	}
	if got := fixtureGitOutput(t, dir, "rev-parse", "--verify", profilePinRef(oldPin)+"^{commit}"); got != oldPin {
		t.Fatalf("local-only pin ref = %s, want %s", got, oldPin)
	}
	gitRun(t, dir, "gc", "--prune=now")
	if got := fixtureGitOutput(t, dir, "rev-parse", "--verify", oldPin+"^{commit}"); got != oldPin {
		t.Fatalf("git gc pruned local-only installed pin %s: %s", oldPin, got)
	}
}

func TestProfileSourceCompactionRefusesAnAbsentPin(t *testing.T) {
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"README.md": "available\n"}, "v1.0.0")
	operand := ids.serve(repo, "https://example.com/team/absent-pin.git")
	manager := newGitManager(home)
	source, err := manager.Identity("skill", "absent-pin", operand)
	if err != nil {
		t.Fatal(err)
	}
	oldCache := manager.repoDir(source)
	if err := os.MkdirAll(filepath.Dir(oldCache), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := gitops.Clone(operand, oldCache); err != nil {
		t.Fatal(err)
	}
	oldHead := fixtureGitOutput(t, oldCache, "rev-parse", "HEAD")
	writeProfileSourcePinLock(t, home, source, strings.Repeat("f", 40))
	if _, err := manager.ensureRepo(source); err == nil || !strings.Contains(err.Error(), "preserve installed profile pin") {
		t.Fatalf("compaction error = %v, want refusal to preserve absent pin", err)
	}
	if got := fixtureGitOutput(t, oldCache, "rev-parse", "HEAD"); got != oldHead {
		t.Fatalf("failed compaction changed old cache head from %s to %s", oldHead, got)
	}
	if shallow, err := gitops.IsProfileSourceClone(oldCache); err != nil || shallow {
		t.Fatalf("failed compaction replaced full cache: shallow=%t err=%v", shallow, err)
	}
	if got := fixtureGitOutput(t, oldCache, "for-each-ref", "--format=%(refname)", "refs/curator/profile-pins"); got != "" {
		t.Fatalf("absent pin created a durable ref: %s", got)
	}
}

func writeProfileSourcePinLock(t *testing.T, home, source, pin string) {
	t.Helper()
	lockPath := filepath.Join(home, "profiles", "p", "lock.json")
	_, err := contextlock.Write(lockPath, &contextlock.Lock{
		SchemaVersion: 2,
		HashVersion:   2,
		Root:          "profile",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "profile", Version: "1.0.0", StateHash: strings.Repeat("a", 64), RequiredBy: []string{}},
			{Kind: contextlock.KindSkill, Name: "compact", Source: source, Commit: pin, RequiredBy: []string{}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func profilePinRef(commit string) string {
	return "refs/curator/profile-pins/" + commit
}

func TestProfileSourceStillRefusesAnUnknownExactPin(t *testing.T) {
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"README.md": "available\n"}, "v1.0.0")
	operand := ids.serve(repo, "https://example.com/team/pin.git")
	manager := newGitManager(home)
	manager.missingPins = map[string]bool{}
	source, err := manager.Identity("skill", "pin", operand)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := manager.ensureRepo(source)
	if err != nil {
		t.Fatal(err)
	}
	missing := strings.Repeat("f", 40)
	before := fixtureGitOutput(t, dir, "show-ref")
	err = manager.acquireMissingPin(dir, source, missing)
	if err == nil || !strings.Contains(err.Error(), "fetch exact pin") {
		t.Fatalf("unknown exact pin error = %v", err)
	}
	if after := fixtureGitOutput(t, dir, "show-ref"); after != before {
		t.Fatalf("unknown pin changed refs:\n%s\n->\n%s", before, after)
	}
}

func assertShallowProfileSource(t *testing.T, repo string) {
	t.Helper()
	shallow, err := gitops.IsProfileSourceClone(repo)
	if err != nil || !shallow {
		t.Fatalf("profile source shallow=%t, err=%v", shallow, err)
	}
}
