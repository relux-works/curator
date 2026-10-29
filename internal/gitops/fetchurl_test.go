package gitops

import (
	"slices"
	"strings"
	"testing"
)

// FetchURL refuses an empty or option-shaped endpoint before running git, and
// the arguments it builds name the endpoint as a positional with the same
// head and tag refspecs as FetchURLIsolated: no configuration override, and
// so no credential helper, is added to the operator's transport.
func TestFetchURLArgs(t *testing.T) {
	repo := t.TempDir()
	for _, url := range []string{"", "   ", "--upload-pack=touch /tmp/x", "-c"} {
		if err := FetchURL(repo, url); err == nil || !strings.Contains(err.Error(), "suspicious") {
			t.Fatalf("FetchURL(%q) err = %v, want the suspicious-URL refusal", url, err)
		}
	}
	if _, err := fetchURLArgs(repo, "git@example.com:team/profiles.git"); err == nil {
		t.Fatal("a directory that is not a repository was accepted")
	}
	if _, err := run(repo, "init", "-q"); err != nil {
		t.Fatal(err)
	}
	args, err := fetchURLArgs(repo, " git@example.com:team/profiles.git ")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"fetch", "--force", "--prune", "--no-write-fetch-head", "git@example.com:team/profiles.git",
		"+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"}
	if !slices.Equal(args, want) {
		t.Fatalf("args = %q, want %q", args, want)
	}
	for _, arg := range args {
		if arg == "-c" || strings.Contains(arg, "credential") {
			t.Fatalf("args carry a configuration override: %q", args)
		}
	}
}

// FetchURL refuses an endpoint that carries credentials before running git,
// and the refusal never echoes the input: a password in any URL form, user
// information on HTTP(S), and a scp-like address with a password. The
// controls show an SSH user name in both spellings and a plain HTTPS URL pass
// the credential check and reach git (here a missing remote, so git fails
// with its own error rather than a credential refusal).
func TestFetchURLRefusesCredentialURLs(t *testing.T) {
	repo := t.TempDir()
	if _, err := run(repo, "init", "-q"); err != nil {
		t.Fatal(err)
	}
	const canary = "canary-secret-7f3a"
	for _, url := range []string{
		"https://user:" + canary + "@example.invalid/team/repo.git",
		"https://" + canary + "@example.invalid/team/repo.git",
		"http://user:" + canary + "@example.invalid/team/repo.git",
		"ssh://git:" + canary + "@example.invalid/team/repo.git",
		"git:" + canary + "@example.invalid:team/repo.git",
	} {
		err := FetchURL(repo, url)
		if err == nil || !strings.HasPrefix(err.Error(), "refusing to fetch") || strings.Contains(err.Error(), canary) {
			t.Fatalf("FetchURL(credential URL) err = %v, want a refusal that does not echo the input", err)
		}
	}
	for _, url := range []string{
		"git@example.invalid:team/repo.git",
		"ssh://git@example.invalid/team/repo.git",
		"https://example.invalid/team/repo.git",
	} {
		if err := refuseCredentialURL(url); err != nil {
			t.Fatalf("refuseCredentialURL(%q) = %v, want admitted", url, err)
		}
	}
	if err := FetchURL(repo, "file://"+t.TempDir()+"/missing.git"); err == nil || strings.HasPrefix(err.Error(), "refusing to fetch") {
		t.Fatalf("a credential-free endpoint did not reach git: %v", err)
	}
}
