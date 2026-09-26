// Package envfragment builds and renders the launch-env-fragment-v2 object
// (environments §10.2): the closed fragment a launcher consumes, with
// profile.lock_sha256, precedence, permissions, env, system_prompt, mcp
// (path, sorted env_names union, channel descriptor), and optional
// path_prepend. Resolution launches nothing and applies no channel; the
// fragment is data about channels. The §10.3 profile-influence boundary is
// enforced at build: variable names come only from the closed adapter
// registry and every value stays an absolute path below the manager-owned
// environments root.
package envfragment

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Version is the fragment version emitted by this manager release.
const Version = "launch-env-fragment-v2"

// Permissions is the resolved per-profile permission policy carried by
// launch-env-fragment-v2 (environments §10.2).
type Permissions struct {
	Mode   string
	Locked bool
	Source string
}

// SystemPrompt carries the inert §5.5 file path and the adapter's §7.3
// channel descriptors. It is present exactly when the lock carries at
// least one applicable system module for the environment.
type SystemPrompt struct {
	Path     string
	Channels []envregistry.Channel
}

// MCP carries the §5.8 file path, the sorted env_names union, and the
// adapter's §7.8 channel descriptor without semantics. It is present
// exactly when the adapter's resolved MCP set is non-empty and the
// adapter declares a channel.
type MCP struct {
	Path     string
	EnvNames []string
	Channels []envregistry.Channel
}

// Fragment is one launch-env-fragment-v2 object. This resolver omits the
// optional path_prepend member (environments §10.2).
type Fragment struct {
	Environment  string
	Profile      string
	LockSHA256   string
	Permissions  Permissions
	Winner       string
	Placement    string
	Env          map[string]string
	SystemPrompt *SystemPrompt
	MCP          *MCP
}

// channelObject renders one descriptor in the §7.3 grammar: flag with
// flag, argument, name exactly when argument is name, and optional with;
// config-key with key; variable with variable; file with filename. A
// system-prompt descriptor carries semantics; an MCP descriptor does not.
func channelObject(channel envregistry.Channel, semantics bool) map[string]any {
	object := map[string]any{"kind": channel.Kind}
	if semantics {
		object["semantics"] = channel.Semantics
	}
	switch channel.Kind {
	case envregistry.KindFlag:
		object["flag"] = channel.Flag
		object["argument"] = channel.Argument
		if channel.Argument == envregistry.ArgumentName {
			object["name"] = channel.Name
		}
		if len(channel.With) > 0 {
			with := make([]any, 0, len(channel.With))
			for _, flag := range channel.With {
				with = append(with, flag)
			}
			object["with"] = with
		}
	case envregistry.KindConfigKey:
		object["key"] = channel.Key
	case envregistry.KindVariable:
		object["variable"] = channel.Variable
	case envregistry.KindFile:
		object["filename"] = channel.Filename
	}
	return object
}

// Object renders the fragment as a JSON-domain value for CCJ-1 encoding.
func (f *Fragment) Object() map[string]any {
	object := map[string]any{
		"fragment":    Version,
		"environment": f.Environment,
		"profile":     map[string]any{"name": f.Profile, "lock_sha256": f.LockSHA256},
		"permissions": map[string]any{"mode": f.Permissions.Mode, "locked": f.Permissions.Locked, "source": f.Permissions.Source},
		"precedence":  map[string]any{"winner": f.Winner, "placement": f.Placement},
	}
	env := map[string]any{}
	for name, value := range f.Env {
		env[name] = value
	}
	object["env"] = env
	if f.SystemPrompt != nil {
		channels := make([]any, 0, len(f.SystemPrompt.Channels))
		for _, channel := range f.SystemPrompt.Channels {
			channels = append(channels, channelObject(channel, true))
		}
		object["system_prompt"] = map[string]any{"path": f.SystemPrompt.Path, "channels": channels}
	}
	if f.MCP != nil {
		names := make([]any, 0, len(f.MCP.EnvNames))
		for _, name := range f.MCP.EnvNames {
			names = append(names, name)
		}
		channels := make([]any, 0, len(f.MCP.Channels))
		for _, channel := range f.MCP.Channels {
			channels = append(channels, channelObject(channel, false))
		}
		object["mcp"] = map[string]any{"path": f.MCP.Path, "env_names": names, "channels": channels}
	}
	return object
}

// Canonical renders the CCJ-1 bytes without the trailing LF, so that the
// works.relux.curator.fragment-digest extension key (Decision 0013 6.4) is
// sha256: over exactly these bytes (environments §10.1).
func (f *Fragment) Canonical() ([]byte, error) {
	return protocoljson.MarshalCanonical(f.Object())
}

