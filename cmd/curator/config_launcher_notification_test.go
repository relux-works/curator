package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Real file-backed config export and normal edits retain schema8, the exact
// notification metadata, existing auth/proxy and unrelated user settings.
func TestConfigLauncherNotificationNormalEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selected.json")
	store := filepath.Join(t.TempDir(), "absent-credential")
	text := `{"schema_version":8,"skills_root":"x","projects":{},"preferred_locale":"en","extensions":{"notification":{"backend":"host","contract":"curator-host-notification/1","workflow_reference":"host-notification-inbox-v1","workflow_sha256":"` + strings.Repeat("a", 64) + `"},"proxy":{"profile":"private"},"claude_auth":{"type":"oauth","store":{"type":"file","path":` + strconvQuote(store) + `}}}}`
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	var initial map[string]any
	if err := json.Unmarshal([]byte(text), &initial); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"config", "launcher"},
		{"env", "config", "set", "credential_store.sample.codex_cli", "file"},
		{"env", "config", "set", "overlay_default_weight", "9"},
		{"env", "config", "unset", "overlay_default_weight"},
		{"config", "launcher"},
	} {
		var out, errs bytes.Buffer
		if code := run(args, fileConfigSource(path), &out, &errs); code != exitOK || errs.Len() != 0 {
			t.Fatalf("%v: code%d: %s", args, code, &errs)
		}
		if args[0] == "config" {
			var got map[string]any
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got["schema"] != "curator.launcher-config/2" || got["config_schema"] != float64(8) || got["config_path"] != path || !reflect.DeepEqual(got["extensions"], initial["extensions"]) {
				t.Fatal("metadata export changed selected fields")
			}
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var final map[string]any
	if err := json.Unmarshal(data, &final); err != nil {
		t.Fatal(err)
	}
	if final["schema_version"] != float64(8) || final["preferred_locale"] != "en" || !reflect.DeepEqual(final["extensions"], initial["extensions"]) {
		t.Fatal("normal edit downgraded or lost selected config")
	}
	if _, err := os.Lstat(store); !os.IsNotExist(err) {
		t.Fatalf("unexpected credential fixture: %v", err)
	}
}

// The selected CLI refuses an unknown notification field with no partial JSON,
// hostile-data reflection, or config mutation; the valid counterpart is above.
func TestConfigLauncherNotificationRefusal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selected.json")
	text := `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":{"backend":"off","HOSTILE_SECRET_KEY":"HOSTILE_SECRET_VALUE"}}}`
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	if code := run([]string{"config", "launcher"}, fileConfigSource(path), &out, &errs); code != exitFail || out.Len() != 0 || strings.Contains(errs.String(), "HOSTILE_SECRET") || !strings.Contains(errs.String(), "extensions") {
		t.Fatalf("incorrect notification refusal: %d %q %q", code, &out, &errs)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != text {
		t.Fatal("refusal changed selected config")
	}
}
