package scriptworker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/godriver"
)

// DeclaredCapabilities is the derivation input: the parsed declaration plus
// the presence of every top-level field in the declared manifest bytes.
// Presence matters because derivation reads the declared bytes, not the
// schema defaults: an absent field derives deny, so an absent `filesystem`
// must not inherit the declared-only `"repo"` default the parser applies.
type DeclaredCapabilities struct {
	// Parsed is the validated declaration. Fields Parse defaulted are
	// still visible here; Present tells which the manifest contained.
	Parsed capabilities.Manifest
	// Present names the top-level capability keys the manifest contained.
	Present map[string]bool
}

// ParseDeclaredCapabilities validates one raw `capabilities` object from
// declared manifest bytes. A nil or empty object means every field is
// absent and derives deny. Malformed bytes are refused fail-closed: the
// declaration is package-shaped input and must never widen the profile.
func ParseDeclaredCapabilities(raw json.RawMessage) (DeclaredCapabilities, error) {
	declared := DeclaredCapabilities{Present: map[string]bool{}}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		parsed, err := capabilities.Parse(nil)
		if err != nil {
			return DeclaredCapabilities{}, diagnostic(CodePackageInfluenceForbidden,
				"capability declaration is not usable: %v", err)
		}
		declared.Parsed = parsed
		return declared, nil
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return DeclaredCapabilities{}, diagnostic(CodePackageInfluenceForbidden,
			"capability declaration is not a JSON object: %v", err)
	}
	for key := range object {
		declared.Present[key] = true
	}
	parsed, err := capabilities.Parse(object)
	if err != nil {
		return DeclaredCapabilities{}, diagnostic(CodePackageInfluenceForbidden,
			"capability declaration is invalid: %v", err)
	}
	declared.Parsed = parsed
	return declared, nil
}

// DerivationReport is the manager's record of what one derivation decided.
// It carries identifiers and decisions only: secret identifiers never
// resolve to values, recorded hosts are reporting-only, and withheld names
// are reported rather than dropped silently.
type DerivationReport struct {
	// NetworkMode is "offline-environment" when the derived network
	// capability is none (absent or "none") and "recorded-hosts" when the
	// declaration lists hosts.
	NetworkMode string `json:"network_mode"`
	// RecordedHosts is the declared host list, recorded and reported but
	// never applied as a filter.
	RecordedHosts []string `json:"recorded_hosts"`
	// WithheldEnv names the env_read entries the manager withheld as
	// manager-owned, in sorted order.
	WithheldEnv []string `json:"withheld_env"`
	// ResolvedExec maps each declared exec name to its manager-resolved
	// fixed path.
	ResolvedExec map[string]string `json:"resolved_exec"`
	// UnresolvedExec names the declared exec names absent from the built
	// PATH, in sorted order.
	UnresolvedExec []string `json:"unresolved_exec"`
	// SecretIDs carries the declared secret identifiers, which remain
	// identifiers: derivation never resolves them to values.
	SecretIDs []string `json:"secret_ids"`
	// FarmEntries names the manager-owned PATH directory entries, sorted.
	FarmEntries []string `json:"farm_entries"`
}

// DerivedProfile is one derived containment profile: the environment and
// working directory the worker applies, plus the launch-boundary exec
// identities and the derivation record.
type DerivedProfile struct {
	// Environment is the complete manager-built session environment in
	// deterministic sorted order.
	Environment []string
	// Path is the manager-built PATH value: exactly the farm directory.
	Path string
	// WorkingDir is the manager-selected working directory.
	WorkingDir string
	// ProjectRoot is the canonical project root the derivation applied,
	// or "" when none was derived (global scope or non-repo filesystem).
	ProjectRoot string
	// WritePaths is the derived path set filesystem-write-confinement
	// grants on Linux: exactly the declared paths beneath the canonical
	// project root (the whole root for a `repo` declaration), in sorted
	// order. Empty when no path set was derived; the private area is
	// granted separately by the worker.
	WritePaths []string
	// NetworkOffline is true when the derived network capability is none.
	NetworkOffline bool
	// Exec is the launch-boundary identity of every resolved exec name.
	Exec []ExecIdentity
	// Report records the derivation decisions.
	Report DerivationReport
}

