//go:build (darwin || linux) && !cgo

package envorigin

import (
	"fmt"
	"os"
	"syscall"
)

func trustedAccountExecutable(info os.FileInfo) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != 0 {
		return fmt.Errorf("native account utility is not OS-owned")
	}
	return nil
}
