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
		FetchedRepos:   map[string]string{},
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
		opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]string{}}
		if err := buildOne(opts, decl("tool"), nil); err != nil {
			t.Fatal(err)
		}
		fixture.assertRefreshedFromOriginOnly(t, h)
	})
	t.Run("dev checkout", func(t *testing.T) {
		h := newHarness(t)
		devRoot := testgitenv.ShortTempDir(t)
		fixture := newRealFetchFixture(t, h, devRoot)
		opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, DevRoot: devRoot, FetchedRepos: map[string]string{}}
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
	opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]string{}}
	if err := buildOne(opts, declaredTool(fixture.checkout), nil); err != nil {
		t.Fatal(err)
	}
	fixture.assertRefreshedFromOriginOnly(t, h)
}

// Claim: with a local declaration, the stored origin must be local and name
// the same location (spelling may differ: file:// or a trailing separator),
// and the fetch is handed the origin's absolute physical path; another local
// path, or a network origin, refuses with no fetch. The reverse, a network
// declaration over a local origin, refuses too.
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
			want, err := filepath.EvalSymlinks(testCase.stored)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(log.endpoints, ",") != want {
				t.Fatalf("fetched endpoints = %v, want the stored origin's physical path %s", log.endpoints, want)
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
	redirectTo := func(_, endpoint string) error {
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
		opts.FetchedRepos = map[string]string{}
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

// Claim: in one run whose scopes share FetchedRepos (the upgrade --all
// wiring), a later scope that reaches an already refreshed checkout is still
// gated and bound to its stored origin: a disallowed or mismatching
// declaration refuses with no further fetch, on the skills-root path and the
// dev path. A second scope with the identical declaration resolves without
// fetching again; one with the same identity spelled differently is bound and
// fetched from its own spelling, because only an identical endpoint is
// deduplicated.
func TestEveryScopeIsBoundBeforeTheFetchIsDeduplicated(t *testing.T) {
	declared := "https://git.example.com/skills/tool.git"
	allowed := []string{"git.example.com/skills"}
	cases := []struct {
		name, second, refusal string
		endpoints             []string
	}{
		{"disallowed declaration", "https://forbidden.example.com/skills/tool.git", "source not allowed for tool", []string{declared}},
		{"mismatching declaration", "https://git.example.com/skills/other.git", "source_origin_mismatch: ", []string{declared}},
		{"identical declaration", declared, "", []string{declared}},
		{"same identity, other spelling", "https://git.example.com/skills/tool", "", []string{declared, "https://git.example.com/skills/tool"}},
	}
	for _, testCase := range cases {
		t.Run("skills root/"+testCase.name, func(t *testing.T) {
			h := newHarness(t)
			h.toolCheckout(declared)
			log := &fetchLog{}
			opts := h.fetchOptions(log, allowed...)
			if err := buildOne(opts, declaredTool(declared), nil); err != nil {
				t.Fatal(err)
			}
			err := buildOne(opts, declaredTool(testCase.second), nil)
			assertSecondScope(t, err, testCase.refusal, log.endpoints, testCase.endpoints)
		})
		t.Run("dev checkout/"+testCase.name, func(t *testing.T) {
			h := newHarness(t)
			dev := newHarness(t)
			dev.toolCheckout(declared)
			log := &fetchLog{}
			opts := h.fetchOptions(log, allowed...)
			opts.DevRoot = dev.skillsRoot
			substitute := func(gitURL string) map[string]devsub.Substitution {
				return map[string]devsub.Substitution{"tool": {SkillName: "tool", Git: gitURL, RefKind: "tag", RefValue: "v1"}}
			}
			if err := buildOne(opts, decl("tool"), substitute(declared)); err != nil {
				t.Fatal(err)
			}
			err := buildOne(opts, decl("tool"), substitute(testCase.second))
			assertSecondScope(t, err, testCase.refusal, log.endpoints, testCase.endpoints)
		})
	}
}

func assertSecondScope(t *testing.T, err error, refusal string, endpoints, want []string) {
	t.Helper()
	if refusal == "" && err != nil {
		t.Fatalf("second scope: %v", err)
	}
	if refusal != "" && (err == nil || !strings.HasPrefix(err.Error(), refusal)) {
		t.Fatalf("second scope err = %v, want %q", err, refusal)
	}
	if strings.Join(endpoints, ",") != strings.Join(want, ",") {
		t.Fatalf("fetched endpoints = %v, want %v", endpoints, want)
	}
}

// Claim: a checkout cloned by an earlier scope of the same run is bound like
// any existing checkout when a later scope reaches it: another local
// declaration refuses with source_origin_mismatch, the same declaration
// resolves. Real git, local paths only.
func TestClonedCheckoutIsBoundForLaterScopes(t *testing.T) {
	sources := newHarness(t)
	upstream := sources.skill("tool", nil, nil)
	elsewhere := sources.skill("elsewhere", nil, nil)
	for _, testCase := range []struct {
		name, second string
		refuse       bool
	}{{"other local path", elsewhere, true}, {"same local path", upstream, false}} {
		t.Run(testCase.name, func(t *testing.T) {
			h := newHarness(t)
			opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]string{}}
			if err := buildOne(opts, declaredTool(upstream), nil); err != nil {
				t.Fatal(err)
			}
			err := buildOne(opts, declaredTool(testCase.second), nil)
			if !testCase.refuse {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") {
				t.Fatalf("err = %v, want source_origin_mismatch", err)
			}
		})
	}
}

