//go:build windows

package scopes

import (
	"testing"

	"golang.org/x/sys/windows"
)

// protectedTestHome builds a manager home the protected build cache accepts as
// its trust anchor: owned by the effective user, with an inheritance-protected
// DACL that grants that user and nobody else.
//
// Building it here rather than borrowing the store's own helper is what lets
// the end-to-end maintenance tests actually run on Windows instead of skipping:
// the sweep they exercise is refused outright on state it cannot prove.
func protectedTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := protectWindowsTestPath(home); err != nil {
		t.Skipf("this host cannot create manager-protected state: %v", err)
	}
	return home
}

func protectWindowsTestPath(path string) error {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	access := []windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.SET_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}
	acl, err := windows.ACLFromEntries(access, nil)
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	)
}
