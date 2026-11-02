//go:build !unix && !windows

package testhomeguard

import "os/exec"

func launchObserved(cmd *exec.Cmd) (func() error, func(), error) {
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	return cmd.Process.Kill, func() {}, nil
}
