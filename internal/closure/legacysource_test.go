package closure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/testgitenv"
)

// fetchLog records every endpoint the injected FetchRepo is asked to contact.
type fetchLog struct {
	endpoints []string
	repos     []string
	fail      error
}

func (f *fetchLog) fetch(repo, endpoint string) error {
	f.repos = append(f.repos, repo)
	f.endpoints = append(f.endpoints, endpoint)
	return f.fail
}

func declaredTool(gitURL string) manifest.Decl {
	return manifest.Decl{Name: "tool", Source: "tool", Ref: manifest.Ref{Kind: "tag", Value: "v1"}, Git: gitURL}
}

// toolCheckout creates skills_root/tool with the given stored origin URL.
func (h *harness) toolCheckout(origin string) string {
	h.t.Helper()
	dir := h.skill("tool", nil, nil)
	if origin != "" {
		h.git(dir, "remote", "add", "origin", origin)
	}
	return dir
}

func (h *harness) fetchOptions(log *fetchLog, allowed ...string) Options {
	return Options{
		SkillsRoot:     h.skillsRoot,
		Home:           h.home,
		AllowedSources: allowed,
		FetchExisting:  true,
		FetchedRepos:   map[string]bool{},
		FetchRepo:      log.fetch,
	}
}

func buildOne(opts Options, d manifest.Decl, subs map[string]devsub.Substitution) error {
	_, err := Build(opts, &manifest.Manifest{Skills: []manifest.Decl{d}}, subs)
	return err
}

// Claim: a stored origin whose identity differs from the declared URL refuses
// with source_origin_mismatch naming both URLs, both identities and the
// checkout, makes no fetch at all and leaves the stored origin untouched; the
// same checkout with a matching origin fetches the declared URL once.
func TestSwappedOriginRefusesWithoutAnyFetch(t *testing.T) {
	declared := "https://git.example.com/skills/tool.git"
	swapped := "https://evil.example.com/skills/tool.git"
	h := newHarness(t)
	repo := h.toolCheckout(swapped)
	log := &fetchLog{}

	err := buildOne(h.fetchOptions(log), declaredTool(declared), nil)
	if err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") {
		t.Fatalf("err = %v, want source_origin_mismatch", err)
	}
	for _, want := range []string{declared, swapped, "git.example.com/skills/tool", "evil.example.com/skills/tool", repo} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("diagnostic %q lacks %q", err, want)
		}
	}
	if len(log.endpoints) != 0 {
		t.Fatalf("a mismatched checkout was fetched: %v", log.endpoints)
	}
	if got := h.git(repo, "config", "--get", "remote.origin.url"); got != swapped {
		t.Fatalf("stored origin rewritten to %q", got)
	}

	// Positive control: the matching origin refreshes from the declaration.
	h.git(repo, "remote", "set-url", "origin", declared)
	if err := buildOne(h.fetchOptions(log), declaredTool(declared), nil); err != nil {
		t.Fatal(err)
	}
	if strings.Join(log.endpoints, ",") != declared {
		t.Fatalf("fetched endpoints = %v, want exactly the declared URL", log.endpoints)
	}
}

// Claim: a stored origin with the declared canonical identity but another
// spelling (.git suffix, host case, scp versus https) is not a mismatch, and
// the refresh contacts the declared URL, never the stored spelling.
func TestEqualIdentitySpelledDifferentlyFetchesDeclaredURL(t *testing.T) {
	cases := []struct{ name, declared, stored string }{
		{"git suffix", "https://git.example.com/skills/tool", "https://git.example.com/skills/tool.git"},
		{"host case", "https://Git.Example.com/skills/tool.git", "https://git.example.com/skills/tool.git"},
		{"scp versus https", "git@git.example.com:skills/tool.git", "https://git.example.com/skills/tool"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			h := newHarness(t)
			h.toolCheckout(testCase.stored)
			log := &fetchLog{}
			if err := buildOne(h.fetchOptions(log), declaredTool(testCase.declared), nil); err != nil {
				t.Fatal(err)
			}
			if len(log.endpoints) != 1 || log.endpoints[0] != testCase.declared {
				t.Fatalf("fetched endpoints = %v, want only %s", log.endpoints, testCase.declared)
			}
		})
	}
}

