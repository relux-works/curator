//go:build !linux

package scriptworker

import "syscall"

// Network namespaces exist only on Linux. Off Linux the probe always
// reports absent, so the worker never isolates the interpreter there.

// ScriptNetNSProbeMode names the hidden probe mode on every platform so
// the production dispatch reads identically; off Linux it is never
// re-executed.
const ScriptNetNSProbeMode = "__curator-script-netns-probe"

// RunNetNSProbe is the hidden probe entry.
func RunNetNSProbe() int { return 0 }

// probeScriptNetNS reports network isolation absent on non-Linux hosts.
func probeScriptNetNS() (bool, error) { return false, nil }

// scriptNetNSAttr refuses: a non-Linux worker must never be asked to
// isolate the interpreter in a namespace.
func scriptNetNSAttr() *syscall.SysProcAttr { return nil }

// checkScriptNetNS refuses: a non-Linux worker must never be asked to
// isolate the interpreter in a namespace.
func checkScriptNetNS(_ *workerRequest) error {
	return diagnostic(CodeCapabilityEvidenceInvalid, "network isolation is a Linux-only control")
}

// confirmScriptNetNS refuses: there is no namespace to confirm off Linux.
func confirmScriptNetNS(_ int) error {
	return diagnostic(CodeCapabilityEvidenceInvalid, "network isolation is a Linux-only control")
}
