package envprofile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/pkgversion"
)

// gitManager is the contextresolve.Source over the profile git cache and
// path snapshots. Git repositories are cloned once below the manager home
// and fetched on install and update; manifests are read from object-database
// extractions, so their bytes are a function of the commit alone
// (environments §1.2).
type gitManager struct {
	home string
}

func newGitManager(home string) *gitManager { return &gitManager{home: home} }

// reposDir holds one clone per canonical git identity.
func (m *gitManager) reposDir() string { return filepath.Join(m.home, "profile-repos") }

// repoDir maps a canonical identity onto its clone path.
func (m *gitManager) repoDir(identity string) string {
	var builder strings.Builder
	for _, r := range identity {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('_')
		}
	}
	name := builder.String()
	if len(name) > 120 {
		name = name[len(name)-120:]
	}
	return filepath.Join(m.reposDir(), name)
}

// ensureRepo clones the identity when absent.
func (m *gitManager) ensureRepo(identity string) (string, error) {
	dir := m.repoDir(identity)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir, nil
	}
	if err := os.MkdirAll(m.reposDir(), 0o755); err != nil {
		return "", err
	}
	if err := gitops.Clone(identity, dir); err != nil {
		return "", fmt.Errorf("%s: clone %s: %v", DiagSourceInvalid, identity, err)
	}
	return dir, nil
}

// fetch updates the cached clone.
func (m *gitManager) fetch(identity string) error {
	dir, err := m.ensureRepo(identity)
	if err != nil {
		return err
	}
	if err := gitops.Fetch(dir); err != nil {
		return fmt.Errorf("fetch %s: %v", identity, err)
	}
	return nil
}

// Identity returns the canonical source identity.
func (m *gitManager) Identity(_, _ string, declared string) (string, error) {
	if declared == "" {
		return "", nil
	}
	return canonicalGit(declared), nil
}

// Candidates lists every version tag of the source peeled to its commit.
// Tags that do not parse as strict versions are silently outside every
// range (environments §1.4).
// A git source hosts exactly one package, so every version tag is a
// candidate of the requested name.
func (m *gitManager) Candidates(_ string, name, source string) ([]contextresolve.Candidate, error) {
	if source == "" {
		return nil, fmt.Errorf("%s: member %s names no source", DiagSourceInvalid, name)
	}
	dir, err := m.ensureRepo(source)
	if err != nil {
		return nil, err
	}
	out, err := exec.Command("git", "-C", dir, "tag", "--list").Output() // #nosec G204 -- dir is the manager cache
	if err != nil {
		return nil, fmt.Errorf("%s: list tags of %s: %v", DiagSourceInvalid, source, err)
	}
	var candidates []contextresolve.Candidate
	for _, tag := range strings.Split(string(out), "\n") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		version, ok := pkgversion.ParseTag(tag)
		if !ok {
			continue
		}
		resolved, err := gitops.Resolve(dir, "tag", tag)
		if err != nil {
			continue
		}
		candidates = append(candidates, contextresolve.Candidate{Tag: tag, Version: version, Commit: resolved.Commit})
	}
	return candidates, nil
}

// ResolveTag peels one exact tag to its commit.
func (m *gitManager) ResolveTag(_ string, name, source, tag string) (string, error) {
	dir, err := m.ensureRepo(source)
	if err != nil {
		return "", err
	}
	resolved, err := gitops.Resolve(dir, "tag", tag)
	if err != nil {
		return "", &contextresolve.Error{Diagnostic: contextresolve.DiagVersionMismatch, Name: name, Tag: tag}
	}
	return resolved.Commit, nil
}

