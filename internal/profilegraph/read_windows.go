package profilegraph

import (
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Handle-relative opens never traverse a reparse point. V1 permits internal
// links but does not require following them; this host refuses all reparse
// points rather than weaken the deny-list or containment during a race.
func (r *Reader) open(root, rel string, directory bool, links ...bool) (*os.File, error) {
	root, rel, err := r.boundary(root, rel)
	if err != nil {
		return nil, err
	}
	name, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return nil, refusal("graph_invalid_input")
	}
	parent, err := windows.CreateFile(name, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	check := func(h windows.Handle) error {
		var info windows.ByHandleFileInformation
		if err := windows.GetFileInformationByHandle(h, &info); err != nil {
			return err
		}
		if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return refusal("graph_symlink_escape")
		}
		return nil
	}
	if err := check(parent); err != nil {
		windows.CloseHandle(parent)
		return nil, err
	}
	if rel == "." {
		return os.NewFile(uintptr(parent), rel), nil
	}
	parts := strings.Split(rel, "/")
	for i, part := range parts {
		if strings.ContainsAny(part, ":") {
			windows.CloseHandle(parent)
			return nil, refusal("graph_invalid_input")
		}
		if err := r.inspect(); err != nil {
			windows.CloseHandle(parent)
			return nil, err
		}
		name, err := windows.NewNTUnicodeString(part)
		if err != nil {
			windows.CloseHandle(parent)
			return nil, refusal("graph_invalid_input")
		}
		oa := windows.OBJECT_ATTRIBUTES{RootDirectory: parent, ObjectName: name, Attributes: windows.OBJ_CASE_INSENSITIVE}
		oa.Length = uint32(unsafe.Sizeof(oa))
		options := uint32(windows.FILE_OPEN_REPARSE_POINT | windows.FILE_SYNCHRONOUS_IO_NONALERT)
		if i < len(parts)-1 || directory {
			options |= windows.FILE_DIRECTORY_FILE
		}
		var next windows.Handle
		var status windows.IO_STATUS_BLOCK
		err = windows.NtCreateFile(&next, windows.FILE_GENERIC_READ, &oa, &status, nil, 0,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN, options, 0, 0)
		windows.CloseHandle(parent)
		if err != nil {
			if code, ok := err.(windows.NTStatus); ok {
				return nil, code.Errno()
			}
			return nil, err
		}
		if err := check(next); err != nil {
			windows.CloseHandle(next)
			return nil, err
		}
		parent = next
	}
	return os.NewFile(uintptr(parent), rel), nil
}
