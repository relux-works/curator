package buildcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func nativeExecutableSnapshot() (executableSnapshot, error) {
	return linuxProcessSnapshot("/proc", os.Geteuid(), true)
}

func linuxExecutablePaths(proc string) ([]string, error) {
	snapshot, err := linuxProcessSnapshot(proc, -1, false)
	return snapshot.paths, err
}

func linuxProcessSnapshot(proc string, uid int, identities bool) (executableSnapshot, error) {
	entries, err := os.ReadDir(proc)
	if err != nil {
		return executableSnapshot{}, err
	}
	snapshot := executableSnapshot{privateOnly: uid > 0, linuxProbe: identities}
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		dir := filepath.Join(proc, entry.Name())
		if uid > 0 {
			status, statusErr := os.ReadFile(filepath.Join(dir, "status")) // #nosec G304 -- fixed proc root and validated PID; kernel UID metadata only.
			if errors.Is(statusErr, os.ErrNotExist) {
				continue
			}
			if statusErr != nil {
				if identities {
					continue
				} // Mandatory inode write probes cover unreadable images.
				return executableSnapshot{}, fmt.Errorf("read credentials of process %d: %w", pid, statusErr)
			}
			foreign, statusErr := linuxForeignUID(string(status), uid)
			if statusErr != nil {
				if identities {
					continue
				} // Mandatory inode write probes cover unreadable images.
				return executableSnapshot{}, fmt.Errorf("read credentials of process %d: %w", pid, statusErr)
			}
			if foreign {
				continue
			}
		}
		path, err := os.Readlink(filepath.Join(dir, "exe"))
		if err == nil {
			snapshot.paths = append(snapshot.paths, strings.TrimSuffix(path, " (deleted)"))
			if identities {
				image, imageErr := os.Stat(filepath.Join(dir, "exe"))
				if imageErr == nil {
					snapshot.images = append(snapshot.images, image)
				} else if !errors.Is(imageErr, os.ErrNotExist) {
					// The reported path remains an extra signal. The mandatory
					// write probe covers the unreadable inode independently.
					continue
				}
			}
			continue
		}
		if errors.Is(err, os.ErrNotExist) {
			// A missing exe can be an exited process, a zombie or a kernel
			// thread. Prove which; a live userspace process is uncertainty.
			stat, statErr := os.ReadFile(filepath.Join(dir, "stat")) // #nosec G304 -- fixed proc root and validated numeric PID; reads only kernel process metadata.
			if errors.Is(statErr, os.ErrNotExist) {
				continue
			}
			if statErr != nil {
				if identities {
					continue
				}
				return executableSnapshot{}, statErr
			}
			noImage, statErr := linuxProcessHasNoImage(string(stat))
			if statErr != nil {
				if identities {
					continue
				}
				return executableSnapshot{}, statErr
			}
			if noImage {
				continue
			}
		}

		if identities {
			continue
		} // Including non-dumpable same-user processes.
		return executableSnapshot{}, fmt.Errorf("read executable of process %d: %w", pid, err)
	}
	return snapshot, nil
}

func linuxProcessHasNoImage(stat string) (bool, error) {
	end := strings.LastIndexByte(stat, ')')
	if end < 0 {
		return false, fmt.Errorf("malformed process stat")
	}
	fields := strings.Fields(stat[end+1:])
	if len(fields) < 7 {
		return false, fmt.Errorf("truncated process stat")
	}
	flags, err := strconv.ParseUint(fields[6], 10, 64) // field 9
	if err != nil {
		return false, err
	}
	return fields[0] == "Z" || fields[0] == "X" || flags&0x00200000 != 0, nil // PF_KTHREAD
}

// Proc-directory ownership changes for non-dumpable/setuid processes. The four
// kernel credential values, not that directory's owner, decide user membership.
func linuxForeignUID(status string, uid int) (bool, error) {
	for _, line := range strings.Split(status, "\n") {
		if !strings.HasPrefix(line, "Uid:") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "Uid:"))
		if len(fields) != 4 {
			return false, fmt.Errorf("malformed process UIDs")
		}
		foreign := true
		for _, field := range fields {
			value, err := strconv.ParseUint(field, 10, 32)
			if err != nil {
				return false, fmt.Errorf("malformed process UIDs")
			}
			if int64(value) == int64(uid) {
				foreign = false
			}
		}
		return foreign, nil
	}
	return false, fmt.Errorf("missing process UIDs")
}
