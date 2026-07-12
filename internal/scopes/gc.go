package scopes

import (
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/marker"
)

// CollectRuntime removes runtime store entries referenced by no marker of
// any registered consumer, the global store, or the hybrid store, and prunes
// consumer entries whose checkout disappeared or holds no markers
// (Spec §8.7). Returns the removed runtime entries as skill/commit pairs.
func CollectRuntime(home string) ([]string, error) {
	referenced := map[string]bool{}
	var liveConsumers []string
	for _, projectRoot := range LoadConsumers(home) {
		markers := markersUnder(filepath.Join(projectRoot, ".agents", "skills"))
		if len(markers) == 0 {
			continue // dead consumer: gone or empty
		}
		liveConsumers = append(liveConsumers, projectRoot)
		for _, m := range markers {
			referenced[m.Name+"/"+m.Commit] = true
		}
	}
	for _, scopeDir := range []string{
		filepath.Join(home, "global", "skills"),
		filepath.Join(home, "hybrid", "skills"),
	} {
		for _, m := range markersUnder(scopeDir) {
			referenced[m.Name+"/"+m.Commit] = true
		}
	}
	if err := ReplaceConsumers(home, liveConsumers); err != nil {
		return nil, err
	}

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

func markersUnder(skillsDir string) []*marker.Marker {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}
	var markers []*marker.Marker
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if m := marker.Read(filepath.Join(skillsDir, entry.Name())); m != nil {
			markers = append(markers, m)
		}
	}
	return markers
}
