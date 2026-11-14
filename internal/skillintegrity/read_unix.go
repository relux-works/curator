//go:build unix

package skillintegrity

import (
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

func singleLink(_ string, info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink == 1
}
func openCaptureFile(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}
