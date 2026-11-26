//go:build darwin

package homelease

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/relux-works/curator/internal/homelease/state"
	"golang.org/x/sys/unix"
)

// errKernelChanged marks an incarnation change inside one observation pass;
// the contract maps it to the mismatch row, never to a retried success.
var errKernelChanged = errors.New("kernel incarnation changed during observation pass")

// errProcessExited is the exact kernel-level refusal for a zombie: the
// process has exited, so its retained start time is not a live incarnation.
// VerifyBind maps it (like every non-change kernel failure) to
// environment_lease_authority_refused/authority_unavailable. Tests assert
// this sentinel, never any error, so an ESRCH from a lookup that cannot
// reach zombies cannot masquerade as the SZOMB refusal.
var errProcessExited = errors.New("process has exited")

func nativeStartTimeKind() string { return state.StartDarwin }

// proc_pidinfo flavor and call discipline from <sys/proc_info.h>. The
// proc_bsdinfo layout is parsed by offset so no C types enter the build.
const (
	procInfoCallPidInfo = 2   // PROC_INFO_CALL_PIDINFO
	procPidTBSDInfo     = 3   // PROC_PIDTBSDINFO
	procBsdInfoSize     = 136 // sizeof(struct proc_bsdinfo)
	bsdInfoOffStatus    = 4
	bsdInfoOffPID       = 12
	bsdInfoOffRUID      = 28
	bsdInfoOffStartSec  = 120
	bsdInfoOffStartUsec = 128
	procStatusZombie    = 5 // SZOMB in <sys/proc.h>
	// procPidTBSDInfoZombieArg is the PROC_PIDTBSDINFO arg that makes
	// XNU's proc_pidinfo_internal look zombies up (proc_find_zombref).
	// With arg 0, proc_find never returns a zombie and the call fails
	// ESRCH before any byte is parsed, so the SZOMB refusal in
	// readKernelProcess could never fire. Passing 1 returns the zombie's
	// proc_bsdinfo with pbi_status == SZOMB and leaves live processes on
	// the same lookup as before.
	procPidTBSDInfoZombieArg = 1
)

func readKernelProcess(pid int64) (state.Process, error) {
	bootBefore, err := darwinBootSessionUUID()
	if err != nil {
		return state.Process{}, err
	}
	info, err := darwinProcBsdInfo(pid)
	if err != nil {
		return state.Process{}, err
	}
	bootAfter, err := darwinBootSessionUUID()
	if err != nil {
		return state.Process{}, err
	}
	if bootBefore != bootAfter {
		return state.Process{}, errKernelChanged
	}
	gotPID := int64(binary.LittleEndian.Uint32(info[bsdInfoOffPID:]))
	if gotPID != pid {
		return state.Process{}, errKernelChanged
	}
	// A zombie has exited; its retained start time is not a live
	// incarnation Curator can bind to.
	if binary.LittleEndian.Uint32(info[bsdInfoOffStatus:]) == procStatusZombie {
		return state.Process{}, errProcessExited
	}
	uid := int64(binary.LittleEndian.Uint32(info[bsdInfoOffRUID:]))
	sec := binary.LittleEndian.Uint64(info[bsdInfoOffStartSec:])
	usec := binary.LittleEndian.Uint64(info[bsdInfoOffStartUsec:])
	if usec > 999999 {
		return state.Process{}, fmt.Errorf("process start microseconds outside range")
	}
	process := state.Process{PID: pid, UID: uid, BootID: bootBefore,
		StartTime: &state.StartTime{Kind: state.StartDarwin, Sec: strconv.FormatUint(sec, 10), Usec: int64(usec)}}
	if err := state.ValidateProcessSemantics(&process); err != nil {
		return state.Process{}, err
	}
	return process, nil
}

// darwinProcBsdInfo invokes the proc_pidinfo system call directly; its one
// call returns the complete proc_bsdinfo for a single incarnation. The
// zombie arg is required for that to be true of an exited-but-unreaped
// process too (see procPidTBSDInfoZombieArg).
func darwinProcBsdInfo(pid int64) ([]byte, error) {
	buf := make([]byte, procBsdInfoSize)
	r1, _, errno := unix.RawSyscall6(unix.SYS_PROC_INFO, procInfoCallPidInfo, //nolint:staticcheck // direct XNU proc_info syscall avoids cgo
		uintptr(pid), procPidTBSDInfo, procPidTBSDInfoZombieArg,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))) // #nosec G103 -- passes caller-owned fixed-size proc_bsdinfo buffer to proc_pidinfo.
	if errno != 0 {
		return nil, errno
	}
	if r1 != procBsdInfoSize {
		return nil, fmt.Errorf("short proc_pidinfo result")
	}
	return buf, nil
}

func darwinBootSessionUUID() (string, error) {
	value, err := unix.Sysctl("kern.bootsessionuuid")
	if err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(value)), nil
}
