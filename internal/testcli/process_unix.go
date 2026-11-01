//go:build !windows

package testcli

import (
	"os"
	"os/exec"
	"syscall"
)

// Every fixture leads a separate group. Cancellation and return cleanup can
// stop only this invocation and descendants that remain in that group.
func configureFixtureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return killFixtureGroup(cmd.Process) }
}

func killFixtureGroup(process *os.Process) error {
	if process == nil {
		return nil
	}
	err := syscall.Kill(-process.Pid, syscall.SIGKILL)
	if err == syscall.ESRCH {
		return os.ErrProcessDone
	}
	return err
}

func cleanupFixtureProcess(cmd *exec.Cmd) { _ = killFixtureGroup(cmd.Process) }
