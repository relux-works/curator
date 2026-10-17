//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !windows

package nativeseed

import (
	"fmt"
	"os"
)

func openSeed(string) (*os.File, error) {
	return nil, fmt.Errorf("native seed security checks are unavailable on this platform")
}
func checkSeedFile(*os.File, os.FileInfo) error {
	return fmt.Errorf("native seed security checks are unavailable on this platform")
}
