// Package graphtest builds deterministic, entirely synthetic profile stores.
package graphtest

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1" // #nosec G505 -- Synthetic Git fixtures require Git object identities.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/privatedir"
)

// Write creates a file inside a test-owned fixture tree.
func Write(root, rel string, raw []byte) error {
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(full, raw, 0600); err != nil {
		return err
	}
	// Mode bits alone do not establish the installed lock's private boundary
	// on Windows. Protect only this test-owned file using the native DACL.
	return privatedir.ProtectFile(full)
}

// Object writes a hash-verified synthetic object, never an authored delivery commit.
func Object(root, kind string, body []byte) (string, error) {
	raw := append([]byte(fmt.Sprintf("%s %d\x00", kind, len(body))), body...)
	sum := sha1.Sum(raw) // #nosec G401 -- Git object format in a synthetic fixture.
	id := hex.EncodeToString(sum[:])
	var compressed bytes.Buffer
	z := zlib.NewWriter(&compressed)
	if _, err := z.Write(raw); err != nil {
		return "", err
	}
	if err := z.Close(); err != nil {
		return "", err
	}
	return id, Write(root, "objects/"+id[:2]+"/"+id[2:], compressed.Bytes())
}
func tree(root string, files map[string][]byte) (string, error) {
	names := map[string]bool{}
	for p := range files {
		first, _, _ := strings.Cut(p, "/")
		names[first] = true
	}
	var order []string
	for name := range names {
		order = append(order, name)
	}
	sort.Strings(order)
	var out bytes.Buffer
	for _, name := range order {
		mode := "100644"
		var id string
		var err error
		if raw, ok := files[name]; ok {
			id, err = Object(root, "blob", raw)
		} else {
			mode = "40000"
			children := map[string][]byte{}
			for p, raw := range files {
				if strings.HasPrefix(p, name+"/") {
					children[strings.TrimPrefix(p, name+"/")] = raw
				}
			}
			id, err = tree(root, children)
		}
		if err != nil {
			return "", err
		}
		oid, _ := hex.DecodeString(id)
		out.WriteString(mode + " " + name)
		out.WriteByte(0)
		out.Write(oid)
	}
	return Object(root, "tree", out.Bytes())
}

// Commit writes synthetic Git objects and a matching immutable test snapshot.
func Commit(home, source string, files map[string][]byte) (string, error) {
	root := filepath.Join(envprofile.GraphRepository(home, source), ".git")
	tree, err := tree(root, files)
	if err != nil {
		return "", err
	}
	return Object(root, "commit", []byte("tree "+tree+"\nauthor Fixture <fixture@example.invalid> 0 +0000\ncommitter Fixture <fixture@example.invalid> 0 +0000\n\nSynthetic fixture\n"))
}
func material(home string, m contextlock.Member, files map[string][]byte) error {
	entry := envprofile.GraphMemberEntry(home, m)
	for p, raw := range files {
		if err := Write(entry, p, raw); err != nil {
			return err
		}
	}
	return nil
}
func manifest(name string, requires map[string]any, module bool) []byte {
	obj := map[string]any{"schema_version": 1, "name": name, "version": "1.0.0"}
	if module {
		obj["context"] = map[string]any{"modules": []any{map[string]any{"path": "base.md"}}}
	}
	if requires != nil {
		obj["requires"] = map[string]any{"contexts": requires}
	}
	raw, _ := json.Marshal(obj)
	return raw
}

// Small converts the contract example into a real immutable cache, replacing
// illustrative commit strings with content-addressed local Git objects.
func Small(home, fixture string) (*contextlock.Lock, error) {
	raw, err := os.ReadFile(filepath.Join(fixture, "context-lock.json")) // #nosec G304 -- Fixed testdata file.
	if err != nil {
		return nil, err
	}
	lock, err := contextlock.Parse(raw)
	if err != nil {
		return nil, err
	}
	for i := range lock.Members {
		source, err := identity.Parse(lock.Members[i].Source)
		if err != nil {
			return nil, err
		}
		lock.Members[i].Source = source
	}
	commits := map[string]string{}
	fileSets := map[string]map[string][]byte{}
	for _, m := range lock.Members {
		if m.Name == lock.Root {
			continue
		}
		dir := "nested"
		if m.Kind == "skill" {
			dir = "skill"
		} else if m.Overlay {
			dir = "overlay"
		}
		files := map[string][]byte{}
		err := filepath.WalkDir(filepath.Join(fixture, dir), func(p string, e os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if e.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(filepath.Join(fixture, dir), p)
			raw, err := os.ReadFile(p) // #nosec G304 -- Test-owned synthetic fixture path.
			if err != nil {
				return err
			}
			files[path.Join(m.Directory, filepath.ToSlash(rel))] = raw
			return nil
		})
		if err != nil {
			return nil, err
		}
		id, err := Commit(home, m.Source, files)
		if err != nil {
			return nil, err
		}
		commits[m.Name] = id
		fileSets[m.Name] = files
	}
	root, _ := lock.RootMember()
	files := map[string][]byte{}
	err = filepath.WalkDir(filepath.Join(fixture, "root"), func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(filepath.Join(fixture, "root"), p)
		raw, err := os.ReadFile(p) // #nosec G304 -- Test-owned synthetic fixture path.
		if err != nil {
			return err
		}
		if rel == "agent-context.json" {
			var obj map[string]any
			if err := json.Unmarshal(raw, &obj); err != nil {
				return err
			}
			requires := obj["requires"].(map[string]any)
			for _, family := range []string{"contexts", "skills"} {
				for name, value := range requires[family].(map[string]any) {
					value.(map[string]any)["revision"] = commits[name]
				}
			}
			raw, err = json.Marshal(obj)
			if err != nil {
				return err
			}
		}
		files[filepath.ToSlash(rel)] = raw
		return nil
	})
	if err != nil {
		return nil, err
	}
	id, err := Commit(home, root.Source, files)
	if err != nil {
		return nil, err
	}
	commits[root.Name] = id
	fileSets[root.Name] = files
	for i := range lock.Members {
		lock.Members[i].Commit = commits[lock.Members[i].Name]
		if err := material(home, lock.Members[i], fileSets[lock.Members[i].Name]); err != nil {
			return nil, err
		}
	}
	raw, err = lock.Canonical()
	if err != nil {
		return nil, err
	}
	if err := Write(home, "profiles/sample/lock.json", raw); err != nil {
		return nil, err
	}
	if err := Write(home, "profiles/sample/source.json", []byte(`{"kind":"git","git":"https://example.invalid/root","requirement":{"revision":"`+commits[root.Name]+`"}}`)); err != nil {
		return nil, err
	}
	cfg := map[string]any{"environments": map[string]any{"overlays": map[string]any{"sample": []any{map[string]any{"source": "https://example.invalid/overlay", "directory": "profiles/extra", "revision": commits["sample-overlay"], "weight": 10}}}}}
	raw, _ = json.Marshal(cfg)
	return lock, Write(home, "config.json", raw)
}

