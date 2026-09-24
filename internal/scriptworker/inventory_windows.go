//go:build windows

package scriptworker

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows native-control probes for the enforced script policy. The three
// job-backed controls create the exact kernel object with the exact limits
// of this invocation; the handle control proves the attribute list the
// launch will use is allocatable. The mechanisms are the same ones the
// go-v1 worker uses; only the bound values are script-fixed.

const (
	// scriptJobProcessBound is the exact active-process bound one enforced
	// script invocation installs. The magnitude matches the go-v1 default.
	scriptJobProcessBound = 64
	// scriptJobMemoryBound is the exact aggregate job and process memory
	// bound in bytes one enforced script invocation installs. The
	// magnitude matches the go-v1 default.
	scriptJobMemoryBound = int64(2 * 1024 * 1024 * 1024)
)

// inventoryPlatformUsable reports whether the test-only platform override
// names an inventory whose mechanisms exist on this host. Windows probes
// only the Windows inventory: unix mechanisms do not exist here.
func inventoryPlatformUsable(platform string) bool {
	return platform == ScriptPlatformWindows
}

// scriptJobLimitFlags maps one inventory control to the Job Object limit
// that carries it.
func scriptJobLimitFlags(name string) (uint32, error) {
	switch name {
	case ScriptControlDescendantDomainTermination:
		return windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE, nil
	case ScriptControlActiveProcessCountLimit:
		return windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS, nil
	case ScriptControlAggregateMemoryLimit:
		return windows.JOB_OBJECT_LIMIT_JOB_MEMORY | windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY, nil
	default:
		return 0, fmt.Errorf("no Windows Job Object limit for inventory control %q", name)
	}
}

// scriptJobLimitInformation is the exact limit block this invocation
// installs for the requested flags.
func scriptJobLimitInformation(flags uint32) windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION {
	information := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	information.BasicLimitInformation.LimitFlags = flags
	if flags&windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS != 0 {
		information.BasicLimitInformation.ActiveProcessLimit = uint32(scriptJobProcessBound)
	}
	if flags&windows.JOB_OBJECT_LIMIT_JOB_MEMORY != 0 {
		information.JobMemoryLimit = uintptr(scriptJobMemoryBound)
		information.ProcessMemoryLimit = uintptr(scriptJobMemoryBound)
	}
	return information
}

// newScriptControlJob creates one private Job Object carrying exactly the
// requested limits for exactly this invocation.
func newScriptControlJob(flags uint32) (windows.Handle, error) {
	information := scriptJobLimitInformation(flags)
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

// probeScriptMechanism performs the per-invocation availability
// determination for one control on Windows. It creates the exact kernel
// object the control will use with the exact limits of this invocation,
// starts no program, and caches nothing.
func probeScriptMechanism(_, name string) (bool, error) {
	switch name {
	case ScriptControlDescendantDomainTermination, ScriptControlActiveProcessCountLimit, ScriptControlAggregateMemoryLimit:
		flags, err := scriptJobLimitFlags(name)
		if err != nil {
			return false, err
		}
		job, err := newScriptControlJob(flags)
		if err != nil {
			return false, err
		}
		return true, windows.CloseHandle(job)
	case ScriptControlInheritedHandleRestriction:
		// os/exec restricts inheritance with
		// PROC_THREAD_ATTRIBUTE_HANDLE_LIST, so the attribute list must
		// be allocatable on this host.
		list, err := windows.NewProcThreadAttributeList(1)
		if err != nil {
			return false, err
		}
		list.Delete()
		return true, nil
	default:
		return false, diagnostic(CodeWorkerProtocolInvalid, "no Windows probe for inventory control %q", name)
	}
}
