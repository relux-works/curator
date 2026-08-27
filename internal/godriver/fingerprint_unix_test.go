//go:build !windows

package godriver

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestFingerprintImplementationMatchesRC4ToolchainVector(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "bin", "go"), []byte("GO"), 0o755)
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../bin/go", filepath.Join(root, "pkg", "tool-link")); err != nil {
		t.Fatal(err)
	}
	digest, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
	if err != nil {
		t.Fatal(err)
	}
	if digest != "sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e" {
		t.Fatalf("digest = %s", digest)
	}

	if err := os.Chmod(filepath.Join(root, "bin", "go"), 0o555); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(root, "bin", "go"), time.Unix(1, 0), time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
	afterMetadata, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
	if err != nil {
		t.Fatal(err)
	}
	if afterMetadata != digest {
		t.Fatalf("mode/time changed digest: %s != %s", afterMetadata, digest)
	}
}

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