// JSON renders --format json: the CCJ-1 bytes followed by exactly one LF.
func (f *Fragment) JSON() ([]byte, error) {
	canonical, err := f.Canonical()
	if err != nil {
		return nil, err
	}
	return append(canonical, '\n'), nil
}

// variables lists the fragment's env members in the adapter's declared
// variable order (environments §10.1): the adapter's EnvVar first, then
// any remaining names sorted. Only the env object is rendered: channel
// paths and env_names never enter these formats, so all three formats
// stay renderings of one object and resolving one activates nothing
// (§10.2, §10.3).
func (f *Fragment) variables(adapter envregistry.Adapter) [][2]string {
	var out [][2]string
	if value, ok := f.Env[adapter.EnvVar]; ok {
		out = append(out, [2]string{adapter.EnvVar, value})
	}
	rest := make([]string, 0, len(f.Env))
	for name := range f.Env {
		if name == adapter.EnvVar {
			continue
		}
		rest = append(rest, name)
	}
	sort.Strings(rest)
	for _, name := range rest {
		out = append(out, [2]string{name, f.Env[name]})
	}
	return out
}

// EnvFormat renders --format env: one NAME=value line per fragment env
// member, LF-terminated, in the adapter's declared variable order.
func (f *Fragment) EnvFormat(adapter envregistry.Adapter) []byte {
	var out strings.Builder
	for _, variable := range f.variables(adapter) {
		out.WriteString(variable[0])
		out.WriteByte('=')
		out.WriteString(variable[1])
		out.WriteByte('\n')
	}
	return []byte(out.String())
}

// ShellFormat renders --format shell: one POSIX export NAME='value' line
// per fragment env member with single-quote escaping. POSIX-only by design;
// automation on Windows or in PowerShell consumes --format json
// (environments §10.1).
func (f *Fragment) ShellFormat(adapter envregistry.Adapter) []byte {
	var out strings.Builder
	for _, variable := range f.variables(adapter) {
		out.WriteString("export ")
		out.WriteString(variable[0])
		out.WriteString("='")
		out.WriteString(strings.ReplaceAll(variable[1], "'", `'\''`))
		out.WriteString("'\n")
	}
	return []byte(out.String())
}

// S4 passthrough diagnostics (environments §2.1, §10.4).
const (
	// DiagPassthroughUnlisted warns, under s4-warn, for each passed
	// operator variable outside an explicitly configured
	// passable_env_names list, or for each passed variable when the knob
	// is absent. It names the variables and the knob and carries the
	// migration hint.
	DiagPassthroughUnlisted = "mcp_env_passthrough_unlisted" // #nosec G101 -- diagnostic code, not a credential
	// DiagPassthroughDropped warns, under s4-enforce, for each requested
	// name dropped from the launch allowlist. It names the variables.
	DiagPassthroughDropped = "mcp_env_passthrough_dropped" // #nosec G101 -- diagnostic code, not a credential
)

// S4Profile is one §10.3 warn-first passthrough profile: the warning
// release keeps the pre-S4 unbounded default with a warning, the flip
// release bounds an absent knob to the empty list.
type S4Profile string

const (
	// S4Warn is the warning release: an absent passable_env_names knob
	// behaves as unbounded, but every resolution that passes an
	// operator variable warns mcp_env_passthrough_unlisted.
	S4Warn S4Profile = "s4-warn"
	// S4Enforce is the flip release, the revision-1 rule: an absent
	// knob is the empty list, and unlisted requested names are dropped
	// with mcp_env_passthrough_dropped.
	S4Enforce S4Profile = "s4-enforce"
)

// ActiveS4Profile is the shipped profile: s4-warn first. A manager MUST
// ship s4-warn before s4-enforce — one release MUST NOT flip the default
// and start dropping in a single step — so the flip is a later release
// that moves this one constant.
const ActiveS4Profile = S4Warn

// MigrationHint is the §10.3 migration hint the unlisted warning carries.
const MigrationHint = "list the named variables to keep passing them after the flip"

// PassthroughVerdict is one S4 resolution: the names that reach the launch
// allowlist, the requested names outside the effective list, and the
// profile warning, if any. Reserved names are excluded before the bound
// and appear in neither list: the §2.2 exclusion is silent.
type PassthroughVerdict struct {
	Passed  []string
	Dropped []string
	Warning string
}

// EffectivePassable returns the effective passable_env_names allowlist for
// the profile: the explicitly configured list when the knob carries one,
// unbounded (nil) when the knob is explicitly null, and — when the knob
// is absent — unbounded under s4-warn, the empty list under s4-enforce.
// Nil always means unbounded, matching BoundEnvNames.
func EffectivePassable(passable []string, knobSet bool, profile S4Profile) []string {
	if knobSet {
		return passable
	}
	if profile == S4Enforce {
		return []string{}
	}
	return nil
}

