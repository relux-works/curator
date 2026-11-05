//go:build curator_test_platform

package godriver

import (
	"os"
	"runtime"
)

// Only an explicitly tagged test manager admits this selector. Release builds
// never read it, and build package data cannot select the manager's build tags.
func nativeControlHost() string {
	if value := os.Getenv("CURATOR_TEST_NATIVE_GOOS"); value != "" {
		return value
	}
	return runtime.GOOS
}
