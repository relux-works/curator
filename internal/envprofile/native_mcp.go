package envprofile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/mcpnative"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
	"github.com/relux-works/curator/internal/whitelist"
)

// ProfileMCPRequest selects the separate first-install writer. It cannot issue
// or consume ownership-action authority. An explicit profile is mandatory.
type ProfileMCPRequest struct {
	Home, Profile, Scope, Mode, Operation string
	Environments                          []string
	DryRun                                bool
	Policy                                Policy
	Config                                *config.Config
	// Test-only vendor identity observation and transaction fault seams. No
	// production caller can replace qualification via a public input value.
	qualified          func(string) bool
	transactionOptions []transaction.Option
	beforePublish      func()
}

// ProfileMCPRow reports one profile server and environment pair.
type ProfileMCPRow struct {
	Name        string  `json:"name"`
	Environment string  `json:"environment"`
	Path        *string `json:"path"`
	State       string  `json:"state"`
	Code        string  `json:"code,omitempty"`
}

// ProfileMCPResult records first-install outcomes and read-only observations.
type ProfileMCPResult struct {
	Code             string          `json:"code,omitempty"`
	Capability       string          `json:"capability"`
	Profile          string          `json:"profile_id"`
	LockSHA256       string          `json:"lock_sha256"`
	DryRun           bool            `json:"dry_run"`
	Rows             []ProfileMCPRow `json:"rows"`
	RecoveryRequired bool            `json:"recovery_required"`
	Recovered        bool            `json:"recovered"`
	Published        bool            `json:"published"`
	Notice           string          `json:"notice,omitempty"`
	Warnings         []string        `json:"warnings"`
	IgnoredOptions   []string        `json:"ignored_options"`
}

func nativeMCPLedger(home, profile, environment string) string {
	return filepath.Join(home, "mcp-native-v1", "ownership", mcpnative.Key(profile, environment, mcpnative.Path(environment))+".json")
}
func nativeMCPQualifiedPath(environment string) string {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return ""
	}
	command, want := "", ""
	switch environment {
	case envregistry.ClaudeCode:
		command, want = "claude", "ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f"
	case envregistry.CodexCLI:
		command, want = "codex", "74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0"
	default:
		return ""
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return ""
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil || !nativeMCPExecutableMatches(path, want) {
		return ""
	}
	return path
}
func nativeMCPExecutableMatches(path, want string) bool {
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	sum := sha256.New()
	if _, err = io.Copy(sum, f); err != nil {
		return false
	}
	return hex.EncodeToString(sum.Sum(nil)) == want
}
func mcpAuthorityError() error {
	return fmt.Errorf("%s: protected ownership actions require the separate authority capability", mcpnative.AuthorityUnavailable)
}
func mcpRead(home, path string, private bool) (stateread.File, error) {
	if err := validateMCPRoute(home, path); err != nil {
		return stateread.File{}, fmt.Errorf("%s: unsafe MCP route: %w", mcpnative.Unreadable, err)
	}
	f, err := stateread.ReadRegularFile(path)
	if err != nil {
		return f, fmt.Errorf("%s: native MCP preimage is unreadable: %w", mcpnative.Unreadable, err)
	}
	if private {
		parent, e := stateread.Lstat(filepath.Dir(path))
		if e == nil && parent.Kind == stateread.KindPresent {
			e = privatedir.Validate(filepath.Dir(path))
		}
		if e == nil && f.Kind == stateread.KindPresent {
			e = privatedir.ValidateFile(path)
		}
		if e != nil {
			return f, fmt.Errorf("%s: protected MCP ownership is untrusted: %w", mcpnative.Unreadable, e)
		}
	}
	return f, nil
}

