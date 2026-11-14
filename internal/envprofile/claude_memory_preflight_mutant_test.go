package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Removing only the early settings preflight must expose real pending recovery
// writes before malformed-input refusal. Both valid-settings recovery controls
// still pass; setup failures and unrelated child failures are not a kill.
func TestClaudeNativeMemoryPreflightNarrowingMutant(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "internal", "envprofile", "managed.go")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	const anchor = "if _, err := prepareClaudeSettings(root, homeRel); err != nil {\n\t\t\treturn err\n\t\t}"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("early preflight mutation anchor is not unique")
	}
	run := func(overlay string) ([]byte, error) {
		args := []string{"test", "-count=1", "-timeout=2m", "-v", "-run", "^TestReviewPendingJournal$/^(referenced|monolithic)$/^(settings-invalid|unchanged)$", "./internal/envprofile"}
		if overlay != "" {
			args = append([]string{"test", "-overlay", overlay}, args[1:]...)
		}
		cmd := exec.Command(testtoolchain.Go(t), args...)
		cmd.Dir = root
		return cmd.CombinedOutput()
	}
	if output, err := run(""); err != nil {
		t.Fatalf("unmodified preflight controls: %v\n%s", err, output)
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "managed.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(data), anchor, "// early settings preflight deliberately omitted", 1)), 0600); err != nil {
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
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "early refusal ran recovery") {
		t.Fatalf("early preflight plant survived or failed elsewhere: %v\n%s", err, output)
	}
	for _, form := range []string{"referenced", "monolithic"} {
		if !strings.Contains(string(output), "--- FAIL: TestReviewPendingJournal/"+form+"/settings-invalid ") || !strings.Contains(string(output), "--- PASS: TestReviewPendingJournal/"+form+"/unchanged ") {
			t.Fatalf("missing intended refusal/control for %s:\n%s", form, output)
		}
	}
	t.Logf("expected-red exit 1; early settings refusal ran recovery:\n%s", output)
}
