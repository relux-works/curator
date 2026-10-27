package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/skillruntime"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// stageProfileRuntime reuses ordinary closed runtime targets and verified cache
// winners. It never calls Global, repairs an unrelated marker, or commits a
// journal. The envprofile owner joins these targets to the new lock records.
func stageProfileRuntime(cfg *config.Config, profile string, lock *contextlock.Lock, nodes []*closure.Node, plan BuildPlan, staged Staged, external stagedExternal, deps BuildDeps, opts Options, private *privateRoot, session *skillruntime.Session, witness transaction.HomeLock) (skillruntime.Publication, []externalAdoption, error) {
	fail := func(err error) (skillruntime.Publication, []externalAdoption, error) {
		return skillruntime.Publication{}, nil, err
	}
	if opts.Platform != "" && opts.Platform != runtimestore.Platform() {
		return fail(fmt.Errorf("profile_skill_companion_required: routing target platform differs"))
	}
	payload, err := lock.Canonical()
	if err != nil {
		return fail(err)
	}
	binding := skillruntime.Binding{Profile: profile, LockHash: contextlock.HashBytes(payload), PlanHash: lock.SkillPlanSHA256}
	exports := []skillruntime.Export{}
	for _, node := range nodes {
		for _, name := range node.ActiveCommandNames() {
			exports = append(exports, skillruntime.Export{Name: name, Skill: node.Name, Profile: profile})
		}
	}
	// Inventory/admitted-base checks apply before protected cache publication,
	// including retained exports. Inactive preparation cannot claim delivery.
	inventory, err := session.Preflight(binding, exports)
	if err != nil {
		return fail(err)
	}
	commit, err := opts.Commit.resolve(cfg.Home(), deps.Cache)
	if err != nil {
		return fail(err)
	}
	commit, err = bindCommitPublisher(deps.Assurance, commit)
	if err != nil {
		return fail(err)
	}
	// The outer profile operation owns release. Preserve its admission-aware
	// AssertHeld for cache publication without handing the callee its closer.
	homeLock := borrowedProfileRuntimeHome{witness}
	winners, err := publishWinners(commitRequest{home: cfg.Home(), scope: "profile:" + profile, plan: plan, staged: staged, commit: commit}, homeLock)
	if err != nil {
		return fail(err)
	}
	// Verified immutable cache entries may be retained after refusal. Mutable
	// runtime/routing/lock state is exclusively in the single outer journal.
	stageRoot, err := private.dir("profile-runtime-")
	if err != nil {
		return fail(err)
	}
	externalRoot := opts.External.resolved(cfg.Home()).StoreRoot
	bin := filepath.Join(cfg.Home(), "global", "bin")
	active := session.Request().Current == profile

	// Runtime trees belong to the complete profile lock generation, not only
	// a Git commit: two selected package directories can share one commit.
	// Inactive preparation must never replace a current profile's script tree.
	runtimeKeys := map[string]string{}
	for _, node := range nodes {
		runtimeKeys[node.Name] = "profile-" + strings.TrimPrefix(binding.LockHash, "sha256:")
	}
	prepared, err := stageRuntimeAndShims(stageRoot, cfg.Home(), bin, nodes, runtimestore.GlobalCanonicalShim, runtimestore.Platform(), scopeCommit{stageRoot: stageRoot, artifacts: winners.artifacts}, plan.plannedInputs(), external.entries, externalRoot, runtimeKeys, "")
	if err != nil {
		return fail(err)
	}
	var targets staging.Plan
	targets.Merge(external.transactionPlan(externalRoot))
	for _, target := range prepared.plan.Targets {
		if active || target.Class == staging.ClassRuntime {
			targets.Targets = append(targets.Targets, target)
		}
	}
	// A protected per-lock generation record keeps artifact/receipt bindings out
	// of the portable lock. Prior generations are retained; PR3 owns release/GC.
	generation := struct {
		Binding     skillruntime.Binding `json:"binding"`
		Builds      any                  `json:"builds"`
		Keys        []string             `json:"build_keys"`
		RuntimeKeys map[string]string    `json:"runtime_keys"`
	}{binding, prepared.builds, prepared.referencedKeys(), runtimeKeys}
	generationBytes, err := json.Marshal(generation)
	if err != nil {
		return fail(err)
	}
	generationStaged := filepath.Join(stageRoot, "generation.json")
	if err := os.WriteFile(generationStaged, generationBytes, 0600); err != nil {
		return fail(err)
	}
	generationLive := filepath.Join(skillruntime.DomainDir(cfg.Home()), "generations", strings.TrimPrefix(binding.LockHash, "sha256:")+".json")
	targets.Replace("84-profile-generation", binding.LockHash, generationLive, generationStaged)
	var publication skillruntime.Publication
	if active {
		forwardable := map[string]bool{}
		for name := range prepared.commands {
			if !prepared.enforced[name] {
				forwardable[name] = true
			}
		}
		for _, retained := range inventory.RetainedAt(session.Domain().UserBin) {
			forwardable[retained.Name] = true
		}
		domain := session.Domain()
		forwarding, err := globalbins.StageForwarding(stageRoot, cfg.Home(), forwardable, runtimestore.Platform(), map[string]string{"PATH": domain.FixedPATH, globalbins.UserBinEnv: domain.UserBin}, domain.UserHome)
		if err != nil {
			return fail(err)
		}
		if forwarding.Path != domain.UserBin || forwarding.Published != len(forwardable) {
			return fail(fmt.Errorf("environment_surface_unmanaged_conflict: forwarding root unavailable: %s", strings.Join(forwarding.Messages, "; ")))
		}
		targets.Merge(forwarding.Plan())
		publication, err = session.Bind(stageRoot, binding, exports, inventory, targets, prepared.referencedKeys())
		if err != nil {
			return fail(err)
		}
	} else {
		publication = skillruntime.Publication{Plan: targets, ReferencedBuildKeys: prepared.referencedKeys(), Diagnostic: skillruntime.Unavailable, Verify: session.Revalidate}
	}

	before := publication.Verify
	publication.Verify = func() error {
		if err := plan.Verify(opts.context()); err != nil {
			return err
		}
		if err := before(); err != nil {
			return err
		}
		current, err := envprofile.Current(cfg.Home())
		if err != nil {
			return err
		}
		if current != session.Request().Current {
			return fmt.Errorf("profile_skill_plan_binding_invalid: machine current changed")
		}
		return nil
	}
	return publication, external.adoptions(externalRoot), nil
}

// Cache publication borrows the outer witness and cannot release its lock.
type borrowedProfileRuntimeHome struct{ transaction.HomeLock }

func (borrowedProfileRuntimeHome) Close() error { return nil }
