package envunit

import (
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
)

// Native describes only explicitly owned in-place surfaces. It is independent
// of the managed home's credential origin and never inventories a native home.
// Marker bytes and directory entry identities survive loss of callbacks.
type Native struct {
	Profile         string   `json:"profile"`
	Adapter         string   `json:"adapter"`
	Root            string   `json:"root"`
	DirectoryBefore Image    `json:"directory_before"`
	DirectoryAfter  Image    `json:"directory_after"`
	MarkerBefore    []byte   `json:"marker_before,omitempty"`
	MarkerAfter     []byte   `json:"marker_after"`
	Paths           []string `json:"paths"`
}

func (u *Unit) validateNative(profiles map[string]ProfileTransition) error {
	seen := map[string]bool{}
	for _, n := range u.Native {
		p, ok := profiles[n.Profile]
		if !ok || !absolute(n.Root) || within(u.Manager.Path, n.Root) || seen[n.Root] || len(n.Paths) == 0 || len(n.Paths) > 4096 {
			return invalid("native custody binding")
		}
		seen[n.Root] = true
		if n.Adapter != "codex_cli" && n.Adapter != "claude_code" && n.Adapter != "pi" && n.Adapter != "opencode" {
			return invalid("native custody adapter")
		}
		if err := n.DirectoryBefore.validate(false); err != nil {
			return err
		}
		if err := n.DirectoryAfter.validate(false); err != nil {
			return err
		}
		if n.DirectoryBefore.Kind != "absent" && n.DirectoryBefore.Kind != "directory" || n.DirectoryAfter.Kind != "directory" {
			return invalid("native directory type")
		}
		after, err := envmarker.Parse(n.MarkerAfter)
		if err != nil {
			return invalid("native marker grammar: %v", err)
		}
		if after.Mode != envmarker.ModeLinked || after.Profile.Name != n.Profile || strings.TrimPrefix(after.Profile.LockSHA256, "sha256:") != strings.TrimPrefix(p.NewLock.Digest, "sha256:") {
			return invalid("native marker/lock binding")
		}
		paths := map[string]bool{}
		for _, rel := range n.Paths {
			if rel == "" || filepath.IsAbs(rel) || filepath.ToSlash(filepath.Clean(filepath.FromSlash(rel))) != rel || rel == ".." || strings.HasPrefix(rel, "../") || paths[rel] {
				return invalid("native owned path grammar")
			}
			paths[rel] = true
			if !nativeSurface(n.Adapter, rel) {
				return invalid("native target is not a context/skill surface")
			}
		}
		if !paths[envmarker.Name] {
			return invalid("native marker omitted from target inventory")
		}
	}
	return nil
}
func nativeSurface(adapter, rel string) bool {
	if rel == envmarker.Name {
		return true
	}
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
		return nativeSurface(adapter, parts[1]) && parts[1] != envmarker.Name && !strings.HasPrefix(parts[1], ".agent-environment-backup/")
	}
	if adapter == "claude_code" && rel == "CLAUDE.md" || adapter != "claude_code" && rel == "AGENTS.md" {
		return true
	}
	prefix := "skills/"
	if adapter == "opencode" {
		prefix = ".agent-context/skills/"
	}
	if strings.HasPrefix(rel, prefix) {
		name := strings.TrimPrefix(rel, prefix)
		return identifier.MatchString(name)
	}
	return false
}
func (u *Unit) nativeTarget(t Target) bool {
	if t.Class == BackupQuarantine {
		_, ok := u.BackupAdapter(t.Path)
		return ok
	}

	for _, n := range u.Native {
		// Missing ancestors have separately pinned directory-only actions; no file or
		// recursive preimage can use this parent allowance.
		if (t.Class == "parent" || t.Class == "home-directory") && (t.Path == n.Root || within(t.Path, n.Root) || within(n.Root, t.Path)) {
			if t.Before.Kind == "absent" || within(n.Root, t.Path) && t.Before.Kind == "directory" {
				return t.After.Kind == "directory"
			}
		}
		for _, rel := range n.Paths {
			if t.Path == filepath.Join(n.Root, filepath.FromSlash(rel)) && (t.Class == "owned-bytes" || t.Class == "owned-tree" || t.Class == "native-marker") {
				return true
			}
		}
	}
	return false
}

// VerifyNativeRoots is a live physical guard, not credential discovery.
func (u *Unit) VerifyNativeRoots() error {
	if err := u.VerifyLegacyRoots(); err != nil {
		return err
	}
	for _, n := range u.Native {
		if err := envorigin.ValidateExternal(u.Manager.Path, n.Root); err != nil {
			return err
		}
	}
	return nil
}
