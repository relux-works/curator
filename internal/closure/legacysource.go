package closure

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/identity"
)

// refreshCheckout refreshes an existing legacy checkout (skills root or dev
// checkout) from exactly one endpoint, chosen without network access by
// bindCheckout. The binding (gate, stored-origin read and comparison) runs
// for every requirer of the checkout, also when an earlier requirer or scope
// in the same run already refreshed it; only the fetch itself is skipped,
// and only when it was already made from the identical endpoint
// (Options.FetchedRepos). A checkout without an origin is not fetched.
func refreshCheckout(opts Options, name, declared, chain, repo string) error {
	endpoint, err := bindCheckout(opts, name, declared, chain, repo)
	if err != nil {
		return err
	}
	if endpoint.location == "" {
		markFetched(opts, repo, "")
		return nil
	}
	if fetchedFrom(opts, repo, endpoint.location) {
		return nil
	}
	if err := fetchSource(opts, name, chain, repo, endpoint); err != nil {
		return err
	}
	markFetched(opts, repo, endpoint.location)
	return nil
}

// boundEndpoint is the one endpoint bindCheckout chose for a checkout, with
// the local/network classification decided while binding it. The fetch
// dispatch uses kind as recorded and never reclassifies location: a local
// location is an absolute physical filesystem path, whose bytes (a literal
// "%", a final space) are not source-URL grammar.
type boundEndpoint struct {
	// location is the network URL or the absolute physical path handed to
	// git; empty when the checkout has no origin and is not fetched.
	location string
	kind     endpointKind
	// declared reports that location is the declaration, not the stored
	// origin (it selects the source_url_redirected wording).
	declared bool
}

// endpointKind is the transport class bindCheckout decided for an endpoint.
type endpointKind int

const (
	// networkEndpoint is fetched on the operator transport, refusing a
	// redirect (gitops.FetchURLRefusingRedirect).
	networkEndpoint endpointKind = iota + 1
	// localEndpoint is an absolute physical path fetched byte for byte on
	// the isolated lane (gitops.FetchLocalPathIsolated).
	localEndpoint
)

// bindCheckout chooses the one endpoint an existing checkout refreshes from,
// without network access:
//
//   - a declared network URL is gated, its canonical identity must equal the
//     stored origin's, and the declared URL itself is the endpoint (Rule A);
//   - with no declaration, or a declared local path, the stored origin is the
//     only refresh source: a network origin is gated, a local origin is
//     resolved to its absolute physical path, and a declared local path must
//     name the stored origin's location or the checkout itself (Rule B). A
//     local origin spelled with leading or trailing whitespace is refused
//     (source_path_whitespace), like such a declaration (gateSource).
//
// The endpoint carries its classification (boundEndpoint.kind), decided here
// from the declared and stored spellings, never from the resolved physical
// path. An empty location means the checkout has no origin and is not
// fetched. Other remotes are
// never consulted, and no checkout remote is rewritten. A mismatch refuses
// with source_origin_mismatch before any fetch.
func bindCheckout(opts Options, name, declared, chain, repo string) (boundEndpoint, error) {
	declaredID := ""
	if declared != "" {
		if err := gateSource(opts, name, declared, chain); err != nil {
			return boundEndpoint{}, err
		}
		// gateSource has already parsed the declaration successfully.
		declaredID, _ = identity.Parse(declared)
	}
	origin, hasOrigin, err := gitops.OriginURL(repo)
	if err != nil {
		return boundEndpoint{}, fmt.Errorf("failed to fetch %s at %s (via %s): cannot read the stored origin: %w", name, repo, chain, err)
	}
	if !hasOrigin {
		return boundEndpoint{}, nil
	}
	originID, originErr := identity.Parse(origin)
	switch {
	case declaredID != "":
		if originErr != nil || originID != declaredID {
			return boundEndpoint{}, originMismatch(name, declared, declaredID, origin, originID, originErr, repo, chain)
		}
		return boundEndpoint{location: declared, kind: networkEndpoint, declared: true}, nil
	case declared != "":
		if originErr != nil || originID != "" {
			return boundEndpoint{}, originMismatch(name, declared, declaredID, origin, originID, originErr, repo, chain)
		}
		if err := originSpelling(name, chain, repo, origin); err != nil {
			return boundEndpoint{}, err
		}
		if !sameLocalSource(declared, origin, repo) {
			return boundEndpoint{}, originMismatch(name, declared, declaredID, origin, originID, originErr, repo, chain)
		}
		return localOriginEndpoint(name, chain, repo, origin)
	case originErr != nil || originID != "":
		if err := gateSource(opts, name, origin, chain); err != nil {
			return boundEndpoint{}, fmt.Errorf("%w; stored origin of checkout %s", err, repo)
		}
		return boundEndpoint{location: origin, kind: networkEndpoint}, nil
	default:
		if err := originSpelling(name, chain, repo, origin); err != nil {
			return boundEndpoint{}, err
		}
		return localOriginEndpoint(name, chain, repo, origin)
	}
}

