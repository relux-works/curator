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
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pkgversion"
)

// gitManager is the contextresolve.Source over the profile git cache and
// path snapshots. Git repositories are cloned once below the manager home
// and fetched on install and update; manifests are read from object-database
// extractions, so their bytes are a function of the commit alone
// (environments §1.2).
//
// The cache is keyed by the core §6.1 canonical source identity
// (environments §1): SSH and HTTPS spellings of one repository share one
// clone. fetchRaw remembers the raw clone URL each canonical identity was
// first seen under so ensureRepo can clone through the operator's transport
// (including git insteadOf rewrites, which hermetic tests use to serve fake
// network identities from local repositories); a canonical identity with no
// recorded raw clones through https://canonical. A transitive requirement
// declared git@host:org/dep is canonicalized before any raw is recorded and
// therefore clones over https://, which can surprise an SSH-only operator.
// File:// remotes are rejected at the identity boundary (see canonicalGit)
// and never reach this cache.
type gitManager struct {
	home           string
	fetchRaw       map[string]string
	allowedSources []string
}

func newGitManager(home string) *gitManager {
	return &gitManager{home: home, fetchRaw: map[string]string{}}
}

// withPolicy attaches the machine source allowlist for the operation.
// An empty allowlist permits every source (core §6.1).
func (m *gitManager) withPolicy(policy Policy) *gitManager {
	m.allowedSources = policy.AllowedSources
	return m
}

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

// ensureRepo clones the identity when absent. Network canonical identities
// clone through their recorded raw URL when one exists, else through
// https://canonical.
//
// The core §6.1 source allowlist is enforced before any network clone
// (environments §9.1, reusing closure.gateSource's shape): a git member
// whose canonical identity is outside the machine's allowed_sources is
// refused with profile_source_invalid. Empty identities (local sources)
// bypass the allowlist; an empty allowlist permits all.
func (m *gitManager) ensureRepo(canonical string) (string, error) {
	if err := m.gateSource(canonical); err != nil {
		return "", err
	}
	dir := m.repoDir(canonical)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir, nil
	}
	if err := os.MkdirAll(m.reposDir(), 0o755); err != nil {
		return "", err
	}
	cloneURL := canonical
	if raw, ok := m.fetchRaw[canonical]; ok {
		cloneURL = raw
	} else if identity.ValidCanonical(canonical) {
		cloneURL = "https://" + canonical
	}
	if err := gitops.Clone(cloneURL, dir); err != nil {
		return "", fmt.Errorf("%s: clone %s: %v", DiagSourceInvalid, canonical, err)
	}
	return dir, nil
}

// gateSource applies the machine allowlist before any network clone. A
// file:// source is rejected here as well: it carries no network identity
// and must never reach a clone.
func (m *gitManager) gateSource(canonical string) error {
	trimmed := strings.TrimSpace(canonical)
	if trimmed == "" {
		return nil
	}
	if isFileRemote(trimmed) {
		return fmt.Errorf("%s: file:// git sources carry no network identity and are not accepted", DiagSourceInvalid)
	}
	if len(m.allowedSources) == 0 {
		return nil
	}
	if !identity.ValidCanonical(trimmed) {
		parsed, err := identity.Parse(trimmed)
		if err != nil || parsed == "" {
			return nil
		}
		trimmed = parsed
	}
	if identity.Allowed(trimmed, m.allowedSources) {
		return nil
	}
	return fmt.Errorf("%s: source %s is outside the machine's allowed sources", DiagSourceInvalid, canonical)
}

// recordRaw remembers the raw clone URL for a canonical identity so later
// clones of the same identity reuse the operator's transport. First spelling
// wins; every spelling of one repository resolves to the same bytes.
func (m *gitManager) recordRaw(canonical, raw string) {
	if m.fetchRaw == nil {
		m.fetchRaw = map[string]string{}
	}
	trimmed := strings.TrimSpace(raw)
	if canonical == "" || trimmed == "" || trimmed == canonical {
		return
	}
	if identity.ValidCanonical(trimmed) {
		return
	}
	if _, exists := m.fetchRaw[canonical]; !exists {
		m.fetchRaw[canonical] = trimmed
	}
}

