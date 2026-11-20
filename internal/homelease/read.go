package homelease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/transaction"
)

// List observes existing authority without recovery, provisioning or liveness
// reconciliation. Terminal tombstones remain visible; no timeout forgets them.
func (m *Manager) List(ctx context.Context, environment, profile string) ([]Record, error) {
	environment = envregistry.NormalizeEnvID(environment)
	if environment != "" && environment != "claude_code" && environment != "codex_cli" && environment != "muse" {
		return nil, state.Refuse("environment_unknown", "environment_absent")
	}
	if profile != "" && !state.ValidID(profile) {
		return nil, state.Refuse("usage", "invalid_selector")
	}
	locks, err := managerlock.New(m.home)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	lock, err := locks.AcquireExistingHomeOnly(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, state.Refuse("environment_lock_unavailable", "mutation_lock_busy")
		}
		if errors.Is(err, os.ErrNotExist) && !state.HasState(m.home) && !projectlease.HasState(m.home) {
			// Missing lock is empty only if no journal inventory can be hidden.
			_, pendingErr := os.Lstat(filepath.Join(m.home, "state", "transactions", "v1"))
			_, legacyErr := os.Lstat(filepath.Join(m.home, "state", "env-migration.json"))
			if errors.Is(pendingErr, os.ErrNotExist) && errors.Is(legacyErr, os.ErrNotExist) {
				return []Record{}, nil
			}
		}
		return nil, state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	defer func() { _ = lock.Close() }()
	if err := state.LegacyPending(m.home); err != nil {
		return nil, err
	}
	engine, err := transaction.New(m.home)
	if err != nil {
		return nil, err
	}
	pending, err := engine.PendingRecovery(lock)
	if err != nil {
		return nil, state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	if pending {
		return nil, state.Refuse("environment_lease_recovery_required", "pending_mutation")
	}
	if _, err := projectlease.ReferencesUnderLock(m.home, lock); err != nil {
		return nil, state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	l, err := state.Load(m.home)
	if err != nil {
		return nil, err
	}
	rows := []Record{}
	for _, r := range state.Records(l) {
		if (environment == "" || r.Environment == environment) && (profile == "" || r.Profile == profile) {
			rows = append(rows, r)
		}
	}
	return rows, nil
}
