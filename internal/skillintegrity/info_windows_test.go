//go:build windows

package skillintegrity

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The earlier handle observation remains bound to the original file after a
// same-byte replacement. A later single-link file cannot borrow that identity;
// the retained original is the positive neighbor. Native Windows only.
func TestWindowsSkillIdentityRetainedBeforeReplacement(t *testing.T) {
	root, _ := fixture(t)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "SKILL.md")
	for _, entry := range entries {
		if entry.Name() != "SKILL.md" {
			continue
		}
		info, err := captureEntryInfo(path, entry)
		if err != nil {
			t.Fatal(err)
		}
		if !singleLink(path, info) {
			t.Fatal("intact identity neighbor refused")
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		old := filepath.Join(root, "old.md")
		if err := os.Rename(path, old); err != nil {
			t.Fatal(err)
		}
		writeFixture(t, root, "SKILL.md", bytes, 0644)
		if singleLink(path, info) {
			t.Fatal("replacement borrowed earlier identity")
		}
		if !singleLink(old, info) {
			t.Fatal("retained original identity refused")
		}
		return
	}
	t.Fatal("fixture entry missing")
}

// A lazy path-backed observation must fail at the replacement assertion, not
// through setup/build failure. The same test verifies the intact neighbor.
func TestWindowsSkillIdentityControl(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "internal/skillintegrity/info_windows.go")
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if strings.Count(text, "return f.Stat()") != 1 {
		t.Fatal("identity plant anchor missing")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "identity.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(text, "return f.Stat()", "return entry.Info()", 1)), 0600); err != nil {
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
	cmd := exec.Command("go", "test", "-p=1", "-count=1", "-timeout=3m", "-overlay", path, "-v", "-run", "^TestWindowsSkillIdentityRetainedBeforeReplacement$", "./internal/skillintegrity")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestWindowsSkillIdentityRetainedBeforeReplacement ") || !strings.Contains(string(output), "replacement borrowed earlier identity") {
		t.Fatalf("identity control missed intended assertion: %v\n%s", err, output)
	}
}
