//go:build windows

package godriver

import (
	"testing"

	"golang.org/x/sys/windows"
)

// denyDirectoryListing takes away the right to list one directory and gives it
// back when the test ends.
//
// Mode bits cannot express that here: Go's os.Chmod on Windows only toggles the
// read-only attribute, so mode 0o000 leaves a directory perfectly listable.
// A DACL deny is not a sound substitute for this test either. Go opens
// directories with backup intent, and a process with SeBackupPrivilege (such as
// the hosted Windows runner) can therefore bypass the deny.
//
// An already-open handle with no sharing is the Windows-native boundary that
// remains: any later read-open of this directory fails with a sharing violation,
// independently of the caller's ACL privileges. The handle asks for read access
// itself so Windows includes it in the share check performed for every later
// Open/OpenRoot call.
//
// The restore has to run before the temporary root is removed. t.Cleanup is
// last-in-first-out and t.TempDir registered its removal first, so registering
// here is what orders them.
func denyDirectoryListing(t *testing.T, path string) {
	t.Helper()
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(
		name,
		windows.GENERIC_READ,
		0,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := windows.CloseHandle(handle); err != nil {
			t.Errorf("close unreadable-directory fixture: %v", err)
		}
	})
}
