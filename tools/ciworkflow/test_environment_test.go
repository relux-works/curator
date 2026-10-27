package ciworkflow

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The Unix boundary clears entire selector families, preserves CI inputs, and
// prevents an inherited index selector from writing outside the child fixture.
// Bypass and broken-family controls must demonstrate their intended violation.
func TestUnixTestEnvironment(t *testing.T) {
	script, err := filepath.Abs("../../.github/ci/isolated-test.sh")
	if err != nil {
		t.Fatal(err)
	}
	script = filepath.ToSlash(script)
	contents, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	base := make([]string, 0)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "GIT_") || strings.HasPrefix(key, "SSH_") || strings.HasPrefix(key, "XDG_") || key == "HOME" || key == "GOFLAGS" || key == "CODEX_HOME" || key == "CLAUDE_CONFIG_DIR" || key == "PI_CODING_AGENT_DIR" {
			continue
		}
		base = append(base, entry)
	}
	base = append(base, "HOME="+root)
	run := func(dir string, env []string, command string, args ...string) []byte {
		t.Helper()
		cmd := exec.Command(command, args...)
		cmd.Dir, cmd.Env = dir, env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", command, args, err, out)
		}
		return out
	}
	decoy, child := filepath.Join(root, "decoy"), filepath.Join(root, "child")
	for _, dir := range []string{decoy, child} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		run(dir, base, "git", "init", "-q")
		if err := os.WriteFile(filepath.Join(dir, "fixture"), []byte(dir), 0o600); err != nil {
			t.Fatal(err)
		}
		run(dir, base, "git", "add", "fixture")
	}
	index := filepath.Join(decoy, ".git", "index")
	before, err := os.ReadFile(index)
	if err != nil {
		t.Fatal(err)
	}
	hostile := append(append([]string{}, base...), "GIT_INDEX_FILE="+index)
	// The unguarded real Git operation must reproduce the forbidden outer write.
	run(child, hostile, "git", "add", "fixture")
	after, err := os.ReadFile(index)
	if err != nil || bytes.Equal(before, after) {
		t.Fatalf("hostile index control did not write the decoy: %v", err)
	}
	if err := os.WriteFile(index, before, 0o600); err != nil {
		t.Fatal(err)
	}
	run(child, hostile, "bash", script, "git", "add", "fixture")
	after, err = os.ReadFile(index)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("isolated child changed decoy index: %v", err)
	}
	probe := `set -eu
[[ "$HOME" = /* && "$HOME" != "$OUTER_HOME" && -d "$HOME" ]] || exit 72
while IFS= read -r name; do
  case "$name" in
    GIT_*|SSH_*|XDG_*|GOFLAGS|CODEX_HOME|CLAUDE_CONFIG_DIR|PI_CODING_AGENT_DIR) exit 73;;
  esac
done < <(compgen -e)
[[ "$CURATOR_CONFORMANCE_ROOT" = preserved && "$GOCACHE" = preserved && "$CI_UNIX_TEST_SHARDS" = preserved ]] || exit 74
`
	hostile = append(hostile, "GIT_FUTURE_SELECTOR=hostile", "SSH_FUTURE_SELECTOR=hostile", "XDG_FUTURE_SELECTOR=hostile", "XDG_CONFIG_HOME="+root, "XDG_RUNTIME_DIR="+root, "SSH_AUTH_SOCK="+root, "GOFLAGS=-invalid", "CODEX_HOME="+root, "CLAUDE_CONFIG_DIR="+root, "PI_CODING_AGENT_DIR="+root, "OUTER_HOME="+root, "CURATOR_CONFORMANCE_ROOT=preserved", "CI_UNIX_TEST_SHARDS=preserved")
	for i, entry := range hostile {
		if strings.HasPrefix(entry, "GOCACHE=") {
			hostile[i] = "GOCACHE=preserved"
		}
	}
	if !strings.Contains(strings.Join(hostile, "\n"), "GOCACHE=") {
		hostile = append(hostile, "GOCACHE=preserved")
	}
	run(child, hostile, "bash", script, "bash", "-c", probe)
	for _, family := range []string{"GIT_*|", "SSH_*|", "XDG_*|"} {
		t.Run(family[:3], func(t *testing.T) {
			broken := strings.Replace(string(contents), family, "", 1)
			if broken == string(contents) {
				t.Fatal("family plant not applied")
			}
			path := filepath.Join(t.TempDir(), "broken.sh")
			if err := os.WriteFile(path, []byte(broken), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", filepath.ToSlash(path), "bash", "-c", probe)
			cmd.Dir, cmd.Env = child, hostile
			out, err := cmd.CombinedOutput()
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
				t.Fatalf("missing-family plant must expose selector (exit73): %v\n%s", err, out)
			}
		})
	}
	t.Run("relative-home", func(t *testing.T) {
		broken := strings.Replace(string(contents), `export HOME="$test_home"`, `export HOME=.`, 1)
		if broken == string(contents) {
			t.Fatal("relative HOME plant not applied")
		}
		path := filepath.Join(t.TempDir(), "broken.sh")
		if err := os.WriteFile(path, []byte(broken), 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("bash", filepath.ToSlash(path), "bash", "-c", probe)
		cmd.Dir, cmd.Env = child, hostile
		out, err := cmd.CombinedOutput()
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 72 {
			t.Fatalf("relative HOME plant must fail absolute HOME check (exit72): %v\n%s", err, out)
		}
	})
}
