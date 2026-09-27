package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/stateread"
)

const systemDeltaHint = "revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given"

type deltaMemberChange struct {
	old *contextlock.Member
	new *contextlock.Member
}

type systemModuleRecord struct {
	path         string
	environments []string
	content      []byte
}

// resolvedDelta renders §9.2 lock-delta rows and identifies member names
// whose system-module inventory or MCP declaration changed.
func resolvedDelta(home string, manager *gitManager, oldLock, newLock *contextlock.Lock) ([]string, []string, error) {
	changes := make(map[string]deltaMemberChange)
	deltas := contextlock.ResolvedDelta(oldLock, newLock)
	lines := make([]string, 0, len(deltas))
	for _, delta := range deltas {
		lines = append(lines, delta.Line())
		member := delta.New
		if member == nil {
			member = delta.Old
		}
		key := contextresolve.Key(member.Kind, member.Name)
		changes[key] = deltaMemberChange{old: delta.Old, new: delta.New}
	}
	triggerNames := map[string]bool{}
	for key, change := range changes {
		var member contextlock.Member
		if change.new != nil {
			member = *change.new
		} else if change.old != nil {
			member = *change.old
		}
		switch member.Kind {
		case contextlock.KindContext:
			if change.new == nil {
				continue
			}
			newInventory, err := systemModuleInventory(home, manager, *change.new, true)
			if err != nil {
				return lines, nil, fmt.Errorf("%s: cannot inspect system modules for %s: %w", DiagUpdateBlocked, key, err)
			}
			if change.old == nil {
				if len(newInventory) > 0 {
					triggerNames[member.Name] = true
				}
				continue
			}
			oldInventory, err := systemModuleInventory(home, manager, *change.old, false)
			if err != nil {
				return lines, nil, fmt.Errorf("%s: cannot inspect prior system modules for %s: %w", DiagUpdateBlocked, key, err)
			}
			if !equalSystemInventory(oldInventory, newInventory) {
				triggerNames[member.Name] = true
			}
		case contextlock.KindMCP:
			if change.new == nil {
				continue
			}
			newDeclaration, err := mcpDeclaration(home, manager, *change.new, true)
			if err != nil {
				return lines, nil, fmt.Errorf("%s: cannot inspect MCP declaration for %s: %w", DiagUpdateBlocked, key, err)
			}
			if change.old == nil {
				triggerNames[member.Name] = true
				continue
			}
			oldDeclaration, err := mcpDeclaration(home, manager, *change.old, false)
			if err != nil {
				return lines, nil, fmt.Errorf("%s: cannot inspect prior MCP declaration for %s: %w", DiagUpdateBlocked, key, err)
			}
			if !bytes.Equal(oldDeclaration, newDeclaration) {
				triggerNames[member.Name] = true
			}
		}
	}
	triggers := make([]string, 0, len(triggerNames))
	for name := range triggerNames {
		triggers = append(triggers, name)
	}
	sort.Strings(triggers)
	return lines, triggers, nil
}

func systemModuleInventory(home string, manager *gitManager, member contextlock.Member, allowExtract bool) ([]systemModuleRecord, error) {
	root, cleanup, err := deltaMemberRoot(home, manager, member, allowExtract)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	manifest, err := contextpkg.LoadManifest(packageRoot(root, member.Directory))
	if err != nil {
		return nil, err
	}
	records := make([]systemModuleRecord, 0)
	for _, module := range manifest.Modules {
		if module.Class != "system" {
			continue
		}
		path := filepath.Join(packageRoot(root, member.Directory), contextpkg.ContextDir, filepath.FromSlash(module.Path))
		content, err := os.ReadFile(path) // #nosec G304 -- declared module under immutable package snapshot
		if err != nil {
			return nil, err
		}
		var environments []string
		if module.Environments != nil {
			environments = make([]string, len(module.Environments))
			copy(environments, module.Environments)
		}
		records = append(records, systemModuleRecord{path: module.Path, environments: environments, content: content})
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].path != records[j].path {
			return records[i].path < records[j].path
		}
		if environmentsKey(records[i].environments) != environmentsKey(records[j].environments) {
			return environmentsKey(records[i].environments) < environmentsKey(records[j].environments)
		}
		return bytes.Compare(records[i].content, records[j].content) < 0
	})
	return records, nil
}

