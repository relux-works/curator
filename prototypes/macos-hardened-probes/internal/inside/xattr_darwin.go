//go:build darwin

package inside

import (
	"syscall"
	"unsafe"
)

// setxattr sets an extended attribute. The Go standard library exposes no
// wrapper on darwin and this module has no external dependencies, so the
// syscall is issued directly.
//
// Section 5.2 lists extended attribute change among the operations a read-only
// view must refuse, so the probe cannot skip it.
func setxattr(path, name string, value []byte) error {
	pathPtr, err := syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}
	namePtr, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	if len(value) == 0 {
		value = []byte{0}
	}
	// The three unsafe conversions below are the documented calling convention
	// for a raw syscall: each pointer is passed directly to the kernel in the
	// same expression that produces it, so nothing here can be moved by the
	// collector between the conversion and the call.
	_, _, errno := syscall.Syscall6(
		syscall.SYS_SETXATTR,
		uintptr(unsafe.Pointer(pathPtr)),   //nolint:gosec // syscall argument, see above
		uintptr(unsafe.Pointer(namePtr)),   //nolint:gosec // syscall argument, see above
		uintptr(unsafe.Pointer(&value[0])), //nolint:gosec // syscall argument, see above
		uintptr(len(value)),
		0, // position, meaningful only for resource forks
		0, // options
	)
	if errno != 0 {
		return errno
	}
	return nil
}
