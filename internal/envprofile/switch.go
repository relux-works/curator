// Package envprofile switching covers linked and copied in-place
// materialization (environments §8.1, §9.2, manager profile §12.2-§12.3).
//
// profile use re-materializes every in-place surface of the scope from the
// profile's store entries: it attempts every entry, reports per-adapter
// results, and records the new current only when the whole scope
// materialized — the M11 transactional shape. A partial scope is reported
// as profile_use_partial and the recorded current is unchanged. A
// machine-scope switch skips adapters that carry a scope record
// (environments §9.3): those homes follow their scoped profiles, not the
// machine current, so the manager never reports env:<id>=<profile> for a
// home it has just overwritten with a different profile.
//
// Serialization and durability (see lock.go): every switch holds the
// manager-home mutation lock, and the manager-home records it moves — the
// current and scope pointers — commit through the internal/transaction
// journal. Each in-place entry uses an enumerated durable publication unit,
// including its versioned backups and marker. Successful earlier entries may
// remain on a partial scope result (§9.2). Managed repair instead joins its
// participating native skills and managed home in one complete unit.
//
// Every replaced file is backed up into a versioned generation
// .agent-environment-backup/<n>/ beside the marker (environments §8.3),
// pruned to the retention default of 5 (0 keeps every generation only via
// env backups scrub, which is out of scope for this stage). A write that
// would touch a file the preceding marker does not record fails the entry
// with environment_surface_unmanaged_conflict and never overwrites.
//
// Managed homes, seeds, passthrough, and secondary fixed-home targets are
// stage (b): this file materializes the four registered adapters' native
// in-place homes only. Root context and profile skill directories are
// recorded as separate surfaces; MCP files remain managed-home surfaces.
package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/stateread"
)

// Diagnostics for switching (environments §8.5, §9.7).
const (
	DiagUsePartial           = "profile_use_partial"
	DiagUnmanagedConflict    = "environment_surface_unmanaged_conflict"
	DiagBackupExists         = "environment_backup_exists"
	DiagUnknownEnvironment   = "environment_unknown"
	DiagWriteWouldFollowLink = "environment_write_would_follow_link"
	profileSkillsSurfaceDir  = ".agent-context/skills"
)

// BackupRetention is the number of backup generations kept (environments
// §8.3, §12.1 default 5). Manager-config schema 2 may change the knob in a
// later stage; this stage always applies the default.
const BackupRetention = 5

// Adapter is one registered environment adapter (environments §7.1).
type Adapter struct {
	ID         string
	EnvVar     string
	DefaultDir string
	Target     string
}

// Adapters is the closed revision-1 registry.
var Adapters = []Adapter{
	{ID: "claude_code", EnvVar: "CLAUDE_CONFIG_DIR", DefaultDir: ".claude", Target: "CLAUDE.md"},
	{ID: "codex_cli", EnvVar: "CODEX_HOME", DefaultDir: ".codex", Target: "AGENTS.md"},
	{ID: "opencode", EnvVar: "XDG_CONFIG_HOME", DefaultDir: ".config", Target: "AGENTS.md"},
	{ID: "pi", EnvVar: "PI_CODING_AGENT_DIR", DefaultDir: ".pi", Target: "AGENTS.md"},
}

// adapterByID returns the adapter or false for an unregistered identifier.
func adapterByID(id string) (Adapter, bool) {
	for _, adapter := range Adapters {
		if adapter.ID == id {
			return adapter, true
		}
	}
	return Adapter{}, false
}

// EntryResult is the per-adapter outcome of a switch. Notice carries the
// section 9.5 takeover notice for the entry; Warnings carries the
// non-blocking onboarding findings (the dotfile-manager heuristic).
type EntryResult struct {
	Adapter  string
	Home     string
	OK       bool
	Detail   string
	Notice   string
	Warnings []string
	cause    error
}

