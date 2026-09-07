// Package envprofile manages installed context profiles (environments
// §9.2, manager profile §12.3): the profile store below the manager home,
// installation from git and path sources through joint resolution and the
// always-strict audit, lock publication, and the builtin local default
// profile that carries the migrated global scope (environments §9.4).
//
// Switching itself (linked materialization, backups, the marker, current
// pointers) lives in switch.go; the resolution source over git caches and
// path directories lives in gitsource.go.
//
// Machine overlays come from the policy's per-profile declarations
// (manager-config schema 2 via PolicyFromConfig) and join the closure
// beside the root; Direct is empty because live direct declarations
// (`global add`/`remove` writing into the lock) await the skill pipeline of
// a later stage. The §9.4 migration freezes the global
// skill set at migration time into the default lock instead — git tag and
// revision declarations only, since a skill member pins a commit with its
// source and branch-pinned or local skills have no representable pin here;
// live direct declarations (`global add`/`remove` writing into the lock)
// await the machine-config surface and skill pipeline of a later stage;
// scoped secret-material waivers (secret_material_waivers) have no
// machine-config surface yet, so the audit runs with no waivers; precedence
// is the default pair (higher-weight, winner-last).
package envprofile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Diagnostics (environments §1.1, §2.1, §9.7, manager profile §12.3).
const (
	DiagNameTaken       = "profile_name_taken"
	DiagRefConflict     = "profile_install_ref_conflict"
	DiagSourceInvalid   = "profile_source_invalid"
	DiagPathMissing     = "profile_source_path_missing"
	DiagPathUnreadable  = "profile_source_path_unreadable"
	DiagUpdateBlocked   = "profile_update_blocked"
	DiagInUse           = "profile_in_use"
	DiagNotFound        = "profile_not_found"
	DiagImportNameTaken = "profile_import_name_taken"
	// DiagLockUnavailable reports that the manager-home mutation lock
	// could not be acquired within the bounded wait (manager §2.5).
	DiagLockUnavailable = "environment_lock_unavailable"
)

// Source kinds (environments §1).
const (
	KindGit   = "git"
	KindPath  = "path"
	KindLocal = "local"
)

// DefaultProfile is the builtin local profile (environments §9.4).
const DefaultProfile = "default"

