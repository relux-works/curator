package config

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/verr"
)

// build_https selects the operator token that authenticates an HTTPS fetch of
// an external build repository. As with build_ssh, credentials are
// operator-owned: no manifest, descriptor, repository, substitution, or
// marker may select them (Spec §12.2). The scope grammar and the
// longest-prefix match are exactly those of build_ssh (Spec §6.1, §6.3),
// reused rather than restated here.
//
// The config never stores a secret. An entry names where the manager should
// read the token from: TokenSourceGitCredentials reads the operator's own
// Git HTTPS credential for the scope's host, TokenSourceKeyring reads the
// manager-namespaced entry the operator's login command stored, and
// token_env names an environment variable read at process entry. Exactly one
// of token or token_env may be set.
//
// Unlike build_ssh, an unmatched scope is not an error at resolution time:
// anonymous HTTPS is a real transport and public repositories must keep
// working, so absence of a selection simply means no credential is offered.

// TokenSourceGitCredentials reads the operator's own Git HTTPS credential for
// the scope's host through the operator's Git credential machinery.
const TokenSourceGitCredentials = "git-credentials"

// TokenSourceKeyring reads the manager-namespaced entry the operator's
// `curator config build-https login` stored through that same machinery.
const TokenSourceKeyring = "keyring"

// BuildHTTPSDefaultUsername is sent alongside the resolved secret when an
// entry names no username of its own.
const BuildHTTPSDefaultUsername = "token"

// BuildHTTPSTokenRule is the human-readable rule for the token field.
var BuildHTTPSTokenRule = "must be one of " + strings.Join(buildHTTPSTokenSourceList, ", ") + "; secrets never live in the config"

// BuildHTTPSTokenEnvRule is the human-readable rule for the token_env field.
const BuildHTTPSTokenEnvRule = "must be an environment variable name" // #nosec G101 -- a rule string, not a credential

// BuildHTTPSSourceRule states the exactly-one-source requirement.
const BuildHTTPSSourceRule = "requires exactly one of 'token' or 'token_env'"

var buildHTTPSTokenSourceList = []string{TokenSourceGitCredentials, TokenSourceKeyring}

var buildHTTPSTokenSources = map[string]bool{
	TokenSourceGitCredentials: true,
	TokenSourceKeyring:        true,
}

// buildHTTPSFields are the only fields a scope entry may carry.
var buildHTTPSFields = map[string]bool{"token": true, "token_env": true, "username": true}

var envVarNameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// BuildHTTPSCredential is one operator token selection for a canonical-
// identity scope. Exactly one of Token or TokenEnv is set.
type BuildHTTPSCredential struct {
	// Scope is the config key this credential was recorded under.
	Scope string
	// Token names the enumerated source the operator selected.
	Token string
	// TokenEnv names an environment variable read at process entry instead of
	// going through the operator's Git credential machinery.
	TokenEnv string
	// Username accompanies the resolved secret. Empty means
	// BuildHTTPSDefaultUsername applies.
	Username string
}

// Empty reports a selection that names no source at all.
func (b BuildHTTPSCredential) Empty() bool { return b.Token == "" && b.TokenEnv == "" }

// EffectiveUsername returns Username, or BuildHTTPSDefaultUsername when unset.
func (b BuildHTTPSCredential) EffectiveUsername() string {
	if b.Username == "" {
		return BuildHTTPSDefaultUsername
	}
	return b.Username
}

// ValidateBuildHTTPS checks one credential against the scope grammar, the
// token-source enumeration, and the exactly-one-source requirement.
func ValidateBuildHTTPS(credential BuildHTTPSCredential) error {
	if !ValidBuildSSHScope(credential.Scope) {
		return verr.New("build_https", "scope %q %s", credential.Scope, BuildSSHScopeRule)
	}
	label := "build_https." + credential.Scope
	if credential.Token != "" && !buildHTTPSTokenSources[credential.Token] {
		return verr.New(label+".token", "%s", BuildHTTPSTokenRule)
	}
	if credential.TokenEnv != "" && !envVarNameRE.MatchString(credential.TokenEnv) {
		return verr.New(label+".token_env", "%s", BuildHTTPSTokenEnvRule)
	}
	if (credential.Token == "") == (credential.TokenEnv == "") {
		return verr.New(label, "%s", BuildHTTPSSourceRule)
	}
	return nil
}

