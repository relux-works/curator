package envprofile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/whitelist"
)

// The publication revision is independent of the audit ruleset: it versions
// the installed tree and surface identity without changing admission inputs.
const skillProjectionVersion = "core-3.1-v1"

func skillSurfaceIdentity(skill skillOf) []byte {
	return []byte(skillProjectionVersion + ":" + skill.hash)
}

// projectSkill plans without writing. Both the digest and the eventual copy
// use the same core context walker, including canonical filesystem aliases.
func projectSkill(home, source string, version hashing.Version) (skillOf, error) {
	spec, err := skillspec.Load(source)
	if err != nil {
		return skillOf{}, err
	}
	skill := skillOf{source: source, includeScripts: len(spec.Commands) == 0,
		excludeRoots: whitelist.ContextExcludedRoots(spec.RuntimeRoots, spec.BuildRoots), version: version}
	files := map[string][]byte{}
	err = whitelist.WalkContext(source, skill.includeScripts, skill.excludeRoots, func(path, posix string) error {
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[posix] = body
		return nil
	})
	if err != nil {
		return skillOf{}, err
	}
	skill.hash, err = hashing.ContentSHA256Files(files, version)
	if err != nil {
		return skillOf{}, err
	}
	skill.root = filepath.Join(contextstore.Root(home), "skill-context", skillProjectionVersion, fmt.Sprintf("v%d", version), hashing.Normalize(skill.hash))
	return skill, nil
}

// publishSkillProjection runs only during materialization under its mutation
// lock. A complete private staging tree is renamed into the shared store;
// raw package snapshots and command/runtime stores remain untouched.
func publishSkillProjection(home string, skill skillOf) error {
	parent := filepath.Dir(skill.root)
	if err := privatedir.MakeAll(parent); err != nil {
		return err
	}
	if err := pathboundary.ValidateRouteWithOwner(contextstore.Root(home), parent, pathboundary.DefaultOwnerLookup()); err != nil {
		return err
	}
	state, err := stateread.Lstat(skill.root)
	if err != nil {
		return err
	}
	if state.Kind == stateread.KindPresent {
		return verifySkillProjection(home, skill)
	}
	staging, err := os.MkdirTemp(parent, ".projection-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if _, err := whitelist.CopyContext(skill.source, staging, skill.includeScripts, skill.excludeRoots); err != nil {
		return err
	}
	digest, err := hashing.ContentSHA256WithVersion(staging, map[string]bool{}, skill.version)
	if err != nil {
		return err
	}
	if digest != skill.hash {
		return fmt.Errorf("skill context changed while projecting")
	}
	if err := privatedir.ProtectTree(staging); err != nil {
		return err
	}
	if err := os.Rename(staging, skill.root); err != nil {
		return err
	}
	return nil
}

func verifySkillProjection(home string, skill skillOf) error {
	if err := pathboundary.ValidateWithin(contextstore.Root(home), skill.root); err != nil {
		return err
	}
	info, err := os.Lstat(skill.root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("skill context projection is not a directory")
	}
	digest, err := hashing.ContentSHA256WithVersion(skill.root, map[string]bool{}, skill.version)
	if err != nil {
		return err
	}
	if digest != skill.hash {
		return fmt.Errorf("skill context projection differs from its identity")
	}
	return nil
}
