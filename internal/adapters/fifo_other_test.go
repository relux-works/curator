//go:build !unix

package adapters

import "errors"

// makeFIFO reports that this platform has no cheap special-file primitive; the
// caller skips instead of asserting.
func makeFIFO(string) error {
	return errors.New("named pipes are not available on this platform")
}
