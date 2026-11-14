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

// The unmodified selector admits outside homes and excludes managed homes.
// Narrowing the exclusion to exempt Codex must fail its managed-home witness
// while the adjacent outside-home control still passes. Setup failures do not
// count as detection, and the production source is changed only through an overlay.
func TestNativeHomeManagedCallerNarrowingMutant(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "internal", "envprofile", "native_home.go")
	payload, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(payload), "\r\n", "\n")
	const before = "if !managed {"
	if strings.Count(source, before) != 1 {
		t.Fatal("managed selector mutation anchor is not unique")
	}
	run := func(overlay string) ([]byte, error) {
		args := []string{"test", "-count=1", "-timeout=2m", "-v", "-run", "^TestNativeHomeSelectionBoundaries$/^codex_cli$/^(outside|managed)$", "./internal/envprofile"}
		if overlay != "" {
			args = append([]string{"test", "-overlay", overlay}, args[1:]...)
		}
		command := exec.Command(testtoolchain.Go(t), args...)
		command.Dir = root
		return command.CombinedOutput()
	}
	if output, err := run(""); err != nil ||
		!strings.Contains(string(output), "--- PASS: TestNativeHomeSelectionBoundaries/codex_cli/managed ") ||
		!strings.Contains(string(output), "--- PASS: TestNativeHomeSelectionBoundaries/codex_cli/outside ") {
		t.Fatalf("unmodified controls failed: %v\n%s", err, output)
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "native_home.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(source, before, "if !managed || adapter.ID == \"codex_cli\" {", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := run(overlayPath)
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 ||
		!strings.Contains(string(output), "--- FAIL: TestNativeHomeSelectionBoundaries/codex_cli/managed ") ||
		!strings.Contains(string(output), "selection got ") ||
		!strings.Contains(string(output), "--- PASS: TestNativeHomeSelectionBoundaries/codex_cli/outside ") {
		t.Fatalf("mutant survived or lacked the intended failure/control: %v\n%s", err, output)
	}
	t.Logf("expected-red exit %d:\n%s", exit.ExitCode(), output)
}
