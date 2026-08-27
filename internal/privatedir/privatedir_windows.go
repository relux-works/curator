//go:build windows

package privatedir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The Windows private shape is the one the reviewed protected-root backend in
// internal/buildcache/protection_windows.go establishes for the build cache:
// an owner-only DACL whose control word carries SE_DACL_PROTECTED, attached at
// creation time so the directory never exists with a DACL inherited from its
// parent. Validation proves the effective user owns the directory, that the
// DACL is protected from inheritance, and that every entry names the owner and
// allows access — a deny, inherited, or foreign entry is a refusal, not an
// access-check exercise. That package cannot be imported from here (it depends
// on internal/closureexec), so the shape is restated with the same semantics.

// windowsPropagationFlags are the ACE flags that describe what a child
// inherits. They grant nothing on the object itself, so they are the only
// flags an owner-only DACL may carry; INHERITED_ACE is deliberately absent.
const windowsPropagationFlags = byte(windows.CONTAINER_INHERIT_ACE | windows.OBJECT_INHERIT_ACE |
	windows.NO_PROPAGATE_INHERIT_ACE | windows.INHERIT_ONLY_ACE)

const windowsFileDeleteChild windows.ACCESS_MASK = 0x00000040

// windowsConcreteMutationRights are the specific rights that together prove
// the owner can mutate and replace the directory, for a DACL that spells
// rights concretely instead of via GENERIC_ALL.
const windowsConcreteMutationRights windows.ACCESS_MASK = windows.FILE_WRITE_DATA |
	windows.FILE_APPEND_DATA | windows.FILE_WRITE_EA | windows.FILE_WRITE_ATTRIBUTES |
	windowsFileDeleteChild | windows.WRITE_DAC | windows.WRITE_OWNER | windows.DELETE

func ownerOnlyACL() (*windows.ACL, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	return windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.SET_ACCESS,
		// Children created by ordinary os calls inherit the owner-only entry;
		// without propagation they would carry an empty DACL and be unopenable
		// even by their owner. The single entry names the owner either way.
		Inheritance: windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}, nil)
}

// makePrivate creates the directory with the owner-only protected descriptor
// supplied to the create call itself. os.Mkdir followed by re-securing is not
// equivalent: between the calls the directory exists with the parent's
// inherited DACL, and a concurrent validator would fail closed on state this
// manager had just created correctly.
func makePrivate(path string) error {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
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
	// The owner is stated explicitly. When the effective user is an
	// administrator, the default owner policy can assign new objects to the
	// Administrators group instead of the personal SID, and Validate would
	// then refuse a directory this manager had just created correctly.
	if err := descriptor.SetOwner(user.User.Sid, false); err != nil {
		return err
	}
	// Without SE_DACL_PROTECTED the supplied DACL is merged with the parent's
	// inheritable entries, which is exactly the state being avoided.
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

func makeAllPrivate(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if info, statErr := os.Lstat(abs); statErr == nil {
		if info.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: abs, Err: windows.ERROR_DIRECTORY}
	}
	if parent := filepath.Dir(abs); parent != abs {
		if err := makeAllPrivate(parent); err != nil {
			return err
		}
	}
	if err := makePrivate(abs); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return nil
}

func validatePrivate(path string) error {
	handle, err := openNoFollow(path)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return fmt.Errorf("read %s attributes: %w", path, err)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("%s is a reparse point, not a private directory", path)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		return fmt.Errorf("%s is not a directory", path)
	}

	descriptor, err := windows.GetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return fmt.Errorf("read %s owner and DACL: %w", path, err)
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		return fmt.Errorf("%s has no valid owner", path)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("resolve effective Windows user: %w", err)
	}
	if !owner.Equals(user.User.Sid) {
		return fmt.Errorf("%s owner does not match the effective user", path)
	}
	control, _, err := descriptor.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return fmt.Errorf("%s DACL is not protected from inheritance", path)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil || dacl == nil {
		return fmt.Errorf("%s has no protected DACL", path)
	}
	return validateOwnerOnlyDACL(dacl, owner, path)
}

func validateOwnerOnlyDACL(dacl *windows.ACL, owner *windows.SID, path string) error {
	var ownerRights windows.ACCESS_MASK
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil || ace == nil {
			return fmt.Errorf("read %s DACL entry: %w", path, err)
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return fmt.Errorf("%s DACL contains a deny or unsupported entry", path)
		}
		if ace.Header.AceFlags&^windowsPropagationFlags != 0 {
			return fmt.Errorf("%s DACL contains an inherited or unsupported entry", path)
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart)) // #nosec G103 -- fixed ACCESS_ALLOWED_ACE layout: the SID begins at SidStart.
		if sid == nil || !sid.IsValid() {
			return fmt.Errorf("%s DACL contains an invalid SID", path)
		}
		if !owner.Equals(sid) {
			return fmt.Errorf("%s DACL grants access to another principal", path)
		}
		if ace.Header.AceFlags&windows.INHERIT_ONLY_ACE == 0 {
			ownerRights |= ace.Mask
		}
	}
	if ownerRights&windows.GENERIC_ALL == 0 && ownerRights&windowsConcreteMutationRights != windowsConcreteMutationRights {
		return fmt.Errorf("%s owner lacks required mutation rights", path)
	}
	return nil
}

func protectPrivate(path string) error {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	acl, err := ownerOnlyACL()
	if err != nil {
		return err
	}
	// The owner is set alongside the DACL for the same reason makePrivate
	// states it: an administrator's new objects may default to the
	// Administrators group as owner. Assigning one's own SID needs no special
	// privilege on an object the caller just created and can already control.
	return windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		user.User.Sid,
		nil,
		acl,
		nil,
	)
}

func openNoFollow(path string) (windows.Handle, error) {
	if strings.TrimSpace(path) == "" {
		return 0, fmt.Errorf("empty path")
	}
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	handle, err := windows.CreateFile(
		path16,
		windows.GENERIC_READ|windows.READ_CONTROL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
			return 0, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
		}
		return 0, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return handle, nil
}