// BuildHTTPSFor returns the credential whose scope is the longest segment-
// aware prefix of a canonical repository identity (Spec §6.3). No match is
// not a fault: the caller falls back to anonymous HTTPS.
func (c *Config) BuildHTTPSFor(canonical string) (BuildHTTPSCredential, bool) {
	return MatchBuildHTTPS(c.BuildHTTPS, canonical)
}

// MatchBuildHTTPS is BuildHTTPSFor over a bare scope map, for a caller that
// carries the operator's scopes without the whole config.
func MatchBuildHTTPS(scopes map[string]BuildHTTPSCredential, canonical string) (BuildHTTPSCredential, bool) {
	scope, ok := longestScope(scopes, canonical)
	if !ok {
		return BuildHTTPSCredential{}, false
	}
	return scopes[scope], true
}

// BuildHTTPSScopes returns the configured scopes in sorted order.
func (c *Config) BuildHTTPSScopes() []string {
	scopes := make([]string, 0, len(c.BuildHTTPS))
	for scope := range c.BuildHTTPS {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	return scopes
}

// BuildHTTPSObject renders credentials back into the config JSON shape.
// Unset fields are omitted and the result parses back to the same
// credentials.
func BuildHTTPSObject(credentials map[string]BuildHTTPSCredential) map[string]any {
	if len(credentials) == 0 {
		return nil
	}
	object := make(map[string]any, len(credentials))
	for scope, credential := range credentials {
		entry := map[string]any{}
		if credential.Token != "" {
			entry["token"] = credential.Token
		}
		if credential.TokenEnv != "" {
			entry["token_env"] = credential.TokenEnv
		}
		if credential.Username != "" {
			entry["username"] = credential.Username
		}
		object[scope] = entry
	}
	return object
}

// parseBuildHTTPS validates the build_https field. Anything the grammar does
// not spell out is rejected rather than ignored, same as build_ssh: a
// credential that silently fails to apply is worse than a config that
// refuses to load.
func parseBuildHTTPS(raw any) (map[string]BuildHTTPSCredential, error) {
	if raw == nil {
		return nil, nil
	}
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("build_https", "must be an object")
	}
	scopes := make([]string, 0, len(object))
	for scope := range object {
		scopes = append(scopes, scope)
	}
	// Sorted, so a config with several faults always reports the same one.
	sort.Strings(scopes)

	credentials := map[string]BuildHTTPSCredential{}
	for _, scope := range scopes {
		if !ValidBuildSSHScope(scope) {
			return nil, verr.New("build_https", "scope %q %s", scope, BuildSSHScopeRule)
		}
		label := "build_https." + scope
		entry, ok := object[scope].(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		var unknown []string
		for key := range entry {
			if !buildHTTPSFields[key] {
				unknown = append(unknown, key)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			return nil, verr.New(label, "has unsupported field(s): %s", strings.Join(unknown, ", "))
		}
		credential := BuildHTTPSCredential{Scope: scope}
		for field, target := range map[string]*string{
			"token": &credential.Token, "token_env": &credential.TokenEnv, "username": &credential.Username,
		} {
			raw, present := entry[field]
			if !present {
				continue
			}
			// Present-but-empty is a fault here, while an empty struct field
			// simply means the operator named nothing for it.
			value, ok := raw.(string)
			if !ok || value == "" || utf8.RuneCountInString(value) > maxCredentialPathLength {
				return nil, verr.New(label+"."+field, "must be a non-empty string when present")
			}
			*target = value
		}
		if err := ValidateBuildHTTPS(credential); err != nil {
			return nil, err
		}
		credentials[scope] = credential
	}
	return credentials, nil
}