// Manifest reads the package manifest at a commit, below directory when set.
func (m *gitManager) Manifest(kind, name, source, directory, commit string) (*contextresolve.Package, error) {
	dir, err := m.ensureRepo(source)
	if err != nil {
		return nil, err
	}
	snapshot, err := os.MkdirTemp("", "curator-profile-manifest-*")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(snapshot) }()
	if err := gitops.Extract(dir, commit, snapshot); err != nil {
		return nil, fmt.Errorf("%s: extract %s at %s: %v", DiagSourceInvalid, name, commit, err)
	}
	root := snapshot
	if directory != "" {
		root = filepath.Join(snapshot, filepath.FromSlash(directory))
	}
	switch kind {
	case contextlock.KindSkill:
		return &contextresolve.Package{}, nil
	case contextlock.KindMCP:
		manifest, err := contextpkg.LoadMCP(root)
		if err != nil {
			if os.IsNotExist(err) {
				return &contextresolve.Package{}, nil
			}
			return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		return &contextresolve.Package{Version: manifest.Version}, nil
	default:
		manifest, err := contextpkg.LoadManifest(root)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		return packageOf(manifest), nil
	}
}

// rootInput builds the resolution input for a fresh git install: it fetches
// the source, reads the root manifest at the required ref, and declares the
// root requirement as written.
func (m *gitManager) rootInput(gitURL, directory string, requirement Requirement) (contextresolve.Input, string, error) {
	identity := canonicalGit(gitURL)
	if err := m.fetch(identity); err != nil {
		return contextresolve.Input{}, "", err
	}
	dir := m.repoDir(identity)
	commit, tag, err := m.commitFor(dir, requirement)
	if err != nil {
		return contextresolve.Input{}, "", err
	}
	manifest, err := m.contextManifestAt(dir, commit, directory)
	if err != nil {
		return contextresolve.Input{}, "", err
	}
	root := contextresolve.Requirement{
		Kind: contextlock.KindContext, Name: manifest.Name, Source: identity,
		Range: requirement.Range, Tag: requirement.Tag, Revision: requirement.Revision,
		Directory: directory,
	}
	_ = tag
	return contextresolve.Input{Root: root}, manifest.Name, nil
}

// inputFor rebuilds the resolution input of an installed git profile from
// its stored requirement.
func (m *gitManager) inputFor(source Source) (contextresolve.Input, error) {
	input, _, err := m.rootInput(source.Git, source.Directory, source.Req)
	return input, err
}

// commitFor maps a declared requirement onto its commit. A range takes the
// highest satisfying version tag; latest spells *.
func (m *gitManager) commitFor(dir string, requirement Requirement) (string, string, error) {
	switch {
	case requirement.Revision != "":
		resolved, err := gitops.Resolve(dir, "revision", requirement.Revision)
		if err != nil {
			return "", "", &contextresolve.Error{Diagnostic: contextresolve.DiagVersionMismatch, Detail: err.Error()}
		}
		return resolved.Commit, "", nil
	case requirement.Tag != "":
		resolved, err := gitops.Resolve(dir, "tag", requirement.Tag)
		if err != nil {
			return "", "", &contextresolve.Error{Diagnostic: contextresolve.DiagVersionMismatch, Tag: requirement.Tag}
		}
		manifest, err := m.contextManifestAt(dir, resolved.Commit, "")
		if err != nil {
			return "", "", err
		}
		if "v"+manifest.Version != requirement.Tag {
			return "", "", &contextresolve.Error{
				Diagnostic:      contextresolve.DiagVersionMismatch,
				Tag:             requirement.Tag,
				ManifestVersion: manifest.Version,
			}
		}
		return resolved.Commit, requirement.Tag, nil
	default:
		parsed, err := pkgversion.ParseRange(requirement.Range)
		if err != nil {
			return "", "", fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		out, err := exec.Command("git", "-C", dir, "tag", "--list").Output() // #nosec G204 -- dir is the manager cache
		if err != nil {
			return "", "", fmt.Errorf("%s: list tags: %v", DiagSourceInvalid, err)
		}
		var versions []pkgversion.Version
		byVersion := map[string]string{}
		for _, tag := range strings.Split(string(out), "\n") {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			version, ok := pkgversion.ParseTag(tag)
			if !ok {
				continue
			}
			versions = append(versions, version)
			byVersion[version.String()] = tag
		}
		best, found := parsed.Highest(versions)
		if !found {
			return "", "", &contextresolve.Error{Diagnostic: contextresolve.DiagRangeConflict, Detail: "no version tag satisfies " + requirement.Range}
		}
		tag := byVersion[best.String()]
		resolved, err := gitops.Resolve(dir, "tag", tag)
		if err != nil {
			return "", "", fmt.Errorf("%s: %v", DiagSourceInvalid, err)
		}
		return resolved.Commit, tag, nil
	}
}

// contextManifestAt reads the context manifest at a commit.
func (m *gitManager) contextManifestAt(dir, commit, directory string) (*contextpkg.Manifest, error) {
	snapshot, err := os.MkdirTemp("", "curator-profile-manifest-*")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(snapshot) }()
	if err := gitops.Extract(dir, commit, snapshot); err != nil {
		return nil, fmt.Errorf("%s: extract at %s: %v", DiagSourceInvalid, commit, err)
	}
	root := snapshot
	if directory != "" {
		root = filepath.Join(snapshot, filepath.FromSlash(directory))
	}
	manifest, err := contextpkg.LoadManifest(root)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	return manifest, nil
}

