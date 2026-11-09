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

const notificationSelectionFixture = `{"backend":"host","contract":"curator-host-notification/1","workflow_reference":"host-notification-inbox-v1","workflow_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`

// Adding a project and writing an unrelated raw setting retain schema8 and all
// selected metadata through the actual file load/edit/write/reload path.
func TestLauncherNotificationNormalProjectEdit(t *testing.T) {
	dir := t.TempDir()
	path, project := filepath.Join(dir, "selected.json"), filepath.Join(dir, "project")
	if err := os.Mkdir(project, 0o700); err != nil {
		t.Fatal(err)
	}
	text := `{"schema_version":8,"skills_root":"x","projects":{},"preferred_locale":"en","extensions":{"notification":` + notificationSelectionFixture + `,"proxy":{"profile":"private"}}}`
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
	if after.Schema != SchemaVersion8 || !reflect.DeepEqual(before.Extensions, after.Extensions) || after.PreferredLocale != "en" || after.AdapterMode != "copy" || len(after.Projects) != 1 {
		t.Fatal("normal project edit downgraded or lost selected config")
	}
}

// Each recorded carrier version retains its projection; only schema 8 carries
// notification metadata. Parsing/exporting metadata establishes no enrollment.
func TestLauncherNotificationVersionedProjection(t *testing.T) {
	for _, schema := range []int{1, 2, 3, 5, 7, 8} {
		t.Run(fmt.Sprint(schema), func(t *testing.T) {
			cfg := loadText(t, fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{}}`, schema))
			want := "curator.launcher-config/1"
			if schema == 8 {
				want = "curator.launcher-config/2"
			}
			if got := cfg.LauncherProjection(); got.Schema != want || got.ConfigSchema != schema {
				t.Fatalf("versioned projection: got %+v, want %s/schema%d", got, want, schema)
			}
		})
	}
	for _, selection := range []string{
		`{"backend":"off"}`,
		notificationSelectionFixture,
		strings.Replace(notificationSelectionFixture, "host-notification-inbox-v1", strings.Repeat("a", 128), 1),
		strings.Replace(notificationSelectionFixture, "host-notification-inbox-v1", "ABCxyz012._:/@+-", 1),
	} {
		cfg := loadText(t, `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":`+selection+`}}`)
		payload, err := json.Marshal(cfg.LauncherProjection())
		if err != nil {
			t.Fatal(err)
		}
		var got struct {
			Schema     string                     `json:"schema"`
			Extensions map[string]json.RawMessage `json:"extensions"`
		}
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatal(err)
		}
		var wantValue, gotValue any
		if err := json.Unmarshal([]byte(selection), &wantValue); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(got.Extensions["notification"], &gotValue); err != nil {
			t.Fatal(err)
		}
		if got.Schema != "curator.launcher-config/2" || !reflect.DeepEqual(gotValue, wantValue) {
			t.Fatalf("selected metadata changed: %s", payload)
		}
	}
}

// The production file loader rejects each malformed/foreign variant without
// changing the file or quoting hostile keys/values. Nearby valid forms pass above.
func TestLauncherNotificationClosedSelection(t *testing.T) {
	bad := []string{
		`null`, `[]`, `{}`, `{"backend":null}`, `{"backend":"native"}`,
		`{"backend":"off","contract":"curator-host-notification/1"}`,
		`{"backend":"off","backend":"host"}`,
		`{"backend":"host"}`,
		strings.Replace(notificationSelectionFixture, `"curator-host-notification/1"`, `"curator-host-notification/2"`, 1),
		strings.Replace(notificationSelectionFixture, `"host-notification-inbox-v1"`, `""`, 1),
		strings.Replace(notificationSelectionFixture, `"host-notification-inbox-v1"`, `"invalid reference"`, 1),
		strings.Replace(notificationSelectionFixture, `"host-notification-inbox-v1"`, `"invalid\u0000reference"`, 1),
		strings.Replace(notificationSelectionFixture, `"host-notification-inbox-v1"`, `"\u03b1"`, 1),
		strings.Replace(notificationSelectionFixture, `"host-notification-inbox-v1"`, `"`+strings.Repeat("a", 129)+`"`, 1),
		strings.Replace(notificationSelectionFixture, strings.Repeat("a", 64), strings.Repeat("A", 64), 1),
		strings.Replace(notificationSelectionFixture, strings.Repeat("a", 64), strings.Repeat("a", 63), 1),
		strings.Replace(notificationSelectionFixture, `"workflow_sha256":"`+strings.Repeat("a", 64)+`"`, `"workflow_sha256":17`, 1),
		strings.TrimSuffix(notificationSelectionFixture, "}") + `,"endpoint":"HOSTILE_SECRET_VALUE"}`,
		strings.TrimSuffix(notificationSelectionFixture, "}") + `,"HOSTILE_SECRET_KEY":"HOSTILE_SECRET_VALUE"}`,
	}
	for i, selection := range bad {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "selected.json")
			text := `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":` + selection + `}}`
			if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path, nil)
			if cfg != nil || err == nil || strings.Contains(err.Error(), "HOSTILE_SECRET") {
				t.Fatalf("unsafe notification refusal: %v, %v", cfg, err)
			}
			wantRefusal := "extensions"
			if selection == `{"backend":"off","backend":"host"}` {
				// The existing closed JSON reader sanitizes duplicate-key errors
				// before the extension parser receives the object.
				wantRefusal = "malformed JSON in config "
			}
			if !strings.Contains(err.Error(), wantRefusal) {
				t.Fatalf("wrong refusal, notification selection was not evaluated: %v", err)
			}
			got, readErr := os.ReadFile(path)
			if readErr != nil || string(got) != text {
				t.Fatal("rejected config read changed fixture")
			}
		})
	}
	for _, schema := range []int{1, 2, 3, 5, 6, 7} {
		path := filepath.Join(t.TempDir(), "old.json")
		text := fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{},"extensions":{"notification":%s}}`, schema, notificationSelectionFixture)
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		// Older carriers must refuse for their own version/selection reason,
		// rather than passing this control on an unrelated loader failure.
		want := "extensions: requires manager-config schema_version 7 or 8"
		if schema == 6 {
			want = "schema_version: unsupported config schema_version 6; this config requires a newer tool"
		} else if schema == 7 {
			want = "extensions: invalid launcher selection"
		}
		if cfg, err := Load(path, nil); cfg != nil || err == nil || err.Error() != want {
			t.Fatalf("schema %d refusal: got %v, want %q", schema, err, want)
		}
		if actual, err := os.ReadFile(path); err != nil || string(actual) != text {
			t.Fatal("old-carrier refusal changed input")
		}
	}
}

// Schema8 preserves the existing versioned state-hash waiver, and still refuses
// an unsupported hash version rather than silently changing its interpretation.
func TestLauncherNotificationStateHashWaiverCompatibility(t *testing.T) {
	text := `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":` + notificationSelectionFixture + `},"environments":{"secret_material_waivers":[{"pin":"` + strings.Repeat("b", 64) + `","hash_version":2,"file":"skill.md","span":[0,1],"reason":"fixture"}]}}`
	cfg := loadText(t, text)
	if len(cfg.Env.SecretWaivers) != 1 || cfg.Env.SecretWaivers[0].HashVersion != 2 {
		t.Fatal("schema8 lost the versioned state-hash waiver")
	}
	path := filepath.Join(t.TempDir(), "bad-waiver.json")
	if err := os.WriteFile(path, []byte(strings.Replace(text, `"hash_version":2`, `"hash_version":3`, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if cfg, err := Load(path, nil); cfg != nil || err == nil || !strings.Contains(err.Error(), "hash_version") || !strings.Contains(err.Error(), "must be 2") {
		t.Fatalf("wrong waiver refusal: %v %v", cfg, err)
	}
}

// Schema8 does not move launcher selection authority into a system policy file.
func TestLauncherNotificationSystemSourceRefused(t *testing.T) {
	_, _, err := loadWithSystem(t, `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":`+notificationSelectionFixture+`}}`, `{"schema_version":2,"locked":[],"extensions":{"notification":`+notificationSelectionFixture+`}}`)
	if err == nil || !strings.Contains(err.Error(), "extensions") {
		t.Fatalf("system notification selection was not refused: %v", err)
	}
}

// Effective rendering round-trips the selected reference/digest alongside the
// existing credential selector; projections own their copy.
func TestLauncherNotificationEffectiveRoundTrip(t *testing.T) {
	text := `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":` + notificationSelectionFixture + `,"proxy":{"profile":"private"}},"environments":{"credential_store":{"p":{"codex_cli":"file"}}}}`
	cfg := loadText(t, text)
	projected := cfg.LauncherProjection()
	if projected.Extensions.Notification == nil {
		t.Fatal("typed projection has no selected notification")
	}
	projected.Extensions.Notification.WorkflowReference = "changed"
	raw := cfg.EffectiveJSON()
	// EffectiveJSON intentionally exposes only effective settings, not the
	// complete file envelope. Restore that caller-owned envelope for Parse.
	raw["schema_version"], raw["skills_root"], raw["projects"] = 8, "x", map[string]any{}
	// EffectiveJSON carries typed extension values; the real JSON boundary must
	// be exercised rather than passing Go structs directly to a raw-object parser.
	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	again, err := Parse(object, cfg.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Extensions, again.Extensions) || !reflect.DeepEqual(cfg.Env.CredentialStore, again.Env.CredentialStore) {
		t.Fatal("effective round trip lost selected metadata")
	}
	if strings.Contains(string(data), `"changed"`) {
		t.Fatal("projection aliases its source config")
	}
}
