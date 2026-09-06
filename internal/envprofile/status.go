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

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// SurfaceState is one recorded surface row.
type SurfaceState struct {
	Key    string
	Paths  []string
	Form   string
	State  string
	Detail string
}

// HomeState is one profile × environment row.
type HomeState struct {
	Profile        string
	Environment    string
	Mode           string
	Form           string
	Home           string
	Provisioned    bool
	Current        bool
	LockHash       string
	MarkerHash     string
	Surfaces       []SurfaceState
	Passthrough    []string
	Seeds          []string
	SeedLinks      []string
	SeededProjects []string
	Backups        int
	BackupsOldest  string
	BackupsNewest  string
	Findings       []string
	Warnings       []string
}

// ScopeHome carries both doors of a current profile (§8.1): the native
// home and the managed home with its provisioning state.
type ScopeHome struct {
	Scope       string
	Profile     string
	Environment string
	Native      string
	Managed     string
	Provisioned bool
}

// AdapterState carries the recorded and detected tool release per adapter
// (§7.9).
type AdapterState struct {
	ID        string
	Recorded  string
	Detected  string
	Supported []string
}

// TargetState carries one secondary-target row (§7.6, §12).
type TargetState struct {
	ID            string
	Adapter       string
	Participating bool
	Consented     bool
	Detail        string
	Ungoverned    string
}

// MemberState is one lock context member with its weight (§12).
type MemberState struct {
	Kind   string
	Name   string
	Weight int64
}

// PrecedenceState carries the precedence primitives per activation (§12).
type PrecedenceState struct {
	Winner    string
	Placement string
}

// ProfileState carries the lock's context members with weights and the
// precedence primitives per activation (§12).
type ProfileState struct {
	Profile    string
	LockHash   string
	Members    []MemberState
	Precedence PrecedenceState
}

// Status is the whole matrix.
type Status struct {
	Homes                    []HomeState
	Scopes                   []ScopeHome
	Adapters                 []AdapterState
	Targets                  []TargetState
	Profiles                 []ProfileState
	UnregisteredEnvironments []string
	Orphans                  []string
	Notes                    []string
	NonCurrent               bool
}

// StatusRequest scopes one status computation. The seams mirror
// ResolveRequest so tests pin homes without touching the process.
type StatusRequest struct {
	Home         string
	Machine      envregistry.MachineConfig
	Detect       func(envregistry.Adapter) string
	NativeHomeOf func(string) (string, error)
	OperatorXDG  string
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
	}
}

// StatusOf recomputes the profile × environment × surface matrix.
func StatusOf(req StatusRequest) (*Status, error) {
	status := &Status{}
	status.Notes = append(status.Notes, "opencode skills come from the machine-current profile, split-brain by construction (§7.1)")
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
	if len(status.Orphans) > 0 {
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
			state.Passthrough = append(state.Passthrough, entry.Path+" ("+entry.Strategy+")")
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
		if strings.HasPrefix(reason, "passthrough entry") && strings.HasSuffix(reason, "is detached") {
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
// profile. Precedence is the effective policy; revision 1 carries the
// default pair until manager-config schema 2 persists the knobs.
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
		out = append(out, ProfileState{
			Profile:  info.Name,
			LockHash: hash,
			Members:  members,
			Precedence: PrecedenceState{
				Winner:    contextmaterialize.DefaultPrecedence.Winner,
				Placement: contextmaterialize.DefaultPrecedence.Placement,
			},
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Profile < out[j].Profile })
	return out
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