// edgeWhitespace reports a source spelling that begins or ends with
// whitespace. git reads such bytes as part of a local path, while every
// curator reading of a source spelling (identity.Parse, the clone lane's
// argument) trims them, so the two readings can name different locations.
func edgeWhitespace(spelling string) bool {
	return spelling != strings.TrimSpace(spelling)
}

// originSpelling refuses a stored local origin spelled with leading or
// trailing whitespace (source_path_whitespace) instead of choosing between
// git's byte reading and the trimmed one.
func originSpelling(name, chain, repo, origin string) error {
	if !edgeWhitespace(origin) {
		return nil
	}
	return fmt.Errorf(
		"source_path_whitespace: stored local origin %q of checkout %s for %s begins or ends with whitespace, which git reads as part of the path; nothing was fetched; set the origin without it (via %s)",
		origin, repo, name, chain)
}

// localOriginEndpoint turns a stored local origin into the endpoint handed
// to git: its absolute physical path, read from the checkout as git fetch
// reads a relative origin, with every byte of every resolved component kept
// (fetchEndpoint hands it to git unmodified). git reads a bare relative word such as "mirror",
// and a file:// URL, as the name of a configured remote before a location.
// A configuration remote name can never begin with "/", so the absolute
// POSIX path is unambiguous; a path that does not begin with "/" (a Windows
// drive path) is refused with source_origin_ambiguous when a remote of that
// exact name is configured.
func localOriginEndpoint(name, chain, repo, origin string) (boundEndpoint, error) {
	path, ok := physicalPath(origin, repo)
	if !ok {
		return boundEndpoint{}, fmt.Errorf("failed to fetch %s at %s (via %s): stored local origin %s does not resolve to a filesystem location", name, repo, chain, origin)
	}
	if !strings.HasPrefix(path, "/") {
		configured, err := gitops.RemoteConfigured(repo, path)
		if err != nil {
			return boundEndpoint{}, fmt.Errorf("failed to fetch %s at %s (via %s): cannot read the configured remotes: %w", name, repo, chain, err)
		}
		if configured {
			return boundEndpoint{}, fmt.Errorf(
				"source_origin_ambiguous: stored local origin %s of checkout %s for %s resolves to %s, which is also the name of a configured remote; nothing was fetched; rename that remote (via %s)",
				origin, repo, name, path, chain)
		}
	}
	return boundEndpoint{location: path, kind: localEndpoint}, nil
}

// fetchSource fetches repo from endpoint through the injected fetch or the
// production lane, and turns a redirect into source_url_redirected.
func fetchSource(opts Options, name, chain, repo string, endpoint boundEndpoint) error {
	var err error
	if opts.FetchRepo != nil {
		err = opts.FetchRepo(repo, endpoint.location)
	} else {
		err = fetchEndpoint(repo, endpoint)
	}
	if err == nil {
		return nil
	}
	if diagnostic := redirectDiagnostic(name, chain, repo, err, endpoint.declared); diagnostic != nil {
		return diagnostic
	}
	return fmt.Errorf("failed to fetch %s at %s from %s (via %s): %w", name, repo, endpoint.location, chain, err)
}

