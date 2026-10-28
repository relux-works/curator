//go:build !unix && !windows

package envorigin

func processAccountIdentity() (uint64, uint64, error) {
	return 0, 0, invalid("trusted process/account transport is unsupported on this target")
}

func accountMatchesProcess(Account, uint64, uint64) bool { return false }

// LookupIdentity refuses targets without a trusted process identity transport.
func LookupIdentity() (Account, error) { return Account{}, invalid("unsupported process identity") }
