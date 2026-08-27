package adapters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/staging"
)

// Mirror is the staged adapter state of one installation.
//
// Every part of it is a journaled transaction target: each mirror entry —
// a copied tree or a symbolic link — the ownership ledger of each adapter root,
// and each stale managed entry the next ledger no longer claims. Nothing here
// touches a live path, so the whole mirror commits with the rest of the
// installation and a failure at any class restores its exact prior state,
// including the destination of a link that was replaced.
//
// Mirror entries commit in the adapter class before the ledger that claims
// them, and stale removals commit in the removal class; both are ordered before
// the machine-wide consumer ledger.
type Mirror struct {
	plan staging.Plan
	// Messages carries operator-facing notes about entries left alone.
	Messages []string
}

// Plan returns the journaled adapter targets.
func (mirror Mirror) Plan() staging.Plan { return mirror.plan }

// StageProject stages the project-level adapter mirrors of the selected agents.
// Native-discovery agents get no project mirror.
func StageProject(stageRoot, projectRoot string, agents []string, groups []Group, mode string) (Mirror, error) {
	roots := map[string]string{}
	for _, agent := range agents {
		if rel, known := AgentPaths[agent]; known {
			roots[agent] = filepath.Join(projectRoot, filepath.FromSlash(rel))
		}
	}
	return stage(stageRoot, roots, groups, mode)
}

// StageGlobal stages the home-level adapter mirrors, including the
// native-discovery mirror.
func StageGlobal(
	stageRoot, home, userHome string,
	agents []string,
	skills []string,
	mode string,
	sources map[string]string,
) (Mirror, error) {
	canonical := filepath.Join(home, "global", "skills")
	roots := map[string]string{}
	for _, agent := range agents {
		if rel, known := AgentPaths[agent]; known {
			roots[agent] = filepath.Join(userHome, filepath.FromSlash(rel))
		}
		if NativeDiscoveryAgents[agent] {
			roots[agent] = filepath.Join(userHome, filepath.FromSlash(NativeDiscoveryHomePath))
		}
	}
	return stage(stageRoot, roots, []Group{{Root: canonical, Skills: skills, Sources: sources}}, mode)
}

func stage(stageRoot string, adapterRoots map[string]string, groups []Group, mode string) (Mirror, error) {
	expected := map[string]bool{}
	for _, group := range groups {
		for _, name := range group.Skills {
			expected[name] = true
		}
	}
	agents := make([]string, 0, len(adapterRoots))
	for agent := range adapterRoots {
		agents = append(agents, agent)
	}
	sort.Strings(agents)

	var mirror Mirror
	staged := map[string]bool{}
	for _, agent := range agents {
		adapterRoot := adapterRoots[agent]
		if staged[adapterRoot] {
			// Two agents may share one native-discovery root; stage it once.
			continue
		}
		staged[adapterRoot] = true
		if err := mirror.stageRoot(stageRoot, adapterRoot, groups, expected, mode); err != nil {
			return Mirror{}, err
		}
	}
	return mirror, nil
}

