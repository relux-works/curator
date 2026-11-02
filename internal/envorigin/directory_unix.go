//go:build unix

package envorigin

import (
	"golang.org/x/sys/unix"
	"os"
)

// Opening requires read access; inspecting "." relative to the held directory
// additionally requires search access. Neither operation enumerates names or
// opens a credential/config leaf. O_DIRECTORY prevents a substituted FIFO from
// blocking this metadata observation, and O_NOFOLLOW rejects a replaced link.
func openDirectoryMetadata(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, err
	}
	var info unix.Stat_t
	if err = unix.Fstatat(fd, ".", &info, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}
