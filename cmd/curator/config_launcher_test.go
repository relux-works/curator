package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The real file-backed CLI source exports the selected config, preserves both
// selections through normal environment edits, and succeeds with an absent store.
func TestConfigLauncherSelectedSourceAndNormalEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "operator-selected.json")
	store := filepath.Join(t.TempDir(), "not-created-token")
	text := `{"schema_version":7,"skills_root":"x","projects":{},"preferred_locale":"en","extensions":{"proxy":{"profile":"private"},"claude_auth":{"type":"oauth","store":{"type":"file","path":` + strconvQuote(store) + `}}}}`
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	source := fileConfigSource(path)
	runSelected := func(args ...string) (int, string, string) {
		t.Helper()
		var out, errs bytes.Buffer
		code := run(args, source, &out, &errs)
		return code, out.String(), errs.String()
	}
	for _, args := range [][]string{
		{"env", "config", "set", "overlay_default_weight", "9"},
		{"env", "config", "set", "credential_store.sample.codex_cli", "file"},
		{"env", "config", "unset", "overlay_default_weight"},
	} {
		if code, _, err := runSelected(args...); code != exitOK {
			t.Fatalf("%v: %d %s", args, code, err)
		}
	}
	code, out, errs := runSelected("config", "launcher")
	if code != exitOK || errs != "" {
		t.Fatalf("projection: %d %s", code, errs)
	}
	var got struct {
		Schema       string `json:"schema"`
		ConfigSchema int    `json:"config_schema"`
		ConfigPath   string `json:"config_path"`
		Extensions   struct {
			Proxy struct {
				Profile string `json:"profile"`
			} `json:"proxy"`
			ClaudeAuth struct {
				Store struct {
					Path string `json:"path"`
				} `json:"store"`
			} `json:"claude_auth"`
		} `json:"extensions"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Schema != "curator.launcher-config/1" || got.ConfigSchema != 7 || got.ConfigPath != path || got.Extensions.Proxy.Profile != "private" || got.Extensions.ClaudeAuth.Store.Path != store {
		t.Fatal("selected projection lost provenance or selections")
	}
	if _, err := os.Lstat(store); !os.IsNotExist(err) {
		t.Fatalf("projection accessed credential fixture: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"preferred_locale": "en"`) {
		t.Fatal("ordinary edit lost unrelated user setting")
	}
}

func strconvQuote(text string) string { data, _ := json.Marshal(text); return string(data) }

// Projection failure emits no partial JSON or supplied hostile key/value, and
// unrelated arguments are refused before a configuration load.
func TestConfigLauncherRefusalHasNoProjection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selected.json")
	text := `{"schema_version":7,"skills_root":"x","projects":{},"extensions":{"HOSTILE_SECRET_KEY":"HOSTILE_SECRET_VALUE"}}`
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	if code := run([]string{"config", "launcher"}, fileConfigSource(path), &out, &errs); code != exitFail || out.Len() != 0 || strings.Contains(errs.String(), "HOSTILE_SECRET") {
		t.Fatalf("unsafe refusal: %d %q %q", code, out.String(), errs.String())
	}
	out.Reset()
	errs.Reset()
	if code := run([]string{"config", "launcher", "extra"}, fileConfigSource("absent-config"), &out, &errs); code != exitUsage || out.Len() != 0 || !strings.Contains(errs.String(), "takes no arguments") {
		t.Fatalf("argument refusal: %d %q %q", code, out.String(), errs.String())
	}
}
