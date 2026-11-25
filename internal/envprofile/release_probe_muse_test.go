package envprofile

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

const museProbeFixtureEnv = "CURATOR_ENVPROFILE_MUSE_PROBE_TEST_HELPER"

func init() {
	executableName := strings.TrimSuffix(strings.ToLower(filepath.Base(os.Args[0])), ".exe")
	if os.Getenv(museProbeFixtureEnv) != "1" || executableName != "muse" {
		return
	}
	if len(os.Args) != 2 || os.Args[1] != "--version" {
		os.Exit(2)
	}
	_, _ = os.Stdout.Write([]byte("Muse version 9.8.7\n"))
	os.Exit(0)
}

func TestMuseReleaseDetectionUsesHarnessProbePolicy(t *testing.T) {
	registered, err := envregistry.ByID(envregistry.Muse)
	if err != nil {
		t.Fatal(err)
	}
	if len(registered.Probe) != 0 {
		t.Fatalf("legacy probe tuple remains: %v", registered.Probe)
	}
	t.Run("registered-probe-parses-version", func(t *testing.T) {
		dir := t.TempDir()
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		alias := filepath.Join(dir, "muse")
		if runtime.GOOS == "windows" {
			alias += ".exe"
		}
		copyProbeExecutable(t, executable, alias)
		t.Setenv(museProbeFixtureEnv, "1")
		t.Setenv("PATH", dir)
		if got := detectRelease(registered); got != "9.8.7" {
			t.Fatalf("release from registered harness probe = %q, want 9.8.7", got)
		}
	})
	t.Run("missing-runtime-remains-unknown", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if got := detectRelease(registered); got != "unknown" {
			t.Fatalf("missing runtime release = %q, want unknown", got)
		}
	})
}

func copyProbeExecutable(t *testing.T, sourcePath, destinationPath string) {
	t.Helper()
	source, err := os.Open(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := source.Close(); err != nil {
			t.Errorf("close source: %v", err)
		}
	}()
	info, err := source.Stat()
	if err != nil {
		t.Fatal(err)
	}
	destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(destination, source); err != nil {
		_ = destination.Close()
		t.Fatal(err)
	}
	if err := destination.Close(); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(destinationPath, info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
	}
}
