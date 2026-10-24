//go:build linux

package scriptworker

import (
	"context"
	"encoding/json"
	"errors"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// A probe-shaped caller, including a malformed request, must report absent
// without re-executing. The normal caller is the positive control on Landlock
// hosts; changing argv here exercises the fallback guard after initialization.
func TestLandlockProbeChildCannotReprobe(t *testing.T) {
	original := os.Args
	defer func() { os.Args = original }()
	if present, err := probeScriptLandlock(ScriptControlDescendantExecDenial); !present || err != nil {
		t.Fatalf("normal caller = %v, %v; requires the Linux Landlock test host", present, err)
	}
	for _, args := range [][]string{
		{original[0], ScriptLandlockProbeMode, t.TempDir()},
		{original[0], ScriptLandlockProbeMode},
		{original[0], ScriptLandlockProbeMode, t.TempDir(), "extra"},
	} {
		os.Args = args
		if present, err := probeScriptLandlock(ScriptControlDescendantExecDenial); present || err != nil {
			t.Fatalf("probe child argc=%d = %v, %v, want absent without error", len(args), present, err)
		}
	}
}

// Each overlay weakens one guard while retaining the guard itself. Run the
// behavioral regression, not a token checker, and require its named failure.
// Overlays leave the checkout untouched and are inherited by nested builds.
func TestLandlockProbeNarrowingMutants(t *testing.T) {
	for _, mutant := range []struct {
		name, file, before, after, run, failure, diagnostic string
	}{
		{
			"dispatch-admits-extra-argument", "landlock_probe.go",
			"if len(os.Args) != 3 {", "if len(os.Args) < 3 {",
			"^TestLandlockProbeModeTerminatesAffectedTestBinaries$/internal/scriptworker/extra-argument$",
			"--- FAIL: TestLandlockProbeModeTerminatesAffectedTestBinaries",
			"probe exit=0 output=\"\", want exit=1 and no test output",
		},
		{
			"reprobe-admits-malformed-child", "landlock_linux.go",
			"if isLandlockProbeChild() {", "if isLandlockProbeChild() && len(os.Args) == 3 {",
			"^TestLandlockProbeChildCannotReprobe$",
			"--- FAIL: TestLandlockProbeChildCannotReprobe",
			"probe child argc=2 = true, <nil>, want absent without error",
		},
		{
			"cleanup-kills-group-only-on-deadline", "landlock_linux.go",
			"err := unix.Kill(-command.Process.Pid, unix.SIGKILL)",
			"if errors.Is(ctx.Err(), context.Canceled) { return command.Process.Kill() }; err := unix.Kill(-command.Process.Pid, unix.SIGKILL)",
			"^TestLandlockProbeCancellationKillsProcessGroup$",
			"--- FAIL: TestLandlockProbeCancellationKillsProcessGroup/cancel",
			"probe descendant survived cancellation",
		},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			source, err := filepath.Abs(mutant.file)
			if err != nil {
				t.Fatal(err)
			}
			payload, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(payload), mutant.before) != 1 {
				t.Fatal("mutant not applied: source anchor must occur exactly once")
			}
			directory := t.TempDir()
			replacement := filepath.Join(directory, mutant.file)
			if err := os.WriteFile(replacement, []byte(strings.Replace(string(payload), mutant.before, mutant.after, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(directory, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(filepath.Join(build.Default.GOROOT, "bin", "go"), "test", "-count=1", "-timeout=60s", "-run", mutant.run, ".") // #nosec G204 -- local toolchain and fixed behavioral tests
			command.Env = append(os.Environ(), "GOFLAGS="+os.Getenv("GOFLAGS")+" "+strconv.Quote("-overlay="+overlayPath))
			output, err := command.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || !strings.Contains(string(output), mutant.failure) || !strings.Contains(string(output), mutant.diagnostic) {
				t.Fatalf("mutant survived or failed outside its named regression: %v\n%s", err, output)
			}
			t.Logf("mutant %s killed: exit=%d; %s; %s", mutant.name, exit.ExitCode(), mutant.failure, mutant.diagnostic)
		})
	}
}

// deadlineCauseContext preserves the actual cancellation cause supplied by the
// timeout created after readiness, including DeadlineExceeded versus Canceled.
// exec.CommandContext observes a stable Done channel throughout process startup.
type deadlineCauseContext struct{ context.Context }

func (ctx deadlineCauseContext) Err() error { return context.Cause(ctx.Context) }

// Both deadline expiry and explicit cancellation kill a real grandchild in
// the production probe command's process group. An isolated subreaper lets the
// test reap it and assert SIGKILL, rather than mistaking a zombie for a survivor.
func TestLandlockProbeCancellationKillsProcessGroup(t *testing.T) {
	const helper = "CURATOR_TEST_LANDLOCK_SUBREAPER"
	if os.Getenv(helper) != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestLandlockProbeCancellationKillsProcessGroup$", "-test.v") // #nosec G204 -- fixed self-test entry
		command.Env = append(os.Environ(), helper+"=1")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("isolated cleanup test: %v\n%s", err, output)
		}
		return
	}
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"timeout", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			directory := t.TempDir()
			manager := filepath.Join(directory, "manager")
			// A private inherited pipe holds the real descendant until the
			// test closes it. A missed group kill then exits normally, giving
			// Wait4 a deterministic wrong status instead of a timed poll.
			release, hold, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer release.Close()
			defer hold.Close()
			if err := os.WriteFile(manager, []byte("#!/bin/sh\n/bin/sh -c 'read release <&3' &\necho $! > \"$2/descendant\"\nwait\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			base, stop := context.WithCancelCause(t.Context())
			ctx := deadlineCauseContext{base}
			cancel := func() { stop(context.Canceled) }
			defer cancel()
			command := landlockProbeCommand(ctx, manager, directory)
			command.ExtraFiles = []*os.File{release}
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			_ = release.Close()
			defer func() { _ = unix.Kill(-command.Process.Pid, unix.SIGKILL) }()
			var pid int
			for t.Context().Err() == nil {
				payload, err := os.ReadFile(filepath.Join(directory, "descendant"))
				if err == nil {
					pid, err = strconv.Atoi(strings.TrimSpace(string(payload)))
					if err == nil && pid > 0 {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
			if pid <= 0 {
				cancel()
				_ = command.Wait()
				t.Fatal("probe fixture did not start its descendant")
			}
			// Start the deadline only after the fixture's real descendant
			// reports readiness; CI startup is not part of this assertion.
			var expired context.Context
			if mode == "cancel" {
				cancel()
			} else {
				var expireCancel context.CancelFunc
				expired, expireCancel = context.WithTimeout(t.Context(), 2*time.Second)
				defer expireCancel()
				stopForwarding := context.AfterFunc(expired, func() { stop(context.Cause(expired)) })
				defer stopForwarding()
			}
			if err := command.Wait(); err == nil || ctx.Err() == nil {
				t.Fatalf("probe did not fail on %s: wait=%v context=%v", mode, err, ctx.Err())
			}
			if mode == "timeout" && !errors.Is(expired.Err(), context.DeadlineExceeded) {
				t.Fatalf("probe ended without deadline expiry: %v", expired.Err())
			}
			// The direct command has joined after cancellation. Close the
			// descendant barrier: a pending SIGKILL must win over normal exit.
			_ = hold.Close()
			var status unix.WaitStatus
			reaped, err := unix.Wait4(pid, &status, 0, nil)
			if err != nil || reaped != pid {
				t.Fatalf("reap probe descendant: pid=%d err=%v", reaped, err)
			}
			if !status.Signaled() || status.Signal() != unix.SIGKILL {
				t.Fatalf("probe descendant survived cancellation: status=%v, want SIGKILL", status)
			}
			if err := unix.Kill(pid, 0); !errors.Is(err, unix.ESRCH) {
				t.Fatalf("descendant still exists after reap: %v", err)
			}
		})
	}
	// A normally exiting probe remains successful with the cancellation hook.
	if err := landlockProbeCommand(context.Background(), "/bin/true", t.TempDir()).Run(); err != nil {
		t.Fatalf("successful probe command: %v", err)
	}
}
