package envregistry

import "fmt"

// CredentialStoreSelection is the machine-owned selector, not a native backend
// or a runtime capability claim. Native is the default, including for old configs.
type CredentialStoreSelection string

// Machine selector values and activation diagnostics (login-readiness section 2).
const (
	CredentialStoreNative          CredentialStoreSelection = "native"
	CredentialStoreFile            CredentialStoreSelection = "file"
	CredentialStoreKeychain        CredentialStoreSelection = "keychain"
	DiagCredentialStoreUnsupported                          = "environment_credential_store_unsupported"
	DiagCodexSettingsUnsupported                            = "codex_settings_unsupported"
)

// Valid reports the closed login-readiness section 2 selector vocabulary.
func (s CredentialStoreSelection) Valid() bool {
	return s == CredentialStoreNative || s == CredentialStoreFile || s == CredentialStoreKeychain
}

// CredentialStoreFor reads only the machine selector. It never probes vendor
// state, normalizes profile names, or interprets package content as policy.
func (c MachineConfig) CredentialStoreFor(profile, id string) (CredentialStoreSelection, error) {
	if _, err := ByID(id); err != nil {
		return "", err
	}
	selection, present := c.CredentialStore[profile][id]
	if !present {
		return CredentialStoreNative, nil
	}
	if !selection.Valid() {
		return "", fmt.Errorf("%s: invalid machine credential_store selector for %s/%s", DiagCredentialStoreUnsupported, profile, id)
	}
	return selection, nil
}
