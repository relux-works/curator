package install

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitcred"
)

const (
	// EnvBuildHTTPSToken is the run-wide HTTPS token captured at process entry.
	// It precedes configured build_https scopes for every host it covers.
	EnvBuildHTTPSToken = "CURATOR_BUILD_HTTPS_TOKEN" // #nosec G101 -- environment variable name, not a credential.
	// EnvBuildHTTPSHost optionally pins the run-wide token to one canonical
	// repository host. Repositories on every other host resolve as though the
	// override were absent (Spec section 12.2).
	EnvBuildHTTPSHost = "CURATOR_BUILD_HTTPS_HOST"
)

// BuildHTTPSCredentials is one materialized HTTPS credential. Its secret is
// deliberately private and both diagnostic formatters redact it, so logging a
// plan or selection cannot disclose authentication material.
type BuildHTTPSCredentials struct {
	Username string
	secret   string
}

// NewBuildHTTPSCredentials constructs material for the fetch broker.
func NewBuildHTTPSCredentials(username, secret string) BuildHTTPSCredentials {
	return BuildHTTPSCredentials{Username: username, secret: secret}
}

// Secret returns the credential only to the broker that must answer Git.
func (c BuildHTTPSCredentials) Secret() string { return c.secret }

// Selected reports whether this is an authenticated rather than anonymous
// HTTPS transport.
func (c BuildHTTPSCredentials) Selected() bool { return c.secret != "" }

func (c BuildHTTPSCredentials) String() string {
	return fmt.Sprintf("BuildHTTPSCredentials{Username:%q, Secret:<redacted>}", c.Username)
}

// GoString returns the same redacted diagnostic representation as String.
func (c BuildHTTPSCredentials) GoString() string { return c.String() }

// CapturedBuildHTTPSOverride carries the run-wide environment override. The
// host may be empty (all HTTPS hosts) or the exact canonical host it covers.
// The secret never participates in its diagnostic representation.
type CapturedBuildHTTPSOverride struct {
	Host   string
	secret string
}

// Secret returns the captured value to the resolver only.
func (o CapturedBuildHTTPSOverride) Secret() string { return o.secret }

func (o CapturedBuildHTTPSOverride) String() string {
	return fmt.Sprintf("CapturedBuildHTTPSOverride{Host:%q, Secret:<redacted>}", o.Host)
}

// GoString returns the same redacted diagnostic representation as String.
func (o CapturedBuildHTTPSOverride) GoString() string { return o.String() }

// BuildHTTPSCredentialReader is the operator credential-store surface needed
// by resolution. gitcred.Access is its production implementation.
type BuildHTTPSCredentialReader interface {
	ReadHost(context.Context, string) (gitcred.HostCredential, bool)
	ReadScoped(context.Context, string, string) (string, bool)
	Discover(context.Context, string, []string) gitcred.HostMaterial
}

// ErrBuildHTTPSAborted reports an operator who explicitly stopped the HTTPS
// candidate prompt. Anonymous HTTPS is the non-interactive fallback; aborting
// an active prompt is instead a request to stop the run.
var ErrBuildHTTPSAborted = errors.New("build_repository_https_credential_selection_aborted: credential selection was aborted")

// BuildHTTPSRequest is one unmatched effective HTTPS repository presented to
// the operator before any repository fetch starts.
type BuildHTTPSRequest struct {
	Skill, Command, Identity string
	Host, DefaultScope       string
}

// BuildHTTPSResolver covers unmatched HTTPS repositories on an operator
// terminal. Returned credentials are run-local even when the resolver also
// persisted their source selection.
type BuildHTTPSResolver func(context.Context, []BuildHTTPSRequest, map[string]gitcred.HostMaterial, BuildHTTPSCredentialReader) (map[string]BuildHTTPSCredentials, error)

