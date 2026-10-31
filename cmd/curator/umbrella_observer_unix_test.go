//go:build darwin || linux

package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// Observing an already-ending original child leaves its real exit23 available
// to the one Cmd.Wait. EOF exposes the early-exit/registration boundary; this
// does not simulate kernel PID reuse or certify every scheduler interleaving.
func TestUmbrellaProviderObserverPreservesOriginalWait(t *testing.T) {
	observer, err := newProviderExitObserver()
	if err != nil {
		t.Fatal(err)
	}
	defer observer.Close()
	command := exec.Command("/bin/sh", "-c", "exit 23")
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			command.Wait()
		}
	}()
	if _, err := io.ReadAll(output); err != nil {
		t.Fatal(err)
	}
	if err := observer.Watch(command.Process); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		ready, err := observer.Ready()
		if err != nil {
			t.Fatal(err)
		}
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("original child exit was not observed")
		}
		time.Sleep(time.Millisecond)
	}
	waited = true
	err = command.Wait()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 23 || command.ProcessState == nil {
		t.Fatalf("original Wait = %v, state = %v", err, command.ProcessState)
	}
}

// The real ownership seam holds its mutex during signal/cancel callbacks and
// rejects both after reaping begins. Callbacks are inert; no fake PID is signaled.
func TestUmbrellaProviderLifecycleRetiresSignalAuthority(t *testing.T) {
	command := exec.Command("/bin/sh", "-c", "exit 0")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer command.Wait()
	lifecycle := &providerChildLifecycle{command: command}
	calls := 0
	deliver := func(process *os.Process, signal os.Signal) error {
		calls++
		if process != command.Process || signal != syscall.SIGTERM {
			t.Error("original process handle or signal changed")
		}
		if lifecycle.mu.TryLock() {
			lifecycle.mu.Unlock()
			t.Error("delivery ran outside the lifecycle lock")
		}
		return nil
	}
	if err := lifecycle.signal(syscall.SIGTERM, deliver); err != nil || calls != 1 {
		t.Fatalf("owned signal: calls=%d err=%v", calls, err)
	}
	cancelCalls := 0
	cancelError := errors.New("inert original cancellation")
	cancel := func() error {
		cancelCalls++
		if lifecycle.mu.TryLock() {
			lifecycle.mu.Unlock()
			t.Error("cancellation ran outside the lifecycle lock")
		}
		return cancelError
	}
	if err := lifecycle.cancel(cancel); !errors.Is(err, cancelError) || cancelCalls != 1 {
		t.Fatalf("owned cancellation: calls=%d err=%v", cancelCalls, err)
	}
	lifecycle.beginWait()
	if err := lifecycle.signal(syscall.SIGTERM, deliver); !errors.Is(err, os.ErrProcessDone) || calls != 1 {
		t.Fatalf("retired signal: calls=%d err=%v", calls, err)
	}
	if err := lifecycle.cancel(cancel); !errors.Is(err, os.ErrProcessDone) || cancelCalls != 1 {
		t.Fatalf("retired cancellation: calls=%d err=%v", cancelCalls, err)
	}
}

// A real observer failure path kills/waits only its still-original child and
// retains the specific failure instead of presenting successful completion.
func TestUmbrellaProviderObservationFailureRemainsFailure(t *testing.T) {
	input, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	defer writer.Close()
	command := exec.Command("/bin/sh", "-c", "read answer")
	command.Stdin = input
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	observationError := errors.New("inert observer refusal")
	lifecycle := &providerChildLifecycle{command: command}
	err = failProviderObservation(lifecycle, observationError)
	if !errors.Is(err, observationError) || command.ProcessState == nil || !lifecycle.reaping {
		t.Fatalf("observation failure = %v, state=%v, reaping=%v", err, command.ProcessState, lifecycle.reaping)
	}
	status, ok := command.ProcessState.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
		t.Fatalf("owned failed-observer cleanup status = %v", command.ProcessState.Sys())
	}
	if err := lifecycle.signal(syscall.SIGTERM, (*os.Process).Signal); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("signal after failed observation cleanup = %v", err)
	}
}

// WaitDelay's internal Kill would bypass the ownership seam; reject it before
// Start and leave the pending notification intact. Ordinary Cmd is covered above.
func TestUmbrellaProviderWaitDelayRefusedBeforeStart(t *testing.T) {
	command := exec.Command("/bin/sh", "-c", "exit 0")
	command.WaitDelay = time.Second
	signals := make(chan os.Signal, 1)
	signals <- syscall.SIGTERM
	err := runProviderWithSignals(command, signals)
	if err == nil || err.Error() != "provider supervision does not support WaitDelay" || command.Process != nil || len(signals) != 1 {
		t.Fatalf("WaitDelay refusal = %v, process=%v, pending=%d", err, command.Process, len(signals))
	}
}