// ProfileMCP installs only absent entries. Status and preview never acquire a
// mutation lock, recover, fetch, execute a server or inspect native credentials.
func ProfileMCP(req ProfileMCPRequest) (result *ProfileMCPResult, resultErr error) {
	result = &ProfileMCPResult{Capability: mcpnative.Capability, Profile: req.Profile, DryRun: req.DryRun, Rows: []ProfileMCPRow{}, IgnoredOptions: []string{}, Warnings: []string{}}
	defer func() {
		if resultErr != nil {
			result.Code = profileMCPDiagnostic(resultErr)
			for i := range result.Rows {
				if result.Rows[i].State == "added" {
					result.Rows[i].State = "refused"
					result.Rows[i].Code = result.Code
				}
				if result.Rows[i].State == "refused" && result.Rows[i].Code == "mcp_contribution_policy_refused" {
					result.Rows[i].Code = result.Code
				}
			}
		}
		if req.DryRun {
			for i := range result.Rows {
				result.Rows[i].State = "would:" + result.Rows[i].State
			}
		}
	}()
	if err := envauthority.Validate(req.Home); err != nil {
		return result, err
	}
	if req.Scope != "profile" {
		return result, fmt.Errorf("mcp_contribution_policy_refused: project_scope_shared_home")
	}
	if !identifiers.Valid(req.Profile) {
		return result, fmt.Errorf("profile_unknown: explicit registered profile required")
	}
	if req.Mode != "install" && req.Mode != "status" && req.Mode != "repair" {
		return result, fmt.Errorf("invalid profile MCP mode")
	}
	// Explicit actions remain unavailable even if no pair could be discovered.
	if req.Operation != "" {
		return result, mcpAuthorityError()
	}
	inputs := map[string]string{}
	capture := func(path string) error {
		// Git prints forward slashes on Windows. Journal inputs require the
		// native absolute clean spelling, including for absent packed-refs.
		path, err := filepath.Abs(filepath.FromSlash(path))
		if err != nil {
			return err
		}
		digest, err := transaction.DigestPath(path)
		if err == nil {
			inputs[path] = digest
		}
		return err
	}
	if req.Config != nil {
		paths := []string{req.Config.Path, config.SystemCandidatePath()}
		for _, path := range paths {
			if path != "" {
				if err := capture(path); err != nil {
					return result, err
				}
			}
		}
		fresh, err := config.Load(req.Config.Path, nil)
		if err != nil {
			return result, err
		}
		if fresh.Home() != req.Home {
			return result, fmt.Errorf("profile_home_mismatch: current manager home changed")
		}
		req.Policy = PolicyFromConfig(fresh)
		// Refuse a changing system selection, including disappearance while loading.
		if config.SystemCandidatePath() != paths[1] {
			return result, fmt.Errorf("transaction_input_changed: system policy selection changed")
		}
	}
	if req.Mode == "status" || req.DryRun {
		inventory := filepath.Join(req.Home, "state", "transactions", "v1")
		if err := validateMCPRoute(req.Home, inventory); err != nil {
			return result, err
		}
		state, err := stateread.ReadDir(inventory)
		if err != nil {
			return result, err
		}
		result.RecoveryRequired = len(state.Entries) > 0
		if result.RecoveryRequired {
			result.Notice = "Pending transaction recovery required; read-only reporting did not recover."
		}
	}
	if len(req.Environments) == 0 {
		req.Environments = []string{envregistry.ClaudeCode, envregistry.CodexCLI}
	}
	seen := map[string]bool{}
	for _, id := range req.Environments {
		if _, err := envregistry.ByID(id); err != nil {
			return result, err
		}
		if seen[id] {
			return result, fmt.Errorf("duplicate MCP environment")
		}
		seen[id] = true
	}
	sort.Strings(req.Environments)
	validateInputs := func() error {
		if err := envauthority.Validate(req.Home); err != nil {
			return err
		}
		for path, want := range inputs {
			got, err := transaction.DigestPath(path)
			if err != nil || got != want {
				return fmt.Errorf("transaction_input_changed: first-install input changed")
			}
		}
		return nil
	}
	if err := validateInputs(); err != nil {
		return result, err
	}
	var op *operation
	if req.Mode != "status" && !req.DryRun {
		var err error
		op, err = beginOperationValidated(req.Home, validateInputs, req.transactionOptions...)
		if err != nil {
			return result, err
		}
		result.Recovered = op.recovered
		defer func() { resultErr = errorsJoinClose(resultErr, op) }()
	}
	if err := capture(lockPath(req.Home, req.Profile)); err != nil {
		return result, err
	}
	if err := capture(sourcePath(req.Home, req.Profile)); err != nil {
		return result, err
	}
	_, lock, hash, err := loadResolveInputs(req.Home, req.Profile)
	if err != nil {
		return result, err
	}
	result.LockSHA256 = hash
	if _, err = validateResolveRoots(req.Home, nil); err != nil {
		return result, err
	}
	defer func() {
		if resultErr != nil {
			present := map[string]bool{}
			for _, r := range result.Rows {
				present[r.Name+"/"+r.Environment] = true
			}
			for _, m := range lock.Members {
				if m.Kind != contextlock.KindMCP {
					continue
				}
				for _, id := range req.Environments {
					if present[m.Name+"/"+id] {
						continue
					}
					var p *string
					if path := mcpnative.Path(id); path != "" {
						p = &path
					}
					row := ProfileMCPRow{Name: m.Name, Environment: id, Path: p, State: "refused", Code: "mcp_contribution_policy_refused"}
					if id == "pi" || id == "opencode" {
						row.State = "unsupported"
						row.Code = ""
					}
					result.Rows = append(result.Rows, row)
				}
			}
			sortMCPRows(result.Rows)
		}
	}()
	if result.RecoveryRequired {
		return result, fmt.Errorf("%s: explicit recovery is required before pair inspection", DiagMaterializeRecoveryRequired)
	}
	manager := newGitManager(req.Home).withPolicy(req.Policy)
	manager.readOnlyAudit = true
	if failure := validateNamedStoreBoundaries(req.Home, lock, nil); failure != nil {
		return result, failure
	}
	if failure := validateNamedStorePins(req.Home, lock); failure != nil {
		return result, failure
	}
	if err = req.Policy.checkHardenedMCP(lock); err != nil {
		return result, err
	}
	if err = req.Policy.checkHardenedPassthrough(); err != nil {
		return result, err
	}
	servers := map[string]contextmaterialize.MCPServer{}
	serverHashes := map[string]string{}
	contentHashes := map[string]string{}
	for _, member := range lock.Members {
		resolved := resolvedOf(member)
		entry := manager.entryPath(req.Home, resolved)
		root := packageRoot(entry, member.Directory)
		if err = capture(entry); err != nil {
			return result, err
		}
		if member.Source != "" && !identity.Allowed(member.Source, req.Policy.AllowedSources) {
			return result, fmt.Errorf("profile_source_invalid: source outside current allowlist")
		}
		if _, err = strictAuditMember(req.Home, manager, resolved, entry, req.Policy); err != nil {
			return result, err
		}
		if member.Kind != contextlock.KindMCP {
			continue
		}
		if member.Source == "" || len(req.Policy.MCPAllowlist) > 0 && !identity.Allowed(member.Source, req.Policy.MCPAllowlist) {
			return result, fmt.Errorf("mcp_package_not_allowed: package source outside current allowlist")
		}
		if e := verifyNativeMCPSigner(manager, member, req.Policy, capture); e != nil {
			return result, e
		}
		manifest, e := projectmcp.LoadNamedPackage(root, member.Name)
		if e != nil {
			return result, e
		}
		raw, e := os.ReadFile(filepath.Join(root, contextpkg.MCPManifestName)) // #nosec G304 -- root is the verified package snapshot already checked by LoadNamedPackage.
		if e != nil {
			return result, e
		}
		var normalized map[string]any
		normalized, e = normalizedMCPManifest(raw)
		if e != nil {
			return result, e
		}
		serverHashes[member.Name], e = projectmcp.Digest(normalized)
		if e != nil {
			return result, e
		}
		contentHashes[member.Name], e = hashing.ContentSHA256WithVersion(root, nil, lock.ContentHashVersion())
		if e != nil {
			return result, e
		}
		passthrough := envfragmentMCPNames(manifest.Server.EnvNames, req.Policy)
		if !passthrough {
			return result, fmt.Errorf("mcp_contribution_policy_refused: environment-name ceiling")
		}
		servers[member.Name] = contextmaterialize.MCPServer{Name: member.Name, Transport: manifest.Server.Transport, Command: manifest.Server.Command, Args: manifest.Server.Args, URL: manifest.Server.URL, EnvNames: manifest.Server.EnvNames, Environments: manifest.Server.Environments}
	}
	planning := staging.Snapshot{Targets: map[string]staging.TargetSnapshot{}}
	writes := map[string][]byte{}
	preimages := map[string]string{}
	var first error
	refuse := func(e error) {
		if first == nil {
			first = e
		}
	}
	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, environment := range req.Environments {
		path := mcpnative.Path(environment)
		var ptr *string
		if path != "" {
			v := path
			ptr = &v
		}
		ownership := mcpnative.NewOwnership(req.Profile, environment)
		owned := map[string]mcpnative.Record{}
		ledgerPath := nativeMCPLedger(req.Home, req.Profile, environment)
		target := filepath.Join(ManagedHomeDir(req.Home, req.Profile, environment), filepath.FromSlash(path))
		if path != "" {
			early, err := (staging.Plan{Targets: []staging.Target{{LivePath: target}, {LivePath: ledgerPath}}}).Snapshot(nil)
			if err != nil {
				return result, err
			}
			for p, snapshot := range early.Targets {
				planning.Targets[p] = snapshot
			}
		}
		ledger, ledgerDigest, readErr := mcpObserve(req.Home, ledgerPath, true)
		if readErr == nil && ledger.Kind == stateread.KindPresent {
			ownership, readErr = mcpnative.DecodeOwnership(ledger.Bytes, req.Profile, environment)
			if readErr == nil {
				for _, r := range ownership.Records {
					owned[r.Name] = r
				}
			}
		}
		current := stateread.File{}
		currentDigest := transaction.DigestAbsent
		entries := map[string][]byte{}
		if path != "" && readErr == nil {
			current, currentDigest, readErr = mcpObserve(req.Home, target, false)
			if readErr == nil && current.Kind == stateread.KindPresent && environment == envregistry.ClaudeCode && len(current.Bytes) == 0 {
				readErr = fmt.Errorf("%s: empty JSON document", mcpnative.Unreadable)
			}
			if readErr == nil {
				entries, readErr = mcpnative.Entries(environment, current.Bytes)
			}
		}
		// A live Codex base is checked for A and B, independent of seed metadata.
		baseNames := map[string][]byte{}
		if environment == envregistry.CodexCLI && readErr == nil {
			basePath := filepath.Join(ManagedHomeDir(req.Home, req.Profile, environment), "config.toml")
			var base stateread.File
			var baseDigest string
			base, baseDigest, readErr = mcpObserve(req.Home, basePath, false)
			if readErr == nil {
				baseNames, readErr = mcpnative.Entries(environment, base.Bytes)
				if readErr == nil {
					inputs[basePath] = baseDigest
				}
			}
		}
		if len(baseNames) > 0 {
			result.Warnings = append(result.Warnings, "mcp_native_servers_ungoverned", "mcp_seed_unstripped")
		}
		set := append([]string{}, names...)
		for name := range owned {
			if _, ok := servers[name]; !ok {
				set = append(set, name)
			}
		}
		sort.Strings(set)
		data := append([]byte{}, current.Bytes...)
		changed := false
		for _, name := range set {
			server, declared := servers[name]
			record, hasMarker := owned[name]
			applies := declared && contextmaterialize.MCPApplies(server, environment)
			if !applies && !hasMarker {
				continue
			}
			row := ProfileMCPRow{Name: name, Environment: environment, Path: ptr}
			switch {
			case readErr != nil:
				row.State = "unreadable"
				row.Code = mcpnative.Unreadable
				refuse(readErr)
			case path == "" && environment != envregistry.Muse:
				row.State = "unsupported"
			case environment == envregistry.Muse || applies && server.Transport != "stdio":
				row.State = "refused"
				row.Code = "mcp_contribution_policy_refused"
				refuse(fmt.Errorf("mcp_contribution_policy_refused: native_writer_unqualified"))
			default:
				nativeBytes, exists := entries[name]
				_, baseConflict := baseNames[name]
				switch {
				case baseConflict || exists && !hasMarker:
					row.State = "conflict"
					refuse(fmt.Errorf("%s: foreign same-NAME entry", mcpnative.Conflict))
				case hasMarker && exists && nativeBytes == nil:
					row.State = "unreadable"
					row.Code = mcpnative.Unreadable
					refuse(fmt.Errorf("%s: owned native entry has no certified span", mcpnative.Unreadable))
				case hasMarker && !applies:
					row.State = "removed-declaration"
				case hasMarker && !exists:
					row.State = "missing"
				case hasMarker && mcpnative.Digest(nativeBytes) != record.InstalledSHA256:
					row.State = "drift"
				case hasMarker && (record.ServerSHA256 != serverHashes[name] || record.ContentSHA256 != contentHashes[name]):
					row.State = "replacement-required"
				case hasMarker:
					row.State = "unchanged"
				case req.Mode == "status":
					row.State = "missing"
				default:
					qualified := false
					if req.qualified != nil {
						qualified = req.qualified(environment)
					} else {
						executable := nativeMCPQualifiedPath(environment)
						if executable != "" {
							before, e := transaction.DigestPath(executable)
							if e == nil && nativeMCPQualifiedPath(environment) == executable {
								after, e := transaction.DigestPath(executable)
								if e == nil && before == after {
									inputs[executable] = before
									qualified = true
								}
							}
						}
					}
					if !qualified {
						row.State = "refused"
						row.Code = "mcp_contribution_policy_refused"
						refuse(fmt.Errorf("mcp_contribution_policy_refused: native_writer_unqualified"))
						break
					}

					next, entryBytes, e := mcpnative.Add(environment, data, server)
					if e != nil {
						row.State = "refused"
						row.Code = mcpnative.Unreadable
						refuse(e)
						break
					}
					data = next
					changed = true
					row.State = "added"
					ownership.Records = append(ownership.Records, mcpnative.Record{Name: name, Owner: "curator", ServerSHA256: serverHashes[name], ContentSHA256: contentHashes[name], InstalledSHA256: mcpnative.Digest(entryBytes)})
				}
			}
			if req.Mode != "status" && (row.State == "drift" || row.State == "missing" || row.State == "replacement-required" || row.State == "removed-declaration") {
				row.State = "refused"
				row.Code = mcpnative.AuthorityUnavailable
				refuse(mcpAuthorityError())
			}
			result.Rows = append(result.Rows, row)
		}
		if changed && first == nil {
			mcpnative.Sort(ownership)
			body, e := ownership.Bytes()
			if e != nil {
				return result, e
			}
			writes[target] = data
			writes[ledgerPath] = body
			preimages[target] = currentDigest
			preimages[ledgerPath] = ledgerDigest
		}
	}
	sortMCPRows(result.Rows)
	if first != nil {
		for i := range result.Rows {
			r := &result.Rows[i]
			if r.State == "added" {
				r.State = "refused"
				r.Code = "mcp_contribution_policy_refused"
			}
		}
		return result, first
	}
	if len(writes) == 0 || req.Mode == "status" {
		return result, nil
	}
	// Applicable skill payloads and the unchanged exact lock share this journal.
	skills, err := skillsOf(req.Home, manager, lock)
	if err != nil {
		return result, err
	}

	for _, skill := range skills {
		for _, id := range req.Environments {
			adapter, _ := envregistry.ByID(id)
			if adapter.SkillsDir == "" || mcpnative.Path(id) == "" {
				continue
			}
			destination := filepath.Join(ManagedHomeDir(req.Home, req.Profile, id), adapter.SkillsDir, skill.name)
			projected := map[string][]byte{}
			if err = stageMCPContextFiles(destination, skill, projected); err != nil {
				return result, err
			}
			state, e := stateread.Lstat(destination)
			if e != nil {
				return result, e
			}
			if state.Kind == stateread.KindPresent {
				// First install never changes existing skill bytes. Reuse a complete
				// identical projection; any changed or extra file needs materialization.
				if e = validateMCPRoute(req.Home, destination); e != nil {
					return result, e
				}
				actual, e := hashing.ContentSHA256WithVersion(destination, nil, skill.version)
				if e != nil || actual != skill.hash {
					return result, fmt.Errorf("%s: existing skill projection differs", mcpnative.AuthorityUnavailable)
				}
				for path, body := range projected {
					if e := privatedir.ValidateFile(path); e != nil {
						return result, fmt.Errorf("%s: existing skill preimage is untrusted: %w", mcpnative.AuthorityUnavailable, e)
					}
					observed, e := mcpRead(req.Home, path, false)
					if e != nil {
						return result, e
					}
					mode := os.FileMode(0o600)
					if runtime.GOOS == "windows" {
						// A writable Windows regular file reports 0666. Its
						// owner-only access was proved by the DACL check above.
						mode = 0o666
					}
					want, e := transaction.DigestRegularPayload(body, mode)
					if e != nil {
						return result, e
					}
					got, e := mcpObservedDigest(path, observed)
					if e != nil || got != want {
						return result, fmt.Errorf("%s: existing skill preimage differs", mcpnative.AuthorityUnavailable)
					}
					preimages[path] = got
				}
			} else {
				for path := range projected {
					preimages[path] = transaction.DigestAbsent
				}
			}
			for path, body := range projected {
				writes[path] = body
			}
		}
	}

	if req.DryRun {
		return result, nil
	}

	lockBytes, err := os.ReadFile(lockPath(req.Home, req.Profile))
	if err != nil {
		return result, err
	}
	writes[lockPath(req.Home, req.Profile)] = lockBytes
	if req.beforePublish != nil {
		req.beforePublish()
	}
	if err = validateInputs(); err != nil {
		return result, err
	}
	if err = publishNativeMCP(op, req, writes, preimages, inputs, planning, validateInputs); err != nil {
		for i := range result.Rows {
			if result.Rows[i].State == "added" {
				result.Rows[i].State = "refused"
				result.Rows[i].Code = "mcp_contribution_policy_refused"
			}
		}
		return result, err
	}
	result.Published = true
	result.Notice = "MCP configuration installed. Authentication not checked."
	return result, nil
}

