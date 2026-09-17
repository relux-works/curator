//go:build windows

package staging

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// inspectedIdentity is the result of one filesystem inspection behind an
// identity capture: the volume serial plus file index read from one
// opened object, the triple os.SameFile compares on Windows.
type inspectedIdentity struct {
	volume uint32
	high   uint32
	low    uint32
}

// inspectPath performs the single filesystem inspection behind an identity
// capture: it opens the object once and reads its identity from that
// handle, so the in-memory pin and the durable token can never name
// different objects. There is deliberately no second pathname-based
// read here: FileInfo carries no public file index on Windows, and
// os.SameFile resolves it lazily from the stored path, so deriving the
// two proof forms from two independent opens would admit a swap between
// them (see captureIdentity). The handle opens with
// FILE_FLAG_BACKUP_SEMANTICS so directories open, without
// FILE_FLAG_OPEN_REPARSE_POINT so links are followed, matching Stat. A
// missing path returns an IsNotExist error (a vanished boundary, proven
// overlap); any other failure returns a plain error the caller reports as
// boundary_identity_unreadable, never as a proven overlap.
func inspectPath(canonical string) (inspectedIdentity, error) {
	path16, err := windows.UTF16PtrFromString(canonical)
	if err != nil {
		return inspectedIdentity{}, &os.PathError{Op: "stat", Path: canonical, Err: err}
	}
	handle, err := windows.CreateFile(
		path16,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
			return inspectedIdentity{}, &os.PathError{Op: "stat", Path: canonical, Err: os.ErrNotExist}
		}
		return inspectedIdentity{}, &os.PathError{Op: "stat", Path: canonical, Err: err}
	}
	defer windows.CloseHandle(handle)
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return inspectedIdentity{}, &os.PathError{Op: "stat", Path: canonical, Err: err}
	}
	return inspectedIdentity{volume: info.VolumeSerialNumber, high: info.FileIndexHigh, low: info.FileIndexLow}, nil
}

// tokenOfInspection serializes the identity of an already-inspected object
// without touching the filesystem. It must never open the path again: a
// second pathname-based read is the capture window a swap between the two
// reads would corrupt (see captureIdentity).
func tokenOfInspection(inspected inspectedIdentity, _ string) (string, error) {
	return fmt.Sprintf("%x:%x:%x", inspected.volume, inspected.high, inspected.low), nil
}

// FileIdentity returns the durable filesystem identity token for the object
// at canonical (a resolved absolute path): volume serial plus file index,
// the triple os.SameFile compares on Windows. Both the same-process
// Recheck and restart recovery verify the live token against the
// planning-time pin (see verifyDurableIdentity). A missing path returns
// an IsNotExist error (a vanished boundary, proven overlap); any other
// failure returns a plain error the caller reports as
// boundary_identity_unreadable, never as a proven overlap.
func FileIdentity(canonical string) (string, error) {
	inspected, err := inspectPath(canonical)
	if err != nil {
		return "", err
	}
	return tokenOfInspection(inspected, canonical)
}
