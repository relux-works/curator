//go:build linux

package scriptworker

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// Network-namespace isolation for the Linux host-conditional control
// `network-isolation-domain`.
//
// The interpreter child starts in a fresh user and network namespace, so it
// sees no host interface — only a loopback that is down. The worker itself
// stays in the host namespaces and confirms the isolation by comparing the
// child's network-namespace identity with its own.

// ScriptNetNSProbeMode is the fixed hidden probe the network-isolation
// availability check re-executes. Like the worker mode it is an
// implementation boundary, not a user-visible command: it proves exactly
// the namespace creation the interpreter child will use, then exits.
const ScriptNetNSProbeMode = "__curator-script-netns-probe"

// scriptNetNSProbeTimeout bounds one network-isolation probe.
const scriptNetNSProbeTimeout = 30 * time.Second

// RunNetNSProbe is the hidden probe entry: it does nothing but exit 0, so
// reaching it proves the namespace creation the parent requested.
func RunNetNSProbe() int { return 0 }

// probeScriptNetNS determines whether this host lets the invocation place
// the interpreter in a fresh user and network namespace: it re-executes
// this manager in the hidden probe mode with exactly the unshare flags and
// identity mapping the interpreter child will use. Every
// expected-absence condition — no user namespaces, no network namespaces,
// a denied mapping — reports absent without an error.
func probeScriptNetNS() (bool, error) {
	manager, err := os.Executable()
	if err != nil {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), scriptNetNSProbeTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, manager, ScriptNetNSProbeMode) // #nosec G204 -- self-re-execution in the fixed hidden probe mode
	command.Env = scriptWorkerEnvironment()
	command.SysProcAttr = &syscall.SysProcAttr{
		Unshareflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}
	if err := command.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

// scriptNetNSAttr returns the process attributes that place the interpreter
// child in a fresh user and network namespace with this process's identity
// mapped to the container root.
func scriptNetNSAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Unshareflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}
}

// checkScriptNetNS accepts the network-isolation installable claim: Linux
// provides the mechanism, and the interpreter spawn confirms it.
func checkScriptNetNS(_ *workerRequest) error { return nil }

// confirmScriptNetNS proves the started interpreter child lives in a
// network namespace of its own: its namespace identity must differ from
// the worker's. A fresh namespace carries no host interface, which is
// exactly the isolated domain the control names.
func confirmScriptNetNS(child int) error {
	own, err := os.Readlink("/proc/self/ns/net")
	if err != nil {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err, "cannot read the worker network namespace")
	}
	other, err := os.Readlink("/proc/" + strconv.Itoa(child) + "/ns/net")
	if err != nil {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err, "cannot read the interpreter network namespace")
	}
	if other == own {
		return diagnostic(CodeCapabilityEvidenceInvalid, "the interpreter shares the worker network namespace")
	}
	return nil
}
