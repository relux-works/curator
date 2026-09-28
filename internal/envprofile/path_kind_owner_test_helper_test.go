package envprofile

import (
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/pathboundary"
)

// forceWrongOwnerForTest injects a foreign owner identity for one path while
// leaving the production owner lookup in place for every other component.
// This drives the boundary through Install/Resolve without chown privileges.
func forceWrongOwnerForTest(path string) (func(), string) {
	previous := pathSourceOwnerLookup
	lookup := pathSourceOwnerLookup
	target := filepath.Clean(path)
	pathSourceOwnerLookup = func(candidate string, info os.FileInfo) (pathboundary.OwnerIdentity, error) {
		if filepath.Clean(candidate) == target {
			return pathboundary.OwnerIdentity("injected:foreign-owner"), nil
		}
		return lookup(candidate, info)
	}
	return func() { pathSourceOwnerLookup = previous }, ""
}
