//go:build windows

package fsunicode

import (
	"testing"
	"unicode/utf8"
)

func TestValidRejectsWindowsReplacementCharacter(t *testing.T) {
	if Valid("path/" + string(utf8.RuneError)) {
		t.Fatal("Windows replacement character was accepted as a lossless filesystem string")
	}
	if !Valid("path/🚀") {
		t.Fatal("valid Windows surrogate pair was rejected")
	}
}
