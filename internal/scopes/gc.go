package scopes

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/stateread"
)

// CollectRuntime removes runtime store entries referenced by no marker of
// any registered consumer, the global store, or the hybrid store, and prunes
// consumer entries whose checkout disappeared or holds no markers
// (Spec §8.7). Returns the removed runtime entries as skill/commit pairs.
//
// It collects the runtime store only. Collect is the locked maintenance entry
// point that also sweeps the protected build cache.
func CollectRuntime(home string) ([]string, error) {
	if projectlease.HasState(home) {
		return nil, fmt.Errorf("project leases require serialized Collect; runtime sweep skipped")
	}
	marked := markScopes(home)
	if err := pruneConsumers(home, marked); err != nil {
		return nil, err
	}
	if len(marked.uncertain) > 0 {
		return nil, fmt.Errorf("runtime sweep skipped: the live reference set could not be proven complete: %s", strings.Join(marked.uncertain, "; "))
	}
	return sweepRuntime(home, marked.runtime)
}

// HomeLock witnesses the caller-held manager-home mutation lock. Maintenance
// never acquires it: install, rollback, recovery, and standalone `curator gc`
// all serialize on the one lock their caller already holds, so a consumer
// update can neither be lost nor be observed half-written.
type HomeLock = buildcache.HomeLock

// BuildCache is the protected immutable build-cache boundary a sweep uses.
type BuildCache interface {
	Sweep(request buildcache.SweepRequest, lock buildcache.HomeLock) (buildcache.SweepResult, error)
}

// MaintenanceRequest is one serialized maintenance pass over a manager home.
type MaintenanceRequest struct {
	Home string
	// Lock is the already-held manager-home mutation lock.
	Lock HomeLock
	// JournalKeys are the build keys named by every recoverable transaction
	// journal, so an in-flight installation keeps the artifacts it depends on.
	JournalKeys []string
	// Cache overrides the protected build cache; nil resolves the real store.
	Cache BuildCache
	// Now and Grace override the sweep clock and retention window. Their zero
	// values mean "now" and buildcache.DefaultGrace.
	Now   time.Time
	Grace time.Duration
	// DryRun reports all removal candidates without changing the manager home.
	DryRun bool
}

// MaintenanceResult reports what one pass removed and what it could not prove.
type MaintenanceResult struct {
	RemovedRuntime             []string
	RemovedBuilds              []string
	RemovedSkillSnapshots      []string
	RemovedSourceSnapshots     []string
	WouldRemoveRuntime         []string
	WouldRemoveBuilds          []string
	WouldRemoveSkillSnapshots  []string
	WouldRemoveSourceSnapshots []string
	Warnings                   []string
}

