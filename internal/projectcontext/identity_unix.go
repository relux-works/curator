//go:build darwin || linux

package projectcontext

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"strconv"
	"syscall"
)

func filesystemIdentity(st os.FileInfo) (map[string]string, error) {
	s, ok := st.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("filesystem identity unavailable")
	}
	return map[string]string{"device": strconv.FormatUint(uint64(s.Dev), 10), "inode": strconv.FormatUint(uint64(s.Ino), 10)}, nil
}

func openCandidate(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|unix.O_NONBLOCK|unix.O_NOFOLLOW, 0)
}

func singleLink(st os.FileInfo) bool { s, ok := st.Sys().(*syscall.Stat_t); return ok && s.Nlink == 1 }
