//go:build tmphygienecontrols

package pnpmsource

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Isolated overlays prove deletion and capture-store-only bypass plants fail the real-body cleanup witness by name.
func TestPNPMTemporaryTreeCleanupControls(t *testing.T) {
	newConcretePNPMRunner(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "internal/pnpmsource/conformance_test.go")
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	old := `if err := makeTreeWritable(root); err != nil && !os.IsNotExist(err) {`
	for _, p := range []struct{ name, replacement string }{
		{"remove-restore", `if err := error(nil); err != nil && !os.IsNotExist(err) {`},
		{"capture-only-bypass", `if err := func() error { if _, err := os.Stat(filepath.Join(root, "capture-store")); err == nil { return nil }; return makeTreeWritable(root) }(); err != nil && !os.IsNotExist(err) {`},
	} {
		t.Run(p.name, func(t *testing.T) {
			if strings.Count(string(raw), old) != 1 {
				t.Fatal("plant target absent or ambiguous")
			}
			scratch := writableTempDir(t)
			childTmp := filepath.Join(scratch, "child-tmp")
			if err := os.Mkdir(childTmp, 0o700); err != nil {
				t.Fatal(err)
			}
			replacement := filepath.Join(scratch, "conformance_test.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(string(raw), old, p.replacement, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			// Paths are JSON encoded to keep the overlay independent of shell quoting.
			overlay := filepath.Join(scratch, "overlay.json")
			data, err := json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(overlay, data, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-overlay", overlay, "-count=1", "-v", "./internal/pnpmsource", "-run", "^TestPNPMTemporaryTreeRemovable$")
			cmd.Dir = root
			cmd.Env = append(os.Environ(), "TMPDIR="+childTmp, "TMP="+childTmp, "TEMP="+childTmp)
			out, runErr := cmd.CombinedOutput()
			exit, ok := runErr.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: TestPNPMTemporaryTreeRemovable") || !strings.Contains(string(out), "non-writable entry after body cleanup") {
				t.Fatalf("survivor or invalid plant: %v: %s", runErr, out)
			}
			t.Logf("%s killed by TestPNPMTemporaryTreeRemovable; child exit 1 (expected red)", p.name)
		})
	}
}
