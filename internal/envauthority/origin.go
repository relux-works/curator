package envauthority

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/stateread"
)

// ValidateForUnit permits only the exact read-side legacy adoption prefix.
// This permits lock acquisition/planning, never generic writes or recovery.
// Prepare must bind and validate its complete actual unit before publication.
func ValidateForUnit(manager string, unit *envunit.Unit) error { return validate(manager, unit) }

func validateRecorded(manager, profile, adapter, target string, marker *envmarker.Marker, intent *envunit.Unit) error {
	raw, present, err := envorigin.ReadPrivateAuthority(manager, filepath.Join(target, envorigin.Name), envorigin.MaxRecordBytes)
	if err != nil {
		return err
	}
	if marker == nil {
		if present {
			return fmt.Errorf("%s: origin sidecar without managed marker", envorigin.DiagInvalid)
		}
		return nil
	}
	lockPath := filepath.Join(manager, "profiles", profile, "lock.json")
	if err := pathboundary.ValidateRouteWithOwner(manager, lockPath, pathboundary.DefaultOwnerLookup()); err != nil {
		return boundary(err)
	}
	lock, hash, err := contextlock.ReadWith(lockPath, stateread.ReadRegularFile)
	if err != nil || lock == nil {
		return fmt.Errorf("%s: recorded home installed lock unreadable: %v", envorigin.DiagInvalid, err)
	}
	binding := envorigin.Binding{Profile: profile, LockSHA256: "sha256:" + strings.TrimPrefix(hash, "sha256:"), ManagedTarget: target}
	if marker.Mode != envmarker.ModeManagedHome || marker.Profile.Name != profile || strings.TrimPrefix(marker.Profile.LockSHA256, "sha256:") != strings.TrimPrefix(hash, "sha256:") {
		return fmt.Errorf("%s: marker/installed lock binding mismatch", envorigin.DiagInvalid)
	}
	if !present {
		if intent != nil && intent.Manager.Path == manager && intent.Action == envunit.Adopt && len(intent.Homes) == 1 && len(intent.Profiles) == 1 {
			h, p := intent.Homes[0], intent.Profiles[0]
			canonical, e := lock.Canonical()
			if e != nil {
				return e
			}
			mark, e := stateread.ReadRegularFile(filepath.Join(target, envmarker.Name))
			if e != nil {
				return e
			}
			if h.Profile == profile && h.Adapter == adapter && h.Target == target && h.Root == envorigin.ManagedRoot(manager, profile, adapter) && h.Mode == "managed-home" && h.Before.Kind == "legacy" && len(h.Before.Origin) == 0 && bytes.Equal(h.Before.Marker, mark.Bytes) && h.After.Kind == "recorded" && bytes.Equal(h.After.Marker, mark.Bytes) && p.Name == profile && bytes.Equal(p.OldLock.Bytes, canonical) && bytes.Equal(p.NewLock.Bytes, canonical) {
				if _, e := envorigin.Parse(h.After.Origin, binding, adapter, manager); e != nil {
					return e
				}
				return nil
			}
		}
		return fmt.Errorf("%s: legacy managed home requires explicit env adopt attestation", envorigin.DiagRequired)
	}
	_, err = envorigin.Parse(raw, binding, adapter, manager)
	return err
}
