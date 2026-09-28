// Package envprofile managed homes (environments §8.1, §7.4, §10.1): one manager-owned home
// per profile × environment below the environments root, provisioned on
// the first env resolve --repair naming it, verified lock-free on every
// resolve, and repaired under the manager-home mutation lock. Managed
// surfaces link into the immutable profile store with copies where a
// surface requires bytes (the always-copied claude_code root-context file,
// §8.1); every surface is recorded in the environment marker with its
// content hash, its form, and every copy's reason. Credential passthrough
// entries and provisioning seeds are recorded but never hashed; the tool
// owns them after provisioning.
package envprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// Diagnostics for managed homes and resolution (environments §7.7, §8.5,
// §9.7, §10.4).
const (
	DiagHomeStale       = "environment_home_stale"
	DiagRepairFailed    = "environment_repair_failed"
	DiagForeignManager  = "environment_foreign_manager_detected"
	DiagForeignSuspect  = "environment_foreign_manager_suspected"
	DiagProfileUnknown  = "profile_unknown"
	DiagReservedCommand = "environment_reserved_command_name"
)

// EnvRootName is the manager-owned environments root below the manager
// home (environments §8.1).
const EnvRootName = "environments"

// EnvRoot returns the manager-owned environments root.
func EnvRoot(home string) string { return filepath.Join(home, EnvRootName) }

// ManagedParent returns the fragment env-value directory for a profile ×
// environment: the tool's home, except for opencode where the variable
// names the managed XDG parent and the tool reads the opencode child
// (environments §7.1).
func ManagedParent(home, profile, envID string) string {
	return filepath.Join(EnvRoot(home), profile, envID)
}

// ManagedHomeDir returns the tool's home directory.
func ManagedHomeDir(home, profile, envID string) string {
	if envID == envregistry.OpenCode {
		return filepath.Join(ManagedParent(home, profile, envID), "opencode")
	}
	return ManagedParent(home, profile, envID)
}

// ResolveRequest is one env resolve or repair operation.
type ResolveRequest struct {
	Home      string
	Profile   string
	EnvID     string
	LaunchDir string
	Machine   envregistry.MachineConfig
	Repair    bool
	Format    string
	// Policy carries the machine gates (environments §12.1): the
	// precedence primitives that drive emission order and the composition
	// policy. The zero value resolves with the pair default.
	Policy Policy
	// Detect maps an adapter to its detected tool release, or "unknown".
	// Nil means probe the machine read-only.
	Detect func(adapter envregistry.Adapter) string
	// NativeHomeOf resolves native homes; nil means the process homes.
	// A test seam: production never sets it.
	NativeHomeOf func(id string) (string, error)
	// OperatorXDG is the operator's effective XDG config home; empty
	// means resolve it from the process environment.
	OperatorXDG string
	// codexSeedRevisionForTest exercises a retained, non-shipped seed
	// revision through Resolve without mutating the registry singleton.
	codexSeedRevisionForTest string
	// transactionOptions is a fault-injection seam for recovery tests. CLI
	// requests leave it nil and use the production transaction engine.
	transactionOptions []transaction.Option
	// passthroughLstat and passthroughReadlink inject entry inspection errors
	// in tests. Production requests leave them nil and use the shared reader.
	passthroughLstat    func(string) (os.FileInfo, error)
	passthroughReadlink func(string) (string, error)
	readStateFile       func(string) (stateread.File, error)
	readRegularFile     func(string) (stateread.File, error)
}

// ResolveResult is the outcome: exactly one of Document and StaleErr is
// set; a first provisioning additionally carries the first-resolve notice.
type ResolveResult struct {
	Document     []byte
	Notice       string
	Warnings     []string
	Provisioned  bool
	StaleReasons []string
}

// versionToken finds the first dotted-decimal token in tool output.
var versionToken = regexp.MustCompile(`\d+(?:\.\d+)+`)

// detectRelease probes the adapter's tool read-only for its release
// (environments §7.9). Any failure reports "unknown", never a match.
func detectRelease(probe []string) string {
	if len(probe) == 0 {
		return "unknown"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, probe[0], probe[1:]...).Output() // #nosec G204 -- probe argv is registry data
	if err != nil {
		return "unknown"
	}
	if token := versionToken.FindString(string(out)); token != "" {
		return token
	}
	return "unknown"
}

// nativeHome resolves the adapter's native home through the seam.
func (req *ResolveRequest) nativeHome(id string) (string, error) {
	if req.NativeHomeOf != nil {
		return req.NativeHomeOf(id)
	}
	adapter, ok := adapterByID(id)
	if !ok {
		return "", fmt.Errorf("%s: unregistered environment %q", envregistry.DiagUnknown, id)
	}
	return NativeHome(adapter)
}

// detected resolves the tool release through the seam.
func (req *ResolveRequest) detected(adapter envregistry.Adapter) string {
	if req.Detect != nil {
		return req.Detect(adapter)
	}
	return detectRelease(adapter.Probe)
}

func applyCodexSeedRevisionTestOverride(adapter envregistry.Adapter, revision string) (envregistry.Adapter, error) {
	if revision == "" {
		return adapter, nil
	}
	if adapter.ID != envregistry.CodexCLI || (revision != envregistry.CodexSeedRevisionA && revision != envregistry.CodexSeedRevisionB) {
		return envregistry.Adapter{}, fmt.Errorf("invalid Codex seed revision test override %q for %s", revision, adapter.ID)
	}
	adapter.CodexSeedRevision = revision
	return adapter, nil
}

// operatorXDG resolves the operator's effective XDG config home.
func (req *ResolveRequest) operatorXDG() string {
	if req.OperatorXDG != "" {
		return req.OperatorXDG
	}
	if value := os.Getenv("XDG_CONFIG_HOME"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}

// mcpsOf loads the lock's MCP members from the store as resolved servers.
func mcpsOf(home string, manager *gitManager, lock *contextlock.Lock) ([]contextmaterialize.MCPServer, error) {
	var servers []contextmaterialize.MCPServer
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindMCP {
			continue
		}
		entry := manager.entryPath(home, resolvedOf(member))
		manifest, err := contextpkg.LoadMCP(packageRoot(entry, member.Directory))
		if err != nil {
			return nil, fmt.Errorf("%s: member %s %v", DiagRepairFailed, member.Name, err)
		}
		servers = append(servers, contextmaterialize.MCPServer{
			Name:         member.Name,
			Transport:    manifest.Server.Transport,
			Command:      manifest.Server.Command,
			Args:         manifest.Server.Args,
			URL:          manifest.Server.URL,
			EnvNames:     manifest.Server.EnvNames,
			Environments: manifest.Server.Environments,
		})
	}
	return servers, nil
}

// skillOf is one lock skill member with its store package root and content
// hash.
type skillOf struct {
	name string
	root string
	hash string
}

// skillsOf loads the lock's skill members: entry package roots with their
// store content hashes. A skill whose materialized name would collide with
// the umbrella discovery namespace is refused with
// environment_reserved_command_name (environments §9.4): profile-
// materialized files must not be able to poison the PATH the §11 dispatch
// trusts.
func skillsOf(home string, manager *gitManager, lock *contextlock.Lock) ([]skillOf, error) {
	var skills []skillOf
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindSkill {
			continue
		}
		if strings.HasPrefix(member.Name, "curator-") {
			return nil, fmt.Errorf("%s: skill %q collides with the curator-* name reservation", DiagReservedCommand, member.Name)
		}
		entry := manager.entryPath(home, resolvedOf(member))
		root := packageRoot(entry, member.Directory)
		hash, err := contextstore.ContentHash(root)
		if err != nil {
			return nil, fmt.Errorf("%s: member %s %v", DiagRepairFailed, member.Name, err)
		}
		skills = append(skills, skillOf{name: member.Name, root: root, hash: hash})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].name < skills[j].name })
	return skills, nil
}

// storeDocPath names the deterministic store document for manager-authored
// bytes: the same inputs always name the same path, so lock-free
// verification recomputes the expected link target without writing.
func storeDocPath(home, profile, envID, rel string) string {
	return filepath.Join(ProfilesDir(home), profile, "rendered", envID, filepath.FromSlash(rel))
}

// homePlan is the fully assembled managed home: every write, link, and
// removal, with the marker recording them.
type homePlan struct {
	adapter    envregistry.Adapter
	parent     string
	homeDir    string
	storeRoot  string
	form       string
	isolation  string
	copies     map[string][]byte
	links      map[string]string
	docs       map[string][]byte
	marker     *envmarker.Marker
	fileHashes map[string][]byte
	warnings   []string
}

// assembleHome builds the desired managed home purely from the lock, the
// store entries it names, and machine configuration: no read of the home
// itself beyond the fallback and merge inputs the caller supplies.
func assembleHome(req *ResolveRequest, source Source, lock *contextlock.Lock, hash string, precedence contextmaterialize.Precedence, order []contextlock.Member, adapter envregistry.Adapter, prior *envmarker.Marker) (*homePlan, error) {
	if err := validateProfilePathSources(req.Profile, source, req.Policy); err != nil {
		return nil, err
	}
	manager := newGitManager(req.Home)
	packages, err := loadMaterial(req.Home, manager, lock)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	servers, err := mcpsOf(req.Home, manager, lock)
	if err != nil {
		return nil, err
	}
	skills, err := skillsOf(req.Home, manager, lock)
	if err != nil {
		return nil, err
	}
	form, err := req.Machine.EffectiveForm(adapter)
	if err != nil {
		return nil, err
	}
	// An unknown release warns (environment_tool_version_unverified, emitted
	// by verification) rather than blocks: the selection scheme stands on
	// the bundle evidence, so resolution proceeds as at-or-above.
	atAbove := true
	if ok, known := adapter.AtOrAbovePinned(req.detected(adapter)); known {
		atAbove = ok
	}
	isolation, err := req.Machine.EffectiveIsolation(req.Profile, adapter, atAbove)
	if err != nil {
		return nil, err
	}
	if err := req.checkIsolatedCredentialStore(adapter, isolation); err != nil {
		return nil, err
	}
	plan := &homePlan{
		adapter:    adapter,
		parent:     ManagedParent(req.Home, req.Profile, adapter.ID),
		homeDir:    ManagedHomeDir(req.Home, req.Profile, adapter.ID),
		storeRoot:  ProfilesDir(req.Home),
		form:       form,
		isolation:  isolation,
		copies:     map[string][]byte{},
		links:      map[string]string{},
		docs:       map[string][]byte{},
		fileHashes: map[string][]byte{},
	}
	if form == envregistry.FormReferenced && adapter.ID == envregistry.OpenCode {
		if blocked, err := referencedBlocked(plan.homeDir, prior); err != nil {
			return nil, err
		} else if blocked {
			plan.warnings = append(plan.warnings, contextmaterialize.DiagFormUnavailable+": opencode.json exists and no marker records it; monolithic emitted")
			form = envregistry.FormMonolithic
			plan.form = form
		}
	}
	surfaces := map[string]envmarker.Surface{}
	if err := plan.rootContext(req, lock, hash, precedence, packages, surfaces); err != nil {
		return nil, err
	}
	if err := plan.systemPrompt(req, lock, precedence, packages, surfaces); err != nil {
		return nil, err
	}
	plan.skills(req, skills, surfaces)
	if err := plan.mcp(req, servers, surfaces); err != nil {
		return nil, err
	}
	marker := &envmarker.Marker{
		Version: 1,
		Profile: envmarker.Profile{
			Name: req.Profile, Root: lock.Root, Kind: markerKind(source),
			LockSHA256: strings.TrimPrefix(hash, "sha256:"),
			Source:     markerSource(source), Requirement: markerRequirement(source),
			Directory: source.Directory, SourcePath: markerSourcePath(source),
			ImportedFromNative: source.ImportedFromNative,
		},
		Precedence: envmarker.Precedence{Winner: precedence.Winner, Placement: precedence.Placement},
		Mode:       envmarker.ModeManagedHome,
		Surfaces:   surfaces,
	}
	for _, member := range order {
		entry := envmarker.Member{Name: member.Name, Version: member.Version, Weight: member.Weight, Overlay: member.Overlay}
		if member.Commit != "" {
			entry.Commit = member.Commit
		} else {
			entry.StateSHA256 = member.StateHash
		}
		marker.Members = append(marker.Members, entry)
	}
	plan.marker = marker
	return plan, nil
}

// referencedBlocked reports whether an unmanaged opencode.json blocks the
// referenced form: the file exists and the preceding marker does not
// record it, so the adapter warns environment_form_unavailable and
// materializes monolithic instead of editing the unmanaged file (§5.3).
func referencedBlocked(homeDir string, prior *envmarker.Marker) (bool, error) {
	recorded := false
	if prior != nil {
		for _, surface := range prior.Surfaces {
			for _, path := range surface.Paths {
				if path == contextmaterialize.OpenCodeConfigName || path == "opencode.json" {
					recorded = true
				}
			}
		}
	}
	if recorded {
		return false, nil
	}
	path := filepath.Join(homeDir, "opencode.json")
	metadata, err := stateread.Lstat(path)
	if err != nil {
		return false, err
	}
	if metadata.Kind == stateread.KindAbsent {
		return false, nil
	}
	if metadata.Kind != stateread.KindPresent {
		return false, stateread.UnusableError(path, fmt.Errorf("path metadata is unreadable"))
	}
	return true, nil
}

