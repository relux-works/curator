//go:build !darwin && !linux

package godriver

import "os"

// Windows closes the query handle during admission, before acknowledgement.
// The manager then holds the last Job Object handle; its death alone triggers
// kill-on-close. Worker-owned termination is not claimed on this platform.
func stopOrphanedWorkerDomain()       { os.Exit(exitWorkerFailure) }
func protectWorkerPipeWrites() func() { return func() {} }
func finishWorkerDomain()             {}
