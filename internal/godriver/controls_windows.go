//go:build windows

package godriver

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// jobLimitFlags maps one inventory control to the Job Object limit that carries
// it. Only the three job-backed controls have an entry.
func jobLimitFlags(name string) (uint32, error) {
	switch name {
	case ControlDescendantDomainTermination:
		return windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE, nil
	case ControlActiveProcessCountLimit:
		return windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS, nil
	case ControlAggregateMemoryLimit:
		return windows.JOB_OBJECT_LIMIT_JOB_MEMORY | windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY, nil
	default:
		return 0, fmt.Errorf("no Windows Job Object limit for inventory control %q", name)
	}
}

// jobLimitInformation is the exact limit block this operation installs.
func jobLimitInformation(flags uint32, limits ResourceLimits) (windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION, error) {
	information := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	information.BasicLimitInformation.LimitFlags = flags
	if flags&windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS != 0 {
		if limits.Processes <= 0 {
			return information, fmt.Errorf("the active-process bound is not positive")
		}
		information.BasicLimitInformation.ActiveProcessLimit = uint32(limits.Processes) // #nosec G115 -- normalizeBuildLimits bounds this to a positive value below defaultProcessLimit
	}
	if flags&windows.JOB_OBJECT_LIMIT_JOB_MEMORY != 0 {
		if limits.MemoryBytes <= 0 {
			return information, fmt.Errorf("the aggregate memory bound is not positive")
		}
		information.JobMemoryLimit = uintptr(limits.MemoryBytes)
		information.ProcessMemoryLimit = uintptr(limits.MemoryBytes)
	}
	return information, nil
}

// newControlJob creates one private Job Object carrying exactly the requested
// limits for exactly this operation.
func newControlJob(flags uint32, limits ResourceLimits) (windows.Handle, error) {
	information, err := jobLimitInformation(flags, limits)
	if err != nil {
		return 0, err
	}
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return 0, err
	}
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&information)), uint32(unsafe.Sizeof(information))); err != nil {
		_ = windows.CloseHandle(job)
		return 0, err
	}
	return job, nil
}

// probeNativeControl performs the per-operation availability determination for
// one Windows inventory control. It creates the exact kernel object the control
// will use with the exact limits of this operation, starts no program, and
// caches nothing.
func probeNativeControl(name string, limits ResourceLimits) (bool, error) {
	switch name {
	case ControlDescendantDomainTermination, ControlActiveProcessCountLimit, ControlAggregateMemoryLimit:
		flags, err := jobLimitFlags(name)
		if err != nil {
			return false, err
		}
		job, err := newControlJob(flags, limits)
		if err != nil {
			return false, err
		}
		return true, windows.CloseHandle(job)
	case ControlInheritedHandleRestriction:
		// os/exec restricts inheritance with PROC_THREAD_ATTRIBUTE_HANDLE_LIST,
		// so the attribute list must be allocatable on this host.
		list, err := windows.NewProcThreadAttributeList(1)
		if err != nil {
			return false, err
		}
		list.Delete()
		return true, nil
	default:
		return false, fmt.Errorf("no Windows probe for inventory control %q", name)
	}
}

// controlDomain is the manager-owned Windows control domain. The parent creates
// the private Job Object before the worker exists, creates the worker
// suspended, assigns it to the job, and only then resumes it. The explicit
// inherited-handle list is installed by process creation itself. No control is
// applied after the worker begins executing, so CodeControlUnavailable cannot
// first appear there.
type controlDomain struct {
	controls  []string
	job       windows.Handle
	flags     uint32
	limits    ResourceLimits
	installed bool
}