// Use switches the machine scope (environment and target both empty) or one
// narrowed scope to name. A machine-scope switch skips adapters that carry
// a scope record: those homes stay on their scoped profiles. A scope record
// equal to the new machine default is removed in the same journaled publish
// as the machine current. With clear, the scope record is dropped and the
// scope re-materializes from the machine default. It returns the per-entry
// results; when any entry failed the recorded current is unchanged and the
// error carries profile_use_partial.
// Use holds the manager-home mutation lock (see lock.go) and records the
// new current through the operation journal only when the whole scope
// materialized.
// The machine gates come from the process configuration (see
// loadMachinePolicy); callers operating on any other manager home must use
// UseWithPolicy.
func Use(home, name, environment, target string, clearScope bool) ([]EntryResult, error) {
	policy, err := loadMachinePolicy()
	if err != nil {
		return nil, err
	}
	return UseWithPolicy(home, name, environment, target, clearScope, policy)
}

// UseWithPolicy switches under the machine gates of policy (environments
// §9.1). The CLI passes the gates of its already-loaded configuration.
func UseWithPolicy(home, name, environment, target string, clearScope bool, policy Policy) ([]EntryResult, error) {
	op, err := beginOperation(home)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()
	return useLocked(op, home, name, environment, target, clearScope, policy)
}

// useLocked switches under the held operation lock.
func useLocked(op *operation, home, name, environment, target string, clearScope bool, policy Policy) (_ []EntryResult, resultErr error) {
	profileToValidate := name
	if clearScope {
		current, err := Current(home)
		if err != nil {
			return nil, err
		}
		profileToValidate = current
		if profileToValidate == "" {
			profileToValidate = DefaultProfile
		}
	}
	if profileToValidate != "" {
		if err := preflightProfilePathSources(home, profileToValidate, policy); err != nil {
			return nil, err
		}
	}
	if err := ensureDefault(op, home, policy); err != nil {
		return nil, err
	}
	if environment != "" {
		if _, ok := adapterByID(environment); !ok {
			return nil, fmt.Errorf("%s: explicit operand names the unregistered environment %q", DiagUnknownEnvironment, environment)
		}
	}
	if target != "" {
		if environment != "" {
			if _, err := envregistry.TargetFor(environment, target); err != nil {
				return nil, err
			}
		} else if _, err := envregistry.TargetByID(target); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("secondary fixed-home target %q writes are deferred: participation, consent, and status are implemented, surface writes are not", target)
	}
	scope := ""
	if environment != "" {
		scope = "env:" + environment
	}
	effective := name
	if clearScope {
		// The CLI admits no operand beside --clear (see cmdProfileUse):
		// a positional here is an undefined form, and silently ignoring
		// it once switched the machine scope to another profile. Fail
		// rather than guess which profile the operator meant.
		if name != "" {
			return nil, fmt.Errorf("profile use --clear takes no profile operand")
		}
		machine, err := Current(home)
		if err != nil {
			return nil, err
		}
		if machine == "" {
			machine = DefaultProfile
		}
		effective = machine
	}
	// The §12.2 fleet-policy gate is the single construction point of a
	// machine-scope switch: every path that will publish the machine
	// current below — profile use (clear or not; the CLI rejects the
	// combination but the seam does not rely on it), install --use,
	// first-install auto-activation, import --use, and resync — funnels
	// through this one gate before any write, not through any one CLI
	// row. It precedes the installed check, so a machine-scope switch to
	// another profile is a configuration error even when that profile is
	// not installed. A scoped switch records a scoped current and is
	// unaffected. CurrentFile has exactly one production writer, the
	// publish below, so no caller reaches the seam in machine scope
	// without passing here.
	if scope == "" {
		if err := policy.CheckMachineUse(effective); err != nil {
			return nil, err
		}
	}
	if !clearScope {
		if _, err := readSource(home, name); err != nil {
			return nil, err
		}
	}
	activation, err := planScopePublication(home, scope, effective, clearScope)
	if err != nil {
		return nil, err
	}
	if !clearScope {
		if results, handled, err := continueRetainedGlobals(op, home, name, policy, &activation); handled || err != nil {
			return results, err
		}
	}
	if hashing.WriteVersion() == hashing.VersionV2 {
		lock, _, err := readLock(home, effective)
		if err != nil {
			return nil, err
		}
		if lock.ContentHashVersion() == hashing.VersionV1 {
			req := &ResolveRequest{Home: home, Profile: effective, Machine: envregistry.DefaultMachineConfig(), Policy: policy}
			migration, err := prepareIdentityMigration(op, req, []string{effective}, nil)
			if err != nil {
				return nil, err
			}
			defer migration.cleanup()
			if err := migration.publish(op); err != nil {
				return nil, err
			}
		}
	}
	g := newHomePublicationGroup()
	defer finishPublicationCleanup(&resultErr, g.close)
	op.publicationGroup = g
	defer func() { op.publicationGroup = nil }()
	results, err := materializeScope(op, home, effective, environment, policy)
	if err != nil {
		return results, err
	}
	failed := false
	for _, result := range results {
		if !result.OK {
			failed = true
		}
	}
	if failed {
		return results, fmt.Errorf("%s: scope planning refused; surfaces and current are unchanged", DiagUsePartial)
	}
	if g.unit != nil {
		if err := op.commitHomeGroup(home, g, activation.files, activation.removals, nil, nil); err != nil {
			return results, err
		}
	} else if err := op.publish(activation.files, activation.removals...); err != nil {
		return results, err
	}
	return results, nil
}

