// Package adapters mirrors installed context into the directories each
// agent reads, with a managed-entries ledger per adapter root (Spec §10).
package adapters

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
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
//
// Sources overrides where a skill's content is read from while it is being
// staged. An installation mirrors the state it is about to commit, which does
// not exist at the canonical root yet, so it maps each name to the staged
// directory instead. A skill that is already current is left out and read from
// the canonical root.
type Group struct {
	Root    string
	Skills  []string
	Sources map[string]string
}

// source resolves where the content of one skill is read from.
func (group Group) source(name string) string {
	if staged, overridden := group.Sources[name]; overridden {
		return staged
	}
	return filepath.Join(group.Root, name)
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
	if err := protocoljson.Validate(payload); err != nil {
		return map[string]bool{}
	}
	var data struct {
		SchemaVersion int      `json:"schema_version"`
		Entries       []string `json:"entries"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil || data.SchemaVersion != LedgerSchemaVersion || data.Entries == nil {
		return map[string]bool{}
	}
	entries := map[string]bool{}
	for _, entry := range data.Entries {
		if !identifiers.Valid(entry) || entries[entry] {
			return map[string]bool{}
		}
		entries[entry] = true
	}
	return entries
}

// ledgerPayload renders the canonical ownership ledger bytes.
func ledgerPayload(entries map[string]bool) ([]byte, error) {
	var names []string
	for name, present := range entries {
		if present {
			names = append(names, name)
		}
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
		return nil, err
	}
	return append(payload, '\n'), nil
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
