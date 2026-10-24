package projectlease

import (
	"fmt"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// RecoverUnderLock shares the conservative lifecycle transition between every
// recovery owner. Completing file publication does not establish runner life:
// all nonreleased leases become unknown until authoritative reconciliation.
// The caller retains the exact same acquisition through both publications.
func RecoverUnderLock(home string, lock transaction.HomeLock, engine *transaction.Engine) (bool, error) {
	recovered, err := engine.RecoverWithOutcome(lock)
	if err != nil || !recovered {
		return recovered, err
	}
	manager, err := managerlock.New(home)
	if err != nil {
		return recovered, err
	}
	home = manager.Home()
	if !HasState(home) {
		return recovered, nil
	}
	held, ok := lock.(*managerlock.HomeLock)
	if !ok {
		return recovered, managerlock.ErrNotHeld
	}
	if err = held.Held().Validate(home); err != nil {
		return recovered, err
	}
	l, err := load(home)
	if err != nil {
		return recovered, fmt.Errorf("recovered project lease evidence unreadable; retain all roots: %w", err)
	}
	changed := false
	for id, e := range l.Leases {
		if e.State != Released && e.State != Unknown {
			e.State = Unknown
			l.Leases[id] = e
			changed = true
		}
	}
	if !changed {
		return recovered, nil
	}
	m := &Manager{home: home}
	return recovered, m.publish(l, held, engine, map[string][]byte{}, nil)
}
