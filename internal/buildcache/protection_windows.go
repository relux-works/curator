//go:build windows

package buildcache

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func protectionSupported() bool { return true }

func openProtectedEntry(home, entryPath, artifactRel string) (*openedEntry, error) {
	rel, err := filepath.Rel(home, entryPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, untrustedf("cache entry crosses the manager-home boundary")
	}
	opened := &openedEntry{}
	root, err := openWindowsProtected(home, true, false)
	if err != nil {
		return nil, err
	}
	opened.extra = append(opened.extra, root)
	current := home
	parts := strings.Split(rel, string(filepath.Separator))
	for index, part := range parts {
		if part == "" || part == "." || part == ".." {
			opened.close()
			return nil, untrustedf("cache entry contains an invalid path component")
		}
		current = filepath.Join(current, part)
		dir, openErr := openWindowsProtected(current, true, false)
		if openErr != nil {
			opened.close()
			if errors.Is(openErr, os.ErrNotExist) {
				return nil, fmt.Errorf("cache entry is incomplete: %w", openErr)
			}
			return nil, openErr
		}
		if index == len(parts)-1 {
			opened.entryDir = dir
		} else {
			opened.extra = append(opened.extra, dir)
		}
	}
	receipt, err := openWindowsProtected(filepath.Join(entryPath, ReceiptFilename), false, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.receipt = receipt
	binDir, err := openWindowsProtected(filepath.Join(entryPath, "bin"), true, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.binDir = binDir

	artifactParts := strings.Split(filepath.Clean(filepath.FromSlash(artifactRel)), string(filepath.Separator))
	if len(artifactParts) != 2 || artifactParts[0] != "bin" || artifactParts[1] == "" || artifactParts[1] == "." || artifactParts[1] == ".." {
		opened.close()
		return nil, untrustedf("artifact path is not a direct bin child")
	}
	artifact, err := openWindowsProtected(filepath.Join(entryPath, "bin", artifactParts[1]), false, true)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.artifact = artifact
	return opened, nil
}

func openWindowsProtected(path string, directory, executable bool) (*os.File, error) {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, untrustedf("encode protected path: %v", err)
	}
	flags := uint32(windows.FILE_FLAG_OPEN_REPARSE_POINT)
	if directory {
		flags |= windows.FILE_FLAG_BACKUP_SEMANTICS
	}
	handle, err := windows.CreateFile(
		path16,
		windows.GENERIC_READ|windows.READ_CONTROL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		flags,
		0,
	)
	if err != nil {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
			return nil, fmt.Errorf("cache entry is incomplete: %w", os.ErrNotExist)
		}
		return nil, untrustedf("open protected path without following reparse points: %v", err)
	}
	if err := validateWindowsHandle(handle, path, directory, executable); err != nil {
		_ = windows.CloseHandle(handle)
		return nil, err
	}
	return os.NewFile(uintptr(handle), path), nil
}

func validateWindowsHandle(handle windows.Handle, label string, directory, _ bool) error {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return untrustedf("inspect %s: %v", label, err)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return untrustedf("%s is a reparse point", label)
	}
	isDirectory := info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0
	if isDirectory != directory {
		return untrustedf("%s has the wrong file type", label)
	}
	if !directory && info.NumberOfLinks != 1 {
		return untrustedf("%s has multiple hard links", label)
	}
	if err := validateWindowsSecurity(handle, label); err != nil {
		return err
	}
	return nil
}

func validateWindowsSecurity(handle windows.Handle, label string) error {
	descriptor, err := windows.GetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return untrustedf("read %s owner and DACL: %v", label, err)
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		return untrustedf("%s has no valid owner", label)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return untrustedf("resolve effective Windows user: %v", err)
	}
	if err := validateWindowsOwner(owner, user.User.Sid, label); err != nil {
		return err
	}
	control, _, err := descriptor.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return untrustedf("%s DACL is not protected from inheritance", label)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil || dacl == nil {
		return untrustedf("%s has no protected DACL", label)
	}
	return validateWindowsDACL(dacl, owner, label)
}