// rootContext assembles the root-context surface: monolithic bytes or the
// referenced file set, linked into the store with copies where the surface
// requires bytes. The claude_code root file is always a copied regular
// file (§8.1); referenced module files link into the immutable store
// entries so link-target identity verifies them lock-free (§10.1).
func (p *homePlan) rootContext(req *ResolveRequest, lock *contextlock.Lock, hash string, precedence contextmaterialize.Precedence, packages map[string]contextmaterialize.Package, surfaces map[string]envmarker.Surface) error {
	target := p.adapter.RootTarget
	if p.form == envregistry.FormReferenced {
		files, written, err := contextmaterialize.Referenced(lock, hash, precedence, p.adapter.ID, packages)
		if err != nil {
			return err
		}
		if !written {
			return nil
		}
		paths := make([]string, 0, len(files))
		for path := range files {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		copies := []envmarker.Copy{}
		for _, path := range paths {
			document := files[path]
			p.fileHashes[path] = document
			if path == target && p.adapter.ID == envregistry.ClaudeCode {
				p.copies[path] = document
				copies = append(copies, envmarker.Copy{Path: path, Reason: envmarker.ReasonClaudeCodeRootContext})
				continue
			}
			if path == target || path == contextmaterialize.OpenCodeConfigName {
				storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, path)
				p.docs[storeDoc] = document
				p.links[path] = storeDoc
				continue
			}
			member, module := splitModulePath(path)
			link, err := storeModuleLink(req, lock, member, module)
			if err != nil {
				return err
			}
			p.links[path] = link
		}
		ordered := append([]string{}, paths...)
		surfaces[envmarker.SurfaceRootContext] = envmarker.Surface{
			Paths:         ordered,
			Form:          p.form,
			ContentSHA256: contextmaterialize.SurfaceHash(p.fileHashesFor(paths)),
			Copies:        &copies,
		}
		return nil
	}
	document, written, err := contextmaterialize.Monolithic(lock, hash, precedence, p.adapter.ID, packages)
	if err != nil {
		return err
	}
	if !written {
		return nil
	}
	p.fileHashes[target] = document
	copies := []envmarker.Copy{}
	if p.adapter.ID == envregistry.ClaudeCode {
		p.copies[target] = document
		copies = append(copies, envmarker.Copy{Path: target, Reason: envmarker.ReasonClaudeCodeRootContext})
	} else {
		storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, target)
		p.docs[storeDoc] = document
		p.links[target] = storeDoc
	}
	// The size advisory warns but changes nothing about the bytes (§5).
	if int64(len(document)) > p.adapter.SizeAdvisoryBytes {
		p.warnings = append(p.warnings, fmt.Sprintf("%s: root context %d bytes exceeds the %s advisory of %d", envregistry.DiagSizeExceeded, len(document), p.adapter.ID, p.adapter.SizeAdvisoryBytes))
	}
	surfaces[envmarker.SurfaceRootContext] = envmarker.Surface{
		Paths:         []string{target},
		Form:          p.form,
		ContentSHA256: contextmaterialize.SurfaceHash(map[string][]byte{target: document}),
		Copies:        &copies,
	}
	return nil
}

// fileHashesFor projects the plan's known bytes onto paths.
func (p *homePlan) fileHashesFor(paths []string) map[string][]byte {
	out := map[string][]byte{}
	for _, path := range paths {
		out[path] = p.fileHashes[path]
	}
	return out
}

// splitModulePath cuts a module file path into its package and module
// segments.
func splitModulePath(path string) (string, string) {
	rest, _ := strings.CutPrefix(path, contextmaterialize.ModulesDir+"/")
	name, module, _ := strings.Cut(rest, "/")
	return name, module
}

// storeModuleLink resolves the immutable store file a referenced module
// links into.
func storeModuleLink(req *ResolveRequest, lock *contextlock.Lock, member, module string) (string, error) {
	manager := newGitManager(req.Home)
	found, ok := lock.Find(contextlock.KindContext, member)
	if !ok {
		return "", fmt.Errorf("%s: lock carries no context member %s", DiagRepairFailed, member)
	}
	entry := manager.entryPath(req.Home, resolvedOf(found))
	return filepath.Join(packageRoot(entry, found.Directory), contextpkg.ContextDir, filepath.FromSlash(module)), nil
}

// publishDocs publishes the plan's manager-authored store documents. It
// runs only under the mutation lock: assembly itself is pure so that
// lock-free verification recomputes expected link targets without
// writing (environments §10.1, §12).
func (p *homePlan) publishDocs() error {
	paths := make([]string, 0, len(p.docs))
	for path := range p.docs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := writeStoreDoc(p.storeRoot, path, p.docs[path]); err != nil {
			return err
		}
	}
	return nil
}

// effectivePassthrough computes the passthrough links a managed home
// records: home-relative path to native absolute target, with the strategy
// (§7.4). isolated homes link nothing; keyring-backed codex homes are
// ambient and link nothing; every file-shaped strategy is watched by the
// liveness row. The native config.toml read is read-only: an absent file
// or an absent key means the default file store, while an unreadable file
// fails instead of defaulting — absence and a failed read are different
// facts (§8.4.1) — and a selector outside the verified file/keyring/auto
// set fails closed with environment_credential_unsupported.
func (req *ResolveRequest) effectivePassthrough(adapter envregistry.Adapter, isolation string) (map[string]string, map[string]string, error) {
	links := map[string]string{}
	strategies := map[string]string{}
	if isolation == envregistry.IsolationIsolated {
		return links, strategies, nil
	}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range adapter.PassthroughFor(runtime.GOOS) {
		switch entry.Strategy {
		case envregistry.StrategyPerHomeKeychain, envregistry.StrategyAmbient:
			continue
		case envregistry.StrategyKeyringPreferred:
			store, err := req.codexCredentialStore(native)
			if err != nil {
				return nil, nil, err
			}
			if store == "keyring" {
				continue
			}
			links[entry.Path] = filepath.Join(native, filepath.FromSlash(entry.FileLinkTarget))
			strategies[entry.Path] = envregistry.StrategyFileLink
		case envregistry.StrategyFileLink:
			links[entry.Path] = filepath.Join(native, filepath.FromSlash(entry.FileLinkTarget))
			strategies[entry.Path] = envregistry.StrategyFileLink
		}
	}
	return links, strategies, nil
}

// credentialRecords computes the schema-v2 record for each declared
// credential strategy. It records links plus supported linkless strategies;
// the latter deliberately have no Path and never enter link inventory.
func (req *ResolveRequest) credentialRecords(adapter envregistry.Adapter, isolation, provenance string) ([]envmarker.Passthrough, error) {
	if isolation == envregistry.IsolationIsolated {
		if adapter.ID == envregistry.ClaudeCode && runtime.GOOS == "darwin" {
			return []envmarker.Passthrough{{
				Isolation: isolation, Strategy: envregistry.StrategyPerHomeKeychain,
				SourceRole: "managed", Backend: "keychain", BackendVersion: adapter.VerifiedRelease,
				Provenance: provenance,
			}}, nil
		}
		return []envmarker.Passthrough{}, nil
	}
	if isolation != envregistry.IsolationShared {
		return nil, fmt.Errorf("unsupported credential isolation %q", isolation)
	}
	entries := adapter.PassthroughFor(runtime.GOOS)
	if len(entries) == 0 {
		return []envmarker.Passthrough{}, nil
	}
	if adapter.VerifiedRelease == "" {
		return nil, fmt.Errorf("%s: %s has no verified credential backend release", envregistry.DiagCredentialUnsupported, adapter.ID)
	}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return nil, err
	}
	links, _, err := req.effectivePassthrough(adapter, isolation)
	if err != nil {
		return nil, err
	}
	var codexStore string
	for _, entry := range entries {
		if entry.Strategy == envregistry.StrategyKeyringPreferred {
			codexStore, err = req.codexCredentialStore(native)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	records := make([]envmarker.Passthrough, 0, len(entries))
	for _, entry := range entries {
		record := envmarker.Passthrough{
			Isolation: isolation, Strategy: entry.Strategy, SourceRole: "native",
			Backend: "file", BackendVersion: adapter.VerifiedRelease, Provenance: provenance,
		}
		switch entry.Strategy {
		case envregistry.StrategyFileLink:
			record.Path = entry.Path
		case envregistry.StrategyKeyringPreferred:
			if codexStore == "keyring" {
				record.Backend = "ambient"
			} else {
				record.Path = entry.Path
			}
		case envregistry.StrategyAmbient:
			record.Backend = "ambient"
		case envregistry.StrategyPerHomeKeychain:
			record.SourceRole = "managed"
			record.Backend = "keychain"
		default:
			return nil, fmt.Errorf("%s: %s declares unsupported credential strategy %q", envregistry.DiagCredentialUnsupported, adapter.ID, entry.Strategy)
		}
		if record.Path != "" {
			if _, linked := links[record.Path]; !linked {
				return nil, fmt.Errorf("%s: %s credential path %s has no effective link", envregistry.DiagCredentialUnsupported, adapter.ID, record.Path)
			}
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Path != records[j].Path {
			return records[i].Path < records[j].Path
		}
		return records[i].Strategy < records[j].Strategy
	})
	return records, nil
}

// codexCredentialStore resolves the operator's native codex credential
// store: the top-level cli_auth_credentials_store selector when the
// native config.toml carries one, else the platform default file — an
// absent file and an absent key both mean file (§7.4). The file is
// parsed as TOML, so every valid spelling of the key — double-quoted,
// single-quoted, multi-line — resolves identically, while the same key
// nested inside a table is not the top-level selector and stays absent.
// Under keyring the credential is ambient and no entry is linked; under
// file or auto the managed auth.json is a file-link. An unreadable file
// or invalid TOML refuses with a diagnostic naming the file, never
// absence, and a parsed value outside the verified file/keyring/auto
// set — any other string or any non-string TOML value — fails closed
// with environment_credential_unsupported. Sharing is defined by the
// native effective storage only: a diverged managed config.toml changes
// nothing and is never realigned.
func (req *ResolveRequest) codexCredentialStore(native string) (string, error) {
	readRegularFile := req.readRegularFile
	if readRegularFile == nil {
		readRegularFile = stateread.ReadRegularFile
	}
	return codexCredentialStoreWith(native, readRegularFile)
}

func codexCredentialStoreWith(native string, readRegularFile func(string) (stateread.File, error)) (string, error) {
	path := filepath.Join(native, "config.toml")
	state, err := readRegularFile(path)
	if err != nil {
		return "", fmt.Errorf("%s: codex native config.toml: %w", envregistry.DiagSeedUnreadable, err)
	}
	if state.Kind == stateread.KindAbsent {
		return "file", nil
	}
	if state.Kind != stateread.KindPresent {
		return "", stateread.UnusableError(path, fmt.Errorf("unknown file state %q", state.Kind))
	}
	payload := state.Bytes
	var doc map[string]any
	if err := toml.Unmarshal(payload, &doc); err != nil {
		return "", fmt.Errorf("%s: %w", envregistry.DiagSeedUnreadable, stateread.UnusableError(path, fmt.Errorf("codex native config.toml is not valid TOML: %w", err)))
	}
	raw, ok := doc["cli_auth_credentials_store"]
	if !ok {
		return "file", nil
	}
	store, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%s: codex native cli_auth_credentials_store must be one of \"file\", \"keyring\", or \"auto\", got TOML type %T", envregistry.DiagCredentialUnsupported, raw)
	}
	switch store {
	case "file", "keyring", "auto":
		return store, nil
	default:
		return "", fmt.Errorf("%s: codex native cli_auth_credentials_store %q is outside the verified file, keyring, and auto set", envregistry.DiagCredentialUnsupported, store)
	}
}

// checkIsolatedCredentialStore gates codex_cli isolated on the native
// effective store (§7.4): isolated is admitted under file storage only —
// including an absent cli_auth_credentials_store key, which resolves to
// effective file. Under keyring or auto storage the keyring identity is
// assumed operator-global, so isolated is
// environment_isolated_unsupported; a selector outside the verified set
// fails closed with environment_credential_unsupported from the store
// reader. Every other adapter × mode passes here: the static matrix in
// the registry already decided it.
func (req *ResolveRequest) checkIsolatedCredentialStore(adapter envregistry.Adapter, isolation string) error {
	if adapter.ID != envregistry.CodexCLI || isolation != envregistry.IsolationIsolated {
		return nil
	}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return err
	}
	store, err := req.codexCredentialStore(native)
	if err != nil {
		return err
	}
	switch store {
	case "file":
		return nil
	case "keyring":
		return fmt.Errorf("%s: isolated is unsupported for codex_cli under native keyring storage: the keyring identity is operator-global, independent of CODEX_HOME", envregistry.DiagIsolatedUnsupported)
	case "auto":
		return fmt.Errorf("%s: isolated is unsupported for codex_cli under native auto storage: the tool uses the keyring when one is available, so an isolated home on a keyring host would authenticate through the operator-global keyring", envregistry.DiagIsolatedUnsupported)
	default:
		return fmt.Errorf("%s: codex native cli_auth_credentials_store %q is outside the verified file, keyring, and auto set", envregistry.DiagCredentialUnsupported, store)
	}
}

