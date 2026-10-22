//go:build darwin || linux

package transaction

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each overlay weakens one exchange contract/layout gate. Only exit 1 with
// the matching behavioral assertion counts; setup errors cannot satisfy it.
func TestExchangeNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, file, before, after, run, assertion string }{
		{"present-contract", "journal.go", "if target.Kind == KindExchange && (", "if false && target.Kind == KindExchange && (", "^TestExchangePreparationContract$", "wrong contract refusal"},
		{"preinstall-layout", "exchange.go", "if current != target.PreimageDigest || backup != target.DesiredDigest {", "if false && (current != target.PreimageDigest || backup != target.DesiredDigest) {", "^TestExchangeUnknownBackupPreserved$/^backup$", "refusal changed published or unknown bytes"},
		{"spare-sidecars", "exchange.go", "} else if digest != DigestAbsent {", "} else if false && digest != DigestAbsent {", "^TestExchangeUnknownBackupPreserved$/^(desired|rollback)$", "refusal changed published or unknown bytes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, "internal/transaction", tc.file)
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(body), tc.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(string(body), tc.before, tc.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("env", "-u", "GIT_DIR", "-u", "GIT_WORK_TREE", "go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "./internal/transaction", "-run", tc.run)
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestExchange") || !strings.Contains(string(output), tc.assertion) {
				t.Fatalf("mutant survived or lacked intended failure: %v\n%s", err, output)
			}
			t.Logf("expected-red exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
