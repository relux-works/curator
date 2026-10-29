package envunit

import (
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/mcpnative"
)

// ValidateReadScope runs before any executor preimage digest. A target label
// alone cannot turn an arbitrary home/session/credential tree into owned data.
// Recursive images are confined to exact skill projection leaves (or their
// fixed versioned backup locations), independently of caller marker contents.
func (u *Unit) ValidateReadScope(class, path string) error {
	if class == "profile-mcp" {
		if u.Action != Maintain {
			return invalid("MCP publication must preserve installed authority")
		}
		for _, home := range u.Homes {
			if !bytesEqualHome(home.Before, home.After) {
				return invalid("MCP publication changes home authority")
			}
			if path == filepath.Join(home.Target, filepath.FromSlash(mcpnative.Path(home.Adapter))) && mcpnative.Path(home.Adapter) != "" {
				return nil
			}
			if path == filepath.Join(u.Manager.Path, "mcp-native-v1", "ownership", mcpnative.Key(home.Profile, home.Adapter, mcpnative.Path(home.Adapter))+".json") {
				return nil
			}
			adapter, err := envregistry.ByID(home.Adapter)
			if err != nil {
				return err
			}
			skillRoot := filepath.Join(home.Target, adapter.SkillsDir)
			if adapter.SkillsDir != "" && within(skillRoot, path) {
				rel, err := filepath.Rel(skillRoot, path)
				parts := strings.Split(filepath.ToSlash(rel), "/")
				if err == nil && len(parts) > 1 && identifier.MatchString(parts[0]) {
					return nil
				}
			}
		}
		for _, p := range u.Profiles {
			if p.OldLock.Digest == p.NewLock.Digest && path == filepath.Join(u.Manager.Path, "profiles", p.Name, "lock.json") {
				return nil
			}
		}
		return invalid("MCP target outside fixed settings/ownership/skill slots")
	}

	if class == "installed-profile" {
		if u.Action == Remove {
			for _, profile := range u.Profiles {
				if profile.OldLock.Kind == "file" && profile.NewLock.Kind == "absent" && path == filepath.Join(u.Manager.Path, "profiles", profile.Name) {
					return nil
				}
			}
		}
		return invalid("installed profile tree lacks exact removal binding")
	}
	// Credential identity is fixed by the adapter, independently of marker
	// surface labels. Check it before any ordinary byte/tree observation.
	for _, home := range u.Homes {
		if within(home.Target, path) && (class == "marker" && path != filepath.Join(home.Target, envmarker.Name) || class == "origin" && path != filepath.Join(home.Target, envorigin.Name)) {
			return invalid("authority class outside its fixed home leaf")
		}
		if credentialPath(home.Adapter, home.Target) == path && class != "credential-entry" {
			return invalid("credential leaf requires metadata-only credential-entry custody")
		}
	}
	for _, native := range u.Native {
		if credentialPath(native.Adapter, native.Root) == path {
			return invalid("native credential entry is outside publication custody")
		}
	}
	if class == "activation" && !ActivationPath(u.Manager.Path, path) {
		return invalid("activation outside fixed current namespace")
	}
	if legacyClass(class) {
		if !u.LegacyOwns(class, path) {
			return invalid("legacy target outside exact inventory")
		}
		return nil
	}
	if class == "owned-bytes" {
		for _, home := range u.Homes {
			if within(home.Target, path) && !managedOwnedLeaf(home.Adapter, home.Target, path) {
				return invalid("tool state is outside enumerated owned publication slots")
			}
		}
		for _, native := range u.Native {
			if within(native.Root, path) {
				rel, err := filepath.Rel(native.Root, path)
				if err != nil || !nativeSurface(native.Adapter, filepath.ToSlash(rel)) || rel == envmarker.Name {
					return invalid("native tool state is outside context/skill publication slots")
				}
			}
		}
	}
	if class != "owned-tree" {
		return nil
	}
	for _, home := range u.Homes {
		adapter, err := envregistry.ByID(home.Adapter)
		if err != nil {
			return err
		}
		dir := adapter.SkillsDir
		if dir == "" {
			dir = ".agent-context/skills"
		}
		if ownedSkillLeaf(home.Target, path, dir) {
			return nil
		}
	}
	for _, native := range u.Native {
		dir := "skills"
		if native.Adapter == "opencode" {
			dir = ".agent-context/skills"
		}
		if ownedSkillLeaf(native.Root, path, dir) {
			return nil
		}
	}
	return invalid("recursive preimage is outside an enumerated skill projection")
}

