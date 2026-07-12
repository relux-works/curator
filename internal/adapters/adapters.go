// Package adapters mirrors installed context into the directories each
// agent reads, with a managed-entries ledger per adapter root (Spec §10).
package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// AgentPaths maps agent identifiers to their project-relative adapter
// directories (Spec §10.1).
var AgentPaths = map[string]string{
	"codex_cli":   ".codex/skills",
	"claude_code": ".claude/skills",
	"gemini":      ".gemini/skills",
	"cursor":      ".cursor/rules",
}

// NativeDiscoveryAgents discover the canonical .agents/skills directory
// natively and receive no project-level mirror; global installs mirror into
// the home-level .agents/skills for them (Spec §10.2).
var NativeDiscoveryAgents = map[string]bool{"windsurf": true, "opencode": true}

// NativeDiscoveryHomePath is the home-relative mirror for native-discovery
// agents in the global scope.
const NativeDiscoveryHomePath = ".agents/skills"

// LedgerName is the managed-entries file inside every adapter root.
const LedgerName = ".csk-managed.json"

// LedgerSchemaVersion is the supported ledger schema.
const LedgerSchemaVersion = 1

// KnownAgents returns every recognized agent identifier.
func KnownAgents() map[string]bool {
	known := map[string]bool{}
	for agent := range AgentPaths {
		known[agent] = true
	}
	for agent := range NativeDiscoveryAgents {
		known[agent] = true
	}
	return known
}

// RequiredGitignoreEntries returns the generated paths a project must
// ignore for the selected agents (Spec §6.3).
func RequiredGitignoreEntries(agents []string) []string {
	entries := map[string]bool{".agents/": true}
	for _, agent := range agents {
		if rel, known := AgentPaths[agent]; known {
			entries[rel+"/"] = true
		}
	}
	var out []string
	for entry := range entries {
		out = append(out, entry)
	}
	sort.Strings(out)
	return out
}

// UnknownAgents returns the unrecognized names among agents, sorted.
func UnknownAgents(agents []string) []string {
	known := KnownAgents()
	seen := map[string]bool{}
	var unknown []string
	for _, agent := range agents {
		if !known[agent] && !seen[agent] {
			seen[agent] = true
			unknown = append(unknown, agent)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// Group is one canonical root with the skill names it holds.
type Group struct {
	Root   string
	Skills []string
}

// RefreshProject mirrors groups of canonical roots into the project-level
// adapter directories of the selected agents. All groups share one ledger
// per adapter root, so entries falling out of every group are removed in
// the same pass. Native-discovery agents get no project mirror.
func RefreshProject(projectRoot string, agents []string, groups []Group, mode string) error {
	roots := map[string]string{}
	for _, agent := range agents {
		if rel, known := AgentPaths[agent]; known {
			roots[agent] = filepath.Join(projectRoot, filepath.FromSlash(rel))
		}
	}
	return refresh(roots, groups, mode)
}

// RefreshGlobal mirrors the global canonical root into home-level adapter
// directories, including the native-discovery mirror.
func RefreshGlobal(home string, userHome string, agents []string, skills []string, mode string) error {
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
	return refresh(roots, []Group{{Root: canonical, Skills: skills}}, mode)
}

func refresh(adapterRoots map[string]string, groups []Group, mode string) error {
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
	for _, agent := range agents {
		adapterRoot := adapterRoots[agent]
		if err := os.MkdirAll(adapterRoot, 0o755); err != nil {
			return err
		}
		managed := readLedger(adapterRoot)
		// remove managed entries that fell out of every group
		var stale []string
		for name := range managed {
			if !expected[name] {
				stale = append(stale, name)
			}
		}
		sort.Strings(stale)
		for _, name := range stale {
			if err := removePath(filepath.Join(adapterRoot, name)); err != nil {
				return err
			}
		}
		for _, group := range groups {
			names := append([]string(nil), group.Skills...)
			sort.Strings(names)
			for _, name := range names {
				source := filepath.Join(group.Root, name)
				if _, err := os.Stat(source); err != nil {
					continue
				}
				target := filepath.Join(adapterRoot, name)
				conflict, err := unmanagedConflict(target, managed[name], source)
				if err != nil {
					return err
				}
				if conflict {
					return fmt.Errorf("adapter target already exists and is not managed: %s", target)
				}
				if err := refreshEntry(source, target, mode); err != nil {
					return err
				}
			}
		}
		if err := writeLedger(adapterRoot, expected); err != nil {
			return err
		}
	}
	return nil
}

func refreshEntry(source, target, mode string) error {
	switch mode {
	case "copy":
		if err := removePath(target); err != nil {
			return err
		}
		return copyTree(source, target)
	case "symlink":
		if err := removePath(target); err != nil {
			return err
		}
		return symlinkRel(source, target)
	default: // auto: symlink with copy fallback
		if err := removePath(target); err != nil {
			return err
		}
		if err := symlinkRel(source, target); err != nil {
			if err := removePath(target); err != nil {
				return err
			}
			return copyTree(source, target)
		}
		return nil
	}
}

func symlinkRel(source, target string) error {
	rel, err := filepath.Rel(filepath.Dir(target), source)
	if err != nil {
		rel = source
	}
	return os.Symlink(rel, target)
}

// unmanagedConflict reports whether target exists but is neither in the
// ledger nor recognizably ours (a symlink to the source, or an install
// marker inside a copied directory).
func unmanagedConflict(target string, inLedger bool, source string) (bool, error) {
	info, err := os.Lstat(target)
	if err != nil {
		return false, nil
	}
	if inLedger {
		return false, nil
	}
	if info.Mode()&os.ModeSymlink != 0 {
		resolvedTarget, err := filepath.EvalSymlinks(target)
		if err != nil {
			return true, nil
		}
		resolvedSource, err := filepath.EvalSymlinks(source)
		if err != nil {
			return true, nil
		}
		return resolvedTarget != resolvedSource, nil
	}
	if _, err := os.Stat(filepath.Join(target, ".csk-install.json")); err == nil {
		return false, nil
	}
	return true, nil
}

func readLedger(adapterRoot string) map[string]bool {
	payload, err := os.ReadFile(filepath.Join(adapterRoot, LedgerName)) // #nosec G304 -- adapter root is tool-managed
	if err != nil {
		return map[string]bool{}
	}
	var data struct {
		SchemaVersion int      `json:"schema_version"`
		Entries       []string `json:"entries"`
	}
	if err := json.Unmarshal(payload, &data); err != nil || data.SchemaVersion != LedgerSchemaVersion {
		return map[string]bool{}
	}
	entries := map[string]bool{}
	for _, entry := range data.Entries {
		entries[entry] = true
	}
	return entries
}

func writeLedger(adapterRoot string, entries map[string]bool) error {
	var names []string
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	if names == nil {
		names = []string{}
	}
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": LedgerSchemaVersion,
		"entries":        names,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(adapterRoot, LedgerName), append(payload, '\n'), 0o644)
}

func removePath(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return nil
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return os.Remove(path)
	}
	return os.RemoveAll(path)
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		payload, err := os.ReadFile(path) // #nosec G304 -- paths come from the walked tree
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, payload, info.Mode().Perm())
	})
}
