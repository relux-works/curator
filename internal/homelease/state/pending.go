package state

import (
	"errors"
	"os"
	"path/filepath"
)

// LegacyPending retains the existing env-migration journal without reading
// credential surfaces or attempting recovery. Only the migration owner may
// validate/reconcile its contents; the existence of its intent prevents a new
// reservation, reconciliation or collector from assuming a settled home.
func LegacyPending(home string) error {
	path := filepath.Join(home, "state", "env-migration.json")
	if err := CheckRoute(home, path); err != nil {
		return Refuse("environment_lease_unreadable", "state_unreadable")
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || !info.Mode().IsRegular() {
		return Refuse("environment_lease_unreadable", "state_unreadable")
	}
	return Refuse("environment_lease_recovery_required", "pending_mutation")
}