// Collect runs one serialized maintenance pass: it marks every live reference
// once, prunes provably dead consumers, sweeps the runtime store, and sweeps
// the protected build cache.
//
// Marking reads every supported marker schema. Runtime entries are marked from
// marker v1 and marker v2 alike, because a valid schema-1 installation stays
// current; logical build keys come only from marker v2 entries that carry them
// and from in-flight journals.
//
// Uncertainty fails safe, and it stays that way across passes. A consumer
// registry, skill scope, or marker that exists but cannot be trusted leaves the
// reference set unprovable, so neither runtime nor build caches are swept. Crucially,
// nothing that could still hold a reference is forgotten either: an unreadable
// registry is never rewritten, and a consumer whose scope could not be proven
// empty stays registered. A second pass therefore sees the same uncertainty and
// refuses the same sweep, instead of inheriting a registry that was quietly
// emptied by the first.
func Collect(request MaintenanceRequest) (MaintenanceResult, error) {
	if request.Home == "" {
		return MaintenanceResult{}, fmt.Errorf("manager home is empty")
	}
	if err := requireHomeLock(request.Lock); err != nil {
		return MaintenanceResult{}, err
	}
	leaseRoots, err := projectlease.ReferencesUnderLock(request.Home, request.Lock)
	if err != nil {
		return MaintenanceResult{}, fmt.Errorf("project lease references unreadable; runtime and build sweeps skipped: %w", err)
	}
	owner, ok := request.Lock.(interface{ Held() managerlock.Held })
	if !ok {
		return MaintenanceResult{}, managerlock.ErrNotHeld
	}
	witness := owner.Held()
	check := func() error { return witness.Validate(request.Home) }
	marked := markScopes(request.Home)
	for _, entry := range leaseRoots.RuntimeEntries {
		marked.runtime[entry] = true
	}
	result := MaintenanceResult{Warnings: marked.warnings()}

	// Consumer pruning stays inside this pass, under the caller's lock.
	if err := check(); err != nil {
		return result, err
	}
	if !request.DryRun {
		if err := pruneConsumersChecked(request.Home, marked, check); err != nil {
			return result, err
		}
	}
	if len(marked.uncertain) > 0 {
		result.Warnings = append(result.Warnings,
			"runtime sweep skipped: the live reference set could not be proven complete",
			"build cache sweep skipped: the live reference set could not be proven complete")
		return result, nil
	}
	if request.DryRun {
		result.WouldRemoveRuntime, err = previewRuntimeChecked(request.Home, marked.runtime, check)
	} else {
		result.RemovedRuntime, err = sweepRuntimeChecked(request.Home, marked.runtime, check)
	}
	if err != nil {
		return result, err
	}

	cache := request.Cache
	if cache == nil {
		store, storeErr := buildcache.New(request.Home)
		if storeErr != nil {
			return result, fmt.Errorf("open the protected build cache: %w", storeErr)
		}
		cache = store
	}
	referenced := append(append([]string(nil), marked.builds...), request.JournalKeys...)
	referenced = append(referenced, leaseRoots.BuildKeys...)
	if err := check(); err != nil {
		return result, err
	}
	sweep, err := cache.Sweep(buildcache.SweepRequest{
		Referenced: referenced,
		Now:        request.Now,
		Grace:      request.Grace,
		DryRun:     request.DryRun,
	}, maintenanceLock{check: check})
	result.RemovedBuilds = sweep.Removed
	result.WouldRemoveBuilds = append(result.WouldRemoveBuilds, sweep.WouldRemove...)
	result.Warnings = append(result.Warnings, sweep.Warnings...)
	if err != nil {
		return result, fmt.Errorf("sweep the protected build cache: %w", err)
	}
	if err := check(); err != nil {
		return result, err
	}
	var externalRemoved []string
	var externalErr error
	if request.DryRun {
		externalRemoved, externalErr = buildrepo.PreviewChecked(filepath.Join(request.Home, "external-build-cache"), referenced, request.Now, request.Grace, check)
		result.WouldRemoveBuilds = append(result.WouldRemoveBuilds, externalRemoved...)
	} else {
		externalRemoved, externalErr = buildrepo.CollectChecked(filepath.Join(request.Home, "external-build-cache"), referenced, request.Now, request.Grace, check)
		result.RemovedBuilds = append(result.RemovedBuilds, externalRemoved...)
	}
	if externalErr != nil {
		return result, fmt.Errorf("sweep the external build cache: %w", externalErr)
	}
	if len(marked.snapshotUncertain) > 0 {
		result.Warnings = append(result.Warnings, "skill snapshot sweep skipped: installed lock references could not be proven complete")
		return result, nil
	}
	result.RemovedSkillSnapshots, result.WouldRemoveSkillSnapshots, err = sweepLegacySkillSnapshots(
		request.Home, marked.skillSnapshots, maintenanceLock{check: check}, request.DryRun)
	if err != nil {
		return result, fmt.Errorf("sweep legacy skill snapshots: %w", err)
	}
	result.RemovedSourceSnapshots, result.WouldRemoveSourceSnapshots, err = sweepSourceSnapshots(
		request.Home, marked.sourceSnapshots, marked.skillSnapshots, maintenanceLock{check: check}, request.DryRun)
	if err != nil {
		return result, fmt.Errorf("sweep source snapshots: %w", err)
	}
	return result, nil
}

// pruneConsumers rewrites the registry with the consumers one pass proved are
// still worth keeping. A registry that could not be trusted is left exactly as
// it is: overwriting it would destroy the only record of which checkouts may
// still reference a build artifact.
func pruneConsumers(home string, marked marks) error {
	return pruneConsumersChecked(home, marked, func() error { return nil })
}

