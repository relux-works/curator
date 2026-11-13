package closure

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/identity"
)

// refreshCheckout refreshes an existing legacy checkout (skills root or dev
// checkout) from exactly one endpoint, chosen without network access:
//
//   - a declared network URL is gated, its canonical identity must equal the
//     stored origin's, and the declared URL itself is fetched over the
//     operator transport (Rule A);
//   - with no declaration, or a declared local path, the stored origin is the
//     only refresh source: a network origin is gated and fetched over the
//     operator transport, a local origin is fetched with the isolated
//     configuration, and a declared local path must name the stored origin's
//     location or the checkout itself (Rule B).
//
// A checkout without an origin is not fetched. Other remotes are never
// contacted, and no checkout remote is rewritten. A mismatch refuses with
// source_origin_mismatch before any fetch.
func refreshCheckout(opts Options, name, declared, chain, repo string) error {
	declaredID := ""
	if declared != "" {
		if err := gateSource(opts, name, declared, chain); err != nil {
			return err
		}
		// gateSource has already parsed the declaration successfully.
		declaredID, _ = identity.Parse(declared)
	}
	origin, hasOrigin, err := gitops.OriginURL(repo)
	if err != nil {
		return fmt.Errorf("failed to fetch %s at %s (via %s): cannot read the stored origin: %w", name, repo, chain, err)
	}
	if !hasOrigin {
		return nil
	}
	originID, originErr := identity.Parse(origin)
	switch {
	case declaredID != "":
		if originErr != nil || originID != declaredID {
			return originMismatch(name, declared, declaredID, origin, originID, originErr, repo, chain)
		}
		return fetchSource(opts, name, chain, repo, declared, true)
	case declared != "":
		if originErr != nil || originID != "" || !sameLocalSource(declared, origin, repo) {
			return originMismatch(name, declared, declaredID, origin, originID, originErr, repo, chain)
		}
		return fetchSource(opts, name, chain, repo, origin, false)
	default:
		if originErr != nil || originID != "" {
			if err := gateSource(opts, name, origin, chain); err != nil {
				return fmt.Errorf("%w; stored origin of checkout %s", err, repo)
			}
		}
		return fetchSource(opts, name, chain, repo, origin, false)
	}
}

// fetchSource fetches repo from endpoint through the injected fetch or the
// production lane, and turns a redirect into source_url_redirected.
func fetchSource(opts Options, name, chain, repo, endpoint string, declared bool) error {
	fetch := opts.FetchRepo
	if fetch == nil {
		fetch = fetchEndpoint
	}
	err := fetch(repo, endpoint)
	if err == nil {
		return nil
	}
	if diagnostic := redirectDiagnostic(name, chain, repo, err, declared); diagnostic != nil {
		return diagnostic
	}
	return fmt.Errorf("failed to fetch %s at %s from %s (via %s): %w", name, repo, endpoint, chain, err)
}

// fetchEndpoint is the production refresh of one endpoint. A network
// endpoint uses the operator transport, like the clone path; a local one
// involves no network transport and uses the isolated configuration. Neither
// consults the checkout's remotes.
func fetchEndpoint(repo, endpoint string) error {
	if id, err := identity.Parse(endpoint); err != nil || id != "" {
		return gitops.FetchURLRefusingRedirect(repo, endpoint)
	}
	return gitops.FetchURLIsolated(repo, endpoint)
}

// redirectDiagnostic returns source_url_redirected for a *gitops.RedirectError
// and nil for every other error.
func redirectDiagnostic(name, chain, repo string, err error, declared bool) error {
	var redirect *gitops.RedirectError
	if !errors.As(err, &redirect) {
		return nil
	}
	if declared {
		return fmt.Errorf(
			"source_url_redirected: declared %s for %s is not canonical; the server redirected to %s; declare %s (via %s)",
			redirect.Endpoint, name, redirect.Target, redirect.Target, chain)
	}
	return fmt.Errorf(
		"source_url_redirected: stored origin %s of checkout %s for %s is not canonical; the server redirected to %s; set the origin to %s (via %s)",
		redirect.Endpoint, repo, name, redirect.Target, redirect.Target, chain)
}

func originMismatch(name, declared, declaredID, origin, originID string, originErr error, repo, chain string) error {
	originLabel := identityLabel(originID)
	if originErr != nil {
		originLabel = "invalid: " + originErr.Error()
	}
	return fmt.Errorf(
		"source_origin_mismatch: %s declares %s (identity %s), but checkout %s has stored origin %s (identity %s); nothing was fetched; repoint the checkout's origin or fix the declaration (via %s)",
		name, declared, identityLabel(declaredID), repo, origin, originLabel, chain)
}

func identityLabel(id string) string {
	if id == "" {
		return "local"
	}
	return id
}

// sameLocalSource reports whether a declared local path names the stored
// local origin, or the checkout itself. A relative declaration is read from
// the working directory, as the clone path's git reads it; a relative stored
// origin is read from the checkout, as git fetch reads it. Paths compare
// cleaned and absolute, with symlinks resolved where they exist.
func sameLocalSource(declared, origin, repo string) bool {
	declaredPath, ok := localSourcePath(declared, "")
	if !ok {
		return false
	}
	originPath, ok := localSourcePath(origin, repo)
	if !ok {
		return false
	}
	key := repoKey(declaredPath)
	return key == repoKey(originPath) || key == repoKey(repo)
}

// localSourcePath returns the filesystem path a local source names. base
// anchors a relative path; empty base means the working directory.
func localSourcePath(raw, base string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", false
	}
	if len(value) >= 5 && strings.EqualFold(value[:5], "file:") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", false
		}
		value = parsed.Path
		if parsed.Opaque != "" {
			value = parsed.Opaque
		}
		// file:///C:/x carries the drive after a leading slash.
		if len(value) >= 3 && value[0] == '/' && identity.IsDrivePath(value[1:]) {
			value = value[1:]
		}
		value = filepath.FromSlash(value)
		if value == "" {
			return "", false
		}
	}
	if !filepath.IsAbs(value) && base != "" {
		value = filepath.Join(base, value)
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return "", false
	}
	return filepath.Clean(absolute), true
}

// declaredSpelling is the first raw spelling seen for one canonical identity.
type declaredSpelling struct {
	raw   string
	chain string
}

// checkSpelling refuses a second raw spelling of a canonical network
// identity within one legacy resolution (source_mixed_spelling). Equal raw
// spellings, local sources and unparsable URLs pass here; the latter are
// refused where they are parsed.
func checkSpelling(spellings map[string]declaredSpelling, item pending) error {
	if item.git == "" {
		return nil
	}
	id, err := identity.Parse(item.git)
	if err != nil || id == "" {
		return nil
	}
	previous, seen := spellings[id]
	if !seen {
		spellings[id] = declaredSpelling{raw: item.git, chain: item.chain}
		return nil
	}
	if previous.raw == item.git {
		return nil
	}
	return fmt.Errorf(
		"source_mixed_spelling: canonical identity %s is declared as %s (via %s) and as %s (via %s); declare one spelling",
		id, previous.raw, previous.chain, item.git, item.chain)
}
