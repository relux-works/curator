package config

// Machine-owned repository endpoint policy (repository-transport
// revisions 1 and 2): loading, validation, and declaration-to-attempt
// planning.
//
// source-policy.json lives beside the manager configuration, outside
// package-controlled trees. It is optional for URL declarations and
// mandatory for logical `repository` declarations. This layer performs
// no network I/O: every structural check fails before acquisition, and
// the sibling transport-application task consumes Resolution without
// re-deriving pin or fallback semantics.
//
// Revision 2 (source-policy-v2.schema.json, schema_version 2) is an
// additive superset of schema 1: every schema-1 repositories/root_inputs
// shape loads with revision-1 semantics, extended with optional
// per-endpoint mirror_of and alias fields, an optional top-level aliases
// table, and an endpoint/pin URL grammar admitting explicit ports
// (repository-transport §§4-5). Canonical host/path stays the only
// portable identity: aliases and mirrors are machine-policy endpoint
// properties and never enter identity, the lock, receipts, markers, or
// allowlist matching. Attempt bounds, failure classification beyond the
// loader's fail-closed rows, and provenance recording (§§6-7) belong to
// the sibling resolution leaf; the loaded model carries what that leaf
// needs (mirror_of, alias, and the resolved connection address per
// attempt).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/gitcred"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// SourcePolicyFileName is the machine policy filename beside the manager
// configuration. SourcePolicySchemaVersion is the schema 1 version;
// SourcePolicySchemaVersionV2 is the additive revision-2 superset. This
// revision-2 reader accepts both; schema 1 documents keep revision-1
// semantics byte-identically (repository-transport §4).
const (
	SourcePolicyFileName        = "source-policy.json"
	SourcePolicySchemaVersion   = 1
	SourcePolicySchemaVersionV2 = 2
)

// Stable transport diagnostics (repository-transport §§2, 6). Every error
// produced here carries its class as a message prefix so callers match
// on prose exactly as the published failure classes spell it.
const (
	CodeRepositoryPolicyInvalid       = "repository_policy_invalid"
	CodeRepositoryEndpointUnavailable = "repository_endpoint_unavailable"
	// CodeRepositoryMirrorUndeclared reports a resolved connection host
	// that differs from the entry-key host without a mirror_of
	// attestation equal to the key (§§5-6). Zero attempts, no fallback.
	CodeRepositoryMirrorUndeclared = "repository_mirror_undeclared"
	// CodeRepositoryAliasUnknown reports an alias field naming no entry
	// in the policy alias table (§§5-6). Zero attempts, no fallback.
	CodeRepositoryAliasUnknown = "repository_alias_unknown"
)

// Fallback modes (repository-transport §2). Pin forces FallbackNone
// irrespective of the configured value; a URL declaration without an
// entry resolves with FallbackNone and no policy provider.
const (
	FallbackNone             = "none"
	FallbackAvailabilityAuth = "availability-auth"
)

// Endpoint is one listed machine endpoint: a closed-grammar URL whose
// port-stripped lane identity path equals the entry key path, plus an
// opaque operator-owned authentication provider reference.
//
// Under schema 2 the URL may carry an explicit port (URI forms only),
// MirrorOf attests a resolved connection host that differs from the
// entry-key host (present if and only if it differs, always equal to the
// key), and Alias names one entry of the operator alias table through
// the alias field only. All three are machine-policy endpoint
// properties, never portable identity (repository-transport §5).
type Endpoint struct {
	URL            string
	Authentication string
	MirrorOf       string
	Alias          string
}

// AliasEntry is one operator-owned host-alias table entry (§5): a
// concrete lowercase target host, an optional port 1-65535, and the
// authentication provider for connections via this alias, which must
// equal the endpoint authentication.
type AliasEntry struct {
	Host           string
	Port           int
	HasPort        bool
	Authentication string
}

// RepositoryPolicy is one validated `repositories` entry keyed by exact
// canonical identity. Pin is "" when absent.
type RepositoryPolicy struct {
	Identity  string
	Endpoints []Endpoint
	Pin       string
	Fallback  string
}

// SourcePolicy is a validated machine source-policy.json. Repositories
// maps exact canonical identity to its entry; Aliases holds the
// operator-owned host-alias table (nil when the document names none);
// RootInputs maps a case-sensitive source alias to its admitted
// source-relative inputs. SchemaVersion records the accepted document
// version (1 or 2).
type SourcePolicy struct {
	Path          string
	SchemaVersion int
	Repositories  map[string]RepositoryPolicy
	Aliases       map[string]AliasEntry
	RootInputs    map[string][]string
}

// Attempt is one planned endpoint attempt. An empty Authentication names
// no policy provider: the caller falls back to the existing lane policy,
// which is the only case a URL declaration without an entry produces.
//
// MirrorOf, Alias, and the resolved connection address carry the §5
// endpoint properties the sibling resolution leaf needs (§§6-7):
// ResolvedHost is the alias target host when Alias is named, else the
// endpoint URL host (both lowercased); ResolvedPort is the alias port
// when present, else the URL port when present, else 0 for the transport
// default; HasExplicitPort reports whether either port was present.
type Attempt struct {
	URL             string
	Authentication  string
	MirrorOf        string
	Alias           string
	ResolvedHost    string
	ResolvedPort    int
	HasExplicitPort bool
}