// Claim: a bare relative stored origin ("mirror") is fetched as the location
// it names beside the checkout, never as the configured remote of the same
// name: the location's tag arrives, the remote's tag and remote-tracking refs
// do not. The same holds for a file:// origin, which git also reads as a
// remote name when one is configured with that exact name. Real git, local
// repositories only.
func TestLocalOriginIsNeverReadAsARemoteName(t *testing.T) {
	sources := newHarness(t)
	forbidden := sources.skill("forbidden", nil, nil)
	sources.git(forbidden, "tag", "only-in-forbidden-remote")
	forbiddenURL := "file://" + filepath.ToSlash(forbidden)
	if filepath.VolumeName(forbidden) != "" {
		forbiddenURL = "file:///" + filepath.ToSlash(forbidden)
	}
	for _, testCase := range []struct{ name, origin string }{
		{"bare relative origin", "mirror"},
		{"file URL origin", ""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			h := newHarness(t)
			checkout := h.skill("tool", nil, nil)
			mirror := filepath.Join(checkout, "mirror")
			h.git(checkout, "clone", "-q", "--bare", checkout, mirror)
			h.git(mirror, "tag", "fresh")
			origin := testCase.origin
			if origin == "" {
				origin = "file://" + filepath.ToSlash(mirror)
				if filepath.VolumeName(mirror) != "" {
					origin = "file:///" + filepath.ToSlash(mirror)
				}
			}
			h.git(checkout, "remote", "add", "origin", origin)
			h.git(checkout, "config", "remote."+origin+".url", forbiddenURL)
			opts := Options{SkillsRoot: h.skillsRoot, Home: h.home, FetchExisting: true, FetchedRepos: map[string]string{}}
			if err := buildOne(opts, decl("tool"), nil); err != nil {
				t.Fatal(err)
			}
			tags := h.git(checkout, "tag", "-l")
			if !strings.Contains(tags, "fresh") || strings.Contains(tags, "only-in-forbidden-remote") {
				t.Fatalf("tags %q: want the location's fresh tag and not the same-named remote's", tags)
			}
			if refs := h.git(checkout, "for-each-ref", "refs/remotes/"+origin); refs != "" {
				t.Fatalf("the same-named remote was fetched: %q", refs)
			}
		})
	}
}

// Claim: RemoteConfigured, the guard for local origins whose physical path
// does not begin with "/" (Windows drive paths), sees a remote configured
// under exactly that name and nothing else; the binding then refuses with
// source_origin_ambiguous. Limit: on POSIX the guard branch is not reached
// (physical paths begin with "/"), so this proves the accessor, not the
// Windows binding end to end.
func TestRemoteConfiguredSeesExactRemoteName(t *testing.T) {
	h := newHarness(t)
	checkout := h.toolCheckout("https://git.example.com/skills/tool.git")
	h.git(checkout, "config", `remote.C:\skills\mirror.url`, "https://other.example.com/x.git")
	for name, want := range map[string]bool{`C:\skills\mirror`: true, `C:\skills`: false, "origin": true, "mirror": false} {
		got, err := gitops.RemoteConfigured(checkout, name)
		if err != nil || got != want {
			t.Fatalf("RemoteConfigured(%q) = %v, %v; want %v", name, got, err, want)
		}
	}
}

// symlinkOrSkip creates link -> target, skipping where symlinks need a
// privilege the test process lacks.
func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
}

