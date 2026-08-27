package config

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/verr"
)

// build_ssh selects the operator SSH credentials used to reach an external
// build repository. Credentials are operator-owned: no manifest, descriptor,
// repository, substitution, or marker may select them (Spec §12.2). The
// operator names a scope, the scope is matched against the canonical
// `host/path` repository identity of Spec §6.3, and the longest matching
// scope wins.
//
// A scope is a lowercase host, optionally followed by `/`-separated path
// segments. Matching is segment-aware exactly like the source allowlist of
// Spec §6.1: scope `h/portals` covers `h/portals/app` and never
// `h/portals-evil`.
//
// An entry names an agent, an identity file, or both, which are the three
// authentication shapes an SSH invocation admits: an identity alone, an agent
// alone, and an agent pinned to one named identity.

// MaxBuildSSHScopeLength bounds a scope by the canonical identity it selects.
const MaxBuildSSHScopeLength = 4096

// maxCredentialPathLength bounds an operator credential path.
const maxCredentialPathLength = 4096

// maxScopeHostLength bounds the host part of a scope.
const maxScopeHostLength = 253

// BuildSSHScopeRule is the human-readable scope rule, phrased once for reuse.
const BuildSSHScopeRule = "must be a lowercase host optionally followed by '/'-separated path segments of letters, digits, dots, underscores, or hyphens"

// BuildSSHPathRule is the human-readable rule for an operator credential path.
const BuildSSHPathRule = "must be an absolute path or start with '~/' and carry no control character"

// BuildSSHAgentRule is the human-readable rule for the agent selector.
const BuildSSHAgentRule = "must be true for the operator's own agent socket, or an agent socket path"