func errorsJoinClose(err error, op *operation) error {
	closeErr := op.close()
	if err != nil {
		return err
	}
	return closeErr
}
func normalizedMCPManifest(raw []byte) (map[string]any, error) {
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	return projectmcp.NormalizeServer(object["server"])
}
func envfragmentMCPNames(names []string, policy Policy) bool {
	return len(envfragment.ResolvePassthrough(names, policy.PassableEnvNames, policy.PassableEnvNamesSet, envfragment.S4Enforce).Dropped) == 0
}

// Stage exactly the shared context projection, excluding executable command
// and runtime surfaces under the same rules as managed-home materialization.
func stageMCPContextFiles(destination string, skill skillOf, writes map[string][]byte) error {
	files := map[string][]byte{}
	err := whitelist.WalkContext(skill.source, skill.includeScripts, skill.excludeRoots, func(path, posix string) error {
		body, e := os.ReadFile(path) // #nosec G304 -- path is selected by the validated context whitelist walk.
		if e == nil {
			files[posix] = body
		}
		return e
	})
	if err != nil {
		return err
	}
	hash, err := hashing.ContentSHA256Files(files, skill.version)
	if err != nil {
		return err
	}
	if hash != skill.hash {
		return fmt.Errorf("transaction_input_changed: skill projection changed")
	}
	for rel, body := range files {
		writes[filepath.Join(destination, filepath.FromSlash(rel))] = body
	}
	return nil
}

