//go:build !darwin && !linux

package godriver

import "os"

// Windows closes the worker's query-only job handle on exit. After manager
// death this releases the last handle and kill-on-close tears down the job;
// with the manager alive its existing parent-side teardown closes the domain.
func stopOrphanedWorkerDomain() { os.Exit(exitWorkerFailure) }
