package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsNativeExecutablePath(t *testing.T) {
	want, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	got, err := windowsExecutablePath(uint32(os.Getpid()))
	if err != nil || !strings.EqualFold(got, want) {
		t.Fatalf("native executable = %q, %v; want %q", got, err, want)
	}
}

func TestBuildInUseWindowsCase(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Build")
	if !buildInUse(dir, dir, []string{strings.ToUpper(filepath.Join(dir, "bin", "runner.exe"))}) {
		t.Fatal("image path casing hid an in-use build")
	}
}
