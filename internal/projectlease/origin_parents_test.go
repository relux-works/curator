package projectlease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/transaction"
)

// The actual initial lease owner has a durable complete plan before creating
// its ledger/binding parents. A fault after parent publication restores their
// absence; the adjacent healthy control reaches a real binding+ledger commit.
func TestProvisioningOriginLeasePublicationParents(t *testing.T) {
	for _, fault := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "parent-fault"}[fault], func(t *testing.T) {
			f := setup(t)
			parent := filepath.Dir(statePath(f.home))
			prepared, hit := false, false
			planted := errors.New("lease-parent-fault")
			f.m.hooks = transaction.Hooks{Fault: func(e transaction.Event) error {
				if e.Point == transaction.PointPrepared {
					prepared = true
					if _, err := os.Lstat(parent); !os.IsNotExist(err) {
						t.Error("lease parents created before durable journal")
					}
				}
				if fault && e.Point == transaction.PointAfterDirectoryInstall && e.LivePath == parent {
					hit = true
					return planted
				}
				return nil
			}}
			_, err := f.m.Acquire(context.Background(), f.reserve("parent-fixture"))
			if !prepared {
				t.Fatalf("actual lease Prepare not reached: %v", err)
			}
			if fault {
				if !hit || !errors.Is(err, planted) {
					t.Fatalf("intended parent fault: hit=%t %v", hit, err)
				}
				if _, err := os.Lstat(parent); !os.IsNotExist(err) {
					t.Fatalf("failed lease left parent namespace: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				l, err := load(f.home)
				if err != nil || len(l.Leases) != 1 {
					t.Fatalf("healthy lease did not publish: %v", err)
				}
				if _, err := os.Lstat(bindingPath(f.home, "parent-fixture")); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
