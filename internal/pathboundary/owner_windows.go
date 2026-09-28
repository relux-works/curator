//go:build windows

package pathboundary

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const mutationRights windows.ACCESS_MASK = windows.GENERIC_WRITE | windows.GENERIC_ALL |
	windows.FILE_WRITE_DATA | windows.FILE_APPEND_DATA | windows.FILE_WRITE_EA |
	windows.FILE_WRITE_ATTRIBUTES | windowsFileDeleteChild | windows.DELETE |
	windows.WRITE_DAC | windows.WRITE_OWNER

const windowsFileDeleteChild windows.ACCESS_MASK = 0x00000040

func protectTree(root string) error {
	owner, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("resolve effective Windows user: %w", err)
	}
	trustee := windows.TRUSTEE{
		TrusteeForm:  windows.TRUSTEE_IS_SID,
		TrusteeType:  windows.TRUSTEE_IS_USER,
		TrusteeValue: windows.TrusteeValueFromSID(owner.User.Sid),
	}
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.ACCESS_MASK(0x001F01FF), // FILE_ALL_ACCESS
		AccessMode:        windows.SET_ACCESS,
		Inheritance:       windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee:           trustee,
	}}, nil)
	if err != nil {
		return fmt.Errorf("build owner-only mutation DACL: %w", err)
	}
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		link, err := isLink(path, info)
		if err != nil {
			return err
		}
		if link {
			return fmt.Errorf("cannot protect a tree containing a symbolic link: %s", path)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("cannot protect a tree containing a non-regular component: %s", path)
		}
		if err := setPrivateMutationDACL(path, acl); err != nil {
			return err
		}
		return nil
	})
}

func setPrivateMutationDACL(path string, dacl *windows.ACL) error {
	handle, err := openNoFollowAccess(path, windows.READ_CONTROL|windows.WRITE_DAC)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	api := windows.NewLazySystemDLL("advapi32.dll")
	setSecurityInfo := api.NewProc("SetSecurityInfo")
	if err := setSecurityInfo.Find(); err != nil {
		return err
	}
	status, _, _ := setSecurityInfo.Call(
		uintptr(handle),
		uintptr(windows.SE_FILE_OBJECT),
		uintptr(windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION),
		0, 0, uintptr(unsafe.Pointer(dacl)), 0,
	)
	if status != 0 {
		return fmt.Errorf("set owner-only mutation DACL on %s: %w", path, syscall.Errno(status))
	}
	runtime.KeepAlive(dacl)
	return nil
}

func isLink(path string, info os.FileInfo) (bool, error) {
	if info.Mode()&os.ModeSymlink != 0 {
		return true, nil
	}
	handle, err := openNoFollow(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	var attributes windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &attributes); err != nil {
		return false, err
	}
	return attributes.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0, nil
}

func effectiveOwner() (OwnerIdentity, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("resolve effective Windows user: %w", err)
	}
	return OwnerIdentity("sid:" + user.User.Sid.String()), nil
}

func lookupOwner(path string, _ os.FileInfo) (OwnerIdentity, error) {
	owner, _, err := securityOwnerAndDACL(path)
	if err != nil {
		return "", err
	}
	return OwnerIdentity("sid:" + owner.String()), nil
}

func checkMutationPermissions(path string, _ os.FileInfo) error {
	owner, dacl, err := securityOwnerAndDACL(path)
	if err != nil {
		return err
	}
	if owner == nil || dacl == nil {
		return fmt.Errorf("%s has no owner-only mutation DACL", path)
	}
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil || ace == nil {
			return fmt.Errorf("read %s DACL entry: %w", path, err)
		}
		switch ace.Header.AceType {
		case windows.ACCESS_ALLOWED_ACE_TYPE:
			sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart)) // #nosec G103 -- fixed ACCESS_ALLOWED_ACE layout.
			if sid == nil || !sid.IsValid() {
				return fmt.Errorf("%s DACL contains an invalid SID", path)
			}
			if !owner.Equals(sid) && ace.Mask&mutationRights != 0 {
				return fmt.Errorf("%s DACL grants mutation rights to another identity", path)
			}
		case windows.ACCESS_DENIED_ACE_TYPE:
			// Denials cannot grant mutation rights.
		default:
			return fmt.Errorf("%s DACL contains an unsupported entry", path)
		}
	}
	return nil
}

func securityOwnerAndDACL(path string) (*windows.SID, *windows.ACL, error) {
	handle, err := openNoFollow(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	descriptor, err := windows.GetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return nil, nil, err
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		return nil, nil, fmt.Errorf("%s has no valid owner", path)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		return nil, nil, err
	}
	return owner, dacl, nil
}

func openNoFollow(path string) (windows.Handle, error) {
	return openNoFollowAccess(path, windows.READ_CONTROL)
}

func openNoFollowAccess(path string, access windows.ACCESS_MASK) (windows.Handle, error) {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	return windows.CreateFile(
		path16,
		uint32(access),
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
}
