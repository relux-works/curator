package envunit

import (
	"github.com/relux-works/curator/internal/buildrepo"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/staging"
)

// LegacyPublication is the existing install publisher's closed, enumerated
// consumer inventory in a global unit. UserHome is mirror custody only; it is
// never a credential origin or an account-discovery anchor. The actual journal
// binds every entry's metadata/preimage/postimage and durable boundary proof.
type LegacyPublication struct {
	UserHome      string         `json:"user_home"`
	UserBin       string         `json:"user_bin,omitempty"`
	ExternalStore string         `json:"external_store,omitempty"`
	Entries       []LegacyTarget `json:"entries"`
}
type LegacyTarget struct {
	Class string `json:"class"`
	Path  string `json:"path"`
}

func legacyClass(class string) bool {
	switch class {
	case staging.ClassContext, staging.ClassRuntime, staging.ClassCanonicalShim,
		staging.ClassForwardingShim, staging.ClassEnvFile, staging.ClassAdapterLedger,
		staging.ClassMirrorLedger, staging.ClassRemoval, "05-external-cache", "05-external-snapshot", "55-global-declaration", "56-hook-approvals":
		return true
	}
	return false
}
func leaf(root, path string) bool {
	return filepath.Dir(path) == root && identifier.MatchString(filepath.Base(path))
}
func (l *LegacyPublication) mirrorRoots() []string {
	var roots []string
	for _, rel := range adapters.AgentPaths {
		roots = append(roots, filepath.Join(l.UserHome, filepath.FromSlash(rel)))
	}
	return append(roots, filepath.Join(l.UserHome, filepath.FromSlash(adapters.NativeDiscoveryHomePath)))
}

