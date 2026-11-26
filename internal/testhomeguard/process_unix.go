//go:build unix

package testhomeguard

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// Only the observed child leads this group. Timeout teardown therefore cannot
// signal the package observer, its caller, or another lane's process group.
//
// A supervisor that is itself a session leader owns a controlling terminal
// some fixtures drive directly: moving its observed child into a background
// group would break the foreground hand-off those fixtures exercise. That
// child stays in the leader's group and teardown then targets its pid only.
func launchObserved(cmd *exec.Cmd) (func() error, func(), error) {
	if sid, err := unix.Getsid(0); err == nil && sid == os.Getpid() {
		if err := cmd.Start(); err != nil {
			return nil, nil, err
		}
		return cmd.Process.Kill, func() {}, nil
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	return func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }, func() {}, nil
}
