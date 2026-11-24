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
	if got := fixtureGitOutput(t, dir, "config", "--get", "remote.origin.url"); got != operand {
		t.Fatalf("compaction changed stored origin to %q", got)
	}
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
