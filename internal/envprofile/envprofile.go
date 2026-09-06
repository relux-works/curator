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
// Stage bounds: overlays come from machine configuration (manager-config
// schema 2), which is out of scope for this stage, so resolution runs with
// no overlays and Direct is empty; the §9.4 migration freezes the global
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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Diagnostics (environments §2.1, §9.7, manager profile §12.3).
const (
	DiagNameTaken       = "profile_name_taken"
	DiagRefConflict     = "profile_install_ref_conflict"
	DiagSourceInvalid   = "profile_source_invalid"
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

// Source is the install record of one profile.
type Source struct {
	Kind      string      `json:"kind"`
	Git       string      `json:"git,omitempty"`
	Path      string      `json:"path,omitempty"`
	Directory string      `json:"directory,omitempty"`
	Req       Requirement `json:"requirement"`
}

// ProfilesDir is the profile store below the manager home.
func ProfilesDir(home string) string { return filepath.Join(home, "profiles") }

// ProfileDir is one profile's directory.
func ProfileDir(home, name string) string { return filepath.Join(ProfilesDir(home), name) }

// CurrentFile records the machine current profile name.
func CurrentFile(home string) string { return filepath.Join(ProfilesDir(home), "current") }

// ScopedDir records per-scope current profiles.
func ScopedDir(home string) string { return filepath.Join(ProfilesDir(home), "scoped") }

func sourcePath(home, name string) string {
	return filepath.Join(ProfileDir(home, name), "source.json")
}
func lockPath(home, name string) string { return filepath.Join(ProfileDir(home, name), "lock.json") }

// Info is one installed or listed profile. Warnings carries the
// non-blocking audit findings of the install or update that produced it:
// context-system-module-present, mcp_command_unresolved, and unmatched
// waivers.
type Info struct {
	Name      string
	Source    Source
	Lock      *contextlock.Lock
	LockHash  string
	Current   bool
	ScopedFor []string
	Warnings  []string
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

// List returns every installed profile with its lock and currency.
func List(home string) ([]Info, error) {
	op, err := beginOperation(home)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()
	return listLocked(op, home)
}

// listLocked lists under the held operation lock.
func listLocked(op *operation, home string) ([]Info, error) {
	if err := ensureDefault(op, home); err != nil {
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
		out[entry.Name()] = strings.TrimSpace(string(payload))
	}
	return out, nil
}

// SetScoped records (or with clear=true, drops) a scope's current profile.
func SetScoped(home, scope, name string, clearScope bool) error {
	if clearScope {
		err := os.Remove(filepath.Join(ScopedDir(home), scope))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(ScopedDir(home), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ScopedDir(home), scope), []byte(name+"\n"), 0o644)
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

// InstallOptions selects the source of one installation.
type InstallOptions struct {
	Operand   string
	Directory string
	Range     string
	Tag       string
	Revision  string
	As        string
	Use       bool
}

// Install installs one root context package as a profile: it resolves the
// closure, audits every member always-strict, writes the lock, and installs
// every member's store entry. A git operand takes at most one requirement
// flag (default range latest); a path operand takes none and no directory.
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
	if err := ensureDefault(op, home); err != nil {
		return Info{}, false, false, err
	}
	isPath := isPathOperand(options.Operand)
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
			return Info{}, false, false, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		name = options.As
		if name == "" {
			name = manifest.Name
		}
		state, err := stateForPath(home, manifest.Name, options.Operand)
		if err != nil {
			return Info{}, false, false, err
		}
		source = Source{Kind: KindPath, Path: options.Operand}
		input = contextresolve.Input{
			Root:      contextresolve.Requirement{Kind: contextlock.KindContext, Name: manifest.Name},
			RootState: state,
		}
	} else {
		manager = newGitManager(home)
		requirement := Requirement{Range: options.Range, Tag: options.Tag, Revision: options.Revision}
		if requirement.Range == "" && requirement.Tag == "" && requirement.Revision == "" {
			requirement.Range = "latest"
		}
		source = Source{Kind: KindGit, Git: canonicalGit(options.Operand), Directory: options.Directory, Req: requirement}
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
		// re-resolves exactly as profile update does and is reported as an
		// update; a different source under the same name is taken.
		if prior == source {
			info, moved, err := updateLocked(op, home, name)
			if err != nil {
				return Info{}, false, false, err
			}
			_ = moved
			return info, false, true, nil
		}
		return Info{}, false, false, fmt.Errorf("%s: profile %q is already installed", DiagNameTaken, name)
	}
	if manager == nil {
		manager = newGitManager(home)
	}
	result, err := contextresolve.Resolve(manager, input)
	if err != nil {
		return Info{}, false, false, err
	}
	warnings, err := auditAndStore(home, manager, result)
	if err != nil {
		return Info{}, false, false, err
	}
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
	activated := false
	if machine == "" || options.Use {
		if err := op.publish(map[string][]byte{CurrentFile(home): []byte(name + "\n")}); err != nil {
			return Info{}, false, false, err
		}
		activated = true
	}
	info := Info{Name: name, Source: source, Lock: result.Lock, LockHash: hash, Current: activated || machine == name, Warnings: warnings}
	return info, activated, false, nil
}

// Update re-resolves the root (and overlays, when the machine declares any)
// from the declared requirement, fetching new candidates. A blocking
// finding on a member new to the lock leaves the old lock in place with
// profile_update_blocked. A root pinned by tag or revision is reported as
// pinned and does not move. A path root re-resolves against its directory.
func Update(home, name string) (Info, bool, error) {
	op, err := beginOperation(home)
	if err != nil {
		return Info{}, false, err
	}
	defer func() { _ = op.close() }()
	return updateLocked(op, home, name)
}

// updateLocked re-resolves under the held operation lock.
func updateLocked(op *operation, home, name string) (Info, bool, error) {
	if err := ensureDefault(op, home); err != nil {
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
	manager := newGitManager(home)
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
		manifest, err := contextpkg.LoadManifest(source.Path)
		if err != nil {
			return Info{}, false, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		state, err := stateForPath(home, manifest.Name, source.Path)
		if err != nil {
			return Info{}, false, err
		}
		input = contextresolve.Input{
			Root:      contextresolve.Requirement{Kind: contextlock.KindContext, Name: manifest.Name},
			RootState: state,
		}
	default:
		return Info{}, false, fmt.Errorf("%s: profile %q is the builtin local profile and does not move", DiagUpdateBlocked, name)
	}
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
	}
	if result.LockHash == oldHash {
		machine, _ := Current(home)
		return Info{Name: name, Source: source, Lock: result.Lock, LockHash: oldHash, Current: machine == name}, false, nil
	}
	warnings, err := auditAndStore(home, manager, result)
	if err != nil {
		return Info{}, false, err
	}
	canonical, err := result.Lock.Canonical()
	if err != nil {
		return Info{}, false, err
	}
	if err := op.publish(map[string][]byte{lockPath(home, name): canonical}); err != nil {
		return Info{}, false, err
	}
	hash := contextlock.HashBytes(canonical)
	if err := resyncCurrentScopes(op, home, name); err != nil {
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
	if purge {
		if err := purgeHomes(name); err != nil {
			return err
		}
	}
	return os.RemoveAll(ProfileDir(home, name))
}

// auditAndStore audits every resolved member always-strict and installs its
// store entry. A blocking finding fails the operation; system modules and
// unresolved MCP commands are reported as warnings.
func auditAndStore(home string, manager *gitManager, result *contextresolve.Result) ([]string, error) {
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

// stateForPath installs the path root into the store and returns its state
// package for resolution.
func stateForPath(home, name, dir string) (*contextresolve.StatePackage, error) {
	entry, hash, err := contextstore.EnsureState(home, contextlock.KindContext, name, dir)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	manifest, err := contextpkg.LoadManifest(entry)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	return &contextresolve.StatePackage{StateHash: hash, Manifest: packageOf(manifest)}, nil
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

// isPathOperand tells a path operand from a git URL syntactically, never by
// probing the network: an existing local directory is a path.
func isPathOperand(operand string) bool {
	info, err := os.Stat(operand)
	return err == nil && info.IsDir()
}

// canonicalGit normalizes a git operand onto its canonical source identity.
func canonicalGit(operand string) string {
	return strings.TrimSuffix(strings.TrimSpace(operand), "/")
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
func EnsureDefault(home string) error {
	op, err := beginOperation(home)
	if err != nil {
		return err
	}
	defer func() { _ = op.close() }()
	return ensureDefault(op, home)
}

func ensureDefault(op *operation, home string) error {
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
	skills, err := migrateGlobalSkills(home)
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
// and audited exactly like an installed member. A declaration the lock
// cannot represent — a branch pin, or a local skill with no source identity,
// since a skill member pins a commit with its source and a state pin admits
// only a context member — is left for the skill-pipeline stage that owns
// live direct declarations (see the stage bounds in the package doc).
func migrateGlobalSkills(home string) ([]contextlock.Member, error) {
	global, err := manifest.Load(filepath.Join(home, "global"))
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	if global == nil {
		return nil, nil
	}
	manager := newGitManager(home)
	var members []contextlock.Member
	for _, decl := range global.Skills {
		if decl.Git == "" || (decl.Ref.Kind != "tag" && decl.Ref.Kind != "revision") {
			continue
		}
		identity := canonicalGit(decl.Git)
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
		members = append(members, contextlock.Member{
			Kind: contextlock.KindSkill, Name: decl.Name, Source: identity, Commit: resolved.Commit,
		})
	}
	return members, nil
}

// resyncCurrentScopes re-materializes every scope whose current profile is
// name after its lock moved, under the held operation lock.
func resyncCurrentScopes(op *operation, home, name string) error {
	machine, err := Current(home)
	if err != nil {
		return err
	}
	if machine == name {
		if _, err := useLocked(op, home, name, "", "", false); err != nil {
			return err
		}
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return err
	}
	for scope, current := range scoped {
		if current != name {
			continue
		}
		environment, target := splitScope(scope)
		if _, err := useLocked(op, home, name, environment, target, false); err != nil {
			return err
		}
	}
	return nil
}
