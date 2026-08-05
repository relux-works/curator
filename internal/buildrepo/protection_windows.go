//go:build windows

package buildrepo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

const windowsFileAllAccess windows.ACCESS_MASK = 0x001f01ff

type windowsFileIdentity struct {
	volume uint32
	index  uint64
}

type windowsHandleState struct {
	identity   windowsFileIdentity
	attributes uint32
	size       uint64
	links      uint32
	writeTime  windows.Filetime
}

type windowsGuardedDir struct {
	name    string
	handle  windows.Handle
	initial windowsHandleState
}

type nativeProtectedDirGuard struct {
	dirs []windowsGuardedDir
}

func openNativeProtectedDirGuard(names []string, hook func(string)) (*nativeProtectedDirGuard, error) {
	guard := &nativeProtectedDirGuard{}
	for _, name := range names {
		handle, err := openWindowsProtectedPath(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES)
		if err != nil {
			guard.close()
			return nil, admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during guarded open: %v", err)
		}
		state, err := validateWindowsHandle(handle, true)
		if err != nil {
			windows.CloseHandle(handle)
			guard.close()
			return nil, admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during guarded open: %v", err)
		}
		guard.dirs = append(guard.dirs, windowsGuardedDir{name: name, handle: handle, initial: state})
	}
	if hook != nil {
		hook("directory-guard")
	}
	return guard, nil
}

func (g *nativeProtectedDirGuard) validate() error {
	for _, dir := range g.dirs {
		retained, err := validateWindowsHandle(dir.handle, true)
		if err != nil || retained.identity != dir.initial.identity {
			return admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during guarded validation")
		}
		selected, err := openWindowsProtectedPath(dir.name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES)
		if err != nil {
			return admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during guarded validation: %v", err)
		}
		selectedState, validateErr := validateWindowsHandle(selected, true)
		windows.CloseHandle(selected)
		if validateErr != nil || selectedState.identity != dir.initial.identity {
			return admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during guarded validation")
		}
	}
	return nil
}

func (g *nativeProtectedDirGuard) close() {
	for index := len(g.dirs) - 1; index >= 0; index-- {
		windows.CloseHandle(g.dirs[index].handle)
	}
	g.dirs = nil
}

// regularFileIdentity is only a shape check for operation-private snapshot
// materializations. The protected store uses handle-bound proof below.
func regularFileIdentity(_ os.FileInfo, _ bool) bool { return true }

// An os.FileInfo cannot carry owner, DACL, reparse, identity, and link-count
// evidence. Keep the legacy seam fail-closed; protected-store callers use the
// native path and retained handles below.
func protectedFileIdentity(_ os.FileInfo, _ bool) bool { return false }

func currentWindowsUserSID() (*windows.SID, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	return user.User.Sid.Copy()
}

func openWindowsProtectedPath(name string, access uint32) (windows.Handle, error) {
	path, err := windows.UTF16PtrFromString(filepath.Clean(name))
	if err != nil {
		return windows.InvalidHandle, err
	}
	return windows.CreateFile(
		path,
		access,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
}

func windowsHandleInfo(handle windows.Handle) (windowsHandleState, error) {
	typ, err := windows.GetFileType(handle)
	if err != nil {
		return windowsHandleState{}, err
	}
	if typ != windows.FILE_TYPE_DISK {
		return windowsHandleState{}, fmt.Errorf("protected object is not a disk file or directory")
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return windowsHandleState{}, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return windowsHandleState{}, fmt.Errorf("protected object is a reparse point")
	}
	return windowsHandleState{
		identity: windowsFileIdentity{
			volume: info.VolumeSerialNumber,
			index:  uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow),
		},
		attributes: info.FileAttributes,
		size:       uint64(info.FileSizeHigh)<<32 | uint64(info.FileSizeLow),
		links:      info.NumberOfLinks,
		writeTime:  info.LastWriteTime,
	}, nil
}

func validateWindowsSecurityDescriptor(sd *windows.SECURITY_DESCRIPTOR, directory bool) error {
	owner, _, err := sd.Owner()
	if err != nil || owner == nil {
		return fmt.Errorf("protected object has no owner SID")
	}
	wantOwner, err := currentWindowsUserSID()
	if err != nil {
		return err
	}
	if !owner.Equals(wantOwner) {
		return fmt.Errorf("protected object owner differs from current user")
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		return fmt.Errorf("protected object DACL is inheritable")
	}
	dacl, _, err := sd.DACL()
	if err != nil || dacl == nil || dacl.AceCount != 1 {
		return fmt.Errorf("protected object DACL is absent or not private")
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(dacl, 0, &ace); err != nil || ace == nil {
		return fmt.Errorf("cannot inspect protected object DACL")
	}
	wantFlags := uint8(0)
	if directory {
		wantFlags = windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE
	}
	aceSID := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != wantFlags || ace.Mask != windowsFileAllAccess || !aceSID.Equals(wantOwner) {
		return fmt.Errorf("protected object DACL differs from owner-only profile")
	}
	return nil
}

func validateWindowsHandle(handle windows.Handle, directory bool) (windowsHandleState, error) {
	state, err := windowsHandleInfo(handle)
	if err != nil {
		return windowsHandleState{}, err
	}
	isDir := state.attributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0
	if isDir != directory || (!directory && state.links != 1) {
		return windowsHandleState{}, fmt.Errorf("protected object type or link count invalid")
	}
	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return windowsHandleState{}, err
	}
	if err := validateWindowsSecurityDescriptor(sd, directory); err != nil {
		return windowsHandleState{}, err
	}
	return state, nil
}

