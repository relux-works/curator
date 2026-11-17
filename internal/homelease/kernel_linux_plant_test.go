//go:build linux

package homelease

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Claim: dropping the thread/zombie refusal at the Linux kernel read lets a thread id produce a tuple, so TestLinuxKernelRefusesThreadAndZombie kills the plant after its control passes.
func TestLinuxKernelThreadZombiePlant(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool := testtoolchain.Go(t)
	run := func(overlay string) ([]byte, error) {
		args := []string{"test", "-count=1", "-timeout=5m", "-v", "./internal/homelease", "-run", "^TestLinuxKernelRefusesThreadAndZombie$"}
		if overlay != "" {
			args = append([]string{"test", "-overlay", overlay}, args[1:]...)
		}
		cmd := exec.Command(goTool, args...)
		cmd.Dir = root
		for _, kv := range os.Environ() {
			if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") {
				cmd.Env = append(cmd.Env, kv)
			}
		}
		testtoolchain.InheritHostGOROOTLock(t, cmd)
		return cmd.CombinedOutput()
	}
	out, err := run("")
	if err != nil || !strings.Contains(string(out), "--- PASS: TestLinuxKernelRefusesThreadAndZombie") {
		t.Fatalf("positive control failed: %v\n%s", err, out)
	}
	path := filepath.Join(root, "internal/homelease/kernel_linux.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(raw), "\r\n", "\n")
	const anchor = `if tgid != pid || stateBefore == "Z" || stateBefore == "X" || stateAfter == "Z" || stateAfter == "X" {`
	if strings.Count(source, anchor) != 1 {
		t.Fatal("mutation anchor not unique")
	}
	scratch := t.TempDir()
	replacement := filepath.Join(scratch, "replacement.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(source, anchor, `if false && (tgid != pid || stateBefore == "Z" || stateBefore == "X" || stateAfter == "Z" || stateAfter == "X") {`, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlay, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	out, err = run(overlay)
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: TestLinuxKernelRefusesThreadAndZombie") ||
		!strings.Contains(string(out), "produced a process tuple") || strings.Contains(string(out), "panic:") {
		t.Fatalf("plant survived or failed for another reason: %v\n%s", err, out)
	}
	t.Logf("intended expected-red exit 1:\n%s", out)
}
