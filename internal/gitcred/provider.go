package gitcred

// Provider-addressed HTTPS secrets for repository-transport-v1 endpoints.
//
// A policy endpoint carries an opaque `authentication` provider identifier
// (repository-transport §2). The secret for that provider on one host lives
// in the operator's own Git credential machinery under a provider namespace,
// exactly like manager-scoped entries: no new secret store, no secrets in
// policy, manifests, locks, receipts, logs, or command arguments. The
// namespace keeps provider entries disjoint from both the operator's own
// host credential and manager scope entries, so none answers for another.

import (
	"context"
	"strings"
)

// ProviderNamespacePrefix opens the username of every provider-held HTTPS
// entry. A provider identifier cannot contain the "/" that separates a
// scope's host from its path, and the distinct prefix keeps provider
// entries disjoint from scope entries even for host-only scopes.
const ProviderNamespacePrefix = "curator-provider-https:"

// ProviderUsername is the username a provider-held credential for provider
// lives under. It is "" unless provider is an admissible opaque identifier,
// so a malformed reference never addresses the credential machinery.
func ProviderUsername(provider string) string {
	if !ValidProvider(provider) {
		return ""
	}
	return ProviderNamespacePrefix + provider
}

// ReadProvider reads the provider-held HTTPS credential for provider on
// host through the operator's Git credential machinery.
//
// The answer is accepted only when the helper answers for the username that
// was asked about. A helper free to answer a near miss must not hand a
// different record back as this provider's material. Like every other read
// here, an absent credential degrades into "nothing here" rather than an
// error; the transport layer classifies that as an unavailable
// authentication method.
func (a Access) ReadProvider(ctx context.Context, provider, host string) (HostCredential, bool) {
	username := ProviderUsername(provider)
	if username == "" || !portableValue(host) {
		return HostCredential{}, false
	}
	answer, ok := a.call(ctx, "fill",
		field{"protocol", "https"}, field{"host", host}, field{"username", username})
	if !ok || answer["password"] == "" {
		return HostCredential{}, false
	}
	if answered := answer["username"]; answered != "" && answered != username {
		return HostCredential{}, false
	}
	name := answer["username"]
	if name == "" {
		name = DefaultUsername
	}
	return HostCredential{Username: name, Secret: answer["password"]}, true
}

// providerNamespaceOwns reports whether username names a provider-held
// entry. ReadHost uses it to keep excluding namespaced records from the
// operator's own credential view, the way scope entries are excluded.
func providerNamespaceOwns(username string) bool {
	return strings.HasPrefix(username, ProviderNamespacePrefix)
}