func publishNativeMCP(op *operation, req ProfileMCPRequest, writes map[string][]byte, preimages, inputs map[string]string, planning staging.Snapshot, validate func() error) error {
	if err := validate(); err != nil {
		return err
	}
	// Private parents are created only after all pair/source refusals. Capture
	// their identities after creation and before staging. No native credential
	// path is inspected or copied by this writer.
	targets := staging.Plan{}
	for path := range writes {
		if err := validateMCPRoute(req.Home, path); err != nil {
			return err
		}
		if err := privatedir.MakeAll(filepath.Dir(path)); err != nil {
			return err
		}
		if err := validateMCPRoute(req.Home, path); err != nil {
			return err
		}
		if filepath.Dir(path) == filepath.Join(req.Home, "mcp-native-v1", "ownership") {
			if err := privatedir.Validate(filepath.Dir(path)); err != nil {
				return fmt.Errorf("%s: protected MCP ownership parent is untrusted: %w", mcpnative.Unreadable, err)
			}
		}
		targets.Targets = append(targets.Targets, staging.Target{LivePath: path, Kind: staging.KindBytes})
	}
	admitted := []string{}
	for path, digest := range inputs {
		if _, target := writes[path]; !target && digest != transaction.DigestAbsent {
			admitted = append(admitted, path)
		}
	}
	snapshot, err := targets.Snapshot(admitted)
	if err != nil {
		return err
	}
	// Preserve planning-time ancestor pins while adding the directories this
	// operation legitimately created for previously absent destinations.
	for path, earlier := range planning.Targets {
		current, ok := snapshot.Targets[path]
		if !ok {
			continue
		}
		if earlier.Canonical != current.Canonical {
			return fmt.Errorf("source_output_overlap: MCP destination retargeted")
		}
		for ancestor, pin := range earlier.Ancestors {
			current.Ancestors[ancestor] = pin
		}
		snapshot.Targets[path] = current
	}
	for _, target := range targets.Targets {
		if err := staging.RecheckOne(target, snapshot, admitted); err != nil {
			return err
		}
	}
	durable, err := snapshot.Durable()
	if err != nil {
		return err
	}
	// Unchanged targets (the exact lock) are guarded by target preimage/desired
	// digests, including while the live name is temporarily backed up.
	frozen := map[string]string{}
	for path, digest := range inputs {
		if _, target := writes[path]; !target {
			frozen[path] = digest
		}
	}
	recheck := func() error {
		for path, want := range frozen {
			got, e := transaction.DigestPath(path)
			if e != nil || got != want {
				return fmt.Errorf("transaction_input_changed: frozen first-install input")
			}
		}
		return nil
	}
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profile-mcp-first-install", ReadInputs: frozen, BoundaryProof: &durable}
	plan.BoundaryCheck = func(_ int, path string) error {
		if err := staging.RecheckOne(staging.Target{LivePath: path, Kind: staging.KindBytes}, snapshot, admitted); err != nil {
			return err
		}
		if err := validateMCPRoute(req.Home, path); err != nil {
			return err
		}
		if err := privatedir.Validate(filepath.Join(req.Home, "mcp-native-v1", "ownership")); err != nil {
			return fmt.Errorf("%s: protected MCP ownership parent changed: %w", mcpnative.Unreadable, err)
		}
		return recheck()
	}
	stage, err := privatedir.TempStaging(req.Home, ".mcp-stage-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	paths := make([]string, 0, len(writes))
	for path := range writes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		preimage, err := transaction.DigestPath(path)
		if err != nil {
			return err
		}
		if want, ok := preimages[path]; ok && want != preimage {
			return fmt.Errorf("transaction_input_changed: native preimage changed")
		}
		file, err := privatedir.CreateTemp(stage, "payload-*")
		if err != nil {
			return err
		}
		_, writeErr := file.Write(writes[path])
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		plan.Targets = append(plan.Targets, transaction.Target{PrivateFile: true, Class: "profile-mcp", Identifier: path, LivePath: path, StagedSource: file.Name(), PreimageDigest: preimage})
	}
	if err = validate(); err != nil {
		return err
	}
	if _, err = op.engine.Prepare(op.lock, plan); err != nil {
		return err
	}
	return op.engine.Commit(op.lock, plan.TransactionID)
}

