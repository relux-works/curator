package gitops

import (
	"path/filepath"
	"testing"
)

// The shared production Clone and Fetch entry points must ignore inherited
// rewrite/core.sshCommand configuration; file transport is a valid sibling.
// Actual Git objects distinguish declared and redirected repositories.
func TestDeclaredCloneFetchIgnoreInheritedRemapping(t *testing.T) {
	declared, evil := makeRepo(t), makeRepo(t)
	gitRun(t, evil, "commit", "--allow-empty", "-m", "evil")
	endpoint := fileURL(t, declared)
	hostile := hostileGitConfig(t, endpoint, fileURL(t, evil))
	t.Setenv("GIT_CONFIG_GLOBAL", hostile)
	t.Setenv("GIT_CONFIG_SYSTEM", hostile)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
	dest := filepath.Join(t.TempDir(), "cache")
	if err := Clone(endpoint, dest); err != nil {
		t.Fatal(err)
	}
	if got, want := gitRun(t, dest, "rev-parse", "HEAD"), gitRun(t, declared, "rev-parse", "HEAD"); got != want {
		t.Fatalf("clone remapped: %s != %s", got, want)
	}
	gitRun(t, declared, "commit", "--allow-empty", "-m", "new declared")
	if err := Fetch(dest); err != nil {
		t.Fatal(err)
	}
	branch := gitRun(t, declared, "symbolic-ref", "--short", "HEAD")
	if got, want := gitRun(t, dest, "rev-parse", "refs/remotes/origin/"+branch), gitRun(t, declared, "rev-parse", "HEAD"); got != want {
		t.Fatalf("fetch remapped: %s != %s", got, want)
	}
}

// Fetch is tested independently of Clone so a remapped clone cannot hide a
// fetch-only regression. The declared and redirected histories differ.
func TestDeclaredFetchIgnoreInheritedRemapping(t *testing.T) {
	declared, evil := makeRepo(t), makeRepo(t)
	endpoint := fileURL(t, declared)
	dest := filepath.Join(t.TempDir(), "cache")
	if err := Clone(endpoint, dest); err != nil {
		t.Fatal(err)
	}
	gitRun(t, declared, "commit", "--allow-empty", "-m", "new declared")
	gitRun(t, evil, "commit", "--allow-empty", "-m", "evil")
	hostile := hostileGitConfig(t, endpoint, fileURL(t, evil))
	t.Setenv("GIT_CONFIG_GLOBAL", hostile)
	if err := Fetch(dest); err != nil {
		t.Fatal(err)
	}
	branch := gitRun(t, declared, "symbolic-ref", "--short", "HEAD")
	if got, want := gitRun(t, dest, "rev-parse", "refs/remotes/origin/"+branch), gitRun(t, declared, "rev-parse", "HEAD"); got != want {
		t.Fatalf("fetch remapped: %s != %s", got, want)
	}
}

// Fetch keeps its existing all-remotes contract: an independent configured
// remote is refreshed after another fails, while the overall result is red.
// This is not generated transport fallback for a declared endpoint.
func TestFetchRetainsAllRemoteRefreshOnFailure(t *testing.T) {
	repo, valid := makeRepo(t), makeRepo(t)
	gitRun(t, valid, "commit", "--allow-empty", "-m", "independent remote")
	gitRun(t, repo, "remote", "add", "a-broken", fileURL(t, filepath.Join(t.TempDir(), "absent")))
	gitRun(t, repo, "remote", "add", "z-valid", fileURL(t, valid))
	if err := Fetch(repo); err == nil {
		t.Fatal("failed remote was swallowed")
	}
	branch := gitRun(t, valid, "symbolic-ref", "--short", "HEAD")
	if got, want := gitRun(t, repo, "rev-parse", "refs/remotes/z-valid/"+branch), gitRun(t, valid, "rev-parse", "HEAD"); got != want {
		t.Fatalf("later configured remote not refreshed: %s != %s", got, want)
	}
}
