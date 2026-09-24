//go:build windows

package scriptworker

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// scriptDomain is the manager-owned worker domain on Windows: a private Job
// Object carrying exactly this invocation's installable limits, which every
// interpreter descendant joins.
type scriptDomain struct {
	job windows.Handle
}

// prepareScriptDomain creates the private job with exactly the installable
// job-backed limits of this invocation, so closing the last handle
// terminates every process still inside it.
func prepareScriptDomain(probes []ScriptControlProbe) (*scriptDomain, error) {
	var flags uint32
	for _, probe := range probes {
		if !scriptControlInstallable(probe) {
			continue
		}
		switch probe.Name {
		case ScriptControlDescendantDomainTermination, ScriptControlActiveProcessCountLimit, ScriptControlAggregateMemoryLimit:
			limit, err := scriptJobLimitFlags(probe.Name)
			if err != nil {
				return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot create inventory control %q", probe.Name)
			}
			flags |= limit
		case ScriptControlInheritedHandleRestriction:
			// Installed by process creation: os/exec passes exactly the
			// standard handles through
			// PROC_THREAD_ATTRIBUTE_HANDLE_LIST and neither the manager
			// nor the worker adds an inherited handle.
		default:
			return nil, diagnostic(CodeWorkerProtocolInvalid, "no Windows mechanism for inventory control %q", probe.Name)
		}
	}
	job, err := newScriptControlJob(flags)
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot create the script worker domain")
	}
	return &scriptDomain{job: job}, nil
}

// launch starts the worker and assigns it to the private job before any
// session byte flows, so the interpreter it starts is born inside the
// domain. No suspended start is needed: the worker's first act is a
// blocking read of the request, which the manager sends only after this
// assignment returns, so no interpreter can exist outside the job.
func (domain *scriptDomain) launch(command *exec.Cmd) error {
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	// The explicit inherited-handle list contains exactly the three
	// standard handles: os/exec builds it, and no additional handle is
	// inherited.
	command.SysProcAttr.AdditionalInheritedHandles = nil
	if err := command.Start(); err != nil {
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot start the script worker")
	}
	handle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(command.Process.Pid))
	if err != nil {
		terminateScriptDomain(command, domain)
		_ = command.Wait()
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot open the script worker for domain assignment")
	}
	defer windows.CloseHandle(handle)
	if err := windows.AssignProcessToJobObject(domain.job, handle); err != nil {
		terminateScriptDomain(command, domain)
		_ = command.Wait()
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot assign the script worker to its domain")
	}
	return nil
}

// close releases the private job. Kill-on-close terminates every process
// still inside it, so no interpreter descendant can outlive the invocation.
func (domain *scriptDomain) close() {
	if domain == nil || domain.job == 0 {
		return
	}
	_ = windows.CloseHandle(domain.job)
	domain.job = 0
}

// cgroupParams reports no invocation cgroup: Windows carries aggregate
// limits on the private job object, which the worker confirms by query.
func (domain *scriptDomain) cgroupParams() (path, root string) { return "", "" }

// fileSizeParam reports no per-file bound: the control is not installable
// on Windows.
func (domain *scriptDomain) fileSizeParam() uint64 { return 0 }

func scriptWorkerSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{AdditionalInheritedHandles: nil}
}

func interpreterSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{AdditionalInheritedHandles: nil}
}

// terminateScriptDomain terminates the worker; closing the private job
// handle then reaps every descendant that outlived it.
func terminateScriptDomain(command *exec.Cmd, _ *scriptDomain) {
	if command == nil || command.Process == nil {
		return
	}
	_ = command.Process.Kill()
}

func indispensableScriptEnvironment() map[string]string {
	result := make(map[string]string, 2)
	for _, key := range []string{"SYSTEMROOT", "WINDIR"} {
		if value, present := os.LookupEnv(key); present && value != "" {
			result[key] = value
		}
	}
	return result
}