// seedBundle is the one-time provisioning class (§7.4): non-credential
// files copied from the native home exactly once, at provisioning, never
// refreshed, never hashed.
type seedBundle struct {
	files map[string][]byte
	// xdg maps parent-relative names to operator absolute targets.
	xdg map[string]string
	// claudeInit marks the written (not copied) .claude.json seed.
	claudeInit bool
	// codexSeedRecord is captured only when a codex_cli home is first
	// provisioned. Repair preserves the marker's existing record and seed.
	codexSeedRecord *envmarker.CodexSeedRecord
	warnings        []string
}

func parseCodexSeedConfig(payload []byte) (map[string]any, []string, bool, error) {
	var document map[string]any
	if err := toml.Unmarshal(payload, &document); err != nil {
		return nil, nil, false, fmt.Errorf("config.toml is not valid TOML: %w", err)
	}
	serversValue, exists := document["mcp_servers"]
	if !exists {
		return document, []string{}, false, nil
	}
	servers, ok := serversValue.(map[string]any)
	if !ok {
		return nil, nil, false, fmt.Errorf("config.toml mcp_servers is not a TOML table")
	}
	names := make([]string, 0, len(servers))
	for name := range servers {
		if name == "" {
			return nil, nil, false, fmt.Errorf("config.toml mcp_servers contains an empty server name")
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return document, names, true, nil
}

// stripCodexSeedMCPServers applies the rc.13 §7.4 revision-B rule: remove
// the whole top-level mcp_servers table (including every subtable), retain
// every other TOML member, and return only the sorted server names.
func stripCodexSeedMCPServers(payload []byte) ([]byte, []string, error) {
	document, names, exists, err := parseCodexSeedConfig(payload)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return payload, names, nil
	}
	delete(document, "mcp_servers")
	var stripped bytes.Buffer
	if err := toml.NewEncoder(&stripped).Encode(document); err != nil {
		return nil, nil, fmt.Errorf("strip config.toml mcp_servers: %w", err)
	}
	return stripped.Bytes(), names, nil
}

// gatherSeeds reads every seed upfront so that an unreadable seed stops
// provisioning before the first write. A seed absent in the native home
// is simply not seeded; absence and unreadability stay different facts
// (§8.4.1). Copy seeds are gathered at provisioning only: repair never
// refreshes them, so an unreadable native copy must not fail a repair
// that would never read it.
func (req *ResolveRequest) gatherSeeds(adapter envregistry.Adapter, provision bool) (*seedBundle, error) {
	bundle := &seedBundle{files: map[string][]byte{}, xdg: map[string]string{}}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return nil, err
	}
	if provision {
		if adapter.ID == envregistry.CodexCLI {
			if adapter.CodexSeedRevision != envregistry.CodexSeedRevisionA && adapter.CodexSeedRevision != envregistry.CodexSeedRevisionB {
				return nil, fmt.Errorf("%s: unsupported codex_cli seed revision %q", envregistry.DiagSeedUnreadable, adapter.CodexSeedRevision)
			}
			bundle.codexSeedRecord = &envmarker.CodexSeedRecord{
				Revision:         adapter.CodexSeedRevision,
				NativeMCPServers: []string{},
			}
		}
		for _, seed := range adapter.Seeds {
			if adapter.SeedWritten[seed] {
				bundle.claudeInit = true
				continue
			}
			path := filepath.Join(native, filepath.FromSlash(seed))
			readRegularFile := req.readRegularFile
			if readRegularFile == nil {
				readRegularFile = stateread.ReadRegularFile
			}
			file, err := readRegularFile(path) // #nosec G304 -- seed names are registry data
			if err != nil {
				return nil, fmt.Errorf("%s: seed %s: %v", envregistry.DiagSeedUnreadable, seed, err)
			}
			switch file.Kind {
			case stateread.KindAbsent:
				continue
			case stateread.KindPresent:
			default:
				return nil, fmt.Errorf("%s: seed %s has unusable read state %q", envregistry.DiagSeedUnreadable, seed, file.Kind)
			}
			payload := file.Bytes
			if adapter.ID == envregistry.CodexCLI && seed == "config.toml" {
				var names []string
				switch adapter.CodexSeedRevision {
				case envregistry.CodexSeedRevisionA:
					_, names, _, err = parseCodexSeedConfig(payload)
				case envregistry.CodexSeedRevisionB:
					payload, names, err = stripCodexSeedMCPServers(payload)
				}
				if err != nil {
					return nil, fmt.Errorf("%s: seed %s: %v", envregistry.DiagSeedUnreadable, seed, err)
				}
				bundle.codexSeedRecord.NativeMCPServers = names
				if len(names) > 0 {
					switch adapter.CodexSeedRevision {
					case envregistry.CodexSeedRevisionA:
						bundle.warnings = append(bundle.warnings, fmt.Sprintf("%s: native Codex MCP servers %s were inherited into this managed home outside the profile lock and MCP allowlist; the next seed revision stops inheriting them. Declare each server in the profile's MCP set, or accept the loss.", envregistry.DiagMCPNativeServersUngoverned, strings.Join(names, ", ")))
					case envregistry.CodexSeedRevisionB:
						bundle.warnings = append(bundle.warnings, fmt.Sprintf("%s: native Codex MCP servers %s were stripped from config.toml and are not inherited", envregistry.DiagMCPNativeServersNotInherited, strings.Join(names, ", ")))
					}
				}
			}
			bundle.files[seed] = payload
		}
	} else if adapter.ID == envregistry.ClaudeCode {
		bundle.claudeInit = true
	}
	if adapter.ID == envregistry.OpenCode {
		xdg := req.operatorXDG()
		if xdg == "" {
			return bundle, nil
		}
		for _, name := range req.Machine.XDGSeedAllowlist {
			if name == "opencode" {
				continue
			}
			target := filepath.Join(xdg, name)
			metadata, err := stateread.Stat(target)
			if err != nil {
				return nil, fmt.Errorf("%s: xdg seed %s: %v", envregistry.DiagSeedUnreadable, name, err)
			}
			if metadata.Kind == stateread.KindAbsent {
				continue
			}
			bundle.xdg[name] = target
		}
	}
	return bundle, nil
}

// writeStoreDoc publishes manager-authored bytes as a store document.
func writeStoreDoc(root, path string, document []byte) error {
	rel, err := managedRelative(root, path)
	if err != nil {
		return err
	}
	return atomicManagedFile(root, rel, document, 0o644)
}

// claudeSeed merges the .claude.json provisioning seed: the exact object
// {"hasCompletedOnboarding":true,"projects":{}} at provisioning, plus one
// project entry per launch directory holding hasTrustDialogAccepted and —
// under the referenced form — hasClaudeMdExternalIncludesApproved (§7.4).
// Later tool writes are its own state: repair only adds the launch
// directory's entry and never rewrites anything else.
func claudeSeed(root, homeRel, launchDir, form string) (seeded []string, err error) {
	path := managedJoin(homeRel, ".claude.json")
	object := map[string]any{"hasCompletedOnboarding": true, "projects": map[string]any{}}
	payload, present, err := readManagedRegular(root, path)
	if err != nil {
		if strings.Contains(err.Error(), DiagWriteWouldFollowLink) {
			return nil, err
		}
		return nil, fmt.Errorf("%s: managed .claude.json: %v", envregistry.DiagSeedUnreadable, err)
	}
	if present {
		parsed, ok := decodeJSONObject(payload)
		if !ok {
			return nil, fmt.Errorf("%s: managed .claude.json is not an object", envregistry.DiagSeedUnreadable)
		}
		object = parsed
		if _, ok := object["projects"].(map[string]any); !ok {
			object["projects"] = map[string]any{}
		}
	} else if full, pathErr := managedPath(root, path, false); pathErr != nil {
		return nil, fmt.Errorf("%s: managed .claude.json: %v", envregistry.DiagSeedUnreadable, pathErr)
	} else if state, statErr := stateread.Lstat(full); statErr != nil {
		return nil, fmt.Errorf("%s: managed .claude.json: %v", envregistry.DiagSeedUnreadable, statErr)
	} else if state.Kind == stateread.KindPresent && !state.Info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s: managed .claude.json is not a regular file", envregistry.DiagSeedUnreadable)
	}
	projects := object["projects"].(map[string]any)
	for name := range projects {
		seeded = append(seeded, name)
	}
	entry, _ := projects[launchDir].(map[string]any)
	if entry == nil {
		entry = map[string]any{}
	}
	entry["hasTrustDialogAccepted"] = true
	if form == envregistry.FormReferenced {
		entry["hasClaudeMdExternalIncludesApproved"] = true
	}
	projects[launchDir] = entry
	found := false
	for _, name := range seeded {
		if name == launchDir {
			found = true
		}
	}
	if !found {
		seeded = append(seeded, launchDir)
	}
	sort.Strings(seeded)
	payload, err = protocoljsonMarshal(object)
	if err != nil {
		return nil, err
	}
	if err := atomicManagedFile(root, path, payload, 0o644); err != nil {
		return nil, err
	}
	return seeded, nil
}

// decodeJSONObject decodes one JSON object value.
func decodeJSONObject(payload []byte) (map[string]any, bool) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, false
	}
	return object, true
}

// protocoljsonMarshal renders manager-authored JSON state as CCJ-1 plus
// one trailing LF.
func protocoljsonMarshal(object map[string]any) ([]byte, error) {
	document, err := protocoljson.MarshalCanonical(object)
	if err != nil {
		return nil, err
	}
	return append(document, '\n'), nil
}

// claudeProjects reads the managed .claude.json project entries for
// verification. It reports the external-includes approval flag per
// launch directory (§5.3): true only when the entry carries
// hasClaudeMdExternalIncludesApproved. A missing file means no entry;
// an unreadable or invalid file is reported, never treated as absence
// (§8.4).
func claudeProjects(homeDir string) (map[string]bool, error) {
	path := filepath.Join(homeDir, ".claude.json")
	state, err := stateread.ReadRegularFile(path) // #nosec G304 -- managed .claude.json below the resolved home
	if err != nil {
		return nil, fmt.Errorf("%s: managed .claude.json: %w", envregistry.DiagSeedUnreadable, err)
	}
	switch state.Kind {
	case stateread.KindAbsent:
		return map[string]bool{}, nil
	case stateread.KindUnreadable:
		return nil, fmt.Errorf("%s: managed .claude.json is unreadable", envregistry.DiagSeedUnreadable)
	case stateread.KindPresent:
	default:
		return nil, fmt.Errorf("%s: managed .claude.json has unknown read state %q", envregistry.DiagSeedUnreadable, state.Kind)
	}
	parsed, ok := decodeJSONObject(state.Bytes)
	if !ok {
		return nil, fmt.Errorf("%s: managed .claude.json is not an object", envregistry.DiagSeedUnreadable)
	}
	entries := map[string]bool{}
	if projects, ok := parsed["projects"].(map[string]any); ok {
		for name, entry := range projects {
			approved := false
			if object, ok := entry.(map[string]any); ok {
				approved, _ = object["hasClaudeMdExternalIncludesApproved"].(bool)
			}
			entries[name] = approved
		}
	}
	return entries, nil
}

// dotfileStateDirs is the closed heuristic list for the §9.5 inventory: a
// well-known dotfile-manager state location elevates the notice for plain
// unmanaged files to environment_foreign_manager_suspected. The heuristic
// never blocks.
//
// Stated bound (TASK-260906-vlrjo1): the entries are POSIX-portable
// relative paths resolved under os.UserHomeDir, matching the §9.5
// spellings verbatim. Whether the closed list is platform-specific —
// chezmoi keeps state under %LOCALAPPDATA% on Windows, home-manager is
// Nix-only — is a spec question §9.5 does not decide, so no Windows
// location is invented here and the heuristic stays inert for those
// layouts on that platform.
var dotfileStateDirs = [][2]string{
	{".local/share/chezmoi", "chezmoi"},
	{".config/home-manager", "home-manager"},
}

// foreignManagerHint scans the operator home for dotfile-manager state.
func foreignManagerHint() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, candidate := range dotfileStateDirs {
		if info, err := os.Stat(filepath.Join(home, filepath.FromSlash(candidate[0]))); err == nil && info.IsDir() {
			return candidate[1]
		}
	}
	return ""
}

