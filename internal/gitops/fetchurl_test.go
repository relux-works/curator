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