func pruneConsumersChecked(home string, marked marks, check func() error) error {
	if err := check(); err != nil {
		return err
	}
	if marked.registryUnknown {
		return nil
	}
	return replaceConsumersChecked(home, marked.consumers, check)
}

func requireHomeLock(lock HomeLock) error {
	if lock == nil {
		return fmt.Errorf("caller-held manager-home lock is required")
	}
	if err := lock.AssertHeld(); err != nil {
		return fmt.Errorf("manager-home lock is not held: %w", err)
	}
	return nil
}

// marks is one complete traversal of every live install scope.
type marks struct {
	consumers       []string
	runtime         map[string]bool
	builds          []string
	skillSnapshots  map[string]bool
	sourceSnapshots map[string]bool
	// uncertain describes state that exists but could not be trusted, so the
	// reference set derived from it is incomplete. Any entry here blocks the
	// runtime and build sweeps and protects every consumer it came from.
	uncertain []string
	// snapshotUncertain blocks only snapshot cleanup. The existing runtime and
	// build sweeps have independent reference formats and can remain useful.
	snapshotUncertain []string
	// notes describe state that was ignored but cannot hide a reference, so
	// they are reported without blocking maintenance.
	notes []string
	// registryUnknown records that the consumer registry exists but could not
	// be trusted, so it must not be rewritten.
	registryUnknown bool
}

func (marked marks) warnings() []string {
	warnings := append([]string(nil), marked.uncertain...)
	warnings = append(warnings, marked.snapshotUncertain...)
	return append(warnings, marked.notes...)
}

func markScopes(home string) marks {
	marked := marks{runtime: map[string]bool{}, skillSnapshots: map[string]bool{}, sourceSnapshots: map[string]bool{}}
	consumers, err := readConsumers(home)
	if err != nil {
		marked.registryUnknown = true
		marked.uncertain = append(marked.uncertain, fmt.Sprintf(
			"consumer registry %s cannot be trusted: %v; no consumer was pruned and no build entry was swept",
			filepath.Join(home, ConsumersName), err))
	}
	for _, projectRoot := range consumers {
		markSourceLock(projectRoot, &marked)
		scope := readScope(filepath.Join(projectRoot, ".agents", "skills"))
		marked.uncertain = append(marked.uncertain, scope.uncertain...)
		marked.notes = append(marked.notes, scope.notes...)
		// A consumer is dropped only once its scope is proven absent or valid
		// and empty. While anything about it is uncertain it stays registered,
		// so a later pass still visits the markers that may name a build key.
		if len(scope.markers) == 0 && len(scope.uncertain) == 0 {
			continue // provably dead consumer: gone or empty
		}
		marked.consumers = append(marked.consumers, projectRoot)
		marked.absorb(scope)
	}
	for _, scopeDir := range []string{
		filepath.Join(home, "global", "skills"),
		filepath.Join(home, "hybrid", "skills"),
	} {
		scope := readScope(scopeDir)
		marked.uncertain = append(marked.uncertain, scope.uncertain...)
		marked.notes = append(marked.notes, scope.notes...)
		marked.absorb(scope)
	}
	for _, root := range []string{filepath.Join(home, "global"), filepath.Join(home, "hybrid")} {
		markSourceLock(root, &marked)
	}
	markProfileSkillLocks(home, &marked)
	return marked
}

