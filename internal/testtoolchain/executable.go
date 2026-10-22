package testtoolchain

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
)

// Go selects the explicitly configured Go executable or the test binary's
// GOROOT. It never searches PATH; only tests requiring a real Go tool skip
// when that selection is absent.
func Go(t testing.TB) string {
	t.Helper()
	path, err := goExecutable(os.Getenv, build.Default.GOROOT)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("real Go toolchain unavailable: %v", err)
		}
		t.Fatalf("invalid real Go toolchain selection: %v", err)
	}
	return path
}

func goExecutable(getenv func(string) string, defaultRoot string) (string, error) {
	path := getenv("CURATOR_GO")
	if path == "" {
		root := getenv("GOROOT")
		if root == "" {
			root = defaultRoot
		}
		if root == "" {
			return "", fmt.Errorf("set CURATOR_GO or GOROOT; the test binary has no GOROOT: %w", os.ErrNotExist)
		}
		path = filepath.Join(root, "bin", executableName("go"))
	}
	return requireExecutable(path)
}

// Rustc selects the same pinned rustup installation used by the native Cargo
// fixtures, rather than whichever rustc happens to be on PATH.
func Rustc(t testing.TB, version string) string {
	t.Helper()
	current, err := user.Current()
	if err != nil {
		t.Skipf("pinned Rust toolchain unavailable: %v", err)
	}
	path, err := rustcExecutable(current.HomeDir, version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("pinned Rust toolchain unavailable: %v", err)
		}
		t.Fatalf("invalid pinned Rust toolchain selection: %v", err)
	}
	return path
}

func rustcExecutable(home, version, goos, goarch string) (string, error) {
	arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[goarch]
	osName := map[string]string{"darwin": "apple-darwin", "linux": "unknown-linux-gnu", "windows": "pc-windows-msvc"}[goos]
	if home == "" || version == "" || arch == "" || osName == "" {
		return "", fmt.Errorf("no pinned Rust selection for %s/%s: %w", goos, goarch, os.ErrNotExist)
	}
	name := "rustc"
	if goos == "windows" {
		name += ".exe"
	}
	return requireExecutable(filepath.Join(home, ".rustup", "toolchains", version+"-"+arch+"-"+osName, "bin", name))
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func requireExecutable(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("selected executable must be absolute")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("selected executable is unavailable: %w", err)
	}
	if !info.Mode().IsRegular() || (runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0) {
		return "", fmt.Errorf("selected executable is not a regular executable file")
	}
	return path, nil
}
