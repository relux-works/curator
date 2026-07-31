//go:build !windows

package godriver

import (
	"encoding/binary"
	"io/fs"
	"path/filepath"
	"syscall"
)

const platformGoName = "go"

// physicalPath resolves an existing path to the physical location the host
// reaches through it, so that what this package pins, joins onto, and compares
// is a location rather than a name that can be re-aimed at another one.
//
// On a POSIX host the only redirection interposed on a path component is a
// symbolic link, and filepath.EvalSymlinks resolves exactly those.
func physicalPath(path string) (string, error) { return filepath.EvalSymlinks(path) }

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

// platformEnvironmentNames lists the indispensable operating-system process
// variables the closed compiler environment may additionally carry.
func platformEnvironmentNames() []string { return nil }

func applyArtifactPermissions(path string) error { return chmod(path, 0o700) }

func artifactHasMultipleLinks(_ string, info fs.FileInfo) (bool, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink != 1, nil
}
