package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const launcherExtensionsFixture = `{"proxy":{"profile":"private"},"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"claude","account":"operator"}}}`

// The complete successor admits either section or both, and normal project and
// environment edits retain both sections and unrelated admitted user settings.
func TestLauncherExtensionsNormalConfigEdits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "selected.json")
	project := filepath.Join(dir, "project")
	if err := os.Mkdir(project, 0o700); err != nil {
		t.Fatal(err)
	}
	text := `{"schema_version":7,"skills_root":"x","projects":{},"preferred_locale":"en","environments":{"credential_store":{"p":{"codex_cli":"file"}}},"extensions":` + launcherExtensionsFixture + `}`
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := AddProject(path, "sample", project, nil); err != nil {
		t.Fatal(err)
	}
	raw, err := ReadRaw(path)
	if err != nil {
		t.Fatal(err)
	}
	raw["adapter_mode"] = "copy"
	if _, err := WriteRaw(path, raw); err != nil {
		t.Fatal(err)
	}
	after, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.Extensions, after.Extensions) || after.Extensions.Proxy.Profile != "private" || after.Extensions.ClaudeAuth.Store.Account != "operator" {
		t.Fatal("normal config edits lost launcher selections")
	}
	if after.PreferredLocale != "en" || after.AdapterMode != "copy" || len(after.Projects) != 1 || !reflect.DeepEqual(before.Env.CredentialStore, after.Env.CredentialStore) {
		t.Fatal("normal edit lost unrelated settings")
	}
	effective := after.EffectiveJSON()
	if _, ok := effective["extensions"]; !ok {
		t.Fatal("effective projection lost extensions")
	}
	if _, ok := effective["environments"].(map[string]any)["credential_store"]; !ok {
		t.Fatal("successor lost credential-store capability")
	}
}

// Absent selections remain absent; malformed, unknown, duplicate and mixed
// forms refuse without reflecting hostile member names or supplied values.
func TestLauncherExtensionsClosedGrammar(t *testing.T) {
	valid := []string{``, `,"extensions":{}`, `,"extensions":{"proxy":{"profile":"private"}}`, `,"extensions":{"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"s","account":"a"}}}`, `,"extensions":` + launcherExtensionsFixture}
	for _, member := range valid {
		cfg := loadText(t, `{"schema_version":7,"skills_root":"x","projects":{}`+member+`}`)
		if member == "" && cfg.Extensions != nil {
			t.Fatal("absence acquired a selection")
		}
	}
	bad := []string{
		`null`, `[]`, `{"HOSTILE_SECRET_KEY":true}`, `{"proxy":null}`, `{"proxy":{"profile":""}}`, `{"proxy":{"profile":"x","HOSTILE_SECRET_KEY":"HOSTILE_SECRET_VALUE"}}`,
		`{"proxy":{"profile":"x","profile":"y"}}`,
		`{"claude_auth":{"type":"token","store":{"type":"file","path":"/tmp/x"}}}`,
		`{"claude_auth":{"type":"oauth","store":{"type":"file","path":"relative"}}}`,
		`{"claude_auth":{"type":"oauth","store":{"type":"file","path":"/tmp/../x"}}}`,
		`{"claude_auth":{"type":"oauth","store":{"type":"file","path":"/tmp/x","service":"s"}}}`,
		`{"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"s","account":"a","path":"/tmp/x"}}}`,
		`{"claude_auth":{"type":"oauth","token":"HOSTILE_SECRET_VALUE","store":{"type":"keychain","service":"s","account":"a"}}}`,
		`{"HOSTILE_SECRET_KEY":1,"HOSTILE_SECRET_KEY":2}`,
	}
	fixtureRoot := t.TempDir()
	for index, value := range bad {
		t.Run(value, func(t *testing.T) {
			// Keep hostile input out of the fixture path: diagnostics may name
			// the selected operator path, but must not quote JSON member data.
			path := filepath.Join(fixtureRoot, fmt.Sprintf("config-%d.json", index))
			text := `{"schema_version":7,"skills_root":"x","projects":{},"extensions":` + value + `}`
			if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path, nil)
			if err == nil || cfg != nil {
				t.Fatal("invalid launcher selection accepted")
			}
			if strings.Contains(err.Error(), "HOSTILE_SECRET") {
				t.Fatal("refusal reflected hostile configuration material")
			}
			got, _ := os.ReadFile(path)
			if string(got) != text {
				t.Fatal("rejected read changed configuration")
			}
		})
	}
}

// Old manager versions cannot carry the new extension object, and a system
// file cannot create a second authority or silently default an auth selection.
func TestLauncherExtensionsVersionAndSourceBoundaries(t *testing.T) {
	for _, version := range []string{"1", "2", "3", "5", "6"} {
		path := filepath.Join(t.TempDir(), "config.json")
		text := `{"schema_version":` + version + `,"skills_root":"x","projects":{},"extensions":` + launcherExtensionsFixture + `}`
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		if cfg, err := Load(path, nil); err == nil || cfg != nil {
			t.Fatalf("old version %s accepted extensions", version)
		}
	}
	_, _, err := loadWithSystem(t, `{"schema_version":7,"skills_root":"x","projects":{},"extensions":`+launcherExtensionsFixture+`}`, `{"schema_version":2,"locked":[],"extensions":`+launcherExtensionsFixture+`}`)
	if err == nil || !strings.Contains(err.Error(), "extensions") {
		t.Fatalf("system selection was not refused: %v", err)
	}
}

// A typed projection encodes references only and is detached from its config;
// serializing it never resolves a credential source.
func TestLauncherExtensionsTypedProjection(t *testing.T) {
	cfg := loadText(t, `{"schema_version":7,"skills_root":"x","projects":{},"extensions":`+launcherExtensionsFixture+`}`)
	p := cfg.LauncherProjection()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["schema"] != "curator.launcher-config/1" || got["config_schema"] != float64(7) {
		t.Fatalf("invalid typed projection: %s", data)
	}
	p.Extensions.Proxy.Profile = "changed"
	if cfg.Extensions.Proxy.Profile != "private" {
		t.Fatal("projection aliases mutable config")
	}
}
