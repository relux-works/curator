//go:build !unix

package testhomeguard

import "os"

func forwardDiagnosticSignals(*os.Process) func() { return func() {} }
