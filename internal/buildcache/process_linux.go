package buildcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func liveExecutablePaths() ([]string, error) {
	return linuxExecutablePaths("/proc")
}

func linuxExecutablePaths(proc string) ([]string, error) {
	entries, err := os.ReadDir(proc)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		dir := filepath.Join(proc, entry.Name())
		path, err := os.Readlink(filepath.Join(dir, "exe"))
		if err == nil {
			paths = append(paths, strings.TrimSuffix(path, " (deleted)"))
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
				return nil, statErr
			}
			noImage, statErr := linuxProcessHasNoImage(string(stat))
			if statErr != nil {
				return nil, statErr
			}
			if noImage {
				continue
			}
		}
		return nil, fmt.Errorf("read executable of process %d: %w", pid, err)
	}
	return paths, nil
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