func applyWindowsPrivateSecurity(handle windows.Handle, directory bool) error {
	owner, err := currentWindowsUserSID()
	if err != nil {
		return err
	}
	flags := ""
	if directory {
		flags = "OICI"
	}
	sd, err := windows.SecurityDescriptorFromString(fmt.Sprintf("O:%sD:P(A;%s;FA;;;%s)", owner.String(), flags, owner.String()))
	if err != nil {
		return err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	return windows.SetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		owner,
		nil,
		dacl,
		nil,
	)
}

func secureWindowsPath(name string, directory bool) error {
	handle, err := openWindowsProtectedPath(name, windows.READ_CONTROL|windows.WRITE_DAC|windows.WRITE_OWNER|windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	state, err := windowsHandleInfo(handle)
	if err != nil {
		return err
	}
	if (state.attributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory || (!directory && state.links != 1) {
		return fmt.Errorf("protected object type or link count invalid")
	}
	if err := applyWindowsPrivateSecurity(handle, directory); err != nil {
		return err
	}
	_, err = validateWindowsHandle(handle, directory)
	return err
}

func createWindowsPrivateDirs(name string) error {
	clean := filepath.Clean(name)
	volume := filepath.VolumeName(clean)
	current := volume + string(os.PathSeparator)
	for _, part := range splitWindowsPath(clean[len(volume):]) {
		current = filepath.Join(current, part)
		err := os.Mkdir(current, 0o700)
		if err == nil {
			if err = secureWindowsPath(current, true); err != nil {
				return err
			}
			continue
		}
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func splitWindowsPath(path string) []string {
	var parts []string
	for path != "" && path != "." && path != string(os.PathSeparator) {
		path = filepath.Clean(path)
		dir, base := filepath.Split(path)
		if base == "" {
			break
		}
		parts = append([]string{base}, parts...)
		path = filepath.Clean(dir)
	}
	return parts
}

func nativeProtectedDir(name string, create bool, hook func(string)) error {
	if create {
		if err := createWindowsPrivateDirs(name); err != nil {
			return err
		}
	}
	handle, err := openWindowsProtectedPath(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	initial, err := validateWindowsHandle(handle, true)
	if err != nil {
		return admissionError(CodeProtectedBoundaryUntrusted, "protected directory cannot be proved private: %v", err)
	}
	if hook != nil {
		hook("directory-proof")
	}
	selected, err := openWindowsProtectedPath(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		return admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during validation: %v", err)
	}
	defer windows.CloseHandle(selected)
	selectedState, err := validateWindowsHandle(selected, true)
	if err != nil || selectedState.identity != initial.identity {
		return admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during validation")
	}
	final, err := validateWindowsHandle(handle, true)
	if err != nil || final.identity != initial.identity {
		return admissionError(CodeProtectedBoundaryUntrusted, "protected directory changed during validation")
	}
	return nil
}

func readNativeProtectedFile(name string, max int64, hook func(string)) ([]byte, error) {
	handle, err := openWindowsProtectedPath(name, windows.GENERIC_READ|windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(handle), name)
	if file == nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("cannot retain protected file handle")
	}
	defer file.Close()
	initial, err := validateWindowsHandle(handle, false)
	if err != nil || initial.size > uint64(max) {
		return nil, fmt.Errorf("protected file shape invalid: %v", err)
	}
	if hook != nil {
		hook("file-proof")
	}
	data, err := io.ReadAll(io.LimitReader(file, max+1))
	if err != nil || int64(len(data)) > max {
		return nil, fmt.Errorf("protected file read invalid: %v", err)
	}
	retained, err := validateWindowsHandle(handle, false)
	if err != nil || retained.identity != initial.identity || retained.size != initial.size || retained.writeTime != initial.writeTime {
		return nil, fmt.Errorf("protected file changed during read")
	}
	if hook != nil {
		hook("file-reopen")
	}
	selected, err := openWindowsProtectedPath(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		return nil, fmt.Errorf("protected file changed during validation: %w", err)
	}
	defer windows.CloseHandle(selected)
	selectedState, err := validateWindowsHandle(selected, false)
	if err != nil || selectedState.identity != initial.identity {
		return nil, fmt.Errorf("protected file changed during validation")
	}
	retained, retainedErr := validateWindowsHandle(handle, false)
	selectedState, selectedErr := validateWindowsHandle(selected, false)
	if retainedErr != nil || selectedErr != nil || retained.identity != initial.identity || selectedState.identity != initial.identity || retained.links != 1 || selectedState.links != 1 {
		return nil, fmt.Errorf("protected file changed during final validation")
	}
	return data, nil
}

func secureProtectedTree(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("protected publication contains a reparse point")
		}
		return secureWindowsPath(path, entry.IsDir())
	})
}
