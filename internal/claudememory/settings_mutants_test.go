package claudememory

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Omitting both owned exclusions must fail the independent literal scope
// assertion, while malformed-input refusal still passes. Build failures cannot
// count as a detected omission.
func TestNativeMemoryExclusionNarrowingMutant(t *testing.T) {
	source, err := filepath.Abs("settings.go")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	const anchor = "excludes = append(excludes, requiredPath)"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("native exclusion mutation anchor is not unique")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "settings.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(data), anchor, "// native exclusion deliberately omitted", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, data, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, testtoolchain.Go(t), "test", "-overlay", overlay, "-count=1", "-timeout=1m", "-v", "-run", "^TestMerge(NativeMemory|RefusesMalformedSettings)$", ".")
	output, err := command.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 ||
		!strings.Contains(string(output), "--- FAIL: TestMergeNativeMemory/operator ") ||
		!strings.Contains(string(output), "merge changed unrelated settings or exclusion scope") ||
		!strings.Contains(string(output), "--- PASS: TestMergeRefusesMalformedSettings ") ||
		strings.Contains(string(output), "build failed") {
		t.Fatalf("omitted exclusion did not reach its intended assertion and valid neighbor: %v\n%s", err, output)
	}
	t.Logf("expected-red exit 1; native exclusion omission:\n%s", output)
}
