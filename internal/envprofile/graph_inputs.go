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
