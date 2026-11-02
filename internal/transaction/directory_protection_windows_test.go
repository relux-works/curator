//go:build windows

package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/privatedir"
	"golang.org/x/sys/windows"
)

// This target-specific control must run on Windows; a Unix pass cannot qualify
// ACL publication. It grants Everyone read only (never mutation), then proves
// the actual journal tightens that owned directory or restores its exact old
// descriptor after a fault, without changing an unrelated child payload.
func TestPrivateDirectoryACLPublication(t *testing.T) {
	for _, fault := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fault], func(t *testing.T) {
			home := t.TempDir()
			if err := privatedir.Protect(home); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(home, "owned")
			if err := privatedir.Make(dir); err != nil {
				t.Fatal(err)
			}
			user, err := windows.GetCurrentProcessToken().GetTokenUser()
			if err != nil {
				t.Fatal(err)
			}
			descriptor, err := windows.SecurityDescriptorFromString("O:" + user.User.Sid.String() + "D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")(A;OICI;FR;;;WD)")
			if err != nil {
				t.Fatal(err)
			}
			name, err := windows.UTF16PtrFromString(dir)
			if err != nil {
				t.Fatal(err)
			}
			h, err := windows.CreateFile(name, windows.WRITE_DAC, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
			if err != nil {
				t.Fatal(err)
			}
			err = windows.SetKernelObjectSecurity(h, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, descriptor)
			windows.CloseHandle(h)
			if err != nil {
				t.Fatal(err)
			}
			before, err := envunit.ObserveDirectory(dir)
			if err != nil || before.Mode != 0755 {
				t.Fatalf("owned public-read preimage: %+v %v", before, err)
			}
			child := filepath.Join(dir, "unrelated")
			if err = os.WriteFile(child, []byte("unrelated fixture bytes"), 0600); err != nil {
				t.Fatal(err)
			}
			target := fileTarget(t, "record", "record", filepath.Join(dir, "record"), filepath.Join(home, "source"), "old", "new")
			hit := false
			planted := errors.New("acl-publication-fault")
			engine := mustEngine(t, home, WithHooks(Hooks{Fault: func(event Event) error {
				if event.Point == PointAfterDirectoryInstall && event.LivePath == dir {
					hit = true
					if fault {
						return planted
					}
				}
				return nil
			}}))
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "directory-acl", ProjectIdentity: "fixture", Directories: []DirectoryAction{{LivePath: dir, Before: before, Mode: 0700}}, Targets: []Target{target}})
			if err != nil {
				t.Fatal(err)
			}
			if journal.Directories[0].After.Protection == before.Protection {
				t.Fatal("durable private protection was not captured")
			}
			err = engine.Commit(testLock{}, journal.TransactionID)
			if !hit {
				t.Fatal("actual directory publication not reached")
			}
			after, observeErr := envunit.ObserveDirectory(dir)
			if observeErr != nil {
				t.Fatal(observeErr)
			}
			if fault {
				if !errors.Is(err, planted) || after != before {
					t.Fatalf("ACL rollback lost preimage: %+v %v", after, err)
				}
				if got := mustRead(t, target.LivePath); got != "old" {
					t.Fatalf("fault changed unrelated target: %q", got)
				}
			} else {
				if err != nil || after.Mode != 0700 {
					t.Fatalf("private directory commit: %+v %v", after, err)
				}
				if err = privatedir.Validate(dir); err != nil {
					t.Fatal(err)
				}
			}
			if got := mustRead(t, child); got != "unrelated fixture bytes" {
				t.Fatalf("protection touched descendant: %q", got)
			}
		})
	}
}
