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

// Each overlay weakens one reviewed boundary for a bounded production witness.
// The unchanged witness must first pass. A kill requires exit 1 and the named
// assertion, not a compiler error, skipped test, timeout or setup failure.
func TestNativeHomeReviewNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct{ name, file, before, after, run, fail, message string }{
		{"xdg-main-refusal", "managed.go", "xdg := req.operatorXDG()\n\tif xdg ==", "xdg := req.operatorXDG(); if req.EnvID == envregistry.OpenCode { home, _ := os.UserHomeDir(); xdg = filepath.Join(home, \".config\") }\n\tif xdg ==", "^TestNativeHomeReviewRepair$/^xdg-managed$", "TestNativeHomeReviewRepair/xdg-managed", "named native refusal"},
		{"xdg-child-alias", "managed.go", "if managed {\n\t\treturn \"\", nativeHomeRefusal(target", "if managed && name != \"git\" {\n\t\treturn \"\", nativeHomeRefusal(target", "^TestNativeHomeReviewRepair$/^xdg-child-alias$", "TestNativeHomeReviewRepair/xdg-child-alias", "named native refusal"},
		{"managed-default-diagnostic", "native_home.go", "return \"\", false, fmt.Errorf(\"%s: native home %s is inside managed environments; run repair with the intended external native home\", envregistry.DiagCredentialConflict, native)", "return \"\", true, fmt.Errorf(\"managed default refused\")", "^TestNativeHomeReviewRepair$/^all-managed$", "TestNativeHomeReviewRepair/all-managed", "named native refusal"},
		{"codex-parent-traversal", "native_home.go", "if part == \"..\" {", "if part == \"..\" && !strings.HasSuffix(path, \"codex_cli\") {", "^TestNativeHomeReviewRepair$/^parent-traversal$", "TestNativeHomeReviewRepair/parent-traversal", "named native refusal"},
		{"unreported-fallback", "switch.go", "if err == nil && fallback {", "if err == nil && fallback && adapter.ID != \"codex_cli\" {", "^TestNativeHomeReviewUnreportedFallbackRefuses$", "TestNativeHomeReviewUnreportedFallbackRefuses", "silent fallback admitted"},
		{"purge-refusal", "switch.go", "native, err := NativeHome(adapter)\n\t\tif err != nil {\n\t\t\treturn err", "native, err := NativeHome(adapter)\n\t\tif err != nil {\n\t\t\tcontinue", "^TestNativeHomeReviewUnreportedFallbackRefuses$", "TestNativeHomeReviewUnreportedFallbackRefuses", "purge hid selector refusal"},
		{"global-notice", "global.go", "result.Warnings = append(result.Warnings, resolved.Warnings...)", "result.Warnings = nil", "^TestNativeHomeReviewGlobalNotice$", "TestNativeHomeReviewGlobalNotice", "global fallback notice lost"},
		{"manager-ancestor", "native_home.go", "candidate, err := nativeComparablePath(path)", "if path == managerHome { return false, nil }; candidate, err := nativeComparablePath(path)", "^TestNativeHomeReviewRepair$/^native-ancestor$", "TestNativeHomeReviewRepair/native-ancestor", "named native refusal"},
		{"absent-case", "native_home.go", "if foldedOverlap {", "if foldedOverlap && false {", "^TestNativeHomeReviewRepair$/^absent-root-case$", "TestNativeHomeReviewRepair/absent-root-case", "named native refusal"},
		{"muse-bypass", "managed.go", "adapter, ok = museNativeAdapter(), true", "return filepath.Join(req.operatorXDG(), \"muse\"), nil", "^TestNativeHomeReviewRepair$/^muse-managed$", "TestNativeHomeReviewRepair/muse-managed", "Muse shared fallback"},
	} {
		t.Run(m.name, func(t *testing.T) {
			original := filepath.Join(root, "internal", "envprofile", m.file)
			b, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(b), "\r\n", "\n")
			if strings.Count(source, m.before) != 1 {
				t.Fatalf("mutation anchor count = %d", strings.Count(source, m.before))
			}
			run := func(overlay string) ([]byte, error) {
				args := []string{"test", "-count=1", "-timeout=2m", "-v", "-run", m.run, "./internal/envprofile"}
				if overlay != "" {
					args = append([]string{"test", "-overlay", overlay}, args[1:]...)
				}
				cmd := exec.Command(testtoolchain.Go(t), args...)
				cmd.Dir = root
				return cmd.CombinedOutput()
			}
			if output, err := run(""); err != nil || !strings.Contains(string(output), "--- PASS: "+m.fail+" ") {
				t.Fatalf("unmodified witness: %v\n%s", err, output)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, m.file)
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, m.before, m.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(path, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			output, err := run(path)
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+m.fail+" ") || !strings.Contains(string(output), m.message) {
				t.Fatalf("plant survived or failed outside its assertion: %v\n%s", err, output)
			}
			t.Logf("expected-red exit 1:\n%s", output)
		})
	}
}
