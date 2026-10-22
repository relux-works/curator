package projectcontext

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/hashing"
	sourceidentity "github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pkgversion"
)

type offlineSource struct {
	root         string
	dependencies map[string]string
	captures     map[string]captured
	manifests    map[string]*contextpkg.Manifest
}

func (s *offlineSource) repo(name, source string) (string, error) {
	p, ok := s.dependencies[name]
	if !ok {
		return "", fail("project_context_invalid", "missing explicit offline dependency: "+name)
	}
	r, err := os.OpenRoot(s.root)
	if err != nil {
		return "", err
	}
	defer r.Close()
	cur := ""
	for _, c := range strings.Split(p, "/") {
		cur = path.Join(cur, c)
		st, err := r.Lstat(cur)
		if err != nil || !st.IsDir() {
			return "", fail("project_context_invalid", "linked or missing dependency repository")
		}
	}
	dir := filepath.Join(s.root, filepath.FromSlash(p))
	b, err := bind(dir)
	if err != nil {
		return "", err
	}
	if b.Repository.PrimaryRemote == nil || *b.Repository.PrimaryRemote != source {
		return "", fail("project_context_invalid", "offline dependency source identity mismatch")
	}
	return dir, nil
}
func (s *offlineSource) Identity(kind, name, declared string) (string, error) {
	if kind != contextlock.KindContext {
		return "", fail("project_context_unsupported", "context cannot admit skill or MCP dependencies")
	}
	v := declared
	if !sourceidentity.ValidCanonical(v) {
		var err error
		v, err = sourceidentity.Parse(v)
		if err != nil {
			return "", err
		}
	}
	if !sourceidentity.ValidCanonical(v) {
		return "", fail("project_context_invalid", "dependency requires canonical network identity")
	}
	_, err := s.repo(name, v)
	return v, err
}
func (s *offlineSource) ResolveTag(kind, name, source, tag string) (string, error) {
	dir, err := s.repo(name, source)
	if err != nil {
		return "", err
	}
	v, err := gitLocal(dir, "rev-parse", "--verify", "refs/tags/"+tag+"^{commit}")
	if err != nil || !contextCommit(v) {
		return "", fail("project_context_invalid", "offline tag is unavailable")
	}
	return v, nil
}
func (s *offlineSource) Candidates(kind, name, source string) ([]contextresolve.Candidate, error) {
	dir, err := s.repo(name, source)
	if err != nil {
		return nil, err
	}
	tags, err := gitLocal(dir, "tag", "--list")
	if err != nil {
		return nil, err
	}
	out := []contextresolve.Candidate{}
	for _, t := range strings.Split(tags, "\n") {
		v, ok := pkgversion.ParseTag(t)
		if !ok {
			continue
		}
		commit, err := s.ResolveTag(kind, name, source, t)
		if err != nil {
			return nil, err
		}
		out = append(out, contextresolve.Candidate{Tag: t, Version: v, Commit: commit})
	}
	return out, nil
}
func captureCommit(dir, commit, directory string) (captured, error) {
	if !contextCommit(commit) {
		return captured{}, fail("project_context_invalid", "dependency must resolve to an exact commit")
	}
	if directory != "" && !relative(directory) {
		return captured{}, fail("project_context_invalid", "invalid selected dependency directory")
	}
	// Enumerate before reading any blob. No checkout, hooks, fetch, install or
	// executable package parser runs in this offline admission path.
	listing, err := gitLocal(dir, "ls-tree", "-r", "-l", "-z", commit)
	if err != nil {
		return captured{}, err
	}
	files := map[string][]byte{}
	modes := map[string]string{}
	total := int64(0)
	count := 0
	for _, line := range strings.Split(listing, "\x00") {
		if line == "" {
			continue
		}
		count++
		if count > MaxFiles {
			return captured{}, fail("project_context_invalid", "Git snapshot file bound exceeded")
		}
		meta, p, ok := strings.Cut(line, "\t")
		v := strings.Fields(meta)
		if !ok || len(v) != 4 || !relative(p) || v[1] != "blob" || (v[0] != "100644" && v[0] != "100755") {
			return captured{}, fail("project_context_invalid", "linked, special or malformed Git tree entry")
		}
		size, err := strconv.ParseInt(v[3], 10, 64)
		if err != nil || size < 0 || size > MaxFileBytes {
			return captured{}, fail("project_context_invalid", "oversized Git blob")
		}
		total += size
		if total > MaxTotalBytes {
			return captured{}, fail("project_context_invalid", "Git snapshot byte bound exceeded")
		}
		selected := p
		if directory != "" {
			if !strings.HasPrefix(p, directory+"/") {
				continue
			}
			selected = strings.TrimPrefix(p, directory+"/")
		}
		b, err := gitBytes(dir, MaxFileBytes, "cat-file", "blob", v[2])
		if err != nil || int64(len(b)) != size {
			return captured{}, fail("project_context_invalid", "malformed Git blob read")
		}
		files[selected] = b
		modes[selected] = "0600"
		if v[0] == "100755" {
			modes[selected] = "0700"
		}
	}
	c := fromFiles(files)
	for i := range c.Manifest.Files {
		c.Manifest.Files[i].Mode = modes[c.Manifest.Files[i].Path]
	}
	if len(files) == 0 {
		return c, fail("project_context_invalid", "empty or missing selected dependency")
	}
	if err := validateInventory(c); err != nil {
		return c, err
	}
	return c, nil
}
func (s *offlineSource) Manifest(kind, name, source, directory, commit string) (*contextresolve.Package, error) {
	dir, err := s.repo(name, source)
	if err != nil {
		return nil, err
	}
	c, err := captureCommit(dir, commit, directory)
	if err != nil {
		return nil, err
	}
	m, err := manifest(c)
	if err != nil {
		return nil, err
	}
	if m.Name != name {
		return nil, fail("project_context_invalid", "dependency manifest identity mismatch")
	}
	s.captures[name] = c
	s.manifests[name] = m
	return packageOf(m), nil
}
func packageOf(m *contextpkg.Manifest) *contextresolve.Package {
	p := &contextresolve.Package{SchemaVersion: m.SchemaVersion, Version: m.Version, Weight: m.Weight, Weights: m.Weights}
	for _, n := range contextpkg.SortedNames(m.Contexts) {
		r := m.Contexts[n]
		p.Requires = append(p.Requires, contextresolve.Requirement{Kind: "context", Name: n, Source: sourceidentity.Canonical(r.Git), Range: r.Range, Tag: r.Tag, Revision: r.Revision, Directory: r.Directory, Weight: r.Weight})
	}
	return p
}
func manifest(c captured) (*contextpkg.Manifest, error) {
	b, ok := c.Bytes[contextpkg.ManifestName]
	if !ok {
		return nil, fail("project_context_invalid", "context package requires agent-context.json")
	}
	m, err := contextpkg.ParseManifest(b)
	if err != nil {
		return nil, err
	}
	if !m.HasContext || len(m.Skills) != 0 || len(m.MCP) != 0 {
		return nil, fail("project_context_unsupported", "context requires explicit modules and context-only dependencies")
	}
	if len(c.Bytes["config/muse/auth.json"]) > 0 {
		return nil, fail("project_context_invalid", "authentication state is not context")
	}
	for _, mod := range m.Modules {
		b, ok := c.Bytes[path.Join("context", mod.Path)]
		if !ok {
			return nil, fail("project_context_invalid", "selected module missing from snapshot")
		}
		if err := contextpkg.ValidateModuleBytes(b); err != nil {
			return nil, err
		}
		if !noNativeImports(b) {
			return nil, fail("project_context_unsupported", "native imports have no leaf-1 dialect")
		}
	}
	if report := contextaudit.DetectFilesAtVersion(c.Bytes, "", hashing.VersionV2, nil); report.Blocking() {
		return nil, fail("project_context_invalid", "context audit blocked candidate")
	}
	return m, nil
}
func (op *operation) context(r registered, req Request, synthetic map[string][]byte) (Candidate, error) {
	root, err := os.OpenRoot(r.Record.ProjectRoot)
	if err != nil {
		return Candidate{}, err
	}
	defer root.Close()
	c, err := capture(root, req.Source)
	if synthetic != nil {
		c = fromFiles(synthetic)
		err = nil
	}
	if err != nil {
		return Candidate{}, err
	}
	if err = validateInventory(c); err != nil {
		return Candidate{}, err
	}
	m, err := manifest(c)
	if err != nil {
		return Candidate{}, err
	}
	pin, err := hashing.ContentSHA256Files(c.Bytes, hashing.VersionV2)
	if err != nil {
		return Candidate{}, err
	}
	_, base, err := profile(filepath.Dir(op.root), r.Record.Profile.Name)
	if err != nil {
		return Candidate{}, err
	}
	if identity(base.Object()) != r.Record.Profile.LockSHA256 {
		return Candidate{}, fail("project_context_drift", "selected base profile lock changed; explicitly rebind")
	}
	src := &offlineSource{r.Record.ProjectRoot, req.Dependencies, map[string]captured{m.Name: c}, map[string]*contextpkg.Manifest{m.Name: m}}
	result, err := contextresolve.Resolve(src, contextresolve.Input{Precedence: r.Precedence, Root: contextresolve.Requirement{Kind: "context", Name: m.Name}, RootState: &contextresolve.StatePackage{StateHash: strings.TrimPrefix(pin, "sha256:"), Manifest: packageOf(m)}})
	if err != nil {
		return Candidate{}, err
	}
	result.Lock.SchemaVersion = contextlock.SchemaVersion4
	result.Lock.HashVersion = 2
	result.Lock.Precedence = r.Precedence
	if err = result.Lock.Validate(); err != nil {
		return Candidate{}, err
	}
	lockBytes, err := result.Lock.Canonical()
	if err != nil {
		return Candidate{}, err
	}
	packages := map[string]contextmaterialize.Package{}
	members := []Member{}
	for _, member := range result.Lock.Members {
		if member.Commit != "" {
			dir, e := src.repo(member.Name, member.Source)
			if e != nil {
				return Candidate{}, e
			}
			capt, e := captureCommit(dir, member.Commit, member.Directory)
			if e != nil {
				return Candidate{}, e
			}
			man, e := manifest(capt)
			if e != nil {
				return Candidate{}, e
			}
			src.captures[member.Name] = capt
			src.manifests[member.Name] = man
		}
		capt, ok := src.captures[member.Name]
		if !ok || member.Kind != "context" {
			return Candidate{}, fail("project_context_invalid", "unresolved context closure member")
		}
		man := src.manifests[member.Name]
		p := contextmaterialize.Package{HasContext: man.HasContext}
		for _, mod := range man.Modules {
			p.Modules = append(p.Modules, contextmaterialize.Module{Module: mod, Bytes: capt.Bytes[path.Join("context", mod.Path)]})
		}
		packages[member.Name] = p
		ref, h, err := op.put(capt)
		if err != nil {
			return Candidate{}, err
		}
		members = append(members, Member{member.Name, ref, h, member.Pin()})
	}
	order := contextmaterialize.Precedence{Winner: r.Precedence.Winner, Placement: r.Precedence.Placement}
	_, written, err := contextmaterialize.Monolithic(result.Lock, "sha256:"+digest(lockBytes), order, req.Environment, packages)
	if err != nil {
		return Candidate{}, err
	}
	if !written {
		return Candidate{}, fail("project_context_unsupported", "context closure has no applicable root module")
	}
	if _, _, _, err = contextmaterialize.SystemPrompt(result.Lock, order, req.Environment, packages, contextmaterialize.Admission{}); err != nil {
		return Candidate{}, err
	}
	sort.Slice(members, func(i, j int) bool { return members[i].Name < members[j].Name })
	ref, h, err := op.put(c)
	if err != nil {
		return Candidate{}, err
	}
	lockRef, lockHash, err := op.putLock(lockBytes)
	if err != nil {
		return Candidate{}, err
	}
	item := Item{"project-item/1", req.ID, "context", req.Environment, req.Required, Payload{ref, h, lockRef, lockHash}}
	return Candidate{item, identity(item), req, identity(c.Manifest), members, AppendAfterProfile}, nil
}
func validateInventory(c captured) error {
	if len(c.Manifest.Files) > MaxFiles {
		return fail("project_context_invalid", "snapshot file bound exceeded")
	}
	total := int64(0)
	aliases := map[string]bool{}
	for _, e := range c.Manifest.Files {
		if !relative(e.Path) || e.Size < 0 || e.Size > MaxFileBytes || !hexRE.MatchString(e.SHA256) || (e.Mode != "0600" && e.Mode != "0700") {
			return fail("project_context_invalid", "invalid snapshot inventory")
		}
		key := portableAlias(e.Path)
		if aliases[key] {
			return fail("project_context_invalid", "case or normalization alias")
		}
		aliases[key] = true
		for p := path.Dir(key); p != "."; p = path.Dir(p) {
			if aliases[p] {
				return fail("project_context_invalid", "file/ancestor collision")
			}
		}
		total += e.Size
		if total > MaxTotalBytes {
			return fail("project_context_invalid", "snapshot total byte bound exceeded")
		}
	}
	for p := range aliases {
		for parent := path.Dir(p); parent != "."; parent = path.Dir(parent) {
			if aliases[parent] {
				return fail("project_context_invalid", "file/ancestor alias collision")
			}
		}
	}

	return nil
}
