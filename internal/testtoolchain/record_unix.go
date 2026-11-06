//go:build darwin || linux

package testtoolchain

import (
	"os"

	"golang.org/x/sys/unix"
)

func openHostGOROOTRecord(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK|unix.O_NOFOLLOW, 0) // #nosec G304 -- fixed diagnostic record, checked regular before reading.
}
