package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/hashing"
)

// OwnsNativeSkill verifies a profile-owned native entry without adopting it.
// requested, when non-nil, is the legacy global closure's resolved identity.
// Environments §§1.4 and 9.4 require that identity to unify with the profile;
// content equality alone is not agreement on a source, pin and directory.
// Callers run this under the manager-home mutation lock before staging mirrors.
func OwnsNativeSkill(home, adapterRoot, name string, requested *contextlock.Member) (bool, error) {
	if filepath.Base(adapterRoot) != "skills" {
		return false, nil
	}
	native := filepath.Dir(adapterRoot)
	if err := checkManagedPrivateTarget(native, envmarker.Name); err != nil {
		return false, err
	}
	marker, err := envmarker.Read(native)
	if err != nil || marker == nil {
		return false, err
	}
	if marker.Mode == envmarker.ModeManagedHome {
		return false, nil
	}
	path := "skills/" + name
	surface := marker.Surfaces[envmarker.SurfaceSkills]
	recorded := false
	for _, entry := range surface.Paths {
		recorded = recorded || entry == path
	}
	if !recorded {
		return false, nil
	}
	lock, hash, err := readLock(home, marker.Profile.Name)
	if err != nil {
		return false, err
	}
	if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, "sha256:") {
		return false, fmt.Errorf("%s: native skills ownership lock differs for profile %q", DiagSurfaceDrift, marker.Profile.Name)
	}
	var member *contextlock.Member
	for i := range lock.Members {
		candidate := &lock.Members[i]
		if candidate.Kind == contextlock.KindSkill && candidate.Name == name {
			member = candidate
		}
	}
	if member == nil {
		return false, fmt.Errorf("%s: native skill %q has no profile lock member", DiagSurfaceDrift, name)
	}
	if requested != nil && (member.Source != requested.Source || member.Commit != requested.Commit || member.Directory != requested.Directory) {
		return false, fmt.Errorf("context_range_conflict: skill %q: profile %q requires %s@%s directory %q; global Skillfile requires %s@%s directory %q", name, marker.Profile.Name, member.Source, member.Commit, member.Directory, requested.Source, requested.Commit, requested.Directory)
	}
	skills, err := skillsOf(home, newGitManager(home), lock)
	if err != nil {
		return false, err
	}
	identities := map[string][]byte{}
	var selected skillOf
	for _, skill := range skills {
		for _, entry := range surface.Paths {
			if entry == "skills/"+skill.name {
				identities[entry] = skillSurfaceIdentity(skill)
			}
		}
		if skill.name == name {
			selected = skill
		}
	}
	digest, err := contextmaterialize.SurfaceHashWithVersion(identities, lock.ContentHashVersion())
	if err != nil {
		return false, err
	}
	if len(identities) != len(surface.Paths) || digest != surface.ContentSHA256 {
		return false, fmt.Errorf("%s: native skills ownership content differs for profile %q", DiagSurfaceDrift, marker.Profile.Name)
	}
	if err := verifySkillProjection(home, selected); err != nil {
		return false, err
	}
	target, err := managedPath(native, path, false)
	if err != nil {
		return false, err
	}
	info, err := os.Lstat(target)
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(target)
		if err != nil {
			return false, err
		}
		if link == selected.root {
			return true, nil
		}
	} else if info.IsDir() {
		// Copied mode and recorded symlink fallback carry the same projected
		// tree identity. A marker alone never authorizes a drifted directory.
		copied := marker.Mode == envmarker.ModeCopied
		if surface.Copies != nil {
			for _, copy := range *surface.Copies {
				copied = copied || (copy.Path == path && copy.Reason == envmarker.ReasonSymlinkFallback)
			}
		}
		if copied {
			hash, err := hashing.ContentSHA256WithVersion(target, map[string]bool{}, lock.ContentHashVersion())
			if err != nil {
				return false, err
			}
			if hash == selected.hash {
				return true, nil
			}
		}
	}
	return false, fmt.Errorf("%s: native skill %q differs from its recorded profile owner", DiagSurfaceDrift, name)
}
