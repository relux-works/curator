// Package runtimestore keeps command runtimes once per machine, keyed by
// skill and commit, and writes command shims (Spec §8.6).
package runtimestore

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/skillspec"
)

// Platform selects shim and path behavior; "unix" or "windows".
func Platform() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}
	return "unix"
}

// Dir returns the runtime store location for a skill at a commit.
func Dir(home, skillName, commit string) string {
	return filepath.Join(home, "runtime", skillName, commit)
}

// InstallRuntimeRoots copies the declared runtime roots of a snapshot into
// the store, atomically and once per commit: an existing entry is reused.
func InstallRuntimeRoots(home, skillName, commit, snapshot string, runtimeRoots []string) (string, error) {
	target := Dir(home, skillName, commit)
	if _, err := os.Stat(target); err == nil {
		return target, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	tmp := filepath.Join(filepath.Dir(target), fmt.Sprintf(".%s.tmp-%d", commit, os.Getpid()))
	if err := os.RemoveAll(tmp); err != nil {
		return "", err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }
	for _, root := range runtimeRoots {
		src := filepath.Join(snapshot, filepath.FromSlash(root))
		info, err := os.Stat(src)
		if err != nil || !info.IsDir() {
			cleanup()
			return "", fmt.Errorf("runtime root not found: %s", root)
		}
		if err := copyTree(src, filepath.Join(tmp, filepath.FromSlash(root))); err != nil {
			cleanup()
			return "", err
		}
	}
	if _, err := os.Stat(target); err == nil {
		cleanup()
		return target, nil
	}
	if err := os.Rename(tmp, target); err != nil {
		cleanup()
		return "", err
	}
	return target, nil
}

// InstallSingleCommand copies one command file (a skill without runtime
// roots) into the store under bin/ and returns the runtime path.
func InstallSingleCommand(home, skillName, commit, snapshot string, command skillspec.Command, platform string) (string, error) {
	rel := commandRel(command, platform)
	if rel == "" {
		return "", fmt.Errorf("command %q has no path for %s", command.Name, platform)
	}
	src := filepath.Join(snapshot, filepath.FromSlash(rel))
	info, err := os.Stat(src)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("command %q source file not found: %s", command.Name, rel)
	}
	suffix := ""
	if platform == "windows" && !strings.HasSuffix(command.Name, ".cmd") {
		suffix = ".cmd"
	}
	runtimePath := filepath.Join(Dir(home, skillName, commit), "bin", command.Name+suffix)
	if err := os.MkdirAll(filepath.Dir(runtimePath), 0o755); err != nil {
		return "", err
	}
	if err := copyFile(src, runtimePath); err != nil {
		return "", err
	}
	if platform != "windows" {
		if err := makeExecutable(runtimePath); err != nil {
			return "", err
		}
	}
	return runtimePath, nil
}

// RuntimeCommandPath returns the path of a command inside installed runtime
// roots, ensuring it exists and is executable on unix.
func RuntimeCommandPath(home, skillName, commit string, command skillspec.Command, platform string) (string, error) {
	rel := commandRel(command, platform)
	if rel == "" {
		return "", fmt.Errorf("command %q has no path for %s", command.Name, platform)
	}
	runtimePath := filepath.Join(Dir(home, skillName, commit), filepath.FromSlash(rel))
	info, err := os.Stat(runtimePath)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("command %q runtime file not found: %s", command.Name, rel)
	}
	if platform != "windows" {
		if err := makeExecutable(runtimePath); err != nil {
			return "", err
		}
	}
	return runtimePath, nil
}

// WriteBinShim writes the project or global launcher for a command. When
// pathEntries is non-empty, the launcher carries the runtime environment and
// does not depend on shell activation or a user profile.
func WriteBinShim(binDir, commandName, runtimePath, platform string, pathEntries []string) (string, error) {
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", err
	}
	if platform == "windows" {
		shim := filepath.Join(binDir, commandName+".cmd")
		content := WindowsShimContent(runtimePath, pathEntries)
		if err := os.WriteFile(shim, []byte(content), 0o755); err != nil {
			return "", err
		}
		return shim, nil
	}
	shim := filepath.Join(binDir, commandName)
	if _, err := os.Lstat(shim); err == nil {
		if err := os.Remove(shim); err != nil {
			return "", err
		}
	}
	if len(pathEntries) > 0 {
		prefix := strings.Join(pathEntries, ":")
		content := "#!/bin/sh\n" +
			"if [ -n \"${PATH:-}\" ]; then\n" +
			"  PATH=" + shellQuote(prefix) + ":\"$PATH\"\n" +
			"else\n" +
			"  PATH=" + shellQuote(prefix) + "\n" +
			"fi\n" +
			"export PATH\n" +
			"exec " + shellQuote(runtimePath) + " \"$@\"\n"
		if err := os.WriteFile(shim, []byte(content), 0o755); err != nil {
			return "", err
		}
		if err := makeExecutable(shim); err != nil {
			return "", err
		}
		return shim, nil
	}
	rel, err := filepath.Rel(binDir, runtimePath)
	if err != nil {
		rel = runtimePath
	}
	if err := os.Symlink(rel, shim); err != nil {
		return "", err
	}
	return shim, nil
}

// WindowsShimContent returns the exact managed Windows launcher bytes. It is
// shared with forwarding-shim ownership checks.
func WindowsShimContent(runtimePath string, pathEntries []string) string {
	content := "@echo off\r\nsetlocal DisableDelayedExpansion\r\n"
	if len(pathEntries) > 0 {
		escaped := make([]string, 0, len(pathEntries))
		for _, entry := range pathEntries {
			escaped = append(escaped, escapeCMDValue(entry))
		}
		content += "set \"PATH=" + strings.Join(escaped, ";") + ";%PATH%\"\r\n"
	}
	content += "call \"" + escapeCMDValue(runtimePath) + "\" %*\r\n"
	content += "exit /b %ERRORLEVEL%\r\n"
	return content
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func escapeCMDValue(value string) string {
	return strings.ReplaceAll(value, "%", "%%")
}

// RemoveStaleShims deletes shims whose command is no longer expected.
func RemoveStaleShims(binDir string, expected map[string]bool, platform string) error {
	entries, err := os.ReadDir(binDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		command := name
		if platform == "windows" && strings.HasSuffix(strings.ToLower(name), ".cmd") {
			command = name[:len(name)-len(".cmd")]
		}
		if !expected[command] {
			if err := os.Remove(filepath.Join(binDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func commandRel(command skillspec.Command, platform string) string {
	if platform == "windows" {
		return command.WinPath
	}
	return command.UnixPath
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type in runtime root: %s", rel)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	payload, err := os.ReadFile(src) // #nosec G304 -- paths come from the walked snapshot
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, payload, info.Mode().Perm())
}

func makeExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.Chmod(path, info.Mode().Perm()|0o111)
}
