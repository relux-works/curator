package transaction

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/stateread"
)

// PendingRecovery observes all durable recovery records without resuming a
// transaction or cleaning a journal/tomb. Invalid or unreadable evidence is an
// error even when another record already establishes pending recovery.
func (engine *Engine) PendingRecovery(lock HomeLock) (bool, error) {
	return engine.pendingRecovery(lock, "")
}

// PendingRecoveryFor observes pending intent for exactly one transaction owner.
// It still validates the complete inventory: unrelated unreadable intent cannot
// be silently ignored. Observation never commits, repairs or cleans a record.
func (engine *Engine) PendingRecoveryFor(lock HomeLock, projectIdentity string) (bool, error) {
	if projectIdentity == "" {
		return false, journalf("pending transaction owner is empty")
	}
	return engine.pendingRecovery(lock, projectIdentity)
}

func (engine *Engine) pendingRecovery(lock HomeLock, projectIdentity string) (bool, error) {
	journals, err := engine.PendingJournals(lock)
	if err != nil {
		return false, err
	}
	for _, journal := range journals {
		if projectIdentity == "" || journal.ProjectIdentity == projectIdentity {
			return true, nil
		}
	}
	return false, nil
}

// PendingJournals returns detached, validated durable intent without resuming
// or cleaning it. Callers must still validate target contents and their own
// authority: a journal's presence alone never admits a publication. The complete
// inventory is checked before any records are returned, including other owners.
func (engine *Engine) PendingJournals(lock HomeLock) ([]Journal, error) {
	if err := requireHomeLock(lock); err != nil {
		return nil, err
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	// Inspect the manager-relative route without following a linked inventory.
	for _, path := range []string{filepath.Dir(filepath.Dir(engine.journalRoot)), filepath.Dir(engine.journalRoot), engine.journalRoot} {
		state, err := stateread.Lstat(path)
		if err != nil {
			return nil, err
		}
		if state.Kind == stateread.KindAbsent {
			return nil, nil
		}
		if !state.Info.IsDir() {
			return nil, stateread.UnusableError(path, journalf("transaction inventory is not a directory"))
		}
	}
	ids, err := engine.journalIDs()
	if err != nil {
		return nil, err
	}
	var journals []Journal
	for _, id := range ids {
		journal, err := engine.loadJournal(id)
		if err != nil {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				return nil, stateread.UnusableError(pathErr.Path, err)
			}
			return nil, err
		}
		journals = append(journals, *cloneJournal(journal))
	}
	return journals, nil
}
