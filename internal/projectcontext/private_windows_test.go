//go:build windows

package projectcontext

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
	"golang.org/x/sys/windows"
)

// Claim: lower-level registration record creation keeps its protected ACL
// through replacement. Directory FlushFileBuffers may remain unavailable;
// this establishes file privacy, not native Windows registration qualification.
func TestWindowsPrivateRegistrationRecord(t *testing.T) {
	root := filepath.Join(t.TempDir(), "records")
	if err := privatedir.Make(root); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{"first", "replacement"} {
		err := writeAtomic(root, "state.json", []byte(payload))
		if err != nil && !errors.Is(err, windows.ERROR_ACCESS_DENIED) && !errors.Is(err, windows.ERROR_INVALID_HANDLE) {
			t.Fatal(err)
		}
		if err != nil {
			t.Log("directory durability unavailable; checking the already-published record ACL only")
		}
		path := filepath.Join(root, "state.json")
		if err := privatedir.ValidateFile(path); err != nil {
			t.Fatalf("registration record lost private DACL: %v", err)
		}
		if raw, err := os.ReadFile(path); err != nil || string(raw) != payload {
			t.Fatalf("record bytes: %q %v", raw, err)
		}
	}
	broadenRegistrationStage(t, root)
	err := writeAtomic(root, "state.json", []byte("forbidden"))
	if err == nil || !strings.Contains(err.Error(), "mutation rights to another identity") {
		t.Fatalf("broadened record root got wrong refusal: %v", err)
	}
	if raw, err := os.ReadFile(filepath.Join(root, "state.json")); err != nil || string(raw) != "replacement" {
		t.Fatalf("refusal replaced private record: %q %v", raw, err)
	}
	staged, err := filepath.Glob(filepath.Join(root, ".write-*"))
	if err != nil || len(staged) != 0 {
		t.Fatalf("refusal left staged records: %v %v", staged, err)
	}
}

// Claim: the real snapshot staging writer protects its root, nested paths,
// manifest and payload before publication; a foreign-grant root refuses with
// no writes. Native CIP identity/admission on Windows remains unsupported.
func TestWindowsPrivateRegistrationSnapshotStage(t *testing.T) {
	c := fromFiles(map[string][]byte{"nested/guide.md": []byte("retained")})
	stage := t.TempDir()
	if err := writeSnapshotStage(stage, c); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{stage, filepath.Join(stage, "files"), filepath.Join(stage, "files", "nested")} {
		if err := privatedir.Validate(dir); err != nil {
			t.Fatalf("registration stage directory lost private DACL: %v", err)
		}
	}
	for _, rel := range []string{"manifest.json", "files/nested/guide.md"} {
		if err := privatedir.ValidateFile(filepath.Join(stage, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("registration stage file lost private DACL: %v", err)
		}
	}
	if raw, err := os.ReadFile(filepath.Join(stage, "files", "nested", "guide.md")); err != nil || string(raw) != "retained" {
		t.Fatalf("staging lost payload: %q %v", raw, err)
	}
	refused := t.TempDir()
	if err := privatedir.Protect(refused); err != nil {
		t.Fatal(err)
	}
	broadenRegistrationStage(t, refused)
	err := writeSnapshotStage(refused, c)
	if err == nil || !strings.Contains(err.Error(), "another principal") {
		t.Fatalf("foreign snapshot staging root got wrong refusal: %v", err)
	}
	entries, err := os.ReadDir(refused)
	if err != nil || len(entries) != 0 {
		t.Fatalf("refusal wrote private staging: %v %v", entries, err)
	}
	if err := privatedir.Validate(refused); err == nil || !strings.Contains(err.Error(), "another principal") {
		t.Fatalf("refusal erased the foreign staging grant: %v", err)
	}
}

func broadenRegistrationStage(t *testing.T, path string) {
	t.Helper()
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	world, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	if err != nil {
		t.Fatal(err)
	}
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{
		{AccessPermissions: windows.GENERIC_ALL, AccessMode: windows.SET_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid)}},
		{AccessPermissions: windows.GENERIC_WRITE, AccessMode: windows.SET_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeValue: windows.TrusteeValueFromSID(world)}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
}
