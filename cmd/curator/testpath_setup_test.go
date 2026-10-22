package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// isolateCLIPath drops the caller's provider and toolchain directories. Tests
// may prepend their own fixtures; helper CLI invocations keep their explicitly
// supplied PATH because TestMain dispatches them before this setup.
func isolateCLIPath() (func(), error) {
	directory, err := os.MkdirTemp("", "curator-cli-path-")
	if err != nil {
		return nil, err
	}
	prior, present := os.LookupEnv("PATH")
	restore := func() {
		if present {
			_ = os.Setenv("PATH", prior)
		} else {
			_ = os.Unsetenv("PATH")
		}
		_ = os.RemoveAll(directory)
	}
	if err := os.Setenv("PATH", cliTestPath(directory)); err != nil {
		restore()
		return nil, err
	}
	return restore, nil
}

func cliTestPath(directory string) string {
	dirs := []string{directory}
	if runtime.GOOS == "windows" {
		// cmd.exe and the standard Git for Windows command directory are the
		// system tools these fixtures need. No caller PATH entries are kept.
		if root := os.Getenv("SystemRoot"); root != "" {
			dirs = append(dirs, filepath.Join(root, "System32"))
		}
		if root := os.Getenv("ProgramFiles"); root != "" {
			dirs = append(dirs, filepath.Join(root, "Git", "cmd"))
		}
	} else {
		dirs = append(dirs, "/usr/bin", "/bin")
	}
	return strings.Join(dirs, string(os.PathListSeparator))
}
