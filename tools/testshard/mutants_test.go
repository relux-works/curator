package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These retain the checks and diagnostic tokens, but admit exactly one
// missing-verdict or failed-test class. The behavioral suites must kill them.
func TestShardGateNarrowingMutants(t *testing.T) {
	lockShardGoTools(t)
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.ReplaceAll(string(source), "\r\n", "\n"))
	original, err := filepath.Abs("main.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, old, killer string }{
		{"missing-example-verdict-admitted", `if !pass {` + "\n" + `                return fmt.Errorf("missing test verdict %s", key)`, "TestShardCLIExecutesEveryCaseBeforeCombinedAcceptance"},
		{"one-failed-test-admitted", `if e.Action == "fail" {` + "\n" + `                return fmt.Errorf("failed event %s/%s", e.Package, e.Test)`, "TestVerifyRequiresEveryShardAndEverySelectedTest/failed-test"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Use exact gofmt-indented anchors, refusing a stale/not-applied plant.
			old := strings.ReplaceAll(tc.old, "                ", "\t\t\t\t")
			var replacement string
			if tc.name == "missing-example-verdict-admitted" {
				replacement = strings.Replace(old, "if !pass {", `if !pass && !strings.HasSuffix(key, "/Example") {`, 1)
			} else {
				replacement = `if e.Action == "fail" && e.Test == "TestB" { e.Action = "pass" }` + "\n\t\t\t" + old
			}
			if strings.Count(string(source), old) != 1 {
				t.Fatalf("mutant anchor count=%d", strings.Count(string(source), old))
			}
			dir := t.TempDir()
			mutant := filepath.Join(dir, "main.go")
			if err := os.WriteFile(mutant, []byte(strings.Replace(string(source), old, replacement, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: mutant}})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(path, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-count=1", "-overlay", path, "-run", `^Test(VerifyRequiresEveryShardAndEverySelectedTest|ShardCLIExecutesEveryCaseBeforeCombinedAcceptance)$`, ".")
			// The parent protects real Go compilation and the child reuses that lock.
			cmd.Env = append(os.Environ(), "CURATOR_TEST_SHARD_PARENT_LOCK=1")
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			// A compiler/setup failure or no named failing test is a survivor.
			named := strings.Contains(string(output), "--- FAIL: "+tc.killer) || strings.Contains(string(output), "--- FAIL: "+strings.Split(tc.killer, "/")[0]) && strings.Contains(string(output), "--- FAIL: "+tc.killer)
			reason := "control: <nil>"
			if tc.name == "one-failed-test-admitted" {
				reason = "error=<nil>, want failed event"
			}
			if !ok || exit.ExitCode() != 1 || !named || !strings.Contains(string(output), reason) {
				t.Fatalf("mutant survived or unrelated failure: %v\n%s", err, output)
			}
			t.Logf("mutant=%s child_exit=1 killer=%s", tc.name, tc.killer)
		})
	}
}
