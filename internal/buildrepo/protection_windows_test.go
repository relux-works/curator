//go:build windows

package buildrepo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func windowsArtifactFixture(t *testing.T) (*DiskProtectedStore, string, map[string]any, string) {
	t.Helper()
	store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
	key := "sha256:" + strings.Repeat("a", 64)
	input := map[string]any{"command": "tool", "target": map[string]any{"goos": "windows"}}
	artifact := []byte("trusted artifact")
	if _, err := store.StoreArtifact(key, input, "tool", artifact, testExecutionReceiptBytes(t, input, artifact)); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(store.Root, "artifacts", strings.TrimPrefix(key, "sha256:"))
	return store, key, input, entry
}

func TestWindowsProtectedStoreReusesOwnerPrivateArtifact(t *testing.T) {
	store, key, input, _ := windowsArtifactFixture(t)
	hit, err := store.LookupArtifact(key, input, false)
	if err != nil {
		t.Fatal(err)
	}
	if hit == nil || string(hit.Bytes) != "trusted artifact" {
		t.Fatalf("hit=%+v", hit)
	}
}

func TestWindowsProtectedStoreReusesOwnerPrivateSnapshot(t *testing.T) {
	snapshot, _, effective := pipelineFixture(t)
	store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
	key, err := SnapshotKey(effective, snapshot.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.StoreSnapshot(key, snapshot); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadSnapshot(key, false)
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil || loaded.Digest != snapshot.Digest || loaded.Commit != snapshot.Commit {
		t.Fatalf("snapshot=%+v", loaded)
	}
}

func TestWindowsProtectedSecurityDescriptorRejectsWrongOwnerAndDACL(t *testing.T) {
	owner, err := currentWindowsUserSID()
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name      string
		sddl      string
		directory bool
	}{
		{name: "wrong-owner", sddl: "O:BAD:P(A;;FA;;;BA)"},
		{name: "unprotected-dacl", sddl: "O:" + owner.String() + "D:(A;;FA;;;" + owner.String() + ")"},
		{name: "additional-principal", sddl: "O:" + owner.String() + "D:P(A;;FA;;;" + owner.String() + ")(A;;FR;;;WD)"},
		{name: "directory-flags-missing", sddl: "O:" + owner.String() + "D:P(A;;FA;;;" + owner.String() + ")", directory: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sd, err := windows.SecurityDescriptorFromString(testCase.sddl)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateWindowsSecurityDescriptor(sd, testCase.directory); err == nil {
				t.Fatal("untrusted descriptor accepted")
			}
		})
	}
}

func TestWindowsProtectedArtifactAdversarialStateQuarantines(t *testing.T) {
	t.Run("hard-link", func(t *testing.T) {
		store, key, input, entry := windowsArtifactFixture(t)
		if err := os.Link(filepath.Join(entry, "artifact"), filepath.Join(entry, "artifact-link")); err != nil {
			t.Fatal(err)
		}
		assertWindowsArtifactQuarantined(t, store, key, input, entry)
	})

	t.Run("private-dacl-lost", func(t *testing.T) {
		store, key, input, entry := windowsArtifactFixture(t)
		owner, err := currentWindowsUserSID()
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(entry, "receipt.json")
		handle, err := openWindowsProtectedPath(path, windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES)
		if err != nil {
			t.Fatal(err)
		}
		sd, err := windows.SecurityDescriptorFromString("O:" + owner.String() + "D:(A;;FA;;;" + owner.String() + ")(A;;FR;;;WD)")
		if err == nil {
			var dacl *windows.ACL
			dacl, _, err = sd.DACL()
			if err == nil {
				err = windows.SetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
			}
		}
		windows.CloseHandle(handle)
		if err != nil {
			t.Fatal(err)
		}
		assertWindowsArtifactQuarantined(t, store, key, input, entry)
	})

	t.Run("boundary-dacl-lost", func(t *testing.T) {
		store, key, input, _ := windowsArtifactFixture(t)
		owner, err := currentWindowsUserSID()
		if err != nil {
			t.Fatal(err)
		}
		handle, err := openWindowsProtectedPath(store.Root, windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES)
		if err != nil {
			t.Fatal(err)
		}
		sd, err := windows.SecurityDescriptorFromString("O:" + owner.String() + "D:(A;OICI;FA;;;" + owner.String() + ")(A;OICI;FR;;;WD)")
		if err == nil {
			var dacl *windows.ACL
			dacl, _, err = sd.DACL()
			if err == nil {
				err = windows.SetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
			}
		}
		windows.CloseHandle(handle)
		if err != nil {
			t.Fatal(err)
		}
		hit, lookupErr := store.LookupArtifact(key, input, true)
		if hit != nil || ErrorCode(lookupErr) != CodeProtectedBoundaryUntrusted {
			t.Fatalf("hit=%v err=%v", hit, lookupErr)
		}
	})

	t.Run("reparse-point", func(t *testing.T) {
		store, key, input, entry := windowsArtifactFixture(t)
		receipt := filepath.Join(entry, "receipt.json")
		target := filepath.Join(t.TempDir(), "outside.json")
		content, err := os.ReadFile(receipt)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, content, 0o600); err != nil {
			t.Fatal(err)
		}
		if err = os.Remove(receipt); err != nil {
			t.Fatal(err)
		}
		if err = os.Symlink(target, receipt); err != nil {
			t.Skipf("Windows host cannot create a file reparse point: %v", err)
		}
		assertWindowsArtifactQuarantined(t, store, key, input, entry)
	})
}

