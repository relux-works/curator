package scriptworker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/scriptpolicy"
)

// TestWindowsRealInterpretersRunDeclaredExec qualifies the Windows manager
// path with the runner's real Python and Node images. Each interpreter must
// resolve the declared cmd.exe through the manager-built PATH farm; the final
// case removes the operator binding and proves the real Python image is not
// launched through ambient PATH.
func TestWindowsRealInterpretersRunDeclaredExec(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only control in script-worker-v1-native-control-inventory-v1 is exercised on the Windows runner")
	}
	managerEnv, err := parseHostEnvironment(os.Environ(), runtime.GOOS)
	if err != nil {
		t.Fatalf("cannot read the manager environment for Windows exec resolution: %v", err)
	}
	systemRoot, present := managerEnv.lookup("SYSTEMROOT")
	if !present || systemRoot == "" {
		t.Fatalf("manager process environment has no SYSTEMROOT; default exec search dirs = %q", DefaultExecSearchDirs())
	}
	wantCmd := filepath.Join(systemRoot, "System32", "cmd.exe")
	if cmdInfo, statErr := os.Lstat(wantCmd); statErr == nil {
		cmdHasMultipleLinks, linkErr := godriver.HasMultipleLinks(wantCmd, cmdInfo)
		t.Logf("manager System32 cmd.exe multiple-link/reparse check=%t (error=%v)", cmdHasMultipleLinks, linkErr)
	} else {
		t.Logf("cannot inspect manager System32 cmd.exe link state: %v", statErr)
	}
	searchDirs := defaultExecSearchDirs(runtime.GOOS, os.Environ())
	if len(searchDirs) == 0 || !strings.EqualFold(filepath.Clean(searchDirs[0]), filepath.Clean(filepath.Dir(wantCmd))) {
		t.Fatalf("manager SYSTEMROOT=%q; default exec search dirs = %q, want System32 first (%q)", systemRoot, searchDirs, filepath.Dir(wantCmd))
	}
	t.Logf("manager SYSTEMROOT=%q; exec search dirs=%q; declared cmd.exe target=%q", systemRoot, searchDirs, wantCmd)
	interpreters := []struct {
		id      string
		file    string
		entry   string
		marker  string
		program func(marker string) string
	}{
		{
			id: "python3-v1", file: "python.exe", entry: "tool.py", marker: "curator-r5-python-declared-exec",
			program: func(marker string) string {
				return fmt.Sprintf("import subprocess\nresult = subprocess.run([\"cmd.exe\", \"/d\", \"/c\", \"echo\", %q], check=True, capture_output=True, text=True)\nprint(result.stdout.strip())\n", marker)
			},
		},
		{
			id: "node-v1", file: "node.exe", entry: "tool.js", marker: "curator-r5-node-declared-exec",
			program: func(marker string) string {
				return fmt.Sprintf("const { execFileSync } = require('node:child_process');\nconst output = execFileSync('cmd.exe', ['/d', '/c', 'echo', %q], { encoding: 'utf8' });\nprocess.stdout.write(output.trim() + '\\n');\n", marker)
			},
		},
	}
	paths := make(map[string]string, len(interpreters))
	for _, interpreter := range interpreters {
		path, err := exec.LookPath(interpreter.file)
		if err != nil {
			t.Fatalf("Windows qualification requires the runner's real %s: %v", interpreter.file, err)
		}
		physical, err := filepath.EvalSymlinks(path)
		if err != nil {
			t.Fatalf("cannot resolve the runner's %s path %q: %v", interpreter.file, path, err)
		}
		absolute, err := filepath.Abs(physical)
		if err != nil {
			t.Fatalf("cannot make the runner's %s path absolute: %v", interpreter.file, err)
		}
		paths[interpreter.id] = absolute
	}

	for _, interpreter := range interpreters {
		t.Run(interpreter.id, func(t *testing.T) {
			fixture := newLaunchFixture(t, os.Args[0])
			fixture.request.InterpreterID = interpreter.id
			fixture.request.Interpreters = map[string]string{interpreter.id: paths[interpreter.id]}
			fixture.request.HostEnvironment = os.Environ()
			fixture.request.ExecSearchDirs = nil // exercise the production System32/SystemRoot search list
			fixture.request.ProjectRoot = fixture.root
			fixture.declareCapabilities(t, `{"exec":["cmd.exe"],"filesystem":"repo"}`)
			writeTestFile(t, fixture.entry, []byte(interpreter.program(interpreter.marker)), 0o644)

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			result, err := Launch(ctx, fixture.request)
			if err != nil {
				t.Fatalf("real %s launch refused: %v", interpreter.file, err)
			}
			if result.ExitCode != 0 {
				t.Fatalf("real %s exit code = %d, stderr %q", interpreter.file, result.ExitCode, result.Stderr)
			}
			if !strings.Contains(string(result.Stdout), interpreter.marker) {
				t.Fatalf("real %s did not execute declared cmd.exe from the manager PATH farm: stdout %q, stderr %q", interpreter.file, result.Stdout, result.Stderr)
			}
			resolved := result.Report.ResolvedExec["cmd.exe"]
			if resolved == "" {
				t.Fatalf("real %s run did not record the manager-resolved cmd.exe; manager SYSTEMROOT=%q search dirs=%q farm entries=%q", interpreter.file, systemRoot, searchDirs, result.Report.FarmEntries)
			}
			if !filepath.IsAbs(resolved) || !strings.EqualFold(filepath.Clean(resolved), filepath.Clean(wantCmd)) {
				t.Fatalf("real %s resolved cmd.exe to %q, want absolute %q", interpreter.file, resolved, wantCmd)
			}
			farmHasCmd := false
			for _, entry := range result.Report.FarmEntries {
				if strings.EqualFold(entry, "cmd.exe") {
					farmHasCmd = true
					break
				}
			}
			if !farmHasCmd {
				t.Fatalf("real %s PATH farm entries %q omit cmd.exe", interpreter.file, result.Report.FarmEntries)
			}
			t.Logf("real %s resolved cmd.exe=%q through PATH farm entries=%q", interpreter.file, resolved, result.Report.FarmEntries)
		})
	}

	t.Run("unbound-python-refuses-before-worker", func(t *testing.T) {
		fixture := newLaunchFixture(t, os.Args[0])
		fixture.request.InterpreterID = "python3-v1"
		fixture.request.Interpreters = map[string]string{}
		fixture.request.HostEnvironment = os.Environ()
		fixture.request.ProjectRoot = fixture.root
		marker := filepath.Join(fixture.root, "unbound-python-ran")
		writeTestFile(t, fixture.entry, []byte(fmt.Sprintf("from pathlib import Path\nPath(%q).write_text('ran')\n", marker)), 0o644)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		_, err := Launch(ctx, fixture.request)
		if DiagnosticCode(err) != scriptpolicy.ControlUnavailable {
			t.Fatalf("unbound real python error = %v, want %s", err, scriptpolicy.ControlUnavailable)
		}
		if err == nil || !strings.Contains(err.Error(), "python3-v1") {
			t.Fatalf("unbound refusal %v does not name python3-v1", err)
		}
		if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
			t.Fatalf("unbound Python marker check = %v, want no interpreter run", statErr)
		}
	})
}
