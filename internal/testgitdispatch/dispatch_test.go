package testgitdispatch

import (
	"debug/buildinfo"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Reinstalling the SSH fixture must preserve the executable inode; only its
// sidecar is mutable. This catches the ETXTBSY mechanism without a timing race.
func TestInstallPreservesExecutable(t *testing.T) {
	target := filepath.Join(t.TempDir(), "ssh")
	Install(t, target)
	stamp := time.Unix(1, 0)
	if err := os.Chtimes(target, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	Install(t, target)
	after, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("installed executable rewritten")
	}
	if !os.SameFile(before, after) {
		t.Fatal("installed executable replaced")
	}
}

// Unknown declared network endpoints fail in the fixture before real Git/DNS;
// ordinary local Git remains a valid direct sibling.
func TestUnknownEndpointRefused(t *testing.T) {
	for _, endpoint := range []string{"https://unregistered.invalid/repository", "unregistered.invalid:repository", "git@unregistered.invalid:repository"} {
		for _, mode := range []string{"clone", "global-options", "named", "all"} {
			t.Run(endpoint+"/"+mode, func(t *testing.T) {
				New(t)
				// A second, independent test-only protocol boundary prevents DNS even
				// if a future fixture regression admits this unknown endpoint.
				t.Setenv("GIT_ALLOW_PROTOCOL", "file")
				repo := t.TempDir()
				run := func(args ...string) {
					t.Helper()
					if out, err := Command(t, args...).CombinedOutput(); err != nil {
						t.Fatalf("setup: %v %s", err, out)
					}
				}
				run("init", repo)
				run("-C", repo, "config", "remote.origin.url", endpoint)
				args := []string{"clone", endpoint, filepath.Join(t.TempDir(), "clone")}
				switch mode {
				case "global-options":
					args = append([]string{"-C", repo, "-c", "advice.detachedHead=false"}, args...)
				case "named":
					args = []string{"-C", repo, "fetch", "origin"}
				case "all":
					args = []string{"-C", repo, "fetch", "--all"}
				}
				out, err := Command(t, args...).CombinedOutput()
				if err == nil || !strings.Contains(string(out), "fixture endpoint unavailable") {
					t.Fatalf("unknown endpoint reached Git: %v %s", err, out)
				}
			})
		}
	}
}

// Direct ordinary calls and acquisition calls share explicit headless settings;
// the local Git sibling succeeds with hostile ambient prompt settings planted.
func TestFixtureCommandsAreHeadless(t *testing.T) {
	t.Setenv("GIT_TERMINAL_PROMPT", "1")
	t.Setenv("GIT_ASKPASS", "forbidden-askpass")
	t.Setenv("GIT_SSH_COMMAND", "forbidden-ssh")
	New(t)
	cmd := Command(t, "--version")
	joined := strings.Join(cmd.Env, "\n")
	for _, want := range []string{"GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=\n", "SSH_ASKPASS_REQUIRE=never", "BatchMode=yes"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing headless setting %q", want)
		}
	}
	if strings.Contains(joined, "forbidden-") {
		t.Fatal("ambient prompt helper retained")
	}
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "git version") {
		t.Fatalf("local sibling: %v %s", err, out)
	}
}

func TestOrdinaryGitBypassesDispatcher(t *testing.T) {
	path := New(t)
	cmd := Command(t, "--version")
	if filepath.Dir(cmd.Path) == filepath.Dir(path) {
		t.Fatal("ordinary Git runs through dispatcher")
	}
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "git version") {
		t.Fatalf("valid local Git: %v %s", err, out)
	}
}

// Race-instrumented package tests still launch a distinct non-race helper,
// avoiding the race test executable's exit delay on each Git operation.
func TestHelperIsDedicatedNonRaceBinary(t *testing.T) {
	if err := Build(); err != nil {
		t.Fatal(err)
	}
	info, err := buildinfo.ReadFile(helperBinary)
	if err != nil {
		t.Fatal(err)
	}
	for _, setting := range info.Settings {
		if setting.Key == "-race" && setting.Value == "true" {
			t.Fatal("fixture helper is race-instrumented")
		}
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	helperInfo, err := os.Stat(helperBinary)
	if err != nil {
		t.Fatal(err)
	}
	selfInfo, err := os.Stat(self)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(helperInfo, selfInfo) || info.Path != "command-line-arguments" {
		t.Fatal("helper is not the dedicated command")
	}
}
