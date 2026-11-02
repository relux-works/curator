//go:build unix

package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each overlay keeps the writer/check present and weakens one permission
// class. Only exit 1 with the named behavioral failure counts as a kill.
// These controls exercise Resolve or its real fallback writer, never a
// source-token checker.
func TestManagedModesNarrowingMutants(t *testing.T) {
	isolateDetectorParent(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, file, before, after, run, fails, reason string }{
		{
			"single-fallback-0644", "managed.go",
			"atomicManagedFile(root, fullRel, payload, managedAuthoredFileMode)",
			"atomicManagedFile(root, fullRel, payload, 0o644)",
			"^TestManagedFallbackPermissions$/^tree-false$",
			"TestManagedFallbackPermissions/tree-false", "copy mode = 0644, want 0600",
		},
		{
			"executable-tree-fallback-0755", "managed.go",
			"info.Mode().Perm()&managedFallbackFileMask)",
			"func() os.FileMode { if info.Mode().Perm() == 0o755 { return 0o755 }; return info.Mode().Perm()&managedFallbackFileMask }())",
			"^TestManagedFallbackPermissions$/^tree-true$",
			"TestManagedFallbackPermissions/tree-true", "0755 mode = 0755, want 0700",
		},
		{
			"private-seed-0644", "home_publication.go",
			"seed.payload, seed.mode, false)",
			"seed.payload, func() os.FileMode { if seed.mode == 0o600 { return 0o644 }; return seed.mode }(), false)",
			"^TestManagedSeedPreservesNativePermissions$/^0600$",
			"TestManagedSeedPreservesNativePermissions/0600", "mode = 0644, want 0600",
		},
		{
			"claude-trust-0644", "managed.go",
			"prepared.payload, managedAuthoredFileMode, false)",
			"prepared.payload, 0o644, false)",
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
			"codex-directory-not-tightened", "../transaction/directories.go",
			"if before.Mode != 0700 {",
			"if before.Mode != 0700 && filepath.Base(parent) != \"codex_cli\" {",
			"^TestManagedHomeRepairTightensDirectories$/^codex_cli$",
			"TestManagedHomeRepairTightensDirectories/codex_cli", "repair failed to restore private directories: environment_repair_failed: managed directory permissions:",
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
			replacement := filepath.Join(dir, filepath.Base(mutant.file))
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
			command.Env = detectorChildEnv()
			var buffer bytes.Buffer
			command.Stdout = &buffer
			command.Stderr = &buffer
			if err := command.Start(); err != nil {
				t.Fatalf("cannot-start: %v", err)
			}
			t.Logf("owned child PID=%d", command.Process.Pid)
			err = command.Wait()
			output := buffer.Bytes()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+mutant.fails+" ") || !strings.Contains(string(output), mutant.reason) {
				t.Fatalf("mutant survived or lacked the intended permission failure: %v\n%s", err, output)
			}
			t.Logf("expected-red go test exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
