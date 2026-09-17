package buildrepo

import (
	"os"
	"os/exec"
)

// setChildProcessGroup has no Windows process-group equivalent in this
// lane: the child keeps the caller's group and cancellation kills the Git
// process itself, while WaitDelay still bounds pipe draining past the
// deadline when an orphaned helper holds the pipes open.
func setChildProcessGroup(_ *exec.Cmd) {}

// killProcessTree terminates the Git child. An orphaned helper that
// survives still cannot hold the caller past WaitDelay.
func killProcessTree(p *os.Process) error {
	if p == nil {
		return nil
	}
	return p.Kill()
}
