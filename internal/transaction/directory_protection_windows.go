//go:build windows

package transaction

import (
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/privatedir"
)

// A new private metadata-only template supplies the actual Windows descriptor
// rather than guessing a chmod/DACL equivalence. It is allocated before the
// journal like other private preparation artifacts; it never touches a live
// parent, native credential, session or unrelated descendant.
func prepareExistingDirectoryProtection(dir DirectoryRecord, index int, id string) (envunit.Image, error) {
	if dir.Before.Mode == dir.After.Mode {
		return dir.After, nil
	}
	path := directorySidecar(dir.BasePath, id, index) + ".protection"
	if err := privatedir.Make(path); err != nil {
		return envunit.Image{}, err
	}
	template, err := envunit.ObserveDirectory(path)
	if err != nil {
		return envunit.Image{}, err
	}
	if template.Owner != dir.Before.Owner || template.Mode != 0700 {
		return envunit.Image{}, fmt.Errorf("private directory template owner/protection differs")
	}
	current, err := envunit.ObserveDirectory(path)
	if err != nil || current != template {
		return envunit.Image{}, fmt.Errorf("private directory template changed: %v", err)
	}
	if err = os.Remove(path); err != nil {
		return envunit.Image{}, err
	}
	after := dir.After
	after.Protection = template.Protection
	return after, nil
}
