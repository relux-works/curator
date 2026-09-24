package scriptworker

// Shared Landlock vocabulary for the Linux host-conditional controls
// `descendant-exec-denial` (execute right) and
// `filesystem-write-confinement` (write rights). This file carries no
// system call: the ABI-gated handled masks, the per-object-type rule
// rights, and the request helpers below are pure, so they compile and
// test identically on every platform. The Linux mechanism — probing,
// ruleset construction, and spawn-time enforcement — lives in
// landlock_linux.go; every other platform reports the controls absent
// in landlock_other.go and never enforces them.

// Landlock filesystem access rights and their ABI introduction, from
// include/uapi/linux/landlock.h (values in landlock_const_linux.go,
// mirrored off Linux in landlock_const_other.go):
//
//	ABI 1: execute, write-file, remove-dir, remove-file, make-char,
//	  make-dir, make-reg, make-sock, make-fifo, make-block, make-sym
//	  (plus read-file, read-dir — never handled: reads stay
//	  unrestricted, so the write confinement is not a read
//	  confinement and the interpreter keeps reading the system
//	  objects outside its grants)
//	ABI 2: refer
//	ABI 3: truncate
//	ABI 4: no new filesystem right (TCP bind/connect are a separate
//	  handled set, never used; the network namespace covers them)
//	ABI 5: ioctl-dev (character and block devices only; the safe
//	  common ioctls stay invokable without it)
//
// Newer filesystem rights (resolve-unix on ABI 9 and later) are beyond
// this build's UAPI and stay unhandled — and therefore allowed — on
// kernels that provide them. A rule carrying a right the running
// kernel's ABI does not know fails with EINVAL, so the handled mask
// below gates every right on the probed ABI and the enforcement never
// requests more than the ABI provides; conversely every right the ABI
// provides for the installable controls IS handled, because Landlock
// permits every unhandled action.
//
// landlockMutationABI1Access is the handled directory-mutation subset
// available since ABI 1: entry removal and entry creation beneath a
// granted directory. REFER (reparenting an entry across directories)
// starts at ABI 2, so it stays out of this set and is gated on the
// probed ABI below.
const landlockMutationABI1Access = landlockAccessFSRemoveDir |
	landlockAccessFSRemoveFile |
	landlockAccessFSMakeChar |
	landlockAccessFSMakeDir |
	landlockAccessFSMakeReg |
	landlockAccessFSMakeSock |
	landlockAccessFSMakeFifo |
	landlockAccessFSMakeBlock |
	landlockAccessFSMakeSym

// landlockDirectoryOnlyAccess is the handled subset that applies only
// to directory content (creation, removal, and reparenting beneath a
// granted directory): a rule on a file or device carrying one of these
// rights returns EINVAL. The remaining handled rights — execute, write,
// truncate, device ioctl — are file rights, valid on files (truncate
// governs truncate(2), ftruncate(2), creat(2), and open(2) with
// O_TRUNC; overwriting an existing file needs it in addition to
// write-file) and, for directory rules, on the hierarchy beneath.
const landlockDirectoryOnlyAccess = landlockMutationABI1Access |
	landlockAccessFSRefer

// landlockHandledForABI returns the handled access set for the
// installable controls on a kernel with the given ABI version: execute
// for descendant-exec-denial, and the full mutation set for
// filesystem-write-confinement — write plus the entry
// creation/removal rights from ABI 1, reparenting (refer) on ABI 2+,
// truncation on ABI 3+, and device ioctl on ABI 5+. An ABI below 1
// handles nothing.
func landlockHandledForABI(abi int, execDenial, writeConfinement bool) uint64 {
	if abi < 1 {
		return 0
	}
	var handled uint64
	if execDenial {
		handled |= landlockAccessFSExecute
	}
	if writeConfinement {
		handled |= landlockAccessFSWriteFile | landlockMutationABI1Access
		if abi >= 2 {
			handled |= landlockAccessFSRefer
		}
		if abi >= 3 {
			handled |= landlockAccessFSTruncate
		}
		if abi >= 5 {
			handled |= landlockAccessFSIoctlDev
		}
	}
	return handled
}

// landlockWriteAccessForABI returns the access rights one writable
// grant requests on a kernel with the given ABI version: the full
// handled write set for that ABI. The caller still narrows the result
// per object type (see landlockRuleRights): directory grants keep the
// mutation rights, file grants keep the file rights.
func landlockWriteAccessForABI(abi int) uint64 {
	return landlockHandledForABI(abi, false, true)
}

// landlockFileAccessForABI returns the file-typed subset of the
// writable grant for the given ABI version: write, truncation, and
// device ioctl where the ABI provides them. The null-device grant the
// write confinement adds carries exactly this set, so the sink stays
// fully usable under confinement (a script may open it with O_TRUNC,
// and descendants inherit redirected standard streams through it)
// while every directory-only right stays off the non-directory.
func landlockFileAccessForABI(abi int) uint64 {
	return landlockWriteAccessForABI(abi) &^ landlockDirectoryOnlyAccess
}

// landlockRuleRights narrows the requested access to the rights valid
// for the grant's object type within the ruleset: a rule may only carry
// rights the ruleset handles, and only the rights that apply to the
// object type — a rule on a file or device with a directory-only right
// returns EINVAL, so non-directory grants shed the creation, removal,
// and reparenting rights and keep the file rights (write, truncate,
// device ioctl, execute). The worker's pre-bound standard input needs
// no rule (Landlock never restricts an already-open descriptor); the
// null-device grant the write confinement adds is file-typed for the
// same reason.
func landlockRuleRights(isDir bool, access, handled uint64) uint64 {
	granted := access & handled
	if !isDir {
		granted &^= landlockDirectoryOnlyAccess
	}
	return granted
}

// landlockGrantsForRequest derives the exact Landlock grant set of one
// invocation from the worker request: the verified interpreter file and
// the manager-owned PATH farm for descendant-exec-denial, and the
// operation-private area plus the derived path set for
// filesystem-write-confinement. Both the pre-ready construction
// confirmation and the spawn-time enforcement derive it here, so the
// confirmed ruleset and the enforced domain cannot diverge. The
// enforcement adds the fixed file-typed null-device grant on Linux; it
// is not request-derived, so it lives in the ruleset constructor, not
// here.
func landlockGrantsForRequest(request *workerRequest) landlockGrants {
	return landlockGrants{
		executables: []string{request.InterpreterPath, request.FarmDir},
		writables:   append([]string{request.PrivateBase}, request.WritePaths...),
	}
}

// requestLandlockInstallable reports whether the request inventory
// installs the Landlock controls for the interpreter child: each
// control present and host-conditional.
func requestLandlockInstallable(request *workerRequest) (execDenial, writeConfinement bool) {
	for _, input := range request.Inventory {
		if !input.Present || input.Availability != ScriptAvailabilityHostConditional {
			continue
		}
		switch input.Name {
		case ScriptControlDescendantExecDenial:
			execDenial = true
		case ScriptControlFilesystemWriteConfinement:
			writeConfinement = true
		}
	}
	return execDenial, writeConfinement
}
