package envprofile

import (
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// prepareClaudeSeedRename validates paths before the final seed read. Renameat
// uses libSystem and atomically replaces the destination without a stat.
func prepareClaudeSeedRename(source, target string) (func() error, func(), error) {
	if strings.IndexByte(source, 0) >= 0 || strings.IndexByte(target, 0) >= 0 {
		return nil, nil, &os.LinkError{Op: "rename", Old: source, New: target, Err: unix.EINVAL}
	}
	return func() error {
		var err error
		for {
			err = unix.Renameat(unix.AT_FDCWD, source, unix.AT_FDCWD, target)
			if err != unix.EINTR {
				break
			}
		}
		if err != nil {
			return &os.LinkError{Op: "rename", Old: source, New: target, Err: err}
		}
		return nil
	}, func() {}, nil
}