// Claim: the allowlist gates the existing-checkout fetch path before any
// fetch, with exactly the clone path's wording; an allowed declaration with
// the same checkout fetches.
func TestGateRunsOnFetchPathWithCloneWording(t *testing.T) {
	forbidden := "git@forbidden.example.com:skills/tool.git"
	allowed := []string{"git.example.com/skills"}

	cloneHarness := newHarness(t) // no checkout: the clone path gates
	cloneErr := buildOne(cloneHarness.fetchOptions(&fetchLog{}, allowed...), declaredTool(forbidden), nil)
	if cloneErr == nil || !strings.Contains(cloneErr.Error(), "source not allowed for tool") {
		t.Fatalf("clone path err = %v", cloneErr)
	}

	h := newHarness(t)
	repo := h.toolCheckout(forbidden)
	log := &fetchLog{}
	fetchErr := buildOne(h.fetchOptions(log, allowed...), declaredTool(forbidden), nil)
	if fetchErr == nil || fetchErr.Error() != cloneErr.Error() {
		t.Fatalf("fetch path err = %v, want the clone path's %v", fetchErr, cloneErr)
	}
	if len(log.endpoints) != 0 {
		t.Fatalf("a disallowed source was fetched: %v", log.endpoints)
	}

	permitted := "git@git.example.com:skills/tool.git"
	h.git(repo, "remote", "set-url", "origin", permitted)
	if err := buildOne(h.fetchOptions(log, allowed...), declaredTool(permitted), nil); err != nil {
		t.Fatal(err)
	}
	if strings.Join(log.endpoints, ",") != permitted {
		t.Fatalf("fetched endpoints = %v, want %s", log.endpoints, permitted)
	}
}

// Claim: an existing git-substitution checkout follows the same rule: a
// swapped origin refuses with no fetch, an equal identity spelled differently
// fetches the substitution's declared URL.
func TestDevCheckoutComparesOriginBeforeFetch(t *testing.T) {
	declared := "https://git.example.com/skills/tool.git"
	cases := []struct {
		name, stored string
		refuse       bool
	}{
		{"swapped origin", "https://evil.example.com/skills/tool.git", true},
		{"equal identity", "https://GIT.example.com/skills/tool", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			h := newHarness(t)
			dev := newHarness(t)
			devRepo := dev.toolCheckout(testCase.stored)
			log := &fetchLog{}
			opts := h.fetchOptions(log)
			opts.DevRoot = dev.skillsRoot
			subs := map[string]devsub.Substitution{"tool": {SkillName: "tool", Git: declared, RefKind: "tag", RefValue: "v1"}}
			err := buildOne(opts, decl("tool"), subs)
			if testCase.refuse {
				if err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") || !strings.Contains(err.Error(), devRepo) {
					t.Fatalf("err = %v, want source_origin_mismatch naming %s", err, devRepo)
				}
				if len(log.endpoints) != 0 {
					t.Fatalf("a mismatched dev checkout was fetched: %v", log.endpoints)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(log.endpoints, ",") != declared {
				t.Fatalf("fetched endpoints = %v, want %s", log.endpoints, declared)
			}
		})
	}
}

// Claim: an undeclared item refreshes from its stored network origin only
// after the allowlist admits that origin; a disallowed origin refuses with the
// clone path's wording plus the checkout path and no fetch, an allowed one is
// fetched exactly once at exactly the origin URL.
func TestUndeclaredSourceGatesStoredOrigin(t *testing.T) {
	allowed := []string{"git.example.com/skills"}
	t.Run("disallowed origin", func(t *testing.T) {
		origin := "https://forbidden.example.com/skills/tool.git"
		h := newHarness(t)
		repo := h.toolCheckout(origin)
		log := &fetchLog{}
		err := buildOne(h.fetchOptions(log, allowed...), decl("tool"), nil)
		want := "source not allowed for tool: " + origin + " (identity forbidden.example.com/skills/tool)"
		if err == nil || !strings.HasPrefix(err.Error(), want) || !strings.Contains(err.Error(), repo) {
			t.Fatalf("err = %v, want %q naming %s", err, want, repo)
		}
		if len(log.endpoints) != 0 {
			t.Fatalf("a disallowed origin was fetched: %v", log.endpoints)
		}
	})
	t.Run("allowed origin", func(t *testing.T) {
		origin := "https://git.example.com/skills/tool.git"
		h := newHarness(t)
		h.toolCheckout(origin)
		log := &fetchLog{}
		if err := buildOne(h.fetchOptions(log, allowed...), decl("tool"), nil); err != nil {
			t.Fatal(err)
		}
		if strings.Join(log.endpoints, ",") != origin {
			t.Fatalf("fetched endpoints = %v, want exactly %s once", log.endpoints, origin)
		}
	})
}

// Claim: a checkout with no origin is not fetched at all, whether or not the
// item is declared, and its local state still resolves.
func TestCheckoutWithoutOriginIsNotFetched(t *testing.T) {
	for _, gitURL := range []string{"", "https://git.example.com/skills/tool.git"} {
		h := newHarness(t)
		repo := h.toolCheckout("")
		h.git(repo, "remote", "add", "other", "https://git.example.com/skills/other.git")
		log := &fetchLog{}
		if err := buildOne(h.fetchOptions(log), declaredTool(gitURL), nil); err != nil {
			t.Fatalf("git %q: %v", gitURL, err)
		}
		if len(log.endpoints) != 0 {
			t.Fatalf("git %q: an origin-less checkout was fetched: %q", gitURL, log.endpoints)
		}
	}
}