// ConnectionURL returns the git connection target for one planned
// attempt: the listed URL with the §5 alias substitution applied. An
// attempt without an alias connects to its listed URL verbatim —
// explicit ports and declared mirrors already address the connection
// there. An attempt naming an alias connects to the resolved address:
// the alias host with the alias port when present, else the URL port
// when present, else the transport default. The listed URL keeps its
// scheme, userinfo, and path; only the host (and port) change, except
// an scp-like URL combined with an explicit port, which has no
// port-bearing spelling and is rendered as the equivalent ssh:// URL
// with a home-relative path.
//
// ok is false when no connection target can be formed from the carried
// values; the caller fails the acquisition closed before any network
// I/O. Loader-planned attempts always form a target: ok is false only
// for a hand-built or mistranslated attempt.
func (a Attempt) ConnectionURL() (target string, ok bool) {
	if a.Alias == "" {
		return a.URL, true
	}
	// A named alias always resolves through the operator table, so the
	// connection host is always an alias-table host; anything else is a
	// mistranslation. The port flag and value must agree, as in the
	// sibling executor's §6 predicate.
	if !validAliasName(a.ResolvedHost) {
		return "", false
	}
	if a.HasExplicitPort != (a.ResolvedPort != 0) {
		return "", false
	}
	port := ""
	if a.HasExplicitPort {
		if a.ResolvedPort < 1 || a.ResolvedPort > 65535 {
			return "", false
		}
		port = ":" + strconv.Itoa(a.ResolvedPort)
	}
	if rest, found := strings.CutPrefix(a.URL, "https://"); found {
		_, path, valid := splitConnectionAuthority(rest, false)
		if !valid {
			return "", false
		}
		return "https://" + a.ResolvedHost + port + path, true
	}
	if rest, found := strings.CutPrefix(a.URL, "ssh://"); found {
		userinfo, path, valid := splitConnectionAuthority(rest, true)
		if !valid {
			return "", false
		}
		return "ssh://" + userinfo + a.ResolvedHost + port + path, true
	}
	return scpConnectionURL(a.URL, a.ResolvedHost, port)
}

// splitConnectionAuthority splits the post-scheme remainder of a
// URI-form endpoint URL into its userinfo prefix ("" or ending in "@")
// and its path (starting in "/"). A listed URL port is dropped: the
// caller applies the resolved port. ok is false when the remainder is
// outside the endpoint grammar.
func splitConnectionAuthority(rest string, userinfoAllowed bool) (userinfo, path string, ok bool) {
	authority, path := rest, ""
	if i := strings.Index(rest, "/"); i >= 0 {
		authority, path = rest[:i], rest[i:]
	}
	if authority == "" || path == "" || strings.ContainsAny(path, "%?#\\") || containsPolicyWhitespaceOrControl(path) {
		return "", "", false
	}
	if i := strings.LastIndex(authority, "@"); i >= 0 {
		if !userinfoAllowed {
			return "", "", false
		}
		userinfo, authority = authority[:i+1], authority[i+1:]
		if !validConnectionUserinfo(userinfo[:len(userinfo)-1]) || authority == "" {
			return "", "", false
		}
	}
	// The lane host grammar carries no colon, so any port suffix starts
	// at the first colon; a second colon, or a non-port suffix, is
	// outside the grammar rather than a silently dropped port.
	host, suffix, hasPort := strings.Cut(authority, ":")
	if host == "" {
		return "", "", false
	}
	if hasPort {
		if _, valid := parseRevision2Port(suffix); !valid {
			return "", "", false
		}
	}
	return userinfo, path, true
}

// scpConnectionURL substitutes the resolved host (and port) into an
// scp-like [user@]host:path endpoint URL. Without an explicit port the
// spelling is kept; with one the URL is rendered as the equivalent
// ssh:// URL, since the scp-like spelling cannot carry a port. Admitted
// scp-like paths are home-relative, so the ~/ prefix preserves the
// remote path.
func scpConnectionURL(raw, resolvedHost, port string) (string, bool) {
	left, path, hasColon := strings.Cut(raw, ":")
	if !hasColon || strings.Contains(raw, "://") || left == "" || path == "" ||
		strings.HasPrefix(path, "/") || strings.ContainsAny(path, "%?#\\:") || containsPolicyWhitespaceOrControl(path) {
		return "", false
	}
	userinfo := ""
	host := left
	if i := strings.LastIndex(left, "@"); i >= 0 {
		userinfo, host = left[:i+1], left[i+1:]
		if !validConnectionUserinfo(userinfo[:len(userinfo)-1]) {
			return "", false
		}
	}
	if host == "" {
		return "", false
	}
	if port == "" {
		return userinfo + resolvedHost + ":" + path, true
	}
	return "ssh://" + userinfo + resolvedHost + port + "/~/" + path, true
}

// validConnectionUserinfo reports whether value can pass through into a
// substituted connection URL without changing the URL structure: no
// separator, whitespace, or control bytes. The loader already enforces
// the lane username grammar; this is the fail-closed floor for
// hand-built attempts.
func validConnectionUserinfo(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if r > unicode.MaxASCII || unicode.IsSpace(r) || unicode.IsControl(r) || strings.ContainsRune("@/:", r) {
			return false
		}
	}
	return true
}

// Resolution is the ordered attempt plan for one declaration. Attempts
// holds one or two entries in attempt order (exactly one when pinned or
// when no entry applies); Fallback is the effective mode after pin is
// applied; Identity is the stable canonical identity allowlist checks
// match before acquisition.
type Resolution struct {
	Identity string
	Attempts []Attempt
	Fallback string
}

// SourcePolicyPath resolves the machine policy path beside the manager
// configuration.
func SourcePolicyPath() string {
	return filepath.Join(filepath.Dir(UserPath()), SourcePolicyFileName)
}

