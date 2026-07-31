//go:build windows

package godriver

import (
	"encoding/binary"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func testExecutableBytes() []byte {
	return []byte{'M', 'Z', 0, 0, 0, 0, 0, 0}
}

// reparseHeaderSize is REPARSE_DATA_BUFFER up to and including Reserved.
const reparseHeaderSize = 8

// mountPointHeaderSize is the four name offset and length fields the mount
// point form places after that header.
const mountPointHeaderSize = 8

// createDirectoryJunction makes link a directory junction naming target and
// returns link.
//
// A junction is the redirection actions/setup-go leaves on the runner's own
// GOROOT, and it is the one this package must resolve. Unlike a symbolic link
// it needs no privilege to create, so it appears on ordinary hosts, and unlike
// a symbolic link Go reports it as fs.ModeIrregular without fs.ModeDir. This
// helper only builds it; assertJunctionPlatformFact states what it must be.
func createDirectoryJunction(t *testing.T, link, target string) string {
	t.Helper()
	if err := os.MkdirAll(link, 0o755); err != nil {
		t.Fatal(err)
	}
	handle, err := openReparsePointDirectory(link, windows.GENERIC_WRITE)
	if err != nil {
		t.Fatalf("cannot open %s to write a reparse point: %v", link, err)
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	buffer := mountPointReparseData(t, target)
	var returned uint32
	if err := windows.DeviceIoControl(handle, windows.FSCTL_SET_REPARSE_POINT,
		&buffer[0], uint32(len(buffer)), nil, 0, &returned, nil); err != nil {
		t.Fatalf("cannot make %s a junction naming %s: %v", link, target, err)
	}
	return link
}

func openReparsePointDirectory(path string, access uint32) (windows.Handle, error) {
	encoded, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	return windows.CreateFile(encoded, access, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
}

// mountPointReparseData encodes REPARSE_DATA_BUFFER for
// IO_REPARSE_TAG_MOUNT_POINT the way mklink /J does: the substitute name is
// the NT spelling the kernel resolves, with the trailing separator a mount
// point carries, and the print name is the ordinary spelling tools display.
// Each name is stored NUL-terminated, and each declared length excludes that
// terminator.
func mountPointReparseData(t *testing.T, target string) []byte {
	t.Helper()
	absolute, err := filepath.Abs(target)
	if err != nil {
		t.Fatal(err)
	}
	clean := filepath.Clean(absolute)
	substitute := windows.StringToUTF16(`\??\` + clean + `\`)
	display := windows.StringToUTF16(clean)

	names := make([]byte, 0, 2*(len(substitute)+len(display)))
	for _, encoded := range [][]uint16{substitute, display} {
		for _, unit := range encoded {
			names = binary.LittleEndian.AppendUint16(names, unit)
		}
	}

	buffer := make([]byte, reparseHeaderSize+mountPointHeaderSize+len(names))
	binary.LittleEndian.PutUint32(buffer[0:], windows.IO_REPARSE_TAG_MOUNT_POINT)
	putReparseUint16(t, buffer[4:], mountPointHeaderSize+len(names)) // ReparseDataLength
	putReparseUint16(t, buffer[6:], 0)                               // Reserved
	putReparseUint16(t, buffer[8:], 0)                               // SubstituteNameOffset
	putReparseUint16(t, buffer[10:], 2*(len(substitute)-1))          // SubstituteNameLength
	putReparseUint16(t, buffer[12:], 2*len(substitute))              // PrintNameOffset
	putReparseUint16(t, buffer[14:], 2*(len(display)-1))             // PrintNameLength
	copy(buffer[reparseHeaderSize+mountPointHeaderSize:], names)
	return buffer
}

// putReparseUint16 refuses a field the reparse buffer cannot represent instead
// of truncating it into a buffer the kernel would then read as a shorter name.
func putReparseUint16(t *testing.T, destination []byte, value int) {
	t.Helper()
	if value < 0 || value > 0xFFFF {
		t.Fatalf("reparse buffer field %d does not fit a USHORT", value)
	}
	binary.LittleEndian.PutUint16(destination, uint16(value))
}

// assertJunctionPlatformFact states the host behaviour the junction cases exist
// for. A Windows or Go release that stopped reporting a junction this way would
// be named here rather than silently making the regression unobservable.
func assertJunctionPlatformFact(t *testing.T, junction string) {
	t.Helper()
	if tag := junctionReparseTag(t, junction); tag != windows.IO_REPARSE_TAG_MOUNT_POINT {
		t.Fatalf("reparse tag of %s = %#08x, want IO_REPARSE_TAG_MOUNT_POINT", junction, tag)
	}
	info, err := os.Lstat(junction)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() || info.Mode()&fs.ModeIrregular == 0 || info.Mode()&fs.ModeSymlink != 0 {
		t.Fatalf("os.Lstat(%s).Mode() = %v, want an irregular non-directory that is not a symlink", junction, info.Mode())
	}
	followed, err := os.Stat(junction)
	if err != nil || !followed.IsDir() {
		t.Fatalf("os.Stat(%s) = %v, %v, want a directory", junction, followed, err)
	}
	if resolved, err := filepath.EvalSymlinks(junction); err != nil || resolved != junction {
		t.Fatalf("filepath.EvalSymlinks(%s) = %q, %v, want the junction returned unresolved", junction, resolved, err)
	}
}

// junctionReparseTag reports the raw reparse tag of path, so a case can prove
// it built the same redirection the runner's tool cache carries rather than a
// different one that merely fails the same way.
func junctionReparseTag(t *testing.T, path string) uint32 {
	t.Helper()
	handle, err := openReparsePointDirectory(path, windows.FILE_READ_ATTRIBUTES)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	// FILE_ATTRIBUTE_TAG_INFO, which x/sys/windows names the class constant for
	// but declares no struct for.
	var info struct {
		FileAttributes uint32
		ReparseTag     uint32
	}
	if err := windows.GetFileInformationByHandleEx(handle, windows.FileAttributeTagInfo,
		(*byte)(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		t.Fatal(err)
	}
	return info.ReparseTag
}
