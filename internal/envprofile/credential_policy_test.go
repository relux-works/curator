package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// Resolve refuses each unavailable explicit mapping before provisioning; native
// Codex remains a real fragment-producing positive control on synthetic homes.
func TestCredentialStorePolicyActivation(t *testing.T) {
	fx := writeManagedFixture(t, "store-policy")
	for _, id := range []string{envregistry.ClaudeCode, envregistry.CodexCLI, envregistry.Muse, envregistry.Pi, envregistry.OpenCode} {
		for _, selection := range []envregistry.CredentialStoreSelection{envregistry.CredentialStoreFile, envregistry.CredentialStoreKeychain} {
			if id == envregistry.Muse && selection == envregistry.CredentialStoreFile {
				continue // Admitted Muse file controls live in TestMuseCredentialPolicyMatrix.
			}
			t.Run(id+"/"+string(selection), func(t *testing.T) {
				req := fx.request(id)
				req.Repair = true
				req.Machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{fx.profile: {id: selection}}
				result, err := Resolve(req)
				want := envregistry.DiagCredentialStoreUnsupported
				if id == envregistry.CodexCLI && selection == envregistry.CredentialStoreFile {
					want = envregistry.DiagCodexSettingsUnsupported
				}
				if err == nil || !strings.Contains(err.Error(), want) || result != nil {
					t.Fatalf("Resolve = %v, %v; want %s and no fragment", result, err, want)
				}
				if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, id)); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("refused selector provisioned a home: %v", err)
				}
			})
		}
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	result, err := Resolve(req)
	if err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("native control = %v, %v", result, err)
	}
}

// Explicit file cannot turn an unreadable, malformed or unknown native selector
// into the absent-key default. Positive readable/absent-key still reaches the
// intentional delivery refusal, not a native-inspection diagnostic.
func TestCredentialStorePolicyNativeInspection(t *testing.T) {
	fx := writeManagedFixture(t, "store-inspection")
	for _, tc := range []struct {
		name, text, want string
		readErr          error
	}{
		{"absent key", "model = 'test'", envregistry.DiagCodexSettingsUnsupported, nil},
		{"keyring", "cli_auth_credentials_store = 'keyring'", envregistry.DiagCodexSettingsUnsupported, nil},
		{"auto", "cli_auth_credentials_store = 'auto'", envregistry.DiagCodexSettingsUnsupported, nil},
		{"unknown", "cli_auth_credentials_store = 'other'", envregistry.DiagCredentialUnsupported, nil},
		{"malformed", "[broken", envregistry.DiagSeedUnreadable, nil},
		{"permission", "", envregistry.DiagSeedUnreadable, os.ErrPermission},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			req.Machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{fx.profile: {envregistry.CodexCLI: envregistry.CredentialStoreFile}}
			req.Machine.Isolation[fx.profile] = map[string]string{envregistry.CodexCLI: envregistry.IsolationIsolated}
			path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
			if err := os.WriteFile(path, []byte(tc.text), 0o600); err != nil {
				t.Fatal(err)
			}
			if tc.readErr != nil {
				req.readRegularFile = func(name string) (stateread.File, error) {
					if name == path {
						return stateread.File{}, tc.readErr
					}
					return stateread.ReadRegularFile(name)
				}
			}
			result, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), tc.want) || result != nil {
				t.Fatalf("Resolve = %v, %v; want %s", result, err, tc.want)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != tc.text {
				t.Fatal("refusal modified native configuration")
			}
			if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("refusal provisioned a home")
			}
		})
	}
}
