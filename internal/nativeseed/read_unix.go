//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package nativeseed

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func openSeed(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func checkSeedFile(_ *os.File, info os.FileInfo) error {
	return checkSeedInfo(info, uint32(os.Geteuid()))
}

func checkSeedInfo(info os.FileInfo, uid uint32) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("cannot prove native seed target ownership")
	}
	if stat.Uid != uid {
		return fmt.Errorf("native seed target is not owned by the current user")
	}
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("native seed target must not be group- or other-writable")
	}
	return nil
}