// DerivationInput is the complete manager-owned derivation context. The
// host environment supplies passthrough values only for non-reserved
// env_read names; everything else is manager-selected.
type DerivationInput struct {
	Declared      DeclaredCapabilities
	InterpreterID string
	Interpreter   InterpreterIdentity
	// HostEnv is the manager's own environment. Nil reads the process
	// environment, which is what production passes.
	HostEnv []string
	// ExecSearchDirs overrides the manager-owned exec search list. Nil
	// uses DefaultExecSearchDirs; tests inject fixture directories.
	ExecSearchDirs []string
	ForbiddenRoots []string
	Private        privateArea
	// FarmParent holds the manager-owned PATH directory derivation
	// creates. It must be the operation-private area.
	FarmParent  string
	ProjectRoot string
}

const (
	networkModeOffline  = "offline-environment"
	networkModeRecorded = "recorded-hosts"
)

// DeriveProfile derives the containment profile for one enforced
// invocation (Protocol Core §4.1.1, manager profile §3.1 step 4): a
// manager-built environment, a manager-built PATH over a manager-owned
// directory farm, offline network configuration, a manager-selected
// working directory, and operation-private root bindings. Every absent
// field derives its deny-by-default meaning, and package data can never
// widen the profile: reserved environment names never pass through,
// unresolved declared exec names stay absent from PATH and are reported,
// declared network hosts are recorded but never filter, secrets stay
// identifiers, and unknown interpreter identifiers are refused rather than
// launched with an unfiltered environment.
func DeriveProfile(input DerivationInput) (DerivedProfile, error) {
	return deriveProfileForPlatform(input, runtime.GOOS)
}

// deriveProfileForPlatform keeps the manager platform explicit for the
// default exec resolver tests. Production enters through DeriveProfile and
// supplies runtime.GOOS.
func deriveProfileForPlatform(input DerivationInput, platform string) (DerivedProfile, error) {
	if !closedInterpreters[input.InterpreterID] {
		return DerivedProfile{}, diagnostic(CodePackageInfluenceForbidden,
			"interpreter identifier %q has no reserved environment set", input.InterpreterID)
	}
	offline := len(input.Declared.Parsed.Network) == 0
	mode := networkModeOffline
	if !offline {
		mode = networkModeRecorded
	}

	execSearchDirs := input.ExecSearchDirs
	useDefaultExecSearchDirs := execSearchDirs == nil
	if useDefaultExecSearchDirs {
		execSearchDirs = defaultExecSearchDirs(platform, input.HostEnv)
	}
	execIdentities, resolved, unresolved, err := resolveDeclaredExecForPlatform(
		input.Declared.Parsed.Exec, execSearchDirs, input.ForbiddenRoots,
		platform, input.HostEnv, useDefaultExecSearchDirs)
	if err != nil {
		return DerivedProfile{}, err
	}
	farmDir, farmEntries, err := buildPathFarm(farmRequest{
		parent:      input.FarmParent,
		interpreter: input.Interpreter,
		exec:        execIdentities,
	})
	if err != nil {
		return DerivedProfile{}, err
	}
	workingDir, projectRoot, err := deriveWorkingDir(
		input.Declared, input.ProjectRoot, input.Private.tmp)
	if err != nil {
		return DerivedProfile{}, err
	}
	writePaths, err := deriveWritePaths(input.Declared, projectRoot)
	if err != nil {
		return DerivedProfile{}, err
	}
	environment, withheld, err := buildSessionEnvironment(environmentRequest{
		declared:      input.Declared,
		interpreterID: input.InterpreterID,
		hostEnv:       input.HostEnv,
		platform:      platform,
		interpreter:   input.Interpreter,
		private:       input.Private,
		farmDir:       farmDir,
		projectRoot:   projectRoot,
		offline:       offline,
	})
	if err != nil {
		return DerivedProfile{}, err
	}
	if err := validateDerivedEnvironment(environment, input.Private, farmDir, projectRoot, offline, platform); err != nil {
		return DerivedProfile{}, err
	}
	report := DerivationReport{
		NetworkMode:    mode,
		RecordedHosts:  append([]string(nil), input.Declared.Parsed.Network...),
		WithheldEnv:    withheld,
		ResolvedExec:   resolved,
		UnresolvedExec: unresolved,
		SecretIDs:      append([]string(nil), input.Declared.Parsed.Secrets...),
		FarmEntries:    farmEntries,
	}
	return DerivedProfile{
		Environment:    environment,
		Path:           farmDir,
		WorkingDir:     workingDir,
		ProjectRoot:    projectRoot,
		WritePaths:     writePaths,
		NetworkOffline: offline,
		Exec:           execIdentities,
		Report:         report,
	}, nil
}

