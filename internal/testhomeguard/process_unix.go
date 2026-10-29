//go:build unix

package testhomeguard

import (
	"os/exec"
	"syscall"
)

// Only the observed child leads this group. Timeout teardown therefore cannot
// signal the package observer, its caller, or another lane's process group.
func launchObserved(cmd *exec.Cmd) (func() error, func(), error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	return func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }, func() {}, nil
}
