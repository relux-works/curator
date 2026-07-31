//go:build windows

package godriver

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

const platformGoName = "go.exe"

// finalPathFlags is FILE_NAME_NORMALIZED|VOLUME_NAME_DOS: the normalized
// spelling on the drive-letter volume name. x/sys/windows declares neither
// constant because both are zero, which is why the call below passes 0.
const finalPathFlags = 0

// maxFinalPathAttempts bounds the grow-and-retry loop around
// GetFinalPathNameByHandle. One growth is enough for every real path; a third
// attempt only exists so a driver that keeps reporting a larger requirement
// fails closed instead of looping.
const maxFinalPathAttempts = 3

// errFinalPathUnbounded reports a final path name that never fit a bounded
// buffer.
var errFinalPathUnbounded = errors.New("the final path name did not fit a bounded buffer")

// errFinalPathNotAbsolute reports a final path name this package cannot join
// onto or compare, such as the volume-GUID spelling a volume with no drive
// letter produces.
var errFinalPathNotAbsolute = errors.New("the final path name is not an absolute drive or UNC path")

// physicalPath resolves an existing path to the physical location the host
// reaches through it, so that what this package pins, joins onto, and compares
// is a location rather than a name that can be re-aimed at another one.
//
// filepath.EvalSymlinks is not that resolver on Windows. It follows a component
// only where os.Lstat reports fs.ModeSymlink, and since Go 1.23 that is
// IO_REPARSE_TAG_SYMLINK alone: every other name-surrogate reparse point --
// a directory junction, IO_REPARSE_TAG_MOUNT_POINT, above all -- is reported as
// fs.ModeIrregular without fs.ModeDir and is left in the path unresolved. A
// junction therefore survives EvalSymlinks unchanged, and the directory checks
// that follow are then asked about the junction rather than about the directory
// it names: the answer is "not a directory" for something that opens, lists,
// and traverses normally.
//
// GetFinalPathNameByHandle on a handle opened without
// FILE_FLAG_OPEN_REPARSE_POINT is the host's own answer. The kernel resolved
// every reparse point while opening, so the final name is where the handle
// actually landed, with junctions, symlinks, and drive substitutions already
// gone. Its result is passed through filepath.EvalSymlinks afterwards, not
// instead: the final path carries no reparse point left to follow, so that call
// only normalizes, and it guarantees the spelling compared here is produced by
// the same function that produces every other canonical path in this package.
func physicalPath(path string) (string, error) {
	final, err := finalPathName(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(final)
}

// finalPathName opens path following every reparse point and reports where the
// handle landed, in the ordinary spelling rather than the extended-length one.
func finalPathName(path string) (string, error) {
	encoded, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", &os.PathError{Op: "canonicalize", Path: path, Err: err}
	}
	// No FILE_FLAG_OPEN_REPARSE_POINT: this open is required to follow the
	// redirections whose resolution is the whole point of the call.
	// FILE_FLAG_BACKUP_SEMANTICS admits directories, and FILE_READ_ATTRIBUTES
	// is all GetFinalPathNameByHandle needs, so nothing here can read content.
	handle, err := windows.CreateFile(encoded, windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", &os.PathError{Op: "canonicalize", Path: path, Err: err}
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	size := uint32(windows.MAX_PATH)
	for attempt := 0; attempt < maxFinalPathAttempts; attempt++ {
		buffer := make([]uint16, size)
		// A length below the buffer size excludes the terminator and is the
		// result; a length at or above it includes the terminator and is the
		// requirement to retry with.
		length, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], size, finalPathFlags)
		if err != nil {
			return "", &os.PathError{Op: "canonicalize", Path: path, Err: err}
		}
		if length < size {
			return ordinaryPathSpelling(path, windows.UTF16ToString(buffer[:length]))
		}
		if length >= windows.MAX_LONG_PATH {
			return "", &os.PathError{Op: "canonicalize", Path: path, Err: errFinalPathUnbounded}
		}
		size = length + 1
	}
	return "", &os.PathError{Op: "canonicalize", Path: path, Err: errFinalPathUnbounded}
}

// ordinaryPathSpelling converts the extended-length name
// GetFinalPathNameByHandle returns back to the spelling this package joins onto
// and compares: \\?\UNC\server\share becomes \\server\share and \\?\C:\dir
// becomes C:\dir. A name that is not absolute once its prefix is gone -- the
// \\?\Volume{GUID}\ spelling of a volume with no drive letter, above all -- is
// rejected rather than returned as a path that would silently be relative.
func ordinaryPathSpelling(path, final string) (string, error) {
	ordinary := final
	if rest, isUNC := strings.CutPrefix(final, `\\?\UNC\`); isUNC {
		ordinary = `\\` + rest
	} else if rest, isExtended := strings.CutPrefix(final, `\\?\`); isExtended {
		ordinary = rest
	}
	if !filepath.IsAbs(ordinary) {
		return "", &os.PathError{Op: "canonicalize", Path: path, Err: errFinalPathNotAbsolute}
	}
	return ordinary, nil
}

func executableMode(mode fs.FileMode) bool { return mode.IsRegular() }

func nativeExecutableHeader(header []byte) bool {
	return len(header) >= 2 && header[0] == 'M' && header[1] == 'Z'
}

func indispensableEnvironment() map[string]string {
	result := make(map[string]string, 2)
	for _, key := range []string{"SYSTEMROOT", "WINDIR"} {
		if value, present := os.LookupEnv(key); present && value != "" {
			result[key] = value
		}
	}
	return result
}

// platformEnvironmentNames lists the indispensable operating-system process
// variables the closed compiler environment may additionally carry.
func platformEnvironmentNames() []string {
	return []string{"APPDATA", "LOCALAPPDATA", "USERPROFILE", "TEMP", "TMP", "SYSTEMROOT", "WINDIR"}
}

func applyArtifactPermissions(path string) error { return chmod(path, 0o700) }

func artifactHasMultipleLinks(path string, _ fs.FileInfo) (bool, error) {
	encoded, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	handle, err := windows.CreateFile(encoded, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(handle)
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return false, err
	}
	return info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.NumberOfLinks != 1, nil
}
