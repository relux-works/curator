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
	// A native child observes os.Environ directly, independently of Bash's
	// variable table. The marker distinguishes child rejection from no dispatch.
	if os.Getenv("CI_TEST_BOUNDARY_CHILD") == "1" {
		if err := os.WriteFile(os.Getenv("CI_TEST_BOUNDARY_MARKER"), []byte("dispatched"), 0o600); err != nil {
			os.Exit(75)
		}
		home := os.Getenv("HOME")
		if (!filepath.IsAbs(home) && !strings.HasPrefix(home, "/")) || home == os.Getenv("OUTER_HOME") {
			os.Exit(72)
		}
		for _, entry := range os.Environ() {
			name, _, _ := strings.Cut(entry, "=")
			if strings.HasPrefix(name, "GIT_") || strings.HasPrefix(name, "SSH_") || strings.HasPrefix(name, "XDG_") || name == "GOFLAGS" || name == "CODEX_HOME" || name == "CLAUDE_CONFIG_DIR" || name == "PI_CODING_AGENT_DIR" {
				os.Exit(73)
			}
		}
		if os.Getenv("CURATOR_CONFORMANCE_ROOT") != "preserved\nsecond=line" || os.Getenv("GOCACHE") != "preserved" || os.Getenv("CI_UNIX_TEST_SHARDS") != "preserved" || os.Args[len(os.Args)-1] != "argument with spaces\nand=equals" {
			os.Exit(74)
		}
		os.Exit(0)
	}
	shell := os.Getenv("CI_TEST_BOUNDARY_BASH")
	if shell == "" {
		shell = "bash"
	}
	script, err := filepath.Abs("../../.github/ci/isolated-test.sh")
	if err != nil {
		t.Fatal(err)
	}
	script = filepath.ToSlash(script)
	wrapper, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(filepath.Dir(script), "isolated-test.cjs"))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	mutant := func(program string) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "isolated-test.sh")
		if err := os.WriteFile(path, wrapper, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "isolated-test.cjs"), []byte(program), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
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
	run(child, hostile, shell, script, "git", "add", "fixture")
	after, err = os.ReadFile(index)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("isolated child changed decoy index: %v", err)
	}
	marker := filepath.Join(root, "child-marker")
	hostile = append(hostile, "GIT_FUTURE_SELECTOR=hostile", "SSH_FUTURE_SELECTOR=hostile", "XDG_FUTURE_SELECTOR=hostile", "XDG_CONFIG_HOME="+root, "XDG_RUNTIME_DIR="+root, "SSH_AUTH_SOCK="+root, "GOFLAGS=-invalid", "CODEX_HOME="+root, "CLAUDE_CONFIG_DIR="+root, "PI_CODING_AGENT_DIR="+root, "OUTER_HOME="+root, "CURATOR_CONFORMANCE_ROOT=preserved\nsecond=line", "CI_UNIX_TEST_SHARDS=preserved", "CI_TEST_BOUNDARY_CHILD=1", "CI_TEST_BOUNDARY_MARKER="+marker)
	for i, entry := range hostile {
		if strings.HasPrefix(entry, "GOCACHE=") {
			hostile[i] = "GOCACHE=preserved"
		}
	}
	if !strings.Contains(strings.Join(hostile, "\n"), "GOCACHE=") {
		hostile = append(hostile, "GOCACHE=preserved")
	}
	probeArgs := []string{filepath.ToSlash(os.Args[0]), "-test.run=^TestUnixTestEnvironment$", "--", "argument with spaces\nand=equals"}
	probe := func(path string, env []string) ([]byte, error) {
		t.Helper()
		if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		cmd := exec.Command(shell, append([]string{filepath.ToSlash(path)}, probeArgs...)...)
		cmd.Dir, cmd.Env = child, env
		return cmd.CombinedOutput()
	}
	if out, err := probe(script, hostile); err != nil {
		t.Fatalf("valid raw environment boundary: %v\n%s", err, out)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("valid boundary did not dispatch raw observer: %v", err)
	}
	for _, family := range []string{"GIT_", "SSH_", "XDG_"} {
		t.Run(family[:3], func(t *testing.T) {
			broken := strings.Replace(string(contents), "'"+family+"'", "'OTHER_'", 1)
			if broken == string(contents) {
				t.Fatal("family plant not applied")
			}
			out, err := probe(mutant(broken), hostile)
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
				t.Fatalf("missing-family plant must expose selector (exit73): %v\n%s", err, out)
			}
			if _, err := os.Stat(marker); err != nil {
				t.Fatalf("missing-family plant did not dispatch raw observer: %v", err)
			}
		})
	}
	t.Run("relative-home", func(t *testing.T) {
		broken := strings.Replace(string(contents), "childEnv.HOME = realpathSync(mkdtempSync(join(tmpdir(), 'curator-test-home-')));", "childEnv.HOME = '.';", 1)
		if broken == string(contents) {
			t.Fatal("relative HOME plant not applied")
		}
		out, err := probe(mutant(broken), hostile)
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 72 {
			t.Fatalf("relative HOME plant must fail absolute HOME check (exit72): %v\n%s", err, out)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("relative HOME plant did not dispatch observer: %v", err)
		}
	})
	for _, name := range []string{"GIT_BAD-NAME", "SSH_BAD.NAME", "XDG_ODD\nKEY"} {
		t.Run(name, func(t *testing.T) {
			out, err := probe(script, append(append([]string{}, hostile...), name+"=hostile\nvalue=two"))
			if err != nil {
				t.Fatalf("raw selector must be removed before child: %v\n%s", err, out)
			}
			if _, err := os.Stat(marker); err != nil {
				t.Fatalf("cleaned raw selector did not dispatch observer: %v", err)
			}
		})
	}
	for name, expression := range map[string]string{"failed": "(() => { throw new Error('control'); })()", "empty": "[]"} {
		t.Run("enumeration-"+name, func(t *testing.T) {
			broken := strings.Replace(string(contents), "Object.entries(process.env)", expression, 1)
			if broken == string(contents) {
				t.Fatal("enumeration plant not applied")
			}
			out, err := probe(mutant(broken), hostile)
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !bytes.Contains(out, []byte("raw environment enumeration failed")) {
				t.Fatalf("want explicit enumeration refusal: %v\n%s", err, out)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("failed enumeration dispatched a child: %v", err)
			}
		})
	}
	t.Run("line-enumeration-blind-spot", func(t *testing.T) {
		// Regression witness: the previously committed Bash variable-table
		// enumeration dispatches with a raw newline name still inherited.
		broken := `set -euo pipefail
while IFS= read -r name; do
  case "$name" in
    GIT_*|SSH_*|XDG_*|GOFLAGS|CODEX_HOME|CLAUDE_CONFIG_DIR|PI_CODING_AGENT_DIR) unset "$name";;
  esac
done < <(compgen -e)
test_home="$(mktemp -d)"
test_home="$(cd "$test_home" && pwd -P)"
export HOME="$test_home"
exec "$@"
`
		path := filepath.Join(t.TempDir(), "broken.sh")
		if err := os.WriteFile(path, []byte(broken), 0o600); err != nil {
			t.Fatal(err)
		}
		out, err := probe(path, append(append([]string{}, hostile...), "XDG_ODD\nKEY=hostile"))
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
			t.Fatalf("raw observer must catch newline selector survivor (exit73): %v\n%s", err, out)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("blind-spot control did not dispatch observer: %v", err)
		}
	})
}
