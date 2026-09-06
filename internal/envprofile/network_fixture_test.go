package envprofile

import (
	"os"
	"path/filepath"
	"testing"
)

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

// serve maps one operand URL onto a local repository and returns the operand
// for installation or declaration. Every call appends one insteadOf block to
// the test's gitconfig; operands must be distinct per repository.
func (g *gitIdentities) serve(repo, operand string) string {
	g.t.Helper()
	block := "[url \"file://" + repo + "\"]\n\tinsteadOf = " + operand + "\n"
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
