//go:build windows

package privatedir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

// Claim: the Windows file/directory validators accept the shared private
// creators and refuse the same objects after an explicit foreign write grant.
func TestWindowsPrivateCreationRefusesBroadenedACL(t *testing.T) {
	for _, directory := range []bool{false, true} {
		t.Run(map[bool]string{false: "file", true: "directory"}[directory], func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "owned")
			validate := ValidateFile
			if directory {
				if err := Make(path); err != nil {
					t.Fatal(err)
				}
				validate = Validate
			} else {
				file, err := Create(path, 0o600)
				if err != nil {
					t.Fatal(err)
				}
				if err := file.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if err := validate(path); err != nil {
				t.Fatalf("private positive control: %v", err)
			}
			user, err := windows.GetCurrentProcessToken().GetTokenUser()
			if err != nil {
				t.Fatal(err)
			}
			world, err := windows.CreateWellKnownSid(windows.WinWorldSid)
			if err != nil {
				t.Fatal(err)
			}
			acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{
				{AccessPermissions: windows.GENERIC_ALL, AccessMode: windows.SET_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeType: windows.TRUSTEE_IS_USER, TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid)}},
				{AccessPermissions: windows.GENERIC_WRITE, AccessMode: windows.SET_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeValue: windows.TrusteeValueFromSID(world)}},
			}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
				t.Fatal(err)
			}
			if err := validate(path); err == nil || !strings.Contains(err.Error(), "another principal") {
				t.Fatalf("broadened ACL was not refused for its foreign grant: %v", err)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("refusal removed the owned object: %v", err)
			}
		})
	}
}