type scopePublication struct {
	files    map[string][]byte
	removals []string
}

// Current pointers are planned before any home publication, then enlisted in
// the same whole-scope unit. Clearing includes both supported filename forms.
func planScopePublication(home, scope, effective string, clear bool) (scopePublication, error) {
	publication := scopePublication{files: map[string][]byte{}}
	if scope == "" {
		removals, err := scopeRecordsEqualToProfile(home, effective)
		if err != nil {
			return publication, err
		}
		publication.files[CurrentFile(home)] = []byte(effective + "\n")
		publication.removals = removals
		return publication, nil
	}
	machine, err := Current(home)
	if err != nil {
		return publication, err
	}
	path := filepath.Join(ScopedDir(home), scopeFileName(scope))
	if !clear && effective != machine {
		publication.files[path] = []byte(effective + "\n")
	}
	for _, spelling := range []string{scopeFileName(scope), scope} {
		candidate := filepath.Join(ScopedDir(home), spelling)
		if _, err := os.Lstat(candidate); err == nil {
			if _, written := publication.files[candidate]; !written {
				found := false
				for _, prior := range publication.removals {
					found = found || prior == candidate
				}
				if !found {
					publication.removals = append(publication.removals, candidate)
				}
			}
		} else if !os.IsNotExist(err) {
			return publication, err
		}
	}
	return publication, nil
}

// scopeRecordsEqualToProfile returns every on-disk spelling of each scope
// record whose value equals profile. The machine current and these removals
// are passed to one operation.publish call so the section 9.3 invariant is
// durable as one journaled state transition.
func scopeRecordsEqualToProfile(home, profile string) ([]string, error) {
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return nil, err
	}
	scopes := make([]string, 0, len(scoped))
	for scope, current := range scoped {
		if current == profile {
			scopes = append(scopes, scope)
		}
	}
	sort.Strings(scopes)

	var removals []string
	for _, scope := range scopes {
		paths, err := scopeRecordPaths(home, scope)
		if err != nil {
			return nil, err
		}
		removals = append(removals, paths...)
	}
	sort.Strings(removals)
	return removals, nil
}

// scopeRecordPaths lists all existing filename spellings for one decoded
// scope key, including a legacy literal spelling when present. The caller
// has already read the records through ScopedCurrents; an absent directory
// or vanished matching entry here is a changed record, not a reason to leave
// the equal-default invariant partially applied.
func scopeRecordPaths(home, scope string) ([]string, error) {
	dir := ScopedDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || scopeKeyName(entry.Name()) != scope {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("%s: scoped current %q changed while switching", DiagSourceInvalid, scope)
	}
	return paths, nil
}

// Sync re-materializes the machine scope and every scoped current from
// their profiles' locks: the stage-(a) actualization path. The machine pass
// skips adapters that carry a scope record and the scoped pass writes each
// of those homes once from its record, so no home is written twice. Managed
// homes are stage (b) and are not provisioned here. Sync holds the
// manager-home mutation lock; it moves no current pointer, so there is
// nothing to journal beyond the recovery the lock entry already performed.
// The machine gates come from the process configuration (see
// loadMachinePolicy); callers operating on any other manager home must use
// SyncWithPolicy.
func Sync(home string) ([]EntryResult, error) {
	policy, err := loadMachinePolicy()
	if err != nil {
		return nil, err
	}
	return SyncWithPolicy(home, policy)
}

