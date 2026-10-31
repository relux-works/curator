package main

import (
	"os"

	"golang.org/x/sys/unix"
)

type providerExitObserver struct{ pid int }

func newProviderExitObserver() (*providerExitObserver, error) {
	return &providerExitObserver{}, nil
}

func (o *providerExitObserver) Watch(process *os.Process) error {
	o.pid = process.Pid
	return nil
}

func (o *providerExitObserver) Ready() (bool, error) {
	var info unix.Siginfo
	err := unix.Waitid(unix.P_PID, o.pid, &info, unix.WEXITED|unix.WNOHANG|unix.WNOWAIT, nil)
	if err == unix.EINTR {
		return false, nil
	}
	// WNOWAIT leaves the original child unreaped until the sole Cmd.Wait.
	return info.Signo != 0, err
}

func (o *providerExitObserver) Close() error { return nil }
