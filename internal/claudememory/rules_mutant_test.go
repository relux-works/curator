package claudememory

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Omitting only native rules must fail plain and metacharacter literal scope
// tests plus the legacy-settings upgrade. Malformed-input refusal still passes.
func TestNativeRulesExclusionNarrowingMutant(t *testing.T) {
	source, err := filepath.Abs("settings.go")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	const anchor = "required := []string{nativeFile, nativeRules}"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("native rules mutation anchor is not unique")
	}
	run := func(overlay string) ([]byte, error) {
		args := []string{"test", "-count=1", "-timeout=1m", "-v", "-run", "^TestMerge(NativeMemory|NativeRulesUpgrade|RefusesMalformedSettings)$", "."}
		if overlay != "" {
			args = append([]string{"test", "-overlay", overlay}, args[1:]...)
		}
		return exec.Command(testtoolchain.Go(t), args...).CombinedOutput()
	}
	if output, err := run(""); err != nil {
		t.Fatalf("unmodified rules controls: %v\n%s", err, output)
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "settings.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(data), anchor, "required := []string{nativeFile}; _ = nativeRules", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := run(path)
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestMergeNativeMemory/operator ") || !strings.Contains(string(output), "--- FAIL: TestMergeNativeMemory/operator[1]") || !strings.Contains(string(output), "native rules upgrade missing or duplicate exclusion") || !strings.Contains(string(output), "--- PASS: TestMergeRefusesMalformedSettings ") {
		t.Fatalf("rules plant survived or failed elsewhere: %v\n%s", err, output)
	}
	t.Logf("expected-red exit 1; omitted native rules:\n%s", output)
}
