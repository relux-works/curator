package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/mcpnative"
)

// Characterize Muse through the shared resolve path before its policy moves
// behind the harness contract.
func TestMuseSharedProjectionCharacterization(t *testing.T) {
	fx, authLink, nativeAuth := museFixture(t)
	result, err := Resolve(fx.request(envregistry.Muse))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Fragment        string            `json:"fragment"`
		Env             map[string]string `json:"env"`
		CredentialStore *struct {
			Selection string `json:"selection"`
		} `json:"credential_store"`
		RootContext *struct {
			Exclusive bool `json:"exclusive"`
		} `json:"root_context"`
		SystemPrompt json.RawMessage `json:"system_prompt"`
		MCP          json.RawMessage `json:"mcp"`
	}
	if err := json.Unmarshal(result.Document, &got); err != nil {
		t.Fatal(err)
	}
	managed := ManagedHomeDir(fx.home, fx.profile, envregistry.Muse)
	wantEnv := map[string]string{
		"XDG_CONFIG_HOME":        filepath.Join(managed, "config"),
		"XDG_DATA_HOME":          filepath.Join(managed, "data"),
		"XDG_STATE_HOME":         filepath.Join(managed, "state"),
		"XDG_CACHE_HOME":         filepath.Join(managed, "cache"),
		"TBH_CREDENTIAL_BACKEND": "file",
	}
	if got.Fragment != "launch-env-fragment-v4" || !reflect.DeepEqual(got.Env, wantEnv) {
		t.Fatalf("fragment=%q env=%v, want v4 and %v", got.Fragment, got.Env, wantEnv)
	}
	if got.CredentialStore == nil || got.CredentialStore.Selection != "file" || got.RootContext == nil || !got.RootContext.Exclusive {
		t.Fatalf("credential/root-context projection: %+v / %+v", got.CredentialStore, got.RootContext)
	}
	if len(got.SystemPrompt) != 0 && string(got.SystemPrompt) != "null" || len(got.MCP) != 0 && string(got.MCP) != "null" {
		t.Fatalf("unqualified surfaces emitted: system_prompt=%s mcp=%s", got.SystemPrompt, got.MCP)
	}
	if got := inPlaceSkillsDir(envregistry.Muse); got != profileSkillsSurfaceDir {
		t.Fatalf("shared native skills path=%q", got)
	}
	if got := mcpnative.Path(envregistry.Muse); got != "" {
		t.Fatalf("unqualified native MCP path=%q", got)
	}
	if target, err := os.Readlink(authLink); err != nil || target != nativeAuth {
		t.Fatalf("managed credential route=%q err=%v", target, err)
	}
	if bytes, err := os.ReadFile(nativeAuth); err != nil || string(bytes) != "synthetic-native\n" {
		t.Fatalf("native credential changed: %q err=%v", bytes, err)
	}
}

// Missing shared native credentials refuse before any managed-home change.
func TestMuseSharedCredentialRefusalCharacterization(t *testing.T) {
	fx := writeManagedFixture(t, "muse-characterization")
	fx.native[envregistry.Muse] = t.TempDir()
	req := fx.request(envregistry.Muse)
	req.Repair = true
	before := snapshotCredentialScope(t, fx)
	result, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) || result != nil {
		t.Fatalf("missing shared source: result=%v err=%v", result, err)
	}
	if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
		t.Fatal("refusal changed managed or native credential state")
	}
}
