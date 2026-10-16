package config

import (
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/verr"
)

// parseCredentialStore validates the v5 machine subtree. CLI aliases must
// already be canonicalized; config files cannot acquire an alternate identity.
func parseCredentialStore(raw any) (map[string]map[string]envregistry.CredentialStoreSelection, error) {
	profiles, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("environments.credential_store", "must be an object")
	}
	result := map[string]map[string]envregistry.CredentialStoreSelection{}
	for profile, rawEntry := range profiles {
		label := "environments.credential_store." + profile
		if !identifiers.Valid(profile) {
			return nil, verr.New("environments.credential_store", "profile %q %s", profile, identifiers.Rule)
		}
		entries, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		result[profile] = map[string]envregistry.CredentialStoreSelection{}
		for id, rawSelection := range entries {
			if _, err := envregistry.ByID(id); err != nil {
				return nil, verr.New(label, "unknown canonical environment %q", id)
			}
			value, ok := rawSelection.(string)
			selection := envregistry.CredentialStoreSelection(value)
			if !ok || !selection.Valid() {
				return nil, verr.New(label+"."+id, "must be native, file, or keychain")
			}
			result[profile][id] = selection
		}
	}
	return result, nil
}