// packageRoot joins a store entry with a member directory requirement to
// the package root every consumer of a resolved member must use: the audit
// scope, the manifest load, and the materialization load all read below the
// directory when one is declared. A third call site must reuse this helper
// rather than re-spelling the join, so directory-addressed packages cannot
// drift out of any consumer's scope (environments §9.1).
func packageRoot(entry, directory string) string {
	if directory != "" {
		return filepath.Join(entry, filepath.FromSlash(directory))
	}
	return entry
}

// entryPath returns the store entry path of a resolved member without
// creating it.
func (m *gitManager) entryPath(home string, resolved contextresolve.Resolved) string {
	if resolved.Commit != "" {
		return contextstore.EntryDir(home, resolved.Kind, resolved.Name, resolved.Commit)
	}
	return contextstore.EntryDir(home, resolved.Kind, resolved.Name, resolved.StateHash)
}

// ensureEntry installs the store entry of a resolved member and returns its
// path: commit members extract from the object database under their commit
// key (a directory requirement reads below that entry at use time), while
// state members are already in the store under their hash.
func (m *gitManager) ensureEntry(home string, resolved contextresolve.Resolved) (string, error) {
	if resolved.Commit != "" {
		return contextstore.EnsureGit(home, resolved.Kind, resolved.Name, m.repoDir(resolved.Source), resolved.Commit)
	}
	return contextstore.EntryDir(home, resolved.Kind, resolved.Name, resolved.StateHash), nil
}

// packageOf maps a validated context manifest onto the resolution package.
func packageOf(manifest *contextpkg.Manifest) *contextresolve.Package {
	pkg := &contextresolve.Package{Version: manifest.Version, Weight: manifest.Weight, Weights: manifest.Weights}
	for _, name := range contextpkg.SortedNames(manifest.Contexts) {
		requirement := manifest.Contexts[name]
		pkg.Requires = append(pkg.Requires, contextresolve.Requirement{
			Kind: contextlock.KindContext, Name: name, Source: requirement.Git,
			Range: requirement.Range, Tag: requirement.Tag, Revision: requirement.Revision,
			Directory: requirement.Directory, Weight: requirement.Weight,
		})
	}
	for _, name := range contextpkg.SortedNames(manifest.Skills) {
		requirement := manifest.Skills[name]
		pkg.Requires = append(pkg.Requires, contextresolve.Requirement{
			Kind: contextlock.KindSkill, Name: name, Source: requirement.Git,
			Range: requirement.Range, Tag: requirement.Tag, Revision: requirement.Revision,
		})
	}
	for _, name := range contextpkg.SortedNames(manifest.MCP) {
		requirement := manifest.MCP[name]
		pkg.Requires = append(pkg.Requires, contextresolve.Requirement{
			Kind: contextlock.KindMCP, Name: name, Source: requirement.Git,
			Range: requirement.Range, Tag: requirement.Tag, Revision: requirement.Revision,
			Directory: requirement.Directory,
		})
	}
	sort.Slice(pkg.Requires, func(i, j int) bool {
		if pkg.Requires[i].Name != pkg.Requires[j].Name {
			return pkg.Requires[i].Name < pkg.Requires[j].Name
		}
		return pkg.Requires[i].Kind < pkg.Requires[j].Kind
	})
	return pkg
}

// lookPath resolves an executable on PATH.
func lookPath(command string) (string, error) { return exec.LookPath(command) }
