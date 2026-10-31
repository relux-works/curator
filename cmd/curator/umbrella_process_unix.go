//go:build darwin || linux

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func runUmbrellaProvider(command *exec.Cmd) error {
	// Subscribe before Start so a notification in that region remains queued.
	// This does not cover signals received before dispatch reaches this function.
	signals := make(chan os.Signal, 4)
	signal.Notify(signals, syscall.SIGTERM)
	defer signal.Stop(signals)
	return runProviderWithSignals(command, signals)
}

func runProviderWithSignals(command *exec.Cmd, signals <-chan os.Signal) (result error) {
	// WaitDelay can call Process.Kill inside os/exec without this lifecycle
	// serialization. Umbrella dispatch uses an ordinary Cmd with no WaitDelay.
	if command.WaitDelay != 0 {
		return errors.New("provider supervision does not support WaitDelay")
	}
	observer, err := newProviderExitObserver()
	if err != nil {
		return fmt.Errorf("observe provider exit: %w", err)
	}
	defer func() {
		if err := observer.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("close provider exit observer: %w", err))
		}
	}()
	lifecycle := &providerChildLifecycle{command: command}
	if cancel := command.Cancel; cancel != nil {
		// CommandContext cancellation must not signal after the sole reap starts.
		// Return ErrProcessDone rather than blocking the watcher that Wait joins.
		command.Cancel = func() error { return lifecycle.cancel(cancel) }
	}
	if err := command.Start(); err != nil {
		return err
	}
	if err := observer.Watch(command.Process); err != nil {
		return failProviderObservation(lifecycle, err)
	}
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	var forwardingError error
	for {
		ready, err := observer.Ready()
		if err != nil {
			return errors.Join(forwardingError, failProviderObservation(lifecycle, err))
		}
		if ready {
			// Observation did not reap. Retire signal authority before the one
			// real Cmd.Wait, preserving its ProcessState and stream cleanup.
			lifecycle.beginWait()
			return providerWaitResult(command.Wait(), forwardingError)
		}
		select {
		case <-tick.C:
		case received, ok := <-signals:
			if !ok {
				signals = nil
				continue
			}
			// The original child is still unreaped: no Wait runs in parallel.
			err := lifecycle.signal(received, (*os.Process).Signal)
			if err != nil && !errors.Is(err, os.ErrProcessDone) && forwardingError == nil {
				forwardingError = fmt.Errorf("forward provider signal: %w", err)
			}
		}
	}
}

func failProviderObservation(lifecycle *providerChildLifecycle, observationError error) error {
	// Observation failed after Start. Clean up only the still-original,
	// unreaped child and preserve the failure, even if its actual Wait succeeds.
	killError := lifecycle.signal(os.Kill, (*os.Process).Signal)
	if errors.Is(killError, os.ErrProcessDone) {
		killError = nil
	}
	lifecycle.beginWait()
	return errors.Join(fmt.Errorf("observe provider exit: %w", observationError), killError, lifecycle.command.Wait())
}

func providerWaitResult(waitError, forwardingError error) error {
	if forwardingError != nil {
		return errors.Join(waitError, forwardingError)
	}
	// Preserve the concrete ExitError used by the umbrella exit-code mapping.
	return waitError
}
