// Package envfragment builds and renders the launch-env-fragment-v1 object
// (environments §10.2): the closed fragment a launcher consumes, with
// profile.lock_sha256, the precedence object, env, system_prompt, mcp (path,
// sorted env_names union, channel descriptor), and the reserved
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

// Version is the only fragment version this release emits.
const Version = "launch-env-fragment-v1"

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

// Fragment is one launch-env-fragment-v1 object. PathPrepend is reserved
// and never emitted in revision 1 (environments §10.2).
type Fragment struct {
	Environment  string
	Profile      string
	LockSHA256   string
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

// variables lists the fragment variables in the adapter's declared
// variable order (environments §10.1): the env member first, then every
// engaged variable-kind channel value.
func (f *Fragment) variables(adapter envregistry.Adapter) [][2]string {
	var out [][2]string
	names := make([]string, 0, len(f.Env))
	for name := range f.Env {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		out = append(out, [2]string{name, f.Env[name]})
	}
	if f.MCP != nil {
		for _, channel := range f.MCP.Channels {
			if channel.Kind == envregistry.KindVariable {
				out = append(out, [2]string{channel.Variable, f.MCP.Path})
			}
		}
	}
	if f.SystemPrompt != nil {
		for _, channel := range f.SystemPrompt.Channels {
			if channel.Kind == envregistry.KindVariable {
				out = append(out, [2]string{channel.Variable, f.SystemPrompt.Path})
			}
		}
	}
	_ = adapter
	return out
}

// EnvFormat renders --format env: one NAME=value line per fragment
// variable, LF-terminated, in the adapter's declared variable order.
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
// per variable with single-quote escaping. POSIX-only by design;
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

// BoundEnvNames applies the two bounds of §10.3 before the composer sees a
// name: the §2.2 reserved-name exclusion, then the lockable
// passable_env_names allowlist (nil means unbounded). The input must
// already be sorted; the output keeps the order.
func BoundEnvNames(names []string, passable []string) []string {
	out := []string{}
	for _, name := range names {
		if contextpkg.ReservedEnvName(name) {
			continue
		}
		if passable != nil {
			allowed := false
			for _, candidate := range passable {
				if candidate == name {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		out = append(out, name)
	}
	return out
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
