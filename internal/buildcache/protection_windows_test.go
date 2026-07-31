//go:build windows

package buildcache

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"golang.org/x/sys/windows"
)

func protectTestHome(t *testing.T, home string) {
	t.Helper()
	if err := protectWindowsPath(home); err != nil {
		t.Fatal(err)
	}
}

// TestWindowsProtectedBaseIsNeverObservableUnprotected pins the property that
// makes the cache root safe to create from more than one publisher at a time:
// a directory below the manager home is protected at the instant it becomes
// observable, never a moment later.
//
// Creating it and then applying the DACL leaves a window in which the new
// directory carries its parent's inherited entries. A concurrent publisher that
// opens it there is not wrong to refuse it -- inherited access is exactly what
// the protection check exists to catch -- so the whole publication fails with an
// untrusted-provenance verdict for state this manager had just created itself.
// Every racer below must therefore succeed.
func TestWindowsProtectedBaseIsNeverObservableUnprotected(t *testing.T) {
	home := t.TempDir()
	protectTestHome(t, home)
	base := filepath.Join(home, "cache", "build", "go-v1")

	const racers = 16
	start := make(chan struct{})
	errs := make(chan error, racers)
	var wait sync.WaitGroup
	for index := 0; index < racers; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			errs <- ensureProtectedBase(home, base)
		}()
	}
	close(start)
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent protected-base preparation: %v", err)
		}
	}

	// The directories are real and each one independently satisfies the same
	// protection check a later publication or sweep applies.
	for _, path := range []string{
		filepath.Join(home, "cache"),
		filepath.Join(home, "cache", "build"),
		base,
	} {
		dir, err := openWindowsProtected(path, true, false)
		if err != nil {
			t.Fatalf("protected cache directory %q: %v", path, err)
		}
		if err := dir.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

// TestWindowsCreateProtectedDirectoryRefusesAnExistingName keeps the atomic
// creator's contract identical to the os.Mkdir call it replaced, so a caller
// that relies on an existing name being an error still gets one.
func TestWindowsCreateProtectedDirectoryRefusesAnExistingName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entry")
	if err := createProtectedDirectory(path); err != nil {
		t.Fatal(err)
	}
	if err := createProtectedDirectory(path); !os.IsExist(err) {
		t.Fatalf("second creation = %v, want an already-exists error", err)
	}
}

func TestWindowsProtectedStateMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, store *Store, hit Result)
	}{
		{
			name: "artifact grants another principal mutation rights",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				grantWorldMutation(t, hit.ArtifactPath)
			},
		},
		{
			name: "artifact denies owner mutation rights",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				owner := currentTestUserSID(t)
				replaceTestDACL(t, hit.ArtifactPath, []windows.EXPLICIT_ACCESS{
					testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.WRITE_DAC, windows.DENY_ACCESS, windows.NO_INHERITANCE),
					testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.GRANT_ACCESS, windows.NO_INHERITANCE),
				})
			},
		},
		{
			name: "artifact has applicable group deny",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				owner := currentTestUserSID(t)
				world := worldTestSID(t)
				replaceTestDACL(t, hit.ArtifactPath, []windows.EXPLICIT_ACCESS{
					testExplicitAccess(world, windows.TRUSTEE_IS_WELL_KNOWN_GROUP, windows.WRITE_DAC, windows.DENY_ACCESS, windows.NO_INHERITANCE),
					testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.GRANT_ACCESS, windows.NO_INHERITANCE),
				})
			},
		},
		{
			name: "artifact has inherit-only owner allow",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				owner := currentTestUserSID(t)
				replaceTestDACL(t, hit.ArtifactPath, []windows.EXPLICIT_ACCESS{
					testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.SET_ACCESS, windows.INHERIT_ONLY),
				})
			},
		},
		{
			name: "artifact owner lacks required mutation rights",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				owner := currentTestUserSID(t)
				replaceTestDACL(t, hit.ArtifactPath, []windows.EXPLICIT_ACCESS{
					testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_WRITE, windows.SET_ACCESS, windows.NO_INHERITANCE),
				})
			},
		},
		{
			name: "artifact hard link",
			mutate: func(t *testing.T, store *Store, hit Result) {
				if err := os.Link(hit.ArtifactPath, filepath.Join(store.Home(), "artifact-hardlink")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "artifact special file",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				if err := os.Remove(hit.ArtifactPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(hit.ArtifactPath, 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "artifact reparse point",
			mutate: func(t *testing.T, store *Store, hit Result) {
				target := filepath.Join(store.Home(), "reparse-target")
				if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(hit.ArtifactPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, hit.ArtifactPath); err != nil {
					t.Skipf("creating Windows symlink requires host support: %v", err)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			if _, err := store.Publish(publication, testHomeLock{}); err != nil {
				t.Fatal(err)
			}
			hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
			if hit.Status != Hit {
				t.Fatalf("initial inspection = %+v", hit)
			}
			test.mutate(t, store, hit)
			result := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
			if result.Status != UntrustedProvenance {
				t.Fatalf("protected-state violation = %+v", result)
			}
		})
	}
}

func TestValidateWindowsSecurityPolicy(t *testing.T) {
	owner := currentTestUserSID(t)
	world := worldTestSID(t)

	tests := []struct {
		name          string
		owner         *windows.SID
		access        []windows.EXPLICIT_ACCESS
		wantUntrusted bool
	}{
		{
			name:   "canonical owner-only full control",
			owner:  owner,
			access: []windows.EXPLICIT_ACCESS{testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.SET_ACCESS, windows.NO_INHERITANCE)},
		},
		{
			name:  "owner deny",
			owner: owner,
			access: []windows.EXPLICIT_ACCESS{
				testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.WRITE_DAC, windows.DENY_ACCESS, windows.NO_INHERITANCE),
				testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.GRANT_ACCESS, windows.NO_INHERITANCE),
			},
			wantUntrusted: true,
		},
		{
			name:  "applicable group deny",
			owner: owner,
			access: []windows.EXPLICIT_ACCESS{
				testExplicitAccess(world, windows.TRUSTEE_IS_WELL_KNOWN_GROUP, windows.WRITE_DAC, windows.DENY_ACCESS, windows.NO_INHERITANCE),
				testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.GRANT_ACCESS, windows.NO_INHERITANCE),
			},
			wantUntrusted: true,
		},
		{
			name:          "inherit-only owner allow",
			owner:         owner,
			access:        []windows.EXPLICIT_ACCESS{testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.SET_ACCESS, windows.INHERIT_ONLY)},
			wantUntrusted: true,
		},
		{
			name:          "owner lacks required mutation rights",
			owner:         owner,
			access:        []windows.EXPLICIT_ACCESS{testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_WRITE, windows.SET_ACCESS, windows.NO_INHERITANCE)},
			wantUntrusted: true,
		},
		{
			name:          "wrong owner",
			owner:         world,
			access:        []windows.EXPLICIT_ACCESS{testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.SET_ACCESS, windows.NO_INHERITANCE)},
			wantUntrusted: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			acl, err := windows.ACLFromEntries(test.access, nil)
			if err != nil {
				t.Fatal(err)
			}
			err = validateWindowsOwner(test.owner, owner, "test path")
			if err == nil {
				err = validateWindowsDACL(acl, test.owner, "test path")
			}
			if test.wantUntrusted && err == nil {
				t.Fatal("security policy accepted untrusted owner or DACL")
			}
			if !test.wantUntrusted && err != nil {
				t.Fatalf("security policy rejected protected owner and DACL: %v", err)
			}
		})
	}
}

func currentTestUserSID(t *testing.T) *windows.SID {
	t.Helper()
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	return user.User.Sid
}

func worldTestSID(t *testing.T) *windows.SID {
	t.Helper()
	world, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	if err != nil {
		t.Fatal(err)
	}
	return world
}

func testExplicitAccess(sid *windows.SID, trusteeType windows.TRUSTEE_TYPE, mask windows.ACCESS_MASK, mode windows.ACCESS_MODE, inheritance uint32) windows.EXPLICIT_ACCESS {
	return windows.EXPLICIT_ACCESS{
		AccessPermissions: mask,
		AccessMode:        mode,
		Inheritance:       inheritance,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  trusteeType,
			TrusteeValue: windows.TrusteeValueFromSID(sid),
		},
	}
}

func replaceTestDACL(t *testing.T, path string, access []windows.EXPLICIT_ACCESS) {
	t.Helper()
	acl, err := windows.ACLFromEntries(access, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	); err != nil {
		t.Fatal(err)
	}
}

func grantWorldMutation(t *testing.T, path string) {
	t.Helper()
	owner := currentTestUserSID(t)
	world := worldTestSID(t)
	access := []windows.EXPLICIT_ACCESS{
		testExplicitAccess(owner, windows.TRUSTEE_IS_USER, windows.GENERIC_ALL, windows.SET_ACCESS, windows.NO_INHERITANCE),
		testExplicitAccess(world, windows.TRUSTEE_IS_WELL_KNOWN_GROUP, windows.GENERIC_WRITE, windows.GRANT_ACCESS, windows.NO_INHERITANCE),
	}
	replaceTestDACL(t, path, access)
}