func environmentsKey(environments []string) string {
	if environments == nil {
		return "null"
	}
	payload, _ := protocoljson.MarshalCanonical(environments)
	return string(payload)
}

func equalSystemInventory(left, right []systemModuleRecord) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].path != right[index].path || environmentsKey(left[index].environments) != environmentsKey(right[index].environments) || !bytes.Equal(left[index].content, right[index].content) {
			return false
		}
	}
	return true
}

func mcpDeclaration(home string, manager *gitManager, member contextlock.Member, allowExtract bool) ([]byte, error) {
	root, cleanup, err := deltaMemberRoot(home, manager, member, allowExtract)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	payload, err := os.ReadFile(filepath.Join(packageRoot(root, member.Directory), contextpkg.MCPManifestName)) // #nosec G304 -- member snapshot is immutable
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var manifest map[string]any
	if err := decoder.Decode(&manifest); err != nil {
		return nil, err
	}
	server, ok := manifest["server"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("server declaration is absent or malformed")
	}
	return protocoljson.MarshalCanonical(server)
}

func deltaMemberRoot(home string, manager *gitManager, member contextlock.Member, allowExtract bool) (string, func(), error) {
	entry := contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey())
	metadata, err := stateread.Stat(entry)
	if err != nil {
		return "", func() {}, err
	}
	switch metadata.Kind {
	case stateread.KindPresent:
		if metadata.Info == nil {
			return "", func() {}, stateread.UnusableError(entry, fmt.Errorf("present snapshot has no metadata"))
		}
		if !metadata.Info.IsDir() {
			return "", func() {}, stateread.UnusableError(entry, fmt.Errorf("locked package snapshot is not a directory"))
		}
		return entry, func() {}, nil
	case stateread.KindAbsent:
		// Only proven absence permits the immutable Git snapshot fallback.
	default:
		return "", func() {}, stateread.UnusableError(entry, fmt.Errorf("unknown snapshot state %q", metadata.Kind))
	}
	if !allowExtract || member.Commit == "" || member.Source == "" {
		return "", func() {}, fmt.Errorf("locked package snapshot is unavailable")
	}
	repo := manager.repoDir(member.Source)
	repoMetadata, err := stateread.Stat(repo)
	if err != nil {
		return "", func() {}, fmt.Errorf("cached source for %s is unavailable: %w", member.Name, err)
	}
	switch repoMetadata.Kind {
	case stateread.KindAbsent:
		return "", func() {}, fmt.Errorf("cached source for %s is unavailable: %w", member.Name, stateread.AbsentError(repo))
	case stateread.KindPresent:
		if repoMetadata.Info == nil {
			return "", func() {}, stateread.UnusableError(repo, fmt.Errorf("cached source has no metadata"))
		}
		if !repoMetadata.Info.IsDir() {
			return "", func() {}, stateread.UnusableError(repo, fmt.Errorf("cached source is not a directory"))
		}
	default:
		return "", func() {}, stateread.UnusableError(repo, fmt.Errorf("unknown cached source state %q", repoMetadata.Kind))
	}
	snapshot, err := os.MkdirTemp("", "curator-profile-delta-*")
	if err != nil {
		return "", func() {}, err
	}
	if err := gitops.Extract(repo, member.Commit, snapshot); err != nil {
		_ = os.RemoveAll(snapshot)
		return "", func() {}, err
	}
	return snapshot, func() { _ = os.RemoveAll(snapshot) }, nil
}

func writeUpdateWarnings(sink io.Writer, profile string, warnings []string) {
	if sink == nil {
		return
	}
	for _, warning := range warnings {
		if profile == "" {
			_, _ = fmt.Fprintln(sink, "warning:", warning)
		} else {
			_, _ = fmt.Fprintf(sink, "%s: warning: %s\n", profile, warning)
		}
	}
}
