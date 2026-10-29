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
		if intent != nil {
			// The selected legacy home is independently attested. A different legacy
			// home remains quarantined and may be admitted only for this adopt action.
			focus := intent.Homes[0]
			if target != focus.Target && !pathIntersects(target, focus.Root) {
				return nil
			}
			canonical, e := lock.Canonical()
			if e != nil {
				return e
			}
			mark, e := stateread.ReadRegularFile(filepath.Join(target, envmarker.Name))
			if e != nil {
				return e
			}
			p := intent.Profiles[0]
			if focus.Profile == profile && focus.Adapter == adapter && bytes.Equal(focus.Before.Marker, mark.Bytes) && bytes.Equal(p.OldLock.Bytes, canonical) {
				return nil
			}
		}
		return fmt.Errorf("%s: legacy managed home requires explicit env adopt attestation", envorigin.DiagRequired)
	}
	_, err = envorigin.Parse(raw, binding, adapter, manager)
	return err
}

// adoptionQuarantine validates the action before permitting any disjoint legacy
// neighbor. It conveys no authority for native, lock, settings or shared writes.
func adoptionQuarantine(manager string, intent *envunit.Unit, pending *Pending) (*envunit.Unit, error) {
	if intent == nil {
		for _, u := range pending.Units {
			if u.Action == envunit.Adopt {
				if intent != nil {
					return nil, fmt.Errorf("%s: multiple adoption journals", envorigin.DiagInvalid)
				}
				intent = u
			}
		}
	}
	if intent == nil || intent.Action != envunit.Adopt {
		return nil, nil
	}
	refuse := func() (*envunit.Unit, error) {
		return nil, fmt.Errorf("%s: adoption quarantine binding/effects differ", envorigin.DiagInvalid)
	}
	if intent.Manager.Path != manager || len(intent.Homes) != 1 || len(intent.Profiles) != 1 || len(intent.Native) != 0 || intent.LedgerPublication != nil || pending.OtherJournals != 0 {
		return refuse()
	}
	for id, u := range pending.Units {
		if id != intent.ID && (u.Action != envunit.Adopt || len(u.Homes) != 1 || u.Homes[0].Target != intent.Homes[0].Target) {
			return refuse()
		}
	}
	h, p := intent.Homes[0], intent.Profiles[0]
	if h.Mode != "managed-home" || h.Profile != p.Name || h.Root != envorigin.ManagedRoot(manager, h.Profile, h.Adapter) || h.Target != envorigin.ManagedTarget(manager, h.Profile, h.Adapter) || h.Before.Kind != "legacy" || len(h.Before.Origin) != 0 || h.After.Kind != "recorded" || !bytes.Equal(h.Before.Marker, h.After.Marker) || !bytes.Equal(p.OldLock.Bytes, p.NewLock.Bytes) || p.OldLock.Digest != p.NewLock.Digest {
		return refuse()
	}
	for _, t := range intent.Targets {
		if t.Class == "parent" || t.Class == "home-directory" {
			if t.Before != t.After {
				return refuse()
			}
		} else if t.Class != "origin" || t.Path != filepath.Join(h.Target, envorigin.Name) {
			return refuse()
		}
	}
	lock, hash, err := contextlock.ReadWith(filepath.Join(manager, "profiles", p.Name, "lock.json"), stateread.ReadRegularFile)
	if err != nil || lock == nil {
		return refuse()
	}
	canonical, err := lock.Canonical()
	if err != nil || !bytes.Equal(canonical, p.OldLock.Bytes) {
		return refuse()
	}
	mark, err := stateread.ReadRegularFile(filepath.Join(h.Target, envmarker.Name))
	if err != nil || !bytes.Equal(mark.Bytes, h.Before.Marker) {
		return refuse()
	}
	binding := envorigin.Binding{Profile: p.Name, LockSHA256: "sha256:" + strings.TrimPrefix(hash, "sha256:"), ManagedTarget: h.Target}
	if _, err := envorigin.Parse(h.After.Origin, binding, h.Adapter, manager); err != nil {
		return nil, err
	}
	return intent, nil
}

func pathIntersects(a, b string) bool {
	return a == b || strings.HasPrefix(a, b+string(filepath.Separator)) || strings.HasPrefix(b, a+string(filepath.Separator))
}
