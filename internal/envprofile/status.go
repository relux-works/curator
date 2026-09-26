// Package envprofile status matrix (environments §12): the profile ×
// environment × surface matrix, read-only. Status recomputes and reports,
// never mutates
// — no fetch, no repair, no adoption, no channel application, no
// onboarding — and derives every row from the same lock-free verifier
// behind env resolve, so the two commands cannot disagree about currency.
// Warnings never make a row non-current (§12): size advisories, tool
// version skew, seed shadows, acknowledged shadowing paths, and the
// foreign-manager suspicion stay warnings.
package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hookapproval"
	"github.com/relux-works/curator/internal/manifest"
)

// SurfaceState is one recorded surface row.
type SurfaceState struct {
	Key    string   `json:"key"`
	Paths  []string `json:"paths"`
	Form   string   `json:"form"`
	State  string   `json:"state"`
	Detail string   `json:"detail"`
}

// HomeState is one profile × environment row.
type HomeState struct {
	Profile        string         `json:"profile"`
	Environment    string         `json:"environment"`
	Mode           string         `json:"mode"`
	Form           string         `json:"form"`
	Home           string         `json:"home"`
	Provisioned    bool           `json:"provisioned"`
	Current        bool           `json:"current"`
	LockHash       string         `json:"lock_hash"`
	MarkerHash     string         `json:"marker_hash"`
	Surfaces       []SurfaceState `json:"surfaces"`
	Passthrough    []string       `json:"passthrough"`
	Seeds          []string       `json:"seeds"`
	SeedLinks      []string       `json:"seed_links"`
	SeededProjects []string       `json:"seeded_projects"`
	Backups        int            `json:"backups"`
	BackupsOldest  string         `json:"backups_oldest"`
	BackupsNewest  string         `json:"backups_newest"`
	Findings       []string       `json:"findings"`
	Warnings       []string       `json:"warnings"`
}

// ScopeHome carries both doors of a current profile (§8.1): the native
// home and the managed home with its provisioning state.
type ScopeHome struct {
	Scope       string `json:"scope"`
	Profile     string `json:"profile"`
	Environment string `json:"environment"`
	Native      string `json:"native"`
	Managed     string `json:"managed"`
	Provisioned bool   `json:"provisioned"`
}

// AdapterState carries the recorded and detected tool release per adapter
// (§7.9).
type AdapterState struct {
	ID        string   `json:"id"`
	Recorded  string   `json:"recorded"`
	Detected  string   `json:"detected"`
	Supported []string `json:"supported"`
}

// TargetState carries one secondary-target row (§7.6, §12).
type TargetState struct {
	ID            string `json:"id"`
	Adapter       string `json:"adapter"`
	Participating bool   `json:"participating"`
	Consented     bool   `json:"consented"`
	Detail        string `json:"detail"`
	Ungoverned    string `json:"ungoverned"`
}

// MemberState is one lock context member with its weight (§12).
type MemberState struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Weight int64  `json:"weight"`
}

// PrecedenceState carries the precedence primitives per activation (§12).
type PrecedenceState struct {
	Winner    string `json:"winner"`
	Placement string `json:"placement"`
}

// ProfileState carries the lock's context members with weights and the
// precedence primitives per activation (§12).
type ProfileState struct {
	Profile    string          `json:"profile"`
	LockHash   string          `json:"lock_hash"`
	Members    []MemberState   `json:"members"`
	Precedence PrecedenceState `json:"precedence"`
	// TransitiveSystemModules is the effective §12.1 admission policy.
	TransitiveSystemModules string `json:"transitive_system_modules"`
	// DroppedSystemModules names every system module the drop policy
	// skips for the profile, in emitted order (§3, §12). It is empty
	// under the error policy, where nothing is skipped.
	DroppedSystemModules []DroppedSystemModule `json:"dropped_system_modules"`
}

// DroppedSystemModule names one system module the drop policy skips, by
// package and manifest path.
type DroppedSystemModule struct {
	Package string `json:"package"`
	Path    string `json:"path"`
}

