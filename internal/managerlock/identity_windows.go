//go:build windows

package managerlock

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func canonicalWithExistingPrefix(existing string, missing []string, label string) (string, error) {
	canonical, caseSensitive, err := canonicalExistingWindowsPath(existing, len(missing) > 0, label)
	if err != nil {
		return "", err
	}
	for index := len(missing) - 1; index >= 0; index-- {
		component := missing[index]
		if !caseSensitive {
			component = strings.ToUpper(component)
		}
		canonical = filepath.Join(canonical, component)

		// A Windows application normally creates an ordinary case-insensitive
		// child, so use that deterministic provisional identity for the rest of
		// a missing suffix. Manager.prepare canonicalizes again after creation,
		// before reserving or opening any lock, and therefore also handles a
		// filesystem whose created child actually inherits case sensitivity.
		caseSensitive = false
	}
	return filepath.Clean(canonical), nil
}

// canonicalExistingWindowsPath normalizes each existing component according
// to the lookup semantics of its containing directory. filepath.EvalSymlinks
// has already resolved links and recovered each component's on-disk spelling,
// but that spelling alone is not a stable identity in a case-insensitive
// parent. Conversely, applying one leaf flag to the complete path would merge
// distinct Foo and foo children of a case-sensitive parent.
func canonicalExistingWindowsPath(existing string, needLeafCase bool, label string) (string, bool, error) {
	volume := filepath.VolumeName(existing)
	if volume == "" {
		return "", false, fmt.Errorf("resolve case sensitivity for %s: path has no volume", label)
	}

	separator := string(filepath.Separator)
	lookup := volume + separator
	canonical := strings.ToUpper(volume) + separator
	remainder := strings.TrimLeft(existing[len(volume):], `\/`)
	components := strings.FieldsFunc(remainder, func(value rune) bool {
		return value == '\\' || value == '/'
	})

	caseSensitive, err := directoryCaseSensitive(lookup)
	if err != nil {
		return "", false, fmt.Errorf("resolve case sensitivity for %s component %q: %w", label, lookup, err)
	}
	for index, component := range components {
		identityComponent := component
		if !caseSensitive {
			identityComponent = strings.ToUpper(identityComponent)
		}
		canonical = filepath.Join(canonical, identityComponent)
		lookup = filepath.Join(lookup, component)

		if index+1 < len(components) || needLeafCase {
			caseSensitive, err = directoryCaseSensitive(lookup)
			if err != nil {
				return "", false, fmt.Errorf("resolve case sensitivity for %s component %q: %w", label, lookup, err)
			}
		}
	}
	return filepath.Clean(canonical), caseSensitive, nil
}

func directoryCaseSensitive(path string) (bool, error) {
	pathUTF16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	handle, err := windows.CreateFile(
		pathUTF16,
		windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(handle)

	var flags uint32
	err = windows.GetFileInformationByHandleEx(
		handle,
		windows.FileCaseSensitiveInfo,
		(*byte)(unsafe.Pointer(&flags)),
		uint32(unsafe.Sizeof(flags)),
	)
	if err == nil {
		return flags&windows.FILE_CS_FLAG_CASE_SENSITIVE_DIR != 0, nil
	}
	if errors.Is(err, windows.ERROR_INVALID_FUNCTION) ||
		errors.Is(err, windows.ERROR_NOT_SUPPORTED) ||
		errors.Is(err, windows.ERROR_INVALID_PARAMETER) ||
		errors.Is(err, windows.ERROR_CALL_NOT_IMPLEMENTED) {
		// Filesystems and Windows versions without per-directory case
		// sensitivity use traditional case-insensitive path lookup.
		return false, nil
	}
	return false, err
}
