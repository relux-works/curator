//go:build unix

package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each overlay keeps the writer/check present and weakens one permission
// class. Only exit 1 with the named behavioral failure counts as a kill.
// These controls exercise Resolve, never a source-token checker.
func TestManagedModesNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, file, before, after, run, fails, reason string }{
		{
			"private-seed-0644", "managed.go",
			"file.payload, file.mode)",
			"file.payload, func() os.FileMode { if file.mode == 0o600 { return 0o644 }; return file.mode }())",
			"^TestManagedSeedPreservesNativePermissions$/^0600$",
			"TestManagedSeedPreservesNativePermissions/0600", "mode = 0644, want 0600",
		},
		{
			"claude-trust-0644", "managed.go",
			"root, path, payload, managedAuthoredFileMode)",
			"root, path, payload, 0o644)",
			"^TestManagedAuthoredFilesArePrivate$",
			"TestManagedAuthoredFilesArePrivate", ".claude.json mode = 0644, want 0600",
		},
		{
			"codex-directory-drift-admitted", "managed.go",
			"if err := privatedir.Validate(path); err != nil {",
			"if err := privatedir.Validate(path); err != nil && adapter.ID != envregistry.CodexCLI {",
			"^TestManagedHomeRepairTightensDirectories$/^codex_cli$",
			"TestManagedHomeRepairTightensDirectories/codex_cli", "wide home must be stale: <nil>",
		},
		{
			"codex-directory-not-tightened", "nofollow.go",
			"if err := privatedir.Protect(current); err != nil {",
			"if err := func() error { if filepath.Base(current) == \"codex_cli\" { return nil }; return privatedir.Protect(current) }(); err != nil {",
			"^TestManagedHomeRepairTightensDirectories$/^codex_cli$",
			"TestManagedHomeRepairTightensDirectories/codex_cli", "managed directory permissions:",
		},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			original := filepath.Join(root, "internal", "envprofile", mutant.file)
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if strings.Count(source, mutant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, mutant.file)
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, mutant.before, mutant.after, 1)), 0o600); err != nil {
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
			command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", mutant.run, "./internal/envprofile")
			command.Dir = root
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+mutant.fails+" ") || !strings.Contains(string(output), mutant.reason) {
				t.Fatalf("mutant survived or lacked the intended permission failure: %v\n%s", err, output)
			}
			t.Logf("expected-red go test exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
