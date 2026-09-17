//go:build !windows

package buildrepo

import (
	"os"
	"os/exec"
	"syscall"
)

// setChildProcessGroup makes one Git child the leader of its own process
// group, so cancelling the fetch can reach the whole process graph: Git's
// SSH/helper grandchildren otherwise keep the stdio pipes open and the
// caller waits past the total deadline.
func setChildProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// killProcessTree terminates the whole group led by p. A group kill that
// reports the group already gone still succeeds: the tree is down either
// way. Anything else falls back to killing the leader alone.
func killProcessTree(p *os.Process) error {
	if p == nil {
		return nil
	}
	if err := syscall.Kill(-p.Pid, syscall.SIGKILL); err == nil || err == syscall.ESRCH {
		return nil
	}
	return p.Kill()
}
