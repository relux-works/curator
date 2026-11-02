//go:build !darwin && !linux && !windows

package envorigin

import "fmt"

// LookupAccount refuses unsupported native account transport without an environment fallback.
func LookupAccount() (Account, error) {
	return Account{}, fmt.Errorf("trusted native account transport is unsupported on this target/toolchain")
}
