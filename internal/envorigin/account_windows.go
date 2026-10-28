//go:build windows

package envorigin

import (
	"fmt"
	"golang.org/x/sys/windows"
)

// LookupIdentity binds the effective thread and process token SIDs. It never
// consults USERPROFILE, HOME or the account database's display name.
func LookupIdentity() (Account, error) {
	process, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return Account{}, err
	}
	effective, err := windows.GetCurrentThreadEffectiveToken().GetTokenUser()
	if err != nil {
		return Account{}, err
	}
	sid := process.User.Sid.String()
	if sid == "" || sid != effective.User.Sid.String() {
		return Account{}, fmt.Errorf("process/effective token SID mismatch")
	}
	return Account{SID: sid}, nil
}
func processAccountIdentity() (uint64, uint64, error) {
	_, err := LookupIdentity()
	return 0, 0, err
}
func accountMatchesProcess(a Account, _, _ uint64) bool {
	identity, err := LookupIdentity()
	return err == nil && a.SID == identity.SID && a.UID == 0 && a.RealUID == 0 && a.EffectiveUID == 0
}

// LookupAccount obtains the profile directory from the same process token,
// with a second SID observation after lookup. This is independent of selectors.
func LookupAccount() (Account, error) {
	before, err := LookupIdentity()
	if err != nil {
		return Account{}, err
	}
	home, err := windows.GetCurrentProcessToken().GetUserProfileDirectory()
	if err != nil {
		return Account{}, err
	}
	after, err := LookupIdentity()
	if err != nil || before.SID != after.SID || !cleanAbsolute(home) {
		return Account{}, fmt.Errorf("account token/profile changed")
	}
	before.Home = home
	return before, nil
}
