package scriptworker

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Moving the availability check into sync.Once must fail the repeated-call
// skip witness. Compilation errors or unrelated worker failures do not count.
func TestFixtureBuilderSkipNarrowingMutant(t *testing.T) {
	source, err := filepath.Abs("main_test.go")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	// Git for Windows may check out this Go source with CRLF endings.
	data = []byte(strings.ReplaceAll(string(data), "\r\n", "\n"))
	const before = "func stubInterpreterBinary(t *testing.T) string {\n\tt.Helper()\n\tgoTool := testtoolchain.Go(t)\n\tstubOnce.Do(func() {"
	const after = "func stubInterpreterBinary(t *testing.T) string {\n\tt.Helper()\n\tstubOnce.Do(func() {\n\t\tgoTool := testtoolchain.Go(t)"
	if strings.Count(string(data), before) != 1 {
		t.Fatal("fixture skip mutation anchor is not unique")
	}
	directory := t.TempDir()
	replacement := filepath.Join(directory, "main_test.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(data), before, after, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlay, data, 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(testtoolchain.Go(t), "test", "-overlay", overlay, "-count=1", "-timeout=1m", "-v", "-run", "^TestFixtureBuildersSkipWithoutGo$", ".")
	output, err := command.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 ||
		!strings.Contains(string(output), "--- FAIL: TestFixtureBuildersSkipWithoutGo/interpreter/second") ||
		!strings.Contains(string(output), "instead of skipping") ||
		strings.Contains(string(output), "build failed") {
		t.Fatalf("sync.Once mutant did not fail the intended skip witness: %v\n%s", err, output)
	}
	t.Log("sync.Once mutant: second dependent call returns a fixture path without skipping, exit 1")
}