// Provider verdicts for the §11/§12 umbrella provider posture rows: the
// trust verdict of one curator-<name> executable under the active
// revision.
const (
	// ProviderTrusted resolves inside the trust roots with no warning.
	ProviderTrusted = "trusted"
	// ProviderOutsideTrustRoots resolves outside the trust roots with
	// the revision-A warning and stays current.
	ProviderOutsideTrustRoots = "outside_trust_roots"
	// ProviderRefused is refused with subcommand_provider_untrusted.
	ProviderRefused = "refused"
	// ProviderMissing is absent with subcommand_provider_missing.
	ProviderMissing = "missing"
	// ProviderUnreadable failed with
	// subcommand_provider_root_unreadable.
	ProviderUnreadable = "unreadable"
)

// ProviderState is one §12 umbrella provider row: the resolved
// absolute provider path with its trust verdict. Resolved is set when
// the revision selected the provider; RefusedPath names the refused
// path on a refusal; Diagnostic names the closed §11.1 outcome when
// the row is not a silent success.
type ProviderState struct {
	Name                string   `json:"name"`
	Executable          string   `json:"executable"`
	Resolved            *string  `json:"resolved"`
	RefusedPath         *string  `json:"refused_path,omitempty"`
	Verdict             string   `json:"verdict"`
	Diagnostic          *string  `json:"diagnostic,omitempty"`
	Current             bool     `json:"current"`
	TrustRoots          []string `json:"trust_roots_consulted"`
	UnreadableDirectory *string  `json:"unreadable_directory,omitempty"`
}

// DeclarationScope repeats the §2.3 surfacing rows for the current
// profile of one reported scope, in row order without the LF.
type DeclarationScope struct {
	Scope   string   `json:"scope"`
	Profile string   `json:"profile"`
	Rows    []string `json:"rows"`
}

// Status is the whole matrix.
type Status struct {
	Homes                    []HomeState     `json:"homes"`
	Scopes                   []ScopeHome     `json:"scopes"`
	Adapters                 []AdapterState  `json:"adapters"`
	Targets                  []TargetState   `json:"targets"`
	Profiles                 []ProfileState  `json:"profiles"`
	Providers                []ProviderState `json:"providers"`
	UnregisteredEnvironments []string        `json:"unregistered_environments"`
	Orphans                  []string        `json:"orphans"`
	Notes                    []string        `json:"notes"`
	NonCurrent               bool            `json:"non_current"`
	// ShellHookTrust is the shell-hook trust posture (Manager profile
	// §8.6): one row per known project env file. A changed file, or a
	// recorded file whose bytes are missing or unreadable, makes the
	// matrix non-current; an unapproved file whose bytes read is a
	// warning row.
	ShellHookTrust []hookapproval.Posture `json:"shell_hook_trust"`
	// ShellHookTrustWarnings names malformed approval lines (skipped,
	// never repaired) or an unreadable approval state. The key is
	// present only when non-empty, so JSON consumers see the same read
	// failures the human output prints; an unreadable approval state
	// additionally makes the matrix non-current.
	ShellHookTrustWarnings []string `json:"shell_hook_trust_warnings,omitempty"`
	// S4Profile is the active §10.3 passthrough profile, and
	// PassableEnvNames the effective allowlist under it: nil renders
	// null and means unbounded (environments §12).
	S4Profile        string   `json:"s4_profile"`
	PassableEnvNames []string `json:"passable_env_names"`
	// Warnings carries machine-level warnings: the
	// mcp_package_allowlist_empty row when the MCP package allowlist is
	// empty, and unreadable-declaration notices. Warnings never make a
	// row non-current (§12).
	Warnings []string `json:"warnings"`
	// MCPDeclarations repeats the §2.3 surfacing rows for the current
	// profile of each scope reported.
	MCPDeclarations []DeclarationScope `json:"mcp_declarations"`
	// RequireCurrentProfile reports the locked require_current_profile
	// requirement (environments §12.2): when the system file locks the
	// key to a profile name, env status reports it. Nil means no
	// requirement.
	RequireCurrentProfile *string `json:"require_current_profile,omitempty"`
	// RequireCurrentLocked reports whether the requirement is locked.
	RequireCurrentLocked bool `json:"require_current_profile_locked,omitempty"`
}

