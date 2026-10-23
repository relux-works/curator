package projectlease

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// The forward policy independently rejects a replaced acquisition. A current
// acquisition accepts the identical valid desired ledger before replacement.
func TestLeaseTransactionPolicyCurrentLock(t *testing.T) {
	f := setup(t)
	acquire(t, f, "one")
	lm, err := managerlock.New(f.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := lm.AcquireHomeOnly(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	path := statePath(lm.Home())
	digest, err := transaction.DigestTarget(transaction.KindBytes, path)
	if err != nil {
		t.Fatal(err)
	}
	journal := transaction.Journal{ProjectIdentity: "project-leases", Targets: []transaction.TargetRecord{{Class: "project-lease", LivePath: path, DesiredDigest: digest}}}
	check := TransactionCheck(f.home)
	if err = check(lock, journal); err != nil {
		t.Fatal("current transaction policy refused", err)
	}
	paths, err := filepath.Glob(filepath.Join(f.home, "state", "locks", "v1", "*.lock"))
	if err != nil || len(paths) != 1 {
		t.Fatal("lock fixture", paths, err)
	}
	if err = os.Rename(paths[0], paths[0]+".old"); err != nil {
		if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(32)) {
			t.Skip("host refuses open-lock replacement")
		}
		t.Fatal(err)
	}
	if err = os.WriteFile(paths[0], nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err = check(lock, journal); !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("stale transaction policy accepted: %v", err)
	}
}
