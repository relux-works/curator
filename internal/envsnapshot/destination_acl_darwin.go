//go:build darwin

package envsnapshot

import (
	"encoding/binary"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// fgetattrlist reads the ACL of the retained object, not its pathname. Layouts
// and bounds are from sys/attr.h and sys/kauth.h (128 ACEs, 24 bytes each).
// A deny-only ACL on an ancestor cannot broaden Unix permissions. All grants,
// unknown ACE kinds and malformed/unreadable evidence refuse conservatively.
func destinationACL(f *os.File) error {
	attrs := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	var buf [4096]byte
	// #nosec G103 -- synchronous kernel ABI: fixed, live attrlist/buffer;
	// no pointers escape and every returned offset/length is checked below.
	_, _, errno := unix.Syscall6(unix.SYS_FGETATTRLIST, f.Fd(), uintptr(unsafe.Pointer(&attrs)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), unix.FSOPT_REPORT_FULLSIZE, 0) //nolint:staticcheck // SA1019: x/sys has no wrapper; qualified by kernel ACL tests, failure refuses closed.
	if errno != 0 {
		return errno
	}
	n := int(binary.NativeEndian.Uint32(buf[:4]))
	if n < 12 || n > len(buf) {
		return os.ErrPermission
	}
	rawOffset := binary.NativeEndian.Uint32(buf[4:8])
	if rawOffset > uint32(len(buf)-4) {
		return os.ErrPermission
	}
	offset := int(rawOffset) + 4
	length := int(binary.NativeEndian.Uint32(buf[8:12]))
	if length == 0 {
		return nil
	}
	if offset < 12 || length < 44 || offset > n || length > n-offset {
		return os.ErrPermission
	}
	sec := buf[offset : offset+length]
	if binary.NativeEndian.Uint32(sec[:4]) != 0x012cc16d {
		return os.ErrPermission
	}
	count := binary.NativeEndian.Uint32(sec[36:40])
	if count == 0xffffffff {
		return nil
	}
	if count > 128 || int(count)*24 > len(sec)-44 {
		return os.ErrPermission
	}
	for i := 0; i < int(count); i++ {
		flags := binary.NativeEndian.Uint32(sec[44+i*24+16 : 44+i*24+20])
		if flags&^uint32(0x1ff) != 0 {
			return os.ErrPermission
		}
		if flags&0xf != 2 { // KAUTH_ACE_DENY; a permit never qualifies.
			return os.ErrPermission
		}
	}
	return nil
}
