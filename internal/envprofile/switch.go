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
// journal. The per-entry agent-home payloads are not transaction targets in
// this stage: one Plan commits all-or-nothing with rollback, while §9.2
// requires partial success persisted, and the engine's sidecar backups
// cannot produce the specified §8.3 versioned generations. Entries stay
// direct writes under the held lock; re-running profile use converges the
// scope from the lock, which is the specified recovery path.
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
// in-place homes only. The skills tree and MCP files are likewise stage
// (b): the recorded surfaces are the root-context file every adapter owns.
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
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
)

// Diagnostics for switching (environments §8.5, §9.7).
const (
	DiagUsePartial         = "profile_use_partial"
	DiagUnmanagedConflict  = "environment_surface_unmanaged_conflict"
	DiagBackupExists       = "environment_backup_exists"
	DiagUnknownEnvironment = "environment_unknown"
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

// NativeHome resolves the adapter's native default home. The mechanism
// variable names the home, except for opencode where it names the XDG
// parent and the tool reads the opencode child (environments §7.1). The
// four defaults below the user home are implementation choices the
// specification's mechanism implies but does not spell; every test pins
// homes through the variables instead.
func NativeHome(adapter Adapter) (string, error) {
	if adapter.ID == "opencode" {
		if value := os.Getenv(adapter.EnvVar); value != "" {
			return filepath.Join(value, "opencode"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("no home directory for %s: set %s", adapter.ID, adapter.EnvVar)
		}
		return filepath.Join(home, ".config", "opencode"), nil
	}
	if value := os.Getenv(adapter.EnvVar); value != "" {
		return value, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no home directory for %s: set %s", adapter.ID, adapter.EnvVar)
	}
	return filepath.Join(home, adapter.DefaultDir), nil
}

// EntryResult is the per-adapter outcome of a switch.
type EntryResult struct {
	Adapter string
	Home    string
	OK      bool
	Detail  string
}

// Use switches the machine scope (environment and target both empty) or one
// narrowed scope to name. A machine-scope switch skips adapters that carry
// a scope record: those homes stay on their scoped profiles. With clear,
// the scope record is dropped and the scope re-materializes from the
// machine default. It returns the per-entry results; when any entry failed
// the recorded current is unchanged and the error carries
// profile_use_partial.
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
func useLocked(op *operation, home, name, environment, target string, clearScope bool, policy Policy) ([]EntryResult, error) {
	if err := ensureDefault(op, home, policy); err != nil {
		return nil, err
	}
	if environment != "" {
		if _, ok := adapterByID(environment); !ok {
			return nil, fmt.Errorf("%s: explicit operand names the unregistered environment %q", DiagUnknownEnvironment, environment)
		}
	}
	if target != "" {
		if _, err := envregistry.TargetByID(target); err != nil {
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
		machine, err := Current(home)
		if err != nil {
			return nil, err
		}
		if machine == "" {
			machine = DefaultProfile
		}
		effective = machine
	} else {
		if _, err := readSource(home, name); err != nil {
			return nil, err
		}
	}
	results, err := materializeScope(home, effective, environment)
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
		return results, fmt.Errorf("%s: the scope is partially switched; the recorded current is unchanged", DiagUsePartial)
	}
	// The recorded current moves only after the whole scope materialized,
	// and it moves through the operation journal, so a crash here recovers
	// to either the old current (journal resumes) or the new one — never a
	// half-written pointer. A journal failure after switched entries is
	// itself a partial switch: the entries no longer match the recorded
	// current.
	if scope == "" {
		if err := op.publish(map[string][]byte{CurrentFile(home): []byte(effective + "\n")}); err != nil {
			return results, fmt.Errorf("%s: the scope is partially switched; the recorded current is unchanged", DiagUsePartial)
		}
	} else {
		machine, err := Current(home)
		if err != nil {
			return results, err
		}
		if clearScope || effective == machine {
			// A scope-clear is one unlink under the held lock: crash
			// before it retries the clear, crash after it has converged.
			// It is not journaled as an absence target in this stage.
			if err := SetScoped(home, scope, "", true); err != nil {
				return results, err
			}
		} else if err := op.publish(map[string][]byte{filepath.Join(ScopedDir(home), scopeFileName(scope)): []byte(effective + "\n")}); err != nil {
			return results, fmt.Errorf("%s: the scope is partially switched; the recorded current is unchanged", DiagUsePartial)
		}
	}
	return results, nil
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
func SyncWithPolicy(home string, policy Policy) ([]EntryResult, error) {
	op, err := beginOperation(home)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()
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
	var results []EntryResult
	machineResults, err := materializeScope(home, machine, "")
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
		scopeResults, err := materializeScope(home, scoped[scope], environment)
		if err != nil {
			return append(results, scopeResults...), err
		}
		results = append(results, scopeResults...)
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
func materializeScope(home, profile, environment string) ([]EntryResult, error) {
	lock, hash, err := readLock(home, profile)
	if err != nil {
		return nil, err
	}
	source, err := readSource(home, profile)
	if err != nil {
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
	precedence := contextmaterialize.DefaultPrecedence
	order, err := contextmaterialize.EmittedOrder(lock, precedence)
	if err != nil {
		return nil, err
	}
	var results []EntryResult
	for _, adapter := range adapters {
		results = append(results, materializeOne(home, source, profile, lock, hash, precedence, order, packages, adapter))
	}
	return results, nil
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
		root := packageRoot(entry, member.Directory)
		manifest, err := contextpkg.LoadManifest(root)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		pkg := contextmaterialize.Package{HasContext: manifest.HasContext}
		for _, module := range manifest.Modules {
			payload, err := os.ReadFile(filepath.Join(root, contextpkg.ContextDir, filepath.FromSlash(module.Path))) // #nosec G304 -- module of the store entry
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
func materializeOne(home string, source Source, profile string, lock *contextlock.Lock, hash string, precedence contextmaterialize.Precedence, order []contextlock.Member, packages map[string]contextmaterialize.Package, adapter Adapter) EntryResult {
	native, err := NativeHome(adapter)
	if err != nil {
		return EntryResult{Adapter: adapter.ID, OK: false, Detail: err.Error()}
	}
	document, written, err := contextmaterialize.Monolithic(lock, hash, precedence, adapter.ID, packages)
	if err != nil {
		return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
	}
	if err := os.MkdirAll(native, 0o755); err != nil {
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
	target := filepath.Join(native, adapter.Target)
	want := map[string]bool{}
	if written {
		want[adapter.Target] = true
	}
	// A write that would touch a file no marker records fails the entry and
	// never overwrites (manager profile §12.2): without the takeover flag
	// the ledger discipline fails the operation rather than overwrite.
	// Takeover and onboarding import are a later stage.
	for path := range want {
		if !recorded[path] {
			if _, err := os.Lstat(filepath.Join(native, path)); err == nil {
				return EntryResult{Adapter: adapter.ID, Home: native, OK: false,
					Detail: DiagUnmanagedConflict + ": " + path + " exists and no marker records it"}
			}
		}
	}
	generation := 0
	if len(want) > 0 {
		needsBackup := false
		for path := range want {
			if _, err := os.Lstat(filepath.Join(native, path)); err == nil {
				needsBackup = true
			}
		}
		if needsBackup {
			generation, err = openBackup(native, want)
			if err != nil {
				return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
			}
		}
	}
	// Remove recorded files the new profile no longer wants.
	if prior != nil {
		for path := range recorded {
			if !want[path] {
				_ = os.Remove(filepath.Join(native, path))
			}
		}
	}
	surface := envmarker.Surface{Paths: []string{}, Form: "monolithic"}
	copies := []envmarker.Copy{}
	if written {
		if adapter.ID == "claude_code" {
			if err := os.WriteFile(target, document, 0o644); err != nil {
				return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
			}
			copies = append(copies, envmarker.Copy{Path: adapter.Target, Reason: envmarker.ReasonClaudeCodeRootContext})
		} else {
			storeFile, err := writeStoreDocument(home, profile, adapter, document)
			if err != nil {
				return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
			}
			if err := replaceLink(target, storeFile); err != nil {
				// Copy fallback (manager §5): record the copy with its reason.
				if writeErr := os.WriteFile(target, document, 0o644); writeErr != nil {
					return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: writeErr.Error()}
				}
				copies = append(copies, envmarker.Copy{Path: adapter.Target, Reason: envmarker.ReasonSymlinkFallback})
			}
		}
		surface.Paths = []string{adapter.Target}
		surface.ContentSHA256 = contextmaterialize.SurfaceHash(map[string][]byte{adapter.Target: document})
	}
	surface.Copies = &copies
	marker := &envmarker.Marker{
		Version: 1,
		Profile: envmarker.Profile{
			Name: profile, Root: lock.Root, Kind: markerKind(source),
			LockSHA256: strings.TrimPrefix(hash, "sha256:"),
			Source:     markerSource(source), Requirement: markerRequirement(source),
			Directory: source.Directory, SourcePath: markerSourcePath(source),
		},
		Precedence: envmarker.Precedence{Winner: precedence.Winner, Placement: precedence.Placement},
		Mode:       envmarker.ModeLinked,
		Surfaces:   map[string]envmarker.Surface{},
	}
	for _, member := range order {
		entry := envmarker.Member{Name: member.Name, Version: member.Version, Weight: member.Weight, Overlay: member.Overlay}
		if member.Commit != "" {
			entry.Commit = member.Commit
		} else {
			entry.StateSHA256 = member.StateHash
		}
		marker.Members = append(marker.Members, entry)
	}
	if written {
		marker.Surfaces[envmarker.SurfaceRootContext] = surface
	}
	payload, err := marker.Marshal()
	if err != nil {
		return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
	}
	if err := os.WriteFile(filepath.Join(native, envmarker.Name), payload, 0o644); err != nil {
		return EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}
	}
	_ = generation
	return EntryResult{Adapter: adapter.ID, Home: native, OK: true}
}

// openBackup copies every file the operation will replace into the next
// versioned generation and prunes beyond the retention count. A next
// generation that already exists fails with environment_backup_exists.
func openBackup(native string, want map[string]bool) (int, error) {
	base := filepath.Join(native, ".agent-environment-backup")
	entries, err := os.ReadDir(base)
	highest := 0
	if err == nil {
		for _, entry := range entries {
			var n int
			if _, err := fmt.Sscanf(entry.Name(), "%d", &n); err == nil && n > highest {
				highest = n
			}
		}
	} else if !os.IsNotExist(err) {
		return 0, err
	}
	next := highest + 1
	generation := filepath.Join(base, fmt.Sprintf("%d", next))
	if _, err := os.Lstat(generation); err == nil {
		return 0, fmt.Errorf("%s: generation %d exists", DiagBackupExists, next)
	}
	if err := os.MkdirAll(generation, 0o755); err != nil {
		return 0, err
	}
	paths := make([]string, 0, len(want))
	for path := range want {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		full := filepath.Join(native, path)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		payload, err := os.ReadFile(full) // #nosec G304 -- recorded file of the home
		if err != nil {
			_ = os.RemoveAll(generation)
			return 0, err
		}
		target := filepath.Join(generation, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			_ = os.RemoveAll(generation)
			return 0, err
		}
		if err := os.WriteFile(target, payload, 0o644); err != nil {
			_ = os.RemoveAll(generation)
			return 0, err
		}
	}
	pruneBackups(base, highest+1)
	return next, nil
}

// pruneBackups removes the oldest generations beyond the retention count.
// Backups are otherwise never removed: never modified, never collected.
func pruneBackups(base string, newest int) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return
	}
	var generations []int
	for _, entry := range entries {
		var n int
		if _, err := fmt.Sscanf(entry.Name(), "%d", &n); err == nil {
			generations = append(generations, n)
		}
	}
	sort.Ints(generations)
	for len(generations) > BackupRetention {
		oldest := generations[0]
		generations = generations[1:]
		_ = os.RemoveAll(filepath.Join(base, fmt.Sprintf("%d", oldest)))
	}
	_ = newest
}

// writeStoreDocument stores the rendered document for link targets.
func writeStoreDocument(home, profile string, adapter Adapter, document []byte) (string, error) {
	dir := filepath.Join(ProfilesDir(home), profile, "rendered", adapter.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, adapter.Target)
	if err := os.WriteFile(path, document, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// replaceLink swaps target to a symlink of storeFile, removing whatever the
// marker recorded there before.
func replaceLink(target, storeFile string) error {
	_ = os.Remove(target)
	return os.Symlink(storeFile, target)
}

// purgeHomes removes the in-place surfaces, markers, and backup generations
// of every home whose marker names profile, and the profile's managed
// homes with their markers and backups after the notice (environments
// §9.2). Retained homes without a profile are orphans env status reports.
func purgeHomes(home, profile string) error {
	for _, adapter := range Adapters {
		native, err := NativeHome(adapter)
		if err != nil {
			continue
		}
		marker, err := envmarker.Read(native)
		if err != nil || marker == nil || marker.Profile.Name != profile {
			continue
		}
		for _, surface := range marker.Surfaces {
			for _, path := range surface.Paths {
				_ = os.Remove(filepath.Join(native, path))
			}
		}
		_ = os.Remove(filepath.Join(native, envmarker.Name))
		_ = os.RemoveAll(filepath.Join(native, ".agent-environment-backup"))
	}
	// Managed homes hold the operator's session data and are retained
	// unless purged; a purge removes the profile's environments directory
	// with every home, marker, and backup generation. The profile name is
	// the only profile-derived component below the environments root, so
	// the directory holds exactly this profile's homes.
	_ = os.RemoveAll(filepath.Join(EnvRoot(home), profile))
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
