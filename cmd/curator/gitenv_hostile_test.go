package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hostileGitConfig is the shape of an operator config that broke the
// signer rows on a runner that copies the host ~/.gitconfig into HOME:
// every commit and tag is signed with an SSH key.
func hostileGitConfig(signingKey string) string {
	return "[user]\n\tname = Hostile Operator\n\temail = hostile@example.invalid\n\tsigningkey = " + signingKey + "\n" +
		"[gpg]\n\tformat = ssh\n[commit]\n\tgpgsign = true\n[tag]\n\tgpgsign = true\n"
}

// hostileGitEnv returns this process's environment with every git config
// entry point (GIT_CONFIG_GLOBAL, HOME/.gitconfig, XDG_CONFIG_HOME/git)
// aimed at a signing-enabled config and system config re-enabled.
func hostileGitEnv(t *testing.T) []string {
	t.Helper()
	private, _ := generateSSHKey(t)
	root := t.TempDir()
	// Forward slashes: git's config parser treats a backslash in a
	// Windows path as an escape and refuses the line.
	body := hostileGitConfig(filepath.ToSlash(private))
	home := filepath.Join(root, "home")
	xdg := filepath.Join(root, "xdg")
	for _, path := range []string{filepath.Join(root, "global.gitconfig"), filepath.Join(home, ".gitconfig"), filepath.Join(xdg, "git", "config")} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	env := []string{}
	for _, entry := range os.Environ() {
		switch strings.SplitN(entry, "=", 2)[0] {
		case "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM", "HOME", "XDG_CONFIG_HOME":
			continue
		}
		env = append(env, entry)
	}
	return append(env, "GIT_CONFIG_GLOBAL="+filepath.Join(root, "global.gitconfig"),
		"HOME="+home, "XDG_CONFIG_HOME="+xdg)
}

// TestSignerRowsSurviveHostileGlobalGitConfig re-runs the two signer rows
// that failed on a signing-enabled runner in a child test process whose
// whole environment carries that hostile config. The child's TestMain
// (testgitenv.Isolate) is the only thing standing between the config and
// the fixture commits, so removing it turns this row red.
func TestSignerRowsSurviveHostileGlobalGitConfig(t *testing.T) {
	requireGit(t)
	if os.Getenv("CURATOR_CONFORMANCE_ROOT") == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set; the signer rows would skip")
	}
	env := hostileGitEnv(t)

	// Precondition: the config is really hostile. A plain commit under
	// it, without the isolation, comes out signed.
	probe := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"commit", "-q", "--allow-empty", "-m", "probe"}} {
		command := exec.Command("git", args...)
		command.Dir = probe
		command.Env = env
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("hostile probe git %v: %v\n%s", args, err, output)
		}
	}
	command := exec.Command("git", "cat-file", "commit", "HEAD")
	command.Dir = probe
	command.Env = env
	raw, err := command.Output()
	if err != nil || !strings.Contains(string(raw), "\ngpgsig ") {
		t.Fatalf("hostile config did not sign the probe commit (err=%v):\n%s", err, raw)
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(executable, "-test.v", "-test.count=1",
		"-test.run", "^(TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature)$")
	// The helper-process branch skips the package GOROOT lock this
	// process already holds; it still runs TestMain's git isolation.
	child.Env = append(env, cliHelperProcess+"=1")
	output, err := child.CombinedOutput()
	if err != nil {
		t.Fatalf("signer rows under a hostile global git config: %v\n%s", err, output)
	}
	for _, want := range []string{
		"--- PASS: TestSourceSignerVectorsAtProfileInstall/unsigned-refused",
		"--- PASS: TestRevisionDoesNotBorrowTagSignature",
	} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("child run is missing %q (skipped or not run):\n%s", want, output)
		}
	}
}
