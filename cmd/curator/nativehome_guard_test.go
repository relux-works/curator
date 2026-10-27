package main

import (
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/testhomeguard"
	"os"
	"testing"
)

var testNativeHomeGuard *testhomeguard.Guard

// All test/helper exits preserve the package's native-home postimage check.
func exitGuardedTests(code int) {
	if testNativeHomeGuard != nil {
		code = testNativeHomeGuard.Finish(code)
	}
	os.Exit(code)
}

// Both child controls exercise this package's actual TestMain exit boundary.
func TestNativeHomeGuard(t *testing.T) {
	// Reuse the parent's prepared Go toolchain; taking its lock again would
	// deadlock the control child before it reaches the package-exit guard.
	testhomeguard.Control(t, "TestNativeHomeGuardChild", cliHelperProcess+"=1")
}

// This named child resolves native defaults only within synthetic fixture homes.
func TestNativeHomeGuardChild(t *testing.T) {
	testhomeguard.Child(t, func() (string, error) {
		return envprofile.NativeHome(envprofile.Adapter{ID: "claude_code", EnvVar: "CLAUDE_CONFIG_DIR", DefaultDir: ".claude", Target: "CLAUDE.md"})
	})
}
