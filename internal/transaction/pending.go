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

// PendingRecoveryMatching observes matching intent through detached journal
// snapshots and validates the complete inventory. It never resumes intent.
func (engine *Engine) PendingRecoveryMatching(lock HomeLock, matches func(Journal) (bool, error)) (bool, error) {
	if matches == nil {
		return false, journalf("pending transaction selector is absent")
	}
	return engine.pendingRecovery(lock, "", matches)
}

func (engine *Engine) pendingRecovery(lock HomeLock, projectIdentity string, selectors ...func(Journal) (bool, error)) (bool, error) {
	if err := requireHomeLock(lock); err != nil {
		return false, err
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	// Inspect the manager-relative route without following a linked inventory.
	for _, path := range []string{filepath.Dir(filepath.Dir(engine.journalRoot)), filepath.Dir(engine.journalRoot), engine.journalRoot} {
		state, err := stateread.Lstat(path)
		if err != nil {
			return false, err
		}
		if state.Kind == stateread.KindAbsent {
			return false, nil
		}
		if !state.Info.IsDir() {
			return false, stateread.UnusableError(path, journalf("transaction inventory is not a directory"))
		}
	}
	ids, err := engine.journalIDs()
	if err != nil {
		return false, err
	}
	pending := false
	for _, id := range ids {
		journal, err := engine.loadJournal(id)
		if err != nil {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				return false, stateread.UnusableError(pathErr.Path, err)
			}
			return false, err
		}
		if projectIdentity == "" || journal.ProjectIdentity == projectIdentity {
			matches := true
			if len(selectors) > 0 {
				matches, err = selectors[0](*cloneJournal(journal))
				if err != nil {
					return false, err
				}
			}
			pending = pending || matches
		}
	}
	return pending, nil
}
