//go:build darwin

package godriver

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

// fileLimitMutex serializes the manager-side RLIMIT_FSIZE window. The soft
// bound is process-wide, so probing and the launch window must not overlap.
var fileLimitMutex sync.Mutex

// wantedFileLimit is the exact per-file soft bound this operation installs. It
// is derived from the manager-owned limit alone, clamped to the inherited hard
// bound, so the parent and the worker compute the same value.
func wantedFileLimit(limits ResourceLimits) (uint64, error) {
	if limits.FileBytes <= 0 {
		return 0, fmt.Errorf("the per-file byte bound is not positive")
	}
	var current unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &current); err != nil {
		return 0, err
	}
	wanted := uint64(limits.FileBytes) // #nosec G115 -- normalizeBuildLimits bounds this to a positive value below defaultFileLimit
	if current.Max != unix.RLIM_INFINITY && wanted > current.Max {
		wanted = current.Max
	}
	return wanted, nil
}

// probeNativeControl performs the per-operation availability determination for
// one macOS inventory control. It performs the exact operation the control will
// perform for this build, starts no program, and caches nothing.
func probeNativeControl(name string, limits ResourceLimits) (bool, error) {
	switch name {
	case ControlDescendantDomainTermination:
		// Session and process-group teardown requires a usable process group and
		// permission to signal it. Signal 0 performs the permission check only.
		// The private session itself is created by the kernel between fork and
		// exec, so it cannot fail once the worker is executing.
		group, err := unix.Getpgid(0)
		if err != nil {
			return false, err
		}
		if err := unix.Kill(-group, 0); err != nil {
			return false, err
		}
		return true, nil
	case ControlPerFileSizeLimit:
		// Prove the exact byte bound this operation will install is settable,
		// then restore the manager's own bound immediately.
		wanted, err := wantedFileLimit(limits)
		if err != nil {
			return false, err
		}
		fileLimitMutex.Lock()
		defer fileLimitMutex.Unlock()
		var previous unix.Rlimit
		if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &previous); err != nil {
			return false, err
		}
		if err := unix.Setrlimit(unix.RLIMIT_FSIZE, &unix.Rlimit{Cur: wanted, Max: previous.Max}); err != nil {
			return false, err
		}
		var applied unix.Rlimit
		readErr := unix.Getrlimit(unix.RLIMIT_FSIZE, &applied)
		restoreErr := unix.Setrlimit(unix.RLIMIT_FSIZE, &previous)
		if restoreErr != nil {
			return false, restoreErr
		}
		if readErr != nil {
			return false, readErr
		}
		return applied.Cur == wanted, nil
	case ControlInheritedHandleRestriction:
		// Close-on-exec must be readable and settable for the descriptors the
		// worker holds when the compiler starts.
		var pair [2]int
		if err := unix.Pipe(pair[:]); err != nil {
			return false, err
		}
		defer func() {
			_ = unix.Close(pair[0])
			_ = unix.Close(pair[1])
		}()
		flags, err := unix.FcntlInt(uintptr(pair[0]), unix.F_GETFD, 0)
		if err != nil {
			return false, err
		}
		if _, err := unix.FcntlInt(uintptr(pair[0]), unix.F_SETFD, flags|unix.FD_CLOEXEC); err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, fmt.Errorf("no macOS probe for inventory control %q", name)
	}
}

// controlDomain is the manager-owned macOS control domain. Every mechanism is
// installed by the parent: the private session by the kernel between fork and
// exec, and the per-file bound by lowering RLIMIT_FSIZE across the fork so the
// worker and every descendant inherit it. Nothing is applied after the worker
// begins executing, so CodeControlUnavailable cannot first appear there.
type controlDomain struct {
	controls  []string
	session   bool
	fileBound bool
	fileBytes uint64
	installed bool
}

// prepareControlDomain builds the macOS domain from this operation's probes. It
// creates no process and rejects with CodeControlUnavailable before the worker
// exists.
func prepareControlDomain(limits ResourceLimits, probes []ControlProbe) (*controlDomain, error) {
	if err := seamFault(seamPrepare); err != nil {
		return nil, diagnosticErr(CodeControlUnavailable, err, "cannot create the manager-owned native control domain")
	}
	domain := &controlDomain{}
	for _, name := range installableControls(probes) {
		switch name {
		case ControlDescendantDomainTermination:
			domain.session = true
		case ControlPerFileSizeLimit:
			bytes, err := wantedFileLimit(limits)
			if err != nil {
				return nil, diagnosticErr(CodeControlUnavailable, err, "cannot derive the per-file byte bound")
			}
			domain.fileBound, domain.fileBytes = true, bytes
		case ControlInheritedHandleRestriction:
			// Installed at fork and exec: the manager passes exactly the three
			// standard descriptors and no extra file, and the Go runtime opens
			// every other descriptor close-on-exec.
		default:
			return nil, diagnostic(CodeControlUnavailable, "no macOS mechanism for inventory control %q", name)
		}
		domain.controls = append(domain.controls, name)
	}
	return domain, nil
}

