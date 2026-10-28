//go:build (!darwin && !linux) || !cgo

package envorigin

import "fmt"

// No env fallback: unsupported native account transport cannot capture authority.
func LookupAccount() (Account, error) {
	return Account{}, fmt.Errorf("trusted native account transport is unsupported on this target/toolchain")
}