func (marked *marks) absorb(scope scopeMarks) {
	if marked.skillSnapshots == nil {
		marked.skillSnapshots = map[string]bool{}
	}
	if marked.sourceSnapshots == nil {
		marked.sourceSnapshots = map[string]bool{}
	}
	for _, installed := range scope.markers {
		if identifiers.Valid(installed.Name) && contextpkg.ValidCommit(installed.Commit) {
			marked.skillSnapshots[installed.Name+"/"+installed.Commit] = true
		}
		if installed.Source != "" && contextpkg.ValidCommit(installed.Commit) {
			marked.sourceSnapshots[sourceSnapshotKey(installed.Source, installed.Commit)] = true
		}
		if installed.Package != nil {
			// A draft marker keys its runtime tree by the frozen package
			// identity in the source-v1 namespace, never by a commit: mark
			// that leaf so the sweep keeps the live tree. An unhashable
			// package cannot name its tree, so it is reported instead of
			// silently unmarked.
			digest, digestErr := installed.Package.Digest()
			key, keyErr := runtimestore.SourceV1Key(digest)
			if digestErr != nil || keyErr != nil {
				marked.uncertain = append(marked.uncertain, fmt.Sprintf(
					"installed skill %s names a draft package with no derivable runtime key; its runtime tree was left alone",
					installed.Name))
				continue
			}
			marked.runtime[installed.Name+"/"+key] = true
		} else {
			marked.runtime[installed.Name+"/"+installed.Commit] = true
		}
		// Every build-bearing schema contributes its live cache keys. A schema
		// missing from this band is not "no builds": its recorded keys would go
		// unmarked and the collector would delete artifacts an installation is
		// still using.
		if !marker.BuildBearingSchema(installed.SchemaVersion) {
			continue
		}
		for _, build := range installed.Builds {
			marked.builds = append(marked.builds, string(build.CacheKey))
		}
	}
}

func markProfileSkillLocks(home string, marked *marks) {
	root := filepath.Join(home, "profiles")
	profiles, err := stateread.ReadDir(root)
	if err != nil {
		marked.snapshotUncertain = append(marked.snapshotUncertain, fmt.Sprintf("profile locks %s cannot be listed: %v", root, err))
		return
	}
	if profiles.Kind == stateread.KindAbsent {
		return
	}
	for _, entry := range profiles.Entries {
		profileDir := filepath.Join(root, entry.Name())
		state, err := stateread.Lstat(profileDir)
		if err != nil {
			marked.snapshotUncertain = append(marked.snapshotUncertain, fmt.Sprintf("profile lock directory %s cannot be inspected: %v", profileDir, err))
			continue
		}
		if state.Kind != stateread.KindPresent {
			continue
		}
		if state.Info == nil || state.Info.Mode()&os.ModeSymlink != 0 {
			marked.snapshotUncertain = append(marked.snapshotUncertain, fmt.Sprintf("profile lock directory %s is not a plain directory", profileDir))
			continue
		}
		if !state.Info.IsDir() {
			continue
		}
		for _, name := range []string{"lock.json", "lock.prev.json"} {
			path := filepath.Join(profileDir, name)
			lock, _, readErr := contextlock.ReadWith(path, stateread.ReadRegularFile)
			if errors.Is(readErr, os.ErrNotExist) {
				continue
			}
			if readErr != nil {
				marked.snapshotUncertain = append(marked.snapshotUncertain, fmt.Sprintf("profile lock %s cannot be trusted: %v", path, readErr))
				continue
			}
			for _, member := range lock.Members {
				if member.Kind != contextlock.KindSkill || !contextpkg.ValidCommit(member.Commit) {
					continue
				}
				marked.skillSnapshots[member.Name+"/"+member.Commit] = true
				if member.Source != "" {
					marked.sourceSnapshots[sourceSnapshotKey(member.Source, member.Commit)] = true
				}
			}
		}
	}
}

func markSourceLock(root string, marked *marks) {
	path := sourcelock.PathIn(root)
	file, err := stateread.ReadRegularFile(path)
	if err != nil {
		marked.snapshotUncertain = append(marked.snapshotUncertain, fmt.Sprintf("source lock %s cannot be read: %v", path, err))
		return
	}
	if file.Kind == stateread.KindAbsent {
		return
	}
	lock, err := sourcelock.Parse(file.Bytes)
	if err != nil {
		marked.snapshotUncertain = append(marked.snapshotUncertain, fmt.Sprintf("source lock %s cannot be trusted: %v", path, err))
		return
	}
	for _, member := range lock.Members {
		pkg := member.Package
		source := ""
		switch pkg.Kind {
		case sourcelock.KindNetworkGit:
			source = pkg.Repository
		case sourcelock.KindConfiguredGit:
			source = pkg.Source
		}
		if source != "" && contextpkg.ValidCommit(pkg.Commit.Hex) {
			marked.sourceSnapshots[sourceSnapshotKey(source, pkg.Commit.Hex)] = true
		}
	}
}