// realFetchFixture is a skills-root or dev checkout cloned from a local
// upstream that gained the tag "fresh" after the clone, plus an extra remote
// "other" whose tag "only-in-other" appears in the checkout only if "other"
// is fetched.
type realFetchFixture struct {
	checkout string
	upstream string
}

func newRealFetchFixture(t *testing.T, h *harness, root string) realFetchFixture {
	t.Helper()
	sources := newHarness(t)
	upstream := sources.skill("tool", nil, nil)
	other := sources.skill("other", nil, nil)
	// Drop the shared tag name so a fetch of "other" succeeds instead of
	// failing on a clobbered tag: the witness is the tag that arrives.
	sources.git(other, "tag", "-d", "v1")
	sources.git(other, "tag", "only-in-other")
	checkout := filepath.Join(root, "tool")
	h.git(root, "clone", "-q", upstream, checkout)
	h.git(checkout, "remote", "add", "other", other)
	sources.git(upstream, "tag", "fresh")
	return realFetchFixture{checkout: checkout, upstream: upstream}
}

func (f realFetchFixture) assertRefreshedFromOriginOnly(t *testing.T, h *harness) {
	t.Helper()
	tags := h.git(f.checkout, "tag", "-l")
	if !strings.Contains(tags, "fresh") {
		t.Fatalf("the origin was not fetched: tags %q", tags)
	}
	if strings.Contains(tags, "only-in-other") {
		t.Fatalf("the non-origin remote was contacted: tags %q", tags)
	}
	if refs := h.git(f.checkout, "for-each-ref", "refs/remotes/other"); refs != "" {
		t.Fatalf("the non-origin remote was contacted: refs %q", refs)
	}
}

// Claim: with the production fetch, a checkout's extra non-origin remote is
// never contacted on the skills-root path or the dev path, while the origin
// is (its new tag arrives). Restoring the stored-remote --all fetch on either
// path fails this test.
func TestExtraRemoteIsNeverContacted(t *testing.T) {
	t.Run("skills root", func(t *testing.T) {
		h := newHarness(t)
		fixture := newRealFetchFixture(t, h, h.skillsRoot)
		opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]bool{}}
		if err := buildOne(opts, decl("tool"), nil); err != nil {
			t.Fatal(err)
		}
		fixture.assertRefreshedFromOriginOnly(t, h)
	})
	t.Run("dev checkout", func(t *testing.T) {
		h := newHarness(t)
		devRoot := testgitenv.ShortTempDir(t)
		fixture := newRealFetchFixture(t, h, devRoot)
		opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, DevRoot: devRoot, FetchedRepos: map[string]bool{}}
		subs := map[string]devsub.Substitution{"tool": {SkillName: "tool", Git: fixture.upstream, RefKind: "tag", RefValue: "v1"}}
		if err := buildOne(opts, decl("tool"), subs); err != nil {
			t.Fatal(err)
		}
		fixture.assertRefreshedFromOriginOnly(t, h)
	})
}

// Claim: a declared local path that names the checkout itself (the lifecycle
// vectors' transitive shape) refreshes from the stored local origin with the
// production fetch.
func TestDeclaredSelfPathRefreshesFromOrigin(t *testing.T) {
	h := newHarness(t)
	fixture := newRealFetchFixture(t, h, h.skillsRoot)
	opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]bool{}}
	if err := buildOne(opts, declaredTool(fixture.checkout), nil); err != nil {
		t.Fatal(err)
	}
	fixture.assertRefreshedFromOriginOnly(t, h)
}

// Claim: with a local declaration, the stored origin must be local and name
// the same location (spelling may differ: file:// or a trailing separator);
// another local path, or a network origin, refuses with no fetch. The
// reverse, a network declaration over a local origin, refuses too.
func TestLocalDeclarationComparesLocalOrigin(t *testing.T) {
	upstream := t.TempDir()
	elsewhere := t.TempDir()
	fileURL := "file://" + filepath.ToSlash(upstream)
	if filepath.VolumeName(upstream) != "" {
		fileURL = "file:///" + filepath.ToSlash(upstream)
	}
	cases := []struct {
		name, declared, stored string
		refuse                 bool
	}{
		{"same path", upstream, upstream, false},
		{"file URL spelling", fileURL, upstream + string(os.PathSeparator), false},
		{"different local path", elsewhere, upstream, true},
		{"local declaration, network origin", upstream, "https://git.example.com/skills/tool.git", true},
		{"network declaration, local origin", "https://git.example.com/skills/tool.git", upstream, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			h := newHarness(t)
			h.toolCheckout(testCase.stored)
			log := &fetchLog{}
			err := buildOne(h.fetchOptions(log), declaredTool(testCase.declared), nil)
			if testCase.refuse {
				if err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") {
					t.Fatalf("err = %v, want source_origin_mismatch", err)
				}
				if len(log.endpoints) != 0 {
					t.Fatalf("a mismatched checkout was fetched: %v", log.endpoints)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(log.endpoints, ",") != testCase.stored {
				t.Fatalf("fetched endpoints = %v, want the stored origin %s", log.endpoints, testCase.stored)
			}
		})
	}
}

