package projectcontext

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// All candidate reads are anchored in os.Root. Component lstat rejects links;
// identity and metadata checks detect substitutions around the bounded read.
// os.Root also prevents a concurrent symlink swap escaping the open root.
func read(root *os.Root, name string, byteLimit int64) ([]byte, os.FileMode, error) {
	if !relative(name) {
		return nil, 0, fail("project_context_invalid", "unsafe relative path")
	}
	current := ""
	for _, part := range strings.Split(name, "/") {
		current = path.Join(current, part)
		st, err := root.Lstat(current)
		if err != nil {
			return nil, 0, err
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return nil, 0, fail("project_context_invalid", "symlink in candidate path")
		}
		if current != name && !st.IsDir() {
			return nil, 0, fail("project_context_invalid", "non-directory component")
		}
	}
	before, err := root.Lstat(name)
	if err != nil {
		return nil, 0, err
	}
	if before.Mode().Perm()&0444 == 0 {
		return nil, 0, fail("project_context_unreadable", "candidate file is unreadable")
	}
	if !before.Mode().IsRegular() || !singleLink(before) || before.Size() > byteLimit {
		return nil, 0, fail("project_context_invalid", "unreadable, special or oversized file")
	}
	f, err := openCandidate(root, name)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = f.Close() }()
	opened, err := f.Stat()
	if err != nil || !same(before, opened) {
		return nil, 0, fail("project_context_drift", "candidate changed before read")
	}
	b, err := io.ReadAll(io.LimitReader(f, byteLimit+1))
	if err != nil {
		return nil, 0, err
	}
	after, err := root.Lstat(name)
	if err != nil || !same(before, after) || int64(len(b)) != before.Size() {
		return nil, 0, fail("project_context_drift", "candidate changed during read")
	}
	opened, err = f.Stat()
	if err != nil || !same(before, opened) {
		return nil, 0, fail("project_context_drift", "candidate changed during read")
	}
	mode := os.FileMode(0600)
	if before.Mode()&0111 != 0 {
		mode = 0700
	}
	return b, mode, nil
}
func same(a, b os.FileInfo) bool {
	return a != nil && b != nil && os.SameFile(a, b) && a.Mode() == b.Mode() && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

type captured struct {
	Manifest Snapshot
	Bytes    map[string][]byte
}

func capture(root *os.Root, prefix string) (captured, error) {
	return captureTree(root, prefix, true)
}

// A source package may contain its own Git metadata; protected snapshots may
// contain only their declared portable content, without that source exclusion.
func captureTree(root *os.Root, prefix string, sourceGitMetadata bool) (captured, error) {
	out := captured{Manifest: Snapshot{Version: 1, Files: []File{}}, Bytes: map[string][]byte{}}
	aliases := map[string]bool{}
	total := int64(0)
	nodes := 0
	var walk func(string, int) error
	walk = func(rel string, depth int) error {
		if depth > 32 {
			return fail("project_context_invalid", "inventory depth exceeds 32")
		}
		dir := prefix
		if rel != "" {
			dir = path.Join(prefix, rel)
		}
		if dir == "" {
			dir = "."
		}
		st, err := root.Lstat(dir)
		if err != nil {
			return err
		}
		if !st.IsDir() || st.Mode().Perm()&0500 == 0 {
			return fail("project_context_invalid", "unreadable or linked candidate directory")
		}
		f, err := root.Open(dir)
		if err != nil {
			return err
		}
		entries, err := f.ReadDir(MaxFiles + 1)
		_ = f.Close()
		if err != nil && err != io.EOF {
			return err
		}
		if len(entries) > MaxFiles {
			return fail("project_context_invalid", "inventory exceeds file bound")
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for _, e := range entries {
			nodes++
			if nodes > MaxFiles {
				return fail("project_context_invalid", "inventory node bound exceeded")
			}
			if e.Name() == ".git" {
				if rel == "" && sourceGitMetadata {
					continue
				}
				return fail("project_context_invalid", "nested Git metadata is not snapshot content")
			}
			p := path.Join(rel, e.Name())
			if !relative(p) {
				return fail("project_context_invalid", "nonportable inventory path")
			}
			alias := cases.Fold().String(norm.NFC.String(p))
			if aliases[alias] {
				return fail("project_context_invalid", "case or normalization alias")
			}
			aliases[alias] = true
			full := path.Join(prefix, p)
			st, err := root.Lstat(full)
			if err != nil {
				return err
			}
			if st.IsDir() {
				if err := walk(p, depth+1); err != nil {
					return err
				}
				continue
			}
			if len(out.Manifest.Files) >= MaxFiles {
				return fail("project_context_invalid", "too many files")
			}
			b, mode, err := read(root, full, MaxFileBytes)
			if err != nil {
				return err
			}
			total += int64(len(b))
			if total > MaxTotalBytes {
				return fail("project_context_invalid", "snapshot exceeds total byte bound")
			}
			out.Bytes[p] = b
			out.Manifest.Files = append(out.Manifest.Files, File{p, int64(len(b)), digest(b), fmt.Sprintf("%04o", mode)})
		}
		return nil
	}
	if err := walk("", 0); err != nil {
		return out, err
	}
	sort.Slice(out.Manifest.Files, func(i, j int) bool { return out.Manifest.Files[i].Path < out.Manifest.Files[j].Path })
	return out, nil
}
func fromFiles(files map[string][]byte) captured {
	c := captured{Manifest: Snapshot{Version: 1, Files: []File{}}, Bytes: files}
	for p, b := range files {
		c.Manifest.Files = append(c.Manifest.Files, File{p, int64(len(b)), digest(b), "0600"})
	}
	sort.Slice(c.Manifest.Files, func(i, j int) bool { return c.Manifest.Files[i].Path < c.Manifest.Files[j].Path })
	return c
}
func equalCapture(a, b captured) bool { return identity(a.Manifest) == identity(b.Manifest) }

func discovery(root *os.Root) ([]Discovery, error) {
	out := []Discovery{}
	count := 0
	total := int64(0)
	files := map[string]string{"AGENTS.md": "instructions", "AGENTS.override.md": "instructions", "CLAUDE.md": "instructions", "CLAUDE.local.md": "instructions", ".mcp.json": "mcp", "agent-context.json": "context-package", "Skillfile.json": "skill-declaration", "opencode.json": "native-config", "opencode.jsonc": "native-config"}
	dirs := map[string]bool{".claude": true, ".codex": true, ".muse": true, ".agents": true, ".opencode": true, ".pi": true, "rules": true, "knowledge": true, "context": true}
	var walk func(string, int) error
	walk = func(dir string, depth int) error {
		if depth > 32 {
			return fail("project_context_invalid", "discovery depth exceeded")
		}
		f, err := root.Open(dir)
		if err != nil {
			return err
		}
		entries, err := f.ReadDir(MaxFiles + 1)
		_ = f.Close()
		if err != nil && err != io.EOF {
			return err
		}
		if len(entries) > MaxFiles {
			return fail("project_context_invalid", "discovery bound exceeded")
		}
		for _, e := range entries {
			p := e.Name()
			if dir != "." {
				p = path.Join(dir, p)
			}
			kind, known := files[e.Name()]
			if !known && dir == "." && !dirs[e.Name()] {
				continue
			}
			count++
			if count > MaxFiles {
				return fail("project_context_invalid", "discovery bound exceeded")
			}
			st, err := root.Lstat(p)
			if err != nil {
				return err
			}
			if st.Mode()&os.ModeSymlink != 0 {
				return fail("project_context_invalid", "linked discovery candidate")
			}
			if st.IsDir() {
				if err := walk(p, depth+1); err != nil {
					return err
				}
				continue
			}
			b, _, err := read(root, p, MaxFileBytes)
			if err != nil {
				return err
			}
			total += int64(len(b))
			if total > MaxTotalBytes {
				return fail("project_context_invalid", "discovery byte bound exceeded")
			}
			if !known {
				kind = "native-config"
				if strings.HasPrefix(p, "knowledge/") {
					kind = "knowledge"
				}
				if strings.Contains(p, "/rules/") || strings.HasPrefix(p, "rules/") {
					kind = "rules"
				}
				if strings.Contains(p, "/skills/") {
					kind = "skill"
				}
			}
			out = append(out, Discovery{p, kind, digest(b), "untrusted-candidate"})
		}
		return nil
	}
	if err := walk(".", 0); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func within(parent, child string) bool {
	r, err := filepath.Rel(parent, child)
	return err == nil && r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator)) && !filepath.IsAbs(r)
}
func noNativeImports(b []byte) bool {
	// Leaf 1 has no import dialect. Reject Claude-style whitespace-prefixed
	// @path tokens, including inline directives; ordinary Markdown links stay
	// opaque. This conservative boundary is reported in preview documentation.
	for _, line := range bytes.Split(b, []byte("\n")) {
		for _, token := range bytes.Fields(line) {
			if len(token) > 1 && token[0] == '@' {
				return false
			}
		}
	}
	return true
}

func portableAlias(p string) string { return cases.Fold().String(norm.NFC.String(p)) }
