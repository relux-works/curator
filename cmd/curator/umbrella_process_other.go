//go:build !darwin && !linux

package main

import "os/exec"

// Other platforms retain ordinary dispatch without a supervised-stop guarantee.
func runUmbrellaProvider(command *exec.Cmd) error {
	return command.Run()
}
