//go:build unix

package testhomeguard

import (
	"os"
	"os/signal"
	"syscall"
)

// A diagnostic targets the real blocked test child; the observer survives to
// join it and check the postimage. SIGKILL/observer death remain residuals.
func forwardDiagnosticSignals(child *os.Process) func() {
	ch := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(ch, syscall.SIGQUIT)
	go func() {
		select {
		case s := <-ch:
			_ = child.Signal(s)
		case <-done:
		}
	}()
	return func() { signal.Stop(ch); close(done) }
}
