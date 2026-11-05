//go:build curator_cold_repository && linux

package testcli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/godriver"
)

// RunColdManager allows one production worker budget plus five minutes for
// full transactional publication. The target-host probe reached journal fsync
// after compilation and exceeded the former two-minute harness cap. This is a
// test watchdog, not an install-level product limit. Only this owned manager
// receives a Go stack dump shortly before the watchdog expires.
func RunColdManager(t *testing.T, dir string, env []string, name string, args ...string) (int, string, string) {
	t.Helper()
	deadline := time.Now().Add(godriver.DefaultBuildTimeout() + 5*time.Minute)
	if outer, ok := t.Deadline(); ok && outer.Add(-2*time.Second).Before(deadline) {
		deadline = outer.Add(-2 * time.Second)
	}
	ctx, cancel := context.WithDeadline(t.Context(), deadline)
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("cold manager elapsed=%s (worker and transactional publication)", time.Since(started)) }()
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
	timer := time.AfterFunc(max(time.Until(deadline)-10*time.Second, 0), func() { defer close(done); _ = command.Process.Signal(syscall.SIGQUIT) })
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