// launch starts the identity-verified worker with every probed-available macOS
// control already installed. The last installation step happens before fork, so
// a failure here leaves no worker process at all.
func (domain *controlDomain) launch(command *exec.Cmd) error {
	if err := seamFault(seamInstall); err != nil {
		return diagnosticErr(CodeControlUnavailable, err, "cannot install the manager-owned native control domain")
	}
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	// A private session makes the worker its own process-group leader, so one
	// killpg terminates the complete worker domain and no descendant can outlive
	// it by changing groups first.
	command.SysProcAttr.Setsid = domain.session
	if !domain.fileBound {
		if err := command.Start(); err != nil {
			return diagnosticErr(CodeWorkerIdentityInvalid, err, "cannot start the identity-verified worker")
		}
		domain.installed = true
		return nil
	}

	fileLimitMutex.Lock()
	var previous unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &previous); err != nil {
		fileLimitMutex.Unlock()
		return diagnosticErr(CodeControlUnavailable, err, "cannot read RLIMIT_FSIZE")
	}
	if err := unix.Setrlimit(unix.RLIMIT_FSIZE, &unix.Rlimit{Cur: domain.fileBytes, Max: previous.Max}); err != nil {
		fileLimitMutex.Unlock()
		return diagnosticErr(CodeControlUnavailable, err, "cannot apply RLIMIT_FSIZE to the worker domain")
	}
	startErr := command.Start()
	restoreErr := unix.Setrlimit(unix.RLIMIT_FSIZE, &previous)
	fileLimitMutex.Unlock()

	if startErr != nil {
		return diagnosticErr(CodeWorkerIdentityInvalid, startErr, "cannot start the identity-verified worker")
	}
	if restoreErr != nil {
		domain.destroy(command)
		return diagnosticErr(CodeControlUnavailable, restoreErr, "cannot restore the manager RLIMIT_FSIZE after the worker fork")
	}
	domain.installed = true
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

// close releases the domain. macOS holds no kernel object beyond the session,
// which is torn down with the process group.
func (domain *controlDomain) close() {}

// observeNativeControls runs inside the worker. It confirms that every control
// the parent installed is really in effect here and completes the worker's own
// descriptor duty. It applies no availability decision: a contradiction is an
// evidence fault, never a mandatory-control rejection.
func observeNativeControls(limits ResourceLimits, probes []ControlProbe, protocol []*os.File) ([]string, error) {
	confirmed := make([]string, 0, len(probes))
	for _, name := range installableControls(probes) {
		switch name {
		case ControlDescendantDomainTermination:
			group, err := unix.Getpgid(0)
			if err != nil || group != os.Getpid() {
				return nil, diagnosticErr(CodeCapabilityEvidenceInvalid, err,
					"the worker is not the leader of its private session and process group")
			}
		case ControlPerFileSizeLimit:
			wanted, err := wantedFileLimit(limits)
			if err != nil {
				return nil, diagnosticErr(CodeCapabilityEvidenceInvalid, err, "cannot derive the per-file byte bound")
			}
			var applied unix.Rlimit
			if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &applied); err != nil || applied.Cur != wanted {
				return nil, diagnosticErr(CodeCapabilityEvidenceInvalid, err,
					"RLIMIT_FSIZE is %d in the worker, want the installed bound %d", applied.Cur, wanted)
			}
		case ControlInheritedHandleRestriction:
			if err := releaseInheritedDescriptors(protocol); err != nil {
				return nil, diagnosticErr(CodeCapabilityEvidenceInvalid, err,
					"cannot confirm close-on-exec on the worker protocol descriptors")
			}
		default:
			return nil, diagnostic(CodeCapabilityEvidenceInvalid, "no macOS confirmation for inventory control %q", name)
		}
		confirmed = append(confirmed, name)
	}
	return confirmed, nil
}

// releaseInheritedDescriptors proves close-on-exec is set on the descriptors the
// worker still holds, so the compiler inherits none of them.
func releaseInheritedDescriptors(keep []*os.File) error {
	for _, file := range keep {
		if file == nil {
			continue
		}
		flags, err := unix.FcntlInt(file.Fd(), unix.F_GETFD, 0)
		if err != nil {
			return err
		}
		if flags&unix.FD_CLOEXEC == 0 {
			if _, err := unix.FcntlInt(file.Fd(), unix.F_SETFD, flags|unix.FD_CLOEXEC); err != nil {
				return err
			}
		}
		confirmed, err := unix.FcntlInt(file.Fd(), unix.F_GETFD, 0)
		if err != nil || confirmed&unix.FD_CLOEXEC == 0 {
			return fmt.Errorf("descriptor %d is not close-on-exec", file.Fd())
		}
	}
	return nil
}

// compilerSysProcAttr keeps the compiler inside the worker's process group and
// session so one killpg terminates the complete domain, and adds no inherited
// descriptor.
func compilerSysProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{} }

// workerSysProcAttr is the base attribute set of the worker launch. The control
// domain adds the private session before the manager forks.
func workerSysProcAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{} }

// terminateWorkerDomain kills the worker's process group first so no compiler
// or tool child outlives the operation, then the worker process itself.
func terminateWorkerDomain(command *exec.Cmd) {
	if command == nil || command.Process == nil {
		return
	}
	pid := command.Process.Pid
	if pid > 0 {
		_ = unix.Kill(-pid, unix.SIGKILL)
	}
	_ = command.Process.Kill()
}
