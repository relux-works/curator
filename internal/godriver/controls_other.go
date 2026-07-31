//go:build !darwin && !windows

package godriver

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// rc5-native-control-inventory-v1 defines records for exactly macOS and
// Windows. On any other host probeNativeControls rejects before the worker
// starts, so these entry points exist only to keep the package buildable and
// must never be reached.

func probeNativeControl(name string, _ ResourceLimits) (bool, error) {
	return false, fmt.Errorf("rc5-native-control-inventory-v1 has no record for control %q on this platform", name)
}

// controlDomain is the manager-owned native control domain. It cannot exist on
// a platform the exhaustive inventory does not cover.
type controlDomain struct{}

func prepareControlDomain(_ ResourceLimits, _ []ControlProbe) (*controlDomain, error) {
	return nil, diagnostic(CodeControlUnavailable, "the portable execution policy is specified for macOS and Windows only")
}

func (domain *controlDomain) launch(_ *exec.Cmd) error {
	return diagnostic(CodeControlUnavailable, "the portable execution policy is specified for macOS and Windows only")
}

func (domain *controlDomain) installedControls() []string { return nil }

// No destroy here: it is only ever called after a launch that succeeded, and
// launch on this platform cannot succeed.

func (domain *controlDomain) close() {}

func observeNativeControls(_ ResourceLimits, _ []ControlProbe, _ []*os.File) ([]string, error) {
	return nil, diagnostic(CodeCapabilityEvidenceInvalid, "the portable execution policy is specified for macOS and Windows only")
}

func compilerSysProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{} }

func workerSysProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{} }

func terminateWorkerDomain(command *exec.Cmd) {
	if command == nil || command.Process == nil {
		return
	}
	_ = command.Process.Kill()
}