// DAG retains two parents of one package and two distinct overlay directories
// at the same locator/revision with different weights.
func DAG(home string) (*contextlock.Lock, error) {
	source := "example.invalid/packages"
	pin := ""
	req := func(name string) map[string]any {
		return map[string]any{"git": "https://" + source, "revision": pin, "directory": name}
	}
	// Dependencies use one independent immutable leaf snapshot, so no commit
	// needs to refer to its own hash.
	leafFiles := map[string][]byte{"agent-context.json": manifest("leaf", nil, true), "context/base.md": []byte("# Shared leaf\n")}
	leafPin, err := Commit(home, "example.invalid/leaf", leafFiles)
	if err != nil {
		return nil, err
	}
	leafReq := map[string]any{"leaf": map[string]any{"git": "https://example.invalid/leaf", "revision": leafPin}}
	files := map[string][]byte{}
	for _, name := range []string{"left", "right", "overlay-a", "overlay-b"} {
		rs := leafReq
		if strings.HasPrefix(name, "overlay") {
			rs = nil
		}
		files[name+"/agent-context.json"] = manifest(name, rs, true)
		files[name+"/context/base.md"] = []byte("# " + name + "\n")
	}
	pin, err = Commit(home, source, files)
	if err != nil {
		return nil, err
	}
	rootFiles := map[string][]byte{"agent-context.json": manifest("sample", map[string]any{"left": req("left"), "right": req("right")}, true), "context/base.md": []byte("# Root\n")}
	h, err := hashing.ContentSHA256Files(rootFiles, hashing.VersionV2)
	if err != nil {
		return nil, err
	}
	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: 2, Root: "sample", Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Members: []contextlock.Member{{Kind: "context", Name: "sample", Version: "1.0.0", StateHash: strings.TrimPrefix(h, "sha256:"), RequiredBy: []string{}}, {Kind: "context", Name: "leaf", Version: "1.0.0", Commit: leafPin, Source: "example.invalid/leaf", RequiredBy: []string{"left", "right"}}}}
	for _, name := range []string{"left", "right", "overlay-a", "overlay-b"} {
		m := contextlock.Member{Kind: "context", Name: name, Source: source, Directory: name, Commit: pin, Version: "1.0.0", RequiredBy: []string{"sample"}}
		if strings.HasPrefix(name, "overlay") {
			m.Overlay = true
			m.RequiredBy = []string{}
			m.Weight = 10
			if name == "overlay-b" {
				m.Weight = 20
			}
		}
		lock.Members = append(lock.Members, m)
	}
	lock.Sort()
	for _, m := range lock.Members {
		selected := files
		switch m.Name {
		case "sample":
			selected = rootFiles
		case "leaf":
			selected = leafFiles
		}
		if err := material(home, m, selected); err != nil {
			return nil, err
		}
	}
	raw, err := lock.Canonical()
	if err != nil {
		return nil, err
	}
	if err := Write(home, "profiles/sample/lock.json", raw); err != nil {
		return nil, err
	}
	if err := Write(home, "profiles/sample/source.json", []byte(`{"kind":"path","path":"source","requirement":{}}`)); err != nil {
		return nil, err
	}
	overlays := []any{}
	for _, name := range []string{"overlay-a", "overlay-b"} {
		weight := 10
		if name == "overlay-b" {
			weight = 20
		}
		overlays = append(overlays, map[string]any{"source": "https://" + source, "directory": name, "revision": pin, "weight": weight})
	}
	raw, _ = json.Marshal(map[string]any{"environments": map[string]any{"overlays": map[string]any{"sample": overlays}}})
	return lock, Write(home, "config.json", raw)
}
