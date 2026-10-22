package projectlease

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildrepo"
)

// verifyArtifacts rechecks entry presence, not cache receipt/hash validity:
// Authority owns the latter and c is its validated immutable projection. The
// home lock stays held from this check through publication, excluding GC.
func (m *Manager) verifyArtifacts(c Composition) error {
	for _, entry := range c.RuntimeEntries {
		if err := m.artifactDirectory(filepath.Join(m.home, "runtime", filepath.FromSlash(entry))); err != nil {
			return fmt.Errorf("%w: runtime %s: %w", ErrArtifactsUnavailable, entry, err)
		}
	}
	for _, key := range c.BuildKeys {
		present := false
		for _, namespace := range []string{
			filepath.Join("cache", "build", buildcache.Namespace),
			filepath.Join("cache", "build", buildcache.SourceAwareNamespace),
			filepath.Join("external-build-cache", buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion)),
			filepath.Join("external-build-cache", buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion)),
		} {
			err := m.artifactDirectory(filepath.Join(m.home, namespace, strings.TrimPrefix(key, "sha256:")))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return fmt.Errorf("%w: build %s: %w", ErrArtifactsUnavailable, key, err)
			}
			present = true
			break
		}
		if !present {
			return fmt.Errorf("%w: build %s: %w", ErrArtifactsUnavailable, key, os.ErrNotExist)
		}
	}
	return nil
}

func (m *Manager) artifactDirectory(path string) error {
	if err := checkedRoute(m.home, path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("artifact entry is not a directory")
	}
	return nil
}
