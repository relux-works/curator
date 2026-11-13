package gitops

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testgitenv"
)

// localSiblingPair creates, under one physical root, a bare repository named
// intended (whose name carries a significant byte) and a bare sibling named
// sibling, each with its own witness tag, plus an empty repository to fetch
// into. It skips a row whose two names git on this host cannot create as
// two distinct repositories (testgitenv.RequireGitRepositoryNames); an
// ordinary-name row never skips.
func localSiblingPair(t *testing.T, intended, sibling string) (root, target string) {
	t.Helper()
	source := makeRepo(t)
	root, err := filepath.EvalSymlinks(cloneTestRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	testgitenv.RequireGitRepositoryNames(t, root, intended, sibling)
	for name, tag := range map[string]string{intended: "fresh-intended", sibling: "only-in-trimmed-sibling"} {
		bare := filepath.Join(root, name)
		gitRun(t, root, "clone", "-q", "--bare", source, bare)
		gitRun(t, bare, "tag", tag)
	}
	target = filepath.Join(root, "target")
	gitRun(t, root, "init", "-q", target)
	return root, target
}

// Claim: FetchLocalPathIsolated hands git the absolute path byte for byte:
// for a path ending in a space, tab or newline, or carrying a leading space
// or newline inside a component, the named repository's witness tag arrives
// and the sibling's never does. The first three are the cases trimming would
// turn into the sibling; the next two prove no other component rewrite;
// the percent rows prove a literal "%" and a "%20" are neither decoded nor
// refused. Control: an ordinary path with a distinct sibling fetches the
// named one. Rows this host's git cannot keep distinct skip with the probe's
// reason; the control never skips.
func TestFetchLocalPathIsolatedKeepsEveryByte(t *testing.T) {
	for _, testCase := range []struct{ name, intended, sibling string }{
		{"final space", "origin.git ", "origin.git"},
		{"final tab", "origin.git\t", "origin.git"},
		{"final newline", "origin.git\n", "origin.git"},
		{"leading space in component", " origin.git", "origin.git"},
		{"embedded newline", "ori\ngin.git", "origin.git"},
		{"literal percent", "origin%marker.git", "origin.git"},
		{"percent escape", "origin%20x.git", "origin x.git"},
		{"ordinary path control", "origin.git", "other.git"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root, target := localSiblingPair(t, testCase.intended, testCase.sibling)
			if err := FetchLocalPathIsolated(target, filepath.Join(root, testCase.intended)); err != nil {
				t.Fatal(err)
			}
			tags := gitRun(t, target, "tag", "-l")
			if !strings.Contains(tags, "fresh-intended") || strings.Contains(tags, "only-in-trimmed-sibling") {
				t.Fatalf("tags %q: want the named repository's tag and not the sibling's", tags)
			}
		})
	}
}

// Claim: FetchLocalPathIsolated refuses an empty or relative path (git could
// read a relative word as a configured remote name) before running git: the
// target gains no tag even though a repository of that relative name exists
// beside it. The fixture uses ordinary names only, so this runs on every
// platform; "origin.git " is refused as a spelling before git runs.
func TestFetchLocalPathIsolatedRefusesRelativePaths(t *testing.T) {
	root, target := localSiblingPair(t, "origin.git", "other.git")
	t.Chdir(root)
	for _, path := range []string{"", "origin.git", "origin.git ", "./origin.git", "../" + filepath.Base(root) + "/origin.git"} {
		if err := FetchLocalPathIsolated(target, path); err == nil || !strings.Contains(err.Error(), "an absolute filesystem path is required") {
			t.Fatalf("FetchLocalPathIsolated(%q) = %v, want the absolute-path refusal", path, err)
		}
	}
	if tags := gitRun(t, target, "tag", "-l"); tags != "" {
		t.Fatalf("a refused path was fetched: tags %q", tags)
	}
}

// Claim: OriginURL returns the stored origin byte for byte, leading and
// trailing whitespace included (the binding refuses such spellings, so it
// must see them); an ordinary value is returned unchanged.
func TestOriginURLKeepsSignificantWhitespace(t *testing.T) {
	repo := makeRepo(t)
	for _, value := range []string{" /srv/origin.git", "/srv/origin.git ", "\t/srv/origin.git\n", "/srv/origin.git"} {
		gitRun(t, repo, "config", "remote.origin.url", value)
		got, present, err := OriginURL(repo)
		if err != nil || !present || got != value {
			t.Fatalf("OriginURL = %q, %v, %v; want %q", got, present, err, value)
		}
	}
}
