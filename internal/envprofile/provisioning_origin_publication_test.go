package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

// Proves the real profile-record publication owner journals absent parents and
// private output, restoring absence on a later file error. This is a record
// primitive control, not complete home/origin provisioning qualification.
func TestProvisioningOriginProfileRecordParents(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "rollback"}[fail], func(t *testing.T) {
			home := t.TempDir()
			live := filepath.Join(home, "profiles", "p", "source.json")
			failure := errors.New("owned profile record fixture failure")
			directoryReached, fileReached := false, false
			op, err := beginOperation(home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
				if event.Point == transaction.PointAfterDirectoryInstall {
					directoryReached = true
				}
				if event.Point == transaction.PointBeforeInstall {
					fileReached = true
					if fail {
						return failure
					}
				}
				return nil
			}}))
			if err != nil {
				t.Fatal(err)
			}
			defer op.close()
			err = op.publish(map[string][]byte{live: []byte(`{"fixture":true}`)})
			if !directoryReached || !fileReached {
				t.Fatalf("publication owner not reached: directories=%v file=%v err=%v", directoryReached, fileReached, err)
			}
			if fail {
				if !errors.Is(err, failure) {
					t.Fatalf("wrong failure: %v", err)
				}
				if _, err := os.Lstat(filepath.Join(home, "profiles")); !os.IsNotExist(err) {
					t.Fatalf("failed record left parent: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if err := privatedir.ValidateFile(live); err != nil {
					t.Fatal(err)
				}
				if data, err := os.ReadFile(live); err != nil || string(data) != `{"fixture":true}` {
					t.Fatalf("record %q %v", data, err)
				}
			}
		})
	}
}