// ResolvePassthrough applies the S4 profile's §10.3 bound to the
// requested names: the §2.2 reserved-name exclusion first, then the
// effective passable_env_names allowlist. knobSet distinguishes an absent
// knob (the profile default applies) from an explicit null (unbounded,
// silent under both profiles). The input must already be sorted; Passed
// and Dropped keep the requested order. Warning is "" when the profile
// is silent and otherwise starts with the diagnostic code.
func ResolvePassthrough(requested []string, passable []string, knobSet bool, profile S4Profile) PassthroughVerdict {
	effective := EffectivePassable(passable, knobSet, profile)
	verdict := PassthroughVerdict{Passed: []string{}, Dropped: []string{}}
	for _, name := range requested {
		if contextpkg.ReservedEnvName(name) {
			continue
		}
		if effective == nil {
			verdict.Passed = append(verdict.Passed, name)
			continue
		}
		allowed := false
		for _, candidate := range effective {
			if candidate == name {
				allowed = true
				break
			}
		}
		if allowed {
			verdict.Passed = append(verdict.Passed, name)
		} else {
			verdict.Dropped = append(verdict.Dropped, name)
		}
	}
	switch profile {
	case S4Enforce:
		if len(verdict.Dropped) > 0 {
			verdict.Warning = DiagPassthroughDropped + ": dropping " +
				strings.Join(verdict.Dropped, ", ") + " outside passable_env_names"
		}
	default:
		// Under s4-warn an explicitly configured list bounds silently
		// (unlisted requested names never reach Passed, so there is
		// nothing to warn about); an absent knob passes unbounded and
		// warns for each passed variable.
		if !knobSet && len(verdict.Passed) > 0 {
			verdict.Warning = DiagPassthroughUnlisted + ": passing " +
				strings.Join(verdict.Passed, ", ") +
				" with passable_env_names absent; " + MigrationHint
		}
	}
	return verdict
}

// BoundEnvNames applies the two bounds of §10.3 before the composer sees a
// name: the §2.2 reserved-name exclusion, then the lockable
// passable_env_names allowlist under the active S4 profile. A nil passable
// is read as an ABSENT knob — unbounded under s4-warn, empty under
// s4-enforce — so explicit-null callers must use ResolvePassthrough with
// knobSet. The input must already be sorted; the output keeps the order.
func BoundEnvNames(names []string, passable []string) []string {
	return ResolvePassthrough(names, passable, passable != nil, ActiveS4Profile).Passed
}

// CheckBoundary enforces the §10.3 profile-influence boundary on an
// assembled fragment: variable names come only from the closed adapter
// registry (the env member plus variable-kind channel names), and every
// path value — env values, system_prompt.path, mcp.path — is absolute,
// carries no .. segment, and stays below the manager-owned environments
// root. Only the profile-name path segment is profile-derived.
func CheckBoundary(adapter envregistry.Adapter, envRoot string, fragment *Fragment) error {
	allowed := map[string]bool{adapter.EnvVar: true}
	collect := func(channels []envregistry.Channel) {
		for _, channel := range channels {
			if channel.Kind == envregistry.KindVariable {
				allowed[channel.Variable] = true
			}
		}
	}
	if fragment.SystemPrompt != nil {
		collect(fragment.SystemPrompt.Channels)
	}
	if fragment.MCP != nil {
		collect(fragment.MCP.Channels)
	}
	checkPath := func(label, value string) error {
		if !filepath.IsAbs(value) {
			return fmt.Errorf("fragment %s %q is not absolute", label, value)
		}
		for _, segment := range strings.Split(filepath.ToSlash(value), "/") {
			if segment == ".." {
				return fmt.Errorf("fragment %s %q carries a .. segment", label, value)
			}
		}
		root := filepath.Clean(envRoot)
		if value != root && !strings.HasPrefix(value, root+string(filepath.Separator)) {
			return fmt.Errorf("fragment %s %q is outside the environments root", label, value)
		}
		return nil
	}
	for name, value := range fragment.Env {
		if !allowed[name] {
			return fmt.Errorf("fragment variable %q is not registry-declared for %s", name, adapter.ID)
		}
		if err := checkPath("env "+name, value); err != nil {
			return err
		}
	}
	if fragment.SystemPrompt != nil {
		if err := checkPath("system_prompt.path", fragment.SystemPrompt.Path); err != nil {
			return err
		}
	}
	if fragment.MCP != nil {
		if err := checkPath("mcp.path", fragment.MCP.Path); err != nil {
			return err
		}
	}
	return nil
}
