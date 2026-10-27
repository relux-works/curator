// Package testhomeguard observes protected native homes around test processes.
// Production code must not import it. Snapshots never read credential files.
package testhomeguard

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

const Changed = "testhomeguard: protected native home changed"

type identity struct {
	info os.FileInfo
	link string
	hash [sha256.Size]byte
}

type Guard struct {
	roots  []string
	before map[string]identity
}

// start observes account authority and, for an ambient test process, inherited
// native homes. Fixture helpers retain their deliberately supplied environment;
// the surrounding package supervisor already watches its original ambient roots.
// This package never pins HOME: testgitenv is the sole isolation owner.
func start(inherited bool) (*Guard, error) {
	roots := map[string]bool{}
	add := func(path string) {
		if path != "" {
			roots[filepath.Clean(path)] = true
		}
	}
	homes := map[string]bool{}
	for _, name := range []string{"HOME", "USERPROFILE"} {
		if home := os.Getenv(name); inherited && home != "" {
			homes[home] = true
		}
	}
	account, err := user.Current()
	if err != nil || account.HomeDir == "" {
		return nil, fmt.Errorf("testhomeguard: OS account home unavailable")
	}
	homes[account.HomeDir] = true
	for home := range homes {
		add(filepath.Join(home, ".claude.json"))
		for _, suffix := range []string{".claude", ".codex", ".config/opencode", ".config/muse", ".local/share/muse", ".pi", ".pi/agent", ".agents"} {
			add(filepath.Join(home, filepath.FromSlash(suffix)))
		}
	}
	if inherited {
		for _, name := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR"} {
			add(os.Getenv(name))
		}
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); inherited && xdg != "" {
		add(filepath.Join(xdg, "opencode"))
		add(filepath.Join(xdg, "muse"))
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); inherited && xdg != "" {
		add(filepath.Join(xdg, "muse"))
	}
	g := &Guard{}
	for root := range roots {
		if !filepath.IsAbs(root) {
			return nil, fmt.Errorf("testhomeguard: relative inherited native home")
		}
		g.roots = append(g.roots, root)
	}
	sort.Strings(g.roots)
	g.before, err = snapshot(g.roots)
	if err != nil {
		return nil, fmt.Errorf("testhomeguard: cannot establish protected preimage")
	}
	return g, nil
}

// Finish fails the package on a changed identity, link, mode or context hash,
// including when the tests otherwise pass. It owns no isolation directory.
func (g *Guard) Finish(code int) int {
	after, err := snapshot(g.roots)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testhomeguard: protected native home unreadable")
		code = 1
	} else if !same(g.before, after) {
		for _, root := range g.roots {
			a, b := map[string]identity{}, map[string]identity{}
			for path, v := range g.before {
				if path == root || strings.HasPrefix(path, root+string(filepath.Separator)) {
					a[path] = v
				}
			}
			for path, v := range after {
				if path == root || strings.HasPrefix(path, root+string(filepath.Separator)) {
					b[path] = v
				}
			}
			if !same(a, b) {
				fmt.Fprintf(os.Stderr, "%s [%s]\n", Changed, fingerprint(root))
				paths := map[string]bool{}
				for path := range a {
					paths[path] = true
				}
				for path := range b {
					paths[path] = true
				}
				for path := range paths {
					if !same(map[string]identity{path: a[path]}, map[string]identity{path: b[path]}) {
						fmt.Fprintf(os.Stderr, "testhomeguard: changed entry [%s]\n", fingerprint(path))
					}
				}
			}
		}
		code = 1
	} else {
		fmt.Fprintln(os.Stderr, "testhomeguard: protected native homes unchanged")
	}
	return code
}

const maxEntries = 65536
const maxDepth = 64

// Installed skill trees can contain large tool assets. Keep the complete
// recursive inventory bounded without silently omitting those regular files.
const maxFileBytes = 128 << 20
const maxSnapshotBytes = 1 << 30

// Native configuration, instructions and Curator environment state only.
// Runtime databases, logs and session histories are deliberately excluded.
var files = []string{
	"CLAUDE.md", "AGENTS.md", "AGENTS.override.md", "SYSTEM.md", "APPEND_SYSTEM.md",
	"settings.json", "settings.local.json", ".claude.json", "config.json", "config.jsonc", "config.toml",
	"opencode.json", "opencode.jsonc", "models.json", "trust.json", ".mcp.json",
	"curator-mcp.config.toml", "curator-mcp.json", ".agent-environment.json",
	"auth.json", ".credentials.json", "credentials.json", "oauth.json", "tokens.json",
}

