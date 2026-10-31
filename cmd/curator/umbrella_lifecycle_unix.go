//go:build darwin || linux

package main

import (
	"os"
	"os/exec"
	"sync"
)

type providerChildLifecycle struct {
	command *exec.Cmd
	mu      sync.Mutex
	reaping bool
}

func (p *providerChildLifecycle) signal(signal os.Signal, deliver func(*os.Process, os.Signal) error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.reaping {
		return os.ErrProcessDone
	}
	return deliver(p.command.Process, signal)
}

func (p *providerChildLifecycle) cancel(cancel func() error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.reaping {
		return os.ErrProcessDone
	}
	return cancel()
}

func (p *providerChildLifecycle) beginWait() {
	p.mu.Lock()
	p.reaping = true
	p.mu.Unlock()
}
