// Package adapters mirrors installed context into the directories each
// agent reads, with a managed-entries ledger per adapter root (Spec §10).
package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/stateread"
)

// AgentPaths maps agent identifiers to their project-relative adapter
// directories (Spec §10.1). Migrated skill projections come from the registry;
// the remaining entries stay in the legacy catalog until their moves.
var AgentPaths = projectAgentPaths()

var legacyAgentPaths = map[string]string{
	"codex_cli":   ".codex/skills",
	"claude_code": ".claude/skills",
}

func projectAgentPaths() map[string]string {
	paths := make(map[string]string, len(legacyAgentPaths)+len(harnessregistry.IDs()))
	for agent, path := range legacyAgentPaths {
		paths[agent] = path
	}
	projectContext, err := harnessapi.NewContext(runtime.GOOS, harnessapi.ProjectScope, harnessapi.VerifiedRelease{})
	if err != nil {
		panic(fmt.Errorf("create project skill projection context: %w", err))
	}
	for _, id := range harnessregistry.IDs() {
		adapter, err := harnessregistry.Lookup(id)
		if err != nil {
			var refusal harnessapi.Refusal
			if errors.As(err, &refusal) && refusal.Code() == harnessapi.UnknownID {
				continue
			}
			panic(fmt.Errorf("lookup harness adapter %q: %w", id, err))
		}
		discovery, supported, err := registeredSkillProjection(string(id), projectContext, harnessapi.ProjectScope, adapter)
		if err != nil {
			panic(err)
		}
		if !supported || !discovery.Mirror() {
			continue
		}
		paths[string(id)] = oneSkillPath(string(id), discovery.Paths())
	}
	return paths
}

func registeredSkillProjection(agent string, context harnessapi.Context, scope harnessapi.Scope, adapter harnessapi.Adapter) (harnessapi.SkillDiscovery, bool, error) {
	capabilities := adapter.Capabilities()
	if !capabilities.Has(harnessapi.SkillProjection) {
		return harnessapi.SkillDiscovery{}, false, nil
	}
	discovery, err := adapter.Skills(context, scope)
	if err != nil {
		if scope == harnessapi.ProjectScope && capabilities.Has(harnessapi.NativeSkillDiscovery) {
			var refusal harnessapi.Refusal
			if errors.As(err, &refusal) && refusal.Code() == harnessapi.CapabilityUnsupported {
				return harnessapi.SkillDiscovery{}, false, nil
			}
		}
		return harnessapi.SkillDiscovery{}, false, fmt.Errorf("read skill projection for %q: %w", agent, err)
	}
	if !validSkillProjection(scope, discovery) {
		return harnessapi.SkillDiscovery{}, false, fmt.Errorf("invalid skill projection declaration for %q", agent)
	}
	return discovery, true, nil
}

func validSkillProjection(scope harnessapi.Scope, discovery harnessapi.SkillDiscovery) bool {
	valid := discovery.Scope() == scope
	switch scope {
	case harnessapi.ProjectScope:
		valid = valid && !discovery.NativeDiscovery() && discovery.Mirror()
	case harnessapi.UserScope:
		// User-scope adapters declare either a mirrored path or a native path.
		valid = valid && discovery.NativeDiscovery() != discovery.Mirror()
	default:
		valid = false
	}
	return valid
}

func oneSkillPath(agent string, paths harnessapi.NamedPaths) string {
	values := paths.Paths()
	if len(values) != 1 {
		panic(fmt.Errorf("skill projection for %q has %d paths; want exactly one", agent, len(values)))
	}
	return values[0].String()
}

func skillPath(agent string, scope harnessapi.Scope) (string, bool, error) {
	discovery, declared, err := registeredSkills(agent, scope)
	if declared {
		if err != nil {
			return "", false, err
		}
		paths := discovery.Paths().Paths()
		if len(paths) == 0 {
			return "", false, fmt.Errorf("harness_skill_path_missing")
		}
		return paths[0].String(), true, nil
	}
	if err != nil {
		return "", false, err
	}
	path, known := AgentPaths[agent]
	return path, known, nil
}

// NativeDiscoveryAgents discover the canonical .agents/skills directory
// natively and receive no project-level mirror; global installs mirror into
// the home-level .agents/skills for them (Spec §10.2).
var NativeDiscoveryAgents = nativeDiscoveryAgents()

func nativeDiscoveryAgents() map[string]bool {
	agents := map[string]bool{}
	for _, adapter := range envregistry.Registry {
		if adapter.SkillsDir == "" {
			agents[adapter.ID] = true
		}
	}
	for _, id := range harnessregistry.IDsWithCapability(harnessapi.NativeSkillDiscovery) {
		agents[string(id)] = true
	}
	return agents
}

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
//
// Admitted lists the local-source inputs this plan must not overwrite:
// canonical or absolute admitted paths. Production planning sets it to
// the closure node snapshots and the planner refuses a mirror that
// would overwrite any entry, in either direction, including through
// links, casing aliases or a changed parent. Empty disables the check.
type Group struct {
	Root     string
	Skills   []string
	Sources  map[string]string
	Admitted []string
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
// marker inside a copied directory). A casing alias of the requested name
// that the ledger does not own is unmanaged even when the requested
// spelling is listed: the ledger covers exact names, not filesystem
// case-equivalents. Inspection failures fail closed, never as absence.
func unmanagedConflict(target string, inLedger bool, source string) (bool, error) {
	info, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if inLedger && !casingAliasUnmanaged(target) {
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

// casingAliasUnmanaged reports whether the live entry's on-disk spelling
// differs by case from the requested target base. A ledger hit for the
// requested spelling does not cover such an alias: the caller falls through
// to symlink/marker proof instead of adopting blindly. Read failures fail
// closed by forcing the proof path.
func casingAliasUnmanaged(target string) bool {
	base := filepath.Base(target)
	entries, err := os.ReadDir(filepath.Dir(target)) // #nosec G304 -- adapter parent under validation
	if err != nil {
		return true
	}
	for _, entry := range entries {
		name := entry.Name()
		if name != base && strings.EqualFold(name, base) {
			return true
		}
	}
	return false
}

func readLedger(adapterRoot string) (map[string]bool, error) {
	path := filepath.Join(adapterRoot, LedgerName)
	state, err := stateread.ReadFile(path) // #nosec G304 -- adapter root is tool-managed
	if err != nil {
		return nil, err
	}
	if state.Kind == stateread.KindAbsent {
		return map[string]bool{}, nil
	}
	if err := protocoljson.Validate(state.Bytes); err != nil {
		return nil, stateread.UnusableError(path, err)
	}
	var data struct {
		SchemaVersion int      `json:"schema_version"`
		Entries       []string `json:"entries"`
	}
	decoder := json.NewDecoder(bytes.NewReader(state.Bytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return nil, stateread.UnusableError(path, err)
	}
	if data.SchemaVersion != LedgerSchemaVersion || data.Entries == nil {
		return nil, stateread.UnusableError(path, fmt.Errorf("invalid adapter ledger schema or entries"))
	}
	entries := map[string]bool{}
	for _, entry := range data.Entries {
		if !identifiers.Valid(entry) || entries[entry] {
			return nil, stateread.UnusableError(path, fmt.Errorf("invalid or repeated adapter entry %q", entry))
		}
		entries[entry] = true
	}
	return entries, nil
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
