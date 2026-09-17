package config

// Machine-owned repository endpoint policy (repository-transport-v1,
// revision 1): loading, validation, and declaration-to-attempt planning.
//
// source-policy.json lives beside the manager configuration, outside
// package-controlled trees. It is optional for URL declarations and
// mandatory for logical `repository` declarations. This layer performs
// no network I/O: every structural check fails before acquisition, and
// the sibling transport-application task consumes Resolution without
// re-deriving pin or fallback semantics.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/gitcred"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// SourcePolicyFileName is the machine policy filename beside the manager
// configuration. SourcePolicySchemaVersion is the one schema this
// revision-1 reader accepts; a schema-2 document is rejected closed
// (repository-transport §4), never read with its revision-2 members
// ignored.
const (
	SourcePolicyFileName      = "source-policy.json"
	SourcePolicySchemaVersion = 1
)

// Stable transport diagnostics (repository-transport §2). Every error
// produced here carries its class as a message prefix so callers match
// on prose exactly as the published failure classes spell it.
const (
	CodeRepositoryPolicyInvalid       = "repository_policy_invalid"
	CodeRepositoryEndpointUnavailable = "repository_endpoint_unavailable"
)

// Fallback modes (repository-transport §2). Pin forces FallbackNone
// irrespective of the configured value; a URL declaration without an
// entry resolves with FallbackNone and no policy provider.
const (
	FallbackNone             = "none"
	FallbackAvailabilityAuth = "availability-auth"
)

// Endpoint is one listed machine endpoint: a closed-grammar URL whose
// lane identity equals the entry key, plus an opaque operator-owned
// authentication provider reference.
type Endpoint struct {
	URL            string
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
// maps exact canonical identity to its entry; RootInputs maps a
// case-sensitive source alias to its admitted source-relative inputs.
type SourcePolicy struct {
	Path         string
	Repositories map[string]RepositoryPolicy
	RootInputs   map[string][]string
}

// Attempt is one planned endpoint attempt. An empty Authentication names
// no policy provider: the caller falls back to the existing lane policy,
// which is the only case a URL declaration without an entry produces.
type Attempt struct {
	URL            string
	Authentication string
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

// ParseSourcePolicy validates one source-policy.json document (schema 1).
// path names the document in diagnostics only.
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
	if err := rejectPolicyFields(obj, "", "schema_version", "repositories", "root_inputs"); err != nil {
		return nil, err
	}
	schema, ok := integerValue(obj["schema_version"])
	if !ok {
		return nil, verr.New("schema_version", "%s: must be integer %d", CodeRepositoryPolicyInvalid, SourcePolicySchemaVersion)
	}
	if schema == 2 {
		return nil, verr.New("schema_version", "%s: schema 2 is a revision-2 policy and requires a revision-2 reader", CodeRepositoryPolicyInvalid)
	}
	if schema != SourcePolicySchemaVersion {
		return nil, verr.New("schema_version", "%s: unsupported source-policy schema_version %d", CodeRepositoryPolicyInvalid, schema)
	}
	rawRepos, present := obj["repositories"]
	if !present || rawRepos == nil {
		return nil, verr.New("repositories", "%s: requires a repositories object", CodeRepositoryPolicyInvalid)
	}
	repos, ok := rawRepos.(map[string]any)
	if !ok {
		return nil, verr.New("repositories", "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	policy := &SourcePolicy{Path: path, Repositories: map[string]RepositoryPolicy{}}
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
	if policy != nil {
		entry, found = policy.Repositories[canonical]
	}
	if !found {
		if repository != "" {
			return Resolution{}, fmt.Errorf("%s: no machine policy entry for %q", CodeRepositoryEndpointUnavailable, canonical)
		}
		return Resolution{Identity: canonical, Attempts: []Attempt{{URL: gitURL}}, Fallback: FallbackNone}, nil
	}
	if len(entry.Endpoints) == 0 {
		return Resolution{}, fmt.Errorf("%s: entry for %q lists no endpoints", CodeRepositoryPolicyInvalid, canonical)
	}
	if entry.Pin != "" {
		for _, endpoint := range entry.Endpoints {
			if endpoint.URL == entry.Pin {
				return Resolution{
					Identity: canonical,
					Attempts: []Attempt{{URL: endpoint.URL, Authentication: endpoint.Authentication}},
					Fallback: FallbackNone,
				}, nil
			}
		}
		return Resolution{}, fmt.Errorf("%s: pin is not a listed endpoint of %q", CodeRepositoryPolicyInvalid, canonical)
	}
	attempts := make([]Attempt, 0, len(entry.Endpoints))
	for _, endpoint := range entry.Endpoints {
		attempts = append(attempts, Attempt(endpoint))
	}
	return Resolution{Identity: canonical, Attempts: attempts, Fallback: entry.Fallback}, nil
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
