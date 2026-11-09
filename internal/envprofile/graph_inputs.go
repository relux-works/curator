package envprofile

import (
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/stateread"
)

// GraphInputs reuses resolve's source/lock admission without provisioning,
// credential checks, runtime probes or a current-profile lookup. The caller
// supplies bounded protected reads of both records.
func GraphInputs(home, profile string, read func(string) (stateread.File, error)) (Source, *contextlock.Lock, string, error) {
	// Admit both records before legacy owner/boundary metadata checks, so an
	// escaping link retains the graph reader's named refusal. Protected readers
	// cache these bytes; the shared resolver still parses and validates them.
	for _, path := range []string{sourcePath(home, profile), lockPath(home, profile)} {
		if _, err := read(path); err != nil {
			return Source{}, nil, "", err
		}
	}
	return loadResolveInputsWith(home, profile, read)
}

// GraphMemberEntry preserves the resolver's Git versus state cache layout.
func GraphMemberEntry(home string, member contextlock.Member) string {
	return memberEntryPath(home, member)
}

// GraphPackage uses install's requirement projection for the offline resolver.
func GraphPackage(manifest *contextpkg.Manifest) *contextresolve.Package { return packageOf(manifest) }

// GraphRepository names the existing local object cache, without acquisition.
func GraphRepository(home, source string) string { return newGitManager(home).repoDir(source) }
