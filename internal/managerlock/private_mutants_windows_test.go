//go:build windows

package managerlock

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: each manager creation/owner/share narrowing plant reaches its intended
// regression assertion. An unchanged child passes first; setup failures,
// missing anchors, skips and unrelated child failures never count.
func TestWindowsPrivateManagerNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, file, before, after, pkg, run, assertion string }{
		{"manager-boundary", "internal/managerlock/open_windows.go", "privatedir.Prepare(filepath.Join(path, filepath.FromSlash(rel)))", "privatedir.MakeAll(filepath.Join(path, filepath.FromSlash(rel)))", "managerlock", "^TestWindowsPrivateLockReplacement$", "manager directory lost private DACL"},
		{"lock-creation", "internal/managerlock/open_windows.go", "privatedir.Create(path, 0o600)", "os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)", "managerlock", "^TestWindowsPrivateLockReplacement$", "manager lock lost private DACL"},
		{"lock-delete-sharing", "internal/managerlock/open_windows.go", "windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE", "windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE", "managerlock", "^TestWindowsPrivateLockReplacement$", "Windows lock replacement witness unavailable"},
		{"explicit-foreign-grant", "internal/privatedir/privatedir_windows.go", "if sid == nil || !sid.IsValid() || (!owner.Equals(sid) && !user.User.Sid.Equals(sid)) {", "if false && (sid == nil || !sid.IsValid() || (!owner.Equals(sid) && !user.User.Sid.Equals(sid))) {", "privatedir", "^TestWindowsPrepareOwnedInheritanceAndForeignGrant$/^unprotected$", "explicit foreign manager grant was normalized"},
		{"protected-descriptor", "internal/privatedir/privatedir_windows.go", "if control&windows.SE_DACL_PROTECTED != 0 {", "if false && control&windows.SE_DACL_PROTECTED != 0 {", "privatedir", "^TestWindowsPreparePreservesProtectedInsufficientRights$", "protected insufficient rights were normalized"},
		{"foreign-owner", "internal/privatedir/privatedir_windows.go", "return owner != nil && owner.IsValid() && ((effective != nil && owner.Equals(effective)) || (defaultOwner != nil && owner.Equals(defaultOwner)))", "return true", "privatedir", "^TestWindowsPrivateDirectoryOwnerScope$", "foreign directory owner admitted"},
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
