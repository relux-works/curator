// Package snapshot maintains the commit-keyed immutable snapshot cache under
// the machine home: cache/<source>/<commit>/snapshot (Spec §8.2).
package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/gitops"
)

// Dir returns the cache location for a source at a commit.
func Dir(home, source, commit string) string {
	return filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot")
}

// Get returns the snapshot directory, producing it from the repository on a
// cache miss. Staging is atomic: a concurrent producer of the same snapshot
// wins harmlessly.
func Get(home, source, repo, commit string) (string, error) {
	target := Dir(home, source, commit)
	if _, err := os.Stat(target); err == nil {
		return target, nil
	}
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp := filepath.Join(parent, fmt.Sprintf(".snapshot-%d.tmp", os.Getpid()))
	if err := os.RemoveAll(tmp); err != nil {
		return "", err
	}
	if err := gitops.Archive(repo, commit, tmp); err != nil {
		_ = os.RemoveAll(tmp)
		return "", err
	}
	if _, err := os.Stat(target); err == nil {
		_ = os.RemoveAll(tmp)
		return target, nil
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.RemoveAll(tmp)
		return "", err
	}
	return target, nil
}
