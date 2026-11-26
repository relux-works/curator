package state

import (
	"path/filepath"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/stateread"
)

// LegacyPending retains the existing env-migration journal without reading
// credential surfaces or attempting recovery. Only the migration owner may
// validate/reconcile its contents; the existence of its intent prevents a new
// reservation, reconciliation or collector from assuming a settled home.
func LegacyPending(home string) error {
	identity, err := managerlock.New(home)
	if err != nil {
		return Refuse("environment_lease_unreadable", "state_unreadable")
	}
	home = identity.Home()
	path := filepath.Join(home, "state", "env-migration.json")
	if err := CheckRoute(home, path); err != nil {
		return Refuse("environment_lease_unreadable", "state_unreadable")
	}
	observed, err := stateread.Lstat(path)
	if err == nil && observed.Kind == stateread.KindAbsent {
		return nil
	}
	if err != nil || !observed.Info.Mode().IsRegular() {
		return Refuse("environment_lease_unreadable", "state_unreadable")
	}
	return Refuse("environment_lease_recovery_required", "pending_mutation")
}
