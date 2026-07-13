package runtimestore

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

func lay(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInstallRuntimeRootsOncePerCommit(t *testing.T) {
	snapshot := t.TempDir()
	lay(t, snapshot, map[string]string{"scripts/tool": "#!/bin/sh\n", "scripts/lib/util.py": "x"})
	home := t.TempDir()

	dir, err := InstallRuntimeRoots(home, "skill-a", "abc123", snapshot, []string{"scripts"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "scripts", "lib", "util.py")); err != nil {
		t.Fatal("runtime tree incomplete")
	}
	// reuse: marker file survives a second install
	probe := filepath.Join(dir, "probe")
	if err := os.WriteFile(probe, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := InstallRuntimeRoots(home, "skill-a", "abc123", snapshot, []string{"scripts"})
	if err != nil || again != dir {
		t.Fatalf("reuse failed: %v %v", again, err)
	}
	if _, err := os.Stat(probe); err != nil {
		t.Fatal("existing runtime entry was rebuilt")
	}
}

func TestRuntimeCommandPath(t *testing.T) {
	snapshot := t.TempDir()
	lay(t, snapshot, map[string]string{"scripts/tool": "#!/bin/sh\n"})
	home := t.TempDir()
	if _, err := InstallRuntimeRoots(home, "skill-a", "c1", snapshot, []string{"scripts"}); err != nil {
		t.Fatal(err)
	}
	command := skillspec.Command{Name: "tool", Type: "script", UnixPath: "scripts/tool", WinPath: "scripts/tool"}
	path, err := RuntimeCommandPath(home, "skill-a", "c1", command, "unix")
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Fatal("runtime command not executable")
	}
	if _, err := RuntimeCommandPath(home, "skill-a", "c1", skillspec.Command{Name: "gone", UnixPath: "scripts/gone"}, "unix"); err == nil {
		t.Fatal("missing runtime file must fail")
	}
}

func TestInstallSingleCommandAndShims(t *testing.T) {
	snapshot := t.TempDir()
	lay(t, snapshot, map[string]string{"run.sh": "#!/bin/sh\necho hi\n"})
	home := t.TempDir()
	command := skillspec.Command{Name: "run", Type: "script", UnixPath: "run.sh", WinPath: "run.sh"}

	runtimePath, err := InstallSingleCommand(home, "skill-b", "c2", snapshot, command, "unix")
	if err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	if runtime.GOOS != "windows" {
		shim, err := WriteBinShim(binDir, "run", runtimePath, "unix", nil)
		if err != nil {
			t.Fatal(err)
		}
		target, err := os.Readlink(shim)
		if err != nil {
			t.Fatalf("unix shim must be a symlink: %v", err)
		}
		if filepath.IsAbs(target) {
			t.Fatalf("symlink must be relative: %s", target)
		}
	}

	// windows shim shape is testable on any platform
	winShim, err := WriteBinShim(binDir, "run", runtimePath, "windows", nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(winShim)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(winShim) != "run.cmd" || string(payload[:9]) != "@echo off" {
		t.Fatalf("windows shim shape: %s %q", winShim, payload)
	}
}

func TestRemoveStaleShims(t *testing.T) {
	binDir := t.TempDir()
	for _, name := range []string{"keep", "drop"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := RemoveStaleShims(binDir, map[string]bool{"keep": true}, "unix"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(binDir, "keep")); err != nil {
		t.Fatal("expected shim removed")
	}
	if _, err := os.Stat(filepath.Join(binDir, "drop")); err == nil {
		t.Fatal("stale shim survived")
	}

	// windows: .cmd suffix maps back to the command name
	winDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(winDir, "tool.cmd"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := RemoveStaleShims(winDir, map[string]bool{"tool": true}, "windows"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(winDir, "tool.cmd")); err != nil {
		t.Fatal("windows shim wrongly removed")
	}
}

func TestUnixLauncherCarriesRuntimePathArgumentsAndExitStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executes a POSIX launcher")
	}
	root := t.TempDir()
	helperBin := filepath.Join(root, "helper's bin")
	inheritedBin := filepath.Join(root, "inherited bin")
	if err := os.MkdirAll(helperBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(inheritedBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(helperBin, "declared-helper"), []byte("#!/bin/sh\necho declared\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inheritedBin, "inherited-helper"), []byte("#!/bin/sh\necho inherited\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	runtimeDir := filepath.Join(root, "runtime's dir")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	runtimePath := filepath.Join(runtimeDir, "tool")
	if err := os.WriteFile(runtimePath, []byte("#!/bin/sh\ndeclared-helper\ninherited-helper\nprintf 'arg=%s\\n' \"$1\"\nexit 23\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	shim, err := WriteBinShim(filepath.Join(root, "project bin"), "tool", runtimePath, "unix", []string{helperBin})
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(shim); err != nil || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("environment launcher must be an executable file: %v, %v", info, err)
	}
	command := exec.Command(shim, "value with spaces")
	command.Env = []string{"PATH=" + inheritedBin}
	output, runErr := command.CombinedOutput()
	exitErr, ok := runErr.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 23 {
		t.Fatalf("launcher exit = %v, output:\n%s", runErr, output)
	}
	for _, expected := range []string{"declared", "inherited", "arg=value with spaces"} {
		if !strings.Contains(string(output), expected) {
			t.Fatalf("launcher output lacks %q:\n%s", expected, output)
		}
	}
}

func TestWindowsLauncherCarriesRuntimePathAndExitStatus(t *testing.T) {
	binDir := t.TempDir()
	runtimePath := filepath.Join(t.TempDir(), "runtime%root", "tool.cmd")
	helperBin := filepath.Join(t.TempDir(), "helper%bin")
	shim, err := WriteBinShim(binDir, "tool", runtimePath, "windows", []string{helperBin})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(shim)
	if err != nil {
		t.Fatal(err)
	}
	content := string(payload)
	for _, expected := range []string{
		"setlocal DisableDelayedExpansion",
		`set "PATH=` + strings.ReplaceAll(helperBin, "%", "%%") + `;%PATH%"`,
		`call "` + strings.ReplaceAll(runtimePath, "%", "%%") + `" %*`,
		"exit /b %ERRORLEVEL%",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("Windows launcher lacks %q:\n%s", expected, content)
		}
	}
}
