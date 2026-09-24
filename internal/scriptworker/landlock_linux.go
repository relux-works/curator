//go:build linux

package scriptworker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Landlock application for the Linux host-conditional controls
// `descendant-exec-denial` (execute right) and `filesystem-write-confinement`
// (write rights).
//
// Enforcement happens in the WORKER process, on a locked OS thread,
// immediately before the interpreter is spawned, from that same thread:
// LockOSThread, no_new_privs, ruleset creation with the ABI-gated
// handled mask, one typed rule per grant path, restrict_self, then the
// fork/exec of the interpreter from the locked thread, so the child
// inherits the domain. landlock_restrict_self returns EPERM unless the
// calling thread has no_new_privs (or CAP_SYS_ADMIN), it restricts the
// calling thread only, and Go is multi-threaded — restricting anywhere
// else, or from an unlocked goroutine the runtime may migrate between
// the privilege change and the restriction, cannot work. The locked
// thread stays locked until the single-session worker exits: returning
// a restricted thread to the runtime pool would sandbox unrelated
// goroutines.
//
// The pre-ready apply never restricts: it only constructs the exact
// ruleset the spawn will enforce and discards it, proving the paths,
// the ABI mask, and the rule typing before the worker claims `applied`.
// The probe never restricts the manager either: it re-executes this
// manager in the hidden probe mode below, and the short-lived probe
// child performs the full sequence — no_new_privs, ruleset, rule,
// restrict — against a directory the parent owns, then exits. A zero
// exit proves the host provides the controls; anything else reports
// absent. x/sys exposes the syscall numbers but no wrappers, so the
// Landlock calls below use unix.Syscall directly.

// ScriptLandlockProbeMode is the fixed hidden probe the Landlock
// availability check re-executes with a manager-owned directory: the
// child enforces the exact sequence the interpreter spawn will use and
// exits 0 only when every step succeeds. Like the worker mode it is an
// implementation boundary, not a user-visible command: no package file,
// manifest value, environment value, PATH lookup, shell, or user option
// selects it, and restricting itself can only reduce what the child may
// do.
const ScriptLandlockProbeMode = "__curator-script-landlock-probe"

// scriptLandlockProbeTimeout bounds one Landlock availability probe.
const scriptLandlockProbeTimeout = 30 * time.Second

// landlockRulePathBeneath is the only rule type used.
const landlockRulePathBeneath = 1

// landlockRulesetAttr is struct landlock_ruleset_attr with the filesystem
// field only: every filesystem ABI shares that first u64, so size 8
// covers the handled filesystem rights on any ABI while the network
// and scope sets stay unhandled (never requested).
type landlockRulesetAttr struct {
	handledAccessFS uint64
}

// landlockPathBeneathAttr is struct landlock_path_beneath_attr, packed.
type landlockPathBeneathAttr struct {
	allowedAccess uint64
	parentFD      int32
}

// landlockABI queries the running kernel's Landlock ABI version. An error
// means the kernel does not implement Landlock.
func landlockABI() (int, error) {
	version, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET, 0, 0, uintptr(unix.LANDLOCK_CREATE_RULESET_VERSION))
	if errno != 0 {
		return 0, errno
	}
	return int(version), nil
}

