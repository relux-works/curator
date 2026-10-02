//go:build windows

package scriptworker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/relux-works/curator/internal/godriver"
	"golang.org/x/sys/windows"
)

var hardlinkKernel = windows.NewLazySystemDLL("kernel32.dll")
var findFirstExecLink = hardlinkKernel.NewProc("FindFirstFileNameW")
var findNextExecLink = hardlinkKernel.NewProc("FindNextFileNameW")
var execFileIDInfo = hardlinkKernel.NewProc("GetFileInformationByHandleEx")

type windowsExecFileID struct {
	Volume uint64
	ID     [16]byte
}

func readWindowsExecFileID(handle windows.Handle) (windowsExecFileID, error) {
	var id windowsExecFileID
	// FileIdInfo (18) provides the full 128-bit ID, including on ReFS.
	ok, _, err := execFileIDInfo.Call(uintptr(handle), 18, uintptr(unsafe.Pointer(&id)), unsafe.Sizeof(id)) // #nosec G103 -- fixed Win32 FILE_ID_INFO output layout
	if ok == 0 {
		return id, fmt.Errorf("read executable file ID: %w", err)
	}
	return id, nil
}

func nativeWindowsExecHardlinks(file *os.File, canonical string) (windowsExecHardlinkOrigin, error) {
	var origin windowsExecHardlinkOrigin
	handle := windows.Handle(file.Fd())
	var before windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &before); err != nil {
		return origin, err
	}
	if before.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return origin, fmt.Errorf("executable is a reparse point")
	}
	id, err := readWindowsExecFileID(handle)
	if err != nil {
		return origin, err
	}
	descriptor, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return origin, err
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		return origin, fmt.Errorf("executable has no readable valid owner SID: %v", err)
	}
	origin.OwnerSID = owner.String()
	encoded, err := windows.UTF16PtrFromString(canonical)
	if err != nil {
		return origin, err
	}
	// Windows path names cannot exceed 32767 UTF-16 units. A larger buffer
	// avoids partial enumeration and ERROR_MORE_DATA retries on either API.
	const bufferSize = 32768
	buffer := make([]uint16, bufferSize)
	if err := windows.GetVolumePathName(encoded, &buffer[0], bufferSize); err != nil {
		return origin, err
	}
	volume := windows.UTF16ToString(buffer)
	size := uint32(bufferSize)
	// #nosec G103 -- Win32 UTF-16 input and bounded output buffer, live for the syscall
	search, _, callErr := findFirstExecLink.Call(uintptr(unsafe.Pointer(encoded)), 0,
		uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&buffer[0])))
	if windows.Handle(search) == windows.InvalidHandle {
		return origin, fmt.Errorf("enumerate executable hard links: %w", callErr)
	}
	defer func() { _ = windows.FindClose(windows.Handle(search)) }()
	for {
		name := windows.UTF16ToString(buffer)
		if !strings.HasPrefix(name, `\`) || strings.HasPrefix(name, `\\`) || strings.Contains(name, ":") {
			return origin, fmt.Errorf("invalid volume-relative hard-link name %q", name)
		}
		path := filepath.Join(volume, strings.TrimPrefix(name, `\`))
		physical, err := godriver.CanonicalPhysicalPath(path)
		if err != nil {
			return origin, err
		}
		alias, err := os.Open(physical) // #nosec G304 -- enumerated platform path, file ID checked below
		if err != nil {
			return origin, err
		}
		aliasID, idErr := readWindowsExecFileID(windows.Handle(alias.Fd()))
		_ = alias.Close()
		if idErr != nil || aliasID != id {
			return origin, fmt.Errorf("hard-link file ID differs from open executable: %v", idErr)
		}
		origin.Links = append(origin.Links, physical)
		if uint64(len(origin.Links)) > uint64(before.NumberOfLinks) {
			return origin, fmt.Errorf("executable hard-link count changed during enumeration")
		}
		size = bufferSize
		// #nosec G103 -- Win32 bounded UTF-16 output buffer and DWORD size
		ok, _, nextErr := findNextExecLink.Call(search, uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&buffer[0])))
		if ok == 0 {
			if nextErr != windows.ERROR_HANDLE_EOF {
				return origin, fmt.Errorf("continue executable hard-link enumeration: %w", nextErr)
			}
			break
		}
	}
	var after windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &after); err != nil {
		return origin, err
	}
	if before.NumberOfLinks != after.NumberOfLinks || uint64(len(origin.Links)) != uint64(after.NumberOfLinks) {
		return origin, fmt.Errorf("incomplete or changed executable hard-link enumeration")
	}
	origin.Count = after.NumberOfLinks
	return origin, nil
}
