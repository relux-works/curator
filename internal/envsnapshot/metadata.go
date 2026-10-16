package envsnapshot

import (
	"reflect"
	"strings"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// Canonical caseless matching normalizes both before and after folding: full
// case folding can introduce decomposed sequences even from normalized input.
func collisionKey(s string) string {
	return norm.NFD.String(cases.Fold().String(norm.NFD.String(s)))
}

// validateMetadata returns the exact derived directory inventory. It performs
// semantic checks beyond the published structural schemas.
func validateMetadata(m Manifest, d Descriptor, launch []byte) (map[string]bool, error) {
	if m.Adapter != d.Environment || m.Profile != d.Profile {
		return nil, refuse("adapter-profile")
	}
	if m.LaunchSHA256 != Digest(launch) {
		return nil, refuse("launch-digest")
	}
	files := map[string]bool{}
	components := map[string]string{}
	dirs := map[string]bool{"home": true}
	previous := ""
	for _, f := range m.Files {
		if !identifiers.PortablePath(f.Path) {
			return nil, refuse("path-invalid")
		}
		if f.Path < previous {
			return nil, refuse("path-order")
		}
		previous = f.Path
		if files[f.Path] {
			return nil, refuse("path-collision")
		}
		files[f.Path] = true
		parts := strings.Split(f.Path, "/")
		for i := 1; i <= len(parts); i++ {
			prefix := strings.Join(parts[:i], "/")
			key := collisionKey(prefix)
			if old, ok := components[key]; ok && old != prefix {
				return nil, refuse("path-collision")
			}
			components[key] = prefix
			if i < len(parts) {
				dirs[prefix] = true
			}
		}
	}
	for f := range files {
		if dirs[f] {
			return nil, refuse("file-ancestor")
		}
	}
	if d.Environment == envregistry.Muse {
		for _, p := range []string{"config", "data", "state", "cache"} {
			dirs["home/"+p] = true
		}
		if (d.RootContext != nil) != files["home/config/muse/AGENTS.md"] {
			return nil, refuse("root-context")
		}
	}
	adapter, err := envregistry.ByID(d.Environment)
	if err != nil {
		return nil, refuse("adapter-profile")
	}
	if d.SystemPrompt != nil {
		if !files[d.SystemPrompt.Path] {
			return nil, refuse("channel-reference")
		}
		if !channelsEqual(d.SystemPrompt.Channels, adapter.SystemPrompt, true) {
			return nil, refuse("channel-semantics")
		}
	}
	if d.MCP != nil {
		if !files[d.MCP.Path] {
			return nil, refuse("channel-reference")
		}
		if adapter.MCP == nil || !channelsEqual(d.MCP.Channels, []envregistry.Channel{*adapter.MCP}, false) {
			return nil, refuse("channel-semantics")
		}
		for i, name := range d.MCP.EnvNames {
			if (i > 0 && name <= d.MCP.EnvNames[i-1]) || contextpkg.ReservedEnvName(name) {
				return nil, refuse("channel-semantics")
			}
		}
		// A named config layer must occupy the registered layer's fixed filename.
		for _, c := range d.MCP.Channels {
			if c.Kind == envregistry.KindFlag && c.Argument == envregistry.ArgumentName && d.MCP.Path != "home/"+c.Name+".config.toml" {
				return nil, refuse("channel-reference")
			}
		}
	}
	return dirs, nil
}

func channelsEqual(got []Channel, want []envregistry.Channel, semantics bool) bool {
	if len(got) != len(want) {
		return false
	}
	for i, c := range want {
		expected := Channel{Kind: c.Kind, Flag: c.Flag, Argument: c.Argument, Name: c.Name, With: c.With, Key: c.Key, Variable: c.Variable, Filename: c.Filename}
		if semantics {
			expected.Semantics = c.Semantics
		}
		if !reflect.DeepEqual(got[i], expected) {
			return false
		}
	}
	return true
}
