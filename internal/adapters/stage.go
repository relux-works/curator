package adapters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/linkcap"
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
// native-discovery mirror. admitted lists the inputs the planned mirrors
// must not overwrite; empty disables that check.
func StageGlobal(
	stageRoot, home, userHome string,
	agents []string,
	skills []string,
	mode string,
	sources map[string]string,
	admitted []string,
) (Mirror, error) {
	return StageGlobalWithOwnership(stageRoot, home, userHome, agents, skills, mode, sources, admitted, nil)
}

// GlobalOwnership verifies an entry owned by another Curator publisher. A true
// result preserves the entry and excludes it from this publisher's ledger.
// The caller must verify both ownership and the requested package identity.
type GlobalOwnership func(adapterRoot, name string) (bool, error)

// StageGlobalWithOwnership leaves verified profile entries with their existing
// owner; the legacy mirror still handles every other entry normally.
func StageGlobalWithOwnership(
	stageRoot, home, userHome string,
	agents, skills []string,
	mode string,
	sources map[string]string,
	admitted []string,
	owner GlobalOwnership,
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
	return stageWithOwnership(stageRoot, roots, []Group{{Root: canonical, Skills: skills, Sources: sources, Admitted: admitted}}, mode, owner)
}

func stage(stageRoot string, adapterRoots map[string]string, groups []Group, mode string) (Mirror, error) {
	return stageWithOwnership(stageRoot, adapterRoots, groups, mode, nil)
}

func stageWithOwnership(stageRoot string, adapterRoots map[string]string, groups []Group, mode string, owner GlobalOwnership) (Mirror, error) {
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
		if err := mirror.stageRoot(stageRoot, adapterRoot, groups, expected, mode, owner); err != nil {
			return Mirror{}, err
		}
	}
	// Production planning pins its admitted inputs on the groups; the
	// planned mirrors must not overwrite them. Empty Admitted skips the
	// gate.
	var admitted []string
	for _, group := range groups {
		admitted = append(admitted, group.Admitted...)
	}
	if len(admitted) > 0 {
		if err := ValidateDestinations(mirror.plan, admitted); err != nil {
			return Mirror{}, err
		}
	}
	return mirror, nil
}

func (mirror *Mirror) stageRoot(stageRoot, adapterRoot string, groups []Group, expected map[string]bool, mode string, owner GlobalOwnership) error {
	key := rootKey(adapterRoot)
	managed, err := readLedger(adapterRoot)
	if err != nil {
		return fmt.Errorf("read adapter ownership ledger in %s: %w", adapterRoot, err)
	}
	// Ownership is per adapter entry. Never add a profile-owned name to the
	// legacy ledger, or later profile switches would treat it as legacy state.
	preserved := map[string]bool{}
	claimed := map[string]bool{}
	for name := range expected {
		claimed[name] = true
	}
	for name := range managed {
		claimed[name] = true
	}
	localExpected := map[string]bool{}
	claimNames := make([]string, 0, len(claimed))
	for name := range claimed {
		claimNames = append(claimNames, name)
	}
	sort.Strings(claimNames)
	for _, name := range claimNames {
		if owner != nil {
			owned, err := owner(adapterRoot, name)
			if err != nil {
				return err
			}
			preserved[name] = owned && !casingAliasUnmanaged(filepath.Join(adapterRoot, name))
		}
		if expected[name] && !preserved[name] {
			localExpected[name] = true
		}
	}

	var stale []string
	for name := range managed {
		if !expected[name] && !preserved[name] {
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
			if preserved[name] {
				continue
			}
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

	ledger, err := ledgerPayload(localExpected)
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
// journaled copy only when an actual staging syscall reports a capability
// limitation. It never probes or changes a live destination during planning.
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
		before := len(mirror.plan.Targets)
		if err := mirror.stageLink(stageRoot, key, name, canonical, target); err != nil {
			if !linkcap.Unavailable(err) {
				return err
			}
			return mirror.stageCopy(stageRoot, key, name, content, target)
		}
		if len(mirror.plan.Targets) != before {
			alternative, err := stageCopySource(stageRoot, key, name, content, target)
			if err != nil {
				return err
			}
			mirror.plan.Targets[len(mirror.plan.Targets)-1].CopyFallback = alternative
		}
		return nil
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
	stagedEntry, err := stageCopySource(stageRoot, key, name, content, target)
	if err != nil {
		return err
	}
	mirror.plan.ReplaceEntry(staging.ClassAdapterLedger, key+"/entry/"+name, target, stagedEntry)
	return nil
}

func stageCopySource(stageRoot, key, name, content, target string) (string, error) {
	stagedEntry := filepath.Join(stageRoot, "adapters", key, "entries", name)
	if err := os.MkdirAll(filepath.Dir(stagedEntry), 0o700); err != nil {
		return "", fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	if err := os.RemoveAll(stagedEntry); err != nil {
		return "", fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	if err := copyTree(content, stagedEntry); err != nil {
		return "", fmt.Errorf("stage adapter mirror %s: %w", target, err)
	}
	// The live entry may currently be a link from an earlier symlink-mode run,
	// so the copy replaces the directory entry itself, not what it points at.
	return stagedEntry, nil
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

func rootKey(adapterRoot string) string {
	digest := sha256.Sum256([]byte(adapterRoot))
	return hex.EncodeToString(digest[:8])
}
