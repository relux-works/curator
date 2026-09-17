package adapters

// Draft adapter destination boundaries (revision 1).
//
// Adapter skill directories are managed output. Destinations must never
// overwrite admitted context, runtime, build or dependency inputs —
// including through links, adapters, casing aliases or a changed parent
// directory — and the root-input allowlist never authorizes takeover of
// unmanaged destinations.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/staging"
)

// ProjectOutputRoots returns the managed output roots for a project: the
// generated .agents tree and every adapter skill directory. Native-discovery
// agents share .agents/skills and need no extra root. Callers add manager
// runtime stores, build caches, transaction staging and the snapshot store.
func ProjectOutputRoots(projectRoot string) []string {
	roots := []string{filepath.Join(projectRoot, ".agents")}
	for _, rel := range AgentPaths {
		roots = append(roots, filepath.Join(projectRoot, filepath.FromSlash(rel)))
	}
	return roots
}

// ValidateDestinations refuses adapter plans whose live paths would
// overwrite admitted inputs. Every target is compared in both directions —
// a destination inside an admitted tree and an admitted input inside a
// destination both fail — using physical identity, so links, casing
// aliases and a changed parent directory all count. Entry locations and
// live link destinations are both checked: a managed entry must neither
// sit inside admitted inputs nor point into them. Proven overwrites refuse
// as source_output_overlap; inspection failures refuse as
// boundary_identity_unreadable, never as a proven overlap.
func ValidateDestinations(plan staging.Plan, admitted []string) error {
	admittedCanonical := make([]string, 0, len(admitted))
	for _, input := range admitted {
		canonical, err := staging.Canonicalize(input)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: admitted input %s: %v", input, err)
		}
		admittedCanonical = append(admittedCanonical, canonical)
	}
	for _, target := range plan.Targets {
		full, err := staging.Canonicalize(target.LivePath)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
		parent, err := staging.Canonicalize(filepath.Dir(target.LivePath))
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
		location := filepath.Join(parent, filepath.Base(target.LivePath))
		for _, candidate := range []string{full, location} {
			for _, input := range admittedCanonical {
				within, err := staging.Within(input, candidate)
				if err != nil {
					return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
				}
				if within {
					return fmt.Errorf("source_output_overlap: %s overwrites admitted input", target.LivePath)
				}
				reverse, err := staging.Within(candidate, input)
				if err != nil {
					return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
				}
				if reverse {
					return fmt.Errorf("source_output_overlap: %s overwrites admitted input", target.LivePath)
				}
			}
		}
		if info, err := os.Lstat(target.LivePath); err == nil && info.Mode()&os.ModeSymlink != 0 { // #nosec G304 -- planned adapter destination under validation
			destination, err := os.Readlink(target.LivePath) // #nosec G304 -- planned adapter destination under validation
			if err != nil {
				return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
			}
			linkTarget := destination
			if !filepath.IsAbs(linkTarget) {
				linkTarget = filepath.Join(filepath.Dir(target.LivePath), linkTarget)
			}
			linkCanonical, err := staging.Canonicalize(linkTarget)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
			}
			for _, input := range admittedCanonical {
				within, err := staging.Within(input, linkCanonical)
				if err != nil {
					return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
				}
				if within {
					return fmt.Errorf("source_output_overlap: %s overwrites admitted input", target.LivePath)
				}
			}
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
	}
	return nil
}