// StaticReport is the install-time derivation record: the decisions that do
// not depend on per-invocation state (private roots, host presence). The
// installer reports it in its result messages so withheld names, recorded
// hosts, and unresolvable exec names are recorded rather than silent.
type StaticReport struct {
	NetworkMode    string
	RecordedHosts  []string
	WithheldEnv    []string
	ResolvedExec   map[string]string
	UnresolvedExec []string
	SecretIDs      []string
}

// DeriveStaticReport derives the install-time record for one enforced
// command. It resolves exec names through the same manager mechanism the
// invocation uses, but builds no farm and reads no host values.
func DeriveStaticReport(declared DeclaredCapabilities, interpreterID string, searchDirs, forbiddenRoots []string) (StaticReport, error) {
	if !closedInterpreters[interpreterID] {
		return StaticReport{}, diagnostic(CodePackageInfluenceForbidden,
			"interpreter identifier %q has no reserved environment set", interpreterID)
	}
	platform := runtime.GOOS
	withheld := withheldEnvNames(declared.Parsed.EnvRead, platform, interpreterID)
	_, resolved, unresolved, err := resolveDeclaredExec(
		declared.Parsed.Exec, searchDirs, forbiddenRoots)
	if err != nil {
		return StaticReport{}, err
	}
	mode := networkModeOffline
	if len(declared.Parsed.Network) != 0 {
		mode = networkModeRecorded
	}
	return StaticReport{
		NetworkMode:    mode,
		RecordedHosts:  append([]string(nil), declared.Parsed.Network...),
		WithheldEnv:    withheld,
		ResolvedExec:   resolved,
		UnresolvedExec: unresolved,
		SecretIDs:      append([]string(nil), declared.Parsed.Secrets...),
	}, nil
}

// resolveDeclaredExec resolves every declared exec name through the manager
// mechanism. Names resolve in sorted order so the farm and the report are
// deterministic. Unresolvable names are reported, never an error.
func resolveDeclaredExec(names []string, searchDirs, forbiddenRoots []string) ([]ExecIdentity, map[string]string, []string, error) {
	return resolveDeclaredExecForPlatform(names, searchDirs, forbiddenRoots,
		runtime.GOOS, nil, searchDirs == nil)
}

func resolveDeclaredExecForPlatform(names []string, searchDirs, forbiddenRoots []string, platform string, managerEnvironment []string, useDefaultSearchDirs bool) ([]ExecIdentity, map[string]string, []string, error) {
	ordered := append([]string(nil), names...)
	sort.Strings(ordered)
	if searchDirs == nil {
		searchDirs = defaultExecSearchDirs(platform, managerEnvironment)
	}
	var identities []ExecIdentity
	resolved := map[string]string{}
	var unresolved []string
	for _, name := range ordered {
		identity, found, err := resolveExecForPlatform(
			name, searchDirs, forbiddenRoots, platform, managerEnvironment, useDefaultSearchDirs)
		if err != nil {
			return nil, nil, nil, err
		}
		if !found {
			unresolved = append(unresolved, name)
			continue
		}
		identities = append(identities, identity)
		resolved[name] = identity.Path
	}
	return identities, resolved, unresolved, nil
}