var (
	scopeHostRE    = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)*$`)
	scopeSegmentRE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	// Recognised without consulting the running platform, so one config file
	// yields the same verdict wherever it is read.
	windowsAbsoluteRE = regexp.MustCompile(`^(?:[A-Za-z]:[\\/]|\\\\)`)
)

// buildSSHFields are the only fields a scope entry may carry.
var buildSSHFields = map[string]bool{"agent": true, "identity": true, "known_hosts": true}

// BuildSSHCredential is one operator credential selection. Paths keep the
// operator's spelling; Expanded resolves a leading `~/`.
type BuildSSHCredential struct {
	// Scope is the config key this credential was recorded under.
	Scope string
	// Agent reports that an SSH agent is selected for this scope.
	Agent bool
	// AgentSocket names the agent socket. Empty with Agent set means the
	// agent socket the operator's own environment provides, which keeps a
	// per-login socket path out of a persisted config.
	AgentSocket string
	// Identity is the identity file offered to the destination. Together
	// with Agent it pins the single key the agent may offer.
	Identity string
	// KnownHosts overrides the operator known-hosts file for this scope.
	KnownHosts string
}

// Empty reports a selection that names no credential at all.
func (b BuildSSHCredential) Empty() bool { return !b.Agent && b.Identity == "" }

// Expanded returns the credential with every `~/` path resolved against the
// operator's home directory.
func (b BuildSSHCredential) Expanded() BuildSSHCredential {
	b.AgentSocket = expandHome(b.AgentSocket)
	b.Identity = expandHome(b.Identity)
	b.KnownHosts = expandHome(b.KnownHosts)
	return b
}

// ValidBuildSSHScope reports whether value is a well-formed credential scope.
func ValidBuildSSHScope(value string) bool {
	if value == "" || utf8.RuneCountInString(value) > MaxBuildSSHScopeLength {
		return false
	}
	host, path, hasPath := strings.Cut(value, "/")
	if len(host) > maxScopeHostLength || !scopeHostRE.MatchString(host) {
		return false
	}
	if !hasPath {
		return true
	}
	for _, segment := range strings.Split(path, "/") {
		// A canonical identity never carries an empty, "." or ".." path
		// component, so a scope spelled with one could never match.
		if segment == "." || segment == ".." || !scopeSegmentRE.MatchString(segment) {
			return false
		}
	}
	return true
}

// ValidBuildSSHPath reports whether value is a usable operator credential
// path, so a caller can tell an omitted path from a malformed one before it
// builds a credential.
func ValidBuildSSHPath(value string) bool { return validCredentialPath(value) }

// ValidateBuildSSH checks one credential against the scope grammar, the path
// rule, and the requirement that a scope actually select something.
func ValidateBuildSSH(credential BuildSSHCredential) error {
	if !ValidBuildSSHScope(credential.Scope) {
		return verr.New("build_ssh", "scope %q %s", credential.Scope, BuildSSHScopeRule)
	}
	label := "build_ssh." + credential.Scope
	if credential.AgentSocket != "" && !credential.Agent {
		return verr.New(label+".agent", "%s", BuildSSHAgentRule)
	}
	for _, field := range []struct{ name, value string }{
		{"agent", credential.AgentSocket},
		{"identity", credential.Identity},
		{"known_hosts", credential.KnownHosts},
	} {
		if field.value != "" && !validCredentialPath(field.value) {
			return verr.New(label+"."+field.name, "%s", BuildSSHPathRule)
		}
	}
	if credential.Empty() {
		return verr.New(label, "requires 'agent', 'identity', or both")
	}
	return nil
}

// BuildSSHFor returns the credential whose scope is the longest segment-aware
// prefix of a canonical repository identity (Spec §6.3). A value that is not
// a canonical `host/path` identity, and an identity no scope covers, select
// no credential; the caller then fails closed rather than falling back to
// ambient SSH state.
func (c *Config) BuildSSHFor(canonical string) (BuildSSHCredential, bool) {
	return MatchBuildSSH(c.BuildSSH, canonical)
}

// MatchBuildSSH is BuildSSHFor over a bare scope map, for a caller that carries
// the operator's scopes without the whole config: the run-wide selection of an
// install already has to travel that way.
func MatchBuildSSH(scopes map[string]BuildSSHCredential, canonical string) (BuildSSHCredential, bool) {
	scope, ok := longestScope(scopes, canonical)
	if !ok {
		return BuildSSHCredential{}, false
	}
	return scopes[scope], true
}

// longestScope returns the key of scopes selected by the longest segment-aware
// prefix match against a canonical repository identity (Spec §6.3), shared by
// every credential surface keyed on a build_ssh-grammar scope so the matching
// rule lives in exactly one place.
func longestScope[T any](scopes map[string]T, canonical string) (string, bool) {
	if !identity.ValidCanonical(canonical) {
		return "", false
	}
	best := ""
	for scope := range scopes {
		if len(scope) > len(best) && identity.MatchesPrefix(canonical, scope) {
			best = scope
		}
	}
	if best == "" {
		return "", false
	}
	return best, true
}

// BuildSSHScopes returns the configured scopes in sorted order.
func (c *Config) BuildSSHScopes() []string {
	scopes := make([]string, 0, len(c.BuildSSH))
	for scope := range c.BuildSSH {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	return scopes
}

// BuildSSHObject renders credentials back into the config JSON shape. Unset
// fields are omitted and the result parses back to the same credentials.
func BuildSSHObject(credentials map[string]BuildSSHCredential) map[string]any {
	if len(credentials) == 0 {
		return nil
	}
	object := make(map[string]any, len(credentials))
	for scope, credential := range credentials {
		entry := map[string]any{}
		switch {
		case credential.AgentSocket != "":
			entry["agent"] = credential.AgentSocket
		case credential.Agent:
			entry["agent"] = true
		}
		if credential.Identity != "" {
			entry["identity"] = credential.Identity
		}
		if credential.KnownHosts != "" {
			entry["known_hosts"] = credential.KnownHosts
		}
		object[scope] = entry
	}
	return object
}

// parseBuildSSH validates the build_ssh field. Anything the grammar does not
// spell out is rejected rather than ignored: a credential that silently fails
// to apply is worse than a config that refuses to load.
func parseBuildSSH(raw any) (map[string]BuildSSHCredential, error) {
	if raw == nil {
		return nil, nil
	}
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("build_ssh", "must be an object")
	}
	scopes := make([]string, 0, len(object))
	for scope := range object {
		scopes = append(scopes, scope)
	}
	// Sorted, so a config with several faults always reports the same one.
	sort.Strings(scopes)

	credentials := map[string]BuildSSHCredential{}
	for _, scope := range scopes {
		if !ValidBuildSSHScope(scope) {
			return nil, verr.New("build_ssh", "scope %q %s", scope, BuildSSHScopeRule)
		}
		label := "build_ssh." + scope
		entry, ok := object[scope].(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		var unknown []string
		for key := range entry {
			if !buildSSHFields[key] {
				unknown = append(unknown, key)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			return nil, verr.New(label, "has unsupported field(s): %s", strings.Join(unknown, ", "))
		}
		credential := BuildSSHCredential{Scope: scope}
		if raw, present := entry["agent"]; present {
			switch value := raw.(type) {
			case bool:
				// Only the affirmative spelling exists: "agent": false
				// would be a second way to write an identity-only entry.
				if !value {
					return nil, verr.New(label+".agent", "%s", BuildSSHAgentRule)
				}
				credential.Agent = true
			case string:
				if !validCredentialPath(value) {
					return nil, verr.New(label+".agent", "socket %s", BuildSSHPathRule)
				}
				credential.Agent = true
				credential.AgentSocket = value
			default:
				return nil, verr.New(label+".agent", "%s", BuildSSHAgentRule)
			}
		}
		for field, target := range map[string]*string{
			"identity": &credential.Identity, "known_hosts": &credential.KnownHosts,
		} {
			raw, present := entry[field]
			if !present {
				continue
			}
			// Present-but-empty is a fault here, while an empty struct
			// field simply means the operator named no path.
			value, ok := raw.(string)
			if !ok || !validCredentialPath(value) {
				return nil, verr.New(label+"."+field, "%s", BuildSSHPathRule)
			}
			*target = value
		}
		if err := ValidateBuildSSH(credential); err != nil {
			return nil, err
		}
		credentials[scope] = credential
	}
	return credentials, nil
}

// validCredentialPath reports whether value is a rooted, control-free path an
// operator can hand to the SSH client unchanged. A relative path is refused
// because a machine-global config has no stable directory to resolve it
// against.
func validCredentialPath(value string) bool {
	if value == "" || utf8.RuneCountInString(value) > maxCredentialPathLength {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	if strings.HasPrefix(value, "~/") {
		return true
	}
	return strings.HasPrefix(value, "/") || windowsAbsoluteRE.MatchString(value)
}
