package scriptworker

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Every current test binary that transitively imports scriptworker must answer
// the hidden mode before running tests, including malformed invocations. The
// process-group watchdog also bounds the unfixed recursion on Unix hosts.
func TestLandlockProbeModeTerminatesAffectedTestBinaries(t *testing.T) {
	for _, pkg := range []string{"cmd/curator", "internal/install", "internal/install/atomicity", "internal/crossconformance", "internal/scriptworker"} {
		t.Run(pkg, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "probe.test")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			compile := exec.Command(testtoolchain.Go(t), "test", "-c", "-o", binary, "../../"+pkg) // #nosec G204 -- fixed repository test packages
			if output, err := compile.CombinedOutput(); err != nil {
				t.Fatalf("build: %v\n%s", err, output)
			}
			owned := mustPhysical(t, t.TempDir())
			for _, tc := range []struct {
				name string
				args []string
				want int
			}{
				{"owned", []string{ScriptLandlockProbeMode, owned}, probeSuccessExit()},
				{"missing", []string{ScriptLandlockProbeMode, filepath.Join(owned, "missing")}, 1},
				{"no-directory", []string{ScriptLandlockProbeMode}, 1},
				{"extra-argument", []string{ScriptLandlockProbeMode, owned, "extra"}, 1},
			} {
				t.Run(tc.name, func(t *testing.T) {
					ctx, cancel := context.WithCancel(t.Context())
					defer cancel()
					command := exec.CommandContext(ctx, binary, tc.args...) // #nosec G204 -- test-owned binary and fixed hidden arguments
					command.Env = scriptWorkerEnvironment()
					protectProbeTestProcess(command)
					output, err := command.CombinedOutput()
					if ctx.Err() != nil {
						t.Fatalf("probe ran tests or hung: %v; output %.500s", ctx.Err(), output)
					}
					code := 0
					var exit *exec.ExitError
					if errors.As(err, &exit) {
						code = exit.ExitCode()
					} else if err != nil {
						t.Fatal(err)
					}
					if code != tc.want || len(output) != 0 {
						t.Fatalf("probe exit=%d output=%.500q, want exit=%d and no test output", code, output, tc.want)
					}
				})
			}
		})
	}
}

func probeSuccessExit() int {
	if runtime.GOOS == "linux" {
		return 0 // The existing Linux probe contract requires a Landlock host.
	}
	return 1
}