// inventoryUnmanaged walks the desired home paths before the first write:
// a managed-surface path that is already a symlink pointing outside the
// manager's store is evidence of another manager and stops the operation
// with environment_foreign_manager_detected; any other unmanaged file
// fails with environment_surface_unmanaged_conflict (§8.3, §9.5).
func inventoryUnmanaged(homeDir string, want map[string]bool, storeRoot string) error {
	paths := make([]string, 0, len(want))
	for path := range want {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		full := filepath.Join(homeDir, filepath.FromSlash(path))
		metadata, err := stateread.Lstat(full)
		if err != nil {
			return fmt.Errorf("%s: inventory of %s: %v", DiagUnmanagedConflict, path, err)
		}
		if metadata.Kind == stateread.KindAbsent {
			continue
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			return fmt.Errorf("%s: inventory of %s: path metadata is unreadable", DiagUnmanagedConflict, path)
		}
		info := metadata.Info
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := stateread.Readlink(full)
			if err != nil || link.Kind != stateread.KindPresent {
				return fmt.Errorf("%s: inventory of %s: symbolic link target is unreadable: %v", DiagUnmanagedConflict, path, err)
			}
			if !sameStoreTree(link.Target, storeRoot) {
				return fmt.Errorf("%s: %s is a symlink outside the manager store; abort, or take over with backup", DiagForeignManager, path)
			}
		}
		return fmt.Errorf("%s: %s exists and no marker records it", DiagUnmanagedConflict, path)
	}
	return nil
}

// preflightManagedWriteTargets checks every managed-surface parent and the
// private marker destination before an operation writes store documents,
// backups, surfaces, or the marker.
func preflightManagedWriteTargets(root, homeRel string, want map[string]bool) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	for path := range want {
		if _, err := managedPath(root, managedJoin(homeRel, path), false); err != nil {
			return err
		}
	}
	return checkManagedPrivateTarget(root, managedJoin(homeRel, envmarker.Name))
}

// sameStoreTree reports whether the link target lives below the store.
func sameStoreTree(target, storeRoot string) bool {
	if !filepath.IsAbs(target) {
		return false
	}
	root := filepath.Clean(storeRoot)
	return target == root || strings.HasPrefix(target, root+string(filepath.Separator))
}

// applyPlan writes the plan into the home under the held mutation lock:
// store documents, copies, links (with copy fallback for files), seed
// files, and the marker. Managed paths the plan no longer wants are
// removed; files the marker does not record are never touched.
func applyPlan(op *operation, req *ResolveRequest, plan *homePlan, seeds *seedBundle, recorded map[string]bool, prior *envmarker.Marker, provisioned bool) error {
	root := EnvRoot(req.Home)
	homeRel, err := managedRelative(root, plan.homeDir)
	if err != nil {
		return err
	}
	want := map[string]bool{}
	for path := range plan.copies {
		want[path] = true
	}
	for path := range plan.links {
		want[path] = true
	}
	if seeds.claudeInit {
		want[".claude.json"] = true
	}
	for seed := range seeds.files {
		want[seed] = true
	}
	// A new member can add a managed path beneath an already-provisioned
	// home. Until the marker records that path it remains unmanaged, including
	// during a repair of an otherwise managed home.
	unmanaged := unmanagedPlanTargets(plan, recorded)
	if len(unmanaged) > 0 {
		if err := inventoryUnmanaged(plan.homeDir, unmanaged, contextstore.Root(req.Home)); err != nil {
			if !req.Policy.Takeover {
				return err
			}
			if prior != nil {
				plan.warnings = append(plan.warnings, "taking over unmanaged files for "+plan.adapter.ID+
					": native global context files are being replaced by managed ones; backups land in "+
					filepath.Join(plan.homeDir, ".agent-environment-backup")+"/")
			}
		}
	}
	if err := preflightManagedWriteTargets(root, homeRel, want); err != nil {
		return err
	}
	if _, err := managedDirectory(root, homeRel); err != nil {
		return err
	}
	if err := plan.publishDocs(); err != nil {
		return err
	}
	// Versioned generations preserve replaced regular-file bytes (§8.3,
	// §9.5) before the first write. An authorized takeover also backs up a
	// replaced unmanaged symlink as a symlink, preserving its link text;
	// manager-recorded links are re-derived from the lock and are not backed
	// up. A next generation that already exists fails with
	// environment_backup_exists.
	replacing := map[string]bool{}
	for path := range want {
		full, err := managedPath(root, managedJoin(homeRel, path), false)
		if err != nil {
			return err
		}
		state, err := stateread.Lstat(full)
		if err != nil {
			return err
		}
		if state.Kind == stateread.KindAbsent {
			continue
		}
		info := state.Info
		if info.Mode()&os.ModeSymlink != 0 {
			if provisioned && req.Policy.Takeover && !recorded[path] {
				replacing[path] = true
			}
			continue
		}
		if info.Mode().IsRegular() {
			replacing[path] = true
		}
	}
	if len(replacing) > 0 {
		if _, err := openBackup(root, homeRel, replacing); err != nil {
			return err
		}
	}
	// Remove recorded files the plan no longer wants.
	for path := range recorded {
		if !want[path] && !isSeedPath(plan, path) {
			if err := removeManagedEntry(root, managedJoin(homeRel, path)); err != nil {
				return err
			}
		}
	}
	// Writes run in sorted path order: the fresh-home provisioning order
	// (§8.1) is deterministic even though the plan accumulates maps. Each
	// destination is replaced from a private same-directory entry, so the
	// existing surface path is never opened through a symlink.
	for _, path := range sortedKeysBytes(plan.copies) {
		document := plan.copies[path]
		if err := atomicManagedFile(root, managedJoin(homeRel, path), document, 0o644); err != nil {
			return err
		}
	}
	for _, path := range sortedKeys(plan.links) {
		target := plan.links[path]
		rel := managedJoin(homeRel, path)
		if err := atomicManagedLink(root, rel, target); err != nil {
			var unavailable *errSymlinkUnavailable
			if !errors.As(err, &unavailable) {
				return fmt.Errorf("link %s: %v", path, err)
			}
			// Copy fallback where the platform takes no symlink: record
			// the copy with its reason so hash drift applies to it.
			if err := copyLinkFallback(root, rel, target, path, plan); err != nil {
				return err
			}
		}
	}
	// Copy seeds are written at provisioning only: after that the tool
	// owns them, so repair never refreshes them (§7.4).
	if provisioned {
		for seed, payload := range seeds.files {
			if err := atomicManagedFile(root, managedJoin(homeRel, seed), payload, 0o644); err != nil {
				return err
			}
		}
	}
	provenance := "repaired"
	if provisioned {
		provenance = "provisioned"
	}
	marker, err := finalizeMarker(req, plan, seeds, prior, provenance)
	if err != nil {
		return err
	}
	payload, err := marker.Marshal()
	if err != nil {
		return err
	}
	markerPath := filepath.Join(plan.homeDir, envmarker.Name)
	if prior != nil && prior.Version == envmarker.VersionV1 && legacyMarkerProjectionMatches(prior, marker) {
		return nil
	}
	return op.publish(map[string][]byte{markerPath: payload})
}

// isSeedPath reports whether a recorded path is seed-owned tool state the
// repair must not remove.
func isSeedPath(plan *homePlan, path string) bool {
	if plan.marker.Seeds == nil {
		return false
	}
	for _, seed := range *plan.marker.Seeds {
		if seed == path {
			return true
		}
	}
	return false
}

// finalizeMarker records the passthrough entries, provisioning seeds,
// seeded launch-directory projects, and XDG seed links on the assembled
// marker, then reconciles the opencode XDG links: newly present allowlisted
// entries are seeded and recorded, recorded seeds whose target is gone are
// removed, and unrecorded entries shadowing allowlisted operator entries
// warn environment_seed_shadowed and are never touched (§7.1, §7.4).
func finalizeMarker(req *ResolveRequest, plan *homePlan, seeds *seedBundle, prior *envmarker.Marker, provenance string) (*envmarker.Marker, error) {
	root := EnvRoot(req.Home)
	homeRel, err := managedRelative(root, plan.homeDir)
	if err != nil {
		return nil, err
	}
	marker := plan.marker
	marker.Version = envmarker.VersionV2
	if prior != nil && prior.CodexSeedRecord != nil {
		record := *prior.CodexSeedRecord
		record.NativeMCPServers = append([]string{}, prior.CodexSeedRecord.NativeMCPServers...)
		marker.CodexSeedRecord = &record
	}
	if seeds.codexSeedRecord != nil {
		record := *seeds.codexSeedRecord
		record.NativeMCPServers = append([]string{}, seeds.codexSeedRecord.NativeMCPServers...)
		marker.CodexSeedRecord = &record
	}
	links, _, err := req.effectivePassthrough(plan.adapter, plan.isolation)
	if err != nil {
		return nil, err
	}
	native, err := req.nativeHome(plan.adapter.ID)
	if err != nil {
		return nil, err
	}
	recordedLinks := map[string]bool{}
	if prior != nil && prior.Passthrough != nil {
		for _, entry := range *prior.Passthrough {
			if entry.Path != "" {
				recordedLinks[entry.Path] = true
			}
		}
	}
	migrateCmd := fmt.Sprintf("curator env migrate --plan --profile %s --env %s, then curator env migrate --apply --expect <plan-hash> --profile %s --env %s", req.Profile, plan.adapter.ID, req.Profile, plan.adapter.ID)
	for _, path := range sortedKeys(links) {
		if err := ensureCredentialLink(root, homeRel, path, links[path], recordedLinks[path], migrateCmd); err != nil {
			return nil, err
		}
	}
	if err := removeStaleCredentialLinks(plan.adapter, plan.homeDir, native, prior, links, migrateCmd); err != nil {
		return nil, err
	}
	passthrough, err := req.credentialRecords(plan.adapter, plan.isolation, provenance)
	if err != nil {
		return nil, err
	}
	marker.Passthrough = &passthrough
	// Copy-seed records survive repair: the tool owns the files, and
	// repair never refreshes them, so the record unions prior with new.
	recordedSeeds := []string{}
	seenSeeds := map[string]bool{}
	if prior != nil && prior.Seeds != nil {
		for _, seed := range *prior.Seeds {
			if !seenSeeds[seed] {
				seenSeeds[seed] = true
				recordedSeeds = append(recordedSeeds, seed)
			}
		}
	}
	for seed := range seeds.files {
		if !seenSeeds[seed] {
			seenSeeds[seed] = true
			recordedSeeds = append(recordedSeeds, seed)
		}
	}
	if seeds.claudeInit || plan.adapter.ID == envregistry.ClaudeCode {
		seeded, err := claudeSeed(root, homeRel, req.LaunchDir, plan.form)
		if err != nil {
			return nil, err
		}
		marker.SeededProjects = seeded
		recordedSeeds = append(recordedSeeds, ".claude.json")
	}
	sort.Strings(recordedSeeds)
	marker.Seeds = &recordedSeeds
	seedLinks := []string{}
	if plan.adapter.ID == envregistry.OpenCode {
		for _, name := range sortedKeys(seeds.xdg) {
			full := filepath.Join(plan.parent, name)
			rel, err := managedRelative(root, full)
			if err != nil {
				return nil, err
			}
			if err := atomicManagedLink(root, rel, seeds.xdg[name]); err != nil {
				return nil, fmt.Errorf("xdg seed %s: %v", name, err)
			}
			seedLinks = append(seedLinks, name)
		}
		// A recorded seed whose target no longer exists is removed (§7.1).
		if prior != nil {
			for _, name := range prior.SeedLinks {
				if _, ok := seeds.xdg[name]; !ok {
					full := filepath.Join(plan.parent, name)
					rel, err := managedRelative(root, full)
					if err != nil {
						return nil, err
					}
					if err := removeManagedEntry(root, rel); err != nil {
						return nil, err
					}
				}
			}
		}
		plan.warnings = append(plan.warnings, shadowWarnings(req, plan, seedLinks)...)
	}
	marker.SeedLinks = seedLinks
	return marker, nil
}

// legacyMarkerProjectionMatches checks whether publishing a schema-2 marker
// would change any schema-1 state beyond adding credential metadata. A true
// result keeps the original schema-1 bytes untouched; any unrelated marker
// change is published as one complete schema-2 replacement.
func legacyMarkerProjectionMatches(prior, candidate *envmarker.Marker) bool {
	if prior == nil || prior.Version != envmarker.VersionV1 || candidate == nil {
		return false
	}
	legacyEntries := []envmarker.Passthrough{}
	if candidate.Passthrough != nil {
		for _, entry := range *candidate.Passthrough {
			if entry.Path == "" {
				continue
			}
			strategy := entry.Strategy
			if strategy == envregistry.StrategyKeyringPreferred && entry.Backend == "file" {
				strategy = envregistry.StrategyFileLink
			}
			legacyEntries = append(legacyEntries, envmarker.Passthrough{Path: entry.Path, Strategy: strategy})
		}
	}
	projected := *candidate
	projected.Version = envmarker.VersionV1
	projected.Passthrough = &legacyEntries
	// codex_seed_record is valid in schema 1 as well as schema 2. Preserve it
	// in the legacy projection so metadata-only repair does not rewrite a
	// schema-1 marker that already carries the rc.13 seed snapshot.
	priorPayload, err := prior.Marshal()
	if err != nil {
		return false
	}
	projectedPayload, err := projected.Marshal()
	if err != nil {
		return false
	}
	return bytes.Equal(priorPayload, projectedPayload)
}

