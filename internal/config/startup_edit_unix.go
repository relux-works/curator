//go:build unix

package config

import (
	"os"
	"syscall"
)

func startupEditInfo(info os.FileInfo) bool {
	if info == nil {
		return false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && info.Mode().IsRegular() && info.Mode()&(0o022|os.ModeSetuid|os.ModeSetgid|os.ModeSticky) == 0 && int(stat.Uid) == os.Geteuid() && stat.Nlink == 1
}
