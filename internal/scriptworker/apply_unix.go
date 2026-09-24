//go:build unix

package scriptworker

import (
	"os"

	"golang.org/x/sys/unix"
)

// confirmScriptTermination proves the worker leads its private session and
// process group: the group the parent addresses at teardown.
func confirmScriptTermination(_ *workerRequest) error {
	group, err := unix.Getpgid(0)
	if err != nil || group != os.Getpid() {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err,
			"the worker is not the leader of its private session and process group")
	}
	return nil
}

// confirmScriptAggregateLimits proves the invocation cgroup membership on
// Linux request platforms. Elsewhere these controls are never installable,
// and the request validation already refused an installable claim for
// them — reaching here with one set is a request contradiction.
func confirmScriptAggregateLimits(request *workerRequest, installable map[string]bool) error {
	pids := installable[ScriptControlActiveProcessCountLimit]
	memory := installable[ScriptControlAggregateMemoryLimit]
	if !pids && !memory {
		return nil
	}
	if request.InventoryPlatform != ScriptPlatformLinux {
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"aggregate cgroup limits are installable only on the Linux inventory")
	}
	if request.CgroupPath == "" {
		return diagnostic(CodeCapabilityEvidenceInvalid, "the request names no invocation cgroup")
	}
	root := request.CgroupRoot
	if root == "" {
		root = defaultScriptCgroupFS().Root
	}
	return confirmScriptCgroup(scriptCgroupFS{Root: root}, request.CgroupPath)
}

// confirmScriptFileSize proves the installed per-file soft bound is in
// effect here.
func confirmScriptFileSize(request *workerRequest) error {
	if request.FileSizeBound == 0 {
		return diagnostic(CodeCapabilityEvidenceInvalid, "the request names no per-file byte bound")
	}
	var applied unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &applied); err != nil || applied.Cur != request.FileSizeBound {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err,
			"RLIMIT_FSIZE is %d in the worker, want the installed bound %d", applied.Cur, request.FileSizeBound)
	}
	return nil
}

// currentScriptFileSizeLimit reports this process's own per-file soft
// bound. Tests use it to parameterize worker fixtures with a bound the
// raw worker observes; production never calls it.
func currentScriptFileSizeLimit() (uint64, bool) {
	var current unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &current); err != nil {
		return 0, false
	}
	return current.Cur, true
}

// confirmScriptHandles completes the worker's descriptor duty: the session
// channel descriptors must be close-on-exec, so the interpreter inherits
// none of them. The check reads the process's own standard descriptors,
// which are the session channel in every real worker.
func confirmScriptHandles() error {
	for _, file := range []*os.File{os.Stdin, os.Stdout} {
		if file == nil {
			continue
		}
		flags, err := unix.FcntlInt(file.Fd(), unix.F_GETFD, 0)
		if err != nil {
			return diagnosticErr(CodeCapabilityEvidenceInvalid, err,
				"cannot confirm close-on-exec on the worker session channel")
		}
		if flags&unix.FD_CLOEXEC == 0 {
			if _, err := unix.FcntlInt(file.Fd(), unix.F_SETFD, flags|unix.FD_CLOEXEC); err != nil {
				return diagnosticErr(CodeCapabilityEvidenceInvalid, err,
					"cannot release the worker session channel")
			}
		}
		confirmed, err := unix.FcntlInt(file.Fd(), unix.F_GETFD, 0)
		if err != nil || confirmed&unix.FD_CLOEXEC == 0 {
			return diagnosticErr(CodeCapabilityEvidenceInvalid, err,
				"the worker session channel is not close-on-exec")
		}
	}
	return nil
}