// Requirement is the declared root requirement as written.
type Requirement struct {
	Range    string `json:"range,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Revision string `json:"revision,omitempty"`
}

// Form names the requirement form: "range", "tag", or "revision".
func (r Requirement) Form() string {
	switch {
	case r.Range != "":
		return "range"
	case r.Tag != "":
		return "tag"
	default:
		return "revision"
	}
}

// Source is the install record of one profile. ImportedFromNative marks a
// path profile reassembled by the section 9.6 onboarding import; its
// environment markers record imported_from_native.
type Source struct {
	Kind               string      `json:"kind"`
	Git                string      `json:"git,omitempty"`
	Path               string      `json:"path,omitempty"`
	Directory          string      `json:"directory,omitempty"`
	Req                Requirement `json:"requirement"`
	ImportedFromNative bool        `json:"imported_from_native,omitempty"`
}

// ProfilesDir is the profile store below the manager home.
func ProfilesDir(home string) string { return filepath.Join(home, "profiles") }

// ProfileDir is one profile's directory.
func ProfileDir(home, name string) string { return filepath.Join(ProfilesDir(home), name) }

// CurrentFile records the machine current profile name.
func CurrentFile(home string) string { return filepath.Join(ProfilesDir(home), "current") }

// ScopedDir records per-scope current profiles.
func ScopedDir(home string) string { return filepath.Join(ProfilesDir(home), "scoped") }

// scopeFileName encodes a scope key (environments §9.3, e.g.
// "env:codex_cli") into a scope-record filename. The colon the key
// vocabulary mandates is a reserved filename character on Windows, so a
// literal scope key cannot name a file there and every scoped switch fails
// with profile_use_partial; the encoding keeps the key intact for every
// reader. Percent escapes first, so the mapping is one-to-one in both
// directions.
func scopeFileName(scope string) string {
	return strings.ReplaceAll(strings.ReplaceAll(scope, "%", "%25"), ":", "%3A")
}

// scopeKeyName decodes a scope-record filename back to its scope key.
// Filenames written before the encoding (a literal "env:<id>") carry no
// escape and decode to themselves, so manager homes created where a colon
// is a legal filename keep reading.
func scopeKeyName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "%3A", ":"), "%25", "%")
}

func sourcePath(home, name string) string {
	return filepath.Join(ProfileDir(home, name), "source.json")
}
func lockPath(home, name string) string { return filepath.Join(ProfileDir(home, name), "lock.json") }

// prevLockPath is the retained previous lock an update keeps beside the
// new one until garbage collection drops it (environments §9.2).
func prevLockPath(home, name string) string {
	return filepath.Join(ProfileDir(home, name), "lock.prev.json")
}

// Info is one installed or listed profile. Warnings carries the
// non-blocking audit findings of the install or update that produced it:
// context-system-module-present, mcp_command_unresolved, and unmatched
// waivers. Activation carries the per-adapter results of the §9.2 switch
// an install activation performed, if any.
type Info struct {
	Name       string
	Source     Source
	Lock       *contextlock.Lock
	LockHash   string
	Current    bool
	ScopedFor  []string
	Warnings   []string
	Activation []EntryResult
}

// validProfileName reports whether name may be installed.
func validProfileName(name string) bool { return identifiers.Valid(name) }

// readSource loads a profile's install record.
func readSource(home, name string) (Source, error) {
	payload, err := os.ReadFile(sourcePath(home, name)) // #nosec G304 -- profile name validated by callers
	if err != nil {
		if os.IsNotExist(err) {
			return Source{}, fmt.Errorf("%s: no installed profile %q", DiagNotFound, name)
		}
		return Source{}, err
	}
	if err := protocoljson.Validate(payload); err != nil {
		return Source{}, fmt.Errorf("%s: profile %q source is malformed: %v", DiagSourceInvalid, name, err)
	}
	var source Source
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	if err := decoder.Decode(&source); err != nil {
		return Source{}, fmt.Errorf("%s: profile %q source is malformed: %v", DiagSourceInvalid, name, err)
	}
	return source, nil
}

// marshalSource renders a profile's install record. Records reach the
// manager home only through the operation journal (op.publish), never by
// direct write.
func marshalSource(source Source) ([]byte, error) {
	payload, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// readLock loads a profile's lock and hash.
func readLock(home, name string) (*contextlock.Lock, string, error) {
	return contextlock.Read(lockPath(home, name))
}

// List returns every installed profile with its lock and currency. The
// machine gates come from the process configuration (see loadMachinePolicy);
// callers operating on any other manager home must use ListWithPolicy.
func List(home string) ([]Info, error) {
	policy, err := loadMachinePolicy()
	if err != nil {
		return nil, err
	}
	return ListWithPolicy(home, policy)
}

// ListWithPolicy lists under the machine gates of policy (environments
// §9.1). The CLI passes the gates of its already-loaded configuration.
func ListWithPolicy(home string, policy Policy) ([]Info, error) {
	op, err := beginOperation(home)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()
	return listLocked(op, home, policy)
}

// listLocked lists under the held operation lock.
func listLocked(op *operation, home string, policy Policy) ([]Info, error) {
	if err := ensureDefault(op, home, policy); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(ProfilesDir(home))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	machine, _ := Current(home)
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return nil, err
	}
	var out []Info
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !validProfileName(name) {
			continue
		}
		if _, err := os.Stat(sourcePath(home, name)); err != nil {
			if os.IsNotExist(err) {
				continue // reserved state (scoped/) is not a profile
			}
			return nil, fmt.Errorf("profile %q: %v", name, err)
		}
		source, err := readSource(home, name)
		if err != nil {
			return nil, fmt.Errorf("profile %q: %v", name, err)
		}
		lock, hash, err := readLock(home, name)
		if err != nil {
			return nil, fmt.Errorf("profile %q: %v", name, err)
		}
		info := Info{Name: name, Source: source, Lock: lock, LockHash: hash, Current: machine == name}
		for scope, current := range scoped {
			if current == name {
				info.ScopedFor = append(info.ScopedFor, scope)
			}
		}
		sort.Strings(info.ScopedFor)
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Current returns the machine current profile name, or "" when none is set.
func Current(home string) (string, error) {
	payload, err := os.ReadFile(CurrentFile(home)) // #nosec G304 -- manager home path
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(payload)), nil
}

// SetCurrent records the machine current profile.
func SetCurrent(home, name string) error {
	if err := os.MkdirAll(ProfilesDir(home), 0o755); err != nil {
		return err
	}
	return os.WriteFile(CurrentFile(home), []byte(name+"\n"), 0o644)
}

// ScopedCurrents returns the per-scope current map (scope key to profile).
// Record filenames decode through scopeKeyName; when a pre-encoding
// literal record and its encoded successor both exist, the encoded record
// wins — a new write always supersedes the stale spelling.
func ScopedCurrents(home string) (map[string]string, error) {
	out := map[string]string{}
	entries, err := os.ReadDir(ScopedDir(home))
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(ScopedDir(home), entry.Name())) // #nosec G304 -- scoped key file
		if err != nil {
			return nil, err
		}
		key := scopeKeyName(entry.Name())
		if _, seen := out[key]; seen && entry.Name() != scopeFileName(key) {
			continue
		}
		out[key] = strings.TrimSpace(string(payload))
	}
	return out, nil
}

// SetScoped records (or with clear=true, drops) a scope's current profile.
func SetScoped(home, scope, name string, clearScope bool) error {
	file := filepath.Join(ScopedDir(home), scopeFileName(scope))
	if clearScope {
		err := os.Remove(file)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		// A pre-encoding literal record superseded by the encoded one
		// must not resurrect the scope: drop it best-effort. It can only
		// exist where a colon is a legal filename, so any error here is
		// not the record's absence on this platform — but the encoded
		// record above is authoritative, and a stale scratch-home file
		// must never fail a clear.
		if legacy := filepath.Join(ScopedDir(home), scope); legacy != file {
			_ = os.Remove(legacy)
		}
		return nil
	}
	if err := os.MkdirAll(ScopedDir(home), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(file, []byte(name+"\n"), 0o644); err != nil {
		return err
	}
	if legacy := filepath.Join(ScopedDir(home), scope); legacy != file {
		_ = os.Remove(legacy)
	}
	return nil
}

// InUseByAnyScope reports whether name is current in any scope.
func InUseByAnyScope(home, name string) (bool, error) {
	machine, err := Current(home)
	if err != nil {
		return false, err
	}
	if machine == name {
		return true, nil
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return false, err
	}
	for _, current := range scoped {
		if current == name {
			return true, nil
		}
	}
	return false, nil
}

// Policy carries the machine gates for profile operations (environments
// §9.1, §12.1): the core §6.1 source allowlist for git members, the MCP
// package allowlist for mcp members, the audit revocations, the
// overlays_allowed composition policy, the per-profile overlay declarations,
// and the precedence primitives. An empty allowlist permits every
// identity (core §6.1); revocation and the audit canary always block
// regardless of any enabled flag — an advisory profile install does not
// exist.
type Policy struct {
	AllowedSources []string
	MCPAllowlist   []string
	Revocations    []string
	// OverlaysAllowed is the effective §12.1 composition policy: false
	// empties every overlay list at resolution (environments §12.2).
	OverlaysAllowed bool
	// Overlays are the machine overlay declarations per installed profile
	// (environments §6); they join the closure beside the root. A nil map
	// declares none.
	Overlays map[string][]OverlaySpec
	// OverlayDefaultWeight is the machine overlay_default_weight knob
	// (environments §12.1, default 1000).
	OverlayDefaultWeight int64
	// PrecedenceWinner and PrecedencePlacement are the section 6 pair of
	// independent primitives; empty means the pair default
	// (higher-weight, winner-last).
	PrecedenceWinner    string
	PrecedencePlacement string
	// RequireCurrent is the effective §12.1 require_current_profile knob:
	// nil means no requirement. RequireCurrentLocked reports whether the
	// system file locks the key (manager §1); only a locked requirement
	// refuses a machine-scope switch (environments §12.2).
	RequireCurrent       *string
	RequireCurrentLocked bool
	// Takeover is the explicit section 9.5 takeover flag carried by a
	// mutating operation (profile install, use, update, sync, env resolve
	// --repair): with it the operation backs up and takes over the
	// unmanaged files it would write, with the replace notice; without it
	// the operation fails rather than overwrite. It is operation-scoped
	// and never set from machine configuration.
	Takeover bool
}

// OverlaySpec is one machine overlay declaration for a profile
// (environments §6, §12.1): a git source with a range or exact form, or a
// path source with no form, an optional directory within a git snapshot,
// and an optional machine-assigned weight (nil takes the default).
type OverlaySpec struct {
	Source    string
	Range     string
	Tag       string
	Revision  string
	Directory string
	Weight    *int64
}

// ForbidsOverlays reports whether the composition policy empties every
// overlay list (environments §12.2).
func (p Policy) ForbidsOverlays() bool { return !p.OverlaysAllowed }

// EffectiveOverlays returns the overlay declarations for profile after the
// §12.2 composition policy: a forbidding policy empties every list, so
// resolution joins the root alone.
func (p Policy) EffectiveOverlays(profile string) []OverlaySpec {
	if p.ForbidsOverlays() {
		return nil
	}
	return p.Overlays[profile]
}

// Precedence returns the effective section 6 precedence pair, defaulting
// each primitive independently to the §12.1 default.
func (p Policy) Precedence() contextmaterialize.Precedence {
	precedence := contextmaterialize.DefaultPrecedence
	if p.PrecedenceWinner != "" {
		precedence.Winner = p.PrecedenceWinner
	}
	if p.PrecedencePlacement != "" {
		precedence.Placement = p.PrecedencePlacement
	}
	return precedence
}

// CheckMachineUse enforces a locked require_current_profile at the §9.2
// machine-scope switch itself (environments §12.2, manager §1): with the
// key locked to a profile name, a machine-scope switch to any other
// profile is a configuration error. A scoped switch, an unlocked knob,
// and the required profile itself carry no refusal. Every path that
// reaches the machine-scope switch — profile use (clear or not), profile
// install --use, first-install auto-activation, import --use, and resync —
// funnels through the single gate in useLocked, so one gate covers them
// all.
func (p Policy) CheckMachineUse(name string) error {
	if p.RequireCurrent == nil || *p.RequireCurrent == name {
		return nil
	}
	if !p.RequireCurrentLocked {
		return nil
	}
	return fmt.Errorf("environments.require_current_profile: profile use of %q is refused: the system configuration requires current profile %q", name, *p.RequireCurrent)
}

// PolicyFromConfig derives the profile machine gates from the loaded
// machine configuration (environments §9.1, §12.1): the core §6.1 source
// allowlist, the MCP package allowlist, and the audit revocations. The CLI
// calls this on its already-loaded cfg — which carries the system overlay
// with its locked keys and the CURATOR_CONFIG override — and threads the
// result into every profile operation, so the builtin default migration
// below enforces exactly the same gates as Install and Update. The
// migration path must never re-parse configuration into a weaker copy.
func PolicyFromConfig(cfg *config.Config) Policy {
	if cfg == nil {
		return Policy{OverlaysAllowed: true, OverlayDefaultWeight: config.DefaultOverlayWeight}
	}
	policy := Policy{
		AllowedSources:       cfg.AllowedSources,
		MCPAllowlist:         cfg.Env.MCPPackageAllowlist,
		Revocations:          cfg.Audit.Revocations,
		OverlaysAllowed:      cfg.Env.OverlaysAllowed,
		OverlayDefaultWeight: int64(cfg.Env.OverlayDefaultWeight),
		PrecedenceWinner:     cfg.Env.Precedence.Winner,
		PrecedencePlacement:  cfg.Env.Precedence.Placement,
		RequireCurrent:       cfg.Env.RequireCurrent,
		RequireCurrentLocked: cfg.Locked["environments.require_current_profile"],
	}
	if len(cfg.Env.Overlays) > 0 {
		policy.Overlays = map[string][]OverlaySpec{}
		for profile, list := range cfg.Env.Overlays {
			specs := make([]OverlaySpec, 0, len(list))
			for _, decl := range list {
				spec := OverlaySpec{
					Source: decl.Source, Range: decl.Range, Tag: decl.Tag,
					Revision: decl.Revision, Directory: decl.Directory,
				}
				if decl.Weight != nil {
					weight := int64(*decl.Weight)
					spec.Weight = &weight
				}
				specs = append(specs, spec)
			}
			policy.Overlays[profile] = specs
		}
	}
	return policy
}

// InstallOptions selects the source of one installation.
type InstallOptions struct {
	Operand   string
	Directory string
	Range     string
	Tag       string
	Revision  string
	As        string
	Use       bool
	Policy    Policy
	// Imported marks an installation reassembled by the section 9.6
	// onboarding import: the install record and the environment markers
	// carry imported_from_native. Only the import sets it.
	Imported bool
}

// Install installs one root context package as a profile: it resolves the
// closure, audits every member always-strict, writes the lock, and installs
// every member's store entry. A git operand takes at most one requirement
// flag (default range latest); a path operand takes none and no directory.
//
// Activation on install performs the §9.2 switch: when the machine has no
// current profile (first install) or the operator passes Use, Install
// attempts every entry, reports per-adapter results in Info.Activation,
// records the new current only when the whole scope materialized, and
// reports profile_use_partial when it did not (the lock is still written;
// only the activation is partial).
//
// Install holds the manager-home mutation lock and publishes its records
// through the operation journal (see lock.go).
func Install(home string, options InstallOptions) (Info, bool, bool, error) {
	op, err := beginOperation(home)
	if err != nil {
		return Info{}, false, false, err
	}
	defer func() { _ = op.close() }()
	return installLocked(op, home, options)
}

// installLocked installs under the held operation lock.
func installLocked(op *operation, home string, options InstallOptions) (Info, bool, bool, error) {
	if err := ensureDefault(op, home, options.Policy); err != nil {
		return Info{}, false, false, err
	}
	kind := installOperandKind(options.Operand)
	if kind == identity.SourceInvalid {
		return Info{}, false, false, sourceKindRefusal("operand", options.Operand)
	}
	isPath := kind == identity.SourcePath
	if isPath && (options.Range != "" || options.Tag != "" || options.Revision != "" || options.Directory != "") {
		return Info{}, false, false, fmt.Errorf("%s: a path operand takes no requirement flag or --directory", DiagRefConflict)
	}
	forms := 0
	for _, flag := range []string{options.Range, options.Tag, options.Revision} {
		if flag != "" {
			forms++
		}
	}
	if !isPath && forms > 1 {
		return Info{}, false, false, fmt.Errorf("%s: at most one of --range, --tag, --revision", DiagRefConflict)
	}
	var (
		source  Source
		input   contextresolve.Input
		manager *gitManager
		name    string
	)
	if isPath {
		manifest, err := contextpkg.LoadManifest(options.Operand)
		if err != nil {
			return Info{}, false, false, pathManifestDiag(options.Operand, err)
		}
		name = options.As
		if name == "" {
			name = manifest.Name
		}
		state, err := stateForPath(home, manifest.Name, options.Operand)
		if err != nil {
			return Info{}, false, false, err
		}
		source = Source{Kind: KindPath, Path: options.Operand, ImportedFromNative: options.Imported}
		input = contextresolve.Input{
			Root:      contextresolve.Requirement{Kind: contextlock.KindContext, Name: manifest.Name},
			RootState: state,
		}
	} else {
		manager = newGitManager(home).withPolicy(options.Policy)
		requirement := Requirement{Range: options.Range, Tag: options.Tag, Revision: options.Revision}
		if requirement.Range == "" && requirement.Tag == "" && requirement.Revision == "" {
			requirement.Range = "latest"
		}
		canonicalOperand, err := canonicalGit(options.Operand)
		if err != nil {
			return Info{}, false, false, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		source = Source{Kind: KindGit, Git: canonicalOperand, Directory: options.Directory, Req: requirement}
		resolved, rootName, err := manager.rootInput(options.Operand, options.Directory, requirement)
		if err != nil {
			return Info{}, false, false, err
		}
		name = options.As
		if name == "" {
			name = rootName
		}
		input = resolved
	}
	if !validProfileName(name) {
		return Info{}, false, false, fmt.Errorf("%s: profile name %q is not a portable identifier", DiagSourceInvalid, name)
	}
	if prior, err := readSource(home, name); err == nil {
		// Re-installing an installed source with the same requirement
		// re-resolves and is reported as an update; a different source
		// under the same name is taken. A git reinstall is an update, so
		// it delegates to updateLocked. A path reinstall is the §1
		// reinstall — the only operator path that refreshes the immutable
		// snapshot — so it re-resolves from the freshly computed
		// stateForPath result already in hand; routing it through
		// updateLocked would throw that snapshot away and re-pin the old
		// one, since update resolves a path root from the store.
		if prior == source {
			if isPath {
				return reinstallPathLocked(op, home, name, source, input, options)
			}
			info, moved, err := updateLocked(op, home, name, options.Policy)
			if err != nil {
				return Info{}, false, false, err
			}
			_ = moved
			return info, false, true, nil
		}
		return Info{}, false, false, fmt.Errorf("%s: profile %q is already installed", DiagNameTaken, name)
	}
	if manager == nil {
		manager = newGitManager(home).withPolicy(options.Policy)
	}
	// Machine overlays join the closure beside the root and resolve jointly
	// with it (environments §6); a forbidding composition policy resolves
	// the root alone.
	overlays, err := resolveOverlays(home, manager, name, options.Policy)
	if err != nil {
		return Info{}, false, false, err
	}
	input.Overlays = overlays
	input.MCPAllowlist = options.Policy.MCPAllowlist
	overlayInputDefaults(&input, options.Policy)
	result, err := contextresolve.Resolve(manager, input)
	if err != nil {
		return Info{}, false, false, err
	}
	warnings := resolutionWarnings(result)
	auditWarnings, err := auditAndStore(home, manager, result, options.Policy)
	if err != nil {
		return Info{}, false, false, err
	}
	warnings = append(warnings, auditWarnings...)
	sourcePayload, err := marshalSource(source)
	if err != nil {
		return Info{}, false, false, err
	}
	canonical, err := result.Lock.Canonical()
	if err != nil {
		return Info{}, false, false, err
	}
	records := map[string][]byte{sourcePath(home, name): sourcePayload, lockPath(home, name): canonical}
	if err := op.publish(records); err != nil {
		return Info{}, false, false, err
	}
	hash := contextlock.HashBytes(canonical)
	machine, err := Current(home)
	if err != nil {
		return Info{}, false, false, err
	}
	if machine != "" && !options.Use {
		info := Info{Name: name, Source: source, Lock: result.Lock, LockHash: hash, Current: machine == name, Warnings: warnings}
		return info, false, false, nil
	}
	// Activation performs the §9.2 switch under the same held operation:
	// every entry is attempted, per-adapter results are collected, and the
	// new current is recorded only when the whole scope materialized.
	// useLocked publishes the current through the journal on success and
	// leaves it unchanged with profile_use_partial on failure.
	results, switchErr := useLocked(op, home, name, "", "", false, options.Policy)
	machine, _ = Current(home)
	info := Info{Name: name, Source: source, Lock: result.Lock, LockHash: hash, Current: machine == name, Warnings: warnings, Activation: results}
	if switchErr != nil {
		return info, false, false, switchErr
	}
	return info, true, false, nil
}

// reinstallPathLocked re-resolves a same-source path reinstall from the fresh
// snapshot the caller already computed with stateForPath (environments §1:
// later edits change nothing until the operator reinstalls). It mirrors
// updateLocked's publish-and-resync shape — the new-member blocking checks,
// the lock-hash comparison, the lock plus previous-lock publication, and the
// resync of every scope already on this profile — but resolves the root from
// the fresh input rather than from the store snapshot, which is what makes a
// reinstall move the pin while an update must not.
//
// A reinstall is still the install row (cli/curator.md), so it honours the
// install row's flags: --use activates the installed root when it is not
// current (a machine with no current activates the same way a first install
// does), and --takeover covers the unmanaged files the activation or the
// resync would write (environments §9.5). Both ride the useLocked seam, so
// the locked require_current_profile gate applies here exactly as on every
// other machine-scope switch. Without --use and with a current recorded, a
// reinstall only re-materializes scopes already on the profile, exactly as
// an update does. The reinstall always reports updated, never a fresh
// install: activated reports whether the install row's activation ran.
func reinstallPathLocked(op *operation, home, name string, source Source, input contextresolve.Input, options InstallOptions) (Info, bool, bool, error) {
	policy := options.Policy
	manager := newGitManager(home).withPolicy(policy)
	oldLock, oldHash, err := readLock(home, name)
	if err != nil {
		return Info{}, false, false, err
	}
	overlays, err := resolveOverlays(home, manager, name, policy)
	if err != nil {
		return Info{}, false, false, err
	}
	input.Overlays = overlays
	input.MCPAllowlist = policy.MCPAllowlist
	overlayInputDefaults(&input, policy)
	result, err := contextresolve.Resolve(manager, input)
	if err != nil {
		return Info{}, false, false, err
	}
	oldMembers := map[string]bool{}
	for _, member := range oldLock.Members {
		oldMembers[contextresolve.Key(member.Kind, member.Name)] = true
	}
	for key, resolved := range result.Members {
		if oldMembers[key] {
			continue
		}
		entry, err := manager.ensureEntry(home, resolved)
		if err != nil {
			return Info{}, false, false, err
		}
		report, err := contextaudit.Detect(packageRoot(entry, resolved.Directory), pinOf(resolved), nil)
		if err != nil {
			return Info{}, false, false, err
		}
		if report.Blocking() {
			return Info{}, false, false, fmt.Errorf("%s: new member %s carries a blocking finding; the old lock stands", DiagUpdateBlocked, key)
		}
		if _, err := strictAuditMember(home, manager, resolved, entry, policy); err != nil {
			return Info{}, false, false, fmt.Errorf("%s: new member %s %v; the old lock stands", DiagUpdateBlocked, key, err)
		}
	}
	if result.LockHash == oldHash {
		// The snapshot is unchanged, so there is no lock to publish — but
		// the install row's activation still runs. The §9.5 stop-and-retry
		// lands here: the stopped attempt already published the lock, so
		// the retry resolves identically and must still take over the
		// unmanaged files and switch, never report success for no work.
		if activate, err := reinstallActivation(home, name, options.Use); err != nil {
			return Info{}, false, false, err
		} else if activate {
			return activateReinstall(op, home, name, source, result.Lock, oldHash, nil, policy)
		}
		machine, _ := Current(home)
		info := Info{Name: name, Source: source, Lock: result.Lock, LockHash: oldHash, Current: machine == name}
		return info, false, true, nil
	}
	warnings := resolutionWarnings(result)
	auditWarnings, err := auditAndStore(home, manager, result, policy)
	if err != nil {
		return Info{}, false, false, err
	}
	warnings = append(warnings, auditWarnings...)
	canonical, err := result.Lock.Canonical()
	if err != nil {
		return Info{}, false, false, err
	}
	oldCanonical, err := oldLock.Canonical()
	if err != nil {
		return Info{}, false, false, err
	}
	if err := op.publish(map[string][]byte{lockPath(home, name): canonical, prevLockPath(home, name): oldCanonical}); err != nil {
		return Info{}, false, false, err
	}
	hash := contextlock.HashBytes(canonical)
	if activate, err := reinstallActivation(home, name, options.Use); err != nil {
		return Info{}, false, false, err
	} else if activate {
		return activateReinstall(op, home, name, source, result.Lock, hash, warnings, policy)
	}
	if err := resyncCurrentScopes(op, home, name, policy); err != nil {
		return Info{}, false, false, err
	}
	machine, _ := Current(home)
	return Info{Name: name, Source: source, Lock: result.Lock, LockHash: hash, Current: machine == name, Warnings: warnings}, false, true, nil
}

// reinstallActivation reports whether a same-source path reinstall must run
// the install row's activation: the machine records no current (first
// installs activate and say so), or the operator passed --use and the
// installed root is not current. A --takeover without --use activates
// nothing by itself — takeover covers only the files the carrying operation
// would write (§9.5), and a reinstall that switches nothing writes no
// native surface — so the flag alone never flips this.
func reinstallActivation(home, name string, use bool) (bool, error) {
	machine, err := Current(home)
	if err != nil {
		return false, err
	}
	if machine == "" {
		return true, nil
	}
	return use && machine != name, nil
}

// activateReinstall runs the install row's activation for a reinstall: the
// §9.2 machine-scope switch through useLocked — with the operation's
// takeover flag and under the locked require_current_profile gate — then
// the resync of every scoped current already on the profile. A partial
// switch leaves the recorded current unchanged and returns the switch
// error, exactly as a fresh install does; the lock work (published above
// when the pin moved) still stands.
func activateReinstall(op *operation, home, name string, source Source, lock *contextlock.Lock, hash string, warnings []string, policy Policy) (Info, bool, bool, error) {
	results, switchErr := useLocked(op, home, name, "", "", false, policy)
	machine, _ := Current(home)
	info := Info{Name: name, Source: source, Lock: lock, LockHash: hash, Current: machine == name, Warnings: warnings, Activation: results}
	if switchErr != nil {
		return info, false, true, switchErr
	}
	if err := resyncScopedScopes(op, home, name, policy); err != nil {
		return info, false, true, err
	}
	machine, _ = Current(home)
	info.Current = machine == name
	return info, true, true, nil
}

// Update re-resolves the root (and overlays, when the machine declares any)
// from the declared requirement, fetching new candidates. A blocking
// finding on a member new to the lock leaves the old lock in place with
// profile_update_blocked. A root pinned by tag or revision is reported as
// pinned and does not move. A path root resolves from the immutable
// snapshot the store already holds under the old lock's state pin and never
// reads the source directory again (environments §1); overlays still
// re-resolve below, so a path root with git overlays is a legitimate
// update. The machine gates come from the process configuration (see
// loadMachinePolicy); callers operating on any other manager home must use
// UpdateWithPolicy.
func Update(home, name string) (Info, bool, error) {
	policy, err := loadMachinePolicy()
	if err != nil {
		return Info{}, false, err
	}
	return UpdateWithPolicy(home, name, policy)
}

// UpdateWithPolicy re-resolves under the machine gates of policy
// (environments §9.1): the source allowlist, the MCP package allowlist, and
// the always-strict audit with revocation and the canary. The CLI passes the
// machine configuration; tests inject narrowing policies directly.
func UpdateWithPolicy(home, name string, policy Policy) (Info, bool, error) {
	op, err := beginOperation(home)
	if err != nil {
		return Info{}, false, err
	}
	defer func() { _ = op.close() }()
	return updateLocked(op, home, name, policy)
}

// updateLocked re-resolves under the held operation lock.
func updateLocked(op *operation, home, name string, policy Policy) (Info, bool, error) {
	if err := ensureDefault(op, home, policy); err != nil {
		return Info{}, false, err
	}
	source, err := readSource(home, name)
	if err != nil {
		return Info{}, false, err
	}
	oldLock, oldHash, err := readLock(home, name)
	if err != nil {
		return Info{}, false, err
	}
	manager := newGitManager(home).withPolicy(policy)
	var input contextresolve.Input
	switch source.Kind {
	case KindGit:
		if source.Req.Tag != "" || source.Req.Revision != "" {
			machine, _ := Current(home)
			info := Info{Name: name, Source: source, Lock: oldLock, Current: machine == name}
			if _, hash, err := readLock(home, name); err == nil {
				info.LockHash = hash
			}
			return info, false, nil
		}
		if err := manager.fetch(source.Git); err != nil {
			return Info{}, false, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		input, err = manager.inputFor(source)
		if err != nil {
			return Info{}, false, err
		}
	case KindPath:
		// Environments §1: installation copies the directory tree into
		// the profile store as an immutable snapshot and never reads
		// the source directory again. Update resolves the root from
		// the snapshot the store already holds under the old lock's
		// state_sha256 pin — never from source.Path, which may have
		// been edited or deleted (the §9.6 import deletes its staging
		// directory, so every imported profile's source.Path names no
		// existing entry). Overlays re-resolve below the switch, so a
		// path root with git overlays still moves on update.
		rootMember, ok := oldLock.RootMember()
		if !ok || rootMember.StateHash == "" {
			return Info{}, false, fmt.Errorf("%s: profile %q lock carries no state pin for its path root", DiagSourceInvalid, name)
		}
		entry := contextstore.EntryDir(home, contextlock.KindContext, oldLock.Root, rootMember.StateHash)
		manifest, err := contextpkg.LoadManifest(entry)
		if err != nil {
			return Info{}, false, fmt.Errorf("%s: profile %q path snapshot cannot be read: %v", DiagSourceInvalid, name, err)
		}
		if manifest.Name != oldLock.Root {
			return Info{}, false, fmt.Errorf("%s: profile %q snapshot names %q, want lock root %q", DiagSourceInvalid, name, manifest.Name, oldLock.Root)
		}
		input = contextresolve.Input{
			Root:      contextresolve.Requirement{Kind: contextlock.KindContext, Name: oldLock.Root},
			RootState: &contextresolve.StatePackage{StateHash: rootMember.StateHash, Manifest: packageOf(manifest)},
		}
	default:
		return Info{}, false, fmt.Errorf("%s: profile %q is the builtin local profile and does not move", DiagUpdateBlocked, name)
	}
	overlays, err := resolveOverlays(home, manager, name, policy)
	if err != nil {
		return Info{}, false, err
	}
	input.Overlays = overlays
	input.MCPAllowlist = policy.MCPAllowlist
	overlayInputDefaults(&input, policy)
	result, err := contextresolve.Resolve(manager, input)
	if err != nil {
		return Info{}, false, err
	}
	oldMembers := map[string]bool{}
	for _, member := range oldLock.Members {
		oldMembers[contextresolve.Key(member.Kind, member.Name)] = true
	}
	for key, resolved := range result.Members {
		if oldMembers[key] {
			continue
		}
		entry, err := manager.ensureEntry(home, resolved)
		if err != nil {
			return Info{}, false, err
		}
		report, err := contextaudit.Detect(packageRoot(entry, resolved.Directory), pinOf(resolved), nil)
		if err != nil {
			return Info{}, false, err
		}
		if report.Blocking() {
			return Info{}, false, fmt.Errorf("%s: new member %s carries a blocking finding; the old lock stands", DiagUpdateBlocked, key)
		}
		if _, err := strictAuditMember(home, manager, resolved, entry, policy); err != nil {
			return Info{}, false, fmt.Errorf("%s: new member %s %v; the old lock stands", DiagUpdateBlocked, key, err)
		}
	}
	if result.LockHash == oldHash {
		machine, _ := Current(home)
		return Info{Name: name, Source: source, Lock: result.Lock, LockHash: oldHash, Current: machine == name}, false, nil
	}
	warnings := resolutionWarnings(result)
	auditWarnings, err := auditAndStore(home, manager, result, policy)
	if err != nil {
		return Info{}, false, err
	}
	warnings = append(warnings, auditWarnings...)
	canonical, err := result.Lock.Canonical()
	if err != nil {
		return Info{}, false, err
	}
	oldCanonical, err := oldLock.Canonical()
	if err != nil {
		return Info{}, false, err
	}
	// The old lock is retained beside the new one until the next garbage
	// collection so that a stale managed home can still be identified
	// (environments §9.2).
	if err := op.publish(map[string][]byte{lockPath(home, name): canonical, prevLockPath(home, name): oldCanonical}); err != nil {
		return Info{}, false, err
	}
	hash := contextlock.HashBytes(canonical)
	if err := resyncCurrentScopes(op, home, name, policy); err != nil {
		return Info{}, false, err
	}
	machine, _ := Current(home)
	return Info{Name: name, Source: source, Lock: result.Lock, LockHash: hash, Current: machine == name, Warnings: warnings}, true, nil
}

// Remove deletes a profile that is current in no scope and an overlay of
// none. With purge, in-place surfaces recorded by its markers, the markers,
// and the backup generations are removed too. Remove holds the
// manager-home mutation lock; the profile directory goes with one RemoveAll
// under it, which is already atomic at the granularity the manager
// observes, so no journal is needed for the removal itself.
func Remove(home, name string, purge bool) error {
	op, err := beginOperation(home)
	if err != nil {
		return err
	}
	defer func() { _ = op.close() }()
	if _, err := readSource(home, name); err != nil {
		return err
	}
	inUse, err := InUseByAnyScope(home, name)
	if err != nil {
		return err
	}
	if inUse {
		return fmt.Errorf("%s: profile %q is current in a scope", DiagInUse, name)
	}
	// A profile that is an overlay member of another installed profile's
	// lock stays until that profile stops declaring it (environments
	// §9.2): removal would strand the overlay declaration.
	if owner, ok := overlayOwner(home, name); ok {
		return fmt.Errorf("%s: profile %q is named as an overlay of installed profile %q", DiagInUse, name, owner)
	}
	if purge {
		if err := purgeHomes(home, name); err != nil {
			return err
		}
	}
	return os.RemoveAll(ProfileDir(home, name))
}

// overlayOwner reports whether the root package of profile name is an
// overlay member of another installed profile's lock (environments §9.2).
// Locks are ground truth: a machine overlay declaration resolves into an
// overlay-flagged lock member, so a lock scan sees exactly what resolution
// joined.
func overlayOwner(home, name string) (string, bool) {
	rootLock, _, err := readLock(home, name)
	if err != nil || rootLock == nil {
		return "", false
	}
	entries, err := os.ReadDir(ProfilesDir(home))
	if err != nil {
		return "", false
	}
	owners := []string{}
	for _, entry := range entries {
		other := entry.Name()
		if !entry.IsDir() || other == name || !validProfileName(other) {
			continue
		}
		otherLock, _, err := readLock(home, other)
		if err != nil || otherLock == nil {
			continue
		}
		for _, member := range otherLock.Members {
			if member.Overlay && member.Name == rootLock.Root {
				owners = append(owners, other)
				break
			}
		}
	}
	sort.Strings(owners)
	if len(owners) == 0 {
		return "", false
	}
	return owners[0], true
}

// auditAndStore audits every resolved member always-strict and installs its
// store entry. A blocking finding fails the operation; system modules and
// unresolved MCP commands are reported as warnings.
//
// Every member passes the manager §7 source audit in strict mode
// (environments §9.1) via strictAuditMember: the static canary (always
// blocking), revocation, and the deterministic detectors — regardless of any
// enabled flag. An advisory profile install does not exist.
func auditAndStore(home string, manager *gitManager, result *contextresolve.Result, policy Policy) ([]string, error) {
	var warnings []string
	keys := make([]string, 0, len(result.Members))
	for key := range result.Members {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		resolved := result.Members[key]
		entry, err := manager.ensureEntry(home, resolved)
		if err != nil {
			return nil, err
		}
		report, err := contextaudit.Detect(packageRoot(entry, resolved.Directory), pinOf(resolved), nil)
		if err != nil {
			return nil, err
		}
		if report.Blocking() {
			return nil, fmt.Errorf("%s: member %s carries a blocking %s finding", DiagSourceInvalid, key, contextaudit.ClassSecretMaterial)
		}
		gateWarnings, err := strictAuditMember(home, manager, resolved, entry, policy)
		if err != nil {
			return nil, fmt.Errorf("%s: member %s %v", DiagSourceInvalid, key, err)
		}
		warnings = append(warnings, gateWarnings...)
		for _, unmatched := range report.Waivers {
			if unmatched.Diagnostic == contextaudit.DiagWaiverUnmatched {
				warnings = append(warnings, unmatched.Diagnostic)
			}
		}
		if resolved.Kind == contextlock.KindContext {
			root := packageRoot(entry, resolved.Directory)
			if manifest, err := contextpkg.LoadManifest(root); err == nil {
				for _, system := range contextaudit.SystemModules(resolved.Name, manifest) {
					warnings = append(warnings, contextaudit.ClassSystemModulePresent+": "+system.Package+"/"+system.Path)
				}
			}
		}
		if resolved.Kind == contextlock.KindMCP {
			if warning, ok := checkMCPCommand(packageRoot(entry, resolved.Directory)); ok {
				warnings = append(warnings, warning)
			}
		}
	}
	return warnings, nil
}

// pinOf renders the member pin as the audit spells it.
func pinOf(resolved contextresolve.Resolved) string {
	if resolved.Commit != "" {
		return "commit " + resolved.Commit
	}
	return "state sha256:" + resolved.StateHash
}

// strictAuditMember runs the manager §7 source audit in strict mode over one
// member (environments §9.1): raw-tree hashing, the static canary whose
// failure always blocks, the deterministic detectors, and revocation.
// Revocation and the canary apply regardless of any enabled flag: an
// advisory profile install does not exist. A path package has no network
// identity: its revocation identity is its state hash, and the core §6.1
// network allowlist does not apply (local sources bypass it).
//
// canaryPasses is a seam for the narrowing test that proves the canary's
// blocking role on this path: forcing it to fail must refuse the member.
var canaryPasses = audit.CanaryPasses

func strictAuditMember(home string, manager *gitManager, resolved contextresolve.Resolved, entry string, policy Policy) ([]string, error) {
	snapshot := packageRoot(entry, resolved.Directory)
	if !canaryPasses() {
		return nil, fmt.Errorf("audit blocked: audit canary failed: detectors are not producing expected findings")
	}
	contentHash, err := hashing.ContentSHA256(snapshot, nil)
	if err != nil {
		return nil, fmt.Errorf("audit blocked: %v", err)
	}
	sourceForRevocation := resolved.Source
	gitForRevocation := resolved.Source
	if raw, ok := manager.fetchRaw[resolved.Source]; ok {
		gitForRevocation = raw
	}
	if resolved.StateHash != "" {
		sourceForRevocation = resolved.StateHash
	}
	if reason := audit.RevocationFor(policy.Revocations, contentHash, sourceForRevocation, gitForRevocation); reason != "" {
		return nil, fmt.Errorf("audit blocked: %s is revoked", reason)
	}
	cfg := &config.Config{
		Path: filepath.Join(home, "config.json"),
		Audit: config.Audit{
			Enabled: true, Mode: "strict", FailOn: "high", Backend: "null",
			Revocations: policy.Revocations,
		},
	}
	subject := audit.Subject{
		Name:          resolved.Name,
		Source:        sourceForRevocation,
		Git:           gitForRevocation,
		Commit:        resolved.Commit,
		Snapshot:      snapshot,
		SchemaVersion: 3,
		Capabilities:  capabilities.ImplicitNone(),
	}
	warnings, errs := audit.Gate(cfg, []audit.Subject{subject})
	if len(errs) > 0 {
		return warnings, fmt.Errorf("audit blocked: %s", strings.Join(errs, "; "))
	}
	return warnings, nil
}

// pathManifestDiag maps a manifest-load failure on a path operand onto the
// section 1.1 diagnostic: an operand naming no existing filesystem entry
// is profile_source_path_missing; an operand that exists but cannot be
// read is profile_source_path_unreadable (§8.4: a failed read is never
// absence and never invalid); every other failure is
// profile_source_invalid. A failed read below an existing tree surfaces
// from the snapshot copy as profile_source_path_unreadable.
func pathManifestDiag(operand string, err error) error {
	if _, statErr := os.Stat(operand); statErr != nil {
		if os.IsNotExist(statErr) {
			return fmt.Errorf("%s: path %q names no existing filesystem entry", DiagPathMissing, operand)
		}
		if errors.Is(statErr, fs.ErrPermission) || os.IsPermission(statErr) {
			return fmt.Errorf("%s: path %q cannot be read: %v", DiagPathUnreadable, operand, statErr)
		}
	}
	if errors.Is(err, fs.ErrPermission) || os.IsPermission(err) {
		return fmt.Errorf("%s: path %q cannot be read: %v", DiagPathUnreadable, operand, err)
	}
	return fmt.Errorf("%s: %v", DiagSourceInvalid, err)
}

// stateForPath installs the path root into the store and returns its state
// package for resolution. The store already reports the §1.1 diagnostic
// that leads — missing, unreadable, or invalid — so a store error keeps
// its leading diagnostic and is never wrapped in a second
// profile_source_invalid (§8.4: unreadable evidence is reported as
// unreadable, never as something else).
func stateForPath(home, name, dir string) (*contextresolve.StatePackage, error) {
	entry, hash, err := contextstore.EnsureState(home, contextlock.KindContext, name, dir)
	if err != nil {
		return nil, preservePathDiag(err)
	}
	manifest, err := contextpkg.LoadManifest(entry)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	return &contextresolve.StatePackage{StateHash: hash, Manifest: packageOf(manifest)}, nil
}

// preservePathDiag keeps a store error's leading §1.1 diagnostic: an
// error the store already classified (missing, unreadable, invalid) is
// returned as-is so the specific diagnostic leads; any other failure is
// an invalid source tree.
func preservePathDiag(err error) error {
	message := err.Error()
	for _, leading := range []string{DiagPathMissing, DiagPathUnreadable, DiagSourceInvalid} {
		if message == leading || strings.HasPrefix(message, leading+":") {
			return err
		}
	}
	return fmt.Errorf("%s: %v", DiagSourceInvalid, err)
}

// checkMCPCommand warns mcp_command_unresolved when a stdio server's command
// does not resolve on the operator's PATH.
func checkMCPCommand(entry string) (string, bool) {
	manifest, err := contextpkg.LoadMCP(entry)
	if err != nil {
		return "", false
	}
	if manifest.Server.Transport != "stdio" {
		return "", false
	}
	if _, err := lookPath(manifest.Server.Command); err != nil {
		return "mcp_command_unresolved: " + manifest.Name, true
	}
	return "", false
}

// sourceKindRefusal is the refusal for a spelling that is neither kind. A
// file:// remote keeps the F12 wording: it is refused because it carries no
// network identity, which is exactly why the landed discriminator refuses
// it as neither kind, and the boundary is the same one canonicalGit and
// gateSource enforce for a source that reaches them.
func sourceKindRefusal(what, spelling string) error {
	if isFileRemote(strings.TrimSpace(spelling)) {
		return fmt.Errorf("%s: file:// git sources carry no network identity and are not accepted", DiagSourceInvalid)
	}
	return fmt.Errorf("%s: %s %q is neither a git source nor a path", DiagSourceInvalid, what, spelling)
}

// installOperandKind classifies a `profile install <git-url|path>` operand
// through the one discriminator (identity.ClassifySource, the same helper
// the config reader and `profile compose add` use) and never by probing the
// filesystem: a directory in the operator's working directory must never
// shadow a git identity, because a `path` source bypasses the core §6.1
// network allowlist by design.
//
// The operand vocabulary of this row is wider than an overlay `source` in
// one place, and the difference is stated rather than silent. `canonicalGit`
// accepts an already-canonical `host/path` identity and `ensureRepo` clones
// it over https, so `curator profile install github.com/example/x` is a git
// install today. The landed manager-config-v2 discriminator classifies that
// same bare spelling as a `path` — it has to, because `packages/team-context`
// is a declarable path overlay and the two are syntactically identical.
// Following the classification here would silently turn a git install into a
// local one and hand a planted `./github.com/example/x` directory the
// allowlist bypass that F14 exists to prevent, so a spelling the
// discriminator calls `path` that is also a valid canonical network identity
// stays `git` on this row. No operand that reached the network before this
// change reaches a local path now -- `packages/team` was a git install
// operand and still is. One class does move, in the other direction: a
// colon in a later segment (`packages/team:context`) was unclassifiable and
// refused at canonicalGit before, and is a path operand now, which is what
// the landed discriminator decides it is.
//
// The operand is trimmed before classification because every downstream git
// entry point (canonicalGit, identity.Parse) already trims; the overlay
// reader does not trim, because the schema does not.
func installOperandKind(operand string) identity.SourceKind {
	trimmed := strings.TrimSpace(operand)
	kind := identity.ClassifySource(trimmed)
	if kind == identity.SourcePath && identity.ValidCanonical(trimmed) {
		return identity.SourceGit
	}
	return kind
}

// canonicalGit normalizes a git operand onto its core §6.1 canonical source
// identity (environments §1) through identity.Parse: SSH and HTTPS spellings
// of one repository yield one identity, a trailing .git is stripped, the
// host is lowercased. A malformed network source is rejected. A file://
// remote carries no network identity, has no valid lock or marker shape
// (context-lock-v1 admits no file member), and is rejected here — the same
// boundary that rejects a malformed network source — with
// profile_source_invalid. Hermetic tests use git insteadOf rewrites onto
// fake network identities instead. An already-canonical host/path passes
// through unchanged.
func canonicalGit(operand string) (string, error) {
	trimmed := strings.TrimSpace(operand)
	if trimmed == "" || identity.ValidCanonical(trimmed) {
		return trimmed, nil
	}
	if isFileRemote(trimmed) {
		return "", fmt.Errorf("file:// git sources carry no network identity and are not accepted")
	}
	canonical, err := identity.Parse(trimmed)
	if err != nil {
		return "", err
	}
	if canonical == "" {
		return trimmed, nil
	}
	return canonical, nil
}

// isFileRemote reports the file: URL prefix (any case), mirroring the local
// classification in identity.Parse so no spelling of it can pass as a local
// source. A bare local path never carries a scheme and is unaffected.
func isFileRemote(raw string) bool {
	return strings.HasPrefix(strings.ToLower(raw), "file:")
}

// defaultManifest is the synthesized local root of the builtin default
// profile (environments §9.4): name default, version 0.0.0, no context
// member, so it declares no root-context surface and materializes skills
// alone. Its store key and pin are the state hash of this exact tree.
const defaultManifest = `{"schema_version": 1, "name": "default", "version": "0.0.0"}` + "\n"

// EnsureDefault creates the builtin local default profile on first use of
// the profile surface (environments §9.4): it materializes the synthesized
// local root into the store under its state hash, pins that hash in the
// lock, and carries the machine's global skill set at migration time as
// skill lock members. A machine that never installs another profile observes
// no behavior change: default simply is the current profile.
//
// The lock is written before the install record so a failed migration
// retries on the next call instead of stranding a sourceless profile.
// EnsureDefault is the locked entry point; ensureDefault runs under the
// held operation lock.
//
// The machine gates come from the process configuration (see
// loadMachinePolicy); callers operating on any other manager home must use
// EnsureDefaultWithPolicy.
func EnsureDefault(home string) error {
	policy, err := loadMachinePolicy()
	if err != nil {
		return err
	}
	return EnsureDefaultWithPolicy(home, policy)
}

// EnsureDefaultWithPolicy creates the builtin default profile under the
// machine gates of policy (environments §9.1). The CLI passes the gates of
// its already-loaded configuration.
func EnsureDefaultWithPolicy(home string, policy Policy) error {
	op, err := beginOperation(home)
	if err != nil {
		return err
	}
	defer func() { _ = op.close() }()
	return ensureDefault(op, home, policy)
}

func ensureDefault(op *operation, home string, policy Policy) error {
	if _, err := readSource(home, DefaultProfile); err == nil {
		return nil
	}
	staging, err := os.MkdirTemp("", "curator-default-root-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(staging) }()
	if err := os.WriteFile(filepath.Join(staging, contextpkg.ManifestName), []byte(defaultManifest), 0o644); err != nil {
		return err
	}
	_, key, err := contextstore.EnsureState(home, contextlock.KindContext, DefaultProfile, staging)
	if err != nil {
		return err
	}
	members := []contextlock.Member{
		{Kind: contextlock.KindContext, Name: DefaultProfile, StateHash: key, Version: "0.0.0"},
	}
	skills, err := migrateGlobalSkills(home, policy)
	if err != nil {
		return err
	}
	members = append(members, skills...)
	lock := &contextlock.Lock{Root: DefaultProfile, Members: members}
	lock.Sort()
	canonical, err := lock.Canonical()
	if err != nil {
		return err
	}
	sourcePayload, err := marshalSource(Source{Kind: KindLocal})
	if err != nil {
		return err
	}
	return op.publish(map[string][]byte{
		lockPath(home, DefaultProfile):   canonical,
		sourcePath(home, DefaultProfile): sourcePayload,
	})
}

// migrateGlobalSkills snapshots the machine's global skill set at migration
// time (environments §9.4) as skill lock members: every global declaration
// that names a commit — a git tag or revision — is fetched, pinned, stored,
// and audited exactly like an installed member, under the machine gates of
// policy (environments §9.1: the source allowlist, revocation, and the
// always-strict audit). A declaration the lock cannot represent — a branch
// pin, or a local skill with no source identity, since a skill member pins a
// commit with its source and a state pin admits only a context member — is
// left for the skill-pipeline stage that owns live direct declarations (see
// the stage bounds in the package doc).
func migrateGlobalSkills(home string, policy Policy) ([]contextlock.Member, error) {
	global, err := manifest.Load(filepath.Join(home, "global"))
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	if global == nil {
		return nil, nil
	}
	manager := newGitManager(home).withPolicy(policy)
	var members []contextlock.Member
	for _, decl := range global.Skills {
		if decl.Git == "" || (decl.Ref.Kind != "tag" && decl.Ref.Kind != "revision") {
			continue
		}
		identity, err := canonicalGit(decl.Git)
		if err != nil {
			return nil, fmt.Errorf("%s: migrate global skill %q: %v", DiagSourceInvalid, decl.Name, err)
		}
		manager.recordRaw(identity, decl.Git)
		if err := manager.fetch(identity); err != nil {
			return nil, fmt.Errorf("%s: migrate global skill %q: %v", DiagSourceInvalid, decl.Name, err)
		}
		dir := manager.repoDir(identity)
		resolved, err := gitops.Resolve(dir, decl.Ref.Kind, decl.Ref.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: migrate global skill %q: %v", DiagSourceInvalid, decl.Name, err)
		}
		entry, err := contextstore.EnsureGit(home, contextlock.KindSkill, decl.Name, dir, resolved.Commit)
		if err != nil {
			return nil, err
		}
		if report, err := contextaudit.Detect(entry, "commit "+resolved.Commit, nil); err != nil {
			return nil, err
		} else if report.Blocking() {
			return nil, fmt.Errorf("%s: migrated global skill %q carries a blocking %s finding", DiagSourceInvalid, decl.Name, contextaudit.ClassSecretMaterial)
		}
		migrated := contextresolve.Resolved{
			Kind: contextlock.KindSkill, Name: decl.Name, Source: identity,
			Commit: resolved.Commit,
		}
		if _, err := strictAuditMember(home, manager, migrated, entry, policy); err != nil {
			return nil, fmt.Errorf("%s: migrated global skill %q %v", DiagSourceInvalid, decl.Name, err)
		}
		members = append(members, contextlock.Member{
			Kind: contextlock.KindSkill, Name: decl.Name, Source: identity, Commit: resolved.Commit,
		})
	}
	return members, nil
}

// loadMachinePolicy resolves the machine gates for profile operations that
// carry no explicit Policy (the bare List, Use, Sync, Update, and
// EnsureDefault entry points, whose migration runs before any caller-held
// policy exists): allowed_sources and audit.revocations from the effective
// machine configuration. It goes through config.Load on the process
// configuration path, so the system overlay with its locked keys and the
// CURATOR_CONFIG override apply exactly as they do for the CLI's
// already-loaded Policy — the migration never re-parses configuration into
// a weaker copy. Callers that already hold a Policy (Install,
// UpdateWithPolicy, and every WithPolicy entry point) never consult it.
//
// Production always calls with home == cfg.Home(), so the loaded path is the
// home's own configuration; library callers operating on any other manager
// home must use the WithPolicy entry points.
//
// A missing configuration file yields an empty policy (permits all, no
// revocations): absence is legitimate. Any other read or parse failure fails
// the caller — a failed read is never an empty policy.
func loadMachinePolicy() (Policy, error) {
	path := config.UserPath()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return Policy{OverlaysAllowed: true}, nil
		}
		return Policy{}, fmt.Errorf("%s: read the machine configuration: %v", DiagSourceInvalid, err)
	}
	cfg, err := config.Load(path, nil)
	if err != nil {
		return Policy{}, fmt.Errorf("%s: read the machine configuration: %v", DiagSourceInvalid, err)
	}
	return PolicyFromConfig(cfg), nil
}

// resyncCurrentScopes re-materializes every scope whose current profile is
// name after its lock moved, under the held operation lock.
func resyncCurrentScopes(op *operation, home, name string, policy Policy) error {
	machine, err := Current(home)
	if err != nil {
		return err
	}
	if machine == name {
		if _, err := useLocked(op, home, name, "", "", false, policy); err != nil {
			return err
		}
	}
	return resyncScopedScopes(op, home, name, policy)
}

// resyncScopedScopes re-materializes every scoped current already on profile
// name, under the held operation lock. It is the scoped half of
// resyncCurrentScopes, split out so a reinstall activation — which switches
// the machine scope itself through useLocked — still converges the scoped
// currents without re-materializing the machine scope a second time.
func resyncScopedScopes(op *operation, home, name string, policy Policy) error {
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return err
	}
	for scope, current := range scoped {
		if current != name {
			continue
		}
		environment, target := splitScope(scope)
		if _, err := useLocked(op, home, name, environment, target, false, policy); err != nil {
			return err
		}
	}
	return nil
}
