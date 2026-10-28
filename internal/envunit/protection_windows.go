//go:build windows

package envunit

import (
	"fmt"
	"os"
	"strings"

	"github.com/relux-works/curator/internal/privatedir"
	"golang.org/x/sys/windows"
)

func openProtection(path string, access uint32) (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return windows.InvalidHandle, err
	}
	return windows.CreateFile(name, access, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
}

// SDDL contains only owner/DACL metadata, never a file payload or credential.
// Go's synthesized 0666/0777 bits do not establish private Windows authority.
// Logical modes describe proven privacy; Protection pins the exact descriptor.
func metadataProtection(path string, info os.FileInfo) (uint32, string, error) {
	h, err := openProtection(path, windows.READ_CONTROL)
	if err != nil {
		return 0, "", err
	}
	defer windows.CloseHandle(h)
	descriptor, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return 0, "", err
	}
	protection := descriptor.String()
	if !ValidProtection(protection) {
		return 0, "", fmt.Errorf("entry owner/DACL metadata unavailable")
	}
	mode := uint32(0777)
	switch {
	case info.Mode()&os.ModeSymlink != 0:
	case info.IsDir():
		mode = 0755
		if privatedir.Validate(path) == nil {
			mode = 0700
		}
	case info.Mode().IsRegular():
		mode = 0644
		if privatedir.ValidateFile(path) == nil {
			mode = 0600
		}
		if info.Mode().Perm()&0200 == 0 {
			mode &^= 0200
		}
	}
	return mode, protection, nil
}
func ValidProtection(value string) bool {
	if value == "" || len(value) > 16384 || strings.ContainsRune(value, 0) {
		return false
	}
	descriptor, err := windows.SecurityDescriptorFromString(value)
	if err != nil {
		return false
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		return false
	}
	dacl, _, err := descriptor.DACL()
	return err == nil && dacl != nil && descriptor.String() == value
}

// The caller checks the exact entry/preimage before this operation. Apply only
// the recorded DACL/control, retaining owner, inode, children and file content.
// A fresh observation must equal the recorded postimage before progress saves.
func ApplyDirectoryProtection(path string, want Image) error {
	if !ValidProtection(want.Protection) {
		return invalid("invalid directory protection")
	}
	descriptor, err := windows.SecurityDescriptorFromString(want.Protection)
	if err != nil {
		return err
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || "sid:"+owner.String() != want.Owner {
		return invalid("directory protection owner differs")
	}
	h, err := openProtection(path, windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	id := fmt.Sprintf("%d:%d", info.VolumeSerialNumber, uint64(info.FileIndexHigh)<<32|uint64(info.FileIndexLow))
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || id != want.Identity {
		return invalid("directory protection entry substituted")
	}
	current, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	currentOwner, _, err := current.Owner()
	if err != nil || currentOwner == nil || !owner.Equals(currentOwner) {
		return invalid("directory protection live owner differs")
	}
	control, _, err := descriptor.Control()
	if err != nil {
		return err
	}
	flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION)
	if control&windows.SE_DACL_PROTECTED != 0 {
		flags = windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION
	}
	return windows.SetKernelObjectSecurity(h, flags, descriptor)
}
