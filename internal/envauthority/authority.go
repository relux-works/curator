// Package envauthority checks the credential authority of managed homes before
// an incapable consumer can recover or publish manager-home state.
package envauthority

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
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
func Validate(home string) error {
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
		if !profile.IsDir() && profile.Type().IsRegular() {
			continue
		}
		dir := filepath.Join(root, profile.Name())
		if err := pathboundary.ValidateRouteWithOwner(root, dir, owner); err != nil {
			return boundary(err)
		}
		for _, adapter := range envregistry.Registry {
			target := filepath.Join(dir, adapter.ID)
			if adapter.ID == envregistry.OpenCode {
				target = filepath.Join(target, "opencode")
			}
			markerPath := filepath.Join(target, envmarker.Name)
			if err := pathboundary.ValidateRouteWithOwner(root, markerPath, owner); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return boundary(err)
			}
			marker, err := envmarker.Read(target)
			if err != nil {
				return err
			}
			if err := RequirePerHome(marker); err != nil {
				return err
			}
		}
	}
	return nil
}

func boundary(err error) error {
	return fmt.Errorf("%s: credential authority marker boundary: %w", envregistry.DiagStoreUntrusted, err)
}
