//go:build unix

package config

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func openClaudeAuthFile(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func privateClaudeAuthInfo(info os.FileInfo) bool {
	s, ok := info.Sys().(*syscall.Stat_t)
	return ok && info.Mode().IsRegular() && info.Mode() == 0600 && s.Uid == uint32(os.Geteuid()) && s.Nlink == 1
}

func claudeAuthFileIdentity(info os.FileInfo) string {
	s, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "unavailable"
	}
	return fmt.Sprintf("%d:%d:%d:%d:%d:%d", s.Dev, s.Ino, s.Uid, s.Mode, s.Size, info.ModTime().UnixNano())
}
