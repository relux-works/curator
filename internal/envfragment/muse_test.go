package envfragment

import (
	"github.com/relux-works/curator/internal/envregistry"
	"path/filepath"
	"strings"
	"testing"
)

// The successor boundary admits only the five-value file projection. Missing
// policy, a conflicting literal and foreign variables/channels remain refused.
func TestMuseFragmentBoundary(t *testing.T) {
	adapter, err := envregistry.ByID(envregistry.Muse)
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"valid", "missing-data", "wrong-parent", "HOME", "channel", "missing-policy", "native-policy", "missing-literal", "keychain-literal", "false-root"} {
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
			if (err == nil) != (state == "valid") {
				t.Fatalf("boundary=%v", err)
			}
			if state == "valid" {
				lines := strings.Split(strings.TrimSpace(string(fragment.EnvFormat(adapter))), "\n")
				if len(lines) != 5 || lines[4] != "TBH_CREDENTIAL_BACKEND=file" {
					t.Fatalf("missing last backend literal: %v", lines)
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
