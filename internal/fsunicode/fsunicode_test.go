package fsunicode

import "testing"

func TestValidRejectsMalformedUTF8(t *testing.T) {
	if Valid(string([]byte{0xff})) {
		t.Fatal("malformed UTF-8 was accepted as a lossless filesystem string")
	}
	if !Valid("valid/é") {
		t.Fatal("valid Unicode filesystem string was rejected")
	}
}