func sourceSnapshotKey(source, commit string) string { return source + "\x00" + commit }

const retiredSourceSnapshotPrefix = ".curator-source-snapshot-gc-"

type snapshotCandidate struct {
	label string
	path  string
}

func sweepLegacySkillSnapshots(home string, referenced map[string]bool, lock HomeLock, dryRun bool) (removed, wouldRemove []string, err error) {
	root := filepath.Join(contextstore.Root(home), "skill")
	listing, err := stateread.ReadDir(root)
	if err != nil {
		return nil, nil, err
	}
	if listing.Kind == stateread.KindAbsent {
		return nil, nil, nil
	}
	var candidates []snapshotCandidate
	for _, skill := range listing.Entries {
		if !identifiers.Valid(skill.Name()) {
			continue
		}
		skillDir := filepath.Join(root, skill.Name())
		skillState, stateErr := stateread.Lstat(skillDir)
		if stateErr != nil {
			return nil, nil, stateErr
		}
		if skillState.Kind != stateread.KindPresent || skillState.Info == nil || skillState.Info.Mode()&os.ModeSymlink != 0 || !skillState.Info.IsDir() {
			continue
		}
		commits, readErr := stateread.ReadDir(skillDir)
		if readErr != nil {
			return nil, nil, readErr
		}
		for _, commit := range commits.Entries {
			if !contextpkg.ValidCommit(commit.Name()) || referenced[skill.Name()+"/"+commit.Name()] {
				continue
			}
			entry := filepath.Join(skillDir, commit.Name())
			entryState, stateErr := stateread.Lstat(entry)
			if stateErr != nil {
				return nil, nil, stateErr
			}
			if entryState.Kind != stateread.KindPresent || entryState.Info == nil || entryState.Info.Mode()&os.ModeSymlink != 0 || !entryState.Info.IsDir() {
				continue
			}
			if boundaryErr := pathboundary.ValidateWithin(root, entry); boundaryErr != nil {
				return nil, nil, fmt.Errorf("refuse snapshot outside the skill store: %w", boundaryErr)
			}
			candidates = append(candidates, snapshotCandidate{label: skill.Name() + "/" + commit.Name(), path: entry})
		}
	}
	for _, candidate := range candidates {
		if dryRun {
			wouldRemove = append(wouldRemove, candidate.label)
			continue
		}
		if err := lock.AssertHeld(); err != nil {
			return removed, wouldRemove, err
		}
		if err := pathboundary.ValidateWithin(root, candidate.path); err != nil {
			return removed, wouldRemove, fmt.Errorf("refuse snapshot outside the skill store: %w", err)
		}
		if err := os.RemoveAll(candidate.path); err != nil {
			return removed, wouldRemove, err
		}
		removed = append(removed, candidate.label)
	}
	sort.Strings(removed)
	sort.Strings(wouldRemove)
	return removed, wouldRemove, nil
}

func sweepSourceSnapshots(home string, sourceRefs, skillRefs map[string]bool, lock HomeLock, dryRun bool) (removed, wouldRemove []string, err error) {
	return sweepSourceSnapshotsWithRemove(home, sourceRefs, skillRefs, lock, dryRun, os.RemoveAll)
}

