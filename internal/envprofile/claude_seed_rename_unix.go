//go:build darwin || linux

package envprofile

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// prepareClaudeSeedRename is seed-specific: unlike os.Rename it does not stat
// the destination on publication. Encode both paths before the final seed read.
func prepareClaudeSeedRename(source, target string) (func() error, func(), error) {
	from, err := unix.BytePtrFromString(source)
	if err != nil {
		return nil, nil, &os.LinkError{Op: "rename", Old: source, New: target, Err: err}
	}
	to, err := unix.BytePtrFromString(target)
	if err != nil {
		return nil, nil, &os.LinkError{Op: "rename", Old: source, New: target, Err: err}
	}
	dirfd := unix.AT_FDCWD
	return func() error {
		var errno syscall.Errno
		for {
			_, _, errno = syscall.Syscall6(claudeSeedRenameTrap,
				uintptr(dirfd), uintptr(unsafe.Pointer(from)), uintptr(dirfd), uintptr(unsafe.Pointer(to)), 0, 0)
			// Preserve os.Rename's interrupted-call retry, without another stat.
			if errno != syscall.EINTR {
				break
			}
		}
		runtime.KeepAlive(from)
		runtime.KeepAlive(to)
		if errno != 0 {
			return &os.LinkError{Op: "rename", Old: source, New: target, Err: errno}
		}
		return nil
	}, func() {}, nil
}
