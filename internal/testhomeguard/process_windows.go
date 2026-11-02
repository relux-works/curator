package testhomeguard

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Start suspended so even a pre-TestMain stalled child cannot create a writer
// before job assignment. All of its descendants inherit this private job.
func launchObserved(cmd *exec.Cmd) (func() error, func(), error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, nil, err
	}
	closeJob := func() { _ = windows.CloseHandle(job) }
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		closeJob()
		return nil, nil, err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
	if err := cmd.Start(); err != nil {
		closeJob()
		return nil, nil, err
	}
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err == nil {
		err = windows.AssignProcessToJobObject(job, process)
		_ = windows.CloseHandle(process)
	}
	if err == nil {
		err = resumeObserved(uint32(cmd.Process.Pid))
	}
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		closeJob()
		return nil, nil, err
	}
	return func() error { return windows.TerminateJobObject(job, 2) }, closeJob, nil
}

// os/exec closes CreateProcess's primary-thread handle. Locate only this owned,
// still-suspended process's thread; no other process/thread is ever resumed.
func resumeObserved(pid uint32) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()
	entry := windows.ThreadEntry32{Size: uint32(unsafe.Sizeof(windows.ThreadEntry32{}))}
	for err := windows.Thread32First(snapshot, &entry); err == nil; err = windows.Thread32Next(snapshot, &entry) {
		if entry.OwnerProcessID != pid {
			continue
		}
		thread, err := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
		if err != nil {
			return err
		}
		previous, err := windows.ResumeThread(thread)
		_ = windows.CloseHandle(thread)
		if err != nil {
			return err
		}
		if previous != 1 {
			return fmt.Errorf("unexpected owned thread suspension")
		}
		return nil
	}
	return fmt.Errorf("owned primary thread unavailable")
}
