package buildcache

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

func liveExecutablePaths() ([]string, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	var paths []string
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if entry.ProcessID == 0 || entry.ProcessID == 4 {
			continue // System Idle Process and the kernel System process
		}
		path, imageErr := windowsExecutablePath(entry.ProcessID)
		if imageErr != nil {
			return nil, fmt.Errorf("read executable of process %d: %w", entry.ProcessID, imageErr)
		}
		paths = append(paths, path)
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return nil, err
	}
	return paths, nil
}

func windowsExecutablePath(pid uint32) (string, error) {
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer func() { _ = windows.CloseHandle(process) }()
	buffer := make([]uint16, 32768)
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(process, 0, &buffer[0], &size); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buffer[:size]), nil
}
