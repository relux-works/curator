//go:build darwin || linux

package projectadapter

import (
	"os"
	"syscall"
)

func museSingleLink(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink == 1
}
