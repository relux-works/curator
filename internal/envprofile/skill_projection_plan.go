package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/skillruntime"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// SkillPlanPreparation performs ordinary trusted audit/build/receipt checks on
// private closure snapshots. Nil is refused. The install package supplies the
// production implementation; this seam avoids a reverse import dependency.
// It grants no runtime, command-publication or adoption authority.
type SkillPlanPreparation func([]*closure.Node) error

// SkillDrift reports one unrelated direct installation without repairing it.
type SkillDrift struct {
	Name     string
	Declared manifest.Ref
	Observed *string // nil means proven absence; "unreadable" accompanies unknown.
	State    string  // in-sync, drift or unknown; absence is drift.
}

// SkillPlanResult contains the prepared lock binding and unrelated drift.
type SkillPlanResult struct {
	Lock       *contextlock.Lock
	LockHash   string
	Drift      []SkillDrift
	Diagnostic string
}

// SkillRuntimePreparation joins verified runtime outputs to the same lock journal.
// The witness is held by the outer operation; callbacks must not take another
// home lock, commit a journal or publish mutable routing paths themselves.
type SkillRuntimePreparation func(*contextlock.Lock, []*closure.Node, transaction.HomeLock) (skillruntime.Publication, error)

// PrepareSkillRuntime retains the admission across compilation, staging and
// one journal publication, without permitting v5 profile adoption.
func PrepareSkillRuntime(home, profile, expectedHash string, policy Policy, session *skillruntime.Session, prepare SkillPlanPreparation, stage SkillRuntimePreparation) (SkillPlanResult, error) {
	if session == nil || stage == nil || session.Request().Home != home || session.Request().Profile != profile {
		return SkillPlanResult{}, fmt.Errorf("profile_skill_routing_unproven: bound routing session required")
	}
	return prepareSkillPlan(home, profile, expectedHash, policy, prepare, session, stage)
}

// PrepareSkillPlan explicitly upgrades an installed profile's stored pins to
// lock-v5, selectively preparing its reachable skill closure. It never calls
// global reconciliation, writes declarations, shims, current or managed homes.
// expectedHash is the caller's observed lock pin; it is checked before work and
// again before publication. The shared manager journal publishes only records.
func PrepareSkillPlan(home, profile, expectedHash string, policy Policy, prepare SkillPlanPreparation) (SkillPlanResult, error) {
	return prepareSkillPlan(home, profile, expectedHash, policy, prepare, nil, nil)
}

