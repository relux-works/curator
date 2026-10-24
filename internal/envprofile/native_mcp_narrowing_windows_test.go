//go:build windows

package envprofile

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
)

// Each overlay removes one Windows requirement and must fail its named
// regression assertion. A compile failure, timeout or unapplied plant is not
// evidence. The checkout is unchanged; child tests use disposable fixtures.
func TestMCPWindowsNarrowingControls(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, file, old, replacement, pkg, test, witness string
	}{
		{"private-publication", "internal/envprofile/native_mcp.go", "PrivateFile: true", "PrivateFile: false", "./internal/envprofile", "^TestProfileMCPPrivatePublishedFiles$/^private$", "production file is not private"},
		{"skill-privacy", "internal/envprofile/native_mcp.go", "if e := privatedir.ValidateFile(path); e != nil {", "if e := privatedir.ValidateFile(path); false && e != nil {", "./internal/envprofile", "^TestProfileMCPPrivateSkillPreimage$/^foreign-read$", "want mcp_native_writer_authority_unavailable"},
		{"writable-mode", "internal/envprofile/native_mcp.go", "mode = 0o666", "mode = 0o600", "./internal/envprofile", "^TestProfileMCPPrivateSkillPreimage$/^private$", "identical private preimage"},
		{"frozen-native-path", "internal/envprofile/native_mcp.go", "path, err := filepath.Abs(filepath.FromSlash(path))", "_, err := filepath.Abs(filepath.FromSlash(path))", "./internal/envprofile", "^TestProfileMCPSignedTagReplay$/^valid$", "invalid frozen read input"},
		{"source-privacy", "internal/transaction/engine.go", "if err := privatedir.ValidateFile(source); err != nil {", "if err := privatedir.ValidateFile(source); false && err != nil {", "./internal/transaction", "^TestPrivateFilePublication$/^commit$/^source$", "shared source accepted"},
		{"staged-privacy", "internal/transaction/staging.go", "if target.PrivateFile {\n\t\tif err := privatedir.ValidateFile(path); err != nil {", "if false && target.PrivateFile {\n\t\tif err := privatedir.ValidateFile(path); err != nil {", "./internal/transaction", "^TestPrivateFilePublication$/^recovery$/^staged$", "shared staged file accepted"},
		{"clone-initialization", "internal/gitops/gitops.go", "runFn(\"\", \"-c\", \"core.longpaths=true\", \"clone\",", "runFn(\"\", \"clone\",", "./internal/gitops", "^TestCloneLongDestination$/^initialization$", "Filename too long"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, filepath.FromSlash(tc.file))
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// A Windows checkout may use CRLF; plants describe Go source lines.
			body = []byte(strings.ReplaceAll(string(body), "\r\n", "\n"))
			if strings.Count(string(body), tc.old) != 1 {
				t.Fatal("narrowing plant must match exactly once")
			}
			dir := t.TempDir()
			plant := filepath.Join(dir, "plant.go")
			if err := os.WriteFile(plant, []byte(strings.Replace(string(body), tc.old, tc.replacement, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			replacements := map[string]string{path: plant}
			// Source privacy has two independent pre-copy checks. Remove both
			// for this control so a surviving check cannot mask the plant.
			if tc.name == "source-privacy" {
				staging := filepath.Join(root, "internal", "transaction", "staging.go")
				body, err := os.ReadFile(staging)
				if err != nil {
					t.Fatal(err)
				}
				body = []byte(strings.ReplaceAll(string(body), "\r\n", "\n"))
				old := "if err := privatedir.ValidateFile(source); err != nil {"
				if strings.Count(string(body), old) != 1 {
					t.Fatal("second source privacy plant must match exactly once")
				}
				plant := filepath.Join(dir, "staging.go")
				if err := os.WriteFile(plant, []byte(strings.Replace(string(body), old, "if err := privatedir.ValidateFile(source); false && err != nil {", 1)), 0o600); err != nil {
					t.Fatal(err)
				}
				replacements[staging] = plant
			}
			overlay, err := json.Marshal(struct{ Replace map[string]string }{replacements})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			// The executable and arguments are fixed toolchain/test inputs.
			cmd := exec.CommandContext(ctx, goExecutable, "test", "-count=1", "-timeout=3m", "-overlay", overlayPath, tc.pkg, "-run", tc.test, "-v") // #nosec G204 -- selected Go toolchain, bounded fixture tests
			cmd.Dir = root
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				if !strings.EqualFold(key, "GIT_DIR") && !strings.EqualFold(key, "GIT_WORK_TREE") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			output, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), tc.witness) || !strings.Contains(string(output), "--- FAIL: ") {
				if len(output) > 8192 {
					output = output[len(output)-8192:]
				}
				t.Fatalf("narrowing control did not reach intended assertion: error=%v context=%v\n%s", err, ctx.Err(), output)
			}
		})
	}
}
