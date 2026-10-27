package testhomeguard

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const supervisorPID = "CURATOR_TEST_HOME_GUARD_SUPERVISOR"

// Supervise reexecutes the test binary under a surviving postimage observer.
// Call before any TestMain dispatch. The child returns (0,false) and runs the
// tests; the parent returns the checked exit code and must exit immediately.
// A test panic, testing timeout or direct os.Exit cannot bypass the parent's
// check. Killing/crashing the supervisor itself remains outside this boundary.
// fixtureChild excludes intentionally supplied fixture homes, never account
// homes. Ambient Git/HOME isolation belongs exclusively to testgitenv.
func Supervise(fixtureChild bool) (int, bool) {
	if os.Getenv(supervisorPID) == strconv.Itoa(os.Getppid()) {
		if err := verifyObserver(); err != nil {
			fmt.Fprintln(os.Stderr, "testhomeguard: live observer handoff required")
			return 2, true
		}
		return 0, false
	}
	g, err := start(!fixtureChild)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2, true
	}
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	observer, err := newObserver()
	if err != nil {
		fmt.Fprintln(os.Stderr, "testhomeguard: cannot establish observer handoff")
		return 2, true
	}
	defer observer.Close()
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != supervisorPID && key != observerAddress && key != observerKey {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, supervisorPID+"="+strconv.Itoa(os.Getpid()))
	cmd.Env = append(cmd.Env, observerAddress+"="+observer.Addr().String(), observerKey+"="+observer.key)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	code := 0
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "testhomeguard: cannot start observed child")
		return g.Finish(2), true
	}
	authenticated := make(chan error, 1)
	go func() { authenticated <- observer.authenticate(cmd.Process.Pid) }()
	stopSignals := forwardDiagnosticSignals(cmd.Process)
	defer stopSignals()
	if err := cmd.Wait(); err != nil {
		code = 2
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() >= 0 {
			code = exit.ExitCode()
		}
	}
	_ = observer.Close()
	if err := <-authenticated; err != nil {
		fmt.Fprintln(os.Stderr, "testhomeguard: child observer handoff not verified")
		code = 2
	}
	return g.Finish(code), true
}

// ControlProcess distinguishes an ambient-isolation control from a fixture
// helper; the CLI control still borrows its parent's prepared Go toolchain.
func ControlProcess() bool { return os.Getenv(controlMode) != "" }