// LoadSourcePolicy reads and validates the machine policy at path, or at
// SourcePolicyPath when path is "". A missing file is an absent policy
// and returns (nil, nil): URL declarations still attempt their declared
// endpoint once. Any present-but-unreadable or invalid document fails
// repository_policy_invalid; it is never treated as absent.
func LoadSourcePolicy(path string) (*SourcePolicy, error) {
	if path == "" {
		path = SourcePolicyPath()
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- policy path comes from the operator
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: cannot read %s: %v", CodeRepositoryPolicyInvalid, path, err)
	}
	return ParseSourcePolicy(payload, path)
}

// ParseSourcePolicy validates one source-policy.json document (schema 1
// or schema 2). path names the document in diagnostics only. Schema 1
// documents load with revision-1 semantics exactly as before; schema 2
// documents load as the additive superset of §4 (ports, mirror_of, the
// operator alias table). Unknown versions fail closed before any
// network I/O.
func ParseSourcePolicy(payload []byte, path string) (*SourcePolicy, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, verr.New("", "%s: malformed JSON in %s: %v", CodeRepositoryPolicyInvalid, path, err)
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, verr.New("", "%s: malformed JSON in %s: %v", CodeRepositoryPolicyInvalid, path, err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("", "%s: %s must contain an object", CodeRepositoryPolicyInvalid, path)
	}
	schema, ok := integerValue(obj["schema_version"])
	if !ok {
		return nil, verr.New("schema_version", "%s: must be integer %d or %d", CodeRepositoryPolicyInvalid, SourcePolicySchemaVersion, SourcePolicySchemaVersionV2)
	}
	if schema != SourcePolicySchemaVersion && schema != SourcePolicySchemaVersionV2 {
		return nil, verr.New("schema_version", "%s: unsupported source-policy schema_version %d", CodeRepositoryPolicyInvalid, schema)
	}
	if schema == SourcePolicySchemaVersion {
		return parseSourcePolicyV1(obj, path)
	}
	return parseSourcePolicyV2(obj, path)
}

// parseSourcePolicyV1 validates a schema-1 document with revision-1
// semantics. This path is byte-identical to the pre-revision-2 loader:
// ports, mirror_of, alias members, and the aliases table are all
// rejected here as unknown fields or lane-grammar violations.
func parseSourcePolicyV1(obj map[string]any, path string) (*SourcePolicy, error) {
	if err := rejectPolicyFields(obj, "", "schema_version", "repositories", "root_inputs"); err != nil {
		return nil, err
	}
	repos, err := policyRepositoriesObject(obj)
	if err != nil {
		return nil, err
	}
	policy := &SourcePolicy{Path: path, SchemaVersion: SourcePolicySchemaVersion, Repositories: map[string]RepositoryPolicy{}}
	for _, key := range sortedPolicyKeys(repos) {
		entry, err := parseRepositoryEntry(key, repos[key])
		if err != nil {
			return nil, err
		}
		policy.Repositories[key] = entry
	}
	if rawInputs, present := obj["root_inputs"]; present {
		inputs, err := parseRootInputs(rawInputs)
		if err != nil {
			return nil, err
		}
		policy.RootInputs = inputs
	}
	return policy, nil
}

// parseSourcePolicyV2 validates a schema-2 document: the schema-1 shape
// plus the optional aliases table and the per-endpoint mirror_of, alias,
// and port extensions of §5.
func parseSourcePolicyV2(obj map[string]any, path string) (*SourcePolicy, error) {
	if err := rejectPolicyFields(obj, "", "schema_version", "repositories", "aliases", "root_inputs"); err != nil {
		return nil, err
	}
	repos, err := policyRepositoriesObject(obj)
	if err != nil {
		return nil, err
	}
	policy := &SourcePolicy{Path: path, SchemaVersion: SourcePolicySchemaVersionV2, Repositories: map[string]RepositoryPolicy{}}
	if rawAliases, present := obj["aliases"]; present {
		aliases, err := parseAliasTable(rawAliases)
		if err != nil {
			return nil, err
		}
		policy.Aliases = aliases
	}
	if err := rejectChainedAliases(policy.Aliases); err != nil {
		return nil, err
	}
	for _, key := range sortedPolicyKeys(repos) {
		entry, err := parseRepositoryEntryV2(key, repos[key], policy.Aliases)
		if err != nil {
			return nil, err
		}
		policy.Repositories[key] = entry
	}
	if rawInputs, present := obj["root_inputs"]; present {
		inputs, err := parseRootInputs(rawInputs)
		if err != nil {
			return nil, err
		}
		policy.RootInputs = inputs
	}
	return policy, nil
}

