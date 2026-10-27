package testgitenv

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Independent regression inputs, not a list used by the boundary. A fresh
// unknown name proves that isolation does not depend on enumerating Git keys.
func gitPrefixProbeKeys(t testing.TB) []string {
	t.Helper()
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	return []string{
		"GIT_TRACE", "GIT_TRACE_SETUP", "GIT_TRACE2_EVENT",
		"GIT_ATTR_SOURCE", "GIT_INDEX_VERSION", "GIT_FOO_BAR",
		fmt.Sprintf("GIT_FOO_BAR_%X", random),
		"GIT_SSH", "GIT_SSH_COMMAND", "GIT_SSH_VARIANT", "GIT_ASKPASS", "SSH_ASKPASS",
	}
}

// An inherited trace destination must not receive bytes from a real scratch
// commit. The clean row and successful child prevent a reject-all false pass.
func TestInheritedGitTraceIsolation(t *testing.T) {
	for _, key := range []string{"clean", "GIT_TRACE", "GIT_TRACE_SETUP", "GIT_TRACE2_EVENT"} {
		t.Run(key, func(t *testing.T) {
			decoy := newDecoy(t)
			trace := filepath.Join(decoy, "trace.log")
			if err := os.WriteFile(trace, []byte("external sentinel\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before := decoyImage(t, decoy)
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestScratchFixtureChild$")
			cmd.Env = append(os.Environ(), "CURATOR_TEST_HOSTILE_FIXTURE=1")
			if key != "clean" {
				cmd.Env = append(cmd.Env, key+"="+trace)
			}
			output, err := cmd.CombinedOutput()
			requireDecoyImage(t, before, decoyImage(t, decoy))
			if err != nil || !strings.Contains(string(output), "--- PASS: TestScratchFixtureChild") {
				t.Fatalf("trace fixture child: %v\n%s", err, output)
			}
		})
	}
}

// A fixture may deliberately enable tracing after entering the boundary; the
// trace must come from real Git and remain inside the isolated HOME.
func TestFixtureLocalGitTrace(t *testing.T) {
	cleanup := Isolate()
	defer cleanup()
	trace := filepath.Join(os.Getenv("HOME"), "fixture.trace")
	t.Setenv("GIT_TRACE", trace)
	fixtureGit(t, t.TempDir(), "init", "-q")
	body, err := os.ReadFile(trace)
	if err != nil || !strings.Contains(string(body), "built-in: git init") {
		t.Fatalf("fixture-local Git trace was not recorded: %v", err)
	}
}