// Claim: "entry/../x" is compared as the filesystem walks it, resolving the
// symlink entry before "..", not as the lexical "x": with entry ->
// elsewhere/child, a declaration root/entry/../origin.git names
// root/elsewhere/origin.git. It is refused against an origin at the lexical
// root/origin.git (no fetch), and accepted against an origin at the physical
// location, which is fetched by its physical path.
func TestSymlinkParentIsResolvedBeforeDotDot(t *testing.T) {
	root := t.TempDir()
	physicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{filepath.Join(root, "elsewhere", "child"), filepath.Join(root, "elsewhere", "origin.git"), filepath.Join(root, "origin.git")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	symlinkOrSkip(t, filepath.Join(root, "elsewhere", "child"), filepath.Join(root, "entry"))
	declared := root + string(filepath.Separator) + "entry" + string(filepath.Separator) + ".." + string(filepath.Separator) + "origin.git"
	for _, testCase := range []struct {
		name, stored string
		refuse       bool
	}{
		{"lexical location refuses", filepath.Join(root, "origin.git"), true},
		{"physical location fetches", filepath.Join(root, "elsewhere", "origin.git"), false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			h := newHarness(t)
			h.toolCheckout(testCase.stored)
			log := &fetchLog{}
			err := buildOne(h.fetchOptions(log), declaredTool(declared), nil)
			if testCase.refuse {
				if err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") || len(log.endpoints) != 0 {
					t.Fatalf("err = %v, endpoints %v; want source_origin_mismatch and no fetch", err, log.endpoints)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(physicalRoot, "elsewhere", "origin.git")
			if strings.Join(log.endpoints, ",") != want {
				t.Fatalf("fetched endpoints = %v, want %s", log.endpoints, want)
			}
		})
	}
	t.Run("stored origin through a symlink parent", func(t *testing.T) {
		// The same rule on the origin side, anchored at the checkout.
		h := newHarness(t)
		repo := h.toolCheckout("")
		symlinkOrSkip(t, filepath.Join(root, "elsewhere", "child"), filepath.Join(repo, "link"))
		h.git(repo, "remote", "add", "origin", "link/../origin.git")
		log := &fetchLog{}
		if err := buildOne(h.fetchOptions(log), declaredTool(filepath.Join(root, "elsewhere", "origin.git")), nil); err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join(physicalRoot, "elsewhere", "origin.git"); strings.Join(log.endpoints, ",") != want {
			t.Fatalf("fetched endpoints = %v, want %s", log.endpoints, want)
		}
		if err := buildOne(h.fetchOptions(log), declaredTool(filepath.Join(repo, "origin.git")), nil); err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") {
			t.Fatalf("lexical reading of the origin accepted: %v", err)
		}
	})
}

// Claim: with a relative skills_root, a declaration naming the checkout's
// absolute path is the self-path case and refreshes; a relative declaration
// of the same checkout does too, and an unrelated absolute path still
// refuses. All operands compare in one absolute coordinate system.
func TestRelativeSkillsRootAcceptsSelfPath(t *testing.T) {
	h := newHarness(t)
	repo := h.toolCheckout("")
	upstream := t.TempDir()
	h.git(repo, "remote", "add", "origin", upstream)
	t.Chdir(filepath.Dir(h.skillsRoot))
	relativeRoot := filepath.Base(h.skillsRoot)
	for _, testCase := range []struct {
		name, declared string
		refuse         bool
	}{
		{"absolute self path", repo, false},
		{"relative self path", filepath.Join(relativeRoot, "tool"), false},
		{"unrelated path", t.TempDir(), true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			log := &fetchLog{}
			opts := h.fetchOptions(log)
			opts.SkillsRoot = relativeRoot
			err := buildOne(opts, declaredTool(testCase.declared), nil)
			if testCase.refuse {
				if err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") || len(log.endpoints) != 0 {
					t.Fatalf("err = %v, endpoints %v; want source_origin_mismatch and no fetch", err, log.endpoints)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want, err := filepath.EvalSymlinks(upstream)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(log.endpoints, ",") != want {
				t.Fatalf("fetched endpoints = %v, want %s", log.endpoints, want)
			}
		})
	}
}

// Claim (R4 audit): the checkout path in the self-path comparison is a
// filesystem path, never read with source-URL grammar: under a relative
// skills root named "file:roots", a declaration naming the checkout itself
// ("./file:roots/tool") refreshes from the stored local origin instead of
// refusing with source_origin_mismatch because "file:roots/tool" looked
// like an opaque file: URL. Control: an unrelated declared path still
// refuses with no fetch. Limit: skipped where the host cannot create a name
// with ':' (Windows).
func TestCheckoutPathIsNotReadAsASourceURL(t *testing.T) {
	parent := testgitenv.ShortTempDir(t)
	skillsRoot := filepath.Join(parent, "file:roots")
	if err := os.Mkdir(skillsRoot, 0o755); err != nil {
		t.Skipf("this host cannot create a directory named %q: %v", filepath.Base(skillsRoot), err)
	}
	h := &harness{t: t, skillsRoot: skillsRoot, home: testgitenv.ShortTempDir(t)}
	repo := h.toolCheckout("")
	upstream := t.TempDir()
	h.git(repo, "remote", "add", "origin", upstream)
	t.Chdir(parent)
	want, err := filepath.EvalSymlinks(upstream)
	if err != nil {
		t.Fatal(err)
	}
	log := &fetchLog{}
	opts := h.fetchOptions(log)
	opts.SkillsRoot = "file:roots"
	if err := buildOne(opts, declaredTool("./file:roots/tool"), nil); err != nil {
		t.Fatal(err)
	}
	if strings.Join(log.endpoints, ",") != want {
		t.Fatalf("fetched endpoints = %v, want %s", log.endpoints, want)
	}
	unrelated := &fetchLog{}
	opts = h.fetchOptions(unrelated)
	opts.SkillsRoot = "file:roots"
	err = buildOne(opts, declaredTool(t.TempDir()), nil)
	if err == nil || !strings.HasPrefix(err.Error(), "source_origin_mismatch: ") || len(unrelated.endpoints) != 0 {
		t.Fatalf("control: err = %v, endpoints %v; want source_origin_mismatch and no fetch", err, unrelated.endpoints)
	}
}