// BuildHTTPSSelection is everything one install resolves HTTPS credentials
// from. Environment values are captured once rather than reread after package
// state has been consulted.
type BuildHTTPSSelection struct {
	Override CapturedBuildHTTPSOverride
	Scopes   map[string]config.BuildHTTPSCredential
	Reader   BuildHTTPSCredentialReader
	Resolve  BuildHTTPSResolver

	environment map[string]string
}

func (s BuildHTTPSSelection) String() string {
	return fmt.Sprintf("BuildHTTPSSelection{Override:%s, Scopes:%d, Environment:<redacted>}", s.Override, len(s.Scopes))
}

// GoString returns the same redacted diagnostic representation as String.
func (s BuildHTTPSSelection) GoString() string { return s.String() }

// CaptureBuildHTTPSSelection reads the override and every configured
// token_env at process entry. A nil environ captures absence rather than
// consulting ambient state later.
func CaptureBuildHTTPSSelection(cfg *config.Config, environ func(string) string) BuildHTTPSSelection {
	if environ == nil {
		environ = func(string) string { return "" }
	}
	selection := BuildHTTPSSelection{
		Override: CapturedBuildHTTPSOverride{
			Host: environ(EnvBuildHTTPSHost), secret: environ(EnvBuildHTTPSToken),
		},
		environment: map[string]string{},
	}
	if cfg == nil {
		return selection
	}
	selection.Scopes = make(map[string]config.BuildHTTPSCredential, len(cfg.BuildHTTPS))
	for scope, credential := range cfg.BuildHTTPS {
		selection.Scopes[scope] = credential
		if credential.TokenEnv != "" {
			selection.environment[credential.TokenEnv] = environ(credential.TokenEnv)
		}
	}
	return selection
}

type buildHTTPSKey struct {
	skill   string
	command string
}

// resolveBuildHTTPS resolves each effective HTTPS repository independently.
// Precedence is the covering run-wide override, then the longest configured
// scope, then anonymous transport. Anonymous is deliberately not an error:
// unlike SSH, HTTPS has a real unauthenticated transport and public build
// repositories must remain usable without an operator credential.
func resolveBuildHTTPS(ctx context.Context, selection BuildHTTPSSelection, rows []plannedExternal) (map[buildHTTPSKey]BuildHTTPSCredentials, []string, error) {
	credentials := map[buildHTTPSKey]BuildHTTPSCredentials{}
	var provenance []string
	var missing []plannedExternal
	reader := selection.Reader
	if reader == nil {
		reader = gitcred.Access{}
	}
	for _, row := range rows {
		if !needsBuildHTTPS(row) {
			continue
		}
		key := buildHTTPSKeyFor(row)
		host := buildHTTPSHost(row.effective.Identity)
		if selection.Override.secret != "" && (selection.Override.Host == "" || selection.Override.Host == host) {
			credentials[key] = NewBuildHTTPSCredentials(config.BuildHTTPSDefaultUsername, selection.Override.secret)
			provenance = append(provenance, buildHTTPSProvenance(row, "operator environment override"))
			continue
		}
		configured, matched := config.MatchBuildHTTPS(selection.Scopes, row.effective.Identity)
		if !matched {
			missing = append(missing, row)
			continue
		}
		resolved, err := resolveConfiguredBuildHTTPS(ctx, selection, reader, host, row.effective.Identity, configured)
		if err != nil {
			return nil, nil, err
		}
		credentials[key] = resolved
		provenance = append(provenance, buildHTTPSProvenance(row, fmt.Sprintf("config scope %q", configured.Scope)))
	}
	if len(missing) == 0 {
		return credentials, provenance, nil
	}
	if selection.Resolve == nil {
		for _, row := range missing {
			provenance = append(provenance, buildHTTPSProvenance(row, "anonymous"))
		}
		return credentials, provenance, nil
	}

	requests := make([]BuildHTTPSRequest, 0, len(missing))
	candidates := map[string]gitcred.HostMaterial{}
	for _, row := range missing {
		host := buildHTTPSHost(row.effective.Identity)
		requests = append(requests, BuildHTTPSRequest{
			Skill: row.node.Name, Command: row.command.Name, Identity: row.effective.Identity,
			Host: host, DefaultScope: defaultBuildSSHScope(row.effective.Identity),
		})
		if _, discovered := candidates[host]; !discovered {
			// Discover returns presence-only material. It cannot authorize or
			// disclose a credential, and resolution reads the secret again only
			// after the operator explicitly selects that candidate.
			candidates[host] = reader.Discover(ctx, host, nil)
		}
	}
	chosen, err := selection.Resolve(ctx, requests, candidates, reader)
	if err != nil {
		return nil, nil, err
	}
	for _, row := range missing {
		selected, scope, ok := matchPromptedBuildHTTPS(chosen, row.effective.Identity)
		if !ok {
			provenance = append(provenance, buildHTTPSProvenance(row, "anonymous"))
			continue
		}
		credentials[buildHTTPSKeyFor(row)] = selected
		provenance = append(provenance, buildHTTPSProvenance(row, fmt.Sprintf("operator prompt scope %q", scope)))
	}
	return credentials, provenance, nil
}