// probeScriptLandlock determines whether this host provides the Landlock
// rights the named control needs: it queries the ABI, then re-executes
// this manager in the hidden probe mode, where the child performs the
// exact enforcement sequence — no_new_privs, ruleset creation, rule,
// self-restriction — the interpreter spawn will use. It never restricts
// the calling process. Every expected-absence condition — no Landlock,
// an older ABI, a denied privilege change, a failed restriction —
// reports absent without an error.
func probeScriptLandlock(_ string) (bool, error) {
	abi, err := landlockABI()
	if err != nil || abi < 1 {
		return false, nil
	}
	manager, err := os.Executable()
	if err != nil {
		return false, nil
	}
	temporary, err := os.MkdirTemp("", "curator-landlock-probe-")
	if err != nil {
		return false, nil
	}
	defer func() { _ = os.Remove(temporary) }()
	ctx, cancel := context.WithTimeout(context.Background(), scriptLandlockProbeTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, manager, ScriptLandlockProbeMode, temporary) // #nosec G204 -- self-re-execution in the fixed hidden probe mode with a manager-created directory
	command.Env = scriptWorkerEnvironment()
	if err := command.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

// RunLandlockProbe is the hidden probe entry: on a locked OS thread it
// sets no_new_privs, creates the invocation-shaped ruleset for the
// probed ABI, adds one typed rule over the manager-owned directory, and
// restricts itself, then exits 0. Any failure exits 1, so the exit
// status alone answers whether this host enforces the controls. It
// prints nothing.
func RunLandlockProbe(directory string) int {
	// Locked for the same reason as the spawn-time enforcement: the
	// privilege change and the restriction must happen on one thread,
	// and the process exits immediately after.
	runtime.LockOSThread()
	if err := landlockSetNoNewPrivs(); err != nil {
		return 1
	}
	abi, err := landlockABI()
	if err != nil || abi < 1 {
		return 1
	}
	handled := landlockHandledForABI(abi, true, true)
	if handled == 0 {
		return 1
	}
	ruleset, err := landlockCreateRuleset(handled)
	if err != nil {
		return 1
	}
	defer func() { _ = unix.Close(ruleset) }()
	if err := landlockAddGrant(ruleset, directory, handled, handled); err != nil {
		return 1
	}
	if err := landlockRestrictSelf(ruleset); err != nil {
		return 1
	}
	return 0
}

// confirmScriptLandlock proves the invocation can enforce the installable
// Landlock controls: it constructs the exact ruleset the interpreter
// spawn will enforce — same ABI mask, same typed rules over the same
// grant paths — and discards it without restricting the caller. Each
// installable control is confirmed independently from its own grants: a
// writable path that cannot be ruled (a derived member that does not
// exist) fails only filesystem-write-confinement, never
// descendant-exec-denial. A nil error confirms the control; a non-nil
// error carries the cause the worker refuses with — the worker never
// reports an installable control `unavailable`, which would contradict
// the probe that found it present.
func confirmScriptLandlock(execDenial, writeConfinement bool, grants landlockGrants) (execErr, writeErr error) {
	if !execDenial && !writeConfinement {
		return nil, nil
	}
	ruleset, execErr, writeErr := landlockConstructRuleset(execDenial, writeConfinement, grants)
	if ruleset >= 0 {
		_ = unix.Close(ruleset)
	}
	return execErr, writeErr
}

// enforceScriptLandlock enforces the installable Landlock controls on
// the calling thread: no_new_privs, the ABI-gated ruleset, one typed
// rule per grant path, then self-restriction. The caller MUST hold
// runtime.LockOSThread across this call and the interpreter fork/exec
// that follows it, so the privilege change, the restriction, and the
// fork all happen on one thread and the child inherits the domain. It
// returns a plain error the caller classifies: an enforcement failure
// is a worker-side refusal before the interpreter starts, never a
// probe/evidence contradiction the worker could report.
func enforceScriptLandlock(execDenial, writeConfinement bool, grants landlockGrants) error {
	if !execDenial && !writeConfinement {
		return nil
	}
	if err := landlockSetNoNewPrivs(); err != nil {
		return fmt.Errorf("cannot set no_new_privs for the Landlock domain: %w", err)
	}
	ruleset, execErr, writeErr := landlockConstructRuleset(execDenial, writeConfinement, grants)
	if ruleset < 0 {
		return errors.Join(execErr, writeErr)
	}
	defer func() { _ = unix.Close(ruleset) }()
	if err := landlockRestrictSelf(ruleset); err != nil {
		return fmt.Errorf("cannot enforce the Landlock ruleset: %w", err)
	}
	return nil
}

// landlockConstructRuleset creates the invocation ruleset for the
// probed ABI and adds every typed rule for the installable controls:
// the execute right over the manager-resolved executables, the full
// handled write set (write, entry creation/removal/reparenting,
// truncation, device ioctl per ABI) over the operation-private area
// and the derived path set, and — under write confinement — a
// file-typed rule over the null device, which the confined interpreter
// legitimately opens (redirected standard streams, descendant stdio).
// Directory grants keep the mutation rights; file grants keep the file
// rights. It never restricts the caller; the caller closes the returned
// ruleset when it is valid (>= 0). Both the pre-ready confirmation and
// the spawn-time enforcement build it here, so the confirmed ruleset
// and the enforced domain cannot diverge. Rule failures are collected
// per control: a grant that cannot be ruled fails only its own control,
// and a ruleset-level failure (no ABI, no ruleset) fails every
// installable control with the same cause.
func landlockConstructRuleset(execDenial, writeConfinement bool, grants landlockGrants) (ruleset int, execErr, writeErr error) {
	fail := func(err error) (int, error, error) {
		if execDenial {
			execErr = err
		}
		if writeConfinement {
			writeErr = err
		}
		return -1, execErr, writeErr
	}
	abi, err := landlockABI()
	if err != nil {
		return fail(fmt.Errorf("landlock is not provided by the running kernel: %w", err))
	}
	if abi < 1 {
		return fail(fmt.Errorf("landlock is not provided by the running kernel"))
	}
	handled := landlockHandledForABI(abi, execDenial, writeConfinement)
	if handled == 0 {
		return fail(fmt.Errorf("no Landlock control is installable for ABI %d", abi))
	}
	ruleset, err = landlockCreateRuleset(handled)
	if err != nil {
		return fail(fmt.Errorf("cannot create the Landlock ruleset: %w", err))
	}
	addGrants := func(paths []string, access uint64) error {
		for _, path := range paths {
			if err := landlockAddGrant(ruleset, path, access, handled); err != nil {
				return err
			}
		}
		return nil
	}
	if execDenial {
		execErr = addGrants(grants.executables, landlockAccessFSExecute)
	}
	if writeConfinement {
		writeErr = addGrants(grants.writables, landlockWriteAccessForABI(abi))
		if writeErr == nil {
			// The null device is reachable under confinement: any
			// script may redirect to it, and descendant spawns with
			// redirected stdio open it for writing. Reads are never
			// handled, so no read rule is needed or possible (a rule
			// may only carry handled rights); the rule carries the
			// file-typed write rights — write plus truncation and
			// device ioctl where the ABI provides them, so an open
			// with O_TRUNC works on the sink too — and everything
			// else stays denied.
			writeErr = landlockAddGrant(ruleset, os.DevNull, landlockFileAccessForABI(abi), handled)
		}
	}
	if execErr != nil || writeErr != nil {
		_ = unix.Close(ruleset)
		return -1, execErr, writeErr
	}
	return ruleset, nil, nil
}

// landlockAddGrant adds one typed path rule: the granted set is the
// requested access narrowed to the handled set and to the rights that
// apply to the grant's object type (see landlockRuleRights).
func landlockAddGrant(ruleset int, path string, access, handled uint64) error {
	directory, err := unix.Open(path, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("cannot grant Landlock rights on %q: %w", path, err)
	}
	defer func() { _ = unix.Close(directory) }()
	var stat unix.Stat_t
	if err := unix.Fstat(directory, &stat); err != nil {
		return fmt.Errorf("cannot inspect the Landlock grant path %q: %w", path, err)
	}
	granted := landlockRuleRights(stat.Mode&unix.S_IFMT == unix.S_IFDIR, access, handled)
	if err := landlockAddPathRule(ruleset, directory, granted); err != nil {
		return fmt.Errorf("cannot add the Landlock rule for %q: %w", path, err)
	}
	return nil
}

// landlockSetNoNewPrivs sets the calling thread's no_new_privs bit, the
// precondition for landlock_restrict_self without CAP_SYS_ADMIN.
func landlockSetNoNewPrivs() error {
	return unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)
}

