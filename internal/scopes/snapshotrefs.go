package scopes

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/stateread"
)

// SnapshotReferences is the reference half of manager profile section 10.1:
// every commit a live record names, and every record source that exists but
// could not be trusted.
//
// Commits are matched without their source key. That over-retains an entry
// whose commit another source happens to share, which section 10.1 permits
// because retaining an entry is always safe.
type SnapshotReferences struct {
	// Commits holds every full commit named by a valid install marker, source
	// lock, or installed profile lock.
	Commits map[string]bool
	// Uncertain describes each record source that exists but could not be
	// read, parsed, or proven link-free. Any entry makes the reference set
	// uncertain, and then retention removes nothing.
	Uncertain []string
	// Notes describe ignored state that cannot hide a reference.
	Notes []string
}

// Certain reports whether the reference set is provably complete.
func (refs SnapshotReferences) Certain() bool { return len(refs.Uncertain) == 0 }

// CollectSnapshotReferences reads, without mutating anything, every record
// that can keep a commit-keyed snapshot reachable:
//
//   - install markers in registered consumers and the global and hybrid scopes
//     (the same traversal garbage collection marks from);
//   - the Skillfile.lock.json of every registered consumer and of the global
//     and hybrid scopes; and
//   - the lock of every installed profile.
//
// In-flight transaction journals are not read here: the caller recovers them
// under the manager-home lock first, exactly as `curator gc` does, so none
// remains in flight when the plan is computed.
func CollectSnapshotReferences(home string) SnapshotReferences {
	refs := SnapshotReferences{Commits: map[string]bool{}}
	marked := markScopes(home)
	refs.Uncertain = append(refs.Uncertain, marked.uncertain...)
	refs.Notes = append(refs.Notes, marked.notes...)
	for key := range marked.runtime {
		// Runtime keys are <skill>/<commit> for Git installs and
		// <skill>/<source-v1 key> for drafts. Only the commit form can name
		// a snapshot; a draft's locked commit is read from its source lock.
		refs.Commits[path.Base(key)] = true
	}
	for _, commit := range marked.commits {
		refs.Commits[commit] = true
	}

	roots := []string{filepath.Join(home, "global"), filepath.Join(home, "hybrid")}
	consumers, err := readConsumers(home)
	if err != nil {
		refs.Uncertain = append(refs.Uncertain, fmt.Sprintf(
			"consumer registry %s cannot be trusted: %v", filepath.Join(home, ConsumersName), err))
	}
	roots = append(roots, consumers...)
	for _, root := range roots {
		refs.readSourceLock(sourcelock.PathIn(root))
	}
	refs.readProfileLocks(envprofile.ProfilesDir(home))
	sort.Strings(refs.Uncertain)
	return refs
}

func (refs *SnapshotReferences) readSourceLock(path string) {
	state, err := stateread.ReadRegularFile(path)
	if err != nil {
		refs.Uncertain = append(refs.Uncertain, fmt.Sprintf("source lock %s cannot be trusted: %v", path, err))
		return
	}
	if state.Kind == stateread.KindAbsent {
		return
	}
	lock, err := sourcelock.Parse(state.Bytes)
	if err != nil {
		refs.Uncertain = append(refs.Uncertain, fmt.Sprintf("source lock %s is invalid: %v", path, err))
		return
	}
	for _, member := range lock.Members {
		if member.Package.Commit.Hex != "" {
			refs.Commits[member.Package.Commit.Hex] = true
		}
	}
}

func (refs *SnapshotReferences) readProfileLocks(profiles string) {
	metadata, err := stateread.Lstat(profiles)
	switch {
	case err != nil:
		refs.Uncertain = append(refs.Uncertain, fmt.Sprintf("profile directory %s cannot be inspected: %v", profiles, err))
		return
	case metadata.Kind == stateread.KindAbsent:
		return
	case isRedirect(metadata.Info) || !metadata.Info.IsDir():
		refs.Uncertain = append(refs.Uncertain, fmt.Sprintf(
			"profile directory %s is not a plain directory; its locks cannot be proven", profiles))
		return
	}
	listing, err := stateread.ReadDir(profiles)
	if err != nil || listing.Kind != stateread.KindPresent {
		refs.Uncertain = append(refs.Uncertain, fmt.Sprintf("profile directory %s is unreadable: %v", profiles, err))
		return
	}
	for _, entry := range listing.Entries {
		dir := filepath.Join(profiles, entry.Name())
		member, err := stateread.Lstat(dir)
		switch {
		case err != nil:
			refs.Uncertain = append(refs.Uncertain, fmt.Sprintf("profile %s cannot be inspected: %v", dir, err))
			continue
		case member.Kind == stateread.KindAbsent:
			continue
		case isRedirect(member.Info):
			refs.Uncertain = append(refs.Uncertain, fmt.Sprintf(
				"profile %s is a symbolic link or reparse point; its lock cannot be proven", dir))
			continue
		case !member.Info.IsDir():
			continue // current-profile pointers and other records hold no lock
		}
		lockPath := filepath.Join(dir, "lock.json")
		state, err := stateread.ReadRegularFile(lockPath)
		if err != nil {
			refs.Uncertain = append(refs.Uncertain, fmt.Sprintf("profile lock %s cannot be trusted: %v", lockPath, err))
			continue
		}
		if state.Kind == stateread.KindAbsent {
			continue
		}
		lock, err := contextlock.Parse(state.Bytes)
		if err != nil {
			refs.Uncertain = append(refs.Uncertain, fmt.Sprintf("profile lock %s is invalid: %v", lockPath, err))
			continue
		}
		for _, locked := range lock.Members {
			if locked.Commit != "" {
				refs.Commits[locked.Commit] = true
			}
		}
	}
}
