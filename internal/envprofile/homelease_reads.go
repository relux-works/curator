package envprofile

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// Legacy homes still need pending-intent certainty for read-only callers. This
// path neither provisions the new namespace nor imposes its private-file format
// on existing unrelated transaction state. A missing lock cannot hide intent.
func homeLeaseReadLegacy(home string, held *managerlock.Held) error {
	if _, err := os.Lstat(filepath.Join(home, "state", "env-migration.json")); !errors.Is(err, os.ErrNotExist) {
		return state.LegacyPending(home)
	}
	var witness transaction.HomeLock
	if held != nil {
		if err := held.Validate(home); err != nil {
			return err
		}
		witness = homeLeaseReadWitness{home: home, held: *held}
	} else {
		manager, err := managerlock.New(home)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
		defer cancel()
		lock, err := manager.AcquireExistingHomeOnly(ctx)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if _, err := os.Lstat(filepath.Join(home, "state", "transactions", "v1")); errors.Is(err, os.ErrNotExist) {
					return nil
				}
			}
			if ctx.Err() != nil {
				return state.Refuse("environment_lock_unavailable", "mutation_lock_busy")
			}
			return state.Refuse("environment_lease_unreadable", "state_unreadable")
		}
		defer func() { _ = lock.Close() }()
		witness = lock
	}
	engine, err := transaction.New(home)
	if err != nil {
		return err
	}
	pending, err := engine.PendingRecovery(witness)
	if err != nil {
		return state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	if pending {
		return state.Refuse("environment_lease_recovery_required", "pending_mutation")
	}
	return nil
}
