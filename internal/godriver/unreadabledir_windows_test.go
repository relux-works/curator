//go:build windows

package godriver

import (
	"testing"

	"golang.org/x/sys/windows"
)

// denyDirectoryListing takes away the right to list one directory and gives it
// back when the test ends.
//
// Mode bits cannot express that here: Go's os.Chmod on Windows only toggles the
// read-only attribute, so mode 0o000 leaves a directory perfectly listable and
// the case it is meant to set up never happens at all. The right this fixture
// removes is FILE_LIST_DIRECTORY, and what removes it is a deny entry in the
// directory's own protected DACL, evaluated ahead of the grant behind it.
// Everything else the entry has to stay usable for — reading its attributes,
// and being deleted with the temporary root — is left granted, so the traversal
// fails where it reads the directory and nowhere earlier.
//
// The restore has to run before the temporary root is removed. t.Cleanup is
// last-in-first-out and t.TempDir registered its removal first, so registering
// here is what orders them.
func denyDirectoryListing(t *testing.T, path string) {
	t.Helper()
	owner := currentUserSID(t)
	// The grant propagates and the denial does not. A protected DACL stops the
	// entry inheriting anything, and that recomputes what its children inherit
	// too, so a grant that does not propagate would leave the file inside this
	// directory reachable by nobody and the temporary root unremovable. The
	// denial is the directory's own listing right and belongs to it alone.
	grant := explicitAccess(owner, windows.GENERIC_ALL, windows.GRANT_ACCESS, windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT)
	setProtectedDACL(t, path, []windows.EXPLICIT_ACCESS{
		explicitAccess(owner, windows.FILE_LIST_DIRECTORY, windows.DENY_ACCESS, windows.NO_INHERITANCE),
		grant,
	})
	t.Cleanup(func() { setProtectedDACL(t, path, []windows.EXPLICIT_ACCESS{grant}) })
}

func currentUserSID(t *testing.T) *windows.SID {
	t.Helper()
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	return user.User.Sid
}

func explicitAccess(sid *windows.SID, mask windows.ACCESS_MASK, mode windows.ACCESS_MODE, inheritance uint32) windows.EXPLICIT_ACCESS {
	return windows.EXPLICIT_ACCESS{
		AccessPermissions: mask,
		AccessMode:        mode,
		Inheritance:       inheritance,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(sid),
		},
	}
}

// setProtectedDACL replaces the entry's DACL with exactly the entries given and
// stops it inheriting any others, so what the fixture states is the whole of
// what the entry permits.
func setProtectedDACL(t *testing.T, path string, access []windows.EXPLICIT_ACCESS) {
	t.Helper()
	acl, err := windows.ACLFromEntries(access, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil,
	); err != nil {
		t.Fatal(err)
	}
}
