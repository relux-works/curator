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
func beginOperation(home string, options ...transaction.Option) (*operation, error) {
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
	engine, err := transaction.New(home, options...)
	if err != nil {
		_ = lock.Close()
		return nil, err
	}
	if err := engine.Recover(lock); err != nil {
		_ = lock.Close()
		return nil, fmt.Errorf("recover incomplete profile transactions: %w", err)
	}
	return &operation{lock: lock, engine: engine}, nil
}

func (op *operation) close() error {
	if op == nil || op.lock == nil {
		return nil
	}
	return op.lock.Close()
}

// publish commits manager-home record writes and removals as one journaled
// transaction. Write paths are staged and swapped under the held lock;
// removal paths use the transaction engine's absent desired state. A crash
// mid-publication resumes from the journal on the next operation instead of
// stranding a half-written record set.
func (op *operation) publish(files map[string][]byte, removals ...string) error {
	if len(files) == 0 && len(removals) == 0 {
		return nil
	}
	type change struct {
		path    string
		payload []byte
		remove  bool
	}
	changesByPath := make(map[string]change, len(files)+len(removals))
	for path := range files {
		changesByPath[path] = change{path: path, payload: files[path]}
	}
	for _, path := range removals {
		if _, exists := changesByPath[path]; exists {
			return fmt.Errorf("profile transaction both writes and removes %s", path)
		}
		changesByPath[path] = change{path: path, remove: true}
	}
	changes := make([]change, 0, len(changesByPath))
	for _, item := range changesByPath {
		changes = append(changes, item)
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].path < changes[j].path })
	var temps []string
	defer func() {
		for _, temp := range temps {
			_ = os.Remove(temp)
		}
	}()
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profiles"}
	for _, item := range changes {
		live := item.path
		if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
			return err
		}
		preimage, err := transaction.DigestTarget(transaction.KindBytes, live)
		if err != nil {
			return fmt.Errorf("read the current state of %s: %w", live, err)
		}
		target := transaction.Target{
			Class: "profile", Identifier: live, Kind: transaction.KindBytes,
			LivePath: live, PreimageDigest: preimage,
		}
		if !item.remove {
			temp, err := os.CreateTemp("", "curator-profile-record-*")
			if err != nil {
				return err
			}
			temps = append(temps, temp.Name())
			if _, err := temp.Write(item.payload); err != nil {
				_ = temp.Close()
				return err
			}
			if err := temp.Close(); err != nil {
				return err
			}
			target.StagedSource = temp.Name()
		}
		plan.Targets = append(plan.Targets, target)
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
