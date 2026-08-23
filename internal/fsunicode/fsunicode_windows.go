//go:build windows

package fsunicode

import (
	"strings"
	"unicode/utf8"
)

func platformValid(value string) bool {
	// Windows filesystem APIs expose names as UTF-16. Go replaces an unpaired
	// surrogate with U+FFFD while converting that name to string, so accepting
	// U+FFFD here could admit a path whose original scalar spelling was lost.
	// A literal U+FFFD is indistinguishable after conversion and is therefore
	// rejected too: portable identities must fail closed rather than hash a
	// laundered spelling.
	return !strings.ContainsRune(value, utf8.RuneError)
}
