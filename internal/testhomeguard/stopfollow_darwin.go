//go:build darwin

package testhomeguard

import "golang.org/x/sys/unix"

// darwinSSTOP matches SSTOP in XNU <sys/proc.h> (SIDL=1, SRUN=2, SSLEEP=3, SSTOP=4, SZOMB=5).
const darwinSSTOP = 4

func childStopped(pid int) bool {
	process, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	return err == nil && process.Proc.P_stat == darwinSSTOP
}