// fetchEndpoint is the production refresh of one bound endpoint, dispatched
// on the kind bindCheckout recorded; the location is never reparsed with
// source-URL grammar. A network endpoint uses the operator transport, like
// the clone path; a local one, always an absolute physical path from
// localOriginEndpoint, involves no network transport, uses the isolated
// configuration and is handed to git byte for byte (no trimming, no percent
// decoding: a final space or a literal "%" is part of a filename). Neither
// consults the checkout's remotes. An endpoint without a recorded kind is
// refused, never guessed.
func fetchEndpoint(repo string, endpoint boundEndpoint) error {
	switch endpoint.kind {
	case networkEndpoint:
		return gitops.FetchURLRefusingRedirect(repo, endpoint.location)
	case localEndpoint:
		return gitops.FetchLocalPathIsolated(repo, endpoint.location)
	default:
		return fmt.Errorf("refusing to fetch from %s: the endpoint has no recorded transport class", endpoint.location)
	}
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
// origin is read from the checkout, as git fetch reads it. Every operand is
// resolved to its absolute physical path (physicalPath), and two locations
// that both exist compare by filesystem identity.
func sameLocalSource(declared, origin, repo string) bool {
	declaredPath, ok := physicalPath(declared, "")
	if !ok {
		return false
	}
	if originPath, ok := physicalPath(origin, repo); ok && sameLocation(declaredPath, originPath) {
		return true
	}
	// repo is a filesystem path, not a source spelling: it is resolved as
	// is, never read through localPathOf's file: URL grammar.
	repoPath, ok := physicalLocation(repo, "")
	return ok && sameLocation(declaredPath, repoPath)
}

// sameLocation compares two physical paths: by os.SameFile when both exist,
// otherwise by their spelling (case-insensitively on Windows, like repoKey).
func sameLocation(left, right string) bool {
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	if leftErr == nil && rightErr == nil {
		return os.SameFile(leftInfo, rightInfo)
	}
	if leftErr == nil || rightErr == nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

// localPathOf returns the filesystem path a local source spells: a file://
// URL's path, or the raw value. It neither trims nor cleans the result: a
// path byte is significant (edgeWhitespace spellings are refused before this
// point), and ".." must be walked after the symlinks before it
// (physicalPath). A file: spelling without "//" is refused: git reads it as
// an scp-like host "file", not as a location.
func localPathOf(raw string) (string, bool) {
	value := raw
	if value == "" {
		return "", false
	}
	if len(value) >= 5 && strings.EqualFold(value[:5], "file:") {
		if !strings.HasPrefix(value[5:], "//") {
			return "", false
		}
		parsed, err := url.Parse(value)
		if err != nil {
			return "", false
		}
		// The host is dropped, as git drops it for file:// (probed: a
		// file://otherhost/x fetch reads /x). Percent escapes never reach
		// this point: identity.Parse refuses them, and the stored-origin
		// case then takes the gated network branch, which refuses too.
		value = parsed.Path
		// file:///C:/x carries the drive after a leading slash.
		if len(value) >= 3 && value[0] == '/' && identity.IsDrivePath(value[1:]) {
			value = value[1:]
		}
		value = filepath.FromSlash(value)
		if value == "" {
			return "", false
		}
	}
	return value, true
}

// physicalPath returns the absolute physical path a local source names, the
// location the filesystem reaches by walking it: every symlink is resolved
// before a ".." that follows it, so "link/../x" names x beside the link's
// target, not beside the link. A relative path is anchored at base, or at
// the working directory when base is empty; a relative base is itself read
// from the working directory. When the location does not exist, its longest
// existing prefix is resolved and the rest appended; a ".." in that rest
// cannot be walked, so the path is refused (ok false).
func physicalPath(raw, base string) (string, bool) {
	value, ok := localPathOf(raw)
	if !ok {
		return "", false
	}
	return physicalLocation(value, base)
}

// physicalLocation is physicalPath for a value that is already a filesystem
// path (no source-URL grammar applies to it): a checkout path, or the path
// localPathOf read from a spelling.
func physicalLocation(value, base string) (string, bool) {
	if value == "" {
		return "", false
	}
	if !filepath.IsAbs(value) {
		anchor := base
		if anchor == "" || !filepath.IsAbs(anchor) {
			working, err := os.Getwd()
			if err != nil {
				return "", false
			}
			if anchor != "" {
				working = working + string(filepath.Separator) + anchor
			}
			anchor = working
		}
		// No filepath.Join: it would clean "link/.." lexically.
		value = anchor + string(filepath.Separator) + value
	}
	if resolved, err := filepath.EvalSymlinks(value); err == nil {
		return filepath.Clean(resolved), true
	}
	volume := filepath.VolumeName(value)
	parts := strings.FieldsFunc(value[len(volume):], func(r rune) bool { return r == '/' || r == filepath.Separator })
	for index := len(parts) - 1; index >= 0; index-- {
		prefix := volume + string(filepath.Separator) + strings.Join(parts[:index], string(filepath.Separator))
		resolved, err := filepath.EvalSymlinks(prefix)
		if err != nil {
			continue
		}
		rest := parts[index:]
		for _, part := range rest {
			if part == ".." {
				return "", false
			}
		}
		return filepath.Join(append([]string{resolved}, rest...)...), true
	}
	return "", false
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
