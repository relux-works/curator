package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
	"github.com/relux-works/curator/internal/whitelist"
)

// The publication revision is independent of the audit ruleset: it versions
// the installed tree and surface identity without changing admission inputs.
const skillProjectionVersion = "core-3.1-v1"

// ErrSkillProjectionIdentity identifies a projection that differs from its pin.
var ErrSkillProjectionIdentity = errors.New("skill context projection differs from its identity")

func skillSurfaceIdentity(skill skillOf) []byte {
	return []byte(skillProjectionVersion + ":" + skill.hash)
}

// projectSkill plans without writing. Both the digest and the eventual copy
// use the same core context walker, including canonical filesystem aliases.
func projectSkill(home, source string, version hashing.Version) (skillOf, error) {
	if err := opaquescan.RefuseNULV1(source, version); err != nil {
		return skillOf{}, err
	}
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
func publishSkillProjection(op *operation, home string, skill skillOf) error {
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
		if err := verifySkillProjection(home, skill); !errors.Is(err, ErrSkillProjectionIdentity) {
			return err
		}
	}
	// Capture the old complete tree before staging. The transaction rechecks
	// this preimage before replacing it; foreign routes and links were refused
	// by verification above, and a failed rebuild never touches this tree.
	preimage, err := transaction.DigestPath(skill.root)
	if err != nil {
		return err
	}
	staging, err := os.MkdirTemp(parent, ".projection-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if _, err := whitelist.CopyContext(skill.source, staging, skill.includeScripts, skill.excludeRoots); err != nil {
		return err
	}
	digest, err := skillProjectionHash(staging, skill.version)
	if err != nil {
		return err
	}
	if digest != skill.hash {
		return fmt.Errorf("%w: skill context changed while projecting", ErrSkillProjectionIdentity)
	}
	if err := privatedir.ProtectTree(staging); err != nil {
		return err
	}
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profile-skill-projection", Targets: []transaction.Target{{
		Class: "skill-projection", Identifier: skill.hash, LivePath: skill.root,
		StagedSource: staging, PreimageDigest: preimage,
	}}}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return err
	}
	return op.engine.Commit(op.lock, plan.TransactionID)
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
	digest, err := skillProjectionHash(skill.root, skill.version)
	if err != nil {
		return err
	}
	if digest != skill.hash {
		return ErrSkillProjectionIdentity
	}
	return nil
}

// Only an explicitly recorded skill fallback is eligible for replacement.
// Unrecorded directories still pass through the ordinary conflict inventory.
func recordedSkillFallback(prior *envmarker.Marker, path string) bool {
	if prior == nil {
		return false
	}
	surface := prior.Surfaces[envmarker.SurfaceSkills]
	if surface.Copies == nil {
		return false
	}
	for _, copy := range *surface.Copies {
		if copy.Path == path && copy.Reason == envmarker.ReasonSymlinkFallback {
			return true
		}
	}
	return false
}

// Replace the complete copied tree through the existing recoverable entry
// transaction. Overlay-copying would retain excluded files from the old tree;
// renaming a symlink over a non-empty directory is unavailable on some hosts.
func publishSkillFallback(op *operation, root, rel, source string) error {
	live, err := managedPath(root, rel, true)
	if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(live), ".curator-skill-copy-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	stageRel, err := managedRelative(root, stage)
	if err != nil {
		return err
	}
	nextRel := managedJoin(stageRel, "next")
	if err := copyTree(source, root, nextRel); err != nil {
		return err
	}
	digest, err := transaction.DigestTarget(transaction.KindEntry, live)
	if err != nil {
		return err
	}
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profile-skill-projection", Targets: []transaction.Target{{
		Class: "skill-context", Identifier: rel, Kind: transaction.KindEntry, LivePath: live,
		StagedSource: filepath.Join(stage, "next"), PreimageDigest: digest,
	}}}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return err
	}
	return op.engine.Commit(op.lock, plan.TransactionID)
}

// Every projection-tree identity retains the legacy no-v1-hash-over-NUL
// invariant, including changed staging input and a damaged projected store.
func skillProjectionHash(root string, version hashing.Version) (string, error) {
	if err := opaquescan.RefuseNULV1(root, version); err != nil {
		return "", err
	}
	return hashing.ContentSHA256WithVersion(root, map[string]bool{}, version)
}
