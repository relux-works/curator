//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package envsnapshot

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// Every component is opened relative to its retained parent, never resolving a
// link. Nonblocking file opens prevent FIFO/device surprises from hanging reads.
func openMember(parent *os.File, name string, directory bool) (*os.File, error) {
	flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
	if directory {
		flags |= unix.O_DIRECTORY
	}
	fd, err := unix.Openat(int(parent.Fd()), name, flags, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), name), nil
}
func openDirectory(p string) (*os.File, error) {
	if p == "" {
		return nil, os.ErrInvalid
	}
	absolute := p
	if !filepath.IsAbs(p) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		// Preserve every caller component, including links before a parent segment.
		absolute = cwd + string(filepath.Separator) + p
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	current := os.NewFile(uintptr(fd), "/")
	for _, part := range strings.Split(strings.TrimPrefix(absolute, "/"), "/") {
		if part == "" {
			continue
		}
		next, err := openMember(current, part, true)
		_ = current.Close()
		if err != nil {
			return nil, err
		}
		current = next
	}
	return current, nil
}
func regularIndependent(info os.FileInfo) bool {
	if info == nil || !info.Mode().IsRegular() {
		return false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink == 1
}
func modeMatches(info os.FileInfo, mode string) bool {
	want := os.FileMode(0600)
	if mode == "0700" {
		want = 0700
	}
	return info.Mode() == want
}
func directoryMode(info os.FileInfo) bool { return info != nil && info.Mode() == os.ModeDir|0700 }
