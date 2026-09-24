//go:build !windows

package scriptworker

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// scriptDomain is the manager-owned worker domain on Unix: the worker's own
// process group and session, which every interpreter descendant inherits,
// plus the per-file bound installed across the fork and the invocation
// cgroup on Linux hosts that delegate one.
type scriptDomain struct {
	fileBound bool
	fileBytes uint64
	cgroup    *scriptCgroupDomain
}

// prepareScriptDomain creates the domain handle from this invocation's
// probes. The process group itself comes into existence with the worker
// fork below; the per-file bound and the invocation cgroup are prepared
// here, before the worker exists, so a failure refuses before any worker
// starts.
func prepareScriptDomain(probes []ScriptControlProbe) (*scriptDomain, error) {
	domain := &scriptDomain{}
	for _, probe := range probes {
		if !scriptControlInstallable(probe) {
			continue
		}
		switch probe.Name {
		case ScriptControlPerFileSizeLimit:
			bound, err := wantedScriptFileLimit()
			if err != nil {
				return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot derive the per-file byte bound")
			}
			domain.fileBound, domain.fileBytes = true, bound
		case ScriptControlDescendantDomainTermination, ScriptControlInheritedHandleRestriction:
			// Installed at fork and exec: the private session by the
			// kernel between fork and exec, and the descriptor hygiene
			// by the runtime plus the worker's explicit release.
		case ScriptControlActiveProcessCountLimit, ScriptControlAggregateMemoryLimit,
			ScriptControlDescendantExecDenial, ScriptControlFilesystemWriteConfinement,
			ScriptControlNetworkIsolationDomain:
			// Installed around the launch: the cgroup by the parent
			// below, Landlock and the network namespace by the worker
			// itself before the interpreter starts.
		default:
			return nil, diagnostic(CodeWorkerProtocolInvalid, "no unix mechanism for inventory control %q", probe.Name)
		}
	}
	cgroup, err := prepareScriptCgroup(probes)
	if err != nil {
		return nil, err
	}
	domain.cgroup = cgroup
	return domain, nil
}

// launch starts the worker in its own session, so its process group is
// private to this invocation and teardown can address the whole domain. The
// per-file bound is lowered across the fork so the worker and every
// descendant inherit it, and the worker is assigned to the invocation
// cgroup while it is still blocked reading the request.
func (domain *scriptDomain) launch(command *exec.Cmd) error {
	if !domain.fileBound {
		if err := command.Start(); err != nil {
			return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot start the script worker")
		}
		return domain.assign(command)
	}

	scriptFileLimitMutex.Lock()
	var previous unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &previous); err != nil {
		scriptFileLimitMutex.Unlock()
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot read RLIMIT_FSIZE")
	}
	if err := unix.Setrlimit(unix.RLIMIT_FSIZE, &unix.Rlimit{Cur: domain.fileBytes, Max: previous.Max}); err != nil {
		scriptFileLimitMutex.Unlock()
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot apply RLIMIT_FSIZE to the worker domain")
	}
	startErr := command.Start()
	restoreErr := unix.Setrlimit(unix.RLIMIT_FSIZE, &previous)
	scriptFileLimitMutex.Unlock()

	if startErr != nil {
		return diagnosticErr(CodeWorkerProtocolInvalid, startErr, "cannot start the script worker")
	}
	if restoreErr != nil {
		terminateScriptDomain(command, domain)
		_ = command.Wait()
		return diagnosticErr(CodeWorkerProtocolInvalid, restoreErr, "cannot restore the manager RLIMIT_FSIZE after the worker fork")
	}
	return domain.assign(command)
}

// assign moves the started worker into the invocation cgroup, if any,
// before any session byte flows. A failure destroys the worker while it is
// still blocked reading the request, so it never runs a program.
func (domain *scriptDomain) assign(command *exec.Cmd) error {
	if domain.cgroup == nil {
		return nil
	}
	if command.Process == nil {
		return diagnostic(CodeWorkerProtocolInvalid, "the started worker has no process identity")
	}
	if err := domain.cgroup.assign(command.Process.Pid); err != nil {
		terminateScriptDomain(command, domain)
		_ = command.Wait()
		return err
	}
	return nil
}

func (domain *scriptDomain) close() {
	if domain == nil {
		return
	}
	if domain.cgroup != nil {
		domain.cgroup.close()
		domain.cgroup = nil
	}
}

// cgroupParams reports the prepared invocation cgroup the worker confirms:
// the hierarchy-relative path plus the effective hierarchy root, so a
// fixture hierarchy propagates to the worker through the request.
func (domain *scriptDomain) cgroupParams() (path, root string) {
	if domain == nil || domain.cgroup == nil {
		return "", ""
	}
	return domain.cgroup.relative, domain.cgroup.filesystem.Root
}

// fileSizeParam reports the installed per-file soft bound, or 0 when the
// control is not installable.
func (domain *scriptDomain) fileSizeParam() uint64 {
	if domain == nil || !domain.fileBound {
		return 0
	}
	return domain.fileBytes
}

// scriptWorkerSysProcAttr places the worker in a new session.
func scriptWorkerSysProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{Setsid: true} }

// interpreterSysProcAttr keeps the interpreter inside the worker's process
// group, so the domain teardown below reaches every descendant.
func interpreterSysProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{} }

// terminateScriptDomain kills the worker's process group first so no
// interpreter descendant outlives the invocation, then the worker itself.
func terminateScriptDomain(command *exec.Cmd, _ *scriptDomain) {
	if command == nil || command.Process == nil {
		return
	}
	pid := command.Process.Pid
	if pid > 0 {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
	_ = command.Process.Kill()
}

func indispensableScriptEnvironment() map[string]string { return nil }
