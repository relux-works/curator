//go:build windows

package projectcontext

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: each registration record/staging privacy narrowing plant reaches its intended
// regression assertion. An unchanged child passes first; setup failures,
// missing anchors, skips and unrelated child failures never count.
func TestWindowsPrivateRegistrationNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, file, before, after, pkg, run, assertion string }{
		{"record-file", "internal/projectcontext/store.go", "privatedir.CreateTemp(root, \".write-*\")", "os.CreateTemp(root, \".write-*\")", "projectcontext", "^TestWindowsPrivateRegistrationRecord$", "registration record lost private DACL"},
		{"record-root-guard", "internal/projectcontext/store.go", "if err := pathboundary.ValidateRoot(root); err != nil {", "if err := pathboundary.ValidateRoot(root); false && err != nil {", "projectcontext", "^TestWindowsPrivateRegistrationRecord$", "broadened record root got wrong refusal"},
		{"snapshot-root-guard", "internal/projectcontext/store.go", "privatedir.Prepare(stage)", "privatedir.MakeAll(stage)", "projectcontext", "^TestWindowsPrivateRegistrationSnapshotStage$", "foreign snapshot staging root got wrong refusal"},
		{"snapshot-files", "internal/projectcontext/store.go", "return privatedir.ProtectTree(stage)", "return nil", "projectcontext", "^TestWindowsPrivateRegistrationSnapshotStage$", "registration stage file lost private DACL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			goTool := os.Getenv("CURATOR_GO")
			if goTool == "" {
				goTool = "go"
			}
			positive := exec.Command(goTool, "test", "-count=1", "-timeout=2m", "-v", "./internal/"+tc.pkg, "-run", tc.run)
			positive.Dir = root
			for _, kv := range os.Environ() {
				if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") {
					positive.Env = append(positive.Env, kv)
				}
			}
			if out, err := positive.CombinedOutput(); err != nil || !strings.Contains(string(out), "--- PASS: TestWindows") || strings.Contains(string(out), "--- SKIP:") {
				t.Fatalf("positive manager control: %v\n%s", err, out)
			}

			path := filepath.Join(root, tc.file)
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
			cmd := exec.Command(goTool, "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "./internal/"+tc.pkg, "-run", tc.run)
			for _, kv := range os.Environ() {
				if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, kv)
				}
			}
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestWindows") || !strings.Contains(string(output), tc.assertion) {
				t.Fatalf("mutant survived or lacked intended failure: %v\n%s", err, output)
			}
			t.Logf("expected-red exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
