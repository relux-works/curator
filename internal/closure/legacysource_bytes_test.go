package closure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/testgitenv"
)

// symlinkSiblingFixture is the review's two-repository shape: the bare
// repository intended (its name carries a significant byte) holds the tag
// "fresh-intended", a distinct bare sibling holds "only-in-trimmed-sibling",
// and root/mirror -> intended. The skills-root checkout "tool" has the
// ordinary origin file://<root>/mirror; nothing about the stored or declared
// spelling carries whitespace.
type symlinkSiblingFixture struct {
	checkout  string
	mirrorURL string
	physical  string // the absolute physical path git must be handed
}

func newSymlinkSiblingFixture(t *testing.T, h *harness, intended, sibling string) symlinkSiblingFixture {
	t.Helper()
	root, err := filepath.EvalSymlinks(testgitenv.ShortTempDir(t))
	if err != nil {
		t.Fatal(err)
	}
	testgitenv.RequireGitRepositoryNames(t, root, intended, sibling)
	checkout := h.skill("tool", nil, nil)
	for name, tag := range map[string]string{intended: "fresh-intended", sibling: "only-in-trimmed-sibling"} {
		bare := filepath.Join(root, name)
		h.git(root, "clone", "-q", "--bare", checkout, bare)
		h.git(bare, "tag", tag)
	}
	symlinkOrSkip(t, filepath.Join(root, intended), filepath.Join(root, "mirror"))
	mirrorURL := "file://" + filepath.ToSlash(filepath.Join(root, "mirror"))
	if filepath.VolumeName(root) != "" {
		mirrorURL = "file:///" + filepath.ToSlash(filepath.Join(root, "mirror"))
	}
	h.git(checkout, "remote", "add", "origin", mirrorURL)
	return symlinkSiblingFixture{checkout: checkout, mirrorURL: mirrorURL, physical: filepath.Join(root, intended)}
}

func (f symlinkSiblingFixture) assertFetchedIntended(t *testing.T, h *harness) {
	t.Helper()
	tags := h.git(f.checkout, "tag", "-l")
	if !strings.Contains(tags, "fresh-intended") || strings.Contains(tags, "only-in-trimmed-sibling") {
		t.Fatalf("tags %q: want the symlink target's fresh-intended and not the trimmed sibling's tag", tags)
	}
}

// Claim (review R2 witness, Build): with origin file://<root>/mirror and
// mirror -> a repository whose name ends in a space, tab or newline, Build
// hands the fetch the symlink target's exact physical path (asserted byte
// for byte through the injected fetch) and the production fetch imports the
// target's tag, never the trimmed sibling's; with no declaration and with
// the matching file:// declaration. The leading-space and embedded-newline
// names prove no other component rewrite. The percent rows (review R3
// witness: an ordinary alias whose physical target carries a literal "%")
// prove the bound local endpoint is fetched as a path, never reclassified
// with source-URL grammar, decoded or refused, also with a final space.
// The ordinary-name row is the positive control and never skips; a row
// this host's git cannot keep distinct skips with the probe's reason.
// Limit: real git and local repositories only.
func TestPhysicalOriginKeepsSignificantBytes(t *testing.T) {
	for _, names := range []struct{ name, intended, sibling string }{
		{"final space", "origin.git ", "origin.git"},
		{"final tab", "origin.git\t", "origin.git"},
		{"final newline", "origin.git\n", "origin.git"},
		{"leading space in component", " origin.git", "origin.git"},
		{"embedded newline", "ori\ngin.git", "origin.git"},
		{"literal percent", "origin%marker.git", "origin.git"},
		{"literal percent and final space", "origin%marker.git ", "origin%marker.git"},
		{"percent escape", "origin%20x.git", "origin x.git"},
		{"ordinary path control", "origin.git", "other.git"},
	} {
		for _, declared := range []bool{false, true} {
			label := names.name + "/undeclared"
			if declared {
				label = names.name + "/declared file URL"
			}
			t.Run(label, func(t *testing.T) {
				h := newHarness(t)
				fixture := newSymlinkSiblingFixture(t, h, names.intended, names.sibling)
				d := decl("tool")
				if declared {
					d = declaredTool(fixture.mirrorURL)
				}
				log := &fetchLog{}
				if err := buildOne(h.fetchOptions(log), d, nil); err != nil {
					t.Fatal(err)
				}
				if len(log.endpoints) != 1 || log.endpoints[0] != fixture.physical {
					t.Fatalf("fetched endpoints = %q, want exactly [%q]", log.endpoints, fixture.physical)
				}
				opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]string{}}
				if err := buildOne(opts, d, nil); err != nil {
					t.Fatal(err)
				}
				fixture.assertFetchedIntended(t, h)
			})
		}
	}
}

