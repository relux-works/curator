package envprofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Only refmap suppression is removed: transport, protocol policy and every
// other fetch guard remain. Behavioral failures, not token searches, kill it.
func TestMissingPinRefmapNarrowingMutant(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "internal", "envprofile", "gitsource.go")
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	const anchor = `"--refmap=", `
	if strings.Count(string(raw), anchor) != 1 {
		t.Fatal("refmap anchor must be unique")
	}
	dir := t.TempDir()
	plant := filepath.Join(dir, "gitsource.go")
	if err := os.WriteFile(plant, []byte(strings.Replace(string(raw), anchor, "", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{source: plant}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=7m", "-json", "./internal/envprofile", "-run", "^TestInstallMissingPinPreservesConfiguredRefs$")
	cmd.Dir = root
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	t.Logf("drop-refmap subprocess PID=%d exit=%d", cmd.Process.Pid, cmd.ProcessState.ExitCode())
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("expected behavioral exit 1, got %v: %s", err, output.String())
	}
	actions, diagnostics := map[string]string{}, map[string]string{}
	for _, line := range strings.Split(output.String(), "\n") {
		var event struct{ Action, Test, Output string }
		if json.Unmarshal([]byte(line), &event) != nil {
			continue
		}
		if event.Action == "pass" || event.Action == "fail" {
			actions[event.Test] = event.Action
		}
		diagnostics[event.Test] += event.Output
	}
	for _, dest := range []string{"refs/tags/v1.0.0", "refs/remotes/origin/preserved"} {
		for _, absent := range []bool{false, true} {
			for _, mode := range []string{"stale", "unsigned", "wrong-signer", "cached-offline"} {
				name := fmt.Sprintf("TestInstallMissingPinPreservesConfiguredRefs/%s/absent-%t/%s", dest, absent, mode)
				if mode == "cached-offline" {
					if actions[name] != "pass" {
						t.Errorf("valid neighbour did not pass: %s: %s", name, diagnostics[name])
					}
				} else if actions[name] != "fail" || !strings.Contains(diagnostics[name], "existing destination ref changed:") {
					t.Errorf("no intended ref-preservation kill: %s: %s", name, diagnostics[name])
				}
				t.Logf("%s: %s\n%s", name, actions[name], diagnostics[name])
			}
		}
	}
}
