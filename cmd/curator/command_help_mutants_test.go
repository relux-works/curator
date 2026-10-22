package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Each overlay weakens one help boundary while leaving the guard present.
// Only exit 1 plus a named behavioral test failure counts as a killed mutant;
// compilation errors, setup failures and timeouts cannot satisfy this control.
func TestCommandHelpNarrowingMutants(t *testing.T) {
	for _, mutant := range []struct {
		name, file, before, after, test, failure string
	}{
		{"init-bypasses-help", "main.go", `command != nil && hasHelpFlag(args[1:])`, `command != nil && args[0] != "init" && hasHelpFlag(args[1:])`, "^TestRegisteredSubcommandHelpNoWrites$/^init$", "--- FAIL: TestRegisteredSubcommandHelpNoWrites/init/1"},
		{"only-leading-flags", "command_help.go", `if arg == "-h" || arg == "--help" {`, `if (arg == "-h" || arg == "--help") && (args[0] == "-h" || args[0] == "--help") {`, "^TestRegisteredSubcommandHelpNoWrites$/^init$", "--- FAIL: TestRegisteredSubcommandHelpNoWrites/init/2"},
		{"help-crosses-separator", "command_help.go", "if arg == \"--\" {\n\t\t\tbreak\n\t\t}\n\t\tif arg == \"-h\"", "if arg == \"--\" {\n\t\t\tcontinue\n\t\t}\n\t\tif arg == \"-h\"", "^TestCommandHelpRunPassthrough$", "--- FAIL: TestCommandHelpRunPassthrough"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			original, err := filepath.Abs(mutant.file)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(data), "\r\n", "\n")
			if strings.Count(source, mutant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, mutant.before, mutant.after, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			data, err = json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlay, data, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(testtoolchain.Go(t), "test", "-count=1", "-timeout=3m", "-v", "-run", mutant.test, ".")
			// GOFLAGS carries the overlay into the real CLI binary built by the
			// behavioral suite too. These child tests do not fingerprint GOROOT;
			// reuse the repository's helper protocol to avoid taking its lock twice.
			command.Env = append(os.Environ(), cliHelperProcess+"=1", "GOFLAGS="+os.Getenv("GOFLAGS")+" "+strconv.Quote("-overlay="+overlay))
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), mutant.failure) || strings.Contains(string(output), "build failed") {
				t.Fatalf("mutant lacks intended behavioral exit 1: %v\n%s", err, output)
			}
			t.Logf("mutant %s: go test -run %q . exit=1; killed by %s", mutant.name, mutant.test, mutant.failure)
		})
	}
}
