//go:build windows

package scriptworker

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func setWindowsExecFixtureOwner(t *testing.T, path, sidText string) {
	t.Helper()
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = token.Close() }()
	name, err := windows.UTF16PtrFromString("SeRestorePrivilege")
	if err != nil {
		t.Fatal(err)
	}
	state := windows.Tokenprivileges{PrivilegeCount: 1}
	if err := windows.LookupPrivilegeValue(nil, name, &state.Privileges[0].Luid); err != nil {
		t.Fatal(err)
	}
	state.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED
	var previous windows.Tokenprivileges
	var returned uint32
	if err := windows.AdjustTokenPrivileges(token, false, &state, uint32(unsafe.Sizeof(previous)), &previous, &returned); err != nil {
		t.Fatalf("hosted Windows fixture requires SeRestorePrivilege: %v", err)
	}
	defer func() {
		if err := windows.AdjustTokenPrivileges(token, false, &previous, 0, nil, nil); err != nil {
			t.Errorf("restore fixture token privileges: %v", err)
		}
	}()
	sid, err := windows.StringToSid(sidText)
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION, sid, nil, nil, nil); err != nil {
		t.Fatalf("set real fixture owner %s: %v", sidText, err)
	}
}

// This test never replaces the OS seam. Hosted Windows must exercise the real
// file IDs, owner descriptor and complete FindFirst/NextFileName enumeration.
func TestWindowsNativeExecHardlinkOrigins(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		owner string
		extra bool
		want  bool
	}{
		{name: "component-store", owner: trustedInstallerSID, want: true},
		{name: "system-owned-component-store", owner: localSystemSID, want: true},
		{name: "noncomponent-store", owner: trustedInstallerSID, extra: true},
		{name: "unowned-file"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := mustPhysical(t, t.TempDir())
			target := filepath.Join(root, "System32", "cmd.exe")
			alias := filepath.Join(root, "WinSxS", "cmd.exe")
			writeTestFile(t, target, []byte("native executable fixture"), 0o755)
			if err := os.MkdirAll(filepath.Dir(alias), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(target, alias); err != nil {
				t.Fatal(err)
			}
			if testCase.extra {
				outside := filepath.Join(root, "OtherLinks", "cmd.exe")
				if err := os.MkdirAll(filepath.Dir(outside), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(target, outside); err != nil {
					t.Fatal(err)
				}
			}
			if testCase.owner != "" {
				setWindowsExecFixtureOwner(t, target, testCase.owner)
			}
			file, err := os.Open(target)
			if err != nil {
				t.Fatal(err)
			}
			origin, err := nativeWindowsExecHardlinks(file, target)
			_ = file.Close()
			if err != nil {
				t.Fatalf("native observation: %v", err)
			}
			wantCount := uint32(2)
			if testCase.extra {
				wantCount = 3
			}
			if origin.Count != wantCount || len(origin.Links) != int(wantCount) {
				t.Fatalf("native links = %+v, want %d", origin, wantCount)
			}
			if testCase.owner != "" && origin.OwnerSID != testCase.owner {
				t.Fatalf("native owner = %q, want %q", origin.OwnerSID, testCase.owner)
			}
			t.Logf("native owner=%s links=%d paths=%q", origin.OwnerSID, origin.Count, origin.Links)
			identity, found, err := resolveExecForPlatform("cmd.exe", nil, nil, "windows", []string{"SYSTEMROOT=" + root}, true)
			if err != nil || found != testCase.want {
				t.Fatalf("native production resolver found=%t err=%v, want %t", found, err, testCase.want)
			}
			if found {
				if err := VerifyExec(identity); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(target, filepath.Join(root, "outside.exe")); err != nil {
					t.Fatal(err)
				}
				if err := VerifyExec(identity); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
					t.Fatalf("native added alias at launch boundary: %v", err)
				}
			}
		})
	}

	t.Run("real-system32-cmd", func(t *testing.T) {
		identity, found, err := resolveExecForPlatform("cmd.exe", nil, nil, "windows", os.Environ(), true)
		if err != nil || !found {
			t.Fatalf("native System32 cmd.exe rejected: %v", err)
		}
		if err := VerifyExec(identity); err != nil {
			t.Fatal(err)
		}
		t.Logf("native platform executable accepted at %s", identity.Path)
	})
}
