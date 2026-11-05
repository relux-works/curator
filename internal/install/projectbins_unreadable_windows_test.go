package install

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

// A mandatory byte-range lock denies a Go ReadFile on these nonempty records,
// independently of backup-intent open privileges. Cleanup restores readability.
// LockFileEx denies access through a second handle, including in this process:
// https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-lockfileex
func denyProjectLedgerRead(t *testing.T, path string) {
	t.Helper()
	if data, err := os.ReadFile(path); err != nil || len(data) == 0 {
		t.Fatalf("readable nonempty fixture control: size=%d error=%v", len(data), err)
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	var offset windows.Overlapped
	if err := windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &offset); err != nil {
		_ = windows.CloseHandle(handle)
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := windows.UnlockFileEx(handle, 0, 1, 0, &offset); err != nil {
			t.Error(err)
		}
		if err := windows.CloseHandle(handle); err != nil {
			t.Error(err)
		}
	})
	if _, err := os.ReadFile(path); !errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		t.Fatalf("unreadable-state fixture did not deny a real read: %v", err)
	}
	t.Log("exclusive byte-range lock: Go read refused with ERROR_LOCK_VIOLATION")
}

// An ordinary native second open respects an exclusive share denial. Record
// the backup-intent comparison separately: hosted Windows permits that open
// against the same file ID, so share denial alone cannot prove Go read denial.
func TestProjectBinWindowsShareDenialBackupIntent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "probe")
	if err := os.WriteFile(path, []byte("probe"), 0600); err != nil {
		t.Fatal(err)
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	second, secondErr := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if secondErr == nil {
		_ = windows.CloseHandle(second)
	}
	if !errors.Is(secondErr, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("ordinary native open did not respect share denial: %v", secondErr)
	}
	backup, backupErr := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if backupErr == nil {
		defer func() { _ = windows.CloseHandle(backup) }()
		var held, readInfo windows.ByHandleFileInformation
		if err := windows.GetFileInformationByHandle(handle, &held); err != nil {
			t.Fatal(err)
		}
		if err := windows.GetFileInformationByHandle(backup, &readInfo); err != nil {
			t.Fatal(err)
		}
		if held.VolumeSerialNumber != readInfo.VolumeSerialNumber || held.FileIndexHigh != readInfo.FileIndexHigh || held.FileIndexLow != readInfo.FileIndexLow {
			t.Fatal("backup-intent open reached a different file identity")
		}
		t.Log("ordinary open refused; backup-intent open permitted against the same file identity")
	} else if errors.Is(backupErr, windows.ERROR_SHARING_VIOLATION) {
		t.Log("ordinary and backup-intent opens both refused by this account's share boundary")
	} else {
		t.Fatalf("unexpected backup-intent outcome: %v", backupErr)
	}
}