// deriveWorkingDir selects the working directory. A repo or path-set
// filesystem derives the canonical project root of the invocation; every
// other shape — absent, home-config, or an empty path set — derives
// nothing, so the invocation works in its private temporary root. A global
// command has no project root, so a repo declaration there also derives
// nothing rather than refusing: absence of a derivable root denies by
// default. A project root that was derived but is no longer a directory is
// corrupt install state and refuses fail-closed.
func deriveWorkingDir(declared DeclaredCapabilities, projectRoot, privateTmp string) (string, string, error) {
	filesystem := declared.Parsed.Filesystem
	derivesRoot := declared.Present["filesystem"] &&
		(filesystem.Keyword == "repo" ||
			(filesystem.Keyword == "" && len(filesystem.Paths) != 0))
	if !derivesRoot || projectRoot == "" {
		return privateTmp, "", nil
	}
	info, err := os.Stat(projectRoot)
	if err != nil || !info.IsDir() {
		return "", "", diagnosticErr(CodeWorkerProtocolInvalid, err,
			"the derived project root is not a directory")
	}
	canonical, err := godriver.CanonicalPhysicalPath(projectRoot)
	if err != nil {
		return "", "", diagnosticErr(CodeWorkerProtocolInvalid, err,
			"the derived project root is not canonical")
	}
	return canonical, canonical, nil
}

// deriveWritePaths carries the derived path set Protocol Core §4.1.1
// promises: exactly the declared paths beneath the canonical project root
// — the whole root for a `repo` declaration. A path-set declaration
// derives exactly its members; absent, `home-config`, and empty shapes
// derive nothing. A global command has no project root, so nothing is
// derived there rather than refusing: absence of a derivable root denies
// by default. A member that escapes the root is corrupt derivation input
// and refuses fail-closed.
func deriveWritePaths(declared DeclaredCapabilities, projectRoot string) ([]string, error) {
	filesystem := declared.Parsed.Filesystem
	if !declared.Present["filesystem"] || projectRoot == "" {
		return nil, nil
	}
	switch filesystem.Keyword {
	case "repo":
		return []string{projectRoot}, nil
	case "home-config":
		return nil, nil
	case "":
		if len(filesystem.Paths) == 0 {
			return nil, nil
		}
		paths := make([]string, 0, len(filesystem.Paths))
		for _, member := range filesystem.Paths {
			absolute := member
			if !filepath.IsAbs(absolute) {
				absolute = filepath.Join(projectRoot, member)
			}
			absolute = filepath.Clean(absolute)
			if absolute != projectRoot && !isBelow(absolute, projectRoot) {
				return nil, diagnostic(CodePackageInfluenceForbidden,
					"capability filesystem path %q escapes the project root", member)
			}
			paths = append(paths, absolute)
		}
		sort.Strings(paths)
		return paths, nil
	default:
		return nil, diagnostic(CodePackageInfluenceForbidden,
			"capability filesystem keyword %q is not usable", filesystem.Keyword)
	}
}

// environmentRequest is the manager-owned context for one session
// environment.
type environmentRequest struct {
	declared      DeclaredCapabilities
	interpreterID string
	hostEnv       []string
	platform      string
	interpreter   InterpreterIdentity
	private       privateArea
	farmDir       string
	projectRoot   string
	offline       bool
}

