//go:build windows

package godriver

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

const platformGoName = "go.exe"

// physicalPath resolves every link in path to the directory or file it really
// names.
//
// filepath.EvalSymlinks alone is not complete on Windows. A directory junction
// (IO_REPARSE_TAG_MOUNT_POINT) is a link, but since Go 1.23 os.Lstat reports it
// as ModeIrregular -- it is a reparse-tag name surrogate, so the ModeDir branch
// is skipped, and its tag is not IO_REPARSE_TAG_SYMLINK, so it falls through to
// ModeIrregular. EvalSymlinks follows only ModeSymlink, so it returns a junction
// path untouched, and a caller that then asks "is this a real directory?" is
// told no about a perfectly ordinary directory. That is not hypothetical: the
// Go installation the GitHub Actions tool cache publishes on Windows is reached
// through exactly such a junction.
//
// GetFinalPathNameByHandle resolves every component, of every link kind, in one
// call. Anchoring on what it returns is stricter than anchoring on the link:
// the identity this driver fingerprints is the physical directory, so
// retargeting the junction afterwards cannot silently move the trusted
// toolchain.
func physicalPath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	encoded, err := windows.UTF16PtrFromString(resolved)
	if err != nil {
		return "", err
	}
	// No FILE_FLAG_OPEN_REPARSE_POINT: the reparse points are meant to be
	// followed here. FILE_FLAG_BACKUP_SEMANTICS is what lets a directory be
	// opened at all.
	handle, err := windows.CreateFile(
		encoded,
		0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return "", err
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	// FILE_NAME_NORMALIZED|VOLUME_NAME_DOS is 0: the normalized, drive-letter
	// spelling. x/sys/windows declares neither constant. A returned length
	// above the supplied size is the call asking for a larger buffer, which one
	// retry at the requested size settles.
	size := uint32(windows.MAX_PATH)
	for attempt := 0; attempt < 2; attempt++ {
		buffer := make([]uint16, size)
		length, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], size, 0)
		if err != nil {
			return "", err
		}
		if length <= size {
			return trimExtendedLengthPrefix(windows.UTF16ToString(buffer[:length])), nil
		}
		size = length
	}
	return "", fmt.Errorf("resolve physical path of %q: name did not fit twice", resolved)
}

// trimExtendedLengthPrefix turns the \\?\ form GetFinalPathNameByHandle returns
// back into the ordinary spelling the rest of the driver compares against.
func trimExtendedLengthPrefix(path string) string {
	if after, found := strings.CutPrefix(path, `\\?\UNC\`); found {
		return `\\` + after
	}
	if after, found := strings.CutPrefix(path, `\\?\`); found {
		return after
	}
	return path
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
