package testcli

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Both build entry points use the configured Go executable even with no Go
// on PATH. The resulting real manager binaries execute their version command.
func TestBuildsUseSelectedGoWithoutAmbientGo(t *testing.T) {
	goTool := testtoolchain.Go(t)
	t.Setenv("CURATOR_GO", goTool)
	t.Setenv("PATH", t.TempDir())
	// This harness test compiles a portable binary, without a C compiler.
	t.Setenv("CGO_ENABLED", "0")
	paths := []string{Binary(t)}
	output := filepath.Join(t.TempDir(), "curator"+testExecutableSuffix())
	GoBuild(t, ModuleRoot(t), output, nil)
	paths = append(paths, output)
	for _, path := range paths {
		if output, err := exec.Command(path, "--version").CombinedOutput(); err != nil || len(output) == 0 {
			t.Fatalf("built manager version: %v, %q", err, output)
		}
	}
}

// Missing Go skips each dependent call, even after a builder has cached a
// result. It must not poison the cache or return an unvalidated empty path.
func TestBuildsSkipWithoutGo(t *testing.T) {
	t.Setenv("CURATOR_GO", filepath.Join(t.TempDir(), "absent-go"))
	for _, name := range []string{"binary-first", "binary-second", "cross-build"} {
		skipped := false
		t.Run(name, func(t *testing.T) {
			defer func() { skipped = t.Skipped() }()
			if name == "cross-build" {
				GoBuild(t, ModuleRoot(t), filepath.Join(t.TempDir(), "unused"), nil)
			} else {
				_ = Binary(t)
			}
			t.Error("missing toolchain returned from a build helper instead of skipping")
		})
		if !skipped {
			t.Error("dependent build call did not skip")
		}
	}
}

func testExecutableSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