func (mirror *Mirror) stageRoot(stageRoot, adapterRoot string, groups []Group, expected map[string]bool, mode string) error {
	key := rootKey(adapterRoot)
	managed := readLedger(adapterRoot)

	var stale []string
	for name := range managed {
		if !expected[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	for _, name := range stale {
		if err := mirror.stageStale(key, name, filepath.Join(adapterRoot, name)); err != nil {
			return err
		}
	}

	for _, group := range groups {
		names := append([]string(nil), group.Skills...)
		sort.Strings(names)
		for _, name := range names {
			// The mirror points at the canonical root the commit publishes,
			// while its bytes are read from wherever the content lives right
			// now — the staged replacement for a skill this run installs, the
			// canonical directory for one that is already current.
			content := group.source(name)
			if _, err := os.Stat(content); err != nil {
				continue
			}
			canonical := filepath.Join(group.Root, name)
			target := filepath.Join(adapterRoot, name)
			conflict, err := unmanagedConflict(target, managed[name], canonical)
			if err != nil {
				return err
			}
			if conflict {
				return fmt.Errorf("adapter target already exists and is not managed: %s", target)
			}
			if err := mirror.stageEntry(stageRoot, key, name, content, canonical, target, mode); err != nil {
				return err
			}
		}
	}

	ledger, err := ledgerPayload(expected)
	if err != nil {
		return err
	}
	stagedLedger := filepath.Join(stageRoot, "adapters", key, LedgerName)
	if err := os.MkdirAll(filepath.Dir(stagedLedger), 0o700); err != nil {
		return fmt.Errorf("stage adapter ledger: %w", err)
	}
	if err := os.WriteFile(stagedLedger, ledger, 0o644); err != nil {
		return fmt.Errorf("stage adapter ledger: %w", err)
	}
	// "ledger" sorts after "entry/..." bytewise, so every mirror this ledger
	// claims is already durable when the claim itself commits.
	mirror.plan.Replace(staging.ClassAdapterLedger, key+"/ledger", filepath.Join(adapterRoot, LedgerName), stagedLedger)
	return nil
}

// stageStale journals a managed entry the next ledger no longer claims. Its
// desired state is absence, so the transaction backs the entry up before
// removing it and restores it — a copied tree byte for byte, a link with its
// exact destination — if any later class fails.
//
// An entry that is neither a directory, a regular file, nor a link is reported
// and left alone rather than removed behind a rollback that could not put it
// back; the same rule the canonical stores use for stale skills.
func (mirror *Mirror) stageStale(key, name, target string) error {
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 && !info.IsDir() && !info.Mode().IsRegular() {
		mirror.Messages = append(mirror.Messages, fmt.Sprintf(
			"stale adapter entry %q was left in %s because it is a special file that cannot be removed transactionally",
			name, filepath.Dir(target)))
		return nil
	}
	mirror.plan.RemoveEntry("adapter/"+key+"/"+name, target)
	return nil
}

// stageEntry stages one mirror entry. "auto" prefers a link and falls back to a
// journaled copy when the staging root or the adapter root refuses one.
//
// A link always points at the canonical path, never at the staged content: the
// staged tree is operation-private and disappears when the run ends.
func (mirror *Mirror) stageEntry(stageRoot, key, name, content, canonical, target, mode string) error {
	switch mode {
	case "copy":
		return mirror.stageCopy(stageRoot, key, name, content, target)
	case "symlink":
		return mirror.stageLink(stageRoot, key, name, canonical, target)
	default:
		if linksSupported(stageRoot, target) {
			return mirror.stageLink(stageRoot, key, name, canonical, target)
		}
		return mirror.stageCopy(stageRoot, key, name, content, target)
	}
}

// stageLink stages one symbolic-link mirror entry. An entry that is already the
// exact link the commit would install produces no target at all: the desired
// state is the live state, and journaling an identical replacement would churn
// every adapter root on every run.
func (mirror *Mirror) stageLink(stageRoot, key, name, canonical, target string) error {
	destination := linkDestination(canonical, target)
	if current, err := os.Readlink(target); err == nil && current == destination {
		return nil
	}
	stagedEntry := filepath.Join(stageRoot, "adapters", key, "links", name)
	if err := os.MkdirAll(filepath.Dir(stagedEntry), 0o700); err != nil {
		return fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	if err := os.RemoveAll(stagedEntry); err != nil {
		return fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	if err := os.Symlink(destination, stagedEntry); err != nil {
		return fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	mirror.plan.ReplaceEntry(staging.ClassAdapterLedger, key+"/entry/"+name, target, stagedEntry)
	return nil
}

func (mirror *Mirror) stageCopy(stageRoot, key, name, content, target string) error {
	stagedEntry := filepath.Join(stageRoot, "adapters", key, "entries", name)
	if err := os.MkdirAll(filepath.Dir(stagedEntry), 0o700); err != nil {
		return fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	if err := os.RemoveAll(stagedEntry); err != nil {
		return fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	if err := copyTree(content, stagedEntry); err != nil {
		return fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	// The live entry may currently be a link from an earlier symlink-mode run,
	// so the copy replaces the directory entry itself, not what it points at.
	mirror.plan.ReplaceEntry(staging.ClassAdapterLedger, key+"/entry/"+name, target, stagedEntry)
	return nil
}

// linkDestination renders the exact destination string a mirror link carries:
// relative to the entry's own directory whenever that is expressible.
func linkDestination(canonical, target string) string {
	rel, err := filepath.Rel(filepath.Dir(target), canonical)
	if err != nil {
		return canonical
	}
	return rel
}

// linksSupported probes every filesystem an auto-mode link has to exist on: the
// operation-private staging root, and the adapter root that will hold both the
// transaction sidecar and the published entry. Discovering the limitation here
// keeps it out of the middle of a commit.
func linksSupported(stageRoot, target string) bool {
	if err := os.MkdirAll(stageRoot, 0o700); err != nil {
		return false
	}
	return linkProbe(stageRoot) && linkProbe(nearestExistingDirectory(filepath.Dir(target)))
}

// linkProbe creates and removes one private link. The name carries the process
// id so a probe is never mistaken for managed state and never collides with
// another manager home that happens to reach the same directory.
func linkProbe(directory string) bool {
	if directory == "" {
		return false
	}
	probe := filepath.Join(directory, fmt.Sprintf(".curator-symlink-probe-%d", os.Getpid()))
	_ = os.Remove(probe)
	if err := os.Symlink(".", probe); err != nil {
		return false
	}
	return os.Remove(probe) == nil
}

// nearestExistingDirectory walks up to the closest directory that exists, which
// is the filesystem the missing ones will be created on.
func nearestExistingDirectory(path string) string {
	current := filepath.Clean(path)
	for {
		if info, err := os.Lstat(current); err == nil {
			if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
				return current
			}
			return ""
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func rootKey(adapterRoot string) string {
	digest := sha256.Sum256([]byte(adapterRoot))
	return hex.EncodeToString(digest[:8])
}
