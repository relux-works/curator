package gitops

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: RedirectTarget extracts the location from git's untranslated
// "warning: redirecting to <url>" line among other stderr lines (LF or CRLF),
// without its trailing slash, and reports nothing for stderr without that line
// or with an empty location.
func TestRedirectTargetClassifiesGitStderr(t *testing.T) {
	cases := []struct {
		name, stderr, target string
		ok                   bool
	}{
		{"git warning", "warning: redirecting to https://git.example.com/skills/renamed.git/\n", "https://git.example.com/skills/renamed.git", true},
		{"among progress lines", "Cloning into 'x'...\r\nwarning: redirecting to https://h.example/a/b/\r\nremote: done\r\n", "https://h.example/a/b", true},
		{"no warning", "Cloning into 'x'...\nwarning: some other warning\n", "", false},
		{"empty stderr", "", "", false},
		{"no location", "warning: redirecting to \n", "", false},
		{"not at line start", "remote: warning: redirecting to https://h.example/a\n", "", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target, ok := RedirectTarget(testCase.stderr)
			if ok != testCase.ok || target != testCase.target {
				t.Fatalf("RedirectTarget = %q, %v; want %q, %v", target, ok, testCase.target, testCase.ok)
			}
		})
	}
}

// fakeRedirectGit puts the recording fake git first on PATH with the given
// success stderr and returns its marker file.
func fakeRedirectGit(t *testing.T, stderr string) string {
	t.Helper()
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "git-ran")
	fakeTool(t, bin, "git")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GIT_MARKER", marker)
	t.Setenv("GIT_FAKE_STDERR", stderr)
	t.Setenv("LC_ALL", "de_DE.UTF-8")
	return marker
}

// Claim: a successful clone and fetch whose git stderr carries the redirect
// warning return *RedirectError naming the endpoint and target, the redirected
// clone's destination is removed, and git runs with the message locale pinned
// to C despite an ambient LC_ALL. Limit: git is a fake that prints the
// warning; no HTTP server redirects here.
func TestRefusingRedirectLanesReadSuccessStderr(t *testing.T) {
	marker := fakeRedirectGit(t, "Cloning into 'x'...\nwarning: redirecting to https://git.example.com/skills/renamed.git/\n")
	endpoint := "https://git.example.com/skills/tool.git"
	want := RedirectError{Endpoint: endpoint, Target: "https://git.example.com/skills/renamed.git"}

	destination := filepath.Join(t.TempDir(), "clone")
	var redirect *RedirectError
	if err := CloneRefusingRedirect(endpoint, destination); !errors.As(err, &redirect) || *redirect != want {
		t.Fatalf("clone err = %v, want %+v", err, want)
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("a redirected clone left its destination: %v", err)
	}

	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	redirect = nil
	if err := FetchURLRefusingRedirect(repo, endpoint); !errors.As(err, &redirect) || *redirect != want {
		t.Fatalf("fetch err = %v, want %+v", err, want)
	}
	recorded, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(recorded); got != "LC_ALL=C\nLC_ALL=C\n" {
		t.Fatalf("git invocations recorded %q, want two with LC_ALL=C", got)
	}
}

// Claim: the same successful clone and fetch without a redirect warning in
// stderr succeed and keep the clone (positive control for the test above).
func TestRefusingRedirectLanesAcceptUnredirectedSuccess(t *testing.T) {
	fakeRedirectGit(t, "Cloning into 'x'...\nwarning: some other warning\n")
	endpoint := "https://git.example.com/skills/tool.git"
	destination := filepath.Join(t.TempDir(), "clone")
	if err := CloneRefusingRedirect(endpoint, destination); err != nil {
		t.Fatalf("clone err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destination, ".git")); err != nil {
		t.Fatalf("an unredirected clone was removed: %v", err)
	}
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := FetchURLRefusingRedirect(repo, endpoint); err != nil {
		t.Fatalf("fetch err = %v", err)
	}
}

// Claim: FetchURLRefusingRedirect keeps FetchURL's endpoint contract: a local
// path, a file URL and a credential-bearing URL refuse with the fixed error
// before git runs.
func TestFetchURLRefusingRedirectKeepsEndpointContract(t *testing.T) {
	marker := fakeRedirectGit(t, "")
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"/srv/skills/tool", "file:///srv/skills/tool", "https://user:secret@git.example.com/skills/tool.git"} {
		if err := FetchURLRefusingRedirect(repo, endpoint); err != errFetchEndpoint {
			t.Fatalf("FetchURLRefusingRedirect(%q) = %v, want the fixed refusal", endpoint, err)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("git ran for a refused endpoint: %v", err)
	}
}

// Claim: a real local clone through CloneRefusingRedirect succeeds and
// resolves like Clone, so the clone path is unchanged when git reports no
// redirect.
func TestCloneRefusingRedirectClonesLocally(t *testing.T) {
	source := makeRepo(t)
	destination := filepath.Join(cloneTestRoot(t), "clone")
	if err := CloneRefusingRedirect(source, destination); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(destination, "tag", "v1"); err != nil {
		t.Fatal(err)
	}
}

// Claim: OriginURL reads the stored origin without rewriting: a configured
// origin is returned verbatim even when an insteadOf rule would rewrite it; a
// checkout without an origin (other remotes notwithstanding) reports absent;
// several or empty origin URLs are refused.
func TestOriginURLReadsStoredConfiguration(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q")
	if got, present, err := OriginURL(repo); err != nil || present || got != "" {
		t.Fatalf("no remotes: %q %v %v", got, present, err)
	}
	gitRun(t, repo, "remote", "add", "other", "https://git.example.com/skills/other.git")
	if _, present, err := OriginURL(repo); err != nil || present {
		t.Fatalf("non-origin remote read as origin: %v %v", present, err)
	}
	origin := "https://git.example.com/skills/tool.git"
	gitRun(t, repo, "remote", "add", "origin", origin)
	gitRun(t, repo, "config", "url.https://evil.example.com/.insteadOf", "https://git.example.com/")
	if got, present, err := OriginURL(repo); err != nil || !present || got != origin {
		t.Fatalf("origin = %q %v %v, want %q verbatim", got, present, err, origin)
	}
	gitRun(t, repo, "config", "--add", "remote.origin.url", "https://git.example.com/skills/second.git")
	if _, _, err := OriginURL(repo); err == nil || !strings.Contains(err.Error(), "2 origin URLs") {
		t.Fatalf("several origin URLs: err = %v", err)
	}
	gitRun(t, repo, "config", "--unset-all", "remote.origin.url")
	gitRun(t, repo, "config", "remote.origin.url", "")
	if _, _, err := OriginURL(repo); err == nil || !strings.Contains(err.Error(), "empty origin URL") {
		t.Fatalf("empty origin URL: err = %v", err)
	}
}
