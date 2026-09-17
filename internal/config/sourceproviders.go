package config

// Operator authentication provider table for draft bounded transport
// resolution (repository-transport-v1 §2).
//
// source-providers.json lives beside the manager configuration, outside
// package-controlled trees, next to source-policy.json. It binds the
// opaque `authentication` provider identifiers a machine policy names to
// operator-owned configuration: per-provider HTTPS usernames and explicit
// anonymity, and per-provider SSH paths. It carries no secrets: HTTPS
// secrets stay in the operator's Git credential machinery under the
// provider namespace and are read through gitcred.Access at fetch time;
// SSH material is admitted operator paths, validated and never executed.
//
// The file is optional. An absent file configures no providers: every
// named provider reference is then unavailable and its attempt records an
// auth failure with no fetch traffic. A present-but-unreadable or invalid
// document fails repository_policy_invalid like any other malformed
// machine policy; it is never treated as absent. This layer performs no
// network I/O and confers no admission: the transport layer revalidates
// every entry before use, and unknown names stay unknown.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/gitcred"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// SourceProvidersFileName is the machine provider-table filename beside
// the manager configuration. SourceProvidersSchemaVersion is the one
// schema this reader accepts.
const (
	SourceProvidersFileName      = "source-providers.json"
	SourceProvidersSchemaVersion = 1
)

// SourceProvidersPath resolves the machine provider-table path beside the
// manager configuration.
func SourceProvidersPath() string {
	return filepath.Join(filepath.Dir(UserPath()), SourceProvidersFileName)
}

// LoadSourceProviders reads and validates the machine provider table at
// path, or at SourceProvidersPath when path is "". A missing file
// configures no providers and returns (nil, nil). Any
// present-but-unreadable or invalid document fails
// repository_policy_invalid; it is never treated as absent.
func LoadSourceProviders(path string) (buildrepo.ProviderSet, error) {
	if path == "" {
		path = SourceProvidersPath()
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- provider path comes from the operator
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: cannot read %s: %v", CodeRepositoryPolicyInvalid, path, err)
	}
	return ParseSourceProviders(payload, path)
}

// ParseSourceProviders validates one source-providers.json document.
// path names the document in diagnostics only.
func ParseSourceProviders(payload []byte, path string) (buildrepo.ProviderSet, error) {
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
	if err := rejectPolicyFields(obj, "", "schema_version", "providers"); err != nil {
		return nil, err
	}
	schema, ok := integerValue(obj["schema_version"])
	if !ok || schema != SourceProvidersSchemaVersion {
		return nil, verr.New("schema_version", "%s: must be integer %d", CodeRepositoryPolicyInvalid, SourceProvidersSchemaVersion)
	}
	rawProviders, present := obj["providers"]
	if !present || rawProviders == nil {
		return nil, verr.New("providers", "%s: requires a providers object", CodeRepositoryPolicyInvalid)
	}
	entries, ok := rawProviders.(map[string]any)
	if !ok {
		return nil, verr.New("providers", "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	set := buildrepo.ProviderSet{}
	for _, name := range sortedPolicyKeys(entries) {
		provider, err := parseProviderEntry(name, entries[name])
		if err != nil {
			return nil, err
		}
		set[name] = provider
	}
	return set, nil
}

func parseProviderEntry(name string, raw any) (buildrepo.Provider, error) {
	label := "providers." + name
	if !gitcred.ValidProvider(name) {
		return buildrepo.Provider{}, verr.New(label, "%s: provider name must be an opaque operator identifier", CodeRepositoryPolicyInvalid)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return buildrepo.Provider{}, verr.New(label, "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	if err := rejectPolicyFields(obj, label, "https", "ssh"); err != nil {
		return buildrepo.Provider{}, err
	}
	if _, hasHTTPS := obj["https"]; !hasHTTPS {
		if _, hasSSH := obj["ssh"]; !hasSSH {
			return buildrepo.Provider{}, verr.New(label, "%s: requires an https or ssh entry", CodeRepositoryPolicyInvalid)
		}
	}
	provider := buildrepo.Provider{}
	if rawHTTPS, present := obj["https"]; present {
		https, err := parseProviderHTTPS(label, rawHTTPS)
		if err != nil {
			return buildrepo.Provider{}, err
		}
		provider.HTTPS = &https
	}
	if rawSSH, present := obj["ssh"]; present {
		ssh, err := parseProviderSSH(label, rawSSH)
		if err != nil {
			return buildrepo.Provider{}, err
		}
		provider.SSH = &ssh
	}
	return provider, nil
}

func parseProviderHTTPS(label string, raw any) (buildrepo.ProviderHTTPS, error) {
	field := label + ".https"
	obj, ok := raw.(map[string]any)
	if !ok {
		return buildrepo.ProviderHTTPS{}, verr.New(field, "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	if err := rejectPolicyFields(obj, field, "username", "anonymous"); err != nil {
		return buildrepo.ProviderHTTPS{}, err
	}
	https := buildrepo.ProviderHTTPS{}
	if rawUsername, present := obj["username"]; present {
		username, ok := rawUsername.(string)
		if !ok {
			return buildrepo.ProviderHTTPS{}, verr.New(field+".username", "%s: must be a string", CodeRepositoryPolicyInvalid)
		}
		https.Username = username
	}
	if rawAnonymous, present := obj["anonymous"]; present {
		anonymous, ok := rawAnonymous.(bool)
		if !ok {
			return buildrepo.ProviderHTTPS{}, verr.New(field+".anonymous", "%s: must be a boolean", CodeRepositoryPolicyInvalid)
		}
		https.Anonymous = anonymous
	}
	if https.Anonymous && https.Username != "" {
		return buildrepo.ProviderHTTPS{}, verr.New(field, "%s: an anonymous provider names no username", CodeRepositoryPolicyInvalid)
	}
	return https, nil
}

func parseProviderSSH(label string, raw any) (buildrepo.ProviderSSH, error) {
	field := label + ".ssh"
	obj, ok := raw.(map[string]any)
	if !ok {
		return buildrepo.ProviderSSH{}, verr.New(field, "%s: must be an object", CodeRepositoryPolicyInvalid)
	}
	if err := rejectPolicyFields(obj, field, "identity", "agent_socket", "known_hosts"); err != nil {
		return buildrepo.ProviderSSH{}, err
	}
	ssh := buildrepo.ProviderSSH{}
	for _, member := range []struct {
		key   string
		value *string
	}{
		{"identity", &ssh.Identity},
		{"agent_socket", &ssh.AgentSocket},
		{"known_hosts", &ssh.KnownHosts},
	} {
		rawValue, present := obj[member.key]
		if !present {
			continue
		}
		value, ok := rawValue.(string)
		if !ok {
			return buildrepo.ProviderSSH{}, verr.New(field+"."+member.key, "%s: must be a string", CodeRepositoryPolicyInvalid)
		}
		*member.value = value
	}
	if ssh.Identity == "" && ssh.AgentSocket == "" {
		return buildrepo.ProviderSSH{}, verr.New(field, "%s: requires an identity or agent_socket entry", CodeRepositoryPolicyInvalid)
	}
	return ssh, nil
}
