// Package processcontext binds manager operations to terminal and service signals.
package processcontext

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Notify cancels on Ctrl-C or termination. Workers dispatch before this manager
// boundary and own their separate process-domain liveness monitor.
func Notify(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}