// SyncWithPolicy syncs under the machine gates of policy (environments
// §9.1). The CLI passes the gates of its already-loaded configuration.
func SyncWithPolicy(home string, policy Policy) (_ []EntryResult, resultErr error) {
	op, err := beginOperation(home)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()
	if results, handled, err := continueRetainedGlobals(op, home, "", policy, nil); handled || err != nil {
		return results, err
	}
	if err := preflightCurrentPathSources(home, policy); err != nil {
		return nil, err
	}
	if err := ensureDefault(op, home, policy); err != nil {
		return nil, err
	}
	machine, err := Current(home)
	if err != nil {
		return nil, err
	}
	if machine == "" {
		machine = DefaultProfile
	}
	g := newHomePublicationGroup()
	defer finishPublicationCleanup(&resultErr, g.close)
	op.publicationGroup = g
	defer func() { op.publicationGroup = nil }()
	var results []EntryResult
	machineResults, err := materializeScope(op, home, machine, "", policy)
	if err != nil {
		return machineResults, err
	}
	results = append(results, machineResults...)
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return results, err
	}
	scopes := make([]string, 0, len(scoped))
	for scope := range scoped {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	for _, scope := range scopes {
		environment, _ := splitScope(scope)
		scopeResults, err := materializeScope(op, home, scoped[scope], environment, policy)
		if err != nil {
			return append(results, scopeResults...), err
		}
		results = append(results, scopeResults...)
	}
	for _, row := range results {
		if !row.OK {
			return results, fmt.Errorf("%s: scope planning refused; all surfaces are unchanged", DiagUsePartial)
		}
	}
	if err := op.commitHomeGroup(home, g, nil, nil, nil, nil); err != nil {
		return results, err
	}
	return results, nil
}

// materializeScope materializes profile name into every adapter home of the
// scope (one adapter when environment names it), attempting every entry.
//
// A machine-scope switch (environment empty) skips adapters that carry a
// scope record (environments §9.3): such a home follows its scoped profile,
// not the machine current, so the machine pass has no business writing it.
// SyncWithPolicy re-materializes each skipped home from its scoped record in
// its own pass, so every home is written exactly once and the recorded state
// and the bytes always agree when the command returns.
func materializeScope(op *operation, home, profile, environment string, policy Policy) ([]EntryResult, error) {
	return materializeScopeWithNativeHome(op, home, profile, environment, policy, nil)
}

func materializeScopeWithNativeHome(op *operation, home, profile, environment string, policy Policy, nativeHomeOf func(string) (string, error)) ([]EntryResult, error) {
	lock, hash, err := readLock(home, profile)
	if err != nil {
		return nil, err
	}
	source, err := readSource(home, profile)
	if err != nil {
		return nil, err
	}
	return materializeScopeObserved(op, home, profile, environment, policy, nativeHomeOf, source, lock, hash)
}

func materializeScopeObserved(op *operation, home, profile, environment string, policy Policy, nativeHomeOf func(string) (string, error), source Source, lock *contextlock.Lock, hash string) ([]EntryResult, error) {
	var err error
	if err := validateProfilePathSources(profile, source, policy); err != nil {
		return nil, err
	}
	var adapters []Adapter
	if environment != "" {
		adapter, _ := adapterByID(environment)
		adapters = []Adapter{adapter}
	} else {
		scoped, err := ScopedCurrents(home)
		if err != nil {
			return nil, err
		}
		for _, adapter := range Adapters {
			if _, ok := scoped["env:"+adapter.ID]; ok {
				continue
			}
			adapters = append(adapters, adapter)
		}
	}
	manager := newGitManager(home)
	packages, err := loadMaterial(home, manager, lock)
	if err != nil {
		return nil, err
	}
	skills, err := skillsOf(home, manager, lock)
	if err != nil {
		return nil, err
	}
	precedence := policy.Precedence()
	order, err := contextmaterialize.EmittedOrder(lock, precedence)
	if err != nil {
		return nil, err
	}
	var results []EntryResult
	for _, adapter := range adapters {
		results = append(results, materializeOne(op, home, source, profile, lock, hash, precedence, order, packages, skills, adapter, policy, nativeHomeOf))
	}
	return results, nil
}

