//go:build !windows

package snapcache

import (
	"fmt"
	"io/fs"
	"syscall"
)

// allocatedSize is the storage the platform reports as allocated to the file:
// st_blocks in 512-byte units.
func allocatedSize(info fs.FileInfo) int64 {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return info.Size()
	}
	return int64(stat.Blocks) * 512
}

// fileIdentity is the device and inode pair of a multiply linked file. A file
// with one link cannot be counted twice, so it needs no identity.
func fileIdentity(info fs.FileInfo) (string, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil || stat.Nlink <= 1 {
		return "", false
	}
	return fmt.Sprintf("%d:%d", stat.Dev, stat.Ino), true
}

func isReparsePoint(fs.FileInfo) bool { return false }