func sortMCPRows(rows []ProfileMCPRow) {
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Environment < b.Environment
	})
}

// Bind targets to inspected bytes, rather than a later read that could already
// contain an external edit. The transaction verifies this digest again.
func mcpObservedDigest(path string, file stateread.File) (string, error) {
	if file.Kind == stateread.KindAbsent {
		return transaction.DigestAbsent, nil
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("transaction_input_changed: observed MCP file changed")
	}
	want, err := transaction.DigestRegularPayload(file.Bytes, info.Mode())
	if err != nil {
		return "", err
	}
	actual, err := transaction.DigestPath(path)
	if err != nil || actual != want {
		return "", fmt.Errorf("transaction_input_changed: observed MCP bytes changed")
	}
	return want, nil
}

// Observe bytes, type and mode coherently before deriving a write preimage.
// An absent base is frozen as absent before parsing, never from a later read.
func mcpObserve(home, path string, private bool) (stateread.File, string, error) {
	before, err := stateread.Lstat(path)
	if err != nil {
		return stateread.File{}, "", err
	}
	file, err := mcpRead(home, path, private)
	if err != nil {
		return file, "", err
	}
	after, err := stateread.Lstat(path)
	if err != nil {
		return file, "", err
	}
	if before.Kind != after.Kind || before.Kind == stateread.KindPresent && (!os.SameFile(before.Info, after.Info) || before.Info.Mode() != after.Info.Mode()) {
		return file, "", fmt.Errorf("transaction_input_changed: MCP file identity changed while reading")
	}
	digest, err := mcpObservedDigest(path, file)
	return file, digest, err
}