func sweepSourceSnapshotsWithRemove(home string, sourceRefs, skillRefs map[string]bool, lock HomeLock, dryRun bool, remove func(string) error) (removed, wouldRemove []string, err error) {
	root := filepath.Join(home, "cache")
	listing, err := stateread.ReadDir(root)
	if err != nil {
		return nil, nil, err
	}
	if listing.Kind == stateread.KindAbsent {
		return nil, nil, nil
	}
	var candidates []snapshotCandidate
	var walk func(path, source string, depth int) error
	walk = func(path, source string, depth int) error {
		if depth > 32 {
			return fmt.Errorf("source cache nesting exceeds the supported bound")
		}
		entries, readErr := stateread.ReadDir(path)
		if readErr != nil {
			return readErr
		}
		if entries.Kind == stateread.KindAbsent {
			return nil
		}
		for _, entry := range entries.Entries {
			if depth == 0 && (entry.Name() == "build" || entry.Name() == "registry" || entry.Name() == "local-store") {
				continue
			}
			child := filepath.Join(path, entry.Name())
			state, stateErr := stateread.Lstat(child)
			if stateErr != nil {
				return stateErr
			}
			if state.Kind != stateread.KindPresent || state.Info == nil || state.Info.Mode()&os.ModeSymlink != 0 || !state.Info.IsDir() {
				continue
			}
			childSource := entry.Name()
			if source != "" {
				childSource = filepath.ToSlash(filepath.Join(source, entry.Name()))
			}
			if contextpkg.ValidCommit(entry.Name()) && (identity.ValidCanonical(source) || (strings.Count(source, "/") == 0 && identifiers.Valid(source))) {
				retiredEntries, retiredErr := stateread.ReadDir(child)
				if retiredErr != nil {
					return retiredErr
				}
				for _, retiredEntry := range retiredEntries.Entries {
					if !strings.HasPrefix(retiredEntry.Name(), retiredSourceSnapshotPrefix) {
						continue
					}
					retiredPath := filepath.Join(child, retiredEntry.Name())
					retiredState, stateErr := stateread.Lstat(retiredPath)
					if stateErr != nil {
						return stateErr
					}
					if retiredState.Kind == stateread.KindAbsent {
						continue
					}
					if retiredState.Info == nil || retiredState.Info.Mode()&os.ModeSymlink != 0 || !retiredState.Info.IsDir() {
						return fmt.Errorf("refuse invalid retired source snapshot %s", retiredPath)
					}
					if boundaryErr := pathboundary.ValidateWithin(root, retiredPath); boundaryErr != nil {
						return fmt.Errorf("refuse retired source snapshot outside the cache store: %w", boundaryErr)
					}
					candidates = append(candidates, snapshotCandidate{
						label: "retired:" + filepath.ToSlash(filepath.Join(source, entry.Name(), retiredEntry.Name())), path: retiredPath,
					})
				}
				snapshotPath := filepath.Join(child, "snapshot")
				snapshotState, snapshotErr := stateread.Lstat(snapshotPath)
				if snapshotErr != nil {
					return snapshotErr
				}
				if snapshotState.Kind == stateread.KindAbsent {
					continue
				}
				if snapshotState.Info == nil || snapshotState.Info.Mode()&os.ModeSymlink != 0 || !snapshotState.Info.IsDir() {
					return fmt.Errorf("refuse non-directory source snapshot %s", snapshotPath)
				}
				if boundaryErr := pathboundary.ValidateWithin(root, snapshotPath); boundaryErr != nil {
					return fmt.Errorf("refuse source snapshot outside the cache store: %w", boundaryErr)
				}
				key := sourceSnapshotKey(source, entry.Name())
				if sourceRefs[key] || skillRefs[source+"/"+entry.Name()] {
					continue
				}
				candidates = append(candidates, snapshotCandidate{label: source + "/" + entry.Name(), path: snapshotPath})
				continue
			}
			if err := walk(child, childSource, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, "", 0); err != nil {
		return nil, nil, err
	}
	for _, candidate := range candidates {
		if dryRun {
			wouldRemove = append(wouldRemove, candidate.label)
			continue
		}
		if err := lock.AssertHeld(); err != nil {
			return removed, wouldRemove, err
		}
		if err := pathboundary.ValidateWithin(root, candidate.path); err != nil {
			return removed, wouldRemove, fmt.Errorf("refuse source snapshot outside the cache store: %w", err)
		}
		if strings.HasPrefix(filepath.Base(candidate.path), retiredSourceSnapshotPrefix) {
			if remove == nil {
				return removed, wouldRemove, fmt.Errorf("remove retired source snapshot: removal operation is unavailable")
			}
			if err := remove(candidate.path); err != nil {
				return removed, wouldRemove, fmt.Errorf("remove retired source snapshot %s: %w", filepath.Base(candidate.path), err)
			}
			removed = append(removed, candidate.label)
			continue
		}
		if err := retireSourceSnapshot(root, candidate.path, remove); err != nil {
			return removed, wouldRemove, err
		}
		removed = append(removed, candidate.label)
	}
	sort.Strings(removed)
	sort.Strings(wouldRemove)
	return removed, wouldRemove, nil
}

// retireSourceSnapshot removes the canonical cache name before deleting its
// tree. A failed recursive removal therefore leaves a hidden tombstone for a
// later sweep, never a partial tree that a cache lookup could mistake for the
// immutable snapshot.
func retireSourceSnapshot(root, path string, remove func(string) error) error {
	if remove == nil {
		return fmt.Errorf("remove source snapshot: removal operation is unavailable")
	}
	retired, err := reserveRetiredSourceSnapshot(filepath.Dir(path))
	if err != nil {
		return err
	}
	if err := pathboundary.ValidateWithin(root, filepath.Dir(retired)); err != nil {
		return fmt.Errorf("refuse retired source snapshot outside the cache store: %w", err)
	}
	if err := os.Rename(path, retired); err != nil {
		return fmt.Errorf("retire source snapshot %s: %w", path, err)
	}
	if err := remove(retired); err != nil {
		return fmt.Errorf("remove retired source snapshot %s: %w", filepath.Base(retired), err)
	}
	return nil
}

func reserveRetiredSourceSnapshot(parent string) (string, error) {
	for attempt := 0; attempt < 16; attempt++ {
		var suffix [8]byte
		if _, err := rand.Read(suffix[:]); err != nil {
			return "", fmt.Errorf("reserve a retired source snapshot name: %w", err)
		}
		path := filepath.Join(parent, retiredSourceSnapshotPrefix+hex.EncodeToString(suffix[:]))
		state, err := stateread.Lstat(path)
		if err != nil {
			return "", fmt.Errorf("inspect retired source snapshot name: %w", err)
		}
		if state.Kind == stateread.KindAbsent {
			return path, nil
		}
	}
	return "", fmt.Errorf("could not reserve a retired source snapshot name")
}

func previewRuntimeChecked(home string, referenced map[string]bool, check func() error) ([]string, error) {
	if err := check(); err != nil {
		return nil, err
	}
	root := filepath.Join(home, "runtime")
	listing, err := stateread.ReadDir(root)
	if err != nil {
		return nil, err
	}
	if listing.Kind == stateread.KindAbsent {
		return nil, nil
	}
	var candidates []string
	for _, skill := range listing.Entries {
		if !skill.IsDir() {
			continue
		}
		commitListing, err := stateread.ReadDir(filepath.Join(root, skill.Name()))
		if err != nil {
			return candidates, err
		}
		for _, commit := range commitListing.Entries {
			key := skill.Name() + "/" + commit.Name()
			if commit.IsDir() && !referenced[key] {
				if err := check(); err != nil {
					return candidates, err
				}
				candidates = append(candidates, key)
			}
		}
	}
	sort.Strings(candidates)
	return candidates, nil
}

// scopeMarks is one skills directory: its valid markers, everything it holds
// that exists but could not be validated, and everything it ignored.
type scopeMarks struct {
	markers   []*marker.Marker
	uncertain []string
	notes     []string
}

// readScope reads one skills directory without following a link into it.
//
// Only a plain directory is traversed. A skills root or member that is a
// symbolic link or a reparse point could redirect the traversal away from the
// installation whose marker holds the live build reference, so it is refused
// and reported instead of followed. Every metadata failure that is not a plain
// absence is likewise an uncertainty: a marker that cannot be inspected may
// name a build key that must not be swept.
func readScope(skillsDir string) scopeMarks {
	var scope scopeMarks
	metadata, err := stateread.Lstat(skillsDir)
	switch {
	case err == nil && metadata.Kind == stateread.KindAbsent:
		return scope
	case err != nil:
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s cannot be inspected: %v", skillsDir, err))
		return scope
	case isRedirect(metadata.Info):
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s is a symbolic link or reparse point; maintenance cannot prove which installations it holds",
			skillsDir))
		return scope
	case !metadata.Info.Mode().IsDir():
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s is not a directory; repair or remove it", skillsDir))
		return scope
	}

	listing, err := stateread.ReadDir(skillsDir)
	if err != nil {
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s is unreadable: %v", skillsDir, err))
		return scope
	}
	if listing.Kind == stateread.KindAbsent {
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s became absent during inspection: %s", skillsDir, stateread.DiagAbsent))
		return scope
	}
	for _, entry := range listing.Entries {
		installedDir := filepath.Join(skillsDir, entry.Name())
		member, err := stateread.Lstat(installedDir)
		switch {
		case err == nil && member.Kind == stateread.KindAbsent:
			continue // removed while the directory was being read
		case err != nil:
			scope.uncertain = append(scope.uncertain, fmt.Sprintf(
				"installed skill %s cannot be inspected: %v", installedDir, err))
			continue
		case isRedirect(member.Info):
			scope.uncertain = append(scope.uncertain, fmt.Sprintf(
				"installed skill %s is a symbolic link or reparse point; maintenance cannot prove what it installs",
				installedDir))
			continue
		case !member.Info.Mode().IsDir():
			// A non-directory member cannot hold an install marker, so it hides
			// no reference; it is reported so an operator can clean it up.
			scope.notes = append(scope.notes, fmt.Sprintf(
				"ignored non-directory member %s of skill directory %s", entry.Name(), skillsDir))
			continue
		}
		installed, kind, markerErr := marker.ReadState(installedDir)
		if markerErr != nil {
			readDescription := "is unreadable or invalid"
			diagnostic := stateread.DiagUnreadable
			var invalid *marker.InvalidError
			if errors.As(markerErr, &invalid) {
				readDescription = "is invalid"
				diagnostic = marker.DiagInvalid
			}
			var pathErr *os.PathError
			if diagnostic == stateread.DiagUnreadable && errors.As(markerErr, &pathErr) {
				readDescription = "cannot be inspected"
			}
			scope.uncertain = append(scope.uncertain, fmt.Sprintf(
				"install marker %s %s (%s): %v; repair or remove that installation",
				filepath.Join(installedDir, marker.Name), readDescription, diagnostic, markerErr))
			continue
		}
		if kind == stateread.KindAbsent {
			continue
		}
		scope.markers = append(scope.markers, installed)
	}
	return scope
}

