//go:build darwin

package testhomeguard

import "golang.org/x/sys/unix"

func childStopped(pid int) bool {
	process, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	return err == nil && process.Proc.P_stat == 3 // SSTOP
}
