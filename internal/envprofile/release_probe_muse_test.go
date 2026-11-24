package envprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

const museProbeFixtureEnv = "CURATOR_ENVPROFILE_MUSE_PROBE_TEST_HELPER"

func init() {
	if os.Getenv(museProbeFixtureEnv) != "1" || filepath.Base(os.Args[0]) != "muse" {
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
		if runtime.GOOS == "windows" {
			t.Skip("the executable alias fixture uses a POSIX symlink")
		}
		dir := t.TempDir()
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(executable, filepath.Join(dir, "muse")); err != nil {
			t.Fatal(err)
		}
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
