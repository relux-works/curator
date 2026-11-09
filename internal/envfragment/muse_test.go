package envfragment

import (
	"github.com/relux-works/curator/internal/envregistry"
	"path/filepath"
	"strings"
	"testing"
)

// The boundary now admits native with four home paths and no backend override,
// alongside the existing five-value file form. Missing/unsupported policy,
// mismatched or empty backend literals, and foreign paths/channels still refuse.
func TestMuseFragmentBoundary(t *testing.T) {
	adapter, err := envregistry.ByID(envregistry.Muse)
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"valid", "valid-native", "native-empty-literal", "native-keychain-literal", "empty-literal", "keychain-policy", "unset-policy", "missing-data", "wrong-parent", "HOME", "channel", "missing-policy", "native-policy", "missing-literal", "keychain-literal", "false-root"} {
		t.Run(state, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "environments")
			parent := filepath.Join(root, "acme", "muse")
			fragment := &Fragment{Environment: envregistry.Muse, Env: map[string]string{}}
			fragment.CredentialStore = &CredentialStore{Selection: envregistry.CredentialStoreFile}
			fragment.Env["TBH_CREDENTIAL_BACKEND"] = "file"
			for _, variable := range adapter.HomeVariables {
				fragment.Env[variable[0]] = filepath.Join(parent, variable[1])
			}
			switch state {
			case "missing-policy":
				fragment.CredentialStore = nil
			case "valid-native":
				fragment.CredentialStore.Selection = envregistry.CredentialStoreNative
				delete(fragment.Env, "TBH_CREDENTIAL_BACKEND")
			case "native-empty-literal":
				fragment.CredentialStore.Selection = envregistry.CredentialStoreNative
				fragment.Env["TBH_CREDENTIAL_BACKEND"] = ""
			case "native-keychain-literal":
				fragment.CredentialStore.Selection = envregistry.CredentialStoreNative
				fragment.Env["TBH_CREDENTIAL_BACKEND"] = "keychain"
			case "empty-literal":
				fragment.Env["TBH_CREDENTIAL_BACKEND"] = ""
			case "keychain-policy":
				fragment.CredentialStore.Selection = envregistry.CredentialStoreKeychain
			case "unset-policy":
				fragment.CredentialStore.Selection = ""
			case "native-policy":
				fragment.CredentialStore.Selection = envregistry.CredentialStoreNative
			case "missing-literal":
				delete(fragment.Env, "TBH_CREDENTIAL_BACKEND")
			case "keychain-literal":
				fragment.Env["TBH_CREDENTIAL_BACKEND"] = "keychain"
			case "false-root":
				fragment.RootContext = &RootContext{Exclusive: false}
			case "missing-data":
				delete(fragment.Env, "XDG_DATA_HOME")
			case "wrong-parent":
				fragment.Env["XDG_DATA_HOME"] = filepath.Join(parent, "config")
			case "HOME":
				fragment.Env["HOME"] = parent
			case "channel":
				fragment.SystemPrompt = &SystemPrompt{Path: filepath.Join(parent, "unverified.md")}
			}
			err := CheckBoundary(adapter, root, fragment)
			if (err == nil) != (state == "valid" || state == "valid-native") {
				t.Fatalf("boundary=%v", err)
			}
			switch state {
			case "missing-policy", "native-policy", "missing-literal", "keychain-literal", "native-empty-literal", "native-keychain-literal", "empty-literal", "keychain-policy", "unset-policy":
				if !strings.Contains(err.Error(), envregistry.DiagCredentialStoreUnsupported) {
					t.Fatalf("policy mismatch must refuse as unsupported credential store: %v", err)
				}
			}
			if state == "valid" || state == "valid-native" {
				lines := strings.Split(strings.TrimSpace(string(fragment.EnvFormat(adapter))), "\n")
				wantLines := 4
				if state == "valid" {
					wantLines = 5
					if lines[len(lines)-1] != "TBH_CREDENTIAL_BACKEND=file" {
						t.Fatalf("missing backend literal: %v", lines)
					}
				}
				if len(lines) != wantLines {
					t.Fatalf("wrong environment line count: %v", lines)
				}
				for i, variable := range adapter.HomeVariables {
					if lines[i] != variable[0]+"="+filepath.Join(parent, variable[1]) {
						t.Fatalf("variable order: %v", lines)
					}
				}
			}
		})
	}
}

// Admitting Muse's native form must not admit its backend variable for another
// adapter. Each native home-only fragment is a valid positive control.
func TestMuseBackendRemainsAdapterScoped(t *testing.T) {
	for _, id := range []string{envregistry.ClaudeCode, envregistry.CodexCLI, envregistry.Pi, envregistry.OpenCode} {
		t.Run(id, func(t *testing.T) {
			adapter, err := envregistry.ByID(id)
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			fragment := &Fragment{Environment: id, Env: map[string]string{adapter.EnvVar: filepath.Join(root, "home")}}
			if err := CheckBoundary(adapter, root, fragment); err != nil {
				t.Fatalf("native control: %v", err)
			}
			fragment.Env["TBH_CREDENTIAL_BACKEND"] = "file"
			if err := CheckBoundary(adapter, root, fragment); err == nil || !strings.Contains(err.Error(), "not registry-declared") {
				t.Fatalf("foreign backend: %v", err)
			}
		})
	}
}
