package buildcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// Linux never removes a build any process is executing, with the accepted
// exception of an executable opened under a fanotify pre-content/HSM watch.
// Linux deliberately exempts those opens from write denial (0357ef03c94e,
// following removal 2a010c412853 and revert 3b832035387f). Readable /proc paths
// and inodes remain extra signals: successful write-open alone is not unused.
// The exception can remove the path without stopping the running process; a
// later exec fails until rebuilt. See docs/linux-build-cache-sweep.md.
// Probes never truncate, create, chmod or write bytes, and close immediately.
func entryExecutionInUse(root *sweepRoot, name string, snapshot executableSnapshot) (bool, error) {
	fd, err := unix.Openat(int(root.validated.dir.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if errors.Is(err, unix.ENOENT) || errors.Is(err, unix.ENOTDIR) || errors.Is(err, unix.ELOOP) {
		return false, nil
	} // Unlinking a retired link cannot remove its target.
	if err != nil {
		return false, fmt.Errorf("open candidate for execution probe: %w", err)
	}
	entry := os.NewFile(uintptr(fd), name)
	defer func() { _ = entry.Close() }()
	return probeLinuxTree(entry, "", snapshot)
}

// Partial retired trees need no receipts: inspect remaining executable inodes
// through the proven root, never through a replaced pathname or a symlink.
func probeLinuxTree(dir *os.File, rel string, snapshot executableSnapshot) (bool, error) {
	if err := validateUnixDir(int(dir.Fd()), "execution probe directory"); err != nil {
		return false, err
	}
	names, err := directoryNames(dir)
	if err != nil {
		return false, fmt.Errorf("list execution probe directory %q: %w", rel, err)
	}
	for _, name := range names {
		child := filepath.Join(rel, name)
		var stat unix.Stat_t
		if err := unix.Fstatat(int(dir.Fd()), name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return false, fmt.Errorf("stat execution probe %q: %w", child, err)
		}
		switch stat.Mode & unix.S_IFMT {
		case unix.S_IFDIR:
			fd, err := unix.Openat(int(dir.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if err != nil {
				return false, fmt.Errorf("open execution probe directory %q: %w", child, err)
			}
			sub := os.NewFile(uintptr(fd), name)
			busy, probeErr := probeLinuxTree(sub, child, snapshot)
			closeErr := sub.Close()
			if probeErr != nil {
				return false, probeErr
			}
			if closeErr != nil {
				return false, fmt.Errorf("close execution probe directory %q: %w", child, closeErr)
			}
			if busy {
				return true, nil
			}
		case unix.S_IFREG:
			// Every bin member is an artifact even if its mode changed after exec.
			if !strings.HasPrefix(child, "bin/") && stat.Mode&0o111 == 0 {
				continue
			}
			probeErr := probeLinuxExecutable(int(dir.Fd()), name, &stat)
			if errors.Is(probeErr, unix.ETXTBSY) {
				return true, nil
			}
			if probeErr != nil {
				return false, fmt.Errorf("write-open execution probe %q: %w", child, probeErr)
			}
			// Independently preserve readable inode and sweep-time mode evidence.
			if linuxSnapshotRetains(snapshot, &stat) {
				return true, nil
			}
		}
	}
	return false, nil
}

func linuxSnapshotRetains(snapshot executableSnapshot, stat *unix.Stat_t) bool {
	if snapshot.privateOnly && (stat.Mode&0o077 != 0 || stat.Mode&(unix.S_ISUID|unix.S_ISGID) != 0) {
		return true
	}
	for _, info := range snapshot.images {
		image, ok := info.Sys().(*syscall.Stat_t)
		if ok && image.Dev == stat.Dev && image.Ino == stat.Ino {
			return true
		}
	}
	return false
}

func probeLinuxExecutable(parentFD int, name string, expected *unix.Stat_t) error {
	// Pin the no-follow regular inode first. Reopen only this process's own
	// descriptor through the kernel magic link, so a concurrent basename
	// replacement cannot redirect a write-open onto a device or another inode.
	readFD, err := unix.Openat(parentFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(readFD) }()
	var pinned unix.Stat_t
	if err := unix.Fstat(readFD, &pinned); err != nil {
		return err
	}
	if pinned.Mode&unix.S_IFMT != unix.S_IFREG || pinned.Dev != expected.Dev || pinned.Ino != expected.Ino {
		return fmt.Errorf("executable inode changed before probe")
	}
	fd, err := unix.Open(fmt.Sprintf("/proc/self/fd/%d", readFD), unix.O_WRONLY|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	var actual unix.Stat_t
	statErr := unix.Fstat(fd, &actual)
	closeErr := unix.Close(fd)
	if statErr != nil {
		return statErr
	}
	if closeErr != nil {
		return closeErr
	}
	if actual.Mode&unix.S_IFMT != unix.S_IFREG || actual.Dev != expected.Dev || actual.Ino != expected.Ino {
		return fmt.Errorf("executable inode changed during probe")
	}
	return nil
}
