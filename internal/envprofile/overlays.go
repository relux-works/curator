// Overlay composition input (environments §6): machine overlay declarations
// join the closure beside the root and resolve jointly with it.
package envprofile

import (
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
)

// resolveOverlays builds the resolution overlays for profile from the
// machine policy (environments §6, §12.1): every declaration in
// overlays.<profile> becomes one closure seed beside the root, with the
// machine-assigned weight (default overlay_default_weight). A forbidding
// composition policy empties every list, so the root resolves alone
// (environments §12.2).
//
// A git overlay resolves its requirement to a commit exactly like a root
// install and reads the overlay name from the manifest at that commit; a
// path overlay snapshots the directory immutably under its state hash and
// reads the name from the snapshot manifest. An unreadable or
// uninstallable overlay source fails with that source's section 1.1
// diagnostic; a declaration that repeats a name already in the closure by
// another declaration fails resolution with
// environment_composition_invalid.
func resolveOverlays(home string, manager *gitManager, profile string, policy Policy) ([]contextresolve.Overlay, error) {
	decls := policy.EffectiveOverlays(profile)
	if len(decls) == 0 {
		return nil, nil
	}
	overlays := make([]contextresolve.Overlay, 0, len(decls))
	for index, decl := range decls {
		overlay, err := resolveOverlay(home, manager, decl)
		if err != nil {
			return nil, fmt.Errorf("overlay %d of profile %q: %v", index, profile, err)
		}
		overlays = append(overlays, overlay)
	}
	return overlays, nil
}

func resolveOverlay(home string, manager *gitManager, decl OverlaySpec) (contextresolve.Overlay, error) {
	weight := decl.Weight
	if isPathOperand(decl.Source) {
		if decl.Range != "" || decl.Tag != "" || decl.Revision != "" || decl.Directory != "" {
			return contextresolve.Overlay{}, fmt.Errorf("%s: a path overlay carries no range, tag, revision, or directory", DiagSourceInvalid)
		}
		manifest, err := contextpkg.LoadManifest(decl.Source)
		if err != nil {
			return contextresolve.Overlay{}, pathManifestDiag(decl.Source, err)
		}
		state, err := stateForPath(home, manifest.Name, decl.Source)
		if err != nil {
			return contextresolve.Overlay{}, err
		}
		return contextresolve.Overlay{
			Name: manifest.Name, Source: decl.Source,
			Weight: weight, State: state,
		}, nil
	}
	identity, err := canonicalGit(decl.Source)
	if err != nil {
		return contextresolve.Overlay{}, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	manager.recordRaw(identity, decl.Source)
	if err := manager.fetch(identity); err != nil {
		return contextresolve.Overlay{}, err
	}
	dir := manager.repoDir(identity)
	commit, _, err := manager.commitFor(dir, Requirement{Range: decl.Range, Tag: decl.Tag, Revision: decl.Revision})
	if err != nil {
		return contextresolve.Overlay{}, err
	}
	manifest, err := manager.contextManifestAt(dir, commit, decl.Directory)
	if err != nil {
		return contextresolve.Overlay{}, err
	}
	return contextresolve.Overlay{
		Name: manifest.Name, Source: identity,
		Range: decl.Range, Tag: decl.Tag, Revision: decl.Revision,
		Directory: decl.Directory, Weight: weight,
	}, nil
}

// overlayInputDefaults carries the overlay weight default into one
// resolution input.
func overlayInputDefaults(input *contextresolve.Input, policy Policy) {
	input.OverlayDefaultWeight = policy.OverlayDefaultWeight
}

// resolutionWarnings renders successful-resolution findings — the rule-3
// weight-conflict downgrade (environments §6: the root has the final word
// and the disagreement is reported as a warning under the same diagnostic)
// — as install and update warnings.
func resolutionWarnings(result *contextresolve.Result) []string {
	var warnings []string
	for _, warning := range result.Warnings {
		parts := make([]string, 0, len(warning.Requirers))
		for _, requirer := range warning.Requirers {
			parts = append(parts, requirer.Requirer+" declares weight")
		}
		warnings = append(warnings, warning.Diagnostic+": "+warning.Name+" ("+strings.Join(parts, ", ")+")")
	}
	return warnings
}
