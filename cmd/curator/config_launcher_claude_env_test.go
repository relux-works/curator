package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// Selected schema9 edits and metadata queries preserve all launcher selections
// on every target OS without opening the platform-absolute credential reference.
func TestConfigLauncherClaudeEnvSelectedEditsAndStatus(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "selected.json")
	credentialPath := filepath.Join(root, "credential-reference")
	encodedPath, err := json.Marshal(credentialPath)
	if err != nil || !filepath.IsAbs(credentialPath) {
		t.Fatalf("fixture credential path must be target-absolute: %v", err)
	}
	text := `{"schema_version":9,"skills_root":"x","projects":{},"preferred_locale":"en","extensions":{"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"8000"},"notification":{"backend":"off"},"proxy":{"profile":"private"},"claude_auth":{"type":"oauth","store":{"type":"file","path":` + string(encodedPath) + `}}}}`
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	// Synthetic profile, lock and native settings witnesses must never change.
	sentinels := []string{"profile.json", "context-lock.json", "settings.json", "hooks.json"}
	for _, name := range sentinels {
		if err := os.WriteFile(filepath.Join(root, name), []byte("unchanged\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv(config.ClaudeEnvSetting, "17")
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
		{"config", "launcher", "--status"},
		{"config", "launcher", "--runtime", "claude_code"},
		{"config", "launcher", "--runtime", "codex_cli"},
		{"config", "launcher", "--runtime", "opencode"},
		{"config", "launcher", "--runtime", "pi"},
		{"config", "launcher", "--runtime", "muse"},
	} {
		var out, errs bytes.Buffer
		if code := run(args, fileConfigSource(path), &out, &errs); code != exitOK || errs.Len() != 0 {
			t.Fatalf("command %v failed: code=%d stderr=%q", args, code, errs.String())
		}
		if len(args) == 2 && args[0] == "config" {
			var got map[string]any
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got["schema"] != "curator.launcher-config/3" || got["config_schema"] != float64(9) || got["config_path"] != path || !reflect.DeepEqual(got["extensions"], initial["extensions"]) {
				t.Fatal("selected fields lost")
			}
		}
		if len(args) == 3 && args[2] == "--status" {
			var rows []config.LauncherClaudeEnvStatus
			if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
				t.Fatal(err)
			}
			if len(rows) != 5 || rows[0].State != "configured" || rows[0].ConfigPath != path || rows[1].State != "runtime-ineligible" || strings.Contains(out.String(), "8000") {
				t.Fatal("unsafe or misleading status")
			}
		}
		if len(args) == 4 && args[0] == "config" && args[2] == "--runtime" {
			got, err := config.ParseClaudeEnvContribution(out.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if args[3] == "claude_code" {
				want = "8000"
			}
			if got.Runtime != args[3] || got.Env.BashMaxOutputLength != want {
				t.Fatal("wrong runtime contribution")
			}
		}
	}
	if _, err := os.Stat(credentialPath); !os.IsNotExist(err) {
		t.Fatal("metadata query created the absent credential reference")
	}
	if os.Getenv(config.ClaudeEnvSetting) != "17" {
		t.Fatal("manager environment was mutated")
	}
	for _, name := range sentinels {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(data) != "unchanged\n" {
			t.Fatal("selection touched profile/lock/native state")
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var final map[string]any
	if err := json.Unmarshal(raw, &final); err != nil {
		t.Fatal(err)
	}
	if final["schema_version"] != float64(9) || final["preferred_locale"] != "en" || !reflect.DeepEqual(final["extensions"], initial["extensions"]) {
		t.Fatal("ordinary edits downgraded or discarded selections")
	}
}

// This test uses Core's real provider resolution and exec plumbing. Its child
// is explicitly a synthetic umbrella provider, never a Claude runtime witness.
// The nearby valid case proves that no-child refusals reach a launch-capable
// boundary, rather than passing only because the provider is unavailable.
func TestConfigLauncherClaudeEnvRefusesBeforeProviderDispatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("synthetic POSIX provider")
	}
	root := t.TempDir()
	t.Setenv(config.ClaudeEnvSetting, "17")
	path := filepath.Join(root, "selected.json")
	provider := filepath.Join(root, "curator-run")
	if err := os.WriteFile(provider, []byte("#!/bin/sh\nprintf 'PROVIDER_EXECUTED BASH_MAX_OUTPUT_LENGTH=%s\\n' \"${BASH_MAX_OUTPUT_LENGTH-unset}\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	prefix := `{"schema_version":9,"skills_root":"x","projects":{},"extensions":{"claude_env":`
	valid := prefix + `{"BASH_MAX_OUTPUT_LENGTH":"8000"}}}`
	fixtures := []struct{ text, want string }{
		{valid, ""},
		{strings.Replace(valid, `"schema_version":9`, `"schema_version":"9"`, 1), "schema_version: must be an integer"},
		{strings.Replace(valid, `"schema_version":9`, `"schema_version":8`, 1), "extensions.claude_env: " + config.ClaudeEnvInvalid},
		{prefix + `{"BASH_MAX_OUTPUT_LENGTH":"0"}}}`, "extensions.claude_env: " + config.ClaudeEnvValueNonpositive},
		{prefix + `{"BASH_MAX_OUTPUT_LENGTH":8000}}`, "malformed JSON in config " + path},
		{prefix + `{"BASH_MAX_OUTPUT_LENGTH":"8000","BASH_MAX_OUTPUT_LENGTH":"1"}}}`, "malformed JSON in config " + path},
		{prefix + `{"BASH_MAX_OUTPUT_LENGTH":"8000","BASH_ENV":"DO_NOT_ECHO"}}}`, "extensions.claude_env: " + config.ClaudeEnvNameNotAllowed},
		{prefix + `{"BASH_MAX_OUTPUT_LENGTH":"DO_NOT_ECHO"}}}`, "extensions.claude_env: " + config.ClaudeEnvValueNotInteger},
		{prefix + `{"BASH_MAX_OUTPUT_LENGTH":"8000"},"UNLISTED":"DO_NOT_ECHO"}}`, "extensions: invalid launcher selection"},
		{strings.Replace(valid, "8000", string([]byte{0xff}), 1), "malformed JSON in config " + path},
	}
	for i, fx := range fixtures {
		if err := os.WriteFile(path, []byte(fx.text), 0600); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"run", "claude"}, {"run", "claude", "--resume", "synthetic-session"}, {"run", "codex"}} {
			var out, errs bytes.Buffer
			command := cli{config: fileConfigSource(path), stdout: &out, stderr: &errs, providerPathOverride: &root, userHome: func() (string, error) { return root, nil }}
			code := command.run(args)
			if fx.want == "" {
				if code != exitOK || out.String() != "PROVIDER_EXECUTED BASH_MAX_OUTPUT_LENGTH=17\n" {
					t.Fatalf("valid provider control %d did not execute", i)
				}
			} else if code != exitFail || out.Len() != 0 || errs.String() != "curator: "+fx.want+"\n" {
				t.Fatalf("fixture %d failed refusal, redaction or no-child check", i)
			}
			after, err := os.ReadFile(path)
			if err != nil || string(after) != fx.text {
				t.Fatal("refusal or dispatch changed selected config")
			}
		}
	}
}

func TestConfigLauncherClaudeEnvReselectionAndInvalidNoWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selected.json")
	for _, value := range []string{"8000", "1", "0"} {
		text := `{"schema_version":9,"skills_root":"x","projects":{},"extensions":{"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"` + value + `"}}}`
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"config", "launcher", "--runtime", "claude_code"}, {"config", "launcher", "--status"}, {"env", "config", "set", "overlay_default_weight", "3"}} {
			var out, errs bytes.Buffer
			code := run(args, fileConfigSource(path), &out, &errs)
			if value == "0" {
				after, err := os.ReadFile(path)
				if code != exitFail || out.Len() != 0 || err != nil || string(after) != text || errs.String() != "curator: extensions.claude_env: claude_env_value_nonpositive\n" {
					t.Fatal("invalid explicit selection fell back or wrote")
				}
			} else if code != exitOK {
				t.Fatal("valid reselected configuration refused")
			} else if len(args) == 4 && !strings.Contains(out.String(), `"BASH_MAX_OUTPUT_LENGTH": "`+value+`"`) {
				t.Fatal("stale selected value")
			}
		}
	}
}
