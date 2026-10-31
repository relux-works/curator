package main

import (
	"errors"
	"syscall"
	"testing"
)

// An unavailable owned observer returns EBADF rather than claiming exit. The
// live observer/real original-child completion control is in the shared test.
func TestUmbrellaProviderClosedObserverIsNotCompletion(t *testing.T) {
	observer, err := newProviderExitObserver()
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.Close(); err != nil {
		t.Fatal(err)
	}
	ready, err := observer.Ready()
	if ready || !errors.Is(err, syscall.EBADF) {
		t.Fatalf("closed observer: ready=%v err=%v", ready, err)
	}
}
