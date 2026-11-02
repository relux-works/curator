//go:build unix

package envunit

import (
	"fmt"
	"os"
	"syscall"
)

func metadataIdentity(_ string, info os.FileInfo) (string, error) {
	s, ok := info.Sys().(*syscall.Stat_t)
	if !ok || s == nil {
		return "", fmt.Errorf("native entry identity unavailable")
	}
	return fmt.Sprintf("%d:%d", s.Dev, s.Ino), nil
}
