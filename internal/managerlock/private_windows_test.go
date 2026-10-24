//go:build windows

package managerlock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/privatedir"
	"golang.org/x/sys/windows"
)

// Claim: a real acquisition hardens its owned inheriting home and creates
// protected lock state; replacement is permitted by Windows sharing but
// invalidates the old witness, while the replacement acquisition stays valid.
func TestWindowsPrivateLockReplacement(t *testing.T) {
	home := t.TempDir()
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	for _, rel := range []string{"", "state", "state/locks", "state/locks/v1"} {
		if err := privatedir.Validate(filepath.Join(home, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("manager directory lost private DACL: %v", err)
		}
	}
	path := lock.lock.file.Name()
	if err := privatedir.ValidateFile(path); err != nil {
		t.Fatalf("manager lock lost private DACL: %v", err)
	}
	if err := lock.Held().Validate(home); err != nil {
		t.Fatal(err)
	}
	// Reopen through the existing-name branch as well as the first creator.
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	lock, err = manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if err := os.Rename(path, path+".old"); err != nil {
		t.Fatalf("Windows lock replacement witness unavailable: %v", err)
	}
	if file, err := privatedir.Create(path, 0o600); err != nil {
		t.Fatal(err)
	} else if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(lock.Held().Validate(home), ErrNotHeld) {
		t.Fatal("replaced Windows lock witness accepted")
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	other, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	current, err := other.AcquireExistingHomeOnly(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = current.Close() }()
	if err := current.Held().Validate(home); err != nil {
		t.Fatalf("replacement positive control: %v", err)
	}
}

// Claim: initialization refuses an explicitly broadened manager boundary and
// does not create lock state or erase the foreign grant while refusing it.
func TestWindowsManagerRefusesBroadenedBoundary(t *testing.T) {
	home := t.TempDir()
	if err := privatedir.Protect(home); err != nil {
		t.Fatal(err)
	}
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
	if err := windows.SetNamedSecurityInfo(home, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if lock, err := manager.AcquireHomeOnly(context.Background(), false); err == nil || !strings.Contains(err.Error(), "another principal") {
		if lock != nil {
			_ = lock.Close()
		}
		t.Fatalf("broadened manager boundary got wrong refusal: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(home, "state")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("refusal created manager state: %v", err)
	}
	if err := privatedir.Validate(home); err == nil || !strings.Contains(err.Error(), "another principal") {
		t.Fatal("refusal erased the broadened ACL")
	}
}

// Claim: creating a missing home below a directory alias preserves the real
// target's spelling and private DACLs. The real and aliased names contend;
// releasing that same acquisition permits the real-name positive control.
func TestWindowsPrivateMissingHomeAlias(t *testing.T) {
	base := t.TempDir()
	realParent := filepath.Join(base, "RealParent")
	if err := os.Mkdir(realParent, 0o700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(realParent, alias); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(alias, "Missing", "CuratorHome")
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(t.Context(), false)
	if err != nil {
		t.Fatalf("aliased private home creation failed: %v", err)
	}
	defer func() { _ = lock.Close() }()
	physical := filepath.Join(realParent, "Missing", "CuratorHome")
	for _, part := range []struct{ parent, name string }{{realParent, "Missing"}, {filepath.Join(realParent, "Missing"), "CuratorHome"}} {
		entries, err := os.ReadDir(part.parent)
		if err != nil || len(entries) != 1 || entries[0].Name() != part.name {
			t.Fatalf("missing suffix spelling changed: %v %v", entries, err)
		}
	}
	for _, rel := range []string{"", "state", "state/locks", "state/locks/v1"} {
		if err := privatedir.Validate(filepath.Join(physical, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("aliased manager directory lost private DACL: %v", err)
		}
	}
	other, err := New(physical)
	if err != nil || other.Home() != manager.Home() {
		t.Fatalf("alias identity differs: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	if unexpected, err := other.AcquireHomeOnly(ctx, false); !errors.Is(err, context.DeadlineExceeded) {
		if unexpected != nil {
			_ = unexpected.Close()
		}
		t.Fatalf("physical twin escaped aliased lock: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	current, err := other.AcquireExistingHomeOnly(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = current.Close() }()
	if err := current.Held().Validate(home); err != nil {
		t.Fatalf("released alias positive control: %v", err)
	}
}
