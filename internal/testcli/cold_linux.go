//go:build curator_cold_repository && linux

package testcli

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// RunColdManager keeps the ordinary two-minute harness ceiling and requests a
// Go stack dump from this owned manager before timeout. It does not signal
// other processes or change the production budget. Cold CI can then identify
// the actual blocked install phase instead of guessing from a silent timeout.
func RunColdManager(t *testing.T, dir string, env []string, name string, args ...string) (int, string, string) {
	t.Helper()
	ctx, cancel := fixtureContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...) // #nosec G204 -- test-only invocation of the freshly built manager binary
	configureFixtureProcess(command)
	command.Dir = dir
	command.Env = append(os.Environ(), env...)
	command.WaitDelay = time.Second
	defer cleanupFixtureProcess(command)
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	timer := time.AfterFunc(110*time.Second, func() { defer close(done); _ = command.Process.Signal(syscall.SIGQUIT) })
	err := command.Wait()
	if !timer.Stop() {
		<-done
	}
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("cold manager: %v\n%s", err, stderr.String())
	}
	return exit.ExitCode(), stdout.String(), stderr.String()
}
