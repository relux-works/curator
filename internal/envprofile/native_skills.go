package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/stateread"
)

// inPlaceSkillsDir preserves opencode staging; only the single-home native
// adapters participate in the native skills ownership contract.
func inPlaceSkillsDir(id string) string {
	switch id {
	case envregistry.ClaudeCode, envregistry.CodexCLI, envregistry.Pi:
		adapter, _ := envregistry.ByID(id)
		return adapter.SkillsDir
	default:
		return profileSkillsSurfaceDir
	}
}

// inPlaceSkillsLedger checks the parent before reading ownership state. A
// failed ledger read is never evidence that an entry is unowned.
func inPlaceSkillsLedger(native string, adapter Adapter, skills []skillOf, recorded map[string]bool) (map[string]bool, error) {
	needed := len(skills) > 0
	for path := range recorded {
		if strings.HasPrefix(path, inPlaceSkillsDir(adapter.ID)+"/") {
			needed = true
		}
	}
	if !needed {
		return map[string]bool{}, nil
	}
	if inPlaceSkillsDir(adapter.ID) == profileSkillsSurfaceDir {
		return map[string]bool{}, nil
	}
	path, err := managedPath(native, filepath.ToSlash(filepath.Join(inPlaceSkillsDir(adapter.ID), ".csk-managed.json")), false)
	if err != nil {
		return nil, err
	}
	return readSkillsLedger(filepath.Dir(path))
}

func ledgerSkillPath(adapter Adapter, path string, ledger map[string]bool) bool {
	dir := inPlaceSkillsDir(adapter.ID)
	return dir != profileSkillsSurfaceDir && ledger[strings.TrimPrefix(path, dir+"/")] && strings.HasPrefix(path, dir+"/")
}

func inPlaceLinkInStore(full, target, store string) bool {
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(full), target)
	}
	return sameStoreTree(target, store)
}

// nativeSkillsState is read-only and checks only the effective current
// profile. It never inventories or adopts names outside the profile.
func nativeSkillsState(home, profile, id string, nativeHomeOf func(string) (string, error)) (*SurfaceState, error) {
	if inPlaceSkillsDir(id) == profileSkillsSurfaceDir {
		return nil, nil
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return nil, err
	}
	current := scoped["env:"+id]
	if current == "" {
		current, err = Current(home)
		if err != nil {
			return nil, err
		}
	}
	if current != profile {
		return nil, nil
	}
	lock, hash, err := readLock(home, profile)
	if err != nil {
		return nil, err
	}
	skills, err := skillsOf(home, newGitManager(home), lock)
	if err != nil {
		return nil, err
	}
	adapter, _ := adapterByID(id)
	native, err := resolveNativeHome(adapter, nativeHomeOf)
	if err != nil {
		return nil, err
	}
	root, err := stateread.Stat(native)
	if err != nil {
		return nil, err
	}
	row := &SurfaceState{Key: envmarker.SurfaceSkills, State: "current", Detail: "current", Paths: []string{}}
	if root.Kind == stateread.KindAbsent {
		if len(skills) == 0 {
			return nil, nil
		}
		row.State = DiagSurfaceMissing
		row.Detail = "native skills home missing"
		return row, nil
	}
	if err := checkManagedPrivateTarget(native, envmarker.Name); err != nil {
		return nil, err
	}
	marker, err := envmarker.Read(native)
	if err != nil {
		return nil, err
	}
	ledger, err := inPlaceSkillsLedger(native, adapter, skills, map[string]bool{})
	if err != nil {
		return nil, err
	}
	expected := map[string][]byte{}
	var reasons []string
	for _, skill := range skills {
		path := filepath.ToSlash(filepath.Join(inPlaceSkillsDir(id), skill.name))
		row.Paths = append(row.Paths, path)
		if ledger[skill.name] {
			reasons = append(reasons, path+" belongs to legacy global scope (adapter ledger); left in place")
			continue
		}
		expected[path] = []byte(skill.hash)
		if marker == nil {
			reasons = append(reasons, path+" is not recorded")
			continue
		}
		full, err := managedPath(native, path, false)
		if err != nil {
			return nil, err
		}
		link, err := stateread.Readlink(full)
		if err != nil || link.Kind != stateread.KindPresent || link.Target != skill.root {
			// Retain the existing recorded symlink-unavailable fallback contract.
			fallback := false
			surface := marker.Surfaces[envmarker.SurfaceSkills]
			if surface.Copies != nil {
				for _, copy := range *surface.Copies {
					if copy.Path == path && copy.Reason == envmarker.ReasonSymlinkFallback {
						fallback = true
					}
				}
			}
			info, statErr := os.Lstat(full)
			if fallback && statErr == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
				digest, err := hashing.ContentSHA256WithVersion(full, map[string]bool{}, lock.ContentHashVersion())
				if err == nil && digest == skill.hash {
					continue
				}
			}
			reasons = append(reasons, path+" native skill link drifted or missing")
		}
	}
	if marker != nil {
		surface, ok := marker.Surfaces[envmarker.SurfaceSkills]
		if len(expected) > 0 {
			digest, _ := contextmaterialize.SurfaceHashWithVersion(expected, lock.ContentHashVersion())
			if !ok || digest != surface.ContentSHA256 || marker.Profile.Name != profile || marker.Profile.LockSHA256 != strings.TrimPrefix(hash, "sha256:") {
				reasons = append(reasons, "native skills record differs from current profile")
			}
		}
		paths := make([]string, 0, len(expected))
		for path := range expected {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		if strings.Join(paths, "\n") != strings.Join(surface.Paths, "\n") {
			reasons = append(reasons, "native skills paths differ from current profile (old staging must converge)")
		}
	}
	if len(skills) == 0 && marker == nil {
		return nil, nil
	}
	if len(reasons) > 0 {
		row.State = DiagSurfaceDrift
		row.Detail = strings.Join(reasons, "; ")
	}
	return row, nil
}

func repairNativeSkills(req *ResolveRequest) error {
	row, err := nativeSkillsState(req.Home, req.Profile, req.EnvID, req.NativeHomeOf)
	if err != nil {
		return err
	}
	if row == nil || row.State == "current" {
		return nil
	}
	results, err := materializeScopeWithNativeHome(req.Home, req.Profile, req.EnvID, req.Policy, req.NativeHomeOf)
	if err != nil {
		return err
	}
	for _, result := range results {
		if !result.OK {
			return fmt.Errorf("%s", result.Detail)
		}
	}
	return nil
}

// An unchanged lock does not imply current native discovery surfaces. Update
// repairs participating effective-current homes without re-resolving path bytes.
func syncCurrentNativeSkills(home, profile string, policy Policy) error {
	for _, adapter := range Adapters {
		if inPlaceSkillsDir(adapter.ID) == profileSkillsSurfaceDir {
			continue
		}
		req := &ResolveRequest{Home: home, Profile: profile, EnvID: adapter.ID, Policy: policy}
		if err := repairNativeSkills(req); err != nil {
			return err
		}
	}
	return nil
}
