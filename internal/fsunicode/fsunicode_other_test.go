//go:build !windows

package fsunicode

import (
	"testing"
	"unicode/utf8"
)

func TestValidPreservesLiteralReplacementCharacterOutsideWindows(t *testing.T) {
	if !Valid("path/" + string(utf8.RuneError)) {
		t.Fatal("literal replacement character was rejected on a lossless string platform")
	}
}
