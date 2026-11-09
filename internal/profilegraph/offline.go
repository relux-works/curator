package profilegraph

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pkgversion"
)

type offlineSource struct{ b *build }

func (s offlineSource) Identity(_, _, declared string) (string, error) {
	if err := s.b.r.inspect(); err != nil {
		return "", err
	}
	if identity.ValidCanonical(declared) {
		return declared, nil
	}
	id, err := identity.Parse(declared)
	if err != nil || id == "" {
		return "", refusal("graph_invalid_input")
	}
	return id, nil
}
func (s offlineSource) Candidates(_, _, source string) ([]contextresolve.Candidate, error) {
	return s.b.gitCandidates(source)
}
func (s offlineSource) ResolveTag(_, _, source, tag string) (string, error) {
	return s.b.gitTag(source, tag)
}
func (s offlineSource) Manifest(kind, name, source, directory, commit string) (*contextresolve.Package, error) {
	if err := s.b.r.inspect(); err != nil {
		return nil, err
	}
	if len(s.b.packages) >= s.b.r.Limits.Nodes {
		return nil, refusal("graph_limit_exceeded")
	}
	m := contextlock.Member{Kind: kind, Name: name, Source: source, Directory: directory, Commit: commit}
	p := packageInput{root: envprofile.GraphMemberEntry(s.b.req.Home, m), label: "snapshot/" + kind + "/" + name + "/" + commit, prefix: directory, member: m}
	if err := s.b.r.bind(p.root, s.b.req.Home); err != nil {
		return nil, err
	}
	if directory == "." {
		p.prefix = ""
	}
	file := contextpkg.ManifestName
	if kind == contextlock.KindSkill {
		file = "SKILL.md"
	}
	if kind == contextlock.KindMCP {
		file = contextpkg.MCPManifestName
	}
	raw, err := p.read(s.b.r, file)
	if isMissing(err) {
		// Local Git objects are read directly. No git process, config, hook,
		// credential helper, lazy fetch or repository lock can run.
		raw, err = s.b.gitFile(source, commit, path.Join(directory, file))
		if err == nil {
			p.root = ""
			p.label = "git/" + source + "/" + commit
		}
	}
	if err != nil {
		return nil, err
	}
	s.b.packages[name] = p
	switch kind {
	case contextlock.KindContext:
		manifest, err := contextpkg.ParseManifest(raw)
		if err != nil {
			return nil, err
		}
		if manifest.Name != name && name != "" {
			return nil, refusal("graph_invalid_input")
		}
		p.manifest = manifest
		s.b.packages[manifest.Name] = p
		if err := s.b.checkKnownDAG(true); err != nil {
			return nil, err
		}
		return envprofile.GraphPackage(manifest), nil
	case contextlock.KindSkill:
		if directory == "" {
			return &contextresolve.Package{}, nil
		}
		return s.b.scopedSkill(p, name)
	default:
		manifest, err := contextpkg.ParseMCP(raw)
		if err != nil {
			return nil, err
		}
		return &contextresolve.Package{Version: manifest.Version}, nil
	}
}
func (b *build) previewOverlay(slot int, o envprofile.OverlaySpec) (contextresolve.Overlay, error) {
	if identity.ClassifySource(o.Source) == identity.SourcePath {
		if o.Directory != "" || o.Tag != "" || o.Range != "" || o.Revision != "" {
			return contextresolve.Overlay{}, refusal("graph_invalid_input")
		}
		root, err := filepath.Abs(o.Source)
		if err != nil {
			return contextresolve.Overlay{}, err
		}
		if err := admittedTree(root); err != nil {
			return contextresolve.Overlay{}, err
		}
		label := "overlay/" + itoa(slot)
		raw, err := b.r.Read(root, label, contextpkg.ManifestName)
		if err != nil {
			return contextresolve.Overlay{}, err
		}
		m, err := contextpkg.ParseManifest(raw)
		if err != nil {
			return contextresolve.Overlay{}, err
		}
		files := map[string][]byte{}
		if err := b.sourceTree(root, label, ".", files, 0); err != nil {
			return contextresolve.Overlay{}, err
		}
		h, err := hashing.ContentSHA256Files(files, hashing.WriteVersion())
		if err != nil {
			return contextresolve.Overlay{}, err
		}
		b.packages[m.Name] = packageInput{root: root, label: label, manifest: m}
		return contextresolve.Overlay{Name: m.Name, Source: o.Source, Weight: o.Weight, State: &contextresolve.StatePackage{StateHash: strings.TrimPrefix(h, "sha256:"), Manifest: envprofile.GraphPackage(m)}}, nil
	}
	source, err := (offlineSource{b}).Identity("", "", o.Source)
	if err != nil {
		return contextresolve.Overlay{}, err
	}
	commit := o.Revision
	if commit == "" && o.Tag != "" {
		commit, err = b.gitTag(source, o.Tag)
	} else if commit == "" {
		candidates, e := b.gitCandidates(source)
		if e != nil {
			return contextresolve.Overlay{}, e
		}
		constraint := o.Range
		if constraint == "" {
			constraint = "latest"
		}
		rng, e := pkgversion.ParseRange(constraint)
		if e != nil {
			return contextresolve.Overlay{}, refusal("graph_invalid_input")
		}
		for _, c := range candidates {
			if rng.Satisfies(c.Version) {
				commit = c.Commit
				break
			}
		}
		if commit == "" {
			return contextresolve.Overlay{}, refusal("graph_missing_input")
		}
	}
	if err != nil {
		return contextresolve.Overlay{}, err
	}
	raw, err := b.gitFile(source, commit, path.Join(o.Directory, contextpkg.ManifestName))
	if err != nil {
		return contextresolve.Overlay{}, err
	}
	manifest, err := contextpkg.ParseManifest(raw)
	if err != nil {
		return contextresolve.Overlay{}, err
	}
	b.packages[manifest.Name] = packageInput{label: "git/" + source + "/" + commit, prefix: o.Directory, manifest: manifest}
	return contextresolve.Overlay{Name: manifest.Name, Source: source, Directory: o.Directory, Revision: o.Revision, Range: o.Range, Tag: o.Tag, Weight: o.Weight}, nil
}
func (b *build) packageRead(p packageInput, rel string) ([]byte, error) {
	if p.root != "" {
		return p.read(b.r, rel)
	}
	return b.gitFile(p.member.Source, p.member.Commit, path.Join(p.member.Directory, rel))
}