func TestWindowsProtectedSnapshotHardLinkQuarantines(t *testing.T) {
	snapshot, _, effective := pipelineFixture(t)
	store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
	key, err := SnapshotKey(effective, snapshot.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.StoreSnapshot(key, snapshot); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(store.Root, "snapshots", strings.TrimPrefix(key, "sha256:"))
	file := filepath.Join(entry, "files", "outside-secret.txt")
	if err = os.Link(file, filepath.Join(entry, "snapshot-file-link")); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadSnapshot(key, true)
	if loaded != nil || err == nil {
		t.Fatalf("snapshot=%v err=%v", loaded, err)
	}
	if _, statErr := os.Stat(entry); !os.IsNotExist(statErr) {
		t.Fatalf("untrusted snapshot was not quarantined: %v", statErr)
	}
}

func TestWindowsProtectedArtifactPathSwapCannotReturnBytes(t *testing.T) {
	store, key, input, entry := windowsArtifactFixture(t)
	receipt := filepath.Join(entry, "receipt.json")
	replacement := filepath.Join(entry, "replacement.json")
	content, err := os.ReadFile(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(replacement, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = secureWindowsPath(replacement, false); err != nil {
		t.Fatal(err)
	}
	swapped := false
	store.proofHook = func(phase string) {
		if phase != "file-reopen" || swapped {
			return
		}
		swapped = true
		if err := os.Rename(receipt, filepath.Join(entry, "retained.json")); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(replacement, receipt); err != nil {
			t.Fatal(err)
		}
	}
	hit, err := store.LookupArtifact(key, input, true)
	if hit != nil || err == nil || !swapped {
		t.Fatalf("hit=%v err=%v swapped=%v", hit, err, swapped)
	}
	if _, statErr := os.Stat(entry); !os.IsNotExist(statErr) {
		t.Fatalf("raced entry was not quarantined: %v", statErr)
	}
}

func TestWindowsProtectedArtifactEntrySwapCannotReturnBytes(t *testing.T) {
	store, key, input, entry := windowsArtifactFixture(t)
	replacement := filepath.Join(filepath.Dir(entry), ".replacement")
	if err := os.Mkdir(replacement, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"artifact", "receipt.json", "execution-receipt.ccj.json"} {
		data, err := os.ReadFile(filepath.Join(entry, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(replacement, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := secureProtectedTree(replacement); err != nil {
		t.Fatal(err)
	}
	swapped := false
	store.proofHook = func(phase string) {
		if phase != "directory-guard" || swapped {
			return
		}
		swapped = true
		if err := os.Rename(entry, entry+"-retained"); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(replacement, entry); err != nil {
			t.Fatal(err)
		}
	}
	hit, err := store.LookupArtifact(key, input, true)
	if hit != nil || err == nil || !swapped {
		t.Fatalf("hit=%v err=%v swapped=%v", hit, err, swapped)
	}
	if _, statErr := os.Stat(entry); !os.IsNotExist(statErr) {
		t.Fatalf("replacement entry was not quarantined: %v", statErr)
	}
}

func assertWindowsArtifactQuarantined(t *testing.T, store *DiskProtectedStore, key string, input map[string]any, entry string) {
	t.Helper()
	hit, err := store.LookupArtifact(key, input, true)
	if hit != nil || err == nil {
		t.Fatalf("hit=%v err=%v", hit, err)
	}
	if _, statErr := os.Stat(entry); !os.IsNotExist(statErr) {
		t.Fatalf("untrusted entry was not quarantined: %v", statErr)
	}
	entries, statErr := os.ReadDir(filepath.Join(store.Root, "quarantine"))
	if statErr != nil || len(entries) != 1 {
		t.Fatalf("quarantine=%v err=%v", entries, statErr)
	}
}
