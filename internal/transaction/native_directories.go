package transaction

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
)

// NativePublicationDirectories requires an explicit enumerated native custody
// root. It creates only absent ancestors, never changes an existing native
// parent or snapshots its contents. Each action is rebound to its exact inode
// before the first journal, with the same durable proof as managed parents.
func (engine *Engine) NativePublicationDirectories(root string, paths ...string) ([]DirectoryAction, error) {
	if err := envorigin.ValidateExternal(engine.home, root); err != nil {
		return nil, err
	}
	byPath := map[string]DirectoryAction{}
	for _, path := range paths {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path || !insideDirectory(path, root) {
			return nil, fmt.Errorf("native target outside custody root")
		}
		parent := filepath.Dir(path)
		for {
			before, err := envunit.ObserveDirectory(parent)
			if err != nil {
				return nil, err
			}
			if before.Kind != "absent" {
				break
			}
			if parent == filepath.Dir(parent) {
				return nil, fmt.Errorf("no native publication ancestor")
			}
			byPath[parent] = DirectoryAction{LivePath: parent, Before: before, Mode: 0700}
			parent = filepath.Dir(parent)
		}
	}
	result := make([]DirectoryAction, 0, len(byPath))
	for _, a := range byPath {
		result = append(result, a)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].LivePath < result[j].LivePath })
	return result, nil
}
func (engine *Engine) verifyNativeDirectoryBase(j *Journal, dir DirectoryRecord) error {
	if j.Unit == nil {
		return corruptionf("external directory lacks typed custody")
	}
	admitted := false
	for _, n := range j.Unit.Native {
		if (insideDirectory(n.Root, dir.BasePath) || insideDirectory(dir.BasePath, n.Root)) && insideDirectory(dir.LivePath, dir.BasePath) && (insideDirectory(dir.LivePath, n.Root) || insideDirectory(n.Root, dir.LivePath)) {
			admitted = true
		}
	}
	if j.Unit.LegacyParent(dir.LivePath) {
		admitted = true
	}
	if !admitted {
		return corruptionf("external directory base is outside native custody")
	}
	if err := j.Unit.VerifyNativeRoots(); err != nil {
		return err
	}
	return verifyDirectoryImage(dir.BasePath, dir.Base)
}