// buildSessionEnvironment builds the manager environment from an empty
// bootstrap: manager-set values first, then exactly the non-reserved
// env_read names present in the host environment, then the offline
// proxy/resolver scrub when the derived network is none. The inherited
// environment is otherwise discarded, and every reserved name keeps its
// manager-set value. On Windows the loader essentials SYSTEMROOT and
// WINDIR are always manager-set (request host value, else the manager
// process value), so the child environment is closed and the runtime
// never injects one past the derived set.
func buildSessionEnvironment(request environmentRequest) ([]string, []string, error) {
	host, err := parseHostEnvironment(request.hostEnv, request.platform)
	if err != nil {
		return nil, nil, err
	}
	values := map[string]string{}
	set := func(key, value string) {
		if value == "" {
			return
		}
		values[key] = value
	}
	// Manager-set values. Every name the manager sets is reserved, so no
	// env_read entry can override one below.
	set("PATH", request.farmDir)
	set("HOME", request.private.base)
	set("TMPDIR", request.private.tmp)
	set("XDG_CONFIG_HOME", request.private.config)
	set("XDG_CACHE_HOME", request.private.cache)
	set("XDG_DATA_HOME", request.private.cache)
	set("XDG_STATE_HOME", request.private.cache)
	if request.platform == "windows" {
		set("TEMP", request.private.tmp)
		set("TMP", request.private.tmp)
		set("USERPROFILE", request.private.base)
		set("APPDATA", request.private.config)
		set("LOCALAPPDATA", request.private.cache)
		set("PATHEXT", windowsPathExtensions)
		for _, key := range []string{"SYSTEMROOT", "WINDIR", "COMSPEC"} {
			if value, present := host.lookup(key); present {
				set(key, value)
			}
		}
		// The Windows loader essentials are always manager-set. Go's
		// os/exec injects SYSTEMROOT from the calling process when the
		// explicit child environment lacks it, which would smuggle an
		// unmanaged value past the derived set. The request host value
		// wins; when the host does not carry one, the manager's own
		// process value stands (the same source the worker bootstrap
		// reads in indispensableScriptEnvironment). COMSPEC stays
		// host-carried only: nothing injects it.
		for _, key := range []string{"SYSTEMROOT", "WINDIR"} {
			if _, already := values[key]; already {
				continue
			}
			if value, present := os.LookupEnv(key); present {
				set(key, value)
			}
		}
	}
	if request.projectRoot != "" {
		set("CSK_PROJECT_ROOT", request.projectRoot)
	}
	var withheld []string
	for _, name := range request.declared.Parsed.EnvRead {
		if !validEnvName(name) {
			return nil, nil, diagnostic(CodePackageInfluenceForbidden,
				"env_read entry %q is not an environment variable name", name)
		}
		if reservedEnvironmentName(name, request.platform, request.interpreterID) {
			withheld = append(withheld, name)
			continue
		}
		if _, shadowed := values[name]; shadowed {
			withheld = append(withheld, name)
			continue
		}
		if value, present := host.lookup(name); present {
			values[name] = value
		}
	}
	if request.offline {
		// Offline scrub, defense in depth over the reserved filter above:
		// when the derived network is none, no proxy or resolver name
		// survives in the final environment even if a manager default
		// ever carried one.
		for key := range values {
			if proxyOrResolverName(key, request.platform) {
				delete(values, key)
			}
		}
	}
	environment := make([]string, 0, len(values))
	for key, value := range values {
		environment = append(environment, key+"="+value)
	}
	sort.Strings(environment)
	sort.Strings(withheld)
	return environment, withheld, nil
}

// windowsPathExtensions is the explicit bare-name extension list for the
// built interpreter environment. It mirrors the documented Windows default
// resolution the process would otherwise inherit implicitly, so the value
// is manager-set rather than host-dependent.
const windowsPathExtensions = ".COM;.EXE;.BAT;.CMD"

// withheldEnvNames reports the env_read entries derivation would withhold
// without reading host values: reserved names and manager-set names never
// pass through regardless of presence.
func withheldEnvNames(names []string, platform, interpreterID string) []string {
	var withheld []string
	for _, name := range names {
		if reservedEnvironmentName(name, platform, interpreterID) || managerSetName(name, platform) {
			withheld = append(withheld, name)
		}
	}
	sort.Strings(withheld)
	return withheld
}

// managerSetName reports whether derivation sets name itself. Every such
// name is reserved; the predicate exists so the install-time report agrees
// with the invocation without reading host values.
func managerSetName(name, platform string) bool {
	for _, managed := range []string{
		"PATH", "HOME", "TMPDIR",
		"XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME",
		"CSK_PROJECT_ROOT",
	} {
		if environmentNameEqual(name, managed, platform) {
			return true
		}
	}
	if platform != "windows" {
		return false
	}
	for _, managed := range []string{
		"TEMP", "TMP", "USERPROFILE", "APPDATA", "LOCALAPPDATA",
		"PATHEXT", "SYSTEMROOT", "WINDIR", "COMSPEC",
	} {
		if environmentNameEqual(name, managed, platform) {
			return true
		}
	}
	return false
}

// hostEnvironment is the manager's own environment indexed for derivation
// lookups. Later entries win, and on Windows keys fold case.
type hostEnvironment struct {
	platform string
	values   map[string]string
}

