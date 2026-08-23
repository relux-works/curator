// Package fsunicode validates filesystem strings before they enter portable
// protocol identities.
package fsunicode

import "unicode/utf8"

// Valid reports whether value can be trusted as the lossless Unicode spelling
// of a filesystem name on the current platform.
func Valid(value string) bool {
	return utf8.ValidString(value) && platformValid(value)
}
