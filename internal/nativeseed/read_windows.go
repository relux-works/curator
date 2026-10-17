//go:build windows

package nativeseed

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/relux-works/curator/internal/pathboundary"
	"golang.org/x/sys/windows"
)

func openSeed(path string) (*os.File, error) { return pathboundary.OpenReadNoFollow(path) }

// Windows uses the owner SID and DACL of the same open handle; pathname-based
// ACL inspection would let a replacement supply permissions for different bytes.
func checkSeedFile(file *os.File, _ os.FileInfo) error {
	current, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	descriptor, err := windows.GetSecurityInfo(windows.Handle(file.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		return fmt.Errorf("cannot prove native seed target ownership")
	}
	if !owner.Equals(current.User.Sid) {
		return fmt.Errorf("native seed target is not owned by the current user")
	}
	dacl, _, err := descriptor.DACL()
	if err != nil || dacl == nil {
		return fmt.Errorf("cannot prove native seed target mutation permissions")
	}
	const mutation = windows.GENERIC_WRITE | windows.GENERIC_ALL | windows.FILE_WRITE_DATA | windows.FILE_APPEND_DATA | windows.FILE_WRITE_EA | windows.FILE_WRITE_ATTRIBUTES | 0x40 | windows.DELETE | windows.WRITE_DAC | windows.WRITE_OWNER
	for i := uint32(0); i < uint32(dacl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, i, &ace); err != nil || ace == nil {
			return fmt.Errorf("cannot inspect native seed target DACL")
		}
		switch ace.Header.AceType {
		case windows.ACCESS_ALLOWED_ACE_TYPE:
			sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart)) // #nosec G103 -- fixed ACCESS_ALLOWED_ACE layout.
			if sid == nil || !sid.IsValid() {
				return fmt.Errorf("native seed target DACL has an invalid SID")
			}
			if !owner.Equals(sid) && ace.Mask&mutation != 0 {
				return fmt.Errorf("native seed target grants mutation rights to another identity")
			}
		case windows.ACCESS_DENIED_ACE_TYPE:
		default:
			return fmt.Errorf("native seed target DACL has an unsupported entry")
		}
	}
	return nil
}