func parseHostEnvironment(environ []string, platform string) (hostEnvironment, error) {
	if environ == nil {
		environ = os.Environ()
	}
	host := hostEnvironment{platform: platform, values: map[string]string{}}
	for _, item := range environ {
		key, value, present := strings.Cut(item, "=")
		if !present || key == "" {
			return hostEnvironment{}, diagnostic(CodeWorkerProtocolInvalid,
				"the manager environment contains a malformed entry")
		}
		if strings.ContainsRune(item, 0) {
			return hostEnvironment{}, diagnostic(CodeWorkerProtocolInvalid,
				"the manager environment contains a NUL byte")
		}
		if platform == "windows" {
			key = strings.ToUpper(key)
		}
		host.values[key] = value
	}
	return host, nil
}

func (host hostEnvironment) lookup(name string) (string, bool) {
	if host.platform == "windows" {
		name = strings.ToUpper(name)
	}
	value, present := host.values[name]
	return value, present
}

// environmentNameEqual compares two environment names under the platform
// rule: case-insensitive on Windows, exact elsewhere.
func environmentNameEqual(one, two, platform string) bool {
	if platform == "windows" {
		return strings.EqualFold(one, two)
	}
	return one == two
}

// environmentNameHasPrefix reports whether name carries the reserved prefix
// under the platform rule.
func environmentNameHasPrefix(name, prefix, platform string) bool {
	if platform == "windows" {
		return len(name) >= len(prefix) && strings.EqualFold(name[:len(prefix)], prefix)
	}
	return strings.HasPrefix(name, prefix)
}

// reservedExactNames are manager-owned on every platform for every
// interpreter identifier. The proxy and resolver names live in the shared
// subset below instead of here.
var reservedExactNames = []string{
	"PATH", "HOME", "TMPDIR", "TEMP", "TMP",
	"XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME",
	"IFS", "CSK_PROJECT_ROOT",
}

// reservedWindowsNames are additionally manager-owned on Windows.
var reservedWindowsNames = []string{
	"USERPROFILE", "APPDATA", "LOCALAPPDATA", "PATHEXT", "COMSPEC", "WINDIR", "SYSTEMROOT",
}

// proxyOrResolverName reports whether name is one of the reserved proxy or
// resolver names the offline scrub removes. It is the shared subset the
// reserved filter and the scrub both enforce, so neither can drift from
// the other.
func proxyOrResolverName(name, platform string) bool {
	for _, reserved := range []string{
		"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "FTP_PROXY", "NO_PROXY",
		"http_proxy", "https_proxy", "all_proxy", "ftp_proxy", "no_proxy",
		"RES_OPTIONS", "HOSTALIASES", "LOCALDOMAIN",
	} {
		if environmentNameEqual(name, reserved, platform) {
			return true
		}
	}
	return false
}

// reservedEnvironmentName reports whether name is manager-owned under the
// profile's reserved enumeration: the portable set, the macOS DYLD prefix,
// the Windows set, and the per-interpreter loader prefixes. An unlisted
// name is not reserved: user-preference program selections such as SHELL
// or EDITOR pass through when named, because withholding them would change
// script behavior without containment benefit — a descendant that can name
// an absolute path gains nothing from them.
func reservedEnvironmentName(name, platform, interpreterID string) bool {
	if name == "" {
		return false
	}
	for _, reserved := range reservedExactNames {
		if environmentNameEqual(name, reserved, platform) {
			return true
		}
	}
	if proxyOrResolverName(name, platform) {
		return true
	}
	if environmentNameHasPrefix(name, "LD_", platform) {
		return true
	}
	if platform == "darwin" && environmentNameHasPrefix(name, "DYLD_", platform) {
		return true
	}
	if platform == "windows" {
		for _, reserved := range reservedWindowsNames {
			if environmentNameEqual(name, reserved, platform) {
				return true
			}
		}
	}
	switch interpreterID {
	case "python3-v1":
		if environmentNameHasPrefix(name, "PYTHON", platform) {
			return true
		}
		if environmentNameEqual(name, "__PYVENV_LAUNCHER__", platform) {
			return true
		}
	case "node-v1":
		if environmentNameHasPrefix(name, "NODE_", platform) {
			return true
		}
		if environmentNameHasPrefix(name, "NPM_CONFIG_", platform) {
			return true
		}
	}
	return false
}