// StatusRequest scopes one status computation. The seams mirror
// ResolveRequest so tests pin homes without touching the process.
type StatusRequest struct {
	Home         string
	Machine      envregistry.MachineConfig
	Detect       func(envregistry.Adapter) string
	NativeHomeOf func(string) (string, error)
	OperatorXDG  string
	// Policy carries the machine gates; the zero value reports the pair
	// default (environments §6, §12.1).
	Policy Policy
	// ProbeTarget reports whether a secondary target's probe path exists.
	// Nil means unprobed: auto participation finds nothing.
	ProbeTarget func(envregistry.Target) bool
	// LaunchDir is the directory project entries are reported against.
	LaunchDir string
}

func (req *StatusRequest) resolve() ResolveRequest {
	return ResolveRequest{
		Home:         req.Home,
		Machine:      req.Machine,
		Detect:       req.Detect,
		NativeHomeOf: req.NativeHomeOf,
		OperatorXDG:  req.OperatorXDG,
		LaunchDir:    req.LaunchDir,
		Policy:       req.Policy,
	}
}

// StatusOf recomputes the profile × environment × surface matrix.
func StatusOf(req StatusRequest) (*Status, error) {
	status := &Status{}
	status.Notes = append(status.Notes, "opencode skills come from the machine-current profile, split-brain by construction (§7.1)")
	// §12.2: env status reports the locked require_current_profile
	// requirement. The policy already carries the effective knob and its
	// locked bit from the loaded configuration.
	if req.Policy.RequireCurrent != nil {
		value := *req.Policy.RequireCurrent
		status.RequireCurrentProfile = &value
		status.RequireCurrentLocked = req.Policy.RequireCurrentLocked
	}
	infos, err := List(req.Home)
	if err != nil {
		return nil, err
	}
	installed := map[string]bool{}
	for _, info := range infos {
		installed[info.Name] = true
		for _, adapter := range envregistry.Registry {
			status.Homes = append(status.Homes, homeState(req, info.Name, adapter))
		}
	}
	for _, state := range status.Homes {
		if !state.Current {
			status.NonCurrent = true
		}
	}
	status.Scopes = scopeHomes(req, installed)
	status.Adapters = adapterStates(req)
	status.Targets = targetStates(req)
	status.Profiles = profileStates(req, infos)
	status.UnregisteredEnvironments = unregisteredEnvironments(req.Machine)
	status.Orphans = orphanHomes(req.Home, installed)
	// §12 posture: the active S4 profile with the effective
	// passable_env_names, the allowlist-empty warning row, and the
	// §2.3 surfacing rows for the current profile of each reported
	// scope. None of them affects currency.
	status.S4Profile = string(envfragment.ActiveS4Profile)
	status.PassableEnvNames = envfragment.EffectivePassable(
		req.Machine.PassableEnvNames, req.Machine.PassableEnvNamesSet,
		envfragment.ActiveS4Profile)
	if warning := AllowlistEmptyWarning(req.Policy.MCPAllowlist); warning != "" {
		status.Warnings = append(status.Warnings, warning)
	}
	declarations, declarationWarnings := declarationScopes(req, installed)
	status.MCPDeclarations = declarations
	status.Warnings = append(status.Warnings, declarationWarnings...)
	if len(status.Orphans) > 0 {
		status.NonCurrent = true
	}
	// Shell-hook trust posture (Manager profile §8.6): every recorded
	// path, plus the env files of the project the launch directory is
	// inside, if any. Read-only; a changed file, a recorded file whose
	// bytes are missing or unreadable, or an unreadable approval state is
	// non-current, while an unapproved file whose bytes read stays a
	// warning row.
	stateUnreadable := false
	status.ShellHookTrust, status.ShellHookTrustWarnings, stateUnreadable = hookapproval.AssessFailClosedDetailed(req.Home, trustProjectCandidates(req.LaunchDir))
	for _, row := range status.ShellHookTrust {
		if row.NonCurrent() {
			status.NonCurrent = true
			break
		}
	}
	if stateUnreadable {
		status.NonCurrent = true
	}
	sort.Slice(status.Homes, func(i, j int) bool {
		if status.Homes[i].Profile != status.Homes[j].Profile {
			return status.Homes[i].Profile < status.Homes[j].Profile
		}
		return status.Homes[i].Environment < status.Homes[j].Environment
	})
	sort.Strings(status.Orphans)
	return status, nil
}

