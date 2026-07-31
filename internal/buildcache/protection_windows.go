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

	artifactName, err := artifactChildName(artifactRel)
	if err != nil {
		opened.close()
		return nil, err
	}
	artifact, err := openWindowsProtected(filepath.Join(entryPath, "bin", artifactName), false, true)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.artifact = artifact
	return opened, nil
}

// openProtectedEntryFrom opens the receipt, artifact directory, and artifact of
// an already-proven cache entry below that entry's own open handle.
//
// Windows answers the exchange threat one layer lower than Unix does. Every
// component from the manager home down to this entry is held open without
// FILE_SHARE_DELETE, so none of them can be renamed or deleted while the sweep
// runs and the entry pathname cannot be made to resolve elsewhere. The identity
// assertion below refuses even a hypothetical exchange rather than trusting the
// share mode alone, and the decisive listing of the entry's members still comes
// from the caller's handle, which no pathname can redirect.
func openProtectedEntryFrom(entry *protectedDir, artifactRel string) (*openedEntry, error) {
	if entry == nil || entry.dir == nil {
		return nil, untrustedf("protected cache entry handle is missing")
	}
	base := entry.dir.Name()
	if err := validateWindowsHandle(windows.Handle(entry.dir.Fd()), base, true, false); err != nil {
		return nil, err
	}
	held, err := entry.dir.Stat()
	if err != nil {
		return nil, untrustedf("stat the proven cache entry: %v", err)
	}
	current, err := os.Lstat(base)
	if err != nil || !os.SameFile(held, current) {
		return nil, untrustedf("the cache entry pathname no longer names the proven entry")
	}
	artifactName, err := artifactChildName(artifactRel)
	if err != nil {
		return nil, err
	}

	opened := &openedEntry{entryDir: entry.dir, borrowedEntryDir: true}
	receipt, err := openWindowsProtected(filepath.Join(base, ReceiptFilename), false, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.receipt = receipt
	binDir, err := openWindowsProtected(filepath.Join(base, "bin"), true, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.binDir = binDir
	artifact, err := openWindowsProtected(filepath.Join(base, "bin", artifactName), false, true)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.artifact = artifact
	return opened, nil
}

// openProtectedDir resolves dirPath below home without following a reparse
// point at any component and validates the type, owner, and protected DACL of
// each one. It is the traversal boundary maintenance revalidates before reading
// the cache root or classifying one entry, and it never creates or repairs
// anything.
func openProtectedDir(home, dirPath string) (*protectedDir, error) {
	rel, err := filepath.Rel(home, dirPath)
	if err != nil || rel == "." || rel == ".." ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, untrustedf("protected directory crosses the manager-home boundary")
	}
	root, err := openWindowsProtected(home, true, false)
	if err != nil {
		return nil, err
	}
	opened := &protectedDir{parents: []*os.File{root}}
	current := home
	parts := strings.Split(rel, string(filepath.Separator))
	for index, part := range parts {
		if part == "" || part == "." || part == ".." {
			opened.close()
			return nil, untrustedf("protected directory contains an invalid path component")
		}
		current = filepath.Join(current, part)
		dir, openErr := openWindowsProtected(current, true, false)
		if openErr != nil {
			opened.close()
			return nil, openErr
		}
		if index == len(parts)-1 {
			opened.dir = dir
		} else {
			opened.parents = append(opened.parents, dir)
		}
	}
	return opened, nil
}

// openProtectedChildFile opens one regular file inside an already-validated
// protected directory handle, without following a reparse point.
func openProtectedChildFile(dir *os.File, name string) (*os.File, error) {
	if dir == nil {
		return nil, untrustedf("protected directory handle is missing")
	}
	return openWindowsProtected(filepath.Join(dir.Name(), name), false, false)
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

// ownerOnlyACL builds the owner-only, non-inheritable ACL that
// validateWindowsDACL accepts. It is the one shape Curator ever writes.
func ownerOnlyACL() (*windows.ACL, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	return windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.SET_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}, nil)
}

func protectWindowsPath(path string) error {
	acl, err := ownerOnlyACL()
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

// createProtectedDirectory creates one directory that is already protected the
// instant it becomes observable, by passing the owner-only protected security
// descriptor to the create call rather than applying it afterwards.
//
// os.Mkdir followed by protectWindowsPath is not equivalent. Between the two
// calls the new directory exists with a DACL inherited from its parent, and a
// concurrent publisher that opens it in that window fails closed on
// "DACL is not protected from inheritance" -- a spurious untrusted-provenance
// verdict for state this manager had just created correctly. Unix has no such
// window because mkdir takes the mode at creation; this gives Windows the same
// property. It returns os.ErrExist for an existing name, like os.Mkdir.
func createProtectedDirectory(path string) error {
	acl, err := ownerOnlyACL()
	if err != nil {
		return err
	}
	descriptor, err := windows.NewSecurityDescriptor()
	if err != nil {
		return err
	}
	if err := descriptor.SetDACL(acl, true, false); err != nil {
		return err
	}
	// Without SE_DACL_PROTECTED the DACL supplied here is merged with the
	// parent's inheritable entries, which is exactly the state being avoided.
	if err := descriptor.SetControl(windows.SE_DACL_PROTECTED, windows.SE_DACL_PROTECTED); err != nil {
		return err
	}
	attributes := windows.SecurityAttributes{SecurityDescriptor: descriptor}
	attributes.Length = uint32(unsafe.Sizeof(attributes))
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	if err := windows.CreateDirectory(path16, &attributes); err != nil {
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			return &os.PathError{Op: "mkdir", Path: path, Err: os.ErrExist}
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: err}
	}
	return nil
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
			if err := createProtectedDirectory(current); err != nil && !errors.Is(err, os.ErrExist) {
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
	return createProtectedDirectory(path)
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

// syncDirHandle is a no-op on Windows: FlushFileBuffers is not defined for a
// directory handle, and the rename that retires an entry is already ordered by
// the filesystem.
func syncDirHandle(*os.File) error { return nil }

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
