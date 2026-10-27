// Package testhomeguard isolates test processes from operator native homes.
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
	home   string
}

// Start records the OS account home as well as inherited home overrides before
// pinning all provider defaults. A caller's belt-and-braces HOME does not hide
// the OS account home on hosts whose account database supplies it.
func Start() (*Guard, error) {
	roots := map[string]bool{}
	add := func(path string) {
		if path != "" {
			roots[filepath.Clean(path)] = true
		}
	}
	homes := map[string]bool{}
	for _, name := range []string{"HOME", "USERPROFILE"} {
		if home := os.Getenv(name); home != "" {
			homes[home] = true
		}
	}
	account, err := user.Current()
	if err != nil || account.HomeDir == "" {
		return nil, fmt.Errorf("testhomeguard: OS account home unavailable")
	}
	homes[account.HomeDir] = true
	for home := range homes {
		for _, suffix := range []string{".claude", ".codex", ".config/opencode", ".config/muse", ".local/share/muse", ".pi", ".pi/agent", ".agents"} {
			add(filepath.Join(home, filepath.FromSlash(suffix)))
		}
	}
	for _, name := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR"} {
		add(os.Getenv(name))
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		add(filepath.Join(xdg, "opencode"))
		add(filepath.Join(xdg, "muse"))
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
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
	g.home, err = os.MkdirTemp("", "curator-test-home-*")
	if err != nil {
		return nil, fmt.Errorf("testhomeguard: cannot create isolated home")
	}
	for _, dir := range []string{".config", ".local/share", ".local/state", ".cache"} {
		if err := os.MkdirAll(filepath.Join(g.home, filepath.FromSlash(dir)), 0700); err != nil {
			_ = os.RemoveAll(g.home)
			return nil, err
		}
	}
	for name, value := range map[string]string{
		"HOME": g.home, "USERPROFILE": g.home,
		"XDG_CONFIG_HOME": filepath.Join(g.home, ".config"),
		"XDG_DATA_HOME":   filepath.Join(g.home, ".local", "share"),
		"XDG_STATE_HOME":  filepath.Join(g.home, ".local", "state"),
		"XDG_CACHE_HOME":  filepath.Join(g.home, ".cache"),
	} {
		if err := os.Setenv(name, value); err != nil {
			_ = os.RemoveAll(g.home)
			return nil, err
		}
	}
	for _, name := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR"} {
		if err := os.Unsetenv(name); err != nil {
			_ = os.RemoveAll(g.home)
			return nil, err
		}
	}
	return g, nil
}

// Finish fails the package on a changed identity, link, mode or context hash,
// including when the tests otherwise pass. It removes only its own fresh home.
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
			}
		}
		code = 1
	} else {
		fmt.Fprintln(os.Stderr, "testhomeguard: protected native homes unchanged")
	}
	if err := os.RemoveAll(g.home); err != nil {
		fmt.Fprintln(os.Stderr, "testhomeguard: isolated home cleanup failed")
		code = 1
	}
	return code
}

func snapshot(roots []string) (map[string]identity, error) {
	out := map[string]identity{}
	for _, root := range roots {
		if err := capture(out, root, false); err != nil {
			return nil, err
		}
		for _, name := range []string{"CLAUDE.md", "AGENTS.md", ".agent-environment.json"} {
			if err := capture(out, filepath.Join(root, name), true); err != nil {
				return nil, err
			}
		}
		skills := filepath.Join(root, "skills")
		if err := capture(out, skills, false); err != nil {
			return nil, err
		}
		i, err := os.Lstat(skills)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !i.IsDir() {
			continue
		}
		f, err := os.Open(skills)
		if err != nil {
			return nil, err
		}
		entries, readErr := f.ReadDir(4097)
		closeErr := f.Close()
		if readErr != nil && readErr != io.EOF {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(entries) > 4096 {
			return nil, fmt.Errorf("skill identity bound exceeded")
		}
		for _, entry := range entries {
			if err := capture(out, filepath.Join(skills, entry.Name()), false); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
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
		if i.Size() > 16<<20 {
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
		n, readErr := io.Copy(h, io.LimitReader(f, (16<<20)+1))
		closeErr := f.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		if n > 16<<20 {
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
		if x.info != nil && (!os.SameFile(x.info, y.info) || x.info.Mode() != y.info.Mode() || x.info.Size() != y.info.Size() || !x.info.ModTime().Equal(y.info.ModTime())) {
			return false
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
