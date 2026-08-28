//go:build !unix && !windows

package closureexec

import "os/exec"

func configurePortableProcess(_ *exec.Cmd) {}

func terminatePortableProcess(command *exec.Cmd) {
	if command != nil && command.Process != nil {
		_ = command.Process.Kill()
	}
}
