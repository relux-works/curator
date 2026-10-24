//go:build unix

package closureexec

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Paired real-worker witnesses reject omitted domain termination at the
// survival assertion. A setup failure, hang, skip or unrelated error never
// counts as detection.
func TestPortableDescendantNarrowingControl(t *testing.T) {
	goTool := os.Getenv("CURATOR_GO")
	if goTool == "" {
		goTool = "go"
	}
	for _, plant := range []struct{ name, file, before, after, witness, message string }{
		{"omit-group-kill", "portable_process_unix.go", "_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)", "_ = command.Process.Kill()", "TestManagerProcessRunnerTimeoutCleansDescendantProcessGroup", "survived portable timeout: survival marker"},
	} {
		t.Run(plant.name, func(t *testing.T) {
			sourcePath, err := filepath.Abs(plant.file)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(raw), "\r\n", "\n")
			if strings.Count(source, plant.before) != 1 {
				t.Fatal("descendant plant anchor not unique")
			}
			run := func(overlay string) ([]byte, error) {
				// The 90-second hang guard is below the former 120-second natural
				// lifetime, but above the fixture's 60-second worker allowance plus
				// 30 seconds for warm setup. Only the assertion supplies a verdict.
				args := []string{"test", "-count=1", "-timeout=90s", "-v", "-run", "^" + plant.witness + "$", "."}
				if overlay != "" {
					args = append([]string{"test", "-overlay", overlay}, args[1:]...)
				}
				command := exec.CommandContext(t.Context(), goTool, args...)
				for _, item := range os.Environ() {
					if !strings.HasPrefix(item, "GIT_DIR=") && !strings.HasPrefix(item, "GIT_WORK_TREE=") {
						command.Env = append(command.Env, item)
					}
				}
				return command.CombinedOutput()
			}
			output, err := run("")
			if err != nil || !strings.Contains(string(output), "--- PASS: "+plant.witness) {
				t.Fatalf("positive descendant control: %v\n%s", err, output)
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, filepath.Base(plant.file))
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, plant.before, plant.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(map[string]any{"Replace": map[string]string{sourcePath: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlay, payload, 0600); err != nil {
				t.Fatal(err)
			}
			output, err = run(overlay)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.witness) || !strings.Contains(string(output), plant.message) {
				t.Fatalf("descendant mutant survived or failed outside its assertion: %v\n%s", err, output)
			}
			t.Logf("intended assertion detected %s:\n%s", plant.name, output)
		})
	}
}
