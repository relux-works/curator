package envprofile

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// The default envprofile recovery engine applies the project-lease policy instead
// of publishing an invalid lease journal with no desired ledger. A journal
// owned by profiles retains the prior recovery behavior and commits normally.
func TestDefaultRecoveryChecksLeaseJournal(t *testing.T) {
	for _, owner := range []string{"profiles", "project-leases"} {
		t.Run(owner, func(t *testing.T) {
			home := t.TempDir()
			if err := os.Chmod(home, 0700); err != nil {
				t.Fatal(err)
			}
			lm, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			home = lm.Home()
			lock, err := lm.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(home, "owned.txt")
			stage := filepath.Join(t.TempDir(), "stage")
			if err = os.WriteFile(target, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(stage, []byte("new"), 0600); err != nil {
				t.Fatal(err)
			}
			pre, err := transaction.DigestTarget(transaction.KindBytes, target)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := transaction.New(home)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Prepare(lock, transaction.Plan{TransactionID: "recovery-owner", ProjectIdentity: owner, Targets: []transaction.Target{{Class: "project-lease", Identifier: target, LivePath: target, StagedSource: stage, PreimageDigest: pre}}}); err != nil {
				t.Fatal(err)
			}
			if err = lock.Close(); err != nil {
				t.Fatal(err)
			}
			op, err := beginOperation(home)
			if op != nil {
				defer op.close()
			}
			raw, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if owner == "project-leases" {
				if err == nil || !strings.Contains(err.Error(), "lacks its ledger") || string(raw) != "old" {
					t.Fatalf("default lease recovery policy missing: %v bytes=%q", err, raw)
				}
			} else if err != nil || string(raw) != "new" {
				t.Fatal("legacy profile recovery changed", err, string(raw))
			}
		})
	}
}