// maintenanceLock retains the acquisition that produced the mark set; each
// downstream AssertHeld must prove that same acquisition is still current.
type maintenanceLock struct{ check func() error }

func (lock maintenanceLock) AssertHeld() error { return lock.check() }

func sweepRuntime(home string, referenced map[string]bool) ([]string, error) {
	return sweepRuntimeChecked(home, referenced, func() error { return nil })
}

func sweepRuntimeChecked(home string, referenced map[string]bool, check func() error) ([]string, error) {
	if err := check(); err != nil {
		return nil, err
	}
	runtimeRoot := filepath.Join(home, "runtime")
	listing, err := stateread.ReadDir(runtimeRoot)
	if err != nil {
		return nil, err
	}
	if listing.Kind == stateread.KindAbsent {
		return nil, nil
	}
	var removed []string
	for _, skill := range listing.Entries {
		if !skill.IsDir() {
			continue
		}
		skillPath := filepath.Join(runtimeRoot, skill.Name())
		commitListing, err := stateread.ReadDir(skillPath)
		if err != nil {
			return removed, err
		}
		if commitListing.Kind == stateread.KindAbsent {
			continue
		}
		for _, commit := range commitListing.Entries {
			if !commit.IsDir() {
				continue
			}
			key := skill.Name() + "/" + commit.Name()
			if !referenced[key] {
				if err := check(); err != nil {
					return removed, err
				}
				if err := os.RemoveAll(filepath.Join(runtimeRoot, skill.Name(), commit.Name())); err != nil {
					return removed, err
				}
				removed = append(removed, key)
			}
		}
		remaining, err := stateread.ReadDir(skillPath)
		if err != nil {
			return removed, err
		}
		if remaining.Kind == stateread.KindAbsent || len(remaining.Entries) == 0 {
			if err := check(); err != nil {
				return removed, err
			}
			_ = os.Remove(skillPath)
		}
	}
	return removed, nil
}