func snapshot(roots []string) (map[string]identity, error) {
	out := map[string]identity{}
	var bytes int64
	for _, root := range roots {
		if err := capture(out, root, false); err != nil {
			return nil, err
		}
		if i := out[root].info; i != nil && !i.IsDir() {
			if i.Mode().IsRegular() && filepath.Base(root) == ".claude.json" {
				if err := captureBounded(out, root, true, &bytes); err != nil {
					return nil, err
				}
			}
			continue // Observe a linked root itself without traversing its target.
		}
		for _, name := range files {
			path := filepath.Join(root, name)
			if err := captureBounded(out, path, !credential(name), &bytes); err != nil {
				return nil, err
			}
		}
		for _, name := range []string{"skills", ".agent-environment-backup"} {
			if err := captureTree(out, filepath.Join(root, name), 0, &bytes); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

// Known credential carriers are metadata-only, even inside copied skills.
func credential(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "credential") || strings.Contains(name, "token") ||
		strings.Contains(name, "secret") || strings.Contains(name, "oauth") ||
		strings.Contains(name, "auth.json") || name == "auth" ||
		strings.HasPrefix(name, "auth-") || strings.HasPrefix(name, "auth_") ||
		strings.HasPrefix(name, "auth.") || strings.HasPrefix(name, "id_rsa") ||
		strings.HasPrefix(name, "id_ed25519") || strings.HasPrefix(name, "id_ecdsa") ||
		name == ".netrc" || name == ".npmrc" || name == ".pypirc" || strings.HasPrefix(name, ".env") ||
		strings.HasSuffix(name, ".key") || strings.HasSuffix(name, ".pem") ||
		strings.HasSuffix(name, ".p12") || strings.HasSuffix(name, ".pfx")
}

func captureBounded(out map[string]identity, path string, hash bool, bytes *int64) error {
	if len(out) >= maxEntries {
		return fmt.Errorf("native snapshot entry bound exceeded")
	}
	if err := capture(out, path, hash); err != nil {
		return err
	}
	if i := out[path].info; hash && i != nil && i.Mode().IsRegular() {
		*bytes += i.Size()
		if *bytes > maxSnapshotBytes {
			return fmt.Errorf("native snapshot byte bound exceeded")
		}
	}
	return nil
}

func captureTree(out map[string]identity, path string, depth int, bytes *int64) error {
	if depth > maxDepth {
		return fmt.Errorf("native snapshot depth bound exceeded")
	}
	if err := captureBounded(out, path, !credential(filepath.Base(path)), bytes); err != nil {
		return err
	}
	i := out[path].info
	if i == nil || !i.IsDir() {
		return nil
	} // Lstat: never follow a skill symlink.
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	opened, err := f.Stat()
	if err != nil || !os.SameFile(i, opened) {
		_ = f.Close()
		return fmt.Errorf("skill directory changed during snapshot")
	}
	entries, readErr := f.ReadDir(maxEntries + 1)
	closeErr := f.Close()
	if readErr != nil && readErr != io.EOF {
		return readErr
	}
	if closeErr != nil {
		return closeErr
	}
	if len(entries) > maxEntries {
		return fmt.Errorf("skill directory entry bound exceeded")
	}
	for _, e := range entries {
		if err := captureTree(out, filepath.Join(path, e.Name()), depth+1, bytes); err != nil {
			return err
		}
	}
	return nil
}

func capture(out map[string]identity, path string, hashContent bool) error {
	i, err := os.Lstat(path)
	if os.IsNotExist(err) {
		out[path] = identity{}
		return nil
	}
	if err != nil {
		return err
	}
	v := identity{info: i}
	if i.Mode()&os.ModeSymlink != 0 {
		v.link, err = os.Readlink(path)
	} else if hashContent && i.Mode().IsRegular() {
		if i.Size() > maxFileBytes {
			return fmt.Errorf("context identity bound exceeded")
		}
		f, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		opened, statErr := f.Stat()
		if statErr != nil || !os.SameFile(i, opened) {
			_ = f.Close()
			return fmt.Errorf("context identity changed during snapshot")
		}
		h := sha256.New()
		n, readErr := io.Copy(h, io.LimitReader(f, (maxFileBytes)+1))
		closeErr := f.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		if n > maxFileBytes {
			return fmt.Errorf("context identity bound exceeded")
		}
		copy(v.hash[:], h.Sum(nil))
	}
	if err != nil {
		return err
	}
	out[path] = v
	return nil
}

func same(a, b map[string]identity) bool {
	if len(a) != len(b) {
		return false
	}
	for path, x := range a {
		y, ok := b[path]
		if !ok || (x.info == nil) != (y.info == nil) {
			return false
		}
		if x.info != nil {
			if !os.SameFile(x.info, y.info) || x.info.Mode() != y.info.Mode() {
				return false
			}
			// Directory timestamps and sizes reflect unrelated runtime files
			// (for example Codex SQLite sidecars), not directory identity.
			// Protected children and the complete skill-entry set are compared
			// separately, including file timestamps, sizes and context hashes.
			if !x.info.IsDir() && (x.info.Size() != y.info.Size() || !x.info.ModTime().Equal(y.info.ModTime())) {
				return false
			}
		}
		if x.link != y.link || x.hash != y.hash {
			return false
		}
	}
	return true
}

func fingerprint(path string) string {
	digest := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", digest[:8])
}