func landlockCreateRuleset(handled uint64) (int, error) {
	attr := landlockRulesetAttr{handledAccessFS: handled}
	fd, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET,
		uintptr(unsafe.Pointer(&attr)), unsafe.Sizeof(attr), 0) // #nosec G103 -- raw syscall args for landlock_create_ruleset; no memory escapes
	if errno != 0 {
		return -1, errno
	}
	return int(fd), nil // #nosec G115 -- the kernel returns a small non-negative descriptor
}

func landlockAddPathRule(ruleset, parent int, allowed uint64) error {
	attr := landlockPathBeneathAttr{allowedAccess: allowed, parentFD: int32(parent)} // #nosec G115 -- unix.Open returns a small non-negative descriptor
	_, _, errno := unix.Syscall(unix.SYS_LANDLOCK_ADD_RULE,
		uintptr(ruleset), uintptr(landlockRulePathBeneath), uintptr(unsafe.Pointer(&attr))) // #nosec G103 -- raw syscall args for landlock_add_rule; no memory escapes
	if errno != 0 {
		return errno
	}
	return nil
}

// landlockRestrictSelf enforces the ruleset on the calling thread. The
// caller must have set no_new_privs on this thread first.
func landlockRestrictSelf(ruleset int) error {
	_, _, errno := unix.Syscall(unix.SYS_LANDLOCK_RESTRICT_SELF, uintptr(ruleset), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