// Claim: a declared local source spelled with leading or trailing whitespace
// is refused with source_path_whitespace before any clone or fetch, on the
// existing-checkout path, the clone path and the dev-substitution path,
// because the clone lane trims it while git reads those bytes as part of the
// path. Control: the same declaration without the whitespace refreshes or
// clones. A network declaration is not affected (identity.Parse decides).
func TestWhitespaceEdgedLocalDeclarationIsRefused(t *testing.T) {
	sources := newHarness(t)
	upstream := sources.skill("tool", nil, nil)
	for _, testCase := range []struct{ name, spelling string }{
		{"final space", upstream + " "},
		{"leading space", " " + upstream},
		{"final newline", upstream + "\n"},
		{"leading tab", "\t" + upstream},
	} {
		t.Run("existing checkout "+testCase.name, func(t *testing.T) {
			h := newHarness(t)
			h.toolCheckout(upstream)
			log := &fetchLog{}
			err := buildOne(h.fetchOptions(log), declaredTool(testCase.spelling), nil)
			if err == nil || !strings.HasPrefix(err.Error(), "source_path_whitespace: tool declares local source ") || len(log.endpoints) != 0 {
				t.Fatalf("err = %v, endpoints %q; want source_path_whitespace and no fetch", err, log.endpoints)
			}
		})
	}
	t.Run("existing checkout control", func(t *testing.T) {
		h := newHarness(t)
		h.toolCheckout(upstream)
		log := &fetchLog{}
		if err := buildOne(h.fetchOptions(log), declaredTool(upstream), nil); err != nil || len(log.endpoints) != 1 {
			t.Fatalf("err = %v, endpoints %q; want one fetch", err, log.endpoints)
		}
	})
	t.Run("clone path", func(t *testing.T) {
		h := newHarness(t)
		opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]string{}}
		err := buildOne(opts, declaredTool(upstream+" "), nil)
		if err == nil || !strings.HasPrefix(err.Error(), "source_path_whitespace: ") {
			t.Fatalf("err = %v, want source_path_whitespace", err)
		}
		if _, statErr := os.Stat(filepath.Join(h.skillsRoot, "tool")); !os.IsNotExist(statErr) {
			t.Fatalf("the refused declaration was cloned: %v", statErr)
		}
		if err := buildOne(opts, declaredTool(upstream), nil); err != nil {
			t.Fatalf("control clone: %v", err)
		}
	})
	t.Run("dev substitution", func(t *testing.T) {
		h := newHarness(t)
		devRoot := testgitenv.ShortTempDir(t)
		opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, DevRoot: devRoot, FetchedRepos: map[string]string{}}
		h.skill("tool", nil, nil)
		subs := func(gitURL string) map[string]devsub.Substitution {
			return map[string]devsub.Substitution{"tool": {SkillName: "tool", Git: gitURL, RefKind: "tag", RefValue: "v1"}}
		}
		err := buildOne(opts, decl("tool"), subs(upstream+" "))
		if err == nil || !strings.HasPrefix(err.Error(), "source_path_whitespace: ") {
			t.Fatalf("err = %v, want source_path_whitespace", err)
		}
		if _, statErr := os.Stat(filepath.Join(devRoot, "tool")); !os.IsNotExist(statErr) {
			t.Fatalf("the refused substitution was cloned: %v", statErr)
		}
		if err := buildOne(opts, decl("tool"), subs(upstream)); err != nil {
			t.Fatalf("control substitution: %v", err)
		}
	})
}

// Claim: a stored local origin spelled with leading or trailing whitespace is
// refused with source_path_whitespace and no fetch, with no declaration and
// with a local declaration; the leading-space row requires OriginURL to read
// the value untrimmed. Control: the same origin without whitespace fetches.
func TestWhitespaceEdgedStoredOriginIsRefused(t *testing.T) {
	sources := newHarness(t)
	upstream := sources.skill("tool", nil, nil)
	for _, origin := range []string{" " + upstream, upstream + " ", upstream + "\t"} {
		for _, d := range []manifest.Decl{decl("tool"), declaredTool(upstream)} {
			h := newHarness(t)
			repo := h.toolCheckout("")
			h.git(repo, "config", "remote.origin.url", origin)
			log := &fetchLog{}
			err := buildOne(h.fetchOptions(log), d, nil)
			if err == nil || !strings.HasPrefix(err.Error(), "source_path_whitespace: stored local origin ") || len(log.endpoints) != 0 {
				t.Fatalf("origin %q, git %q: err = %v, endpoints %q; want source_path_whitespace and no fetch", origin, d.Git, err, log.endpoints)
			}
		}
	}
	for _, d := range []manifest.Decl{decl("tool"), declaredTool(upstream)} {
		h := newHarness(t)
		h.toolCheckout(upstream)
		log := &fetchLog{}
		if err := buildOne(h.fetchOptions(log), d, nil); err != nil || len(log.endpoints) != 1 {
			t.Fatalf("control git %q: err = %v, endpoints %q; want one fetch", d.Git, err, log.endpoints)
		}
	}
}

// Claim: a file: spelling without "//" is not read as a location (git reads
// it as an scp-like host named "file"): as a stored origin it refuses with no
// fetch. Control: the file:// spelling of the same location fetches.
func TestOpaqueFileSpellingIsNotALocation(t *testing.T) {
	sources := newHarness(t)
	upstream := sources.skill("tool", nil, nil)
	h := newHarness(t)
	repo := h.toolCheckout("file:" + filepath.ToSlash(upstream))
	log := &fetchLog{}
	err := buildOne(h.fetchOptions(log), decl("tool"), nil)
	if err == nil || !strings.Contains(err.Error(), "does not resolve to a filesystem location") || len(log.endpoints) != 0 {
		t.Fatalf("err = %v, endpoints %q; want the location refusal and no fetch", err, log.endpoints)
	}
	h.git(repo, "config", "remote.origin.url", "file://"+filepath.ToSlash(upstream))
	if err := buildOne(h.fetchOptions(log), decl("tool"), nil); err != nil || len(log.endpoints) != 1 {
		t.Fatalf("control: err = %v, endpoints %q; want one fetch", err, log.endpoints)
	}
}
