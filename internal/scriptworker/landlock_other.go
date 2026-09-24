//go:build !linux

package scriptworker

import "errors"

// Landlock exists only on Linux. Off Linux the probe always reports
// absent, so the worker never confirms or enforces these controls there.

// ScriptLandlockProbeMode names the hidden probe mode on every platform
// so the production dispatch reads identically; off Linux it is never
// re-executed.
const ScriptLandlockProbeMode = "__curator-script-landlock-probe"

// probeScriptLandlock reports the Landlock controls absent on non-Linux
// hosts.
func probeScriptLandlock(_ string) (bool, error) { return false, nil }

// landlockABI reports no Landlock ABI off Linux. The probe never
// consults it there; tests call it only behind a Linux gate, and the
// stub keeps the untagged test files portable.
func landlockABI() (int, error) { return 0, errors.New("landlock is a Linux-only control") }

// confirmScriptLandlock confirms nothing off Linux: a non-Linux worker
// must never be asked to apply a Linux-only control, so an installable
// claim for one fails that control and the worker refuses fail-closed
// instead of contradicting the probe.
func confirmScriptLandlock(execDenial, writeConfinement bool, _ landlockGrants) (execErr, writeErr error) {
	if execDenial {
		execErr = errors.New("landlock is a Linux-only control")
	}
	if writeConfinement {
		writeErr = errors.New("landlock is a Linux-only control")
	}
	return execErr, writeErr
}

// enforceScriptLandlock refuses: a non-Linux worker must never be asked
// to enforce a Linux-only control.
func enforceScriptLandlock(_, _ bool, _ landlockGrants) error {
	return errors.New("landlock is a Linux-only control")
}

// RunLandlockProbe is the hidden probe entry. It is never re-executed
// off Linux; the nonzero exit answers the probe truthfully.
func RunLandlockProbe(_ string) int { return 1 }
