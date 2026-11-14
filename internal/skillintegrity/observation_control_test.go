package skillintegrity

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSkillHardlinkObservationControl(t *testing.T) {
	file, before, after := "read_unix.go", "stat.Nlink == 1", "stat.Nlink <= 2"
	witness := "hardlink missed named admission refusal"
	if runtime.GOOS == "windows" {
		file, before, after = "info_windows.go", "openCaptureMetadata(path)", "openCaptureFile(path)"
		witness = "identity observation prematurely rejected hardlink"
	}
	runObservationControl(t, file, before, after, "TestSkillHardlinkObservationKeepsNamedRefusal", witness)
}

// This narrowing control proves that the indicated regression is detected at
// its intended assertion, with a valid neighbor in the same child invocation.
func runObservationControl(t *testing.T, file, before, after, test, witness string) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "internal", "skillintegrity", file)
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if strings.Count(text, before) != 1 {
		t.Fatal("observation control anchor must occur once")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "plant.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(text, before, after, 1)), 0600); err != nil {
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
	cmd := exec.Command("go", "test", "-p=1", "-count=1", "-timeout=3m", "-overlay", path, "-v", "-run", "^"+test+"$", "./internal/skillintegrity")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+test+" ") || !strings.Contains(string(output), witness) {
		t.Fatalf("observation control missed intended assertion: %v\n%s", err, output)
	}
	t.Logf("narrowing control reached %s: %s", test, witness)
}