// trustProjectCandidates returns the two hook-sourced env files of the
// project the launch directory is inside, or nothing when it is inside no
// project. An empty launch directory means the process working directory,
// matching the resolve verifier.
func trustProjectCandidates(launchDir string) []string {
	if launchDir == "" {
		var err error
		launchDir, err = os.Getwd()
		if err != nil {
			return nil
		}
	}
	root, ok := trustProjectRoot(launchDir)
	if !ok {
		return nil
	}
	return []string{
		filepath.Join(root, ".agents", "env.sh"),
		filepath.Join(root, ".agents", "env.ps1"),
	}
}

// trustProjectRoot searches upward from start for the nearest directory
// carrying Skillfile.json.
func trustProjectRoot(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, manifest.Name)); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// homeState verifies one profile × environment through the resolve
// verifier and projects the verdict onto a status row.
func homeState(req StatusRequest, profile string, adapter envregistry.Adapter) HomeState {
	resolve := req.resolve()
	resolve.Profile = profile
	resolve.EnvID = adapter.ID
	if resolve.LaunchDir == "" {
		resolve.LaunchDir, _ = os.Getwd()
	}
	state := HomeState{
		Profile:     profile,
		Environment: adapter.ID,
		Mode:        envmarker.ModeManagedHome,
		Home:        ManagedHomeDir(req.Home, profile, adapter.ID),
	}
	source, lock, hash, err := loadResolveInputs(req.Home, profile)
	if err != nil {
		state.Findings = append(state.Findings, DiagProfileUnknown+": "+err.Error())
		return state
	}
	state.LockHash = hash
	verdict := verifyHome(&resolve, adapter, source, lock, hash)
	state.Warnings = append(state.Warnings, verdict.warnings...)
	if verdict.marker == nil {
		state.Findings = append(state.Findings, verdict.reasons...)
		return state
	}
	state.Provisioned = true
	marker := verdict.marker
	state.Form = markerForm(marker)
	state.MarkerHash = marker.Profile.LockSHA256
	if verdict.plan != nil {
		for _, key := range marker.SortedSurfaceKeys() {
			surface := marker.Surfaces[key]
			detail := verdict.surfaceState[key]
			if detail == "" {
				detail = "current"
			}
			state.Surfaces = append(state.Surfaces, SurfaceState{
				Key: key, Paths: surface.Paths, Form: surface.Form,
				State: detail, Detail: detail,
			})
		}
	}
	if marker.Passthrough != nil {
		for _, entry := range *marker.Passthrough {
			if entry.Path == "" {
				state.Passthrough = append(state.Passthrough, entry.Strategy)
			} else {
				state.Passthrough = append(state.Passthrough, entry.Path+" ("+entry.Strategy+")")
			}
		}
	}
	if marker.Seeds != nil {
		state.Seeds = append([]string{}, *marker.Seeds...)
	}
	state.SeedLinks = append([]string{}, marker.SeedLinks...)
	state.SeededProjects = append([]string{}, marker.SeededProjects...)
	state.Backups, state.BackupsOldest, state.BackupsNewest = backupAges(state.Home)
	for _, reason := range verdict.reasons {
		diagnostic := reason
		if strings.HasPrefix(reason, "passthrough entry") && strings.Contains(reason, "is detached") {
			diagnostic = envregistry.DiagPassthroughDetached + ": " + reason
		}
		state.Findings = append(state.Findings, diagnostic)
	}
	// A declared shadowing path that exists is non-current by default and
	// a current warning under shadow_acknowledged (§7.5).
	if verdict.plan != nil {
		for _, shadow := range verdict.plan.adapter.Shadows {
			if _, err := os.Lstat(filepath.Join(verdict.plan.homeDir, filepath.FromSlash(shadow.Path))); err != nil {
				continue
			}
			if req.Machine.ShadowAcknowledges(adapter.ID, shadow.Path) {
				state.Warnings = append(state.Warnings, fmt.Sprintf("%s: %s is acknowledged and stays a warning", envregistry.DiagShadowingPresent, shadow.Path))
			} else {
				state.Findings = append(state.Findings, fmt.Sprintf("%s: %s exists and makes %s inert", envregistry.DiagShadowingPresent, shadow.Path, shadow.Surface))
			}
		}
	}
	state.Current = len(state.Findings) == 0
	return state
}

