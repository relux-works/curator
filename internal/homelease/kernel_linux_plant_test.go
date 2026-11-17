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

// Claim: each refusal at the Linux kernel read kills its own plant by name, after the unchanged control passes. Dropping only the thread check lets the parked thread id produce a tuple, and dropping only the zombie/dead check lets the zombie child keep its tuple, so neither half hides behind the other row's failure.
func TestLinuxKernelThreadZombiePlant(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool := testtoolchain.Go(t)
	cases := []struct{ name, anchor, replacement, message string }{
		{"drop-thread-refusal",
			"if tgid != pid {",
			"if false && tgid != pid {",
			"produced a process tuple"},
		{"drop-zombie-refusal",
			`if stateBefore == "Z" || stateBefore == "X" || stateAfter == "Z" || stateAfter == "X" {`,
			`if false && (stateBefore == "Z" || stateBefore == "X" || stateAfter == "Z" || stateAfter == "X") {`,
			"zombie child still produced a tuple"},
	}
	controlPassed := false
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
			if !controlPassed {
				out, err := run("")
				if err != nil || !strings.Contains(string(out), "--- PASS: TestLinuxKernelRefusesThreadAndZombie") {
					t.Fatalf("positive control failed: %v\n%s", err, out)
				}
				controlPassed = true
			}
			path := filepath.Join(root, "internal/homelease/kernel_linux.go")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(raw), "\r\n", "\n")
			if strings.Count(source, tc.anchor) != 1 {
				t.Fatal("mutation anchor not unique")
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, tc.anchor, tc.replacement, 1)), 0o600); err != nil {
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
			out, err := run(overlay)
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: TestLinuxKernelRefusesThreadAndZombie") ||
				!strings.Contains(string(out), tc.message) || strings.Contains(string(out), "panic:") {
				t.Fatalf("plant survived or failed for another reason: %v\n%s", err, out)
			}
			t.Logf("intended expected-red exit 1:\n%s", out)
		})
	}
}