func validateWindowsOwner(owner, effectiveUser *windows.SID, label string) error {
	if owner == nil || !owner.IsValid() || effectiveUser == nil || !effectiveUser.IsValid() || !owner.Equals(effectiveUser) {
		return untrustedf("%s owner does not match the effective user", label)
	}
	return nil
}

// validateWindowsDACL deliberately accepts only an unambiguous owner-only ACL.
// Curator creates this shape itself. Rejecting deny, inherited, inherit-only,
// and non-owner entries avoids treating the ACL text as an access check while
// still proving that the owner has every right needed to replace the entry.
func validateWindowsDACL(dacl *windows.ACL, owner *windows.SID, label string) error {
	if dacl == nil || owner == nil || !owner.IsValid() {
		return untrustedf("%s has no valid owner-only DACL", label)
	}
	var ownerRights windows.ACCESS_MASK
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil || ace == nil {
			return untrustedf("read %s DACL entry: %v", label, err)
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return untrustedf("%s DACL contains a deny or unsupported entry", label)
		}
		if ace.Header.AceFlags != 0 {
			return untrustedf("%s DACL contains an inherited or inheritable entry", label)
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if sid == nil || !sid.IsValid() {
			return untrustedf("%s DACL contains an invalid SID", label)
		}
		if !owner.Equals(sid) {
			return untrustedf("%s DACL grants access to another principal", label)
		}
		ownerRights |= ace.Mask
	}
	if ownerRights&windows.GENERIC_ALL == 0 && ownerRights&windowsConcreteMutationRights != windowsConcreteMutationRights {
		return untrustedf("%s owner lacks required mutation rights", label)
	}
	return nil
}

const (
	windowsFileDeleteChild        windows.ACCESS_MASK = 0x00000040
	windowsConcreteMutationRights windows.ACCESS_MASK = windows.FILE_WRITE_DATA | windows.FILE_APPEND_DATA |
		windows.FILE_WRITE_EA | windows.FILE_WRITE_ATTRIBUTES | windowsFileDeleteChild |
		windows.WRITE_DAC | windows.WRITE_OWNER | windows.DELETE
)

func protectWindowsPath(path string) error {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	access := []windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.SET_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}
	acl, err := windows.ACLFromEntries(access, nil)
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	)
}

func ensureProtectedBase(home, base string) error {
	root, err := openWindowsProtected(home, true, false)
	if err != nil {
		return err
	}
	defer root.Close()
	rel, err := filepath.Rel(home, base)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return untrustedf("cache root crosses the manager-home boundary")
	}
	current := home
	var opened []*os.File
	defer func() {
		for index := len(opened) - 1; index >= 0; index-- {
			_ = opened[index].Close()
		}
	}()
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		dir, openErr := openWindowsProtected(current, true, false)
		if openErr != nil && errors.Is(openErr, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			if err := protectWindowsPath(current); err != nil {
				return err
			}
			dir, openErr = openWindowsProtected(current, true, false)
		}
		if openErr != nil {
			return openErr
		}
		opened = append(opened, dir)
	}
	return nil
}

func makeProtectedTempDir(parent, pattern string) (string, error) {
	path, err := os.MkdirTemp(parent, pattern)
	if err != nil {
		return "", err
	}
	if err := protectWindowsPath(path); err != nil {
		_ = os.RemoveAll(path)
		return "", err
	}
	return path, nil
}

func createProtectedDir(path string) error {
	if err := os.Mkdir(path, 0o700); err != nil {
		return err
	}
	return protectWindowsPath(path)
}

func writeProtectedFile(path string, _ os.FileMode, source io.Reader) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- path is manager-derived staging
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if err := protectWindowsPath(path); err != nil {
		return err
	}
	if _, err := io.Copy(file, source); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func syncDirectory(string) error { return nil }

func renameDirectoryNoReplace(from, to string) error {
	from16, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	to16, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	return windows.MoveFile(from16, to16)
}
