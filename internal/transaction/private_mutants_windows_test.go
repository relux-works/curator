//go:build windows

package transaction

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: each private-publication narrowing plant reaches its intended DACL
// regression assertion. An unchanged child passes first; setup failures,
// missing anchors, skips and unrelated child failures never count.
func TestWindowsPrivatePublicationNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool := os.Getenv("CURATOR_GO")
	if goTool == "" {
		goTool = "go"
	}
	positive := exec.Command(goTool, "test", "-count=1", "-timeout=2m", "-v", "./internal/transaction", "-run", "^TestWindowsTransactionPrivatePublication$")
	positive.Dir = root
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") {
			positive.Env = append(positive.Env, kv)
		}
	}
	if out, err := positive.CombinedOutput(); err != nil || !strings.Contains(string(out), "--- PASS: TestWindowsTransactionPrivatePublication") {
		t.Fatalf("positive publication control: %v\n%s", err, out)
	}
	for _, tc := range []struct{ name, file, before, after, run, assertion string }{
		{"private-file-creator", "internal/privatedir/privatedir.go", "return createPrivateNamedFile(path, mode)", "return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, mode)", "^TestWindowsTransactionPrivatePublication$/^bytes$", "published file lost private DACL"},
		{"transaction-file-copy", "internal/transaction/private_windows.go", "return privatedir.Create(path, mode)", "return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)", "^TestWindowsTransactionPrivatePublication$/^bytes$", "published file lost private DACL"},
		{"transaction-directory-copy", "internal/transaction/private_windows.go", "return privatedir.Make(path)", "return os.Mkdir(path, 0o700)", "^TestWindowsTransactionPrivatePublication$/^tree$", "published directory lost private DACL"},
		{"journal-file-creation", "internal/transaction/journal.go", "privatedir.CreateTemp(engine.journalRoot, \".journal-*.tmp\")", "os.CreateTemp(engine.journalRoot, \".journal-*.tmp\")", "^TestWindowsTransactionPrivatePublication$/^bytes$", "journal file lost private DACL"},
		{"journal-directory-creation", "internal/transaction/journal.go", "privatedir.Make(path)", "os.Mkdir(path, 0o700)", "^TestWindowsTransactionPrivatePublication$/^bytes$", "journal directory lost private DACL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
			goTool := os.Getenv("CURATOR_GO")
			if goTool == "" {
				goTool = "go"
			}
			cmd := exec.Command(goTool, "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "./internal/transaction", "-run", tc.run)
			for _, kv := range os.Environ() {
				if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, kv)
				}
			}
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestWindowsTransactionPrivatePublication") || !strings.Contains(string(output), tc.assertion) {
				t.Fatalf("mutant survived or lacked intended failure: %v\n%s", err, output)
			}
			t.Logf("expected-red exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
