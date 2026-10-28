package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/stateread"
)

// AdoptOriginRequest is a distinct, explicit authority action. It has no
// environment/account defaults, repair/takeover flag or inferred profile alias.
type AdoptOriginRequest struct {
	Home        string
	Attestation envorigin.AdoptRequest
	Machine     envregistry.MachineConfig
}

// AdoptOrigin publishes only missing historical authority. It never replaces
// an existing differing record or rerenders a surface, repairs a credential
// link, refreshes a seed or touches native session/credential contents.
func AdoptOrigin(request AdoptOriginRequest) error {
	attestation := request.Attestation
	if attestation.Binding.Profile == "" || attestation.Adapter == "" {
		return fmt.Errorf("%s: explicit installed profile and adapter required", envorigin.DiagInvalid)
	}
	req := ResolveRequest{Home: request.Home, Profile: attestation.Binding.Profile, EnvID: attestation.Adapter, Machine: request.Machine}
	adapter, err := envregistry.ByID(req.EnvID)
	if err != nil {
		return err
	}
	if attestation.Binding.ManagedTarget != ManagedHomeDir(req.Home, req.Profile, req.EnvID) {
		return fmt.Errorf("%s: attested managed target differs from installed namespace", envorigin.DiagInvalid)
	}
	// The same complete inspection is repeated under the manager lock, before
	// any recovery may publish and before the adoption journal is prepared.
	var marker *envmarker.Marker
	var markerRaw, originRaw []byte
	var changed bool
	validate := func() error {
		if _, _, hash, err := loadResolveInputs(req.Home, req.Profile); err != nil {
			return err
		} else if attestation.Binding.LockSHA256 != "sha256:"+strings.TrimPrefix(hash, "sha256:") {
			return fmt.Errorf("%s: attested lock differs from current installed lock", envorigin.DiagInvalid)
		}
		if err := validateMarkerBoundary(req.Home, req.Profile, req.EnvID, nil); err != nil {
			return err
		}
		var err error
		marker, err = envmarker.Read(attestation.Binding.ManagedTarget)
		if err != nil {
			return err
		}
		if marker == nil || marker.Mode != envmarker.ModeManagedHome || marker.Profile.Name != req.Profile || marker.Profile.LockSHA256 != strings.TrimPrefix(attestation.Binding.LockSHA256, "sha256:") {
			return fmt.Errorf("%s: adoption needs the exact legacy managed marker", envorigin.DiagInvalid)
		}
		if err := requirePerHomeCredentialSource(marker); err != nil {
			return err
		}
		markerRaw, err = os.ReadFile(filepath.Join(attestation.Binding.ManagedTarget, envmarker.Name))
		if err != nil {
			return err
		}
		existing, _, err := envorigin.ReadPrivateAuthority(req.Home, filepath.Join(attestation.Binding.ManagedTarget, envorigin.Name), envorigin.MaxRecordBytes)
		if err != nil {
			return err
		}
		originRaw, changed, err = envorigin.Attest(req.Home, attestation, existing)
		if err != nil {
			return err
		}
		record, err := envorigin.Parse(originRaw, attestation.Binding, req.EnvID, req.Home)
		if err != nil {
			return err
		}
		req.origin = &originObservation{record: record, raw: originRaw}
		if err := req.preflightOriginCredentials(adapter); err != nil {
			return err
		}
		policy, err := req.resolveCredentialPolicyObserved(adapter, "unknown")
		if err != nil {
			return err
		}
		for path, target := range policy.links {
			full := filepath.Join(attestation.Binding.ManagedTarget, filepath.FromSlash(path))
			state, err := stateread.Readlink(full)
			if err != nil {
				return err
			}
			if state.Kind != stateread.KindPresent {
				return fmt.Errorf("%s: adoption cannot repair a missing credential link", envregistry.DiagCredentialConflict)
			}
			if adapter.ID == envregistry.Muse {
				if auth := req.inspectMuseAuth(full, target); auth.err != nil {
					return auth.err
				}
			}
		}
		return nil
	}
	if err := validate(); err != nil {
		return err
	}
	intent, err := req.homeUnit(newTransactionID(), &homePlan{adapter: adapter, homeDir: attestation.Binding.ManagedTarget}, marker, markerRaw)
	if err != nil {
		return err
	}
	intent.Action = envunit.Adopt
	intent.Homes[0].Before.Kind = "legacy"
	intent.Homes[0].Before.Origin = nil
	op, err := beginAdoptOperation(req.Home, intent, validate)
	if err != nil {
		return err
	}
	defer op.close()
	if err := validate(); err != nil {
		return op.recoveryRefusal(err)
	}
	if !changed {
		return nil
	}
	pub, err := newHomePublication(op)
	if err != nil {
		return err
	}
	defer pub.close()
	if err := pub.file(filepath.Join(attestation.Binding.ManagedTarget, envorigin.Name), "origin", originRaw, 0600, true); err != nil {
		return err
	}
	plan := &homePlan{adapter: adapter, homeDir: attestation.Binding.ManagedTarget}
	unit, err := req.homeUnit(newTransactionID(), plan, marker, markerRaw)
	if err != nil {
		return err
	}
	unit.Action = envunit.Adopt
	unit.Homes[0].Before.Kind = "legacy"
	unit.Homes[0].Before.Origin = nil
	// Marker bytes and every existing entry remain exact adoption preimages.
	if !bytes.Equal(unit.Homes[0].Before.Marker, markerRaw) {
		return fmt.Errorf("%s: marker changed during adoption", envorigin.DiagInvalid)
	}
	pub.unit = unit
	return pub.publish(&req, plan, marker, markerRaw)
}
