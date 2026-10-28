package envauthority

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
)

// AdmitEffects is a read-side planning guard before even hashing/staging live
// target preimages. It does not grant a pending exception: Prepare must still
// bind and validate the full actual unit, phase and durable target inventory.
func AdmitEffects(manager string, unit *envunit.Unit, paths []string) error {
	pending, err := PendingUnits(manager)
	if err != nil {
		return err
	}
	for id, effects := range pending.Effects {
		// During commit this is the same already verified durable unit. At
		// prepare an existing ID is independently refused before publication.
		if unit != nil && unit.ID == id {
			continue
		}
		for _, path := range paths {
			for _, effect := range effects {
				intersects := overlaps(path, effect.Path)
				if effect.DirectoryOnly {
					// A parent metadata action does not own all siblings below it;
					// replacing that parent or an ancestor still intersects it.
					intersects = path == effect.Path || strings.HasPrefix(effect.Path, filepath.Clean(path)+string(filepath.Separator))
				}
				if intersects {
					return fmt.Errorf("%s: effect intersects a different pending unit", envorigin.DiagInvalid)
				}
			}
		}
	}
	root := filepath.Join(manager, "environments")
	for _, path := range paths {
		if path == root || strings.HasPrefix(path, root+string(filepath.Separator)) {
			rel, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			parts := strings.Split(rel, string(filepath.Separator))
			if parts[0] == "project-views" {
				continue
			} // separate composition/lease grammar
			if unit == nil {
				return fmt.Errorf("%s: managed home effect requires a complete environment unit", envorigin.DiagInvalid)
			}
			bound := false
			for _, home := range unit.Homes {
				if overlaps(home.Root, path) {
					bound = true
				}
			}
			if !bound {
				return fmt.Errorf("%s: home effect is outside the unit's exact home inventory", envorigin.DiagInvalid)
			}
		}
		profiles := filepath.Join(manager, "profiles")
		rel, e := filepath.Rel(profiles, path)
		if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) != 1 && !(len(parts) == 2 && (parts[1] == "lock.json" || parts[1] == "source.json")) {
			continue
		}
		profile := parts[0]
		if profile == "." || profile == "current" || profile == "scoped" {
			continue
		}
		// Any installed authority replacement/deletion includes every already
		// materialized adapter, irrespective of current/scoped/default selection.
		for _, adapter := range envregistry.Registry {
			target := envorigin.ManagedTarget(manager, profile, adapter.ID)
			marker, err := envmarker.Read(target)
			if err != nil {
				return err
			}
			if marker == nil {
				continue
			}
			bound := false
			if unit != nil {
				for _, home := range unit.Homes {
					if home.Profile == profile && home.Adapter == adapter.ID && home.Target == target && home.Root == envorigin.ManagedRoot(manager, profile, adapter.ID) {
						bound = true
					}
				}
			}
			if !bound {
				return fmt.Errorf("%s: installed authority publication omits a materialized home", envorigin.DiagInvalid)
			}
		}
	}
	return nil
}