// Claim: one canonical identity declared with two raw spellings in one
// resolution refuses with source_mixed_spelling naming both spellings, the
// identity and both requirer chains, whether the spellings meet on one skill
// name or two; equal spellings resolve.
func TestMixedSpellingRefused(t *testing.T) {
	plain := "https://git.example.com/skills/leaf.git"
	other := "https://git.example.com/skills/leaf"
	t.Run("one name through two requirers", func(t *testing.T) {
		for _, spellings := range [][2]string{{plain, other}, {plain, plain}} {
			h := newHarness(t)
			h.skill("leaf", nil, nil)
			left := requirement("leaf", "full")
			left["git"] = spellings[0]
			right := requirement("leaf", "full")
			right["git"] = spellings[1]
			h.skill("left", nil, map[string]map[string]any{"leaf": left})
			h.skill("right", nil, map[string]map[string]any{"leaf": right})
			_, err := h.build([]manifest.Decl{decl("left"), decl("right")}, nil)
			if spellings[0] == spellings[1] {
				if err != nil {
					t.Fatalf("equal spellings must resolve: %v", err)
				}
				continue
			}
			if err == nil || !strings.HasPrefix(err.Error(), "source_mixed_spelling: ") {
				t.Fatalf("err = %v, want source_mixed_spelling", err)
			}
			for _, want := range []string{plain, other, "git.example.com/skills/leaf", "<project> -> left -> leaf", "<project> -> right -> leaf"} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("diagnostic %q lacks %q", err, want)
				}
			}
		}
	})
	t.Run("two names", func(t *testing.T) {
		h := newHarness(t)
		h.skill("one", nil, nil)
		h.skill("two", nil, nil)
		first := manifest.Decl{Name: "one", Source: "one", Ref: manifest.Ref{Kind: "tag", Value: "v1"}, Git: plain}
		second := manifest.Decl{Name: "two", Source: "two", Ref: manifest.Ref{Kind: "tag", Value: "v1"}, Git: other}
		if _, err := h.build([]manifest.Decl{first, second}, nil); err == nil || !strings.HasPrefix(err.Error(), "source_mixed_spelling: ") {
			t.Fatalf("err = %v, want source_mixed_spelling", err)
		}
		second.Git = plain
		if _, err := h.build([]manifest.Decl{first, second}, nil); err != nil {
			t.Fatalf("equal spellings must resolve: %v", err)
		}
	})
}

// Claim: a fetch git reports as redirected refuses with source_url_redirected
// and the canonical hint, for a declared URL and for a stored origin; the
// same fetch without a redirect succeeds. The redirect itself is injected:
// the classifier of git's stderr is proven in gitops (no HTTP fixture here).
func TestRedirectedFetchRefusesWithCanonicalHint(t *testing.T) {
	target := "https://git.example.com/skills/renamed.git"
	redirectTo := func(repo, endpoint string) error {
		return &gitops.RedirectError{Endpoint: endpoint, Target: target}
	}
	t.Run("declared", func(t *testing.T) {
		declared := "https://git.example.com/skills/tool.git"
		h := newHarness(t)
		h.toolCheckout(declared)
		opts := h.fetchOptions(&fetchLog{})
		opts.FetchRepo = redirectTo
		err := buildOne(opts, declaredTool(declared), nil)
		want := "source_url_redirected: declared " + declared + " for tool is not canonical; the server redirected to " + target + "; declare " + target
		if err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("err = %v, want %q", err, want)
		}
		opts.FetchRepo = (&fetchLog{}).fetch
		opts.FetchedRepos = map[string]bool{}
		if err := buildOne(opts, declaredTool(declared), nil); err != nil {
			t.Fatalf("positive control: %v", err)
		}
	})
	t.Run("stored origin", func(t *testing.T) {
		origin := "https://git.example.com/skills/tool.git"
		h := newHarness(t)
		repo := h.toolCheckout(origin)
		opts := h.fetchOptions(&fetchLog{})
		opts.FetchRepo = redirectTo
		err := buildOne(opts, decl("tool"), nil)
		want := "source_url_redirected: stored origin " + origin + " of checkout " + repo + " for tool is not canonical; the server redirected to " + target + "; set the origin to " + target
		if err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("err = %v, want %q", err, want)
		}
	})
}
