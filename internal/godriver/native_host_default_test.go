//go:build !curator_test_platform

package godriver

import (
	"runtime"
	"testing"
)

// A normal release build ignores the test-tag-only host selector.
func TestProductionHostIgnoresPlatformTestSelector(t *testing.T) {
	t.Setenv("CURATOR_TEST_NATIVE_GOOS", "unsupported-host")
	if nativeControlHost() != runtime.GOOS {
		t.Fatal("release host accepts test selector")
	}
}