// preflightInPlaceScope checks every native target a scope will materialize
// before the caller starts writing any surface. Global lock publication uses
// this while holding the manager-home lock so a static unmanaged conflict
// cannot leave earlier adapter surfaces partially updated.
func preflightInPlaceScope(home, profile, environment string, policy Policy, nativeHomeOf func(string) (string, error)) error {
	lock, hash, err := readLock(home, profile)
	if err != nil {
		return err
	}
	source, err := readSource(home, profile)
	if err != nil {
		return err
	}
	return preflightInPlaceObserved(home, profile, environment, policy, nativeHomeOf, source, lock, hash)
}

func preflightInPlaceObserved(home, profile, environment string, policy Policy, nativeHomeOf func(string) (string, error), source Source, lock *contextlock.Lock, hash string) error {
	var err error
	if err := validateProfilePathSources(profile, source, policy); err != nil {
		return err
	}
	manager := newGitManager(home)
	packages, err := loadMaterial(home, manager, lock)
	if err != nil {
		return err
	}
	skills, err := skillsOf(home, manager, lock)
	if err != nil {
		return err
	}
	precedence := policy.Precedence()
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return err
	}
	var adapters []Adapter
	if environment != "" {
		adapter, ok := adapterByID(environment)
		if !ok {
			return fmt.Errorf("%s: explicit operand names the unregistered environment %q", DiagUnknownEnvironment, environment)
		}
		adapters = []Adapter{adapter}
	} else {
		for _, adapter := range Adapters {
			if _, hasScope := scoped["env:"+adapter.ID]; !hasScope {
				adapters = append(adapters, adapter)
			}
		}
	}
	for _, adapter := range adapters {
		native, err := resolveNativeHome(home, adapter, nativeHomeOf)
		if err != nil {
			return err
		}
		rootState, err := stateread.Stat(native)
		if err != nil {
			return err
		}
		if rootState.Kind == stateread.KindAbsent {
			continue
		}
		if rootState.Info == nil || !rootState.Info.IsDir() {
			return fmt.Errorf("managed root %s is not a directory", native)
		}
		if err := checkManagedPrivateTarget(native, envmarker.Name); err != nil {
			return err
		}
		prior, err := envmarker.Read(native)
		if err != nil {
			return err
		}
		recorded := map[string]bool{}
		if prior != nil {
			for _, surface := range prior.Surfaces {
				for _, path := range surface.Paths {
					recorded[path] = true
				}
			}
		}
		want := map[string]bool{}
		_, written, err := contextmaterialize.Monolithic(lock, hash, precedence, adapter.ID, packages)
		if err != nil {
			return err
		}
		if written {
			want[adapter.Target] = true
		}
		for _, skill := range skills {
			want[filepath.ToSlash(filepath.Join(inPlaceSkillsDir(adapter.ID), skill.name))] = true
		}
		ledger, err := inPlaceSkillsLedger(home, native, adapter, skills, recorded)
		if err != nil {
			return err
		}
		for path := range recorded {
			if _, err := managedPath(native, path, false); err != nil {
				return err
			}
		}
		for path := range want {
			if _, err := managedPath(native, path, false); err != nil {
				return err
			}
			if ledgerSkillPath(adapter, path, ledger) {
				continue
			}
			if recorded[path] {
				continue
			}
			full, err := managedPath(native, path, false)
			if err != nil {
				return err
			}
			metadata, err := stateread.Lstat(full)
			if err != nil {
				return fmt.Errorf("%s: inspect %s: %w", DiagUnmanagedConflict, inPlaceConflictPath(path), err)
			}
			if metadata.Kind == stateread.KindAbsent {
				continue
			}
			if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
				return fmt.Errorf("%s: inspect %s: path state is unreadable", DiagUnmanagedConflict, inPlaceConflictPath(path))
			}
			if metadata.Info.Mode()&os.ModeSymlink != 0 {
				link, err := stateread.Readlink(full)
				if err != nil || link.Kind != stateread.KindPresent {
					return fmt.Errorf("%s: inspect symlink %s: %v", DiagUnmanagedConflict, inPlaceConflictPath(path), err)
				}
				if !inPlaceLinkInStore(full, link.Target, contextstore.Root(home)) && !policy.Takeover {
					return fmt.Errorf("%s: %s is a symlink outside the manager store; abort or take over with backup", DiagForeignManager, inPlaceConflictPath(path))
				}
			}
			if !policy.Takeover {
				class := "owned-bytes"
				if metadata.Info.IsDir() {
					class = "owned-tree"
				}
				// A directory conflict proof may inspect only the exact skill leaf.
				// Neither a marker label nor a directory at AGENTS.md grants tree custody.
				scope := envunit.Unit{Native: []envunit.Native{{Adapter: adapter.ID, Root: native}}}
				if err := scope.ValidateReadScope(class, full); err != nil {
					return err
				}
				image, err := envunit.ObserveTarget(full, class)
				if err != nil {
					return err
				}
				return &globalInPlaceConflict{error: fmt.Errorf("%s: %s exists and no marker records it", DiagUnmanagedConflict, inPlaceConflictPath(path)), capture: envlifecycle.ConflictReceipt{Code: DiagUnmanagedConflict, Phase: "profile_in_place", Profile: profile, Adapter: adapter.ID, Target: full, Surface: path, Entry: image, Publication: "sha256:" + strings.TrimPrefix(hash, "sha256:"), Sequence: 1}}
			}
		}
	}
	return nil
}

