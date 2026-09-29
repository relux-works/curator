//go:build windows

package envprofile

import (
	"fmt"
	"runtime"

	"github.com/relux-works/curator/internal/pathboundary"
	"golang.org/x/sys/windows"
)

func makeFIFOVectorForTest(string) error {
	return fmt.Errorf("Windows does not expose POSIX named pipes")
}

func makeWorldWritableDirectoryForTest(path string) (func() error, error) {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(
		path16,
		uint32(windows.READ_CONTROL|windows.WRITE_DAC),
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	descriptor, err := windows.GetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return nil, err
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		return nil, err
	}
	everyone, err := windows.StringToSid("S-1-1-0")
	if err != nil {
		return nil, err
	}
	worldWrite, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_WRITE,
		AccessMode:        windows.GRANT_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
			TrusteeValue: windows.TrusteeValueFromSID(everyone),
		},
	}}, dacl)
	if err != nil {
		return nil, err
	}
	if err := windows.SetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, worldWrite, nil); err != nil {
		return nil, err
	}
	return func() error { return pathboundary.ProtectTree(path) }, nil
}

// makeWorldWritableSingleDirectoryForTest is makeWorldWritableDirectoryForTest
// for directories that hold manager-created links: restore puts back the
// original DACL of that one path instead of protecting the tree.
func makeWorldWritableSingleDirectoryForTest(path string) (func() error, error) {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(
		path16,
		uint32(windows.READ_CONTROL|windows.WRITE_DAC),
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	descriptor, err := windows.GetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return nil, err
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		return nil, err
	}
	everyone, err := windows.StringToSid("S-1-1-0")
	if err != nil {
		return nil, err
	}
	worldWrite, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_WRITE,
		AccessMode:        windows.GRANT_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
			TrusteeValue: windows.TrusteeValueFromSID(everyone),
		},
	}}, dacl)
	if err != nil {
		return nil, err
	}
	if err := windows.SetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, worldWrite, nil); err != nil {
		return nil, err
	}
	// Restore exactly the original DACL of this one path: the checked
	// directory may contain manager-created links that a tree walk must not
	// follow or refuse.
	return func() error {
		restoreHandle, err := windows.CreateFile(
			path16,
			uint32(windows.READ_CONTROL|windows.WRITE_DAC),
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,
			0,
		)
		if err != nil {
			return err
		}
		defer func() { _ = windows.CloseHandle(restoreHandle) }()
		err = windows.SetSecurityInfo(restoreHandle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
		runtime.KeepAlive(descriptor)
		return err
	}, nil
}
