package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
)

const (
	cacheOriginHTTPS = "https://example.com/team/profiles.git"
	cacheOriginSSH   = "git@example.com:team/profiles.git"
)

// staleOriginCache creates a profile cache through the HTTPS spelling, then
// models the reported machine: the HTTPS transport stops working, the SSH
// spelling serves the same repository, and the repository gains tag v1.1.0.
func staleOriginCache(t *testing.T) (home, canonical, dir string, ids *gitIdentities, repo string) {
	t.Helper()
	home = gitFixtureHome(t)
	ids = newGitIdentities(t)
	repo = gitRepo(t, map[string]string{"README": "one\n"}, "v1.0.0")
	ids.serve(repo, cacheOriginHTTPS)
	first := newGitManager(home)
	canonical, err := first.Identity("", "", cacheOriginHTTPS)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.fetch(canonical); err != nil {
		t.Fatalf("create the cache through HTTPS: %v", err)
	}
	dir = first.repoDir(canonical)
	if err := os.WriteFile(ids.path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ids.serve(filepath.Join(t.TempDir(), "unreachable"), cacheOriginHTTPS)
	ids.serve(repo, cacheOriginSSH)
	gitRun(t, repo, "commit", "--allow-empty", "-m", "two")
	gitRun(t, repo, "tag", "v1.1.0")
	return home, canonical, dir, ids, repo
}

func cacheHasTag(dir, tag string) bool {
	return exec.Command("git", "-C", dir, "rev-parse", "--verify", "--quiet", "refs/tags/"+tag).Run() == nil // #nosec G204 -- fixed test argv
}

// The reported shape, witnessed in an isolated Curator home: an HTTPS-origin
// cache whose transport no longer works, and an operation naming the SSH
// descriptor of the same repository. The refresh comes from the SSH
// descriptor and finds the new tag, and the stored origin keeps its HTTPS
// spelling. The control shows that the stored-origin refresh fails on the
// same cache, so the fixture models the defect rather than a working origin.
func TestProfileCacheRefreshesFromTheRequestedDescriptor(t *testing.T) {
	home, canonical, dir, _, _ := staleOriginCache(t)
	if err := gitops.Fetch(dir); err == nil {
		t.Fatal("control: the stored-origin refresh succeeded, so the fixture no longer models the stale origin")
	}
	second := newGitManager(home)
	if got, err := second.Identity("", "", cacheOriginSSH); err != nil || got != canonical {
		t.Fatalf("SSH descriptor identity = %q, %v; want %q", got, err, canonical)
	}
	if err := second.fetch(canonical); err != nil {
		t.Fatalf("refresh from the requested descriptor: %v", err)
	}
	if !cacheHasTag(dir, "v1.1.0") {
		t.Fatal("the refresh did not reach the descriptor's new tag")
	}
	if origin := fixtureGitOutput(t, dir, "config", "--get", "remote.origin.url"); origin != cacheOriginHTTPS {
		t.Fatalf("stored origin rewritten to %q", origin)
	}
}

// A cache whose stored origin names another repository, or has no single
// origin, refuses before any fetch: the new tag never arrives and the
// configuration is unchanged. The SSH descriptor stays the same, so the
// refusal is the cache binding's, not the transport's.
func TestProfileCacheRefusesAForeignOrUnboundOrigin(t *testing.T) {
	for _, c := range []struct {
		name  string
		setup func(t *testing.T, dir string)
		want  string
	}{
		{"foreign", func(t *testing.T, dir string) {
			gitRun(t, dir, "config", "remote.origin.url", "https://example.com/other/profiles.git")
		}, "origin differs"},
		{"missing", func(t *testing.T, dir string) { gitRun(t, dir, "remote", "remove", "origin") }, "no single stored origin"},
		{"repeated", func(t *testing.T, dir string) {
			gitRun(t, dir, "config", "--add", "remote.origin.url", cacheOriginSSH)
		}, "no single stored origin"},
	} {
		t.Run(c.name, func(t *testing.T) {
			home, canonical, dir, _, _ := staleOriginCache(t)
			c.setup(t, dir)
			before := fixtureGitOutput(t, dir, "config", "--local", "--list")
			second := newGitManager(home)
			if _, err := second.Identity("", "", cacheOriginSSH); err != nil {
				t.Fatal(err)
			}
			err := second.fetch(canonical)
			if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("fetch err = %v, want %s %q", err, DiagSourceInvalid, c.want)
			}
			if cacheHasTag(dir, "v1.1.0") {
				t.Fatal("a refused cache was fetched")
			}
			if after := fixtureGitOutput(t, dir, "config", "--local", "--list"); after != before {
				t.Fatalf("refusal changed the cache configuration:\n%s\n->\n%s", before, after)
			}
		})
	}
}

// An identity this operation recorded no spelling for still refreshes
// through the cache's stored remotes, exactly as before: with the HTTPS
// transport restored it succeeds, and the descriptor path is not taken.
func TestProfileCacheWithoutARecordedSpellingUsesItsStoredOrigin(t *testing.T) {
	home, canonical, dir, ids, repo := staleOriginCache(t)
	if err := os.WriteFile(ids.path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ids.serve(repo, cacheOriginHTTPS)
	if err := newGitManager(home).fetch(canonical); err != nil {
		t.Fatalf("stored-origin refresh: %v", err)
	}
	if !cacheHasTag(dir, "v1.1.0") {
		t.Fatal("the stored-origin refresh did not reach the new tag")
	}
}