// sameCredentialRecordSet compares all durable record fields except
// provenance, which describes the last successful mutation and therefore
// changes on repair and migration without changing credential identity.
func sameCredentialRecordSet(left, right []envmarker.Passthrough) bool {
	if len(left) != len(right) {
		return false
	}
	identityCounts := func(records []envmarker.Passthrough) map[string]int {
		counts := make(map[string]int, len(records))
		for _, record := range records {
			record.Provenance = ""
			payload, _ := json.Marshal(record)
			counts[string(payload)]++
		}
		return counts
	}
	a, b := identityCounts(left), identityCounts(right)
	if len(a) != len(b) {
		return false
	}
	for key, count := range a {
		if b[key] != count {
			return false
		}
	}
	return true
}

// ensureCredentialLink fix-first ensures one wanted credential link (§7.4,
// §10.1): an absent link path is linked, a symlink still targeting the
// declared native store is left alone, and an empty directory is replaced
// by the link — none of which moves credential ownership. A recorded
// symlink to any other target — the mis-targeted state, including a Pi
// home still linked at the pre-0017 native root — is never re-pointed
// here: repair refuses with environment_credential_conflict naming the
// path and points at the explicit migration (migrateCmd). A regular
// file, an unrecorded symlink to an unexpected target, a non-empty
// directory, or a link whose state cannot be established refuses the
// same way, removing nothing. Native bytes are never touched.
func ensureCredentialLink(root, homeRel, path, target string, recorded bool, migrateCmd string) error {
	rel := managedJoin(homeRel, path)
	full, err := managedPath(root, rel, false)
	if err != nil {
		return err
	}
	metadata, err := stateread.Lstat(full)
	if err != nil {
		return fmt.Errorf("%s: %s cannot be inspected: %v: refusing to touch it; restore access out of band and re-run", envregistry.DiagCredentialConflict, full, err)
	}
	if metadata.Kind == stateread.KindAbsent {
		if err := atomicManagedLink(root, rel, target); err != nil {
			return fmt.Errorf("passthrough %s: %v", path, err)
		}
		return nil
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		return fmt.Errorf("%s: %s cannot be inspected: %v: refusing to touch it; restore access out of band and re-run", envregistry.DiagCredentialConflict, full, stateread.UnusableError(full, fmt.Errorf("unknown metadata state %q", metadata.Kind)))
	}
	info := metadata.Info
	if info.Mode()&os.ModeSymlink != 0 {
		got, err := os.Readlink(full)
		if err != nil {
			return fmt.Errorf("%s: %s link target cannot be read: refusing to touch it; restore access out of band and re-run", envregistry.DiagCredentialConflict, full)
		}
		if got == target {
			return nil
		}
		if !recorded {
			return fmt.Errorf("%s: %s links to %s, expected %s: refusing to remove or re-point it; remove the unrecorded link out of band and re-run", envregistry.DiagCredentialConflict, full, got, target)
		}
		return fmt.Errorf("%s: %s links to %s, expected %s: migration needed: re-point the recorded link with `%s`; refusing to re-point it here", envregistry.DiagCredentialConflict, full, got, target, migrateCmd)
	}
	if info.IsDir() {
		entries, err := os.ReadDir(full)
		if err != nil {
			return fmt.Errorf("%s: %s holds a directory that cannot be listed: refusing to touch it; restore access out of band and re-run", envregistry.DiagCredentialConflict, full)
		}
		if len(entries) != 0 {
			return fmt.Errorf("%s: %s holds a non-empty directory, expected a link to %s: refusing to remove it; clear the directory out of band and re-run", envregistry.DiagCredentialConflict, full, target)
		}
		if err := removeManagedEntry(root, rel); err != nil {
			return err
		}
		if err := atomicManagedLink(root, rel, target); err != nil {
			return fmt.Errorf("passthrough %s: %v", path, err)
		}
		return nil
	}
	return fmt.Errorf("%s: %s holds a regular file, expected a link to %s: refusing to remove or replace it; decide which credential bytes win and move the loser aside out of band (the manager never moves credential bytes)", envregistry.DiagCredentialConflict, full, target)
}

// removeStaleCredentialLinks refuses recorded credential links the
// effective set no longer wants — the shared→isolated turn, or a native
// store gone ambient (§7.4, §10.1): a mode change that leaves a recorded
// link behind it is the explicit migration, never silent inside repair.
// A stale recorded symlink still targeting the declared native store
// refuses with environment_credential_conflict naming the path and
// points at the explicit migration (migrateCmd); an already-absent path
// is already gone and the caller drops the record, so hand-removed and
// crash-recovery shapes converge here. Anything else — a regular file,
// a symlink to any other target, or a directory — refuses the same way,
// removing nothing: the operator resolves the conflict out of band and
// re-runs.
func removeStaleCredentialLinks(adapter envregistry.Adapter, homeDir, native string, prior *envmarker.Marker, links map[string]string, migrateCmd string) error {
	if prior == nil || prior.Passthrough == nil {
		return nil
	}
	for _, entry := range *prior.Passthrough {
		if entry.Path == "" {
			continue
		}
		if _, wanted := links[entry.Path]; wanted {
			continue
		}
		full := filepath.Join(homeDir, filepath.FromSlash(entry.Path))
		metadata, err := stateread.Lstat(full)
		if err != nil {
			return fmt.Errorf("%s: %s cannot be inspected: %v: refusing to touch it; restore access out of band and re-run", envregistry.DiagCredentialConflict, full, err)
		}
		if metadata.Kind == stateread.KindAbsent {
			continue
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			return fmt.Errorf("%s: %s cannot be inspected: %v: refusing to touch it; restore access out of band and re-run", envregistry.DiagCredentialConflict, full, stateread.UnusableError(full, fmt.Errorf("unknown metadata state %q", metadata.Kind)))
		}
		info := metadata.Info
		if info.Mode()&os.ModeSymlink == 0 {
			if info.IsDir() {
				return fmt.Errorf("%s: %s holds a directory where the recorded credential link %s was: refusing to remove it; clear the directory out of band and re-run", envregistry.DiagCredentialConflict, full, entry.Path)
			}
			return fmt.Errorf("%s: %s holds a regular file where the recorded credential link %s was: refusing to remove it; decide which credential bytes win and move the loser aside out of band (the manager never moves credential bytes)", envregistry.DiagCredentialConflict, full, entry.Path)
		}
		got, err := os.Readlink(full)
		if err != nil {
			return fmt.Errorf("%s: %s link target cannot be read: refusing to touch it; restore access out of band and re-run", envregistry.DiagCredentialConflict, full)
		}
		declared, ok := declaredPassthroughTarget(adapter, native, entry.Path)
		if !ok {
			return fmt.Errorf("%s: %s has no declared native store for the recorded credential link %s: refusing to remove it; remove the link out of band and re-run", envregistry.DiagCredentialConflict, full, entry.Path)
		}
		if got != declared {
			return fmt.Errorf("%s: %s links to %s, not the recorded credential target %s: refusing to remove it; remove the link out of band and re-run", envregistry.DiagCredentialConflict, full, got, declared)
		}
		return fmt.Errorf("%s: %s is a stale recorded credential link (still targeting the declared store %s): migration needed: unlink it with `%s`; refusing to unlink it here", envregistry.DiagCredentialConflict, full, declared, migrateCmd)
	}
	return nil
}

// declaredPassthroughTarget reconstructs the native target a recorded
// credential link must still point at: the adapter entry's file target
// below the native home. It reports false where the recorded path names
// no linkable adapter entry, in which case nothing at the path can be
// proven ours and repair refuses rather than unlinks.
func declaredPassthroughTarget(adapter envregistry.Adapter, native, path string) (string, bool) {
	for _, entry := range adapter.PassthroughFor(runtime.GOOS) {
		if entry.Path != path {
			continue
		}
		switch entry.Strategy {
		case envregistry.StrategyFileLink, envregistry.StrategyKeyringPreferred:
			return filepath.Join(native, filepath.FromSlash(entry.FileLinkTarget)), true
		default:
			return "", false
		}
	}
	return "", false
}

// shadowWarnings reports unrecorded parent entries shadowing allowlisted
// operator entries as environment_seed_shadowed, left as they are (§7.1).
func shadowWarnings(req *ResolveRequest, plan *homePlan, recorded []string) []string {
	known := map[string]bool{}
	for _, name := range recorded {
		known[name] = true
	}
	xdg := req.operatorXDG()
	var warnings []string
	for _, name := range req.Machine.XDGSeedAllowlist {
		if name == "opencode" || known[name] {
			continue
		}
		if xdg == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(xdg, name)); err != nil {
			continue
		}
		if _, err := os.Lstat(filepath.Join(plan.parent, name)); err == nil {
			warnings = append(warnings, fmt.Sprintf("%s: %s shadows the allowlisted operator entry and is left as it is", envregistry.DiagSeedShadowed, name))
		}
	}
	return warnings
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysBytes(values map[string][]byte) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// verification is the lock-free currency verdict over exactly the surfaces
// the marker records (§10.1).
type verification struct {
	current               bool
	reasons               []string
	warnings              []string
	passthroughReadFailed bool
	markerReadFailed      bool
	// surfaceState carries the per-surface outcome for status rows: ""
	// (current), environment_surface_drift, environment_surface_missing,
	// or environment_surface_unreadable.
	surfaceState map[string]string
	plan         *homePlan
	marker       *envmarker.Marker
	hash         string
	order        []contextlock.Member
	servers      []contextmaterialize.MCPServer
}

// SurfaceDiagnostics for status rows (environments §8.5).
const (
	DiagSurfaceDrift      = "environment_surface_drift"
	DiagSurfaceMissing    = "environment_surface_missing"
	DiagSurfaceUnreadable = "environment_surface_unreadable"
)

// verifyHome verifies the managed home lock-free: it reads the marker and
// covers exactly the surfaces the marker records — no more. For a symlinked
// surface into an immutable store entry, link-target identity is sufficient
// currency (§10.1); a copied surface is verified by its recorded content
// hash. Assembly errors from the store are reported as staleness reasons
// here (fail-closed, no fragment); the same error under the repair lock
// surfaces as environment_repair_failed.
func verifyHome(req *ResolveRequest, adapter envregistry.Adapter, source Source, lock *contextlock.Lock, hash string) *verification {
	verdict := &verification{surfaceState: map[string]string{}}
	precedence := req.Policy.Precedence()
	order, err := contextmaterialize.EmittedOrder(lock, precedence)
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("emitted order: %v", err))
		return verdict
	}
	verdict.order = order
	readStateFile := req.readStateFile
	if readStateFile == nil {
		readStateFile = stateread.ReadFile
	}
	marker, err := envmarker.ReadWith(ManagedHomeDir(req.Home, req.Profile, adapter.ID), readStateFile)
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("marker invalid: %v", err))
		if strings.Contains(err.Error(), envmarker.DiagMarkerUnreadable) {
			verdict.markerReadFailed = true
		}
		return verdict
	}
	if marker == nil {
		verdict.reasons = append(verdict.reasons, "home unprovisioned")
		return verdict
	}
	verdict.marker = marker
	manager := newGitManager(req.Home)
	servers, err := mcpsOf(req.Home, manager, lock)
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("mcp members: %v", err))
		return verdict
	}
	verdict.servers = servers
	plan, err := assembleHome(req, source, lock, hash, precedence, order, adapter, marker)
	if err != nil {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("expected home: %v", err))
		return verdict
	}
	verdict.plan = plan
	verdict.hash = hash
	verdict.warnings = append(verdict.warnings, plan.warnings...)
	if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, "sha256:") {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("lock %s is not the current %s", marker.Profile.LockSHA256, strings.TrimPrefix(hash, "sha256:")))
	}
	if marker.Mode != envmarker.ModeManagedHome {
		verdict.reasons = append(verdict.reasons, fmt.Sprintf("mode %s is not managed-home", marker.Mode))
	}
	if marker.Precedence.Winner != precedence.Winner || marker.Precedence.Placement != precedence.Placement {
		verdict.reasons = append(verdict.reasons, "precedence does not match the effective policy")
	}
	if len(marker.Members) != len(order) {
		verdict.reasons = append(verdict.reasons, "member list does not match the lock")
	} else {
		for i, member := range order {
			recorded := marker.Members[i]
			pin := member.Commit
			if pin == "" {
				pin = member.StateHash
			}
			have := recorded.Commit
			if have == "" {
				have = recorded.StateSHA256
			}
			if recorded.Name != member.Name || recorded.Weight != member.Weight || recorded.Overlay != member.Overlay || have != pin {
				verdict.reasons = append(verdict.reasons, fmt.Sprintf("member %s does not match the lock", member.Name))
				break
			}
		}
	}
	verdict.checkSurfaces(req, plan, marker)
	verdict.checkPassthrough(req, plan, marker)
	verdict.checkClaudeProject(req, plan)
	verdict.checkXDG(req, plan, marker)
	verdict.checkShadows(plan, marker)
	verdict.checkToolRelease(req, adapter)
	verdict.current = len(verdict.reasons) == 0
	return verdict
}