// ActivationPath is the closed current-pointer namespace. It permits both
// supported scope filename spellings; target-specific surface writes remain
// outside this native-origin unit's registered native adapters.
func ActivationPath(manager, path string) bool {
	root := filepath.Join(manager, "profiles")
	if path == filepath.Join(root, "current") {
		return true
	}
	if filepath.Dir(path) != filepath.Join(root, "scoped") {
		return false
	}
	for _, adapter := range envregistry.Registry {
		key := "env:" + adapter.ID
		if filepath.Base(path) == key || filepath.Base(path) == strings.ReplaceAll(key, ":", "%3A") {
			return true
		}
	}
	return false
}
func ownedSkillLeaf(root, path, dir string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, ".agent-environment-backup/") {
		parts := strings.SplitN(strings.TrimPrefix(rel, ".agent-environment-backup/"), "/", 2)
		if len(parts) != 2 || parts[0] == "" {
			return false
		}
		for _, c := range parts[0] {
			if c < '0' || c > '9' {
				return false
			}
		}
		rel = parts[1]
	}
	prefix := dir + "/"
	return strings.HasPrefix(rel, prefix) && identifier.MatchString(strings.TrimPrefix(rel, prefix))
}

func credentialPath(adapter, root string) string {
	switch adapter {
	case "codex_cli", "pi":
		return filepath.Join(root, "auth.json")
	case "claude_code":
		return filepath.Join(root, ".credentials.json")
	case "muse":
		return filepath.Join(root, "config", "muse", "auth.json")
	default:
		return ""
	}
}

// Managed byte custody is independent of marker-provided surface labels.
// Module projections are under the manager's fixed context directory; the
// adapter's unrelated tool/session state is never an ordinary target slot.
func managedOwnedLeaf(adapterID, root, path string) bool {
	a, err := envregistry.ByID(adapterID)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, ".agent-environment-backup/") {
		parts := strings.SplitN(strings.TrimPrefix(rel, ".agent-environment-backup/"), "/", 2)
		if len(parts) != 2 || parts[0] == "" {
			return false
		}
		for _, c := range parts[0] {
			if c < '0' || c > '9' {
				return false
			}
		}
		rel = parts[1]
	}
	if rel == a.RootTarget {
		return true
	}
	for _, seed := range a.Seeds {
		if rel == seed {
			return true
		}
	}
	if ownedSkillLeaf(root, path, a.SkillsDir) {
		return true
	}
	if a.SkillsDir == "" && ownedSkillLeaf(root, path, ".agent-context/skills") {
		return true
	}
	if adapterID != "muse" && rel == ".agent-context/system-prompt.md" {
		return true
	}
	switch adapterID {
	case "claude_code", "opencode":
		if rel == ".agent-context/mcp/"+adapterID+".json" {
			return true
		}
		if adapterID == "opencode" && rel == "opencode.json" {
			return true
		}
		const prefix = ".agent-context/modules/"
		if strings.HasPrefix(rel, prefix) {
			parts := strings.Split(strings.TrimPrefix(rel, prefix), "/")
			return len(parts) >= 2 && identifier.MatchString(parts[0]) && !strings.Contains(rel, "/../")
		}
	case "codex_cli":
		return rel == "curator-mcp.config.toml"
	case "pi":
		return rel == "SYSTEM.md" || rel == "APPEND_SYSTEM.md"
	}
	return false
}

func bytesEqualHome(a, b HomeImage) bool {
	return a.Kind == b.Kind && string(a.Marker) == string(b.Marker) && string(a.Origin) == string(b.Origin)
}