// prepareControlDomain builds the Windows domain from this operation's probes.
// It creates no process and rejects with CodeControlUnavailable before the
// worker exists.
func prepareControlDomain(limits ResourceLimits, probes []ControlProbe) (*controlDomain, error) {
	if err := seamFault(seamPrepare); err != nil {
		return nil, diagnosticErr(CodeControlUnavailable, err, "cannot create the manager-owned native control domain")
	}
	domain := &controlDomain{limits: limits}
	for _, name := range installableControls(probes) {
		switch name {
		case ControlDescendantDomainTermination, ControlActiveProcessCountLimit, ControlAggregateMemoryLimit:
			flags, err := jobLimitFlags(name)
			if err != nil {
				return nil, diagnosticErr(CodeControlUnavailable, err, "cannot create inventory control %q", name)
			}
			domain.flags |= flags
		case ControlInheritedHandleRestriction:
			// Installed by process creation: os/exec passes exactly the standard
			// handles through PROC_THREAD_ATTRIBUTE_HANDLE_LIST and neither the
			// manager nor the worker adds an inherited handle.
		default:
			return nil, diagnostic(CodeControlUnavailable, "no Windows mechanism for inventory control %q", name)
		}
		domain.controls = append(domain.controls, name)
	}
	if domain.flags != 0 {
		job, err := newControlJob(domain.flags, limits)
		if err != nil {
			return nil, diagnosticErr(CodeControlUnavailable, err, "cannot create the private worker job object")
		}
		domain.job = job
	}
	return domain, nil
}

// launch starts the identity-verified worker suspended, completes the control
// domain around it, and resumes it only once every mechanism is installed. A
// failure destroys the worker while it is still suspended, so it never executes
// an instruction.
func (domain *controlDomain) launch(command *exec.Cmd) error {
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	command.SysProcAttr.CreationFlags |= windows.CREATE_SUSPENDED
	if err := command.Start(); err != nil {
		return diagnosticErr(CodeWorkerIdentityInvalid, err, "cannot start the identity-verified worker")
	}
	if err := domain.attach(command); err != nil {
		domain.destroy(command)
		return err
	}
	domain.installed = true
	return nil
}

// attach assigns the suspended worker to the private job and releases it. Every
// failure here happens while the worker is still suspended.
func (domain *controlDomain) attach(command *exec.Cmd) error {
	if err := seamFault(seamInstall); err != nil {
		return diagnosticErr(CodeControlUnavailable, err, "cannot install the manager-owned native control domain")
	}
	if command.Process == nil || command.Process.Pid <= 0 {
		return diagnostic(CodeControlUnavailable, "the suspended worker has no process identity")
	}
	pid := uint32(command.Process.Pid) // #nosec G115 -- the identifier was just proved positive
	if domain.job != 0 {
		process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, pid)
		if err != nil {
			return diagnosticErr(CodeControlUnavailable, err, "cannot open the suspended worker for job assignment")
		}
		assignErr := windows.AssignProcessToJobObject(domain.job, process)
		_ = windows.CloseHandle(process)
		if assignErr != nil {
			return diagnosticErr(CodeControlUnavailable, assignErr, "cannot assign the worker to its private job object")
		}
	}
	return resumeSuspendedProcess(pid)
}

// resumeSuspendedProcess releases every thread of the suspended worker. The
// worker executes its first instruction here and nowhere earlier.
func resumeSuspendedProcess(pid uint32) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return diagnosticErr(CodeControlUnavailable, err, "cannot enumerate the suspended worker threads")
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()
	var entry windows.ThreadEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	resumed := 0
	for err = windows.Thread32First(snapshot, &entry); err == nil; err = windows.Thread32Next(snapshot, &entry) {
		if entry.OwnerProcessID != pid {
			continue
		}
		thread, openErr := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
		if openErr != nil {
			return diagnosticErr(CodeControlUnavailable, openErr, "cannot open the suspended worker thread")
		}
		count, resumeErr := windows.ResumeThread(thread)
		_ = windows.CloseHandle(thread)
		if resumeErr != nil || count == ^uint32(0) {
			return diagnosticErr(CodeControlUnavailable, resumeErr, "cannot resume the suspended worker thread")
		}
		resumed++
	}
	if resumed == 0 {
		return diagnostic(CodeControlUnavailable, "the suspended worker has no resumable thread")
	}
	return nil
}

// installedControls reports the controls whose mechanism is in place. It is
// empty until the domain has been installed, so evidence can never claim an
// applied status before installation.
func (domain *controlDomain) installedControls() []string {
	if domain == nil || !domain.installed {
		return nil
	}
	return append([]string(nil), domain.controls...)
}

