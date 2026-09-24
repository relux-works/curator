//go:build unix

package scriptworker

import (
	"sync"

	"golang.org/x/sys/unix"
)

// scriptFileSizeBound is the exact per-file soft bound one enforced script
// invocation installs where the inventory marks per-file-size-limit
// available. It is manager-selected and fixed: the profile names the
// mechanism, not a value, and the magnitude matches the go-v1 default so
// one number needs no per-policy justification.
const scriptFileSizeBound = int64(512 * 1024 * 1024)

// scriptFileLimitMutex serializes the manager-side RLIMIT_FSIZE window. The
// soft bound is process-wide, so probing and the launch window must not
// overlap.
var scriptFileLimitMutex sync.Mutex

// wantedScriptFileLimit is the exact per-file soft bound this invocation
// installs: the manager bound clamped to the inherited hard bound, so the
// parent and the worker compute the same value.
func wantedScriptFileLimit() (uint64, error) {
	var current unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &current); err != nil {
		return 0, err
	}
	wanted := uint64(scriptFileSizeBound)
	if current.Max != unix.RLIM_INFINITY && wanted > current.Max {
		wanted = current.Max
	}
	return wanted, nil
}

// inventoryPlatformUsable reports whether the test-only platform override
// names an inventory whose mechanisms exist on this host. Unix hosts probe
// the Linux and macOS inventories; the Windows inventory needs Job Objects.
func inventoryPlatformUsable(platform string) bool {
	return platform == ScriptPlatformLinux || platform == ScriptPlatformMacOS
}

// probeScriptMechanism performs the per-invocation availability
// determination for one control on a unix host. It performs the exact
// operation the control will perform for this invocation, starts no
// program except the self-reexec probes (network isolation and Landlock),
// and caches nothing.
func probeScriptMechanism(_, name string) (bool, error) {
	switch name {
	case ScriptControlDescendantDomainTermination:
		// Session and process-group teardown requires a usable process
		// group and permission to signal it. Signal 0 performs the
		// permission check only. The private session itself is created
		// by the kernel between fork and exec, so it cannot fail once
		// the worker is executing.
		group, err := unix.Getpgid(0)
		if err != nil {
			return false, err
		}
		if err := unix.Kill(-group, 0); err != nil {
			return false, err
		}
		return true, nil
	case ScriptControlPerFileSizeLimit:
		// Prove the exact byte bound this invocation will install is
		// settable, then restore the manager's own bound immediately.
		wanted, err := wantedScriptFileLimit()
		if err != nil {
			return false, err
		}
		scriptFileLimitMutex.Lock()
		defer scriptFileLimitMutex.Unlock()
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
	case ScriptControlInheritedHandleRestriction:
		// Close-on-exec must be readable and settable for the
		// descriptors the worker holds when the interpreter starts.
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
	case ScriptControlActiveProcessCountLimit, ScriptControlAggregateMemoryLimit:
		return probeScriptCgroup(name)
	case ScriptControlDescendantExecDenial, ScriptControlFilesystemWriteConfinement:
		return probeScriptLandlock(name)
	case ScriptControlNetworkIsolationDomain:
		return probeScriptNetNS()
	default:
		return false, diagnostic(CodeWorkerProtocolInvalid, "no unix probe for inventory control %q", name)
	}
}
