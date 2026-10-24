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

// Claim: Prepare adopts the owned default-inheriting manager boundary but
// refuses an explicit foreign grant in either a protected or unprotected ACL.
// Refusal retains that descriptor instead of silently normalizing it.
func TestWindowsPrepareOwnedInheritanceAndForeignGrant(t *testing.T) {
	positive := t.TempDir()
	if err := Prepare(positive); err != nil {
		t.Fatalf("owned inherited boundary refused: %v", err)
	}
	if err := Validate(positive); err != nil {
		t.Fatalf("adopted manager boundary remains unsafe: %v", err)
	}
	for _, protected := range []bool{false, true} {
		t.Run(map[bool]string{false: "unprotected", true: "protected"}[protected], func(t *testing.T) {
			path := t.TempDir()
			if err := Protect(path); err != nil {
				t.Fatal(err)
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
				{AccessPermissions: windows.GENERIC_ALL, AccessMode: windows.SET_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid)}},
				{AccessPermissions: windows.GENERIC_WRITE, AccessMode: windows.SET_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeValue: windows.TrusteeValueFromSID(world)}},
			}, nil)
			if err != nil {
				t.Fatal(err)
			}
			flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION)
			if protected {
				flags = windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION
			}
			if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, flags, nil, nil, acl, nil); err != nil {
				t.Fatal(err)
			}
			before, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
			if err != nil {
				t.Fatal(err)
			}
			want := "explicit foreign grant"
			if protected {
				want = "another principal"
			}
			if err := Prepare(path); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("explicit foreign manager grant was normalized or got wrong refusal: %v", err)
			}
			after, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
			if err != nil {
				t.Fatal(err)
			}
			if before.String() != after.String() {
				t.Fatal("refusal changed the foreign descriptor")
			}
		})
	}
}

// Claim: Windows directory adoption accepts only the effective identity or
// its creation-default owner, and rejects a distinct valid owner and no owner.
func TestWindowsPrivateDirectoryOwnerScope(t *testing.T) {
	effective, err := windows.CreateWellKnownSid(windows.WinLocalServiceSid)
	if err != nil {
		t.Fatal(err)
	}
	creationDefault, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	if err != nil {
		t.Fatal(err)
	}
	if !privateDirectoryOwnerAllowed(effective, effective, creationDefault) || !privateDirectoryOwnerAllowed(creationDefault, effective, creationDefault) {
		t.Fatal("valid directory owner refused")
	}
	if privateDirectoryOwnerAllowed(foreign, effective, creationDefault) || privateDirectoryOwnerAllowed(nil, effective, creationDefault) {
		t.Fatal("foreign directory owner admitted")
	}
}

// Claim: an already-protected descriptor lacking file mutation rights is
// refused without being upgraded; permissions to inspect/rewrite its ACL are
// deliberately present so the protected-descriptor refusal is exercised.
func TestWindowsPreparePreservesProtectedInsufficientRights(t *testing.T) {
	path := t.TempDir()
	if err := Protect(path); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{AccessPermissions: windows.GENERIC_READ | windows.READ_CONTROL | windows.WRITE_DAC | windows.WRITE_OWNER, AccessMode: windows.SET_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid)}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := Protect(path); err != nil {
			t.Error(err)
		}
	})
	if err := Prepare(path); err == nil {
		t.Fatal("protected insufficient rights were normalized")
	}
	if err := Validate(path); err == nil || !strings.Contains(err.Error(), "owner lacks required mutation rights") {
		t.Fatalf("protected descriptor refusal changed its rights: %v", err)
	}
}
