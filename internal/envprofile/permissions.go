package envprofile

import (
	"fmt"

	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// PermissionsForProfile resolves the fragment permission policy (§10.2).
// A global lock forces native even for a profile absent from the system map.
func PermissionsForProfile(machine envregistry.MachineConfig, profile string) envfragment.Permissions {
	if machine.PermissionsLocked {
		return envfragment.Permissions{Mode: "native", Locked: true, Source: "global"}
	}
	if mode, configured := machine.Permissions[profile]; configured {
		return envfragment.Permissions{Mode: mode, Source: "profile"}
	}
	return envfragment.Permissions{Mode: "native", Source: "default"}
}

// InstalledPermissions reports fragment permissions without provisioning homes
// or requiring a runtime adapter. Failed inventory reads remain errors.
func InstalledPermissions(home string, machine envregistry.MachineConfig) (map[string]envfragment.Permissions, error) {
	// Some directory enumeration implementations return an empty listing for a
	// regular file. Establish the inventory's type before accepting emptiness.
	inventory := ProfilesDir(home)
	metadata, err := stateread.Stat(inventory)
	if err != nil {
		return nil, err
	}
	if metadata.Kind == stateread.KindPresent && !metadata.Info.IsDir() {
		return nil, stateread.UnusableError(inventory, fmt.Errorf("profile inventory must be a directory"))
	}
	profiles, err := installedProfiles(home)
	if err != nil {
		return nil, err
	}
	result := make(map[string]envfragment.Permissions, len(profiles))
	for _, profile := range profiles {
		result[profile] = PermissionsForProfile(machine, profile)
	}
	return result, nil
}