func inPlaceConflictPath(path string) string {
	if strings.HasPrefix(path, profileSkillsSurfaceDir+"/") || strings.HasPrefix(path, "skills/") {
		return filepath.ToSlash(filepath.Join(path, "SKILL.md"))
	}
	return path
}

// loadMaterial loads the module bytes of every context member from its
// store entry, below the member directory when the lock records one.
func loadMaterial(home string, manager *gitManager, lock *contextlock.Lock) (map[string]contextmaterialize.Package, error) {
	packages := map[string]contextmaterialize.Package{}
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindContext {
			continue
		}
		entry := manager.entryPath(home, resolvedOf(member))
		// Spec §8 interim rule for v1 readers: refuse before the
		// lock-declared v1 materialized identities assembled from
		// these bytes are computed or trusted downstream
		// (surfaceHash, skill hashes, marker comparisons). The scan
		// covers the full store entry — including files the module
		// projection omits — for commit-pinned and state-pinned
		// members alike.
		if err := opaquescan.RefuseNULV1(entry, lock.ContentHashVersion()); err != nil {
			return nil, fmt.Errorf("member %s: %v", member.Name, err)
		}
		root := packageRoot(entry, member.Directory)
		manifest, err := contextpkg.LoadManifest(root)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		pkg := contextmaterialize.Package{HasContext: manifest.HasContext}
		for _, module := range manifest.Modules {
			payload, err := contextpkg.ReadModule(root, module)
			if err != nil {
				return nil, fmt.Errorf("module %s of %s: %v", module.Path, member.Name, err)
			}
			pkg.Modules = append(pkg.Modules, contextmaterialize.Module{Module: module, Bytes: payload})
		}
		packages[member.Name] = pkg
	}
	return packages, nil
}

