//go:build windows

package nativeseed

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
	"golang.org/x/sys/windows"
)

// A constructed current-user, owner-only DACL admits bytes; an explicit foreign
// mutation grant refuses with no bytes. Restoring the valid fixture admits again.
func TestNativeSeedWindowsDACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seed")
	payload := []byte("fixture seed\n")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := privatedir.ProtectFile(path); err != nil {
		t.Fatal(err)
	}
	checkValid := func() {
		t.Helper()
		got, err := ReadFile(path)
		if err != nil || got.Kind != stateread.KindPresent || !bytes.Equal(got.Bytes, payload) {
			t.Fatalf("private native fixture: %+v %v", got, err)
		}
	}
	checkValid()
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.READ_CONTROL|windows.WRITE_DAC, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	everyone, err := windows.StringToSid("S-1-1-0")
	if err != nil {
		t.Fatal(err)
	}
	unsafeACL, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{AccessPermissions: windows.GENERIC_WRITE, AccessMode: windows.GRANT_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeType: windows.TRUSTEE_IS_WELL_KNOWN_GROUP, TrusteeValue: windows.TrusteeValueFromSID(everyone)}}}, acl)
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, unsafeACL, nil); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err == nil || !strings.Contains(err.Error(), "grants mutation rights to another identity") || got.Kind != stateread.KindUnreadable || len(got.Bytes) != 0 {
		t.Fatalf("foreign mutation grant admitted: %+v %v", got, err)
	}
	if err := privatedir.ProtectFile(path); err != nil {
		t.Fatal(err)
	}
	checkValid()
	t.Run("relative-link", func(t *testing.T) {
		link := filepath.Join(filepath.Dir(path), "link")
		if err := os.Symlink("seed", link); err != nil {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		got, err := ReadFile(link)
		if err != nil || !bytes.Equal(got.Bytes, payload) {
			t.Fatalf("private linked seed: %+v %v", got, err)
		}
	})
}