func profileMCPDiagnostic(err error) string {
	for _, code := range []string{envregistry.DiagCredentialInjectionUnavailable, mcpnative.AuthorityUnavailable, mcpnative.Conflict, mcpnative.Unreadable, "transaction_input_changed", DiagMaterializeRecoveryRequired, "profile_source_invalid", "mcp_package_not_allowed", "source_signers_missing", "mcp_contribution_policy_refused"} {
		if strings.Contains(err.Error(), code) {
			return code
		}
	}
	return "mcp_contribution_policy_refused"
}

// A profile lock retains a version and exact commit, rather than the selected
// tag spelling. Local verification accepts either standard version tag only
// when it still peels to that exact commit. It never fetches or reselects.
func verifyNativeMCPSigner(manager *gitManager, member contextlock.Member, policy Policy, capture func(string) error) error {
	pin := contextresolve.Requirement{Name: member.Name, Source: member.Source, Revision: member.Commit}
	original := contextresolve.VerifyMCPPin(localMCPSource{manager}, pin, policy.SourceSigners, policy.RequireSourceSigners)
	if original == nil || member.Version == "" {
		return original
	}
	for _, tag := range []string{"v" + member.Version, member.Version} {
		for _, ref := range []string{"refs/tags/" + tag, "packed-refs"} {
			output, err := exec.Command("git", "-C", manager.repoDir(member.Source), "rev-parse", "--path-format=absolute", "--git-path", ref).Output() // #nosec G204 -- fixed Git argv inspects a validated member ref without a shell.
			if err != nil {
				return original
			}
			path := strings.TrimSpace(string(output))
			if err = capture(path); err != nil {
				return err
			}
		}
		pin.Tag = tag
		if err := contextresolve.VerifyMCPPin(localMCPSource{manager}, pin, policy.SourceSigners, policy.RequireSourceSigners); err == nil {
			return nil
		}
	}
	return original
}

// The legacy writer only asks whether enrollment exists. Check each protected
// route node before descending, while retaining the existing operator-directory
// boundary scope. No ledger bytes, including malformed bytes, are adopted.
func refuseNativeMCPEnrollment(home, profile, environment string) error {
	ledger := nativeMCPLedger(home, profile, environment)
	for _, path := range []string{filepath.Join(home, "mcp-native-v1"), filepath.Dir(ledger), ledger} {
		state, err := stateread.Lstat(path)
		if err != nil {
			return fmt.Errorf("%s: ownership route unreadable: %w", mcpnative.Unreadable, err)
		}
		if state.Kind == stateread.KindAbsent {
			return nil
		}
		if err = pathboundary.ValidateLeafWithOwner(home, path, pathboundary.DefaultOwnerLookup()); err != nil {
			return fmt.Errorf("%s: unsafe ownership route: %w", mcpnative.Unreadable, err)
		}
		if path != ledger && !state.Info.IsDir() {
			return fmt.Errorf("%s: ownership ancestor is not a directory", mcpnative.Unreadable)
		}
	}
	return mcpAuthorityError()
}
