package main

import (
	"fmt"
	"os"
	"syscall"
)

type providerExitObserver struct {
	fd    int
	pid   uint64
	ready bool
}

func newProviderExitObserver() (*providerExitObserver, error) {
	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}
	syscall.CloseOnExec(fd)
	return &providerExitObserver{fd: fd}, nil
}

func (o *providerExitObserver) Watch(process *os.Process) error {
	// This PID belongs to the original, unreaped child. Registration observes
	// that child; it neither adopts a discovered process nor consumes its wait.
	o.pid = uint64(process.Pid)
	change := syscall.Kevent_t{}
	syscall.SetKevent(&change, process.Pid, syscall.EVFILT_PROC, syscall.EV_ADD|syscall.EV_ONESHOT)
	change.Fflags = syscall.NOTE_EXIT
	var err error
	for {
		_, err = syscall.Kevent(o.fd, []syscall.Kevent_t{change}, nil, &syscall.Timespec{})
		if err != syscall.EINTR {
			break
		}
	}
	if err == syscall.ESRCH {
		// The original child exited before registration. No reaper has run;
		// still perform the real Cmd.Wait and preserve any error it returns.
		o.ready = true
		return nil
	}
	return err
}

func (o *providerExitObserver) Ready() (bool, error) {
	if o.ready {
		return true, nil
	}
	var events [1]syscall.Kevent_t
	n, err := syscall.Kevent(o.fd, nil, events[:], &syscall.Timespec{})
	if err == syscall.EINTR {
		return false, nil
	}
	if err != nil || n == 0 {
		return false, err
	}
	event := events[0]
	if event.Flags&syscall.EV_ERROR != 0 {
		return false, fmt.Errorf("provider exit event: %w", syscall.Errno(event.Data))
	}
	if event.Ident != o.pid || event.Filter != syscall.EVFILT_PROC || event.Fflags&syscall.NOTE_EXIT == 0 {
		return false, fmt.Errorf("unexpected provider exit event")
	}
	o.ready = true
	return true, nil
}

func (o *providerExitObserver) Close() error {
	if o.fd < 0 {
		return nil
	}
	fd := o.fd
	o.fd = -1
	return syscall.Close(fd)
}
