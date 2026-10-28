//go:build !unix

package envorigin

func processAccountIdentity() (uint64, uint64, error) {
	return 0, 0, invalid("trusted process/account transport is unsupported on this target")
}
