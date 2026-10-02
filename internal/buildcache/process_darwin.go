package buildcache

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func liveExecutablePaths() ([]string, error) {
	processes, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, process := range processes {
		pid := int(process.Proc.P_pid)
		// PID zero is the kernel; zombies no longer execute an image.
		if pid == 0 || process.Proc.P_stat == 5 { // SZOMB
			continue
		}
		path, err := darwinExecutablePath(pid)
		if errors.Is(err, unix.ESRCH) {
			continue // process exited during enumeration
		}
		if err != nil {
			// procargs2 can return EINVAL after the original snapshot. Prove
			// exit or zombie state before treating that read as absence.
			if errors.Is(unix.Kill(pid, 0), unix.ESRCH) {
				continue
			}
			current, stateErr := unix.SysctlKinfoProc("kern.proc.pid", pid)
			if stateErr == nil && current.Proc.P_stat == 5 {
				continue
			}
			return nil, fmt.Errorf("read executable of process %d: %w", pid, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func darwinExecutablePath(pid int) (string, error) {
	args, err := unix.SysctlRaw("kern.procargs2", pid)
	if err != nil {
		return "", err
	}
	return executableFromDarwinArgs(args)
}

func executableFromDarwinArgs(args []byte) (string, error) {
	// KERN_PROCARGS2 starts with a 32-bit argc and the saved exec_path,
	// followed by argv and environment. Read only exec_path, never argv[0].
	if len(args) <= 4 {
		return "", fmt.Errorf("truncated process arguments")
	}
	end := bytes.IndexByte(args[4:], 0)
	if end < 0 || !filepath.IsAbs(string(args[4:4+end])) {
		return "", fmt.Errorf("missing or invalid saved executable path")
	}
	return string(args[4 : 4+end]), nil
}