func matchPromptedBuildHTTPS(scopes map[string]BuildHTTPSCredentials, identity string) (BuildHTTPSCredentials, string, bool) {
	best := ""
	for scope := range scopes {
		if credentialScopeCovers(scope, identity) && len(scope) > len(best) {
			best = scope
		}
	}
	if best == "" {
		return BuildHTTPSCredentials{}, "", false
	}
	return scopes[best], best, true
}

func resolveConfiguredBuildHTTPS(ctx context.Context, selection BuildHTTPSSelection, reader BuildHTTPSCredentialReader, host, identity string, credential config.BuildHTTPSCredential) (BuildHTTPSCredentials, error) {
	username := credential.EffectiveUsername()
	switch credential.Token {
	case config.TokenSourceGitCredentials:
		material, ok := reader.ReadHost(ctx, host)
		if !ok {
			return BuildHTTPSCredentials{}, fmt.Errorf(
				"build_https scope %q selected for %s has no Git HTTPS credential for host %q; store one with 'git credential approve' for protocol=https and host=%s, or change/remove that build_https scope",
				credential.Scope, identity, host, host)
		}
		if credential.Username == "" {
			username = material.Username
		}
		return NewBuildHTTPSCredentials(username, material.Secret), nil
	case config.TokenSourceKeyring:
		secret, ok := reader.ReadScoped(ctx, credential.Scope, host)
		if !ok {
			return BuildHTTPSCredentials{}, fmt.Errorf(
				"build_https scope %q selected for %s has no manager credential; run 'curator config build-https login %s', or change/remove that build_https scope",
				credential.Scope, identity, credential.Scope)
		}
		return NewBuildHTTPSCredentials(username, secret), nil
	default:
		secret := selection.environment[credential.TokenEnv]
		if secret == "" {
			return BuildHTTPSCredentials{}, fmt.Errorf(
				"build_https scope %q selected for %s but environment variable %s is empty; set %s, or change/remove that build_https scope",
				credential.Scope, identity, credential.TokenEnv, credential.TokenEnv)
		}
		return NewBuildHTTPSCredentials(username, secret), nil
	}
}

func needsBuildHTTPS(row plannedExternal) bool {
	return row.effective.IdentityKind == "network-git" && row.effective.Transport == "https"
}

func buildHTTPSKeyFor(row plannedExternal) buildHTTPSKey {
	return buildHTTPSKey{skill: row.node.Name, command: row.command.Name}
}

func buildHTTPSHost(identity string) string {
	host, _, _ := strings.Cut(identity, "/")
	return host
}

func buildHTTPSProvenance(row plannedExternal, source string) string {
	return fmt.Sprintf("external build https: %s (command %q of skill %q) <- %s",
		row.effective.Identity, row.command.Name, row.node.Name, source)
}
