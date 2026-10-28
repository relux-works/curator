package projectlease

import (
	"fmt"
	"github.com/relux-works/curator/internal/managerlock"
)

// ProfileRemoval is the exact before/after of the shared reference ledger.
// Immutable binding history and project-home/session data remain untouched.
// The caller must enlist this CAS with installed-profile removal and its receipt.
type ProfileRemoval struct {
	Path          string
	Before, After []byte
}

func PrepareProfileRemoval(home, profile string, held managerlock.Held) (*ProfileRemoval, error) {
	if err := held.Validate(home); err != nil {
		return nil, err
	}
	l, err := load(home)
	if err != nil {
		return nil, err
	}
	changed := false
	for id, e := range l.Leases {
		if e.Composition.Profile == profile {
			if e.State != Released {
				return nil, fmt.Errorf("%s: profile has a live or uncertain project lease", Busy)
			}
			delete(l.Leases, id)
			changed = true
		}
	}
	for id, v := range l.Views {
		if v.Composition.Profile == profile {
			delete(l.Views, id)
			changed = true
		}
	}
	if !changed {
		return nil, nil
	}
	before, err := read(home, statePath(home))
	if err != nil {
		return nil, err
	}
	after, err := canonical(l)
	if err != nil {
		return nil, err
	}
	return &ProfileRemoval{Path: statePath(home), Before: before, After: after}, nil
}
