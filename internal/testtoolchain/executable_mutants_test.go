package testtoolchain

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each fallback mutant must fail its matching ambient-tool rejection claim,
// rather than merely crashing, failing compilation, or running no cases.
func TestExecutableSelectionNarrowingMutants(t *testing.T) {
	for _, tc := range []struct{ name, before, after, test, failure string }{
		{
			"Go-PATH-fallback",
			`return "", fmt.Errorf("set CURATOR_GO or GOROOT; the test binary has no GOROOT: %w", os.ErrNotExist)`,
			`return exec.LookPath("go")`,
			"^TestGoExecutableSelection$", "--- FAIL: TestGoExecutableSelection/absent",
		},
		{
			"Rust-PATH-fallback",
			`return requireExecutable(filepath.Join(home, ".rustup", "toolchains", version+"-"+arch+"-"+osName, "bin", name))`,
			`path, err := requireExecutable(filepath.Join(home, ".rustup", "toolchains", version+"-"+arch+"-"+osName, "bin", name)); if errors.Is(err, os.ErrNotExist) { return exec.LookPath("rustc") }; return path, err`,
			"^TestRustcExecutableSelection$", "--- FAIL: TestRustcExecutableSelection/darwin",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, err := filepath.Abs("executable.go")
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(data), tc.before) != 1 {
				t.Fatal("toolchain fallback mutation anchor is not unique")
			}
			mutated := strings.Replace(string(data), tc.before, tc.after, 1)
			mutated = strings.Replace(mutated, "import (", "import (\n\"os/exec\"", 1)
			directory := t.TempDir()
			replacement := filepath.Join(directory, "executable.go")
			if err := os.WriteFile(replacement, []byte(mutated), 0o600); err != nil {
				t.Fatal(err)
			}
			data, err = json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(directory, "overlay.json")
			if err := os.WriteFile(overlay, data, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(Go(t), "test", "-overlay", overlay, "-count=1", "-timeout=1m", "-v", "-run", tc.test, ".")
			output, err := command.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 ||
				!strings.Contains(string(output), tc.failure) ||
				!strings.Contains(string(output), "ambient") ||
				strings.Contains(string(output), "build failed") {
				t.Fatalf("fallback mutant did not fail its intended rejection: %v\n%s", err, output)
			}
			t.Logf("%s: exit 1 at %s", tc.name, tc.failure)
		})
	}
}