// checkSurfaces covers exactly the recorded surfaces.
func (v *verification) checkSurfaces(req *ResolveRequest, plan *homePlan, marker *envmarker.Marker) {
	want := plan.marker.Surfaces
	for key, recorded := range marker.Surfaces {
		expected, ok := want[key]
		if !ok {
			v.reasons = append(v.reasons, fmt.Sprintf("surface %s is no longer materialized", key))
			v.mark(key, DiagSurfaceMissing)
			continue
		}
		if !equalStrings(recorded.Paths, expected.Paths) || recorded.ContentSHA256 != expected.ContentSHA256 {
			v.reasons = append(v.reasons, fmt.Sprintf("surface %s record does not match", key))
			v.mark(key, DiagSurfaceDrift)
			continue
		}
		if key == envmarker.SurfaceRootContext && recorded.Form != plan.form {
			v.reasons = append(v.reasons, fmt.Sprintf("root-context form %s is not the effective %s", recorded.Form, plan.form))
			v.mark(key, DiagSurfaceDrift)
		}
		copiedPaths := map[string]bool{}
		if recorded.Copies != nil {
			for _, copy := range *recorded.Copies {
				copiedPaths[copy.Path] = true
			}
		}
		for _, path := range recorded.Paths {
			full := filepath.Join(plan.homeDir, filepath.FromSlash(path))
			target, linked := plan.links[path]
			if linked && !copiedPaths[path] {
				metadata, err := stateread.Lstat(full)
				if err != nil || metadata.Kind == stateread.KindUnreadable {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is unreadable: %v", path, err))
					v.mark(key, DiagSurfaceUnreadable)
					continue
				}
				if metadata.Kind == stateread.KindAbsent {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is missing", path))
					v.mark(key, DiagSurfaceMissing)
					continue
				}
				info := metadata.Info
				if info == nil {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is unreadable: metadata is unavailable", path))
					v.mark(key, DiagSurfaceUnreadable)
					continue
				}
				if info.Mode()&os.ModeSymlink == 0 {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: expected a store link", path))
					v.mark(key, DiagSurfaceDrift)
					continue
				}
				link, err := stateread.Readlink(full)
				if err != nil || link.Kind == stateread.KindUnreadable {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s link target is unreadable: %v", path, err))
					v.mark(key, DiagSurfaceUnreadable)
					continue
				}
				if link.Kind == stateread.KindAbsent {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is missing", path))
					v.mark(key, DiagSurfaceMissing)
					continue
				}
				if link.Target != target {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: link target changed", path))
					v.mark(key, DiagSurfaceDrift)
					continue
				}
				// A link into the immutable profile store is verified by
				// target identity alone (the store entry's integrity is
				// the store's own invariant, §10.1). Every other link
				// targets a manager-authored rendered document published
				// through p.docs, so its bytes must still match the
				// recorded hash (§8.4): a write through the intact link
				// is drift even though the link is unchanged.
				if sameStoreTree(target, contextstore.Root(req.Home)) {
					continue
				}
				expected, ok := plan.fileHashes[path]
				if !ok {
					expected, ok = plan.docs[target]
				}
				if !ok {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: no recorded bytes for link target", path))
					v.mark(key, DiagSurfaceDrift)
					continue
				}
				state, err := stateread.ReadFile(target) // #nosec G304 -- link target recomputed from the verified plan
				if err != nil || state.Kind == stateread.KindUnreadable {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s link target is unreadable: %v", path, err))
					v.mark(key, DiagSurfaceUnreadable)
					continue
				}
				if state.Kind == stateread.KindAbsent {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s link target is missing", path))
					v.mark(key, DiagSurfaceMissing)
					continue
				}
				if contextmaterialize.FileHash(state.Bytes) != contextmaterialize.FileHash(expected) {
					v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: link target bytes differ", path))
					v.mark(key, DiagSurfaceDrift)
				}
				continue
			}
			// A copied surface — the claude_code root file, a pi live
			// channel, or a recorded symlink-fallback copy — is verified
			// by the content hash (§10.1), read against the plan bytes
			// or, for a fallback copy, against the store source.
			if document, ok := plan.copies[path]; ok {
				v.mark(key, v.checkCopy(path, full, document))
				continue
			}
			if linked {
				v.mark(key, v.checkFallbackCopy(path, full, target))
				continue
			}
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is not in the plan", path))
			v.mark(key, DiagSurfaceDrift)
		}
	}
	for key := range want {
		if _, ok := marker.Surfaces[key]; !ok {
			v.reasons = append(v.reasons, fmt.Sprintf("surface %s is missing", key))
			v.mark(key, DiagSurfaceMissing)
		}
	}
}

// checkCopy verifies one copied file by its content hash (§10.1). A
// missing file and an unreadable file are different facts (§8.4). It
// returns the surface state for the status row.
func (v *verification) checkCopy(path, full string, document []byte) string {
	state, err := stateread.ReadFile(full) // #nosec G304 -- home path from the verified plan
	if err != nil || state.Kind == stateread.KindUnreadable {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is unreadable: %v", path, err))
		return DiagSurfaceUnreadable
	}
	if state.Kind == stateread.KindAbsent {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is missing", path))
		return DiagSurfaceMissing
	}
	if contextmaterialize.FileHash(state.Bytes) != contextmaterialize.FileHash(document) {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: content hash differs", path))
		return DiagSurfaceDrift
	}
	return ""
}

// checkFallbackCopy verifies a recorded symlink-fallback copy against its
// store source: file bytes for files, the store content hash for trees.
func (v *verification) checkFallbackCopy(path, full, source string) string {
	metadata, err := stateread.Stat(source)
	if err != nil || metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s store source is unreadable: %v", path, err))
		return DiagSurfaceUnreadable
	}
	info := metadata.Info
	if info.IsDir() {
		want, err := contextstore.ContentHash(source)
		if err != nil {
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s store source is unreadable: %v", path, err))
			return DiagSurfaceUnreadable
		}
		got, err := contextstore.ContentHash(full)
		if err != nil {
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is unreadable: %v", path, err))
			return DiagSurfaceUnreadable
		}
		if got != want {
			v.reasons = append(v.reasons, fmt.Sprintf("surface file %s is drifted: content hash differs", path))
			return DiagSurfaceDrift
		}
		return ""
	}
	payload, err := os.ReadFile(source) // #nosec G304 -- store source from the verified plan
	if err != nil {
		v.reasons = append(v.reasons, fmt.Sprintf("surface file %s store source is unreadable: %v", path, err))
		return DiagSurfaceUnreadable
	}
	return v.checkCopy(path, full, payload)
}