// fetch updates the cached clone.
func (m *gitManager) fetch(canonical string) error {
	dir, err := m.ensureRepo(canonical)
	if err != nil {
		return err
	}
	if err := gitops.Fetch(dir); err != nil {
		return fmt.Errorf("fetch %s: %v", canonical, err)
	}
	return nil
}

// Identity returns the core §6.1 canonical source identity for a declared
// git URL. A malformed network source — and a file:// remote, which carries
// no network identity — is rejected with profile_source_invalid. Every
// successful call records the raw-to-canonical mapping so later clones reuse
// the operator's transport.
func (m *gitManager) Identity(_, _ string, declared string) (string, error) {
	if strings.TrimSpace(declared) == "" {
		return "", nil
	}
	canonical, err := canonicalGit(declared)
	if err != nil {
		return "", fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	m.recordRaw(canonical, declared)
	return canonical, nil
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
// root requirement under its canonical source identity. A malformed network
// source is rejected with profile_source_invalid.
func (m *gitManager) rootInput(gitURL, directory string, requirement Requirement) (contextresolve.Input, string, error) {
	identity, err := canonicalGit(gitURL)
	if err != nil {
		return contextresolve.Input{}, "", fmt.Errorf("%s: %v", DiagSourceInvalid, err)
	}
	m.recordRaw(identity, gitURL)
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
// Every requirement source is canonicalized through identity.Parse at this
// boundary (environments §1): two spellings of one repository enter
// resolution as one identity, so they never produce a spurious
// context_source_mismatch. A malformed network source passes through here
// and is rejected at the Identity boundary with profile_source_invalid, as
// is a file:// requirement source, which carries no network identity.
func packageOf(manifest *contextpkg.Manifest) *contextresolve.Package {
	pkg := &contextresolve.Package{Version: manifest.Version, Weight: manifest.Weight, Weights: manifest.Weights}
	for _, name := range contextpkg.SortedNames(manifest.Contexts) {
		requirement := manifest.Contexts[name]
		pkg.Requires = append(pkg.Requires, contextresolve.Requirement{
			Kind: contextlock.KindContext, Name: name, Source: canonicalRequirementSource(requirement.Git),
			Range: requirement.Range, Tag: requirement.Tag, Revision: requirement.Revision,
			Directory: requirement.Directory, Weight: requirement.Weight,
		})
	}
	for _, name := range contextpkg.SortedNames(manifest.Skills) {
		requirement := manifest.Skills[name]
		pkg.Requires = append(pkg.Requires, contextresolve.Requirement{
			Kind: contextlock.KindSkill, Name: name, Source: canonicalRequirementSource(requirement.Git),
			Range: requirement.Range, Tag: requirement.Tag, Revision: requirement.Revision,
		})
	}
	for _, name := range contextpkg.SortedNames(manifest.MCP) {
		requirement := manifest.MCP[name]
		pkg.Requires = append(pkg.Requires, contextresolve.Requirement{
			Kind: contextlock.KindMCP, Name: name, Source: canonicalRequirementSource(requirement.Git),
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

// canonicalRequirementSource canonicalizes one manifest requirement source
// for resolution. Network URLs become their core §6.1 identity; a file://
// remote or a malformed network source passes through here and is rejected
// at the Identity boundary with profile_source_invalid.
func canonicalRequirementSource(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || identity.ValidCanonical(trimmed) {
		return trimmed
	}
	canonical, err := identity.Parse(trimmed)
	if err != nil || canonical == "" {
		return trimmed
	}
	return canonical
}

// lookPath resolves an executable on PATH.
func lookPath(command string) (string, error) { return exec.LookPath(command) }
