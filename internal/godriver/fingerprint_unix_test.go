//go:build !windows

package godriver

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestFingerprintRejectsSpecialFile(t *testing.T) {
	root := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(root, "pipe"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
	if DiagnosticCode(err) != "special_file_forbidden" {
		t.Fatalf("error = %v", err)
	}
}

func TestFingerprintRejectsInvalidUnicodePath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, string([]byte{0xff}))
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Skipf("filesystem rejects invalid UTF-8 names: %v", err)
	}
	_, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
	if DiagnosticCode(err) != "invalid_unicode" {
		t.Fatalf("error = %v", err)
	}
}

func TestUnixLauncherRequiresExecuteBit(t *testing.T) {
	root := t.TempDir()
	launcher := filepath.Join(root, "bin", platformGoName)
	writeTestFile(t, launcher, testExecutableBytes(), 0o644)
	if err := validateLauncher(launcher); DiagnosticCode(err) != "untrusted_go_executable" {
		t.Fatalf("error = %v", err)
	}
}