// materializeOne writes one adapter home: backup, root-context surface,
// marker. The claude_code root-context surface is always a copied regular
// file (environments §8.1); every other surface links into the store with
// copy fallback.
// takeover carries the explicit section 9.5 takeover flag: with it the
// entry backs up and takes over the unmanaged files the switch would
// write, with the replace notice; without it the entry fails with
// environment_surface_unmanaged_conflict (or
// environment_foreign_manager_detected for a foreign-manager symlink)
// rather than overwrite.
func materializeOne(op *operation, home string, source Source, profile string, lock *contextlock.Lock, hash string, precedence contextmaterialize.Precedence, order []contextlock.Member, packages map[string]contextmaterialize.Package, skills []skillOf, adapter Adapter, policy Policy, nativeHomeOf func(string) (string, error)) (out EntryResult) {
	takeover := policy.Takeover
	native, err := resolveNativeHome(home, adapter, nativeHomeOf)
	if err != nil {
		return EntryResult{Adapter: adapter.ID, OK: false, Detail: err.Error()}
	}
	document, written, err := contextmaterialize.Monolithic(lock, hash, precedence, adapter.ID, packages)
	if err != nil {
		return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
	}
	if err := checkManagedPrivateTarget(native, envmarker.Name); err != nil {
		return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
	}
	prior, err := envmarker.Read(native)
	if err != nil {
		return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
	}
	recorded := map[string]bool{}
	if prior != nil {
		for _, surface := range prior.Surfaces {
			for _, path := range surface.Paths {
				recorded[path] = true
			}
		}
	}
	want := map[string]bool{}
	if written {
		want[adapter.Target] = true
	}
	ledger, err := inPlaceSkillsLedger(home, native, adapter, skills, recorded)
	if err != nil {
		return EntryResult{Adapter: adapter.ID, Home: native, Detail: err.Error()}
	}
	for path := range recorded {
		if _, err := nativeSurfacePath(home, native, path); err != nil {
			return EntryResult{Adapter: adapter.ID, Home: native, Detail: err.Error()}
		}
		if ledgerSkillPath(adapter, path, ledger) {
			delete(recorded, path)
		}
	}
	var ledgerDetails []string
	skillLinks := map[string]string{}
	skillHashes := map[string][]byte{}
	skillPaths := make([]string, 0, len(skills))
	for _, skill := range skills {
		path := filepath.ToSlash(filepath.Join(inPlaceSkillsDir(adapter.ID), skill.name))
		if ledgerSkillPath(adapter, path, ledger) {
			ledgerDetails = append(ledgerDetails, path+" belongs to legacy global scope (adapter ledger); left in place")
			continue
		}
		skillLinks[path] = skill.root
		skillHashes[path] = skillSurfaceIdentity(skill)
		skillPaths = append(skillPaths, path)
		want[path] = true
	}
	for path := range want {
		if _, err := nativeSurfacePath(home, native, path); err != nil {
			return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
		}
	}
	// Section 9.5 inventory: an unrecorded surface symlink pointing outside
	// the profile store is evidence of another manager and stops with
	// environment_foreign_manager_detected unless this operation explicitly
	// authorizes takeover. Takeover backs up the exact link entry (including
	// its link text) before replacement; §8.3.1 requires that backup and each
	// managed write to avoid following the link. Other unrecorded targets
	// fail with environment_surface_unmanaged_conflict without takeover.
	result := EntryResult{Adapter: adapter.ID, Home: native, Detail: strings.Join(ledgerDetails, "; "), Warnings: append([]string{}, ledgerDetails...)}
	var taken []string
	for path := range want {
		if recorded[path] {
			continue
		}
		full, err := nativeSurfacePath(home, native, path)
		if err != nil {
			result.OK = false
			result.Detail = err.Error()
			return result
		}
		state, err := stateread.Lstat(full)
		if err != nil {
			result.OK = false
			result.Detail = fmt.Sprintf("%s: inspect %s: %v", DiagUnmanagedConflict, inPlaceConflictPath(path), err)
			return result
		}
		if state.Kind == stateread.KindAbsent {
			continue
		}
		if state.Kind != stateread.KindPresent || state.Info == nil {
			result.OK = false
			result.Detail = fmt.Sprintf("%s: inspect %s: path state is unreadable", DiagUnmanagedConflict, inPlaceConflictPath(path))
			return result
		}
		info := state.Info
		if info.Mode()&os.ModeSymlink != 0 {
			link, readErr := stateread.Readlink(full)
			if readErr != nil || link.Kind != stateread.KindPresent {
				result.OK = false
				result.Detail = fmt.Sprintf("%s: inspect symlink %s: %v", DiagUnmanagedConflict, inPlaceConflictPath(path), readErr)
				return result
			}
			if !inPlaceLinkInStore(full, link.Target, contextstore.Root(home)) {
				if !takeover {
					result.OK = false
					result.Detail = DiagForeignManager + ": " + inPlaceConflictPath(path) + " is a symlink outside the manager store; abort, or take over with backup"
					return result
				}
				taken = append(taken, path+" (foreign-manager symlink)")
				continue
			}
		}
		if !takeover {
			result.OK = false
			result.Detail = DiagUnmanagedConflict + ": " + inPlaceConflictPath(path) + " exists and no marker records it"
			return result
		}
		taken = append(taken, path)
	}
	if len(taken) > 0 {
		sort.Strings(taken)
		result.Notice = "taking over unmanaged " + strings.Join(taken, ", ") + " for " + adapter.ID +
			": native global context files are being replaced by managed ones; backups land in " +
			filepath.Join(native, ".agent-environment-backup") + "/"
	}
	if hint := foreignManagerHint(); hint != "" && len(taken) > 0 {
		result.Warnings = append(result.Warnings, DiagForeignSuspect+": "+hint+" appears to manage this machine and will overwrite managed surfaces on its next apply")
	}
	if err := planNativePublication(op, home, native, profile, source, lock, hash, precedence, order, skills, adapter, prior, recorded, want, skillLinks, skillHashes, skillPaths, document, written); err != nil {
		result.Detail = err.Error()
		result.cause = err
		return result
	}
	result.OK = true
	return result
}

