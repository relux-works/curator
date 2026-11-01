package testcli

import "os/exec"

// POSIX transport fixtures skip Windows. Other Windows fixture invocations
// retain direct-child cancellation and bounded pipe wait; this seam does not
// claim descendant-tree termination on Windows.
func configureFixtureProcess(_ *exec.Cmd) {}
func cleanupFixtureProcess(_ *exec.Cmd)   {}
