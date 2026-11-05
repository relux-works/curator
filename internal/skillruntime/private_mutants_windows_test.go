//go:build windows

package skillruntime

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The routing writer alone must create a protected owner-only Windows DACL.
// Bind has no generation writer, so its private-source assertion is reachable
// independently. The unchanged transition passes; an inherited-DACL creator
// must fail that assertion, never a setup error, skip or generation refusal.
func TestWindowsProfileRoutingPrivateWriterMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool := os.Getenv("CURATOR_GO")
	if goTool == "" {
		goTool = "go"
	}
	run := func(overlay string) ([]byte, error) {
		args := []string{"test", "-count=1", "-timeout=5m", "-v", "./internal/skillruntime", "-run", "^TestBindRequiresCompleteTransition$/^complete$"}
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
		return cmd.CombinedOutput()
	}
	positive, err := run("")
	if err != nil || !strings.Contains(string(positive), "--- PASS: TestBindRequiresCompleteTransition/complete") || strings.Contains(string(positive), "--- SKIP:") {
		t.Fatalf("routing positive control: %v\n%s", err, positive)
	}
	path := filepath.Join(root, "internal", "skillruntime", "commands.go")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(raw), "\r\n", "\n")
	before := "file, err := privatedir.Create(staged, 0600)"
	if strings.Count(source, before) != 1 {
		t.Fatal("routing writer mutation anchor is not unique")
	}
	// The original os.WriteFile 0600 and this creator both inherit the directory
	// DACL. Preserve all subsequent write/close logic and every other creator.
	after := "file, err := os.OpenFile(staged, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)"
	privateImport := `"github.com/relux-works/curator/internal/privatedir"`
	if strings.Count(source, privateImport) != 1 {
		t.Fatal("private creator import anchor is not unique")
	}
	source = strings.Replace(source, privateImport, `"os"`, 1)
	scratch := t.TempDir()
	replacement := filepath.Join(scratch, "commands.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(source, before, after, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlay, payload, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := run(overlay)
	exit, ok := err.(*exec.ExitError)
	text := string(output)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(text, "--- FAIL: TestBindRequiresCompleteTransition/complete") || !strings.Contains(text, "routing journal source is not private:") || !strings.Contains(text, "DACL is not protected from inheritance") || strings.Contains(text, "panic:") || strings.Contains(text, "--- SKIP:") {
		t.Fatalf("routing mutant survived or failed outside its assertion: %v\n%s", err, output)
	}
	t.Logf("intended routing-writer kill, exit 1:\n%s", output)
}
