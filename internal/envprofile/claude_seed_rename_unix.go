//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package envprofile

import (
	"os"

	"golang.org/x/sys/unix"
)

// The seed is staged beside a validated regular-file destination. Unlike
// os.Rename, unix.Rename does not Lstat the target before the native rename.
// Do not use the generic entry replacement's directory/link fallback here.
func prepareClaudeSeedRename(source, target string) (func() error, error) {
	return func() error {
		var err error
		for {
			err = unix.Rename(source, target)
			if err != unix.EINTR {
				break
			}
		}
		if err != nil {
			return &os.LinkError{Op: "rename", Old: source, New: target, Err: err}
		}
		return nil
	}, nil
}
