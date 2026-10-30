// Package testgitenv keeps scratch repositories and native homes independent
// of the operator's repository selectors, Git configuration and SSH agent.
//
// TestMain in every package whose tests run git calls Isolate before
// m.Run. A test that needs its own global config (an insteadOf rewrite, a
// credential helper) still sets GIT_CONFIG_GLOBAL with t.Setenv, which
// overrides the process-wide value for that test. Production code never
// imports this package.
package testgitenv

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Isolate scrubs inherited Git repository/config selectors and SSH_AUTH_SOCK,
// pins HOME/XDG_CONFIG_HOME to private scratch directories, and installs an
// empty global Git config with system config disabled. Native provider and
// XDG data/state/cache root overrides are cleared too. Call before m.Run;
// deliberate per-test overrides remain possible. Cleanup restores the caller's
// managed variables and removes the scratch home. Owned paths and the temporary
// directory variables are absolute and physical. Setup failures fail closed.
func Isolate() func() {
	base, err := physicalPath(os.TempDir())
	if err != nil {
		fmt.Fprintln(os.Stderr, "testgitenv: resolve temporary root:", err)
		os.Exit(2)
	}
	dir, err := os.MkdirTemp(base, "curator-test-gitconfig-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "testgitenv: create isolated git config dir:", err)
		os.Exit(2)
	}
	dir, err = physicalPath(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testgitenv: resolve isolated root:", err)
		os.Exit(2)
	}
	path := filepath.Join(dir, "gitconfig")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "testgitenv: create isolated git config:", err)
		os.Exit(2)
	}
	home := filepath.Join(dir, "home")
	xdg := filepath.Join(dir, "xdg")
	for _, directory := range []string{home, xdg} {
		if err := os.Mkdir(directory, 0o700); err != nil {
			fmt.Fprintln(os.Stderr, "testgitenv: create isolated home:", err)
			os.Exit(2)
		}
	}
	previous := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, _ := strings.Cut(entry, "=")
		if managed(key) {
			previous[key] = value
			mustEnv(os.Unsetenv(key))
		}
	}
	for key, value := range map[string]string{
		"GIT_CONFIG_GLOBAL": path, "GIT_CONFIG_NOSYSTEM": "1",
		"HOME": home, "XDG_CONFIG_HOME": xdg,
		"TMPDIR": base, "TMP": base, "TEMP": base,
	} {
		mustEnv(os.Setenv(key, value))
	}
	// os.UserHomeDir uses USERPROFILE on Windows rather than HOME.
	if runtime.GOOS == "windows" {
		mustEnv(os.Setenv("USERPROFILE", home))
	}
	return func() {
		for _, entry := range os.Environ() {
			key, _, _ := strings.Cut(entry, "=")
			if managed(key) {
				mustEnv(os.Unsetenv(key))
			}
		}
		for key, value := range previous {
			mustEnv(os.Setenv(key, value))
		}
		if err := removeIsolatedHome(dir); err != nil {
			fmt.Fprintln(os.Stderr, "testgitenv: remove isolated home:", err)
			os.Exit(2)
		}
	}
}

// Go's module cache leaves read-only directories under an otherwise private
// HOME. Restore access only on directories in this owned tree. The permission
// walk does not follow symlinks or chmod regular files.
func removeIsolatedHome(dir string) error {
	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return os.Chmod(path, info.Mode()|0o700)
	}); err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func managed(key string) bool {
	key = strings.ToUpper(key)
	if runtime.GOOS == "windows" && key == "USERPROFILE" {
		return true
	}
	if strings.HasPrefix(key, "GIT_CONFIG") {
		return true
	}
	switch key {
	case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY",
		"GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_NAMESPACE", "GIT_TEMPLATE_DIR", "GIT_CEILING_DIRECTORIES",
		"GIT_DISCOVERY_ACROSS_FILESYSTEM", "GIT_IMPLICIT_WORK_TREE", "GIT_GRAFT_FILE", "GIT_SHALLOW_FILE",
		"GIT_NO_REPLACE_OBJECTS", "GIT_REPLACE_REF_BASE", "GIT_PREFIX", "GIT_QUARANTINE_PATH",
		"GIT_REFERENCE_BACKEND", "GIT_DEFAULT_HASH", "GIT_DEFAULT_REF_FORMAT",
		"SSH_AUTH_SOCK", "HOME", "XDG_CONFIG_HOME", "TMPDIR", "TMP", "TEMP",
		"CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME":
		return true
	}
	return false
}

// Resolve before deriving exported paths or cleanup ownership: relative TMPDIR
// and symlinked temporary volumes must not move those paths with a later chdir.
func physicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func mustEnv(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "testgitenv: isolate environment:", err)
		os.Exit(2)
	}
}
