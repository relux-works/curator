//go:build linux || darwin

package atomicity

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func gitFixtureRlimits() string {
	var nofile, nproc unix.Rlimit
	nofileErr := unix.Getrlimit(unix.RLIMIT_NOFILE, &nofile)
	nprocErr := unix.Getrlimit(unix.RLIMIT_NPROC, &nproc)
	nofileText := fmt.Sprintf("cur=%d max=%d", nofile.Cur, nofile.Max)
	if nofileErr != nil {
		nofileText = nofileErr.Error()
	}
	nprocText := fmt.Sprintf("cur=%d max=%d", nproc.Cur, nproc.Max)
	if nprocErr != nil {
		nprocText = nprocErr.Error()
	}
	return fmt.Sprintf("NOFILE %s; NPROC %s", nofileText, nprocText)
}
