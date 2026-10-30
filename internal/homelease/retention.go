package homelease

import (
	"sort"

	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

type Roots struct{ Homes, Protected, BuildKeys, RuntimeEntries []string }

// ReferencesUnderLock joins the existing collector's mark set. Unknown and
// prepared reservations retain identical roots; no liveness or age is queried.
func ReferencesUnderLock(home string, lock transaction.HomeLock) (Roots, error) {
	var roots Roots
	w, ok := lock.(interface{ Held() managerlock.Held })
	if !ok {
		return roots, managerlock.ErrNotHeld
	}
	if err := w.Held().Validate(home); err != nil {
		return roots, err
	}
	// Stored root identity and state routes use the same physical manager
	// identity as acquisition, including when the caller supplied an alias.
	identity, err := managerlock.New(home)
	if err != nil {
		return roots, err
	}
	home = identity.Home()
	if err := w.Held().Validate(home); err != nil {
		return roots, err
	}
	if err := state.LegacyPending(home); err != nil {
		return roots, err
	}
	// Preserve the existing project collector owner filter when no home-lease
	// namespace exists. Pending home-lease publication provisions its namespace
	// before staging, so it cannot be hidden by this legacy fast path.
	if !state.HasState(home) {
		return roots, nil
	}
	engine, err := transaction.New(home)
	if err != nil {
		return roots, err
	}
	pending, err := engine.PendingRecovery(lock)
	if err != nil {
		return roots, state.Refuse("environment_lease_unreadable", "state_unreadable")
	}
	if pending {
		return roots, state.Refuse("environment_lease_recovery_required", "pending_mutation")
	}
	l, err := state.Load(home)
	if err != nil {
		return roots, err
	}
	sets := []map[string]bool{{}, {}, {}, {}}
	for _, e := range l.Leases {
		if !e.Record.Blocking() {
			continue
		}
		sets[0][e.Admission.Home] = true
		for i, values := range [][]string{e.Admission.Protected, e.Admission.BuildKeys, e.Admission.RuntimeEntries} {
			for _, v := range values {
				sets[i+1][v] = true
			}
		}
	}
	for i, out := range []*[]string{&roots.Homes, &roots.Protected, &roots.BuildKeys, &roots.RuntimeEntries} {
		for value := range sets[i] {
			*out = append(*out, value)
		}
		sort.Strings(*out)
	}
	return roots, nil
}