func markerForm(marker *envmarker.Marker) string {
	if surface, ok := marker.Surfaces[envmarker.SurfaceRootContext]; ok {
		return surface.Form
	}
	return ""
}

// backupAges counts the versioned backup generations beside the marker
// and reports the oldest and newest ages (§8.3, §12).
func backupAges(homeDir string) (int, string, string) {
	entries, err := os.ReadDir(filepath.Join(homeDir, ".agent-environment-backup"))
	if err != nil {
		return 0, "-", "-"
	}
	count := 0
	var oldest, newest time.Time
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		count++
		if oldest.IsZero() || info.ModTime().Before(oldest) {
			oldest = info.ModTime()
		}
		if newest.IsZero() || info.ModTime().After(newest) {
			newest = info.ModTime()
		}
	}
	if count == 0 {
		return 0, "-", "-"
	}
	now := time.Now()
	return count, ageString(now.Sub(oldest)), ageString(now.Sub(newest))
}

func ageString(duration time.Duration) string {
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	return fmt.Sprintf("%dd", int(duration.Hours()/24))
}

// profileStates reports the lock's context members with weights and the
// precedence primitives per activation (§12): one row per installed
// profile. Precedence is the effective machine policy.
func profileStates(req StatusRequest, infos []Info) []ProfileState {
	var out []ProfileState
	for _, info := range infos {
		_, lock, hash, err := loadResolveInputs(req.Home, info.Name)
		if err != nil || lock == nil {
			continue
		}
		members := make([]MemberState, 0, len(lock.Members))
		for _, member := range lock.Members {
			members = append(members, MemberState{Kind: member.Kind, Name: member.Name, Weight: member.Weight})
		}
		sort.Slice(members, func(i, j int) bool {
			if members[i].Kind != members[j].Kind {
				return members[i].Kind < members[j].Kind
			}
			return members[i].Name < members[j].Name
		})
		precedence := req.Policy.Precedence()
		out = append(out, ProfileState{
			Profile:  info.Name,
			LockHash: hash,
			Members:  members,
			Precedence: PrecedenceState{
				Winner:    precedence.Winner,
				Placement: precedence.Placement,
			},
			TransitiveSystemModules: effectiveTransitivePolicy(req.Policy),
			DroppedSystemModules:    droppedSystemModulesOf(req.Home, lock, req.Policy),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Profile < out[j].Profile })
	return out
}

// effectiveTransitivePolicy reports the policy value status prints: the
// configured knob, defaulting to drop exactly as resolution and
// materialization do.
func effectiveTransitivePolicy(policy Policy) string {
	if effective, err := policy.Admission().Policy(); err == nil {
		return effective
	}
	return contextmaterialize.TransitiveDrop
}

// droppedSystemModulesOf recomputes, read-only, the system modules the
// drop policy skips for the profile: every non-admitted system module in
// emitted order that applies to at least one registered adapter. Under
// the error policy nothing is skipped and the list is empty. A member
// whose manifest cannot be read contributes nothing here; the profile's
// home rows already report that same failure as non-current through the
// resolve verifier, so the gap is never silent.
func droppedSystemModulesOf(home string, lock *contextlock.Lock, policy Policy) []DroppedSystemModule {
	dropped := []DroppedSystemModule{}
	admission := policy.Admission()
	if effective, err := admission.Policy(); err != nil || effective != contextmaterialize.TransitiveDrop {
		return dropped
	}
	order, err := contextmaterialize.EmittedOrder(lock, policy.Precedence())
	if err != nil {
		return dropped
	}
	direct := contextmaterialize.DirectSet(lock)
	envIDs := registryEnvIDs()
	manager := newGitManager(home)
	for _, member := range order {
		if direct[member.Name] || admission.Waived(member.Name) {
			continue
		}
		manifest, err := contextpkg.LoadManifest(packageRoot(manager.entryPath(home, resolvedOf(member)), member.Directory))
		if err != nil {
			continue
		}
		for _, module := range manifest.Modules {
			class := module.Class
			if class == "" {
				class = "root"
			}
			if class != "system" {
				continue
			}
			applies := false
			for _, env := range envIDs {
				if module.Applies(env) {
					applies = true
					break
				}
			}
			if !applies {
				continue
			}
			dropped = append(dropped, DroppedSystemModule{Package: member.Name, Path: module.Path})
		}
	}
	return dropped
}

// unregisteredEnvironments reports env-ids named in machine configuration
// that the closed registry does not declare (§12).
func unregisteredEnvironments(machine envregistry.MachineConfig) []string {
	seen := map[string]bool{}
	var ids []string
	consider := func(id string) {
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	for id := range machine.Forms {
		consider(id)
	}
	for _, perProfile := range machine.Isolation {
		for id := range perProfile {
			consider(id)
		}
	}
	for id := range machine.InPlaceMode {
		consider(id)
	}
	for _, ack := range machine.ShadowAcknowledged {
		consider(ack.Env)
	}
	registered := map[string]bool{}
	for _, adapter := range envregistry.Registry {
		registered[adapter.ID] = true
	}
	var out []string
	for _, id := range ids {
		if !registered[id] {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// scopeHomes reports both homes of the current profile per scope.
func scopeHomes(req StatusRequest, installed map[string]bool) []ScopeHome {
	var out []ScopeHome
	machine, _ := Current(req.Home)
	scoped, _ := ScopedCurrents(req.Home)
	type scope struct{ name, profile string }
	scopes := []scope{{"machine", machine}}
	for key, profile := range scoped {
		scopes = append(scopes, scope{key, profile})
	}
	for _, item := range scopes {
		if item.profile == "" || !installed[item.profile] {
			continue
		}
		for _, adapter := range envregistry.Registry {
			native := ""
			if req.NativeHomeOf != nil {
				native, _ = req.NativeHomeOf(adapter.ID)
			} else if legacy, ok := adapterByID(adapter.ID); ok {
				native, _ = NativeHome(legacy)
			}
			managed := ManagedHomeDir(req.Home, item.profile, adapter.ID)
			provisioned := false
			if marker, err := envmarker.Read(managed); err == nil && marker != nil {
				provisioned = true
			}
			out = append(out, ScopeHome{
				Scope: item.name, Profile: item.profile,
				Environment: adapter.ID, Native: native,
				Managed: managed, Provisioned: provisioned,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return out[i].Environment < out[j].Environment
	})
	return out
}

// declarationScopes repeats the §2.3 surfacing rows for the current
// profile of each scope reported (environments §12): the machine scope
// first, then every scoped current in key order. A scope whose profile
// is not installed reports nothing; a scope with an empty MCP set
// reports no rows; an installed profile whose lock or declaration
// cannot be read is reported as unreadable, never as an empty set
// (§8.4).
func declarationScopes(req StatusRequest, installed map[string]bool) ([]DeclarationScope, []string) {
	machine, _ := Current(req.Home)
	scoped, _ := ScopedCurrents(req.Home)
	type scope struct{ name, profile string }
	scopes := []scope{{"machine", machine}}
	keys := make([]string, 0, len(scoped))
	for key := range scoped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		scopes = append(scopes, scope{key, scoped[key]})
	}
	manager := newGitManager(req.Home)
	var out []DeclarationScope
	var warnings []string
	for _, item := range scopes {
		if item.profile == "" || !installed[item.profile] {
			continue
		}
		lock, _, err := readLock(req.Home, item.profile)
		if err != nil || lock == nil {
			warnings = append(warnings, "profile "+item.profile+" lock cannot be read: "+readLockError(err))
			continue
		}
		rows, unreadable := surfacingRows(req.Home, manager, lock)
		for _, warning := range unreadable {
			warnings = append(warnings, item.profile+": "+warning)
		}
		if len(rows) == 0 {
			continue
		}
		out = append(out, DeclarationScope{Scope: item.name, Profile: item.profile, Rows: rows})
	}
	return out, warnings
}

// readLockError renders a lock-read failure without dereferencing nil.
func readLockError(err error) string {
	if err == nil {
		return "no lock"
	}
	return err.Error()
}

// adapterStates reports the recorded and detected tool release per
// adapter (§7.9). Detection is read-only; an unreadable tool reports
// unknown, never matching.
func adapterStates(req StatusRequest) []AdapterState {
	var out []AdapterState
	for _, adapter := range envregistry.Registry {
		var detected string
		if req.Detect != nil {
			detected = req.Detect(adapter)
		} else {
			detected = detectRelease(adapter.Probe)
		}
		if detected == "" {
			detected = "unknown"
		}
		recorded := adapter.VerifiedRelease
		if recorded == "" {
			recorded = "unrecorded"
		}
		out = append(out, AdapterState{
			ID: adapter.ID, Recorded: recorded, Detected: detected,
			Supported: append([]string{}, adapter.Forms...),
		})
	}
	return out
}

// targetStates evaluates secondary-target participation with the standing
// ungoverned note (§7.6, §12).
func targetStates(req StatusRequest) []TargetState {
	var out []TargetState
	for _, target := range envregistry.Targets {
		probe := false
		if req.ProbeTarget != nil {
			probe = req.ProbeTarget(target)
		}
		participating := req.Machine.TargetParticipates(target, func(string) bool { return probe })
		detail := "auto: probe path absent, nothing materialized"
		if req.Machine.TargetParticipation[target.ID] == "enabled" {
			detail = "explicitly enabled"
		} else if probe {
			detail = "auto: probe path exists"
		} else if req.Machine.TargetParticipation[target.ID] == "off" {
			detail = "off"
		}
		out = append(out, TargetState{
			ID: target.ID, Adapter: target.Adapter, Participating: participating,
			Consented: req.Machine.TargetConsented[target.ID] || req.Machine.TargetParticipation[target.ID] == "enabled",
			Detail:    detail, Ungoverned: target.Ungoverned,
		})
	}
	return out
}

// orphanHomes reports managed homes whose profile is no longer installed
// (§9.2): retained homes without a profile, removable by a later --purge.
func orphanHomes(home string, installed map[string]bool) []string {
	var out []string
	entries, err := os.ReadDir(EnvRoot(home))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() || seen[entry.Name()] {
			continue
		}
		seen[entry.Name()] = true
		for _, adapter := range envregistry.Registry {
			managed := ManagedHomeDir(home, entry.Name(), adapter.ID)
			marker, err := envmarker.Read(managed)
			if err != nil || marker == nil {
				continue
			}
			if !installed[marker.Profile.Name] {
				out = append(out, managed)
			}
		}
	}
	return out
}
