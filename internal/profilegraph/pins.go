package profilegraph

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/stateread"
)

// verifyState reuses install's versioned content hash over protected bounded
// reads. Every admitted snapshot file participates; denied siblings are never
// read. The remaining tree must match the lock, so an altered ordinary sibling
// or a pin containing excluded payloads cannot borrow the old identity.
func (b *build) verifyState(p packageInput) error {
	files := map[string][]byte{}
	if err := b.sourceTree(p.root, p.label, ".", files, 0); err != nil {
		return err
	}
	if b.lock.ContentHashVersion() == hashing.VersionV1 {
		for _, raw := range files {
			if bytes.IndexByte(raw, 0) >= 0 {
				return refusal("graph_invalid_input")
			}
		}
	}
	h, err := hashing.ContentSHA256Files(files, b.lock.ContentHashVersion())
	if err != nil {
		return err
	}
	if strings.TrimPrefix(h, "sha256:") != p.member.StateHash {
		return refusal("graph_invalid_input")
	}
	return nil
}
func (b *build) verifyGit(p packageInput) error {
	g := b.db(p.member.Source)
	kind, raw, err := g.object(p.member.Commit, 0)
	if err != nil {
		return err
	}
	if kind != "commit" {
		return refusal("graph_invalid_input")
	}
	line, _, _ := bytes.Cut(raw, []byte{'\n'})
	expected := strings.TrimPrefix(string(line), "tree ")
	format := "sha1"
	if len(p.member.Commit) == 64 {
		format = "sha256"
	}
	if p.root == "" {
		return nil
	}
	rel := func(full string) (string, error) {
		r, err := filepath.Rel(p.root, full)
		return filepath.ToSlash(r), err
	}
	actual, err := gitops.TreeObjectIDWith(p.root, format, gitops.TreeReader{
		ReadDir: func(full string) (stateread.Directory, error) {
			r, err := rel(full)
			if err != nil {
				return stateread.Directory{}, err
			}
			if len(strings.Split(r, "/")) > b.r.Limits.Depth {
				return stateread.Directory{}, refusal("graph_limit_exceeded")
			}
			entries, err := b.r.directory(p.root, r)
			return stateread.Directory{Kind: stateread.KindPresent, Entries: entries}, err
		},
		Lstat: func(full string) (stateread.Metadata, error) {
			r, err := rel(full)
			if err != nil {
				return stateread.Metadata{}, err
			}
			if denied(r) {
				return stateread.Metadata{}, refusal("graph_path_denied")
			}
			var info os.FileInfo
			f, err := b.r.open(p.root, r, false)
			if err != nil {
				return stateread.Metadata{}, readError(err)
			}
			defer func() { _ = f.Close() }()
			info, err = f.Stat()
			return stateread.Metadata{Kind: stateread.KindPresent, Info: info}, err
		},
		ReadFile: func(full string) (stateread.File, error) {
			r, err := rel(full)
			if err != nil {
				return stateread.File{}, err
			}
			raw, err := b.r.Read(p.root, p.label, r)
			return stateread.File{Kind: stateread.KindPresent, Bytes: raw}, err
		},
	})
	if err != nil {
		return err
	}
	// A directory-scoped skill cache is compared with the resolver's projected
	// tree identity, preserving its directory ancestors and excluding siblings.
	if p.member.Kind == contextlock.KindSkill && p.member.Directory != "" && p.member.Directory != "." && strings.Contains(filepath.Base(p.root), ".dir-") {
		selected, err := g.treeAt(expected, p.member.Directory, len(p.member.Commit)/2)
		if err != nil {
			return err
		}
		expected, err = gitops.ProjectedTreeID(format, p.member.Directory, selected)
		if err != nil {
			return err
		}
	}
	if actual != expected {
		return refusal("graph_invalid_input")
	}
	return nil
}
