//go:build !windows

package runtimestore

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func TestUnixPostInstallShimPropagatesSignal(t *testing.T) {
	root := t.TempDir()
	artifact := filepath.Join(root, "artifact")
	if err := os.WriteFile(artifact, []byte("#!/bin/sh\nkill -TERM $$\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	shim := mustManagedShim(t, ProjectShim, filepath.Join(root, "live"), "tool", "unix")
	plan, err := StageShimTransition(filepath.Join(root, "stage"), []ShimSpec{{Destination: shim, Target: compiledTargetFixture(t, artifact, runtime.GOOS)}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	installFixtureTarget(t, plan.Desired[0])
	err = exec.Command(plan.Desired[0].LivePath).Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("signal exit = %v", err)
	}
	wait, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok || !wait.Signaled() || wait.Signal() != syscall.SIGTERM {
		t.Fatalf("wrapper did not propagate SIGTERM: %v", exitErr.ProcessState)
	}
}
