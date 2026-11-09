//go:build !unix

package main

import "os/exec"

func runUmbrellaProvider(command *exec.Cmd) error { return command.Run() }
