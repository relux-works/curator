package envprofile

import (
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/envauthority"
)

// RemoveOptions binds the exact native target set for an explicit purge.
// Without Purge, both native and managed homes remain on disk.
type RemoveOptions struct {
	Purge       bool
	NativeHomes NativeHomeBindings
	// ManagedOnly purges the enumerated managed identity without native targets.
	ManagedOnly bool
}

// RemoveWithOptions applies the ordinary scope, overlay and authority gates,
// then purges only explicitly bound native targets and this profile's managed
// homes. Missing bindings fail before acquiring a mutation lock or recovering.
func RemoveWithOptions(home, name string, opts RemoveOptions) error {
	if err := refuseBuiltinProfileAlias(home, name); err != nil {
		return err
	}
	if !validProfileName(name) {
		return fmt.Errorf("%s: profile name %q is not a portable identifier", DiagSourceInvalid, name)
	}
	if err := envauthority.Validate(home); err != nil {
		return err
	}
	if opts.ManagedOnly && !opts.Purge {
		return fmt.Errorf("managed-only requires purge")
	}
	if opts.ManagedOnly && len(opts.NativeHomes) != 0 {
		return fmt.Errorf("managed-only does not accept native bindings")
	}
	if opts.Purge && !opts.ManagedOnly {
		ids := make([]string, 0, len(Adapters))
		supported := map[string]bool{}
		for _, adapter := range Adapters {
			ids = append(ids, adapter.ID)
			supported[adapter.ID] = true
		}
		for _, id := range bindingKeys(opts.NativeHomes) {
			if !supported[id] {
				return fmt.Errorf("%s: %s has no native purge surface", DiagNativeHomeBindingRequired, id)
			}
		}
		if _, err := nativeBindingResolver(home, opts.NativeHomes, ids); err != nil {
			return err
		}
	}
	op, err := beginOperationGuarded(home, nil, homeLeaseProfileTargets(home, name, ""))
	if err != nil {
		return err
	}
	defer func() { _ = op.close() }()
	if _, err := readSource(home, name); err != nil {
		return err
	}
	inUse, err := InUseByAnyScope(home, name)
	if err != nil {
		return err
	}
	if inUse {
		return fmt.Errorf("%s: profile %q is current in a scope", DiagInUse, name)
	}
	// A profile that is an overlay member of another installed profile's
	// lock stays until that profile stops declaring it (environments
	// §9.2): removal would strand the overlay declaration.
	owner, ok, ownerErr := overlayOwner(home, name)
	if ownerErr != nil {
		return ownerErr
	}
	if ok {
		return fmt.Errorf("%s: profile %q is named as an overlay of installed profile %q", DiagInUse, name, owner)
	}
	if opts.Purge {
		if err := preflightProfileRemoval(home, name); err != nil {
			return fmt.Errorf("%s: profile target: %w", DiagProfilePurgePreflightRefused, err)
		}
	}
	if opts.Purge {
		var purgeErr error
		if opts.ManagedOnly {
			purgeErr = purgeManagedHomes(home, name)
		} else {
			purgeErr = purgeHomes(home, name, opts.NativeHomes)
		}
		if purgeErr != nil {
			return purgeErr
		}
	}
	if err := os.RemoveAll(ProfileDir(home, name)); err != nil {
		if opts.Purge {
			return purgeIncomplete("profile directory", err)
		}
		return err
	}
	return nil
}
