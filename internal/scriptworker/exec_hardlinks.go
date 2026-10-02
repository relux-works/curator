package scriptworker

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/godriver"
)

// The service SID is fixed by Windows, independent of localized account names.
const trustedInstallerSID = "S-1-5-80-956008885-3418522649-1831038044-1853292631-2271478464"
const localSystemSID = "S-1-5-18"

type windowsExecHardlinkOrigin struct {
	OwnerSID string
	Links    []string // complete, physical paths to the same file ID as the open executable
	Count    uint32
}

// Tests substitute OS observations, while driving the production resolver and
// VerifyExec. Neither callers nor package metadata can supply these observations.
var inspectWindowsExecHardlinks = nativeWindowsExecHardlinks

func trustedWindowsExecHardlinks(file *os.File, canonical, system32 string) bool {
	origin, err := inspectWindowsExecHardlinks(file, canonical)
	if err != nil || origin.Count < 2 || uint64(len(origin.Links)) != uint64(origin.Count) {
		return false
	}
	if origin.OwnerSID != trustedInstallerSID && origin.OwnerSID != localSystemSID {
		return false
	}
	store, err := godriver.CanonicalPhysicalPath(filepath.Join(filepath.Dir(system32), "WinSxS"))
	// A redirected component store cannot authorize links outside captured root.
	if err != nil || !strings.EqualFold(store, filepath.Join(filepath.Dir(system32), "WinSxS")) {
		return false
	}
	seen := make(map[string]bool, len(origin.Links))
	foundTarget := false
	for _, link := range origin.Links {
		key := strings.ToLower(filepath.Clean(link))
		if !filepath.IsAbs(link) || seen[key] {
			return false
		}
		seen[key] = true
		if strings.EqualFold(link, canonical) {
			foundTarget = true
			continue
		}
		if !pathWithinPlatform(link, store, "windows") {
			return false
		}
	}
	return foundTarget
}
