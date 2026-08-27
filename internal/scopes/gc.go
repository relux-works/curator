package scopes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/marker"
)

// CollectRuntime removes runtime store entries referenced by no marker of
// any registered consumer, the global store, or the hybrid store, and prunes
// consumer entries whose checkout disappeared or holds no markers
// (Spec §8.7). Returns the removed runtime entries as skill/commit pairs.
//
// It collects the runtime store only. Collect is the locked maintenance entry
// point that also sweeps the protected build cache.
func CollectRuntime(home string) ([]string, error) {
	marked := markScopes(home)
	if err := pruneConsumers(home, marked); err != nil {
		return nil, err
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
}

// MaintenanceResult reports what one pass removed and what it could not prove.
type MaintenanceResult struct {
	RemovedRuntime []string
	RemovedBuilds  []string
	Warnings       []string
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
// reference set unprovable, so the build cache is not swept at all. Crucially,
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
	marked := markScopes(request.Home)
	result := MaintenanceResult{Warnings: marked.warnings()}

	// Consumer pruning stays inside this pass, under the caller's lock.
	if err := pruneConsumers(request.Home, marked); err != nil {
		return result, err
	}
	removedRuntime, err := sweepRuntime(request.Home, marked.runtime)
	result.RemovedRuntime = removedRuntime
	if err != nil {
		return result, err
	}

	if len(marked.uncertain) > 0 {
		result.Warnings = append(result.Warnings,
			"build cache sweep skipped: the live reference set could not be proven complete")
		return result, nil
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
	sweep, err := cache.Sweep(buildcache.SweepRequest{
		Referenced: referenced,
		Now:        request.Now,
		Grace:      request.Grace,
	}, request.Lock)
	result.RemovedBuilds = sweep.Removed
	result.Warnings = append(result.Warnings, sweep.Warnings...)
	if err != nil {
		return result, fmt.Errorf("sweep the protected build cache: %w", err)
	}
	return result, nil
}

// pruneConsumers rewrites the registry with the consumers one pass proved are
// still worth keeping. A registry that could not be trusted is left exactly as
// it is: overwriting it would destroy the only record of which checkouts may
// still reference a build artifact.
func pruneConsumers(home string, marked marks) error {
	if marked.registryUnknown {
		return nil
	}
	return ReplaceConsumers(home, marked.consumers)
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
	consumers []string
	runtime   map[string]bool
	builds    []string
	// uncertain describes state that exists but could not be trusted, so the
	// reference set derived from it is incomplete. Any entry here blocks the
	// build sweep and protects every consumer it came from.
	uncertain []string
	// notes describe state that was ignored but cannot hide a reference, so
	// they are reported without blocking maintenance.
	notes []string
	// registryUnknown records that the consumer registry exists but could not
	// be trusted, so it must not be rewritten.
	registryUnknown bool
}

func (marked marks) warnings() []string {
	warnings := append([]string(nil), marked.uncertain...)
	return append(warnings, marked.notes...)
}

func markScopes(home string) marks {
	marked := marks{runtime: map[string]bool{}}
	consumers, err := readConsumers(home)
	if err != nil {
		marked.registryUnknown = true
		marked.uncertain = append(marked.uncertain, fmt.Sprintf(
			"consumer registry %s cannot be trusted: %v; no consumer was pruned and no build entry was swept",
			filepath.Join(home, ConsumersName), err))
	}
	for _, projectRoot := range consumers {
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
	return marked
}

func (marked *marks) absorb(scope scopeMarks) {
	for _, installed := range scope.markers {
		marked.runtime[installed.Name+"/"+installed.Commit] = true
		if installed.SchemaVersion != marker.SchemaVersion {
			continue
		}
		for _, build := range installed.Builds {
			marked.builds = append(marked.builds, string(build.CacheKey))
		}
	}
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
	info, err := os.Lstat(skillsDir)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return scope
	case err != nil:
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s cannot be inspected: %v", skillsDir, err))
		return scope
	case isRedirect(info):
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s is a symbolic link or reparse point; maintenance cannot prove which installations it holds",
			skillsDir))
		return scope
	case !info.Mode().IsDir():
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s is not a directory; repair or remove it", skillsDir))
		return scope
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		scope.uncertain = append(scope.uncertain, fmt.Sprintf(
			"skill directory %s is unreadable: %v", skillsDir, err))
		return scope
	}
	for _, entry := range entries {
		installedDir := filepath.Join(skillsDir, entry.Name())
		member, err := entry.Info()
		switch {
		case errors.Is(err, os.ErrNotExist):
			continue // removed while the directory was being read
		case err != nil:
			scope.uncertain = append(scope.uncertain, fmt.Sprintf(
				"installed skill %s cannot be inspected: %v", installedDir, err))
			continue
		case isRedirect(member):
			scope.uncertain = append(scope.uncertain, fmt.Sprintf(
				"installed skill %s is a symbolic link or reparse point; maintenance cannot prove what it installs",
				installedDir))
			continue
		case !member.Mode().IsDir():
			// A non-directory member cannot hold an install marker, so it hides
			// no reference; it is reported so an operator can clean it up.
			scope.notes = append(scope.notes, fmt.Sprintf(
				"ignored non-directory member %s of skill directory %s", entry.Name(), skillsDir))
			continue
		}
		if installed := marker.Read(installedDir); installed != nil {
			scope.markers = append(scope.markers, installed)
			continue
		}
		scope.uncertain = append(scope.uncertain, unreadableMarker(installedDir)...)
	}
	return scope
}

// unreadableMarker classifies why an installed directory produced no valid
// marker. Only a plainly absent marker means "not an installation".
func unreadableMarker(installedDir string) []string {
	markerPath := filepath.Join(installedDir, marker.Name)
	info, err := os.Lstat(markerPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil
	case err != nil:
		return []string{fmt.Sprintf(
			"install marker %s cannot be inspected: %v; repair or remove that installation", markerPath, err)}
	case isRedirect(info) || !info.Mode().IsRegular():
		return []string{fmt.Sprintf(
			"install marker %s is not a regular file; repair or remove that installation", markerPath)}
	default:
		return []string{fmt.Sprintf(
			"install marker %s is unreadable or invalid; repair or remove that installation", markerPath)}
	}
}

func sweepRuntime(home string, referenced map[string]bool) ([]string, error) {
	runtimeRoot := filepath.Join(home, "runtime")
	skills, err := os.ReadDir(runtimeRoot)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var removed []string
	for _, skill := range skills {
		if !skill.IsDir() {
			continue
		}
		commits, err := os.ReadDir(filepath.Join(runtimeRoot, skill.Name()))
		if err != nil {
			continue
		}
		for _, commit := range commits {
			if !commit.IsDir() {
				continue
			}
			key := skill.Name() + "/" + commit.Name()
			if !referenced[key] {
				if err := os.RemoveAll(filepath.Join(runtimeRoot, skill.Name(), commit.Name())); err != nil {
					return removed, err
				}
				removed = append(removed, key)
			}
		}
		remaining, _ := os.ReadDir(filepath.Join(runtimeRoot, skill.Name()))
		if len(remaining) == 0 {
			_ = os.Remove(filepath.Join(runtimeRoot, skill.Name()))
		}
	}
	return removed, nil
}
