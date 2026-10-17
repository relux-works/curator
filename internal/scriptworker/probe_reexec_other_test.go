//go:build !linux && !darwin

package scriptworker

import "os/exec"

func protectProbeTestProcess(_ *exec.Cmd) {}
