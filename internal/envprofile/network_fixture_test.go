package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// gitFileURL renders a local fixture repository path as the file:// remote
// a git [url ...] section matches. The value is written into git
// configuration, where a backslash is an escape: a raw Windows temporary
// path (C:\Users\...) collapses (\U, \A, \T consumed) and the clone fails
// with "does not appear to be a git repository". Forward slashes parse
// identically on every platform, and the three-slash form keeps a Windows
// volume (file:///C:/...) out of the host position. This is the one helper
// every insteadOf fixture in this package routes through — never
// interpolate a raw t.TempDir() into a git config value.
func gitFileURL(path string) string {
	return gitFileURLFor(filepath.ToSlash(path), filepath.VolumeName(path))
}

// gitFileURLFor renders an already-slashed fixture path with its volume
// name as the file:// remote of a git [url ...] section. The split keeps
// the Windows branch unit-testable off Windows: filepath.VolumeName
// reports "" for every input on unix, so the volume branch only runs on
// Windows itself.
func gitFileURLFor(slashed, volume string) string {
	if volume != "" {
		return "file:///" + slashed
	}
	return "file://" + slashed
}

// TestMain pins the process configuration at paths the tests control, so the
// machine-policy loader (loadMachinePolicy) never observes the operator's
// real configuration: package tests are hermetic by construction. Tests that
// exercise the loader set CURATOR_CONFIG/CURATOR_SYSTEM_CONFIG to their own
// files with t.Setenv, which overrides this base for the test.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "curator-test-noconfig-*")
	if err == nil {
		// Never created: the loader maps an absent file to an empty policy.
		_ = os.Setenv("CURATOR_CONFIG", filepath.Join(dir, "config.json"))
		_ = os.Setenv("CURATOR_SYSTEM_CONFIG", "")
	}
	os.Exit(m.Run())
}

// gitIdentities serves local git repositories under fake canonical network
// identities for one test, through git insteadOf rewrites. A file:// remote
// is rejected at the identity boundary (see canonicalGit) and must never
// appear as an operand or a requirement source; these fixtures exercise the
// production network identity path — canonicalization, the source allowlist,
// revocation, and schema-shaped locks — hermetically and offline.
type gitIdentities struct {
	t    *testing.T
	path string
}

func newGitIdentities(t *testing.T) *gitIdentities {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", path)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	return &gitIdentities{t: t, path: path}
}

// TestGitFileURLForKeepsVolumesOutOfTheHost narrows the fixture-remote
// gate: a Windows volume stays in the path (file:///C:/...), never in the
// host position (file://C:/...), and a volumeless path keeps the two-slash
// form. A mutant that drops the volume branch (always "file://" + slashed)
// maps C:/... onto host C: and must fail this test on every platform.
func TestGitFileURLForKeepsVolumesOutOfTheHost(t *testing.T) {
	if got := gitFileURLFor("C:/Users/runner/Temp/x/repo", "C:"); got != "file:///C:/Users/runner/Temp/x/repo" {
		t.Fatalf("windows remote %q", got)
	}
	if got := gitFileURLFor("/tmp/x/repo", ""); got != "file:///tmp/x/repo" {
		t.Fatalf("unix remote %q", got)
	}
	for _, got := range []string{
		gitFileURLFor("C:/Users/runner/Temp/x/repo", "C:"),
		gitFileURLFor("/tmp/x/repo", ""),
		gitFileURL(t.TempDir()),
	} {
		if strings.Contains(got, "\\") {
			t.Fatalf("remote %q carries a backslash git parses as an escape", got)
		}
	}
}

// TestServeWritesTheGitSafeRemote pins the wiring: serve must route the
// fixture remote through gitFileURL, never a raw directory. A mutant that
// restores the raw interpolation fails the backslash clause on Windows,
// where the temporary directory carries separators git would consume.
func TestServeWritesTheGitSafeRemote(t *testing.T) {
	ids := newGitIdentities(t)
	repo := t.TempDir()
	ids.serve(repo, "https://example.com/pinned")
	payload, err := os.ReadFile(ids.path) // #nosec G304 -- test gitconfig
	if err != nil {
		t.Fatal(err)
	}
	want := "[url \"" + gitFileURL(repo) + "\"]\n\tinsteadOf = https://example.com/pinned\n"
	if string(payload) != want {
		t.Fatalf("gitconfig:\n%s\nwant:\n%s", payload, want)
	}
	if strings.Contains(string(payload), "\\") {
		t.Fatalf("gitconfig carries a backslash git parses as an escape:\n%s", payload)
	}
}

// serve maps one operand URL onto a local repository and returns the operand
// for installation or declaration. Every call appends one insteadOf block to
// the test's gitconfig; operands must be distinct per repository.
func (g *gitIdentities) serve(repo, operand string) string {
	g.t.Helper()
	block := "[url \"" + gitFileURL(repo) + "\"]\n\tinsteadOf = " + operand + "\n"
	file, err := os.OpenFile(g.path, os.O_APPEND|os.O_WRONLY, 0o644) // #nosec G304 -- test gitconfig
	if err != nil {
		g.t.Fatal(err)
	}
	if _, err := file.WriteString(block); err != nil {
		_ = file.Close()
		g.t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		g.t.Fatal(err)
	}
	return operand
}