// validEnvName reports whether name is a well-formed environment variable
// name. It mirrors the manifest parser's rule so hand-built declarations
// cannot smuggle a malformed entry past derivation.
func validEnvName(name string) bool {
	if name == "" {
		return false
	}
	for index, character := range name {
		alpha := (character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') || character == '_'
		digit := character >= '0' && character <= '9'
		if index == 0 && !alpha {
			return false
		}
		if !alpha && !digit {
			return false
		}
	}
	return true
}

// validateDerivedEnvironment checks the closed form of one derived session
// environment. The manager runs it as a self-check before the worker
// starts, and the worker runs the same function on the request it
// receives, so a manager-side derivation defect refuses at both layers
// instead of reaching the interpreter.
func validateDerivedEnvironment(environment []string, private privateArea, farmDir, projectRoot string, offline bool, platform string) error {
	values := map[string]string{}
	for _, item := range environment {
		key, value, present := strings.Cut(item, "=")
		if !present || key == "" || !validEnvName(key) {
			return diagnostic(CodeWorkerProtocolInvalid, "derived environment contains a malformed entry")
		}
		if strings.ContainsRune(item, 0) {
			return diagnostic(CodeWorkerProtocolInvalid, "derived environment contains a NUL byte")
		}
		folded := key
		if platform == "windows" {
			folded = strings.ToUpper(key)
		}
		if _, repeated := values[folded]; repeated {
			return diagnostic(CodeWorkerProtocolInvalid, "derived environment repeats %s", key)
		}
		values[folded] = value
	}
	lookup := func(name string) (string, bool) {
		if platform == "windows" {
			name = strings.ToUpper(name)
		}
		value, present := values[name]
		return value, present
	}
	// The manager-built PATH is exactly the farm directory.
	path, present := lookup("PATH")
	if !present || path == "" || !filepath.IsAbs(path) {
		return diagnostic(CodeWorkerProtocolInvalid, "derived environment PATH is not the absolute farm directory")
	}
	if platform == "windows" {
		if strings.Contains(path, ";") {
			return diagnostic(CodeWorkerProtocolInvalid, "derived environment PATH carries more than the farm directory")
		}
	} else if strings.Contains(path, ":") {
		return diagnostic(CodeWorkerProtocolInvalid, "derived environment PATH carries more than the farm directory")
	}
	if path != farmDir {
		return diagnostic(CodeWorkerProtocolInvalid, "derived environment PATH is not the derived farm directory")
	}
	// The private roots are bound through the platform environment.
	bindings := map[string]string{
		"TMPDIR":           private.tmp,
		"XDG_CONFIG_HOME":  private.config,
		"XDG_CACHE_HOME":   private.cache,
		"XDG_DATA_HOME":    private.cache,
		"XDG_STATE_HOME":   private.cache,
		"HOME":             private.base,
		"CSK_PROJECT_ROOT": projectRoot,
	}
	if platform == "windows" {
		bindings["TEMP"] = private.tmp
		bindings["TMP"] = private.tmp
		bindings["USERPROFILE"] = private.base
		bindings["APPDATA"] = private.config
		bindings["LOCALAPPDATA"] = private.cache
	}
	for key, want := range bindings {
		value, present := lookup(key)
		if want == "" {
			if present {
				return diagnostic(CodeWorkerProtocolInvalid,
					"derived environment sets %s without a derived value", key)
			}
			continue
		}
		if !present || value != want {
			return diagnostic(CodeWorkerProtocolInvalid,
				"derived environment does not bind %s to the derived root", key)
		}
	}
	if offline {
		for _, item := range environment {
			key, _, _ := strings.Cut(item, "=")
			if proxyOrResolverName(key, platform) {
				return diagnostic(CodeWorkerProtocolInvalid,
					"derived offline environment carries proxy or resolver configuration %s", key)
			}
		}
	}
	return nil
}
