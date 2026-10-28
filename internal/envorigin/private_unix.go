//go:build unix

package envorigin

import (
	"os"
	"syscall"
)

func singleLink(_ string, info os.FileInfo) bool {
	s, ok := info.Sys().(*syscall.Stat_t)
	return ok && s.Nlink == 1
}
func privateMode(_ string, info os.FileInfo) bool { return info.Mode().Perm()&0077 == 0 }
