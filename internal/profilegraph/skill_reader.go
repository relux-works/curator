package profilegraph

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/skillspec"
)

// scopedSkill uses install's complete selected-skill validation, including
// command names and build/runtime layout, without staging an extraction.
func (b *build) scopedSkill(p packageInput, name string) (*contextresolve.Package, error) {
	var refusalErr error
	admit := func(err error) error {
		if isMissing(err) {
			return fs.ErrNotExist
		}
		if err != nil {
			refusalErr = err
		}
		return err
	}
	stat := func(rel string) (os.FileInfo, error) {
		rel = filepath.ToSlash(rel)
		if !cleanRelative(rel) || denied(rel) {
			return nil, admit(refusal("graph_path_denied"))
		}
		if p.root != "" {
			f, err := b.r.open(p.root, path.Join(p.prefix, rel), false, false)
			if err != nil {
				return nil, admit(readError(err))
			}
			defer func() { _ = f.Close() }()
			info, err := f.Stat()
			return info, admit(err)
		}
		kind, raw, err := b.db(p.member.Source).pathObject(p.member.Commit, path.Join(p.member.Directory, rel))
		if err != nil {
			return nil, admit(err)
		}
		mode := os.FileMode(0444)
		if kind == "tree" {
			mode |= os.ModeDir
		}
		if kind != "blob" && kind != "tree" {
			return nil, admit(refusal("graph_invalid_input"))
		}
		return objectInfo{rel, int64(len(raw)), mode}, nil
	}
	r := skillspec.Reader{Lstat: stat, Stat: stat, ReadFile: func(rel string) ([]byte, error) {
		raw, err := b.packageRead(p, filepath.ToSlash(rel))
		return raw, admit(err)
	}}
	skillName, spec, err := manifest.ReadSkillPackageWith(".", r)
	if refusalErr != nil {
		return nil, refusalErr
	}
	if err != nil || skillName != name {
		return nil, refusal("graph_invalid_input")
	}
	commands := map[string]bool{}
	for name := range spec.Commands {
		commands[name] = true
	}
	return &contextresolve.Package{SkillCommands: commands}, nil
}

type objectInfo struct {
	name string
	size int64
	mode os.FileMode
}

func (i objectInfo) Name() string       { return path.Base(i.name) }
func (i objectInfo) Size() int64        { return i.size }
func (i objectInfo) Mode() os.FileMode  { return i.mode }
func (i objectInfo) ModTime() time.Time { return time.Time{} }
func (i objectInfo) IsDir() bool        { return i.mode.IsDir() }
func (i objectInfo) Sys() any           { return nil }
