package install

import (
	"errors"
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

// An exclusive native handle makes an ordinary state read fail with a real
// sharing violation. This tests unreadable state without assuming that the
// hosted account cannot bypass a deny-read DACL. Cleanup restores readability.
func denyProjectLedgerRead(t *testing.T, path string) {
	t.Helper()
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := windows.CloseHandle(handle); err != nil {
			t.Error(err)
		}
	})
	if _, err := os.ReadFile(path); !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("unreadable-ledger fixture did not deny a real read: %v", err)
	}
}