func policyRepositoriesObject(obj map[string]any) (map[string]any, error) {
	rawRepos, present := obj["repositories"]
	if !present || rawRepos == nil {
		return nil, verr.New("repositories", "%s: requires a repositories object", CodeRepositoryPolicyInvalid)
	}
	repos, ok := rawRepos.(map[string]any)
	if !ok {
		return nil, verr.New("repositories", "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	return repos, nil
}

// ResolveRepositoryEndpoints plans endpoint attempts for one source
// declaration: exactly one of gitURL (a declared URL) or repository (a
// logical canonical identity) must be set. policy may be nil for an
// absent machine policy. No network I/O occurs here.
//
//   - URL with an entry: the machine list in list order (pin selects one);
//     the declared URL never reorders the list.
//   - URL without an entry: the declared URL once, no policy provider,
//     no fallback.
//   - Logical with an entry: the machine list as above.
//   - Logical without an entry: repository_endpoint_unavailable.
//   - A malformed declaration: source_selection_invalid.
func ResolveRepositoryEndpoints(policy *SourcePolicy, gitURL, repository string) (Resolution, error) {
	if (gitURL == "") == (repository == "") {
		return Resolution{}, fmt.Errorf("source_selection_invalid: requires exactly one of git or repository")
	}
	var canonical string
	if gitURL != "" {
		parsed, err := buildrepo.ParseSource(gitURL)
		if err != nil {
			return Resolution{}, fmt.Errorf("source_selection_invalid: invalid repository endpoint")
		}
		canonical = parsed.Identity
	} else {
		if !identity.DraftCanonicalKey(repository) {
			return Resolution{}, fmt.Errorf("source_selection_invalid: repository must be canonical host/path")
		}
		canonical = repository
	}
	var entry RepositoryPolicy
	found := false
	var aliases map[string]AliasEntry
	if policy != nil {
		entry, found = policy.Repositories[canonical]
		aliases = policy.Aliases
	}
	if !found {
		if repository != "" {
			return Resolution{}, fmt.Errorf("%s: no machine policy entry for %q", CodeRepositoryEndpointUnavailable, canonical)
		}
		host, _, _ := strings.Cut(canonical, "/")
		return Resolution{Identity: canonical, Attempts: []Attempt{{URL: gitURL, ResolvedHost: host}}, Fallback: FallbackNone}, nil
	}
	if len(entry.Endpoints) == 0 {
		return Resolution{}, fmt.Errorf("%s: entry for %q lists no endpoints", CodeRepositoryPolicyInvalid, canonical)
	}
	// Identity is always the entry key: ports, mirrors, and aliases are
	// machine-policy endpoint properties carried per attempt, never
	// portable identity (§5).
	if entry.Pin != "" {
		for _, endpoint := range entry.Endpoints {
			if endpoint.URL == entry.Pin {
				attempt, err := planAttempt(canonical, endpoint, aliases)
				if err != nil {
					return Resolution{}, err
				}
				return Resolution{
					Identity: canonical,
					Attempts: []Attempt{attempt},
					Fallback: FallbackNone,
				}, nil
			}
		}
		return Resolution{}, fmt.Errorf("%s: pin is not a listed endpoint of %q", CodeRepositoryPolicyInvalid, canonical)
	}
	attempts := make([]Attempt, 0, len(entry.Endpoints))
	for _, endpoint := range entry.Endpoints {
		attempt, err := planAttempt(canonical, endpoint, aliases)
		if err != nil {
			return Resolution{}, err
		}
		attempts = append(attempts, attempt)
	}
	return Resolution{Identity: canonical, Attempts: attempts, Fallback: entry.Fallback}, nil
}

// planAttempt derives one attempt's §5 connection address from its
// validated endpoint and the policy alias table. Parsed policies already
// satisfy these checks; hand-built policies fail here with the same
// fail-closed classes the loader reports.
func planAttempt(key string, endpoint Endpoint, aliases map[string]AliasEntry) (Attempt, error) {
	keyHost, keyPath, _ := strings.Cut(key, "/")
	urlHost, urlPort, hasURLPort, epIdentity, err := parseRevision2EndpointURL(endpoint.URL)
	if err != nil {
		return Attempt{}, verr.New("", "%s: endpoint %q is outside the endpoint grammar", CodeRepositoryPolicyInvalid, endpoint.URL)
	}
	_, epPath, _ := strings.Cut(epIdentity, "/")
	resolvedHost, resolvedPort, hasExplicitPort, err := checkEndpointSemantics("", key, keyHost, keyPath, urlHost, epPath, hasURLPort, urlPort, endpoint.MirrorOf, endpoint.Alias, endpoint.Authentication, aliases)
	if err != nil {
		return Attempt{}, err
	}
	return Attempt{
		URL:             endpoint.URL,
		Authentication:  endpoint.Authentication,
		MirrorOf:        endpoint.MirrorOf,
		Alias:           endpoint.Alias,
		ResolvedHost:    resolvedHost,
		ResolvedPort:    resolvedPort,
		HasExplicitPort: hasExplicitPort,
	}, nil
}

func parseRepositoryEntry(key string, raw any) (RepositoryPolicy, error) {
	label := "repositories." + key
	if !identity.DraftCanonicalKey(key) {
		return RepositoryPolicy{}, verr.New(label, "%s: entry key must be exact canonical host/path", CodeRepositoryPolicyInvalid)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return RepositoryPolicy{}, verr.New(label, "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	if err := rejectPolicyFields(obj, label, "endpoints", "pin", "fallback"); err != nil {
		return RepositoryPolicy{}, err
	}
	rawEndpoints, present := obj["endpoints"]
	if !present || rawEndpoints == nil {
		return RepositoryPolicy{}, verr.New(label+".endpoints", "%s: requires one or two endpoints", CodeRepositoryPolicyInvalid)
	}
	list, ok := rawEndpoints.([]any)
	if !ok || len(list) < 1 || len(list) > 2 {
		return RepositoryPolicy{}, verr.New(label+".endpoints", "%s: requires one or two endpoints", CodeRepositoryPolicyInvalid)
	}
	entry := RepositoryPolicy{Identity: key}
	seen := map[string]bool{}
	for index, raw := range list {
		endpoint, err := parsePolicyEndpoint(label, index, key, raw)
		if err != nil {
			return RepositoryPolicy{}, err
		}
		if seen[endpoint.URL] {
			return RepositoryPolicy{}, verr.New(label+".endpoints", "%s: endpoints must be distinct URLs", CodeRepositoryPolicyInvalid)
		}
		seen[endpoint.URL] = true
		entry.Endpoints = append(entry.Endpoints, endpoint)
	}
	if rawPin, present := obj["pin"]; present {
		pin, ok := rawPin.(string)
		if !ok {
			return RepositoryPolicy{}, verr.New(label+".pin", "%s: must be a listed endpoint URL", CodeRepositoryPolicyInvalid)
		}
		if _, err := buildrepo.ParseSource(pin); err != nil {
			return RepositoryPolicy{}, verr.New(label+".pin", "%s: pin is outside the closed repository grammar", CodeRepositoryPolicyInvalid)
		}
		if !seen[pin] {
			return RepositoryPolicy{}, verr.New(label+".pin", "%s: pin must equal a listed endpoint URL exactly", CodeRepositoryPolicyInvalid)
		}
		entry.Pin = pin
	}
	rawFallback, present := obj["fallback"]
	if !present || rawFallback == nil {
		return RepositoryPolicy{}, verr.New(label+".fallback", "%s: requires fallback %q or %q", CodeRepositoryPolicyInvalid, FallbackNone, FallbackAvailabilityAuth)
	}
	fallback, ok := rawFallback.(string)
	if !ok || (fallback != FallbackNone && fallback != FallbackAvailabilityAuth) {
		return RepositoryPolicy{}, verr.New(label+".fallback", "%s: requires fallback %q or %q", CodeRepositoryPolicyInvalid, FallbackNone, FallbackAvailabilityAuth)
	}
	entry.Fallback = fallback
	return entry, nil
}

func parsePolicyEndpoint(label string, index int, key string, raw any) (Endpoint, error) {
	field := fmt.Sprintf("%s.endpoints[%d]", label, index)
	obj, ok := raw.(map[string]any)
	if !ok {
		return Endpoint{}, verr.New(field, "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	if err := rejectPolicyFields(obj, field, "url", "authentication"); err != nil {
		return Endpoint{}, err
	}
	rawURL, present := obj["url"]
	if !present || rawURL == nil {
		return Endpoint{}, verr.New(field+".url", "%s: requires an endpoint URL", CodeRepositoryPolicyInvalid)
	}
	rawAuth, present := obj["authentication"]
	if !present || rawAuth == nil {
		return Endpoint{}, verr.New(field+".authentication", "%s: requires an authentication provider reference", CodeRepositoryPolicyInvalid)
	}
	urlValue, ok := rawURL.(string)
	if !ok {
		return Endpoint{}, verr.New(field+".url", "%s: must be a string", CodeRepositoryPolicyInvalid)
	}
	parsed, err := buildrepo.ParseSource(urlValue)
	if err != nil {
		return Endpoint{}, verr.New(field+".url", "%s: endpoint is outside the closed repository grammar", CodeRepositoryPolicyInvalid)
	}
	if parsed.Identity != key {
		return Endpoint{}, verr.New(field+".url", "%s: endpoint canonicalizes to %q, want entry key %q", CodeRepositoryPolicyInvalid, parsed.Identity, key)
	}
	provider, ok := rawAuth.(string)
	if !ok || !gitcred.ValidProvider(provider) {
		return Endpoint{}, verr.New(field+".authentication", "%s: must be an opaque operator provider identifier", CodeRepositoryPolicyInvalid)
	}
	return Endpoint{URL: urlValue, Authentication: provider}, nil
}

// aliasNameRE is the §5 operator alias grammar: lowercase ASCII labels.
// Alias names, alias target hosts, and the endpoint alias field all use
// it, and every comparison is case-sensitive and exact.
var aliasNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*$`)

// maxAliasNameBytes bounds alias names and target hosts (schema maxLength).
const maxAliasNameBytes = 253

func validAliasName(value string) bool {
	return value != "" && len(value) <= maxAliasNameBytes && aliasNameRE.MatchString(value)
}

// parseRepositoryEntryV2 validates one schema-2 repositories entry: the
// schema-1 shape plus per-endpoint mirror_of, alias, and port-bearing
// URLs against the operator alias table.
func parseRepositoryEntryV2(key string, raw any, aliases map[string]AliasEntry) (RepositoryPolicy, error) {
	label := "repositories." + key
	if !identity.DraftCanonicalKey(key) {
		return RepositoryPolicy{}, verr.New(label, "%s: entry key must be exact canonical host/path", CodeRepositoryPolicyInvalid)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return RepositoryPolicy{}, verr.New(label, "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	if err := rejectPolicyFields(obj, label, "endpoints", "pin", "fallback"); err != nil {
		return RepositoryPolicy{}, err
	}
	rawEndpoints, present := obj["endpoints"]
	if !present || rawEndpoints == nil {
		return RepositoryPolicy{}, verr.New(label+".endpoints", "%s: requires one or two endpoints", CodeRepositoryPolicyInvalid)
	}
	list, ok := rawEndpoints.([]any)
	if !ok || len(list) < 1 || len(list) > 2 {
		return RepositoryPolicy{}, verr.New(label+".endpoints", "%s: requires one or two endpoints", CodeRepositoryPolicyInvalid)
	}
	entry := RepositoryPolicy{Identity: key}
	seen := map[string]bool{}
	for index, raw := range list {
		endpoint, err := parsePolicyEndpointV2(label, index, key, raw, aliases)
		if err != nil {
			return RepositoryPolicy{}, err
		}
		if seen[endpoint.URL] {
			return RepositoryPolicy{}, verr.New(label+".endpoints", "%s: endpoints must be distinct URLs", CodeRepositoryPolicyInvalid)
		}
		seen[endpoint.URL] = true
		entry.Endpoints = append(entry.Endpoints, endpoint)
	}
	if rawPin, present := obj["pin"]; present {
		pin, ok := rawPin.(string)
		if !ok {
			return RepositoryPolicy{}, verr.New(label+".pin", "%s: must be a listed endpoint URL", CodeRepositoryPolicyInvalid)
		}
		if _, _, _, _, err := parseRevision2EndpointURL(pin); err != nil {
			return RepositoryPolicy{}, verr.New(label+".pin", "%s: pin is outside the revision-2 endpoint grammar", CodeRepositoryPolicyInvalid)
		}
		if !seen[pin] {
			return RepositoryPolicy{}, verr.New(label+".pin", "%s: pin must equal a listed endpoint URL exactly, including any port", CodeRepositoryPolicyInvalid)
		}
		entry.Pin = pin
	}
	rawFallback, present := obj["fallback"]
	if !present || rawFallback == nil {
		return RepositoryPolicy{}, verr.New(label+".fallback", "%s: requires fallback %q or %q", CodeRepositoryPolicyInvalid, FallbackNone, FallbackAvailabilityAuth)
	}
	fallback, ok := rawFallback.(string)
	if !ok || (fallback != FallbackNone && fallback != FallbackAvailabilityAuth) {
		return RepositoryPolicy{}, verr.New(label+".fallback", "%s: requires fallback %q or %q", CodeRepositoryPolicyInvalid, FallbackNone, FallbackAvailabilityAuth)
	}
	entry.Fallback = fallback
	return entry, nil
}

func parsePolicyEndpointV2(label string, index int, key string, raw any, aliases map[string]AliasEntry) (Endpoint, error) {
	field := fmt.Sprintf("%s.endpoints[%d]", label, index)
	obj, ok := raw.(map[string]any)
	if !ok {
		return Endpoint{}, verr.New(field, "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	if err := rejectPolicyFields(obj, field, "url", "authentication", "mirror_of", "alias"); err != nil {
		return Endpoint{}, err
	}
	rawURL, present := obj["url"]
	if !present || rawURL == nil {
		return Endpoint{}, verr.New(field+".url", "%s: requires an endpoint URL", CodeRepositoryPolicyInvalid)
	}
	rawAuth, present := obj["authentication"]
	if !present || rawAuth == nil {
		return Endpoint{}, verr.New(field+".authentication", "%s: requires an authentication provider reference", CodeRepositoryPolicyInvalid)
	}
	urlValue, ok := rawURL.(string)
	if !ok {
		return Endpoint{}, verr.New(field+".url", "%s: must be a string", CodeRepositoryPolicyInvalid)
	}
	provider, ok := rawAuth.(string)
	if !ok || !gitcred.ValidProvider(provider) {
		return Endpoint{}, verr.New(field+".authentication", "%s: must be an opaque operator provider identifier", CodeRepositoryPolicyInvalid)
	}
	mirrorOf := ""
	if rawMirror, present := obj["mirror_of"]; present {
		value, ok := rawMirror.(string)
		if !ok || !identity.DraftCanonicalKey(value) {
			return Endpoint{}, verr.New(field+".mirror_of", "%s: mirror_of must be an exact canonical host/path", CodeRepositoryPolicyInvalid)
		}
		mirrorOf = value
	}
	aliasName := ""
	if rawAlias, present := obj["alias"]; present {
		value, ok := rawAlias.(string)
		if !ok || !validAliasName(value) {
			return Endpoint{}, verr.New(field+".alias", "%s: alias must be a lowercase operator alias name", CodeRepositoryPolicyInvalid)
		}
		aliasName = value
	}
	urlHost, urlPort, hasURLPort, epIdentity, err := parseRevision2EndpointURL(urlValue)
	if err != nil {
		return Endpoint{}, verr.New(field+".url", "%s: endpoint is outside the revision-2 endpoint grammar", CodeRepositoryPolicyInvalid)
	}
	keyHost, keyPath, _ := strings.Cut(key, "/")
	_, epPath, _ := strings.Cut(epIdentity, "/")
	if _, _, _, err := checkEndpointSemantics(field, key, keyHost, keyPath, urlHost, epPath, hasURLPort, urlPort, mirrorOf, aliasName, provider, aliases); err != nil {
		return Endpoint{}, err
	}
	return Endpoint{URL: urlValue, Authentication: provider, MirrorOf: mirrorOf, Alias: aliasName}, nil
}

// checkEndpointSemantics enforces the §5 identity rules shared by the
// loader and the attempt planner: the port-stripped path must equal the
// key path; a URL host equal to an alias key is refused; an alias
// reference must resolve to a concrete entry with matching
// authentication and no double port; and the single resolved-host
// predicate governs mirror_of. It returns the resolved connection host
// (lowercased), the resolved port (0 for the transport default), and
// whether an explicit port was present. field prefixes diagnostics and
// is empty at the planning entry.
func checkEndpointSemantics(field, key, keyHost, keyPath, urlHost, epPath string, hasURLPort bool, urlPort int, mirrorOf, aliasName, endpointAuth string, aliases map[string]AliasEntry) (string, int, bool, error) {
	label := field
	if label == "" {
		label = "repositories." + key
	}
	if epPath != keyPath {
		return "", 0, false, verr.New(label, "%s: endpoint path %q is not the entry-key path %q; only the host may differ, and only when attested", CodeRepositoryPolicyInvalid, epPath, keyPath)
	}
	// A URL host that equals an alias key is never substituted
	// implicitly. The comparison uses the lowercased URL host because DNS
	// is case-insensitive; an uppercase spelling of an alias key still
	// addresses the alias.
	if _, embedded := aliases[urlHost]; embedded {
		return "", 0, false, verr.New(label, "%s: endpoint URL host %q names an operator alias; aliases apply only through the alias field", CodeRepositoryPolicyInvalid, urlHost)
	}
	resolvedHost := urlHost
	resolvedPort := 0
	hasExplicitPort := false
	if aliasName != "" {
		target, ok := aliases[aliasName]
		if !ok {
			return "", 0, false, verr.New(label, "%s: unknown operator alias %q", CodeRepositoryAliasUnknown, aliasName)
		}
		if urlHost != keyHost {
			return "", 0, false, verr.New(label, "%s: a mirror URL must not be combined with an alias", CodeRepositoryPolicyInvalid)
		}
		if hasURLPort && target.HasPort {
			return "", 0, false, verr.New(label, "%s: a URL port and an alias port must not both be present", CodeRepositoryPolicyInvalid)
		}
		if target.Authentication != endpointAuth {
			return "", 0, false, verr.New(label, "%s: alias authentication must equal the endpoint authentication", CodeRepositoryPolicyInvalid)
		}
		resolvedHost = target.Host
		if target.HasPort {
			resolvedPort, hasExplicitPort = target.Port, true
		} else if hasURLPort {
			resolvedPort, hasExplicitPort = urlPort, true
		}
	} else if hasURLPort {
		resolvedPort, hasExplicitPort = urlPort, true
	}
	if resolvedHost != keyHost {
		if mirrorOf == "" {
			return "", 0, false, verr.New(label, "%s: resolved host %q differs from the entry-key host %q without a mirror_of attestation", CodeRepositoryMirrorUndeclared, resolvedHost, keyHost)
		}
		if mirrorOf != key {
			return "", 0, false, verr.New(label, "%s: mirror_of must equal the entry key %q exactly", CodeRepositoryPolicyInvalid, key)
		}
	} else if mirrorOf != "" {
		return "", 0, false, verr.New(label, "%s: mirror_of is forbidden when the resolved host equals the entry-key host", CodeRepositoryPolicyInvalid)
	}
	return resolvedHost, resolvedPort, hasExplicitPort, nil
}

// parseRevision2EndpointURL validates a §5 endpoint or pin URL: the
// revision-1 closed grammar plus an optional explicit port on URI forms
// only (https://host[:port]/path, ssh://[user@]host[:port]/path). The
// port is decimal 1-65535 with no leading zeros. Scp-like spellings
// carry no port: the segment after ":" is always the path. It returns
// the lowercased URL host, the explicit port, and the port-stripped lane
// identity the key-path check applies to.
func parseRevision2EndpointURL(raw string) (string, int, bool, string, error) {
	if raw == "" || !utf8.ValidString(raw) || utf8.RuneCountInString(raw) > 4096 || strings.ContainsAny(raw, "%?#\\") || containsPolicyWhitespaceOrControl(raw) {
		return "", 0, false, "", fmt.Errorf("outside the revision-2 endpoint grammar")
	}
	stripped := raw
	port := 0
	hasPort := false
	if scheme, rest, ok := cutRevision2Scheme(raw); ok {
		authority := rest
		path := ""
		if i := strings.Index(rest, "/"); i >= 0 {
			authority, path = rest[:i], rest[i:]
		}
		hostport := authority
		var userinfo string
		hasUserinfo := false
		if i := strings.LastIndex(authority, "@"); i >= 0 {
			userinfo, hostport, hasUserinfo = authority[:i], authority[i+1:], true
		}
		if scheme == "https" && hasUserinfo {
			return "", 0, false, "", fmt.Errorf("HTTPS carries no userinfo")
		}
		host := hostport
		if i := strings.LastIndex(hostport, ":"); i >= 0 {
			value, ok := parseRevision2Port(hostport[i+1:])
			if !ok {
				return "", 0, false, "", fmt.Errorf("invalid explicit port")
			}
			host, port, hasPort = hostport[:i], value, true
		}
		if host == "" {
			return "", 0, false, "", fmt.Errorf("missing endpoint host")
		}
		rebuilt := scheme + "://"
		if hasUserinfo {
			rebuilt += userinfo + "@"
		}
		stripped = rebuilt + host + path
	}
	parsed, err := buildrepo.ParseSource(stripped)
	if err != nil {
		return "", 0, false, "", err
	}
	urlHost, _, _ := strings.Cut(parsed.Identity, "/")
	return urlHost, port, hasPort, parsed.Identity, nil
}

// cutRevision2Scheme splits the two URI forms that may carry a port.
// Every other spelling is scp-like and port-free.
func cutRevision2Scheme(raw string) (string, string, bool) {
	if rest, ok := strings.CutPrefix(raw, "https://"); ok {
		return "https", rest, true
	}
	if rest, ok := strings.CutPrefix(raw, "ssh://"); ok {
		return "ssh", rest, true
	}
	return "", "", false
}

// parseRevision2Port validates an explicit endpoint port: decimal
// 1-65535 with no leading zeros.
func parseRevision2Port(value string) (int, bool) {
	if value == "" || len(value) > 5 {
		return 0, false
	}
	if len(value) > 1 && value[0] == '0' {
		return 0, false
	}
	number := 0
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		number = number*10 + int(c-'0')
	}
	if number < 1 || number > 65535 {
		return 0, false
	}
	return number, true
}

func containsPolicyWhitespaceOrControl(value string) bool {
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// parseAliasTable validates the operator-owned §5 aliases table: each
// name maps to a concrete lowercase target host, an optional integer
// port 1-65535, and an opaque authentication provider reference.
func parseAliasTable(raw any) (map[string]AliasEntry, error) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("aliases", "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	aliases := make(map[string]AliasEntry, len(obj))
	for _, name := range sortedPolicyKeys(obj) {
		label := "aliases." + name
		if !validAliasName(name) {
			return nil, verr.New(label, "%s: alias name must be lowercase [a-z0-9][a-z0-9.-]*", CodeRepositoryPolicyInvalid)
		}
		entry, ok := obj[name].(map[string]any)
		if !ok {
			return nil, verr.New(label, "%s: must be an object", CodeRepositoryPolicyInvalid)
		}
		if err := rejectPolicyFields(entry, label, "host", "port", "authentication"); err != nil {
			return nil, err
		}
		rawHost, present := entry["host"]
		if !present || rawHost == nil {
			return nil, verr.New(label+".host", "%s: requires a concrete target host", CodeRepositoryPolicyInvalid)
		}
		host, ok := rawHost.(string)
		if !ok || !validAliasName(host) {
			return nil, verr.New(label+".host", "%s: target host must be lowercase [a-z0-9][a-z0-9.-]*", CodeRepositoryPolicyInvalid)
		}
		rawAuth, present := entry["authentication"]
		if !present || rawAuth == nil {
			return nil, verr.New(label+".authentication", "%s: requires an authentication provider reference", CodeRepositoryPolicyInvalid)
		}
		provider, ok := rawAuth.(string)
		if !ok || !gitcred.ValidProvider(provider) {
			return nil, verr.New(label+".authentication", "%s: must be an opaque operator provider identifier", CodeRepositoryPolicyInvalid)
		}
		parsed := AliasEntry{Host: host, Authentication: provider}
		if rawPort, present := entry["port"]; present {
			if rawPort == nil {
				return nil, verr.New(label+".port", "%s: port must be an integer 1-65535", CodeRepositoryPolicyInvalid)
			}
			value, ok := integerValue(rawPort)
			if !ok || value < 1 || value > 65535 {
				return nil, verr.New(label+".port", "%s: port must be an integer 1-65535", CodeRepositoryPolicyInvalid)
			}
			parsed.Port, parsed.HasPort = value, true
		}
		aliases[name] = parsed
	}
	return aliases, nil
}

// rejectChainedAliases enforces the single-substitution rule: an alias
// target must be a concrete host, never another alias key.
func rejectChainedAliases(aliases map[string]AliasEntry) error {
	for _, name := range sortedAliasKeys(aliases) {
		if _, chained := aliases[aliases[name].Host]; chained {
			return verr.New("aliases."+name, "%s: alias target %q is itself an alias; alias targets must be concrete hosts", CodeRepositoryPolicyInvalid, aliases[name].Host)
		}
	}
	return nil
}

func sortedAliasKeys(aliases map[string]AliasEntry) []string {
	keys := make([]string, 0, len(aliases))
	for key := range aliases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func parseRootInputs(raw any) (map[string][]string, error) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("root_inputs", "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	inputs := map[string][]string{}
	for _, alias := range sortedPolicyKeys(obj) {
		label := "root_inputs." + alias
		if !identifiers.Valid(alias) {
			return nil, verr.New(label, "%s: alias must be a portable identifier", CodeRepositoryPolicyInvalid)
		}
		list, ok := obj[alias].([]any)
		if !ok || len(list) == 0 {
			return nil, verr.New(label, "%s: requires a non-empty input list", CodeRepositoryPolicyInvalid)
		}
		entries := make([]string, 0, len(list))
		seen := map[string]bool{}
		for _, raw := range list {
			value, ok := raw.(string)
			if !ok || !identifiers.PortablePath(value) {
				return nil, verr.New(label, "%s: entries must be portable source-relative paths", CodeRepositoryPolicyInvalid)
			}
			if seen[value] {
				return nil, verr.New(label, "%s: duplicate root-input entry %q", CodeRepositoryPolicyInvalid, value)
			}
			for _, prior := range entries {
				if rootInputOverlaps(prior, value) {
					return nil, verr.New(label, "%s: overlapping root-input entries %q and %q", CodeRepositoryPolicyInvalid, prior, value)
				}
			}
			seen[value] = true
			entries = append(entries, value)
		}
		inputs[alias] = entries
	}
	return inputs, nil
}

// rootInputOverlaps reports entries that select overlapping bytes: equal
// paths or one nested under the other. Comparison is byte-exact;
// filesystem case-equivalence is an acquisition-time check.
func rootInputOverlaps(a, b string) bool {
	return a == b || strings.HasPrefix(b, a+"/") || strings.HasPrefix(a, b+"/")
}

func rejectPolicyFields(obj map[string]any, field string, allowed ...string) error {
	set := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		set[key] = true
	}
	var unknown []string
	for key := range obj {
		if !set[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		if field == "" {
			field = "source-policy.json"
		}
		return verr.New(field, "%s: unsupported field(s): %s", CodeRepositoryPolicyInvalid, strings.Join(unknown, ", "))
	}
	return nil
}

func sortedPolicyKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