// removeManagedTree removes a managed surface path without following a
// symlink at the target or in any descendant. Directories are removed only
// after their entries have each passed the same managed-path checks.
func removeManagedTree(root, rel string) error {
	full, err := managedPath(root, rel, false)
	if err != nil {
		return err
	}
	state, err := stateread.Lstat(full)
	if err != nil {
		return err
	}
	if state.Kind == stateread.KindAbsent {
		return nil
	}
	if state.Kind != stateread.KindPresent || state.Info == nil {
		return fmt.Errorf("%s: inspect %s: path state is unreadable", DiagUnmanagedConflict, rel)
	}
	if state.Info.IsDir() && state.Info.Mode()&os.ModeSymlink == 0 {
		listing, err := stateread.ReadDir(full)
		if err != nil || listing.Kind != stateread.KindPresent {
			return fmt.Errorf("%s: inspect directory %s: %v", DiagUnmanagedConflict, rel, err)
		}
		for _, entry := range listing.Entries {
			if err := removeManagedTree(root, managedJoin(rel, entry.Name())); err != nil {
				return err
			}
		}
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// splitScope splits a scope key into environment and target.
func splitScope(scope string) (string, string) {
	if rest, ok := strings.CutPrefix(scope, "env:"); ok {
		return rest, ""
	}
	if rest, ok := strings.CutPrefix(scope, "target:"); ok {
		return "", rest
	}
	return "", ""
}

// markerKind maps the install kind onto the marker vocabulary.
func markerKind(source Source) string {
	switch source.Kind {
	case KindGit:
		return KindGit
	case KindPath:
		return KindPath
	default:
		return KindLocal
	}
}

// markerSource renders the core §6.1 canonical source identity for git
// roots (environments §1.3, §8.2). New install records already carry the
// canonical identity; old records carrying a raw URL canonicalize on the
// fly so the next switch heals the marker. A record no boundary accepts
// anymore (such as a pre-rejection file:// remote) passes through as
// written: no valid marker shape exists for it, and such records can only
// predate the rejection.
func markerSource(source Source) string {
	if source.Kind != KindGit {
		return ""
	}
	if source.Git == "" {
		return ""
	}
	canonical, err := canonicalGit(source.Git)
	if err != nil || canonical == "" {
		return source.Git
	}
	return canonical
}

// markerRequirement renders the declared requirement as written.
func markerRequirement(source Source) *envmarker.Requirement {
	if source.Kind != KindGit {
		return nil
	}
	requirement := &envmarker.Requirement{}
	switch {
	case source.Req.Range != "":
		requirement.Range = source.Req.Range
	case source.Req.Tag != "":
		requirement.Tag = source.Req.Tag
	default:
		requirement.Revision = source.Req.Revision
	}
	return requirement
}

// markerSourcePath records a path operand exactly as supplied: informative
// provenance whose bytes never enter any identity.
func markerSourcePath(source Source) string {
	if source.Kind == KindPath {
		return source.Path
	}
	return ""
}

// resolvedOf rebuilds the store lookup of a lock member.
func resolvedOf(member contextlock.Member) contextresolve.Resolved {
	return contextresolve.Resolved{
		Kind: member.Kind, Name: member.Name, Source: member.Source,
		Directory: member.Directory, Commit: member.Commit, StateHash: member.StateHash,
	}
}

// _ keeps identifiers referenced for the scope-key grammar.
var _ = identifiers.Valid

// _ keeps contextstore referenced for entry reuse checks.
var _ = contextstore.Exists
