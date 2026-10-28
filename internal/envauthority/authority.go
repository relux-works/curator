// Package envauthority checks the credential authority of managed homes before
// an incapable consumer can recover or publish manager-home state.
package envauthority

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/stateread"
)

// RequirePerHome preserves schema acceptance without downgrading node authority
// to the legacy per-home policy. It does not read credential material.
func RequirePerHome(marker *envmarker.Marker) error {
	if marker != nil && marker.CredentialSource != nil && marker.CredentialSource.Source != nil {
		return fmt.Errorf("%s: node credential source activation is not supported", envregistry.DiagCredentialInjectionUnavailable)
	}
	return nil
}

// Validate checks all managed markers, including orphaned profiles: generic
// transaction recovery is manager-home scoped and can publish outside the
// current command's selected profile. Only the fixed marker paths are read;
// payloads, native homes and credential stores are never walked.
// Callers revalidate under their mutation lock and after recovery. Absence is
// allowed for first provisioning; unreadable or unsafe authority fails closed.
func Validate(home string) error { return validate(home, nil) }

func validate(home string, intent *envunit.Unit) error {
	if _, err := ReadCompletedLegacyMigration(home); err != nil {
		return err
	}
	pending, err := PendingUnits(home)
	if err != nil {
		return err
	}
	lifecycle, err := inspectLifecycle(home, pending)
	if err != nil {
		return err
	}
	if err := validateLeaseReferences(home, lifecycle, pending); err != nil {
		return err
	}
	root := filepath.Join(home, "environments")
	state, err := stateread.Lstat(root)
	if err != nil {
		return err
	}
	if state.Kind == stateread.KindAbsent {
		return nil
	}
	owner := pathboundary.DefaultOwnerLookup()
	if err := pathboundary.ValidateLeafWithOwner(home, root, owner); err != nil {
		return boundary(err)
	}
	profiles, err := stateread.ReadDir(root)
	if err != nil {
		return err
	}
	for _, profile := range profiles.Entries {
		if pending.Artifacts[filepath.Join(root, profile.Name())] {
			continue
		}
		if !profile.IsDir() {
			return fmt.Errorf("%s: unknown environment namespace entry", envorigin.DiagInvalid)
		}
		// Project views have a separate composition/lease owner and directory
		// grammar; they are not profile × adapter managed-home namespaces.
		if profile.Name() == "project-views" {
			continue
		}
		dir := filepath.Join(root, profile.Name())
		if err := pathboundary.ValidateRouteWithOwner(root, dir, owner); err != nil {
			return boundary(err)
		}
		listing, err := stateread.ReadDir(dir)
		if err != nil {
			return err
		}
		known := map[string]bool{}
		for _, adapter := range envregistry.Registry {
			known[adapter.ID] = true
		}
		for _, child := range listing.Entries {
			if pending.Artifacts[filepath.Join(dir, child.Name())] {
				continue
			}
			if !known[child.Name()] || !child.IsDir() {
				return fmt.Errorf("%s: unknown managed adapter namespace", envorigin.DiagInvalid)
			}
		}
		for _, adapter := range envregistry.Registry {
			target := filepath.Join(dir, adapter.ID)
			if adapter.ID == envregistry.OpenCode {
				target = filepath.Join(target, "opencode")
			}
			markerPath := filepath.Join(target, envmarker.Name)
			if err := pathboundary.ValidateRouteWithOwner(root, markerPath, owner); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return boundary(err)
			}
			// Missing marker is still inspected for an orphan origin sidecar.
			// It does not skip authority validation for the adapter namespace.
			marker, err := envmarker.Read(target)
			if err != nil {
				return err
			}
			if err := RequirePerHome(marker); err != nil {
				return err
			}
			if !lifecycle.retained[target] {
				if _, isPending := pending.Homes[target]; !isPending {
					if err := validateRecorded(home, profile.Name(), adapter.ID, target, marker, intent); err != nil {
						return err
					}
				}
			}
			if marker != nil && !lifecycle.retained[target] {
				if _, pending := pending.Homes[target]; !pending {
					_, present, err := envorigin.ReadOwnedRecord(home, filepath.Join(home, "profiles", profile.Name(), "source.json"), 16<<20)
					if err != nil {
						return err
					}
					if !present {
						return fmt.Errorf("%s: unexplained orphan managed home", envorigin.DiagInvalid)
					}
				}
			}
		}
	}
	return nil
}

func boundary(err error) error {
	return fmt.Errorf("%s: credential authority marker boundary: %w", envregistry.DiagStoreUntrusted, err)
}