// PublicationRoots enumerates external mirror/bin custody, not credential
// discovery. Only absent parent entries of declared leaf targets may use it.
func (l *LegacyPublication) PublicationRoots() []string {
	roots := l.mirrorRoots()
	if l.UserBin != "" {
		roots = append(roots, l.UserBin)
	}
	if l.ExternalStore != "" {
		roots = append(roots, l.ExternalStore)
	}
	return roots
}
func (u *Unit) validateLegacyPath(class, path string) error {
	l := u.Legacy
	if l == nil || (u.Action != GlobalExtend && u.Action != Takeover) || !absolute(path) || !absolute(l.UserHome) || within(u.Manager.Path, l.UserHome) {
		return invalid("legacy publication binding")
	}
	global := filepath.Join(u.Manager.Path, "global")
	allowed := false
	switch class {
	case staging.ClassEnvFile:
		allowed = path == filepath.Join(global, "env.sh") || path == filepath.Join(global, "env.ps1")
	case "56-hook-approvals":
		allowed = path == filepath.Join(u.Manager.Path, "hook-approvals.tsv")
	case "55-global-declaration":
		allowed = path == filepath.Join(global, "Skillfile.json")
	case staging.ClassContext, staging.ClassRemoval:
		allowed = leaf(filepath.Join(global, "skills"), path) || leaf(filepath.Join(global, ".curator-generations"), path)
		if class == staging.ClassRemoval {
			for _, root := range l.mirrorRoots() {
				allowed = allowed || leaf(root, path) && filepath.Base(path) != adapters.LedgerName
			}
		}
		if class == staging.ClassRemoval {
			allowed = allowed || leaf(filepath.Join(global, "bin"), path) || l.UserBin != "" && leaf(l.UserBin, path)
		}
	case staging.ClassRuntime:
		rel, e := filepath.Rel(filepath.Join(u.Manager.Path, "runtime"), path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		allowed = e == nil && len(parts) == 2 && identifier.MatchString(parts[0]) && identifier.MatchString(parts[1])
	case staging.ClassCanonicalShim:
		allowed = leaf(filepath.Join(global, "bin"), path)
	case staging.ClassForwardingShim:
		allowed = l.UserBin != "" && leaf(l.UserBin, path) && filepath.Base(path) != ".curator-managed.json"
	case staging.ClassMirrorLedger:
		allowed = l.UserBin != "" && path == filepath.Join(l.UserBin, ".curator-managed.json")
	case staging.ClassAdapterLedger:
		for _, root := range l.mirrorRoots() {
			allowed = allowed || path == filepath.Join(root, adapters.LedgerName) || leaf(root, path)
		}
	case "05-external-cache", "05-external-snapshot":
		rel, e := filepath.Rel(l.ExternalStore, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		allowed = l.ExternalStore != "" && e == nil && len(parts) == 2 && externalKey.MatchString(parts[1]) &&
			(class == "05-external-snapshot" && parts[0] == "snapshots" || class == "05-external-cache" && (parts[0] == buildrepo.ArtifactsDir(2) || parts[0] == buildrepo.ArtifactsDir(3)))
	}
	if !allowed {
		return invalid("legacy %s target outside closed consumer namespace: %s", class, path)
	}
	return nil
}

var externalKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (u *Unit) validateExternalStore() error {
	l := u.Legacy
	root := l.ExternalStore
	if root == "" {
		return nil
	}
	if !absolute(root) || root == filepath.Dir(root) || within(root, u.Manager.Path) || within(u.Manager.Path, root) && root != filepath.Join(u.Manager.Path, "external-build-cache") || within(root, l.UserHome) {
		return invalid("legacy external store custody")
	}
	roots := []string{l.UserBin}
	for _, rel := range []string{".codex", ".claude", ".pi", ".config", ".local", ".agents", ".gemini", ".cursor"} {
		roots = append(roots, filepath.Join(l.UserHome, rel))
	}
	for _, n := range u.Native {
		roots = append(roots, n.Root)
	}
	for _, h := range u.Homes {
		roots = append(roots, h.Target)
	}
	for _, other := range roots {
		if other != "" && (within(root, other) || within(other, root)) {
			return invalid("legacy external store overlaps consumer/native custody")
		}
	}
	return nil
}
func (u *Unit) validateLegacy() error {
	l := u.Legacy
	if l == nil {
		return nil
	}
	if len(l.Entries) == 0 || len(l.Entries) > 8192 {
		return invalid("legacy target inventory size")
	}
	if l.UserBin != "" {
		if !absolute(l.UserBin) || l.UserBin == l.UserHome || !within(l.UserHome, l.UserBin) || within(u.Manager.Path, l.UserBin) {
			return invalid("legacy user-bin custody")
		}
		// A PATH entry cannot turn a native account/config/skills namespace
		// into a command bin. This is independent of credential origin capture.
		for _, rel := range []string{".codex", ".claude", ".pi", ".config", ".agents", ".gemini", ".cursor"} {
			root := filepath.Join(l.UserHome, rel)
			if within(root, l.UserBin) || within(l.UserBin, root) {
				return invalid("legacy user bin overlaps native custody")
			}
		}
	}
	if err := u.validateExternalStore(); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, entry := range l.Entries {
		if seen[entry.Path] || !legacyClass(entry.Class) {
			return invalid("legacy duplicate/class")
		}
		seen[entry.Path] = true
		if err := u.validateLegacyPath(entry.Class, entry.Path); err != nil {
			return err
		}
	}
	return nil
}
func (u *Unit) LegacyOwns(class, path string) bool {
	if u.Legacy == nil || u.validateLegacy() != nil {
		return false
	}
	for _, entry := range u.Legacy.Entries {
		if entry.Class == class && entry.Path == path {
			return true
		}
	}
	return false
}
func (u *Unit) LegacyParent(path string) bool {
	if u.Legacy == nil || u.validateLegacy() != nil {
		return false
	}
	for _, entry := range u.Legacy.Entries {
		if within(path, filepath.Dir(entry.Path)) {
			return true
		}
	}
	return false
}
func (u *Unit) VerifyLegacyRoots() error {
	if u.Legacy == nil {
		return nil
	}
	if err := u.validateLegacy(); err != nil {
		return err
	}
	for _, root := range u.Legacy.mirrorRoots() {
		if err := envorigin.ValidateExternal(u.Manager.Path, root); err != nil {
			return err
		}
	}
	if u.Legacy.UserBin != "" {
		if err := envorigin.ValidateExternal(u.Manager.Path, u.Legacy.UserBin); err != nil {
			return err
		}
	}
	if u.Legacy.ExternalStore != "" {
		if err := envorigin.ValidateExternal(u.Manager.Path, u.Legacy.ExternalStore); err != nil {
			return err
		}
		if err := (&buildrepo.DiskProtectedStore{Root: u.Legacy.ExternalStore}).CheckNamespaces(2, 3); err != nil {
			return err
		}
	}
	return nil
}

// A recursive legacy target is still a single closed context/runtime/cache
// leaf, never a whole account, native home, credential or session tree.
func recursiveClass(class string) bool {
	return class == "installed-profile" || class == "owned-tree" || class == staging.ClassContext || class == staging.ClassAdapterLedger || class == staging.ClassRuntime || class == staging.ClassRemoval || class == "05-external-cache" || class == "05-external-snapshot"
}
