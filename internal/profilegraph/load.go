package profilegraph

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/stateread"
)

// Request contains explicit selectors, never a request to discover arbitrary
// home content. Source previews need no config or installed profile.
type Request struct {
	Profile, Source, Home, Environment, Version string
	Reader                                      *Reader
	Projection                                  Projection
}

// Projection is the restricted composition/environment selection. No provider,
// authentication, passthrough, telemetry or unrelated machine settings enter it.
type Projection struct {
	SystemPromptFiles string                        `json:"system_prompt_files"`
	Precedence        contextmaterialize.Precedence `json:"precedence"`
	Forms             map[string]string             `json:"forms"`
	Overlays          []envprofile.OverlaySpec      `json:"overlays"`
	Admission         contextmaterialize.Admission  `json:"system_modules"`
}

type packageInput struct {
	root, label, prefix string
	manifest            *contextpkg.Manifest
	member              contextlock.Member
}

func (p packageInput) read(r *Reader, rel string) ([]byte, error) {
	return r.Read(p.root, p.label, path.Join(p.prefix, rel))
}

type build struct {
	err      error
	source   envprofile.Source
	req      Request
	r        *Reader
	g        Graph
	lock     *contextlock.Lock
	lockHash string
	packages map[string]packageInput
	material map[string]contextmaterialize.Package
	layers   map[string]*Node
	modules  map[string]*Node
	nodes    map[string]*Node
	profile  *Node
	objects  map[string]gitObject
}

// Build resolves admitted local inputs and returns either one graph or nil.
func Build(req Request) (*Graph, error) {
	if !identifiers.Valid(req.Profile) {
		return nil, refusal("graph_invalid_input")
	}
	if req.Reader == nil {
		req.Reader = NewReader(V1Limits)
	}
	b := &build{req: req, r: req.Reader, packages: map[string]packageInput{}, material: map[string]contextmaterialize.Package{}, layers: map[string]*Node{}, modules: map[string]*Node{}, nodes: map[string]*Node{}}
	b.g.SchemaVersion = SchemaVersion
	b.g.Generator.Name = "curator"
	b.g.Generator.Version = req.Version
	b.g.Scope.Profile = req.Profile
	b.g.Scope.Mode = "installed"
	if req.Source != "" {
		b.g.Scope.Mode = "source"
	}
	b.g.Inputs = []Input{}
	b.g.Nodes = []*Node{}
	b.g.Edges = []Edge{}
	b.g.Documents = []Document{}
	b.g.Diagnostics = []Diagnostic{}
	if req.Environment != "" {
		id := envregistry.NormalizeEnvID(req.Environment)
		if _, err := envregistry.ByID(id); err != nil {
			return nil, refusal("graph_invalid_input")
		}
		b.g.Scope.Environments = []string{id}
	} else {
		for _, a := range envregistry.Registry {
			b.g.Scope.Environments = append(b.g.Scope.Environments, a.ID)
		}
		sort.Strings(b.g.Scope.Environments)
	}
	if len(b.g.Scope.Environments) > b.r.Limits.Environments {
		return nil, refusal("graph_limit_exceeded")
	}
	if req.Source != "" {
		if err := b.loadSource(); err != nil {
			return nil, readError(err)
		}
	} else {
		if err := b.loadInstalled(); err != nil {
			return nil, readError(err)
		}
	}
	if err := b.loadPackages(); err != nil {
		return nil, readError(err)
	}
	if err := b.checkDAG(); err != nil {
		return nil, err
	}
	for _, p := range b.packages {
		if p.member.StateHash != "" {
			if err := b.verifyState(p); err != nil {
				return nil, readError(err)
			}
		} else if p.member.Commit != "" {
			if err := b.verifyGit(p); err != nil {
				return nil, readError(err)
			}
		}
	}
	if err := b.graphNodes(); err != nil {
		return nil, readError(err)
	}
	if err := b.render(); err != nil {
		return nil, readError(err)
	}
	if err := b.finish(); err != nil {
		return nil, readError(err)
	}
	return &b.g, nil
}

func (b *build) loadInstalled() error {
	if b.req.Home == "" {
		return refusal("graph_missing_input")
	}
	read := func(full string) (stateread.File, error) {
		rel, err := filepath.Rel(b.req.Home, full)
		if err != nil {
			return stateread.File{}, refusal("graph_invalid_input")
		}
		raw, err := b.r.Read(b.req.Home, "managed", filepath.ToSlash(rel))
		return stateread.File{Kind: stateread.KindPresent, Bytes: raw}, err
	}
	source, lock, h, err := envprofile.GraphInputs(b.req.Home, b.req.Profile, read)
	if err != nil {
		return err
	}
	b.lock, b.lockHash, b.source = lock, h, source
	for _, m := range lock.Members {
		entry := envprofile.GraphMemberEntry(b.req.Home, m)
		if err := b.r.bind(entry, b.req.Home); err != nil {
			return err
		}
		// State snapshots are already selected packages; Git entries retain the
		// selected directory within the immutable repository snapshot.
		prefix := ""
		if m.Commit != "" && m.Directory != "." {
			prefix = m.Directory
		}
		b.packages[m.Name] = packageInput{root: entry, label: "snapshot/" + m.Kind + "/" + m.Name + "/" + m.PinKey(), prefix: prefix, member: m}
	}
	return nil
}

