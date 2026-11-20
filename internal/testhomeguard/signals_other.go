//go:build !unix

package testhomeguard

import "os"

func forwardDiagnosticSignals(*os.Process) func() { return func() {} }

func signaledExit(error) (os.Signal, bool) { return nil, false }

func reraiseSignal(os.Signal) {}
