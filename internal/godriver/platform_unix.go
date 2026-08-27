//go:build !windows

package godriver

import (
	"encoding/binary"
	"io/fs"
)

const platformGoName = "go"

func executableMode(mode fs.FileMode) bool { return mode.Perm()&0o111 != 0 }

func nativeExecutableHeader(header []byte) bool {
	if len(header) >= 4 && string(header[:4]) == "\x7fELF" {
		return true
	}
	if len(header) < 4 {
		return false
	}
	magic := binary.BigEndian.Uint32(header[:4])
	switch magic {
	case 0xfeedface, 0xfeedfacf, 0xcefaedfe, 0xcffaedfe, 0xcafebabe, 0xbebafeca, 0xcafebabf, 0xbfbafeca:
		return true
	default:
		return false
	}
}

func indispensableEnvironment() map[string]string { return nil }