func (b *build) loadSource() error {
	if err := admittedTree(b.req.Source); err != nil {
		return err
	}
	raw, err := b.r.Read(b.req.Source, "source", contextpkg.ManifestName)
	if err != nil {
		return err
	}
	manifest, err := contextpkg.ParseManifest(raw)
	if err != nil {
		return err
	}
	// Compute install's state identity over the bounded source tree, without
	// publishing a snapshot. The .git control directory is never content.
	files := map[string][]byte{}
	if err := b.sourceTree(b.req.Source, "source", ".", files, 0); err != nil {
		return err
	}
	state, err := hashing.ContentSHA256Files(files, hashing.WriteVersion())
	if err != nil {
		return err
	}
	state = strings.TrimPrefix(state, "sha256:")
	b.packages[manifest.Name] = packageInput{root: b.req.Source, label: "source", manifest: manifest}
	input := contextresolve.Input{Root: contextresolve.Requirement{Name: manifest.Name, Kind: contextlock.KindContext}, RootState: &contextresolve.StatePackage{StateHash: state, Manifest: envprofile.GraphPackage(manifest)}}
	p := b.req.Projection.Precedence
	input.Precedence = contextlock.Precedence{Winner: p.Winner, Placement: p.Placement}
	// Explicit machine overlays use the same resolver seed semantics as install.
	for i, o := range b.req.Projection.Overlays {
		overlay, err := b.previewOverlay(i, o)
		if err != nil {
			return err
		}
		input.Overlays = append(input.Overlays, overlay)
	}
	resolved, err := contextresolve.Resolve(offlineSource{b}, input)
	if err != nil {
		return err
	}
	b.lock, b.lockHash = resolved.Lock, resolved.LockHash
	for _, m := range b.lock.Members {
		p := b.packages[m.Name]
		p.member = m
		b.packages[m.Name] = p
	}
	return nil
}
func (b *build) sourceTree(root, label, rel string, files map[string][]byte, depth int) error {
	if depth > b.r.Limits.Depth {
		return refusal("graph_limit_exceeded")
	}
	entries, err := b.r.directory(root, rel)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" {
			if rel != "." {
				return refusal("graph_invalid_input")
			}
			continue
		}
		p := path.Join(rel, name)
		if entry.IsDir() {
			if err := b.sourceTree(root, label, p, files, depth+1); err != nil {
				return err
			}
			continue
		}
		raw, err := b.r.Read(root, label, p)
		if err != nil {
			return err
		}
		files[p] = raw
	}
	return nil
}

func (b *build) loadPackages() error {
	for _, m := range b.lock.Members {
		p, ok := b.packages[m.Name]
		if !ok {
			return refusal("graph_missing_input")
		}
		p.member = m
		if m.Kind == contextlock.KindContext {
			raw, err := b.packageRead(p, contextpkg.ManifestName)
			if err != nil {
				return err
			}
			manifest, err := contextpkg.ParseManifest(raw)
			if err != nil {
				return err
			}
			if manifest.Name != m.Name || manifest.Version != m.Version {
				return refusal("graph_invalid_input")
			}
			p.manifest = manifest
			pkg := contextmaterialize.Package{HasContext: manifest.HasContext}
			if len(manifest.Modules) > b.r.Limits.Nodes {
				return refusal("graph_limit_exceeded")
			}
			for _, module := range manifest.Modules {
				raw, err := b.packageRead(p, path.Join(contextpkg.ContextDir, module.Path))
				if err != nil {
					return err
				}
				if err := contextpkg.ValidateModuleBytes(raw); err != nil {
					return err
				}
				pkg.Modules = append(pkg.Modules, contextmaterialize.Module{Module: module, Bytes: raw})
			}
			b.material[m.Name] = pkg
		}
		b.packages[m.Name] = p
	}
	return nil
}

func (b *build) checkDAG() error { return b.checkKnownDAG(false) }
func (b *build) checkKnownDAG(partial bool) error {
	visiting, heights := map[string]bool{}, map[string]int{}
	var visit func(string, int) (int, error)
	visit = func(name string, depth int) (int, error) {
		if depth > b.r.Limits.Depth {
			return 0, refusal("graph_limit_exceeded")
		}
		if visiting[name] {
			return 0, refusal("graph_cycle")
		}
		if height, done := heights[name]; done {
			if depth+height > b.r.Limits.Depth {
				return 0, refusal("graph_limit_exceeded")
			}
			return height, nil
		}
		p, ok := b.packages[name]
		if !ok {
			if partial {
				return 0, nil
			}
			return 0, refusal("graph_missing_input")
		}
		visiting[name] = true
		height := 0
		if p.manifest != nil {
			for _, req := range envprofile.GraphPackage(p.manifest).Requires {
				h, err := visit(req.Name, depth+1)
				if err != nil {
					return 0, err
				}
				height = max(height, h+1)
			}
		}
		delete(visiting, name)
		heights[name] = height
		return height, nil
	}
	for name := range b.packages {
		if _, err := visit(name, 0); err != nil {
			return err
		}
	}
	return nil
}

// Explicit selection does not authorize recursive inspection of a home or
// filesystem root. Package subdirectories and immutable snapshot roots do.
func admittedTree(root string) error {
	full, err := filepath.Abs(root)
	if err != nil {
		return refusal("graph_invalid_input")
	}
	home, _ := os.UserHomeDir()
	if full == filepath.Dir(full) || (home != "" && filepath.Clean(full) == filepath.Clean(home)) {
		return refusal("graph_path_denied")
	}
	return nil
}

func isMissing(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Class == "graph_missing_input"
}