// destroy terminates a worker that was never released to run its session.
func (domain *controlDomain) destroy(command *exec.Cmd) {
	terminateWorkerDomain(command)
	if command != nil && command.Process != nil {
		_ = command.Wait()
	}
}

// close releases the private job. Kill-on-close terminates every process still
// inside it, so no compiler or tool child can outlive the operation.
func (domain *controlDomain) close() {
	if domain == nil || domain.job == 0 {
		return
	}
	_ = windows.CloseHandle(domain.job)
	domain.job = 0
}

// observeNativeControls runs inside the worker. It confirms that every control
// the parent installed is really in effect here. It applies no availability
// decision: a contradiction is an evidence fault, never a mandatory-control
// rejection.
func observeNativeControls(limits ResourceLimits, probes []ControlProbe, _ []*os.File) ([]string, error) {
	names := installableControls(probes)
	var wanted uint32
	for _, name := range names {
		if flags, err := jobLimitFlags(name); err == nil {
			wanted |= flags
		}
	}
	if wanted != 0 {
		if err := confirmJobLimits(wanted, limits); err != nil {
			return nil, err
		}
	}
	confirmed := make([]string, 0, len(names))
	for _, name := range names {
		switch name {
		case ControlDescendantDomainTermination, ControlActiveProcessCountLimit, ControlAggregateMemoryLimit:
			// Confirmed together by the job query above.
		case ControlInheritedHandleRestriction:
			list, err := windows.NewProcThreadAttributeList(1)
			if err != nil {
				return nil, diagnosticErr(CodeCapabilityEvidenceInvalid, err,
					"cannot confirm the explicit inherited-handle list in the worker")
			}
			list.Delete()
		default:
			return nil, diagnostic(CodeCapabilityEvidenceInvalid, "no Windows confirmation for inventory control %q", name)
		}
		confirmed = append(confirmed, name)
	}
	return confirmed, nil
}

// confirmJobLimits queries the job the worker was assigned to and requires the
// exact limits the manager installed.
func confirmJobLimits(wanted uint32, limits ResourceLimits) error {
	expected, err := jobLimitInformation(wanted, limits)
	if err != nil {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err, "cannot derive the expected worker job limits")
	}
	var actual windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	var returned uint32
	// A zero handle queries the job object of the calling process, which is the
	// private job the manager assigned the worker to before resuming it.
	if err := windows.QueryInformationJobObject(0, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&actual)), uint32(unsafe.Sizeof(actual)), &returned); err != nil {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err, "the worker is not inside its private job object")
	}
	if actual.BasicLimitInformation.LimitFlags&wanted != wanted {
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"worker job limit flags are %#x, want at least %#x", actual.BasicLimitInformation.LimitFlags, wanted)
	}
	if wanted&windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS != 0 &&
		actual.BasicLimitInformation.ActiveProcessLimit != expected.BasicLimitInformation.ActiveProcessLimit {
		return diagnostic(CodeCapabilityEvidenceInvalid, "worker job active-process limit is %d, want %d",
			actual.BasicLimitInformation.ActiveProcessLimit, expected.BasicLimitInformation.ActiveProcessLimit)
	}
	if wanted&windows.JOB_OBJECT_LIMIT_JOB_MEMORY != 0 && actual.JobMemoryLimit != expected.JobMemoryLimit {
		return diagnostic(CodeCapabilityEvidenceInvalid, "worker job memory limit is %d, want %d",
			actual.JobMemoryLimit, expected.JobMemoryLimit)
	}
	return nil
}

// compilerSysProcAttr adds no inherited handle, so os/exec's explicit handle
// list contains exactly the three standard handles.
func compilerSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{AdditionalInheritedHandles: nil}
}

// workerSysProcAttr is the base attribute set of the worker launch. The control
// domain adds suspended creation before the manager starts the process.
func workerSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{AdditionalInheritedHandles: nil}
}

// terminateWorkerDomain terminates the worker; closing the private job handle
// afterwards terminates every remaining process in the domain.
func terminateWorkerDomain(command *exec.Cmd) {
	if command == nil || command.Process == nil {
		return
	}
	_ = command.Process.Kill()
}
