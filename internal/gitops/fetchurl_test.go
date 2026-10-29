package gitops

import (
	"os"
	"path/filepath"
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
		if err := FetchURL(repo, url); err == nil || !strings.HasPrefix(err.Error(), "refusing to fetch") {
			t.Fatalf("FetchURL(%q) err = %v, want a refusal", url, err)
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

// FetchURL admits only the https, ssh and scp-like endpoint forms and refuses
// everything else before git runs, without echoing the input. A fake git on
// PATH records every invocation: each refused case leaves the marker empty and
// keeps the canary out of the error, and each admitted case reaches git. The
// refused cases cover remote-helper transports of every spelling, upper-case
// schemes, passwords and user information in every URL and scp form,
// percent-encoding, whitespace and newlines, empty and extra-slash authorities,
// queries, fragments, opaque and option-shaped spellings, http, git:// and
// file forms, a drive-letter host and a scheme-named host. A benign '@' in a
// path is refused as well: the grammar is deliberately narrower than what git
// accepts.
func TestFetchURLAdmitsOnlyTheEndpointGrammar(t *testing.T) {
	const c = "canary-secret-7f3a"
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "git-ran")
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\necho ran >> \"$GIT_MARKER\"\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GIT_MARKER", marker)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	refused := []string{
		"", " ", "-" + c, "--upload-pack=" + c,
		"https::https://user:" + c + "@example.invalid/team/repo.git",
		"http::https://example.invalid/" + c + ".git",
		"ext::sh -c " + c, "fd::7", "file::/tmp/" + c, "x::" + c, "1::" + c, "a_b::" + c,
		"HTTPS://example.invalid/" + c + ".git", "Https://example.invalid/" + c + ".git", "SSH://git@example.invalid/" + c + ".git",
		"https://user:" + c + "@example.invalid/team/repo.git",
		"https://" + c + "@example.invalid/team/repo.git",
		"https://user%3A" + c + "@example.invalid/team/repo.git",
		"http://user:" + c + "@example.invalid/team/repo.git",
		"http://example.invalid/" + c + ".git",
		"git://example.invalid/" + c + ".git",
		"git://user:" + c + "@example.invalid/team/repo.git",
		"file:///tmp/" + c + ".git",
		"ssh://git:" + c + "@example.invalid/team/repo.git",
		"ssh://git:" + c + "@[::1]/team/repo.git",
		"ssh://example.invalid/-" + c,
		"git:" + c + "@example.invalid:team/repo.git",
		"user:" + c + "@[::1]:team/repo.git",
		"https:opaque-" + c,
		"https:///user:" + c + "@example.invalid/team/repo.git",
		"https:////user:" + c + "@example.invalid/team/repo.git",
		"https://example.invalid/team/repo.git?password=" + c,
		"https://example.invalid/team/repo.git?access_token=" + c,
		"https://example.invalid/team/repo.git?next=https://user:" + c + "@example.invalid/",
		"https://example.invalid/team/repo.git#" + c,
		"https://example.invalid/team/%2F" + c,
		" https://example.invalid/" + c + ".git",
		"https://example.invalid/" + c + ".git ",
		"https://example.invalid/team/repo.git\n" + c,
		"https://example.invalid/team/r@" + c + ".git",
		"https://exa_mple.invalid/" + c + ".git",
		"https://example.invalid",
		"https://example.invalid/",
		"https://example.invalid:99999999/" + c + ".git",
		"c:" + c,
		"ssh:" + c,
		"https:" + c,
		"example.invalid:-" + c,
		"/tmp/" + c + ".git",
		"./" + c,
	}
	for _, url := range refused {
		err := FetchURL(repo, url)
		if err == nil || !strings.HasPrefix(err.Error(), "refusing to fetch") || strings.Contains(err.Error(), c) {
			t.Errorf("FetchURL(%q) err = %v, want an input-free refusal", url, err)
		}
		if _, statErr := os.Stat(marker); statErr == nil {
			t.Fatalf("FetchURL(%q) ran git before refusing", url)
		}
	}
	for _, url := range []string{
		"https://example.invalid/team/repo.git",
		"https://example.invalid:8443/team/repo.git",
		"https://[::1]/team/repo.git",
		"ssh://git@example.invalid/team/repo.git",
		"ssh://git@example.invalid:2222/team/repo.git",
		"ssh://example.invalid/team/repo.git",
		"ssh://git@[::1]/team/repo.git",
		"git@example.invalid:team/repo.git",
		"git@[::1]:team/repo.git",
		"github-work:team/repo.git",
		"example.invalid:/srv/team/repo.git",
	} {
		_ = os.Remove(marker)
		if err := FetchURL(repo, url); err != nil && strings.HasPrefix(err.Error(), "refusing to fetch") {
			t.Fatalf("FetchURL(%q) refused an admitted endpoint: %v", url, err)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("FetchURL(%q) did not reach git", url)
		}
	}
}