func prepareSkillPlan(home, profile, expectedHash string, policy Policy, prepare SkillPlanPreparation, session *skillruntime.Session, stage SkillRuntimePreparation) (SkillPlanResult, error) {
	if !validProfileName(profile) || expectedHash == "" || prepare == nil {
		return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: profile, observed lock and preparation required")
	}
	if err := policy.checkHardenedProfileMutation(); err != nil {
		return SkillPlanResult{}, err
	}
	var op *operation
	var err error
	if session == nil {
		op, err = beginOperation(home)
	} else {
		// Admission was acquired before this function, hence before lock/recovery.
		op, err = beginOperationAdmitted(home, session.Revalidate)
	}
	if err != nil {
		return SkillPlanResult{}, err
	}
	defer func() { _ = op.close() }()
	if session != nil {
		current, err := Current(home)
		if err != nil {
			return SkillPlanResult{}, err
		}
		if current != session.Request().Current {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_binding_invalid: stale machine current")
		}
	}
	old, hash, err := readLock(home, profile)
	if err != nil {
		return SkillPlanResult{}, err
	}
	if hash != expectedHash {
		return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_binding_invalid: stale observed lock")
	}
	// A prepared v5 profile is not yet an activatable runtime. PR2 supplies the
	// journal/routing seam; PR3 qualifies companions. Re-preparation is explicit.
	manager := newSkillPlanGitManager(home).withPolicy(policy)
	scratch, err := os.MkdirTemp("", "curator-profile-plan-")
	if err != nil {
		return SkillPlanResult{}, err
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	roots := []closure.ProfileRequirement{}
	contextTrees := map[string]string{}
	for _, member := range old.Contexts() {
		var entry string
		if member.Commit != "" {
			canonical, err := manager.Identity("context", member.Name, member.Source)
			if err != nil {
				return SkillPlanResult{}, err
			}
			if canonical != member.Source {
				return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: context source differs")
			}
			repo, err := manager.ensureRepo(canonical)
			if err != nil {
				return SkillPlanResult{}, err
			}
			if err := contextresolve.VerifyContextPin(manager, contextresolve.Requirement{Name: member.Name, Source: canonical, Revision: member.Commit}, policy.SourceSigners, policy.RequireSourceSigners); err != nil {
				return SkillPlanResult{}, err
			}
			entry, err = os.MkdirTemp(scratch, "context-")
			if err != nil {
				return SkillPlanResult{}, err
			}
			if err := gitops.Extract(repo, member.Commit, entry); err != nil {
				return SkillPlanResult{}, err
			}
		} else {
			if old.ContentHashVersion() != hashing.VersionV2 {
				return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_required: reinstall state context with hash-version 2 before upgrade")
			}
			entry = contextstore.EntryDir(home, "context", member.Name, member.StateHash)
			digest, err := hashing.ContentSHA256WithVersion(entry, map[string]bool{}, hashing.VersionV2)
			if err != nil {
				return SkillPlanResult{}, err
			}
			if strings.TrimPrefix(digest, "sha256:") != member.StateHash {
				return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: context state differs")
			}
		}
		contextDigest, err := hashing.ContentSHA256WithVersion(entry, map[string]bool{}, hashing.VersionV2)
		if err != nil {
			return SkillPlanResult{}, err
		}
		contextTrees[entry] = contextDigest
		selected, err := selectedPackageRoot(entry, member.Directory)
		if err != nil {
			return SkillPlanResult{}, err
		}
		pkg, err := contextpkg.LoadManifest(selected)
		if err != nil {
			return SkillPlanResult{}, err
		}
		if pkg.Name != member.Name || pkg.Version != member.Version {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: context manifest identity differs")
		}
		for _, name := range contextpkg.SortedNames(pkg.Skills) {
			requirement := pkg.Skills[name]
			ref := manifest.Ref{Kind: "revision", Value: requirement.Revision}
			if requirement.Tag != "" {
				ref = manifest.Ref{Kind: "tag", Value: requirement.Tag}
			}
			if requirement.Range != "" {
				return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: skill requirement must be exact")
			}
			roots = append(roots, closure.ProfileRequirement{Consumer: member.Name, Name: name, Git: requirement.Git, Directory: requirement.Directory, Ref: ref, Mode: requirement.Mode, Commands: requirement.Commands})
		}
	}
	globalPath := manifest.PathIn(filepath.Join(home, "global"))
	global, err := stateread.ReadRegularFile(globalPath)
	if err != nil {
		return SkillPlanResult{}, err
	}
	globalDigest, err := transaction.DigestPath(globalPath)
	if err != nil {
		return SkillPlanResult{}, err
	}
	direct := []manifest.Decl{}
	declarations := map[string]any{}
	if global.Kind != stateread.KindAbsent {
		parsed, err := manifest.ParseBytes(global.Bytes, globalPath)
		if err != nil {
			return SkillPlanResult{}, err
		}
		if parsed.SchemaVersion != manifest.SchemaVersion {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: global declarations require schema 1")
		}
		direct = parsed.Skills
		var raw struct {
			Skills []map[string]any `json:"skills"`
		}
		decoder := json.NewDecoder(bytes.NewReader(global.Bytes))
		decoder.UseNumber()
		if err := decoder.Decode(&raw); err != nil {
			return SkillPlanResult{}, err
		}
		for _, declaration := range raw.Skills {
			if name, ok := declaration["name"].(string); ok {
				declarations[name] = declaration
			}
		}
	}
	snapshots := map[string]*closure.Node{}
	refreshedTags := map[string]bool{}
	acquire := func(name, source, directory string, ref manifest.Ref) (*closure.Node, error) {
		canonical, err := manager.Identity("skill", name, source)
		if err != nil {
			return nil, err
		}
		repo, err := manager.ensureRepo(canonical)
		if err != nil {
			return nil, err
		}
		if ref.Kind == "revision" {
			if !contextpkg.ValidCommit(ref.Value) {
				return nil, fmt.Errorf("profile_skill_plan_invalid: exact revision required")
			}
			if err := manager.acquireMissingPin(repo, canonical, ref.Value); err != nil {
				return nil, err
			}
		} else if ref.Kind == "tag" && !refreshedTags[canonical] {
			if err := manager.fetch(canonical); err != nil {
				return nil, err
			}
			refreshedTags[canonical] = true
		}
		pin, err := gitops.Resolve(repo, ref.Kind, ref.Value)
		if err != nil {
			return nil, err
		}
		requirement := contextresolve.Requirement{Name: name, Source: canonical, Revision: pin.Commit}
		if ref.Kind == "tag" {
			requirement.Tag = ref.Value
		}
		if err := contextresolve.VerifySkillPin(manager, requirement, policy.SourceSigners, policy.RequireSourceSigners); err != nil {
			return nil, err
		}
		key := canonical + "\x00" + pin.Commit + "\x00" + directory
		if node := snapshots[key]; node != nil {
			copyNode := *node
			copyNode.Resolved = pin
			return &copyNode, nil
		}
		tree, err := os.MkdirTemp(scratch, "skill-")
		if err != nil {
			return nil, err
		}
		if err := gitops.Extract(repo, pin.Commit, tree); err != nil {
			return nil, err
		}
		selected, spec, err := manifest.SelectSkillPackage(tree, directory, name)
		if err != nil {
			return nil, fmt.Errorf("profile_skill_plan_invalid: %w", err)
		}
		// The strict profile audit consumes this exact extraction, never a stale
		// context-store copy. Its underlying diagnostics are retained.
		resolved := contextresolve.Resolved{Kind: "skill", Name: name, Source: canonical, Commit: pin.Commit, Directory: directory, HashVersion: hashing.VersionV2}
		if _, err := strictAuditMember(home, manager, resolved, tree, policy); err != nil {
			return nil, err
		}
		node := &closure.Node{Name: name, Identity: canonical, Directory: directory, Repo: repo, Snapshot: selected, Spec: spec, Resolved: pin, Decl: manifest.Decl{Name: name, Source: name, Git: source, Ref: ref}}
		snapshots[key] = node
		return node, nil
	}
	plan, err := closure.BuildProfile(old, roots, direct, declarations, acquire)
	if err != nil {
		return SkillPlanResult{}, err
	}
	if err := prepare(plan.Nodes); err != nil {
		return SkillPlanResult{}, err
	}
	// Recheck the same admitted bytes after compilation and runtime staging.
	verifySources := func() error {
		for tree, expected := range contextTrees {
			actual, err := hashing.ContentSHA256WithVersion(tree, map[string]bool{}, hashing.VersionV2)
			if err != nil {
				return err
			}
			if actual != expected {
				return fmt.Errorf("profile_skill_plan_invalid: context changed during preparation")
			}
		}

		for _, pin := range plan.Lock.SkillPlan.Nodes {
			for _, node := range plan.Nodes {
				if node.Name == pin.Name {
					digest, err := hashing.ContentSHA256WithVersion(node.Snapshot, map[string]bool{}, hashing.VersionV2)
					if err != nil {
						return err
					}
					if strings.TrimPrefix(digest, "sha256:") != pin.ContentSHA256 {
						return fmt.Errorf("profile_skill_plan_invalid: frozen source changed during preparation")
					}
				}
			}
		}

		return nil
	}
	if err := verifySources(); err != nil {
		return SkillPlanResult{}, err
	}
	currentGlobal, err := stateread.ReadRegularFile(globalPath)
	if err != nil {
		return SkillPlanResult{}, err
	}
	if currentGlobal.Kind != global.Kind || !bytes.Equal(currentGlobal.Bytes, global.Bytes) {
		return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: direct declarations changed during preparation")
	}
	_, currentHash, err := readLock(home, profile)
	if err != nil {
		return SkillPlanResult{}, err
	}
	if currentHash != hash {
		return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_binding_invalid: lock changed during preparation")
	}
	drift := skillPlanDrift(home, plan.Nodes, direct)
	// Selected immutable snapshots only; unrelated direct installations and
	// retained lock members are neither opened for repair nor rewritten.
	persistentInputs := []string{}
	for _, node := range plan.Nodes {
		entry, err := contextstore.EnsureGit(home, "skill", node.Name, node.Repo, node.Resolved.Commit)
		if err != nil {
			return SkillPlanResult{}, err
		}
		selected, _, err := manifest.SelectSkillPackage(entry, node.Directory, node.Name)
		if err != nil {
			return SkillPlanResult{}, err
		}
		actual, err := hashing.ContentSHA256WithVersion(selected, map[string]bool{}, hashing.VersionV2)
		if err != nil {
			return SkillPlanResult{}, err
		}
		expected, err := hashing.ContentSHA256WithVersion(node.Snapshot, map[string]bool{}, hashing.VersionV2)
		if err != nil {
			return SkillPlanResult{}, err
		}
		if actual != expected {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_invalid: selected store bytes differ from admitted source")
		}
		persistentInputs = append(persistentInputs, selected)
	}

	canonical, err := plan.Lock.Canonical()
	if err != nil {
		return SkillPlanResult{}, err
	}
	oldCanonical, err := old.Canonical()
	if err != nil {
		return SkillPlanResult{}, err
	}
	var publication skillruntime.Publication
	if stage != nil {
		publication, err = stage(plan.Lock, plan.Nodes, op.witness())
		if err != nil {
			return SkillPlanResult{}, err
		}
		if publication.Verify == nil {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_binding_invalid: runtime revalidation required")
		}
		if err := publication.Verify(); err != nil {
			return SkillPlanResult{}, err
		}
	}
	if stage != nil {
		if err := verifySources(); err != nil {
			return SkillPlanResult{}, err
		}
		current, err := Current(home)
		if err != nil {
			return SkillPlanResult{}, err
		}
		if current != session.Request().Current {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_binding_invalid: machine current changed during runtime staging")
		}
		// Durable source-boundary proofs use persistent immutable stores. Private
		// extraction/build paths are released after an interrupted caller returns.
		// Prepared journal bytes seal the operation; recovery does not re-resolve
		// declarations from a later operation.
		publication.Admitted = persistentInputs
		currentGlobal, err := stateread.ReadRegularFile(globalPath)
		if err != nil {
			return SkillPlanResult{}, err
		}
		currentDigest, err := transaction.DigestPath(globalPath)
		if err != nil || currentDigest != globalDigest || currentGlobal.Kind != global.Kind || !bytes.Equal(currentGlobal.Bytes, global.Bytes) {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_binding_invalid: direct declarations changed during runtime staging")
		}
		_, currentHash, err := readLock(home, profile)
		if err != nil {
			return SkillPlanResult{}, err
		}
		if currentHash != hash {
			return SkillPlanResult{}, fmt.Errorf("profile_skill_plan_binding_invalid: lock changed during runtime staging")
		}
	}
	files := map[string][]byte{lockPath(home, profile): canonical, prevLockPath(home, profile): oldCanonical}
	if err := op.publishCombined(files, nil, nil, publication); err != nil {
		return SkillPlanResult{}, err
	}
	return SkillPlanResult{Lock: plan.Lock, LockHash: contextlock.HashBytes(canonical), Drift: drift, Diagnostic: publication.Diagnostic}, nil
}

func skillPlanDrift(home string, nodes []*closure.Node, direct []manifest.Decl) []SkillDrift {
	selected := map[string]bool{}
	for _, node := range nodes {
		selected[node.Name] = true
	}
	out := []SkillDrift{}
	for _, decl := range direct {
		if selected[decl.Name] {
			continue
		}
		state := "unknown"
		unreadable := "unreadable"
		observed := &unreadable
		entry := filepath.Join(home, "global", "skills", decl.Name)
		metadata, err := stateread.Lstat(entry)
		if err == nil && metadata.Kind == stateread.KindAbsent {
			state = "drift"
			observed = nil
		} else if err == nil && metadata.Kind == stateread.KindPresent && metadata.Info.IsDir() {
			recorded, _, readErr := marker.ReadState(entry)
			if readErr == nil && recorded != nil {
				state = "drift"
				pin := recorded.Commit
				if recorded.Package != nil && recorded.Package.Commit != nil {
					pin = recorded.Package.Commit.Hex
				}
				observed = &pin
				sameSource := decl.Git != "" && identity.Canonical(decl.Git) == identity.Canonical(recorded.Git)
				sameRef := recorded.RefKind == decl.Ref.Kind && recorded.Ref == decl.Ref.Value && (decl.Ref.Kind != "revision" || decl.Ref.Value == pin)
				if recorded.Package != nil {
					sameSource = identity.Canonical(decl.Git) == recorded.Package.Repository
					if recorded.Package.Directory != "." {
						sameRef = false
					}
				}
				if sameSource && sameRef {
					current, currentErr := marker.Current(entry, recorded)
					if currentErr != nil {
						state = "unknown"
					} else if current {
						state = "in-sync"
					}
				}
			}
		}

		out = append(out, SkillDrift{Name: decl.Name, Declared: decl.Ref, Observed: observed, State: state})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// refusePreparedSkillPlanAdoption preserves the companion diagnostic before a
// legacy entry point attempts domain admission or journal recovery. This is a
// read-only refusal probe; existing under-lock validation remains authoritative
// for absent, malformed or concurrently replaced records.
func refusePreparedSkillPlanAdoption(home, profile string) error {
	if profile == "" {
		var err error
		profile, err = Current(home)
		if err != nil {
			return nil
		}
		if profile == "" {
			profile = DefaultProfile
		}
	}
	if validProfileName(profile) {
		if candidate, _, err := readLock(home, profile); err == nil && candidate.SchemaVersion == contextlock.SchemaVersion5 {
			return RefuseSkillPlanAdoption()
		}
	}
	return nil
}

// RefuseSkillPlanAdoption is unconditional until accepted companion allocations
// and their authoritative qualification readers are implemented in PR3. A
// prepared lock is never a completion token or a caller-selectable bypass.
func RefuseSkillPlanAdoption() error {
	return fmt.Errorf("profile_skill_companion_required: profile-command-admission and managed-session-skill-cli qualification required")
}

// Reuse the existing identity-bound cache/admission implementation, with a
// distinct profile-plan namespace so lossy historical clone keys cannot alias.
func newSkillPlanGitManager(home string) *gitManager {
	manager := newGitManager(home)
	manager.sourceCacheRoot = filepath.Join(home, "profile-skill-repos")
	manager.missingPins = map[string]bool{}
	return manager
}
