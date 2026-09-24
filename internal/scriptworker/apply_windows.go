//go:build windows

package scriptworker

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// confirmScriptTermination proves nothing by itself on Windows: every
// job-backed control is confirmed together by the job query in
// confirmScriptAggregateLimits, which requires at least the kill-on-close
// flag. Termination is installable on every Windows invocation, so the
// query always runs.
func confirmScriptTermination(_ *workerRequest) error { return nil }

// confirmScriptAggregateLimits queries the job the worker was assigned to
// and requires exactly the limits the manager installed for the
// installable job-backed controls.
func confirmScriptAggregateLimits(request *workerRequest, installable map[string]bool) error {
	var wanted uint32
	for _, name := range []string{
		ScriptControlDescendantDomainTermination,
		ScriptControlActiveProcessCountLimit,
		ScriptControlAggregateMemoryLimit,
	} {
		if !installable[name] {
			continue
		}
		flags, err := scriptJobLimitFlags(name)
		if err != nil {
			return diagnostic(CodeCapabilityEvidenceInvalid, "no Windows confirmation for inventory control %q", name)
		}
		wanted |= flags
	}
	if wanted == 0 {
		return nil
	}
	expected := scriptJobLimitInformation(wanted)
	var actual windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	var returned uint32
	// A zero handle queries the job object of the calling process, which
	// is the private job the manager assigned the worker to before any
	// session byte flowed.
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

// confirmScriptFileSize refuses: per-file-size-limit is never installable
// on the Windows inventory, so reaching here with it set is a request
// contradiction.
func confirmScriptFileSize(_ *workerRequest) error {
	return diagnostic(CodeCapabilityEvidenceInvalid, "per-file-size-limit is installable only on the unix inventories")
}

// currentScriptFileSizeLimit reports no bound: Windows has no per-file
// soft bound to observe.
func currentScriptFileSizeLimit() (uint64, bool) { return 0, false }

// confirmScriptHandles proves the explicit inherited-handle list mechanism
// is usable here: the attribute list the launch used must be allocatable.
func confirmScriptHandles() error {
	list, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err,
			"cannot confirm the explicit inherited-handle list in the worker")
	}
	list.Delete()
	return nil
}
