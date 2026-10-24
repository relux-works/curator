package registry

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The fake-clock witnesses must still reject both a missing elapsed allowance
// and a missing future-timestamp gate. Only the named assertion with exit 1
// counts; build errors, missing plants and unrelated failures do not.
func TestSnapshotClockNarrowingControls(t *testing.T) {
	sourcePath, err := filepath.Abs("snapshot.go")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(raw), "\r\n", "\n")
	goTool := os.Getenv("CURATOR_GO")
	if goTool == "" {
		goTool = "go"
	}
	for _, plant := range []struct{ name, before, after, run, message string }{
		{"omit-elapsed-allowance", ".Add(time.Since(start))", ".Add(0*time.Since(start))", "TestSnapshotFutureBoundToleratesCheckerLatency", "a snapshot minted during its own fetch must be accepted"},
		{"widen-future-bound-one-second", "now.Add(clockSkew).Add(time.Since(start))", "now.Add(clockSkew+time.Second).Add(time.Since(start))", "TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew", "refused=false want true"},
		{"omit-future-refusal", "if parsed.CreatedAt.After(", "if false && parsed.CreatedAt.After(", "TestSnapshotZeroClockSkewIsLiteral", "a zero skew must refuse a sub-second future timestamp"},
	} {
		t.Run(plant.name, func(t *testing.T) {
			run := func(overlay string) ([]byte, error) {
				args := []string{"test", "-count=1", "-timeout=3m", "-v", "-run", "^" + plant.run + "$", "."}
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
			if err != nil || !strings.Contains(string(output), "--- PASS: "+plant.run) {
				t.Fatalf("positive clock control: %v\n%s", err, output)
			}
			if strings.Count(source, plant.before) != 1 {
				t.Fatal("clock plant anchor not unique")
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, "snapshot.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, plant.before, plant.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(map[string]any{"Replace": map[string]string{sourcePath: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlay, raw, 0600); err != nil {
				t.Fatal(err)
			}
			output, err = run(overlay)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.run) || !strings.Contains(string(output), plant.message) {
				t.Fatalf("clock mutant survived or failed outside its assertion: %v\n%s", err, output)
			}
		})
	}
}
