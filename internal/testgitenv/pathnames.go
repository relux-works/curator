package testgitenv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// portableName is a repository name every supported platform and git build
// must represent: an ordinary positive control is never skipped.
var portableName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// GitRepositoryNamesDistinct reports whether git on this host can create a
// bare repository at each of names directly inside dir and keep them apart:
// every `git init --bare` must succeed, the directory must then list exactly
// those names, byte for byte (so no name was stripped, folded or merged into
// another on disk), and git must read back from each name the distinct mark
// written through it (so git, too, reaches a different repository per name).
// An os.Mkdir probe proves none of this: Go can create a name through a
// verbatim path that Git for Windows then refuses or reads as its stripped
// sibling. The probe works in a private subdirectory of dir, on the same
// filesystem, and removes it. detail explains a false result in one line;
// err reports a probe that could not run at all.
func GitRepositoryNamesDistinct(dir string, names ...string) (ok bool, detail string, err error) {
	probe, err := os.MkdirTemp(dir, "git-name-probe-")
	if err != nil {
		return false, "", err
	}
	defer func() { _ = os.RemoveAll(probe) }()
	for index, name := range names {
		if name == "" || strings.ContainsAny(name, `/`+string(filepath.Separator)) {
			return false, "", fmt.Errorf("repository name %q is not a single path component", name)
		}
		path := filepath.Join(probe, name)
		if output, runErr := probeGit("init", "-q", "--bare", path); runErr != nil {
			return false, fmt.Sprintf("git init --bare %q failed: %v: %s", name, runErr, oneLine(output)), nil
		}
		if output, runErr := probeGit("--git-dir="+path, "config", "curator.probe", strconv.Itoa(index)); runErr != nil {
			return false, fmt.Sprintf("git config in %q failed: %v: %s", name, runErr, oneLine(output)), nil
		}
	}
	entries, err := os.ReadDir(probe)
	if err != nil {
		return false, "", err
	}
	listed := make([]string, 0, len(entries))
	for _, entry := range entries {
		listed = append(listed, entry.Name())
	}
	wanted := append([]string(nil), names...)
	sort.Strings(listed)
	sort.Strings(wanted)
	if strings.Join(listed, "\x00") != strings.Join(wanted, "\x00") {
		return false, fmt.Sprintf("the filesystem lists %q after creating %q", listed, wanted), nil
	}
	for index, name := range names {
		output, runErr := probeGit("--git-dir="+filepath.Join(probe, name), "config", "--get", "curator.probe")
		if runErr != nil {
			return false, fmt.Sprintf("git cannot read repository %q: %v: %s", name, runErr, oneLine(output)), nil
		}
		if got := strings.TrimSpace(output); got != strconv.Itoa(index) {
			return false, fmt.Sprintf("git reads repository %q and finds the mark %q written for another name", name, got), nil
		}
	}
	return true, "", nil
}

// RequireGitRepositoryNames runs GitRepositoryNamesDistinct before a
// byte-sensitive fixture is built in dir. A row whose names this host and
// its git cannot keep distinct is skipped with a one-line reason naming
// GOOS, the names and git's own complaint (CI's skip-class table accepts it
// as host-capability: "this host cannot create"). A row of portable names
// (letters, digits, '.', '_', '-') is an ordinary positive control and
// fails instead of skipping, as does a probe that cannot run.
func RequireGitRepositoryNames(t testing.TB, dir string, names ...string) {
	t.Helper()
	ok, detail, err := GitRepositoryNamesDistinct(dir, names...)
	if err != nil {
		t.Fatalf("git repository name probe for %q could not run: %v", names, err)
	}
	if ok {
		return
	}
	for _, name := range names {
		if !portableName.MatchString(name) {
			t.Skipf("this host cannot create git repositories named %q as distinct repositories (GOOS=%s): %s; this significant-byte row is skipped, its ordinary positive control runs", names, runtime.GOOS, detail)
			return
		}
	}
	t.Fatalf("ordinary positive control: git cannot keep repositories named %q distinct (GOOS=%s): %s", names, runtime.GOOS, detail)
}

func probeGit(args ...string) (string, error) {
	output, err := exec.Command("git", args...).CombinedOutput() // #nosec G204 -- test-only probe: fixed git executable, closed argument construction over fixture names.
	return string(output), err
}

// oneLine keeps a git complaint on the single line CI's skip-reason
// extraction reads.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
