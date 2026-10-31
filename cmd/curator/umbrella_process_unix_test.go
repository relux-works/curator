//go:build darwin || linux

package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Ordinary completion preserves argv, environment, streams and nonzero status.
func TestUmbrellaProviderCompletion(t *testing.T) {
	command := exec.Command("/bin/sh", "-c", `printf '%s:%s' "$1" "$STOP_FIXTURE"; printf warning >&2; exit 23`, "fixture", "literal argument")
	command.Env = []string{"STOP_FIXTURE=explicit"}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := runUmbrellaProvider(command)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 23 {
		t.Fatalf("completion = %v, want exit 23", err)
	}
	if stdout.String() != "literal argument:explicit" || stderr.String() != "warning" {
		t.Fatalf("streams = %q, %q", stdout.String(), stderr.String())
	}
}

// A signal queued before Start reaches the original child rather than disappearing.
func TestUmbrellaProviderQueuedTermination(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "/bin/sh", "-c", "while :; do :; done")
	signals := make(chan os.Signal, 1)
	signals <- syscall.SIGTERM
	err := runProviderWithSignals(command, signals)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || ctx.Err() != nil {
		t.Fatalf("queued termination = %v, context = %v", err, ctx.Err())
	}
	status, ok := exit.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() || status.Signal() != syscall.SIGTERM {
		t.Fatalf("queued termination status = %v", exit.Sys())
	}
}

// A cooperative original child receives TERM and exits through the original Wait.
// This exercises forwarding with an injected notification, not OS subscription delivery.
func TestUmbrellaProviderCooperativeTermination(t *testing.T) {
	signals := make(chan os.Signal, 1)
	command, input, output, result := startUmbrellaStopFixture(t, signals)
	defer input.Close()
	readUmbrellaFixtureReady(t, output)
	signals <- syscall.SIGTERM
	err := <-result
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 42 {
		t.Fatalf("cooperative termination = %v, want exit 42", err)
	}
	if command.ProcessState == nil || !command.ProcessState.Exited() {
		t.Fatal("original child was not waited")
	}
}

// Real Notify delivery to this isolated test process reaches only the supervised
// original child. A separate, also test-owned child completes without TERM.
func TestUmbrellaProviderSubscribedTermination(t *testing.T) {
	_, otherInput, otherOutput, otherResult := startUmbrellaStopFixture(t, make(chan os.Signal))
	defer otherInput.Close()
	readUmbrellaFixtureReady(t, otherOutput)
	_, input, output, result := startUmbrellaStopFixture(t, nil)
	defer input.Close()
	readUmbrellaFixtureReady(t, output)
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	err := <-result
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 42 {
		t.Fatalf("subscribed termination = %v, want exit 42", err)
	}
	if _, err := io.WriteString(otherInput, "complete\n"); err != nil {
		t.Fatal(err)
	}
	if err := <-otherResult; err != nil {
		t.Fatalf("unrelated original fixture was affected: %v", err)
	}
}

// A non-cooperative direct child ends only at the fixture's forced cutoff.
// This proves its non-successful direct-child result, not descendant absence.
func TestUmbrellaProviderNonCooperativeCutoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	input, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	defer writer.Close()
	command := exec.CommandContext(ctx, "/bin/sh", "-c", "trap '' TERM; printf 'ready\\n'; read answer")
	command.Stdin = input
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	signals := make(chan os.Signal)
	result := make(chan error, 1)
	finished := make(chan struct{})
	go func() {
		result <- runProviderWithSignals(command, signals)
		close(finished)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(2 * time.Second):
			t.Error("non-cooperative original fixture wait did not finish")
		}
	})
	readUmbrellaFixtureReady(t, bufio.NewReader(output))
	select {
	case signals <- syscall.SIGTERM:
	case err := <-result:
		t.Fatalf("fixture ended before TERM delivery: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("TERM notification delivery did not finish")
	}
	err = <-result
	var exit *exec.ExitError
	if !errors.As(err, &exit) || !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("forced cutoff = %v, context = %v", err, ctx.Err())
	}
	status, ok := exit.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
		t.Fatalf("forced cutoff status = %v", exit.Sys())
	}
}

type unsupportedUmbrellaSignal struct{}

func (unsupportedUmbrellaSignal) Signal()        {}
func (unsupportedUmbrellaSignal) String() string { return "unsupported fixture signal" }

// Forwarding failure cannot turn a later successful child exit into success.
func TestUmbrellaProviderForwardFailure(t *testing.T) {
	signals := make(chan os.Signal)
	_, input, output, result := startUmbrellaStopFixture(t, signals)
	defer input.Close()
	readUmbrellaFixtureReady(t, output)
	select {
	case signals <- unsupportedUmbrellaSignal{}:
	case err := <-result:
		t.Fatalf("fixture ended before unsupported signal delivery: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("unsupported signal notification delivery did not finish")
	}
	if _, err := io.WriteString(input, "complete\n"); err != nil {
		t.Fatal(err)
	}
	err := <-result
	if err == nil || !strings.Contains(err.Error(), "forward provider signal") {
		t.Fatalf("forwarding failure = %v", err)
	}
}

// Start failure leaves a queued notification unconsumed and starts no child.
func TestUmbrellaProviderStartFailure(t *testing.T) {
	signals := make(chan os.Signal, 1)
	signals <- syscall.SIGTERM
	command := exec.Command("/nonexistent/curator-stop-fixture")
	if err := runProviderWithSignals(command, signals); err == nil {
		t.Fatal("missing executable was accepted")
	}
	if command.Process != nil || len(signals) != 1 {
		t.Fatal("failed Start acquired a child or consumed its pending signal")
	}
}

// Completion stops consuming notifications; the completed handle is not reused.
func TestUmbrellaProviderNoForwardAfterWait(t *testing.T) {
	signals := make(chan os.Signal, 1)
	command := exec.Command("/bin/sh", "-c", "exit 0")
	if err := runProviderWithSignals(command, signals); err != nil {
		t.Fatal(err)
	}
	signals <- syscall.SIGTERM
	if len(signals) != 1 {
		t.Fatal("a forwarding consumer survived the original Wait")
	}
	if err := command.Process.Signal(syscall.SIGTERM); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("completed original handle = %v, want ErrProcessDone", err)
	}
}

func startUmbrellaStopFixture(t *testing.T, signals <-chan os.Signal) (*exec.Cmd, *os.File, *bufio.Reader, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	input, writer, err := os.Pipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	command := exec.CommandContext(ctx, "/bin/sh", "-c", "trap 'exit 42' TERM; printf 'ready\\n'; read answer; exit 0")
	command.Stdin = input
	output, err := command.StdoutPipe()
	if err != nil {
		cancel()
		input.Close()
		writer.Close()
		t.Fatal(err)
	}
	result := make(chan error, 1)
	finished := make(chan struct{})
	go func() {
		if signals == nil {
			result <- runUmbrellaProvider(command)
		} else {
			result <- runProviderWithSignals(command, signals)
		}
		close(finished)
	}()
	t.Cleanup(func() {
		cancel()
		input.Close()
		writer.Close()
		output.Close()
		select {
		case <-finished:
		case <-time.After(2 * time.Second):
			t.Error("original fixture wait did not finish after cancellation")
		}
	})
	return command, writer, bufio.NewReader(output), result
}

func readUmbrellaFixtureReady(t *testing.T, output *bufio.Reader) {
	t.Helper()
	line, err := output.ReadString('\n')
	if err != nil || line != "ready\n" {
		t.Fatalf("fixture readiness = %q, %v", line, err)
	}
}
