//go:build windows

package godriver

import (
	"io/fs"
	"os"

	"golang.org/x/sys/windows"
)

const platformGoName = "go.exe"

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
