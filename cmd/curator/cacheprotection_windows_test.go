//go:build windows

package main

import (
	"testing"

	"golang.org/x/sys/windows"
)

// restoreCacheProtection puts back the protected boundary of one restored cache
// node.
//
// The snapshot restore recreates the tree with ordinary os calls, which is
// enough on unix because the mode bits it also restores are the whole boundary
// there. On Windows the boundary is the DACL, an ordinary directory inherits
// its parent's, and a restored entry therefore comes back outside the protected
// boundary — so every case after the first one in a shared fixture reported
// untrusted provenance instead of the state it was written to pin.
func restoreCacheProtection(t *testing.T, path string, directory bool) {
	t.Helper()
	inheritance := uint32(windows.NO_INHERITANCE)
	if directory {
		inheritance = windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT
	}
	setWindowsDACL(t, path, []windows.EXPLICIT_ACCESS{
		testAccessEntry(t, currentUserSID(t), windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.SET_ACCESS, inheritance),
	})
}

// breakCacheProtection is the Windows spelling of chmod 0o777 on a cache entry:
// it grants a principal other than the owner the right to mutate it, which is
// exactly what makes the protected boundary unprovable. Mode bits cannot
// express that here — Go's os.Chmod only toggles the read-only attribute on
// Windows, so the unix fixture left the boundary perfectly provable.
func breakCacheProtection(t *testing.T, path string) {
	t.Helper()
	owner := currentUserSID(t)
	world, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	if err != nil {
		t.Fatal(err)
	}
	setWindowsDACL(t, path, []windows.EXPLICIT_ACCESS{
		testAccessEntry(t, owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.SET_ACCESS, windows.NO_INHERITANCE),
		testAccessEntry(t, world, windows.TRUSTEE_IS_WELL_KNOWN_GROUP, windows.GENERIC_WRITE, windows.GRANT_ACCESS, windows.NO_INHERITANCE),
	})
}

func currentUserSID(t *testing.T) *windows.SID {
	t.Helper()
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	return user.User.Sid
}

func testAccessEntry(t *testing.T, sid *windows.SID, trusteeType windows.TRUSTEE_TYPE,
	mask windows.ACCESS_MASK, mode windows.ACCESS_MODE, inheritance uint32) windows.EXPLICIT_ACCESS {
	t.Helper()
	return windows.EXPLICIT_ACCESS{
		AccessPermissions: mask,
		AccessMode:        mode,
		Inheritance:       inheritance,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  trusteeType,
			TrusteeValue: windows.TrusteeValueFromSID(sid),
		},
	}
}

func setWindowsDACL(t *testing.T, path string, access []windows.EXPLICIT_ACCESS) {
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

// foreignArtifactPath names an artifact this target does not derive, so a
// marker recording it is genuine target drift. A windows target derives
// bin/build-tool.exe, so the foreign spelling here is the unix one.
func foreignArtifactPath() string { return "bin/build-tool" }
