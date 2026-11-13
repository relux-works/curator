//go:build unix

package pathboundary

import (
	"fmt"
	"os"
	"syscall"
)

func singleLink(_ string, info os.FileInfo) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return fmt.Errorf("regular-file link metadata is unavailable")
	}
	if stat.Nlink != 1 {
		return fmt.Errorf("regular file must have exactly one link (observed %d)", stat.Nlink)
	}
	return nil
}