// mark records the per-surface outcome, keeping the worst state: missing
// beats unreadable beats drifted beats current.
func (v *verification) mark(key, state string) {
	rank := map[string]int{"": 0, DiagSurfaceDrift: 1, DiagSurfaceUnreadable: 2, DiagSurfaceMissing: 3}
	if rank[state] > rank[v.surfaceState[key]] {
		v.surfaceState[key] = state
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// checkPassthrough runs the liveness row: every recorded entry must still
// be a symlink targeting the native entry, and the recorded set must
// match the effective one (§7.4). A missing link path is plain detached
// and is re-linked by --repair; a link path holding a regular file or a
// symlink to an unexpected target — including a Pi home still linked at
// the pre-0017 native target — is detached with
// environment_credential_conflict-class wording, never silence:
// --repair refuses without touching bytes and points at the explicit
// migration where it can fix the state. A
// correctly targeted link whose native target does not exist yet is the
// distinct detached-pending state — the normal provisioning shape before
// the first login (claude_code on Linux) — reported as a warning, never
// a stale reason and never silence: provisioning, repair, and bare
// resolve all succeed loudly. A native target that cannot be inspected
// is a conflict/inspection diagnostic and stays stale, never absence
// and never silence.
func (v *verification) checkPassthrough(req *ResolveRequest, plan *homePlan, marker *envmarker.Marker) {
	links, strategies, err := req.effectivePassthrough(plan.adapter, plan.isolation)
	if err != nil {
		v.reasons = append(v.reasons, fmt.Sprintf("passthrough: %v", err))
		return
	}
	if marker.Version == envmarker.VersionV2 {
		expected, err := req.credentialRecords(plan.adapter, plan.isolation, "provisioned")
		if err != nil {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough: %v", err))
			return
		}
		var recorded []envmarker.Passthrough
		if marker.Passthrough != nil {
			recorded = *marker.Passthrough
		}
		if !sameCredentialRecordSet(expected, recorded) {
			v.reasons = append(v.reasons, "passthrough credential records do not match the effective set")
			return
		}
	}
	recorded := map[string]string{}
	if marker.Passthrough != nil {
		for _, entry := range *marker.Passthrough {
			if entry.Path != "" {
				recorded[entry.Path] = entry.Strategy
			}
		}
	}
	if len(recorded) != len(links) {
		v.reasons = append(v.reasons, "passthrough entries do not match the effective set")
		return
	}
	for path, target := range links {
		strategy, ok := recorded[path]
		if !ok || (marker.Version == envmarker.VersionV1 && strategy != strategies[path]) {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is not recorded", path))
			continue
		}
		full := filepath.Join(plan.homeDir, filepath.FromSlash(path))
		var metadata stateread.Metadata
		if req.passthroughLstat == nil {
			metadata, err = stateread.Lstat(full)
		} else {
			metadata, err = stateread.LstatWith(full, req.passthroughLstat)
		}
		if err != nil || metadata.Kind == stateread.KindUnreadable {
			v.passthroughReadFailed = true
			if err == nil {
				err = stateread.UnusableError(full, fmt.Errorf("entry inspection returned unreadable without an error"))
			}
			v.reasons = append(v.reasons, fmt.Sprintf("%s: passthrough entry %s cannot be inspected: %v", envregistry.DiagPassthroughUnreadable, path, err))
			continue
		}
		if metadata.Kind == stateread.KindAbsent {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is detached", path))
			continue
		}
		info := metadata.Info
		if info == nil {
			v.passthroughReadFailed = true
			v.reasons = append(v.reasons, fmt.Sprintf("%s: passthrough entry %s has no readable metadata", envregistry.DiagPassthroughUnreadable, path))
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			if !info.IsDir() {
				v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is detached: link path holds a regular file, expected a link to %s (%s)", path, target, envregistry.DiagCredentialConflict))
				continue
			}
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is detached", path))
			continue
		}
		var link stateread.Link
		if req.passthroughReadlink == nil {
			link, err = stateread.Readlink(full)
		} else {
			link, err = stateread.ReadlinkWith(full, req.passthroughReadlink)
		}
		if err != nil || link.Kind == stateread.KindUnreadable {
			v.passthroughReadFailed = true
			if err == nil {
				err = stateread.UnusableError(full, fmt.Errorf("link inspection returned unreadable without an error"))
			}
			v.reasons = append(v.reasons, fmt.Sprintf("%s: passthrough entry %s target cannot be read: %v", envregistry.DiagPassthroughUnreadable, path, err))
			continue
		}
		if link.Kind == stateread.KindAbsent {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is detached", path))
			continue
		}
		got := link.Target
		if got != target {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s is detached: link targets %s, expected %s (%s)", path, got, target, envregistry.DiagCredentialConflict))
			continue
		}
		// The link targets the declared native store; liveness asks
		// whether the native target itself exists — a stat of the
		// target, never a read of credential bytes. An absent target
		// is detached-pending: the link is correct and the native
		// credential simply is not there yet, so the state warns
		// loudly without going stale. An inspection failure is a
		// conflict/inspection diagnostic and stays stale, never
		// absence and never silence.
		metadata, err := stateread.Stat(target)
		if err != nil {
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s target %s cannot be inspected: %v (%s)", path, target, err, envregistry.DiagCredentialConflict))
			continue
		}
		switch metadata.Kind {
		case stateread.KindAbsent:
			v.warnings = append(v.warnings, fmt.Sprintf("passthrough entry %s is detached-pending: link target %s does not exist yet — log in to %s to populate it", path, target, plan.adapter.ID))
		case stateread.KindPresent:
			if metadata.Info == nil {
				v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s target %s cannot be inspected: %v (%s)", path, target, stateread.UnusableError(target, fmt.Errorf("present target has no metadata")), envregistry.DiagCredentialConflict))
			}
		default:
			v.reasons = append(v.reasons, fmt.Sprintf("passthrough entry %s target %s cannot be inspected: %v (%s)", path, target, stateread.UnusableError(target, fmt.Errorf("unknown metadata state %q", metadata.Kind)), envregistry.DiagCredentialConflict))
		}
	}
}

// checkClaudeProject enforces the per-launch-directory project entry under
// the referenced form: a launch directory without its entry makes the home
// stale for that directory (§7.4, §10.1). Under monolithic the entry is
// still merged on repair, but its absence is not staleness.
func (v *verification) checkClaudeProject(req *ResolveRequest, plan *homePlan) {
	if plan.adapter.ID != envregistry.ClaudeCode {
		return
	}
	entries, err := claudeProjects(plan.homeDir)
	if err != nil {
		v.reasons = append(v.reasons, fmt.Sprintf("managed .claude.json is unreadable: %v", err))
		return
	}
	if plan.form == envregistry.FormReferenced && !entries[req.LaunchDir] {
		v.reasons = append(v.reasons, fmt.Sprintf("launch directory %s has no external-includes approval", req.LaunchDir))
	}
}

// checkXDG verifies the recorded XDG seed links and warns on shadows.
func (v *verification) checkXDG(req *ResolveRequest, plan *homePlan, marker *envmarker.Marker) {
	if plan.adapter.ID != envregistry.OpenCode {
		return
	}
	xdg := req.operatorXDG()
	expected := map[string]string{}
	if xdg != "" {
		for _, name := range req.Machine.XDGSeedAllowlist {
			if name == "opencode" {
				continue
			}
			if _, err := os.Stat(filepath.Join(xdg, name)); err == nil {
				expected[name] = filepath.Join(xdg, name)
			}
		}
	}
	recorded := map[string]bool{}
	for _, name := range marker.SeedLinks {
		recorded[name] = true
		target, ok := expected[name]
		if !ok {
			v.reasons = append(v.reasons, fmt.Sprintf("xdg seed %s target is gone", name))
			continue
		}
		got, err := os.Readlink(filepath.Join(plan.parent, name))
		if err != nil || got != target {
			v.reasons = append(v.reasons, fmt.Sprintf("xdg seed %s is detached", name))
		}
	}
	for name := range expected {
		if !recorded[name] {
			// A newly present allowlisted entry is reconciliation
			// input for the next repair, not staleness: the home is
			// still exactly as recorded (§7.1).
			v.warnings = append(v.warnings, fmt.Sprintf("xdg seed %s newly present: repair to seed", name))
		}
	}
	v.warnings = append(v.warnings, shadowWarnings(req, plan, marker.SeedLinks)...)
}

// checkShadows reports declared shadowing paths. The surface is genuinely
// inert so the status row is non-current by default; resolution warns and
// leaves staleness to the recorded surfaces (§7.5).
func (v *verification) checkShadows(plan *homePlan, marker *envmarker.Marker) {
	_ = marker
	for _, shadow := range plan.adapter.Shadows {
		if _, err := os.Lstat(filepath.Join(plan.homeDir, filepath.FromSlash(shadow.Path))); err != nil {
			continue
		}
		v.warnings = append(v.warnings, fmt.Sprintf("%s: %s exists and makes %s inert", envregistry.DiagShadowingPresent, shadow.Path, shadow.Surface))
	}
}

// checkToolRelease warns when the detected release differs from the
// recorded one. An undetectable tool reports unknown, never matching.
func (v *verification) checkToolRelease(req *ResolveRequest, adapter envregistry.Adapter) {
	detected := req.detected(adapter)
	if detected == "unknown" || detected == "" {
		v.warnings = append(v.warnings, fmt.Sprintf("%s: %s detected release unknown, recorded %s", envregistry.DiagToolVersionUnverifed, adapter.ID, recordedRelease(adapter)))
		return
	}
	if adapter.VerifiedRelease != "" && detected != adapter.VerifiedRelease {
		if comparison, ok := envregistry.CompareVersions(detected, adapter.VerifiedRelease); !ok || comparison != 0 {
			v.warnings = append(v.warnings, fmt.Sprintf("%s: %s detected %s, recorded %s", envregistry.DiagToolVersionUnverifed, adapter.ID, detected, adapter.VerifiedRelease))
		}
	}
}

func recordedRelease(adapter envregistry.Adapter) string {
	if adapter.VerifiedRelease == "" {
		return "unrecorded"
	}
	return adapter.VerifiedRelease
}

// buildFragment assembles the launch fragment from a current home (§10.2):
// profile.lock_sha256, precedence, the effective profile permission mode,
// env, the system_prompt section exactly when the home carries the inert
// file, and the mcp
// section exactly when the home carries the channel file, with the sorted
// env_names union under the §10.3 double bound. The optional path_prepend
// member is not produced by this resolver.
func buildFragment(req *ResolveRequest, adapter envregistry.Adapter, verdict *verification) (*envfragment.Fragment, error) {
	plan := verdict.plan
	fragment := &envfragment.Fragment{
		Environment: adapter.ID,
		Profile:     req.Profile,
		LockSHA256:  strings.TrimPrefix(verdict.hash, "sha256:"),
		Winner:      req.Policy.Precedence().Winner,
		Placement:   req.Policy.Precedence().Placement,
		Env:         map[string]string{adapter.EnvVar: plan.parent},
	}
	if req.Machine.PermissionsLocked {
		fragment.Permissions = envfragment.Permissions{Mode: "native", Locked: true, Source: "global"}
	} else if mode, configured := req.Machine.Permissions[req.Profile]; configured {
		fragment.Permissions = envfragment.Permissions{Mode: mode, Source: "profile"}
	} else {
		fragment.Permissions = envfragment.Permissions{Mode: "native", Source: "default"}
	}
	if _, ok := plan.marker.Surfaces[envmarker.SurfaceSystemPrompt]; ok {
		fragment.SystemPrompt = &envfragment.SystemPrompt{
			Path:     filepath.Join(plan.homeDir, ".agent-context", "system-prompt.md"),
			Channels: adapter.SystemPrompt,
		}
	}
	if surface, ok := plan.marker.Surfaces[envmarker.SurfaceMCP]; ok && len(surface.Paths) == 1 {
		set, err := contextmaterialize.MCPSet(verdict.servers, adapter.ID)
		if err != nil {
			return nil, err
		}
		if adapter.MCP == nil {
			return nil, fmt.Errorf("managed home carries an mcp surface the %s adapter declares no channel for", adapter.ID)
		}
		// The §10.3 S4 bound, once per resolution: the active profile
		// bounds the requested names and warns — unlisted under
		// s4-warn, dropped under s4-enforce — without failing.
		passthrough := envfragment.ResolvePassthrough(
			contextmaterialize.MCPEnvNames(set),
			req.Machine.PassableEnvNames, req.Machine.PassableEnvNamesSet,
			envfragment.ActiveS4Profile)
		if passthrough.Warning != "" {
			verdict.warnings = append(verdict.warnings, passthrough.Warning)
		}
		fragment.MCP = &envfragment.MCP{
			Path:     filepath.Join(plan.homeDir, filepath.FromSlash(surface.Paths[0])),
			EnvNames: passthrough.Passed,
			Channels: []envregistry.Channel{*adapter.MCP},
		}
	}
	if err := envfragment.CheckBoundary(adapter, EnvRoot(req.Home), fragment); err != nil {
		return nil, err
	}
	return fragment, nil
}

// renderFragment renders the fragment in the requested format.
func renderFragment(fragment *envfragment.Fragment, adapter envregistry.Adapter, format string) ([]byte, error) {
	switch format {
	case "", "json":
		return fragment.JSON()
	case "env":
		return fragment.EnvFormat(adapter), nil
	case "shell":
		return fragment.ShellFormat(adapter), nil
	default:
		return nil, fmt.Errorf("format %q is not json, env, or shell", format)
	}
}

// firstResolveNotice accompanies the first provisioning and every first
// resolve of a home (§8.1): the managed-home path, the own-state-root
// statement, and the first-run steps the seeds do not cover. Informative,
// never suppressed by configuration.
func firstResolveNotice(adapter envregistry.Adapter, plan *homePlan) string {
	var steps string
	switch adapter.ID {
	case envregistry.ClaudeCode:
		steps = "Log in inside this home on first use; accept the project trust dialog on the first interactive launch; approve MCP servers when prompted."
	case envregistry.CodexCLI:
		steps = "Log in unless the native credential is shared through; pass --skip-git-repo-check outside a git repository."
	case envregistry.OpenCode:
		steps = "Apply the resolved fragment by hand (OPENCODE_CONFIG names the MCP file); skills come from the machine-current profile, split-brain by construction."
	case envregistry.Pi:
		steps = "No trust wall; authentication is shared from the native home."
	}
	return fmt.Sprintf("Managed home provisioned at %s.\nThe tool treats this home as its own state root: sessions, trust records, and approvals accrue here, not in the native home.\nFirst-run steps the seeds do not cover: %s", plan.homeDir, steps)
}

// loadResolveInputs reads the profile source and lock for resolution. A
// missing profile is profile_unknown (§10.4).
func loadResolveInputs(home, profile string) (Source, *contextlock.Lock, string, error) {
	return loadResolveInputsWith(home, profile, nil)
}

func loadResolveInputsWith(home, profile string, readRegularFile func(string) (stateread.File, error)) (Source, *contextlock.Lock, string, error) {
	if readRegularFile == nil {
		readRegularFile = stateread.ReadRegularFile
	}
	source, err := readSource(home, profile)
	if err != nil {
		if isStateAbsent(err) {
			return Source{}, nil, "", fmt.Errorf("%s: profile %q is not installed", DiagProfileUnknown, profile)
		}
		return Source{}, nil, "", fmt.Errorf("%s: profile %q source cannot be trusted: %w", envregistry.DiagStoreUntrusted, profile, err)
	}
	lock, hash, err := readLockWith(home, profile, readRegularFile)
	if err != nil {
		if isStateAbsent(err) {
			return Source{}, nil, "", fmt.Errorf("%s: profile %q is not installed", DiagProfileUnknown, profile)
		}
		return Source{}, nil, "", err
	}
	return source, lock, hash, nil
}

// currentProfileFor resolves the profile operand: the named profile, else
// the scoped current for the environment, else the machine current
// (§9.3, §10.1).
func currentProfileFor(home, envID, named string) (string, error) {
	if named != "" {
		return named, nil
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return "", fmt.Errorf("%s: read scoped current profile: %w", DiagProfileUnknown, err)
	}
	if profile, ok := scoped["env:"+envID]; ok && profile != "" {
		return profile, nil
	}
	current, err := Current(home)
	if err != nil {
		return "", fmt.Errorf("%s: read machine current profile: %w", DiagProfileUnknown, err)
	}
	if current == "" {
		return "", fmt.Errorf("%s: no profile is current", DiagProfileUnknown)
	}
	return current, nil
}

// Resolve verifies the managed home lock-free and, when current, emits the
// launch fragment without touching any lock. A stale home emits no
// fragment without repair (§10.1: environment_home_stale with the
// reasons); under repair the home is provisioned or repaired from the
// store under the mutation lock and the fragment is emitted.
func Resolve(req ResolveRequest) (*ResolveResult, error) {
	adapter, err := envregistry.ByID(req.EnvID)
	if err != nil {
		return nil, err
	}
	adapter, err = applyCodexSeedRevisionTestOverride(adapter, req.codexSeedRevisionForTest)
	if err != nil {
		return nil, err
	}
	profile, err := currentProfileFor(req.Home, req.EnvID, req.Profile)
	if err != nil {
		return nil, err
	}
	req.Profile = profile
	if !identifiers.Valid(profile) {
		return nil, fmt.Errorf("%s: profile %q is not installed", DiagProfileUnknown, profile)
	}
	managedRoot := EnvRoot(req.Home)
	homeRel, err := managedRelative(managedRoot, ManagedHomeDir(req.Home, profile, req.EnvID))
	if err != nil {
		return nil, err
	}
	if err := checkManagedPrivateTarget(managedRoot, managedJoin(homeRel, envmarker.Name)); err != nil {
		return nil, err
	}
	source, lock, hash, err := loadResolveInputsWith(req.Home, profile, req.readRegularFile)
	if err != nil {
		return nil, err
	}
	if err := validateProfilePathSources(profile, source, req.Policy); err != nil {
		return nil, err
	}
	verdict := verifyHome(&req, adapter, source, lock, hash)
	if verdict.markerReadFailed {
		return &ResolveResult{Warnings: verdict.warnings, StaleReasons: verdict.reasons}, fmt.Errorf("%s: %s", envmarker.DiagMarkerUnreadable, strings.Join(verdict.reasons, "; "))
	}
	if verdict.passthroughReadFailed {
		return &ResolveResult{Warnings: verdict.warnings, StaleReasons: verdict.reasons}, fmt.Errorf("%s: %s", envregistry.DiagPassthroughUnreadable, strings.Join(verdict.reasons, "; "))
	}
	if verdict.current {
		fragment, err := buildFragment(&req, adapter, verdict)
		if err != nil {
			return nil, err
		}
		document, err := renderFragment(fragment, adapter, req.Format)
		if err != nil {
			return nil, err
		}
		return &ResolveResult{Document: document, Warnings: verdict.warnings}, nil
	}
	if !req.Repair {
		return &ResolveResult{Warnings: verdict.warnings, StaleReasons: verdict.reasons}, fmt.Errorf("%s: %s", DiagHomeStale, strings.Join(verdict.reasons, "; "))
	}
	return repairUnderLock(&req, adapter, source, lock, hash)
}

// repairUnderLock takes the mutation lock with the bounded wait and a
// distinct lock-acquisition diagnostic, provisions or repairs the home
// from the store entries the lock names as one transaction, then emits
// the fragment (§10.1). A current home emits without touching any state;
// repair restores managed bytes from the store and never adopts candidate
// bytes found in the home.
func repairUnderLock(req *ResolveRequest, adapter envregistry.Adapter, source Source, lock *contextlock.Lock, hash string) (*ResolveResult, error) {
	op, err := beginOperation(req.Home, req.transactionOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()
	if err := validateProfilePathSources(req.Profile, source, req.Policy); err != nil {
		return nil, err
	}
	verdict := verifyHome(req, adapter, source, lock, hash)
	if verdict.markerReadFailed {
		return &ResolveResult{Warnings: verdict.warnings, StaleReasons: verdict.reasons}, fmt.Errorf("%s: %s", envmarker.DiagMarkerUnreadable, strings.Join(verdict.reasons, "; "))
	}
	if verdict.passthroughReadFailed {
		return &ResolveResult{Warnings: verdict.warnings, StaleReasons: verdict.reasons}, fmt.Errorf("%s: %s", envregistry.DiagPassthroughUnreadable, strings.Join(verdict.reasons, "; "))
	}
	if verdict.current {
		fragment, err := buildFragment(req, adapter, verdict)
		if err != nil {
			return nil, err
		}
		document, err := renderFragment(fragment, adapter, req.Format)
		if err != nil {
			return nil, err
		}
		return &ResolveResult{Document: document, Warnings: verdict.warnings}, nil
	}
	precedence := req.Policy.Precedence()
	order, err := contextmaterialize.EmittedOrder(lock, precedence)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	provisioned := verdict.marker == nil
	if provisioned {
		if err := checkProfileCollision(req.Home, req.Profile); err != nil {
			return nil, err
		}
	}
	plan, err := assembleHome(req, source, lock, hash, precedence, order, adapter, verdict.marker)
	if err != nil {
		return nil, err
	}
	seeds, err := req.gatherSeeds(adapter, provisioned)
	if err != nil {
		return nil, err
	}
	root := EnvRoot(req.Home)
	homeRel, err := managedRelative(root, plan.homeDir)
	if err != nil {
		return nil, err
	}
	writePaths := map[string]bool{}
	for path := range plan.copies {
		writePaths[path] = true
	}
	for path := range plan.links {
		writePaths[path] = true
	}
	for path := range seeds.files {
		writePaths[path] = true
	}
	if seeds.claudeInit {
		writePaths[".claude.json"] = true
	}
	if seeds.claudeInit {
		if err := checkManagedPrivateTarget(root, managedJoin(homeRel, ".claude.json")); err != nil {
			return nil, err
		}
	}
	passthroughLinks, _, err := req.effectivePassthrough(adapter, plan.isolation)
	if err != nil {
		if provisioned || isCredentialRefusal(err) {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	for path := range passthroughLinks {
		writePaths[path] = true
	}
	if err := preflightManagedWriteTargets(root, homeRel, writePaths); err != nil {
		return nil, err
	}
	if adapter.ID == envregistry.OpenCode {
		for name := range seeds.xdg {
			rel, err := managedRelative(root, filepath.Join(plan.parent, name))
			if err != nil {
				return nil, err
			}
			if _, err := managedPath(root, rel, false); err != nil {
				return nil, err
			}
		}
	}
	// The credential store must be established before the first write: an
	// unknown native selector fails here, not halfway through applyPlan
	// leaving a markerless partial home behind that the §9.5 inventory
	// would then refuse to provision over.
	if provisioned {
		want := map[string]bool{}
		for path := range plan.copies {
			want[path] = true
		}
		for path := range plan.links {
			want[path] = true
		}
		// Section 9.5 onboarding inventory: unmanaged files stop the
		// repair unless the carrying operation passes --takeover, which
		// backs every replaced file up before the first write (applyPlan
		// below, subject to environment_backup_exists) and reports the
		// replace notice.
		if err := inventoryUnmanaged(plan.homeDir, want, contextstore.Root(req.Home)); err != nil {
			if !req.Policy.Takeover {
				return nil, err
			}
			plan.warnings = append(plan.warnings, "taking over unmanaged files for "+adapter.ID+
				": native global context files are being replaced by managed ones; backups land in "+
				filepath.Join(plan.homeDir, ".agent-environment-backup")+"/")
		}
		if hint := foreignManagerHint(); hint != "" {
			plan.warnings = append(plan.warnings, fmt.Sprintf("%s: %s appears to manage this machine and will overwrite managed surfaces on its next apply", DiagForeignSuspect, hint))
		}
	}
	recorded := map[string]bool{}
	if verdict.marker != nil {
		for _, surface := range verdict.marker.Surfaces {
			for _, path := range surface.Paths {
				recorded[path] = true
			}
		}
	}
	if err := applyPlan(op, req, plan, seeds, recorded, verdict.marker, provisioned); err != nil {
		if provisioned {
			return nil, err
		}
		// A credential refusal is not a store failure: the repair
		// stops with the refusal diagnostic itself (§10.4), not
		// wrapped as environment_repair_failed.
		if isCredentialRefusal(err) {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %v", DiagRepairFailed, err)
	}
	after := verifyHome(req, adapter, source, lock, hash)
	if !after.current {
		return nil, fmt.Errorf("%s: %s", DiagRepairFailed, strings.Join(after.reasons, "; "))
	}
	fragment, err := buildFragment(req, adapter, after)
	if err != nil {
		return nil, err
	}
	document, err := renderFragment(fragment, adapter, req.Format)
	if err != nil {
		return nil, err
	}
	warnings := append([]string{}, after.warnings...)
	warnings = append(warnings, seeds.warnings...)
	result := &ResolveResult{Document: document, Warnings: warnings, Provisioned: provisioned}
	if provisioned {
		result.Notice = firstResolveNotice(adapter, plan)
	}
	return result, nil
}

// isCredentialRefusal reports whether a repair error is already a §7.7
// credential refusal, which surfaces as itself rather than wrapped in
// environment_repair_failed.
func isCredentialRefusal(err error) bool {
	return err != nil && (strings.Contains(err.Error(), envregistry.DiagCredentialConflict) ||
		strings.Contains(err.Error(), envregistry.DiagCredentialUnsupported))
}

// checkProfileCollision fails provisioning with environment_path_collision
// when two profile names map to one platform path below the environments
// root (§5).
func checkProfileCollision(home, profile string) error {
	listing, err := stateread.ReadDir(EnvRoot(home))
	if err != nil {
		return err
	}
	if listing.Kind == stateread.KindAbsent {
		return nil
	}
	if listing.Kind != stateread.KindPresent {
		return stateread.UnusableError(EnvRoot(home), fmt.Errorf("directory listing is unreadable"))
	}
	for _, entry := range listing.Entries {
		if entry.Name() != profile && strings.EqualFold(entry.Name(), profile) {
			return fmt.Errorf("%s: profile names %q and %q map to one platform path", contextmaterialize.DiagPathCollision, entry.Name(), profile)
		}
	}
	return nil
}

// copyLinkFallback materializes a file link as a copy when the platform
// takes no symlink, recording the manager §5 fallback reason.
func copyLinkFallback(root, fullRel, target, path string, plan *homePlan) error {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("link %s: %v", path, err)
	}
	if info.IsDir() {
		if err := copyTree(target, root, fullRel); err != nil {
			return fmt.Errorf("link %s: %v", path, err)
		}
	} else {
		payload, err := os.ReadFile(target) // #nosec G304 -- link target resolved from the store plan
		if err != nil {
			return fmt.Errorf("link %s: %v", path, err)
		}
		if err := atomicManagedFile(root, fullRel, payload, 0o644); err != nil {
			return err
		}
	}
	for key, surface := range plan.marker.Surfaces {
		for _, candidate := range surface.Paths {
			if candidate == path {
				copies := append([]envmarker.Copy{}, *surface.Copies...)
				copies = append(copies, envmarker.Copy{Path: path, Reason: envmarker.ReasonSymlinkFallback})
				surface.Copies = &copies
				plan.marker.Surfaces[key] = surface
			}
		}
	}
	delete(plan.links, path)
	return nil
}

// copyTree copies a directory tree for symlink-fallback copies.
func copyTree(source, root, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		outRel := destination
		if rel != "." {
			outRel = managedJoin(destination, filepath.ToSlash(rel))
		}
		if info.IsDir() {
			_, err := managedDirectory(root, outRel)
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("store tree contains a symlink at %s", path)
		}
		payload, err := os.ReadFile(path) // #nosec G304 -- store tree walked from the plan
		if err != nil {
			return err
		}
		return atomicManagedFile(root, outRel, payload, info.Mode().Perm())
	})
}

// systemPrompt assembles the system-prompt surface into managed homes only
// (§5.5): the inert .agent-context/system-prompt.md, linked from the
// store, plus the pi live channels machine configuration explicitly
// materializes (off by default, so a plain launch carries no active
// system prompt). The surface carries only admitted system modules (§3):
// under drop every skipped module warns context_system_module_dropped
// naming package and path, and under error the first skipped module
// refuses the assembly with context_system_module_transitive.
func (p *homePlan) systemPrompt(req *ResolveRequest, lock *contextlock.Lock, precedence contextmaterialize.Precedence, packages map[string]contextmaterialize.Package, surfaces map[string]envmarker.Surface) error {
	document, written, dropped, err := contextmaterialize.SystemPrompt(lock, precedence, p.adapter.ID, packages, req.Policy.Admission())
	if err != nil {
		return err
	}
	for _, module := range dropped {
		p.warnings = append(p.warnings, fmt.Sprintf("%s: package %s system module %s is transitive; skipped under the drop policy",
			contextmaterialize.DiagSystemModuleDropped, module.Package, module.Path))
	}
	if !written {
		return nil
	}
	const inert = ".agent-context/system-prompt.md"
	storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, inert)
	p.docs[storeDoc] = document
	p.links[inert] = storeDoc
	p.fileHashes[inert] = document
	paths := []string{inert}
	if p.adapter.ID == envregistry.Pi {
		switch req.Machine.SystemPromptFiles[req.Profile] {
		case "append":
			p.copies["APPEND_SYSTEM.md"] = document
			p.fileHashes["APPEND_SYSTEM.md"] = document
			paths = append(paths, "APPEND_SYSTEM.md")
		case "replace":
			p.copies["SYSTEM.md"] = document
			p.fileHashes["SYSTEM.md"] = document
			paths = append(paths, "SYSTEM.md")
		}
	}
	copies := []envmarker.Copy{}
	surfaces[envmarker.SurfaceSystemPrompt] = envmarker.Surface{
		Paths:         paths,
		ContentSHA256: contextmaterialize.SurfaceHash(p.fileHashesFor(paths)),
		Copies:        &copies,
	}
	return nil
}

// skills assembles the per-profile skills surface: one symlink per lock
// skill member into its immutable store entry (§7.1, §7.4 matrix). The
// surface hash binds each link to its entry's store content hash, so
// identical inputs hash identically on every machine (§5.6). opencode
// carries no skills surface: its skills target is the machine-global
// native surface, split-brain by construction (§7.1).
func (p *homePlan) skills(req *ResolveRequest, skills []skillOf, surfaces map[string]envmarker.Surface) {
	_ = req
	if p.adapter.SkillsDir == "" {
		return
	}
	hashes := map[string][]byte{}
	paths := []string{}
	copies := []envmarker.Copy{}
	for _, skill := range skills {
		path := p.adapter.SkillsDir + "/" + skill.name
		paths = append(paths, path)
		p.links[path] = skill.root
		hashes[path] = []byte(skill.hash)
	}
	surfaces[envmarker.SurfaceSkills] = envmarker.Surface{
		Paths:         paths,
		ContentSHA256: contextmaterialize.SurfaceHash(hashes),
		Copies:        &copies,
	}
}

// mcp assembles the §5.8 channel file surface: one inert file per
// adapter, managed homes only, with the sorted env_names union carried
// for the fragment.
func (p *homePlan) mcp(req *ResolveRequest, servers []contextmaterialize.MCPServer, surfaces map[string]envmarker.Surface) error {
	path, document, written, err := contextmaterialize.MCPFile(p.adapter.ID, servers)
	if err != nil {
		return err
	}
	if !written {
		return nil
	}
	storeDoc := storeDocPath(req.Home, req.Profile, p.adapter.ID, path)
	p.docs[storeDoc] = document
	p.links[path] = storeDoc
	p.fileHashes[path] = document
	copies := []envmarker.Copy{}
	surfaces[envmarker.SurfaceMCP] = envmarker.Surface{
		Paths:         []string{path},
		ContentSHA256: contextmaterialize.SurfaceHash(map[string][]byte{path: document}),
		Copies:        &copies,
	}
	return nil
}
