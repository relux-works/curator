//go:build unix

package transaction

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func syncRegular(path string) error {
	file, err := os.Open(path) // #nosec G304 -- transaction-owned target
	if err != nil {
		return err
	}
	return errors.Join(file.Sync(), file.Close())
}

func journalMode() os.FileMode {
	return 0o600
}

func syncDirectory(path string) error {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	return errors.Join(unix.Fsync(fd), unix.Close(fd))
}

func durableRename(from, to string) error {
	if err := os.Rename(from, to); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(to)); err != nil {
		return err
	}
	if filepath.Dir(from) != filepath.Dir(to) {
		return syncDirectory(filepath.Dir(from))
	}
	return nil
}
