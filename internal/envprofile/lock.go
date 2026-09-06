package envprofile

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// lockTimeout bounds one wait for the manager-home mutation lock. Tests
// shorten it for contention cases; production waits up to 30 seconds and
// then reports environment_lock_unavailable.
var lockTimeout = 30 * time.Second

// operation binds one profile mutation to the held manager-home mutation
// lock and the profile transaction journal.
type operation struct {
	home   string
	lock   *managerlock.HomeLock
	engine *transaction.Engine
}

// beginOperation takes the manager-home mutation lock and recovers
// incomplete profile journals before the caller mutates anything: every
// profile mutation runs serialized like any other manager-home
// transaction, and its manager-home records publish through op.publish.
// The per-entry agent-home payloads are not transaction targets in this
// stage (see the switch.go package doc for why); they stay direct writes
// under the held lock.
func beginOperation(home string) (*operation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	manager, err := managerlock.New(home)
	if err != nil {
		return nil, err
	}
	lock, err := manager.AcquireHomeOnly(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire the manager-home mutation lock: %v", DiagLockUnavailable, err)
	}
	engine, err := transaction.New(home)
	if err != nil {
		_ = lock.Close()
		return nil, err
	}
	if err := engine.Recover(lock); err != nil {
		_ = lock.Close()
		return nil, fmt.Errorf("recover incomplete profile transactions: %w", err)
	}
	return &operation{home: home, lock: lock, engine: engine}, nil
}

func (op *operation) close() error {
	if op == nil || op.lock == nil {
		return nil
	}
	return op.lock.Close()
}

// publish commits manager-home record files as one journaled transaction:
// every path's desired bytes are staged and swapped under the held lock,
// so a crash mid-publication resumes from the journal on the next
// operation instead of stranding a half-written record.
func (op *operation) publish(files map[string][]byte) error {
	if len(files) == 0 {
		return nil
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var temps []string
	defer func() {
		for _, temp := range temps {
			_ = os.Remove(temp)
		}
	}()
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profiles"}
	for _, live := range paths {
		if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
			return err
		}
		temp, err := os.CreateTemp("", "curator-profile-record-*")
		if err != nil {
			return err
		}
		temps = append(temps, temp.Name())
		if _, err := temp.Write(files[live]); err != nil {
			_ = temp.Close()
			return err
		}
		if err := temp.Close(); err != nil {
			return err
		}
		preimage, err := transaction.DigestTarget(transaction.KindBytes, live)
		if err != nil {
			return fmt.Errorf("read the current state of %s: %w", live, err)
		}
		plan.Targets = append(plan.Targets, transaction.Target{
			Class: "profile", Identifier: live, Kind: transaction.KindBytes,
			LivePath: live, StagedSource: temp.Name(), PreimageDigest: preimage,
		})
	}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return fmt.Errorf("prepare profile records: %w", err)
	}
	for _, temp := range temps {
		_ = os.Remove(temp)
	}
	temps = nil
	if err := op.engine.Commit(op.lock, plan.TransactionID); err != nil {
		return fmt.Errorf("commit profile records: %w", err)
	}
	return nil
}

// newTransactionID names one profile journal.
func newTransactionID() string {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("profile-%d", time.Now().UnixNano())
	}
	return "profile-" + hex.EncodeToString(random[:])
}
