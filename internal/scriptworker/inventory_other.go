//go:build !unix && !windows

package scriptworker

// The script inventory covers Linux, macOS, and Windows. On any other host
// probeScriptInventory refuses before the worker starts, so these entry
// points exist only to keep the package buildable and must never be
// reached.

// inventoryPlatformUsable refuses every test-only platform override: no
// inventory mechanism exists here.
func inventoryPlatformUsable(_ string) bool { return false }

func probeScriptMechanism(_, name string) (bool, error) {
	return false, diagnostic(CodeWorkerProtocolInvalid, "the enforced script policy is specified for Linux, macOS, and Windows only")
}
