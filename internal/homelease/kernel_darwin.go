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

func nativeStartTimeKind() string { return state.StartDarwin }

// proc_pidinfo flavor and call discipline from <sys/proc_info.h>. The
// proc_bsdinfo layout is parsed by offset so no C types enter the build.
const (
	procInfoCallPidInfo  = 2   // PROC_INFO_CALL_PIDINFO
	procPidTBSDInfo      = 3   // PROC_PIDTBSDINFO
	procBsdInfoSize      = 136 // sizeof(struct proc_bsdinfo)
	bsdInfoOffPID        = 12
	bsdInfoOffRUID       = 28
	bsdInfoOffStartSec   = 120
	bsdInfoOffStartUsec  = 128
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
// call returns the complete proc_bsdinfo for a single incarnation.
func darwinProcBsdInfo(pid int64) ([]byte, error) {
	buf := make([]byte, procBsdInfoSize)
	r1, _, errno := unix.RawSyscall6(unix.SYS_PROC_INFO, procInfoCallPidInfo,
		uintptr(pid), procPidTBSDInfo, 0,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
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
