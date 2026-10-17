package scriptworker

import "os"

// Dispatch before any importing package's initialization, TestMain, or main.
// Re-execution is safe even in binaries that do not know about worker modes:
// importing the probe also installs its entry point. Malformed hidden requests
// fail silently rather than falling through to unrelated application work.
func init() {
	if !isLandlockProbeChild() {
		return
	}
	if len(os.Args) != 3 {
		os.Exit(1)
	}
	os.Exit(RunLandlockProbe(os.Args[2]))
}

func isLandlockProbeChild() bool {
	return len(os.Args) > 1 && os.Args[1] == ScriptLandlockProbeMode
}
