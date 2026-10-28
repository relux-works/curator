//go:build !unix && !windows

package linkcap

// Unavailable classifies platform errors that permit a declared projection-copy alternative.
func Unavailable(error) bool { return false }
