//go:build unix

package envorigin

import "os"

// Process identity is independent of the injectable account-database lookup.
// These scalar reads never discover an account home or inspect credentials.
func processAccountIdentity() (uint64, uint64, error) {
	realUID, effective := os.Getuid(), os.Geteuid()
	if realUID < 0 || effective < 0 {
		return 0, 0, invalid("process account identity unavailable")
	}
	return uint64(realUID), uint64(effective), nil
}

func accountMatchesProcess(a Account, realUID, effective uint64) bool {
	return a.SID == "" && a.UID == effective && a.RealUID == realUID && a.EffectiveUID == effective
}

// LookupIdentity reads process identity only, without discovering an account home.
func LookupIdentity() (Account, error) {
	realUID, effective, err := processAccountIdentity()
	return Account{UID: effective, RealUID: realUID, EffectiveUID: effective}, err
}
