package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Raw version tokens enforce the new carrier rule without changing released
// predecessor numeric forms. Refusal never rewrites the selected file.
func TestLauncherNotificationVersionTokens(t *testing.T) {
	for _, token := range []string{"8", "8.0", "8e0", "7.0", "7e0", "5.0", "5e0", "1.0"} {
		t.Run(token, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			text := fmt.Sprintf(`{"schema_version":%s,"skills_root":"x","projects":{}}`, token)
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path, nil)
			if token == "8.0" || token == "8e0" {
				var diagnostic *Diagnostic
				if cfg != nil || !errors.As(err, &diagnostic) || diagnostic.Code != "config_schema_unsupported" || diagnostic.Path != "schema_version" || err.Error() != "schema_version: must be an integer" {
					t.Fatalf("integer-token refusal: %v", err)
				}
			} else if err != nil || cfg == nil {
				t.Fatalf("valid numeric neighbour: %v", err)
			}
			actual, _ := os.ReadFile(path)
			if string(actual) != text {
				t.Fatal("read changed source")
			}
		})
	}
}

// Fixed diagnostic codes/paths preserve compatibility text and precedence;
// valid carriers beside the negatives prevent reject-all implementations.
func TestLauncherNotificationDiagnosticContract(t *testing.T) {
	for _, row := range []struct{ name, text, code, path string }{
		{"valid", `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":{"backend":"off"}}}`, "", ""},
		{"json", `{"schema_version":8,"schema_version":8,"HOSTILE_SECRET":1}`, "config_json_invalid", "$"},
		{"schema", `{"schema_version":6}`, "config_schema_unsupported", "schema_version"},
		{"old-extensions", `{"schema_version":5,"extensions":{}}`, "launcher_extensions_schema", "extensions"},
		{"selection", `{"schema_version":8,"extensions":{"HOSTILE_SECRET":"HOSTILE_SECRET"}}`, "launcher_selection_invalid", "extensions"},
		{"shape-before-version", `{"schema_version":9,"HOSTILE_SECRET":1}`, "config_shape_invalid", "$"},
		{"shape-before-selection", `{"schema_version":8,"HOSTILE_SECRET":1,"extensions":{"notification":null}}`, "config_shape_invalid", "$"},
	} {
		t.Run(row.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(row.text), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path, nil)
			if row.code == "" {
				if err != nil || cfg == nil {
					t.Fatal(err)
				}
			} else {
				var diagnostic *Diagnostic
				if cfg != nil || !errors.As(err, &diagnostic) || diagnostic.Code != row.code || diagnostic.Path != row.path || strings.Contains(err.Error(), "HOSTILE_SECRET") {
					t.Fatalf("diagnostic: %v", err)
				}
			}
			actual, _ := os.ReadFile(path)
			if string(actual) != row.text {
				t.Fatal("read changed source")
			}
		})
	}
}

// System metadata and selection locks cannot supply machine authority, and
// hostile lock suffixes are never reflected. An unrelated valid lock works.
func TestLauncherNotificationSystemLockDiagnostic(t *testing.T) {
	machine := `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":{"backend":"off"}}}`
	for _, policy := range []string{`{"schema_version":2,"locked":[]}`, `{"schema_version":2,"locked":["extensions"]}`, `{"schema_version":2,"locked":["extensions.notification"]}`, `{"schema_version":2,"locked":["extensions.HOSTILE_SECRET"]}`, `{"schema_version":2,"locked":[],"extensions":{}}`} {
		t.Run(policy, func(t *testing.T) {
			cfg, _, err := loadWithSystem(t, machine, policy)
			if policy == `{"schema_version":2,"locked":[]}` {
				if cfg == nil || err != nil {
					t.Fatal(err)
				}
				return
			}
			var diagnostic *Diagnostic
			if cfg != nil || !errors.As(err, &diagnostic) || diagnostic.Code != "launcher_selection_machine_only" || diagnostic.Path != "extensions" || err.Error() != "extensions: launcher selections belong only to the selected machine configuration" {
				t.Fatalf("system refusal: %v", err)
			}
		})
	}
}

// Unspecified system-policy failures retain their text/cause without borrowing
// selected-file corpus codes. Valid policy and typed machine-only locks above
// remain positive and independently specified controls. Reads change no bytes.
func TestLauncherSystemPolicyDiagnosticBoundary(t *testing.T) {
	for _, row := range []struct{ name, policy, want string }{
		{"valid", `{"schema_version":2,"locked":[]}`, ""},
		{"missing", "", "global config not found: %s"},
		{"malformed", "{", "malformed JSON in config %s"},
		{"non-object", "[]", "config %s must be a JSON object"},
		{"unsupported", `{"schema_version":9,"locked":[]}`, "schema_version: system config %s requires schema_version 1 or 2"},
		{"locked-shape", `{"schema_version":2,"locked":"wrong"}`, "locked: system config %s field 'locked' must be a list of strings"},
	} {
		t.Run(row.name, func(t *testing.T) {
			root := t.TempDir()
			machine := filepath.Join(root, "machine.json")
			policy := filepath.Join(root, "system.json")
			text := `{"schema_version":8,"skills_root":"x","projects":{}}`
			if err := os.WriteFile(machine, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			if row.policy != "" {
				if err := os.WriteFile(policy, []byte(row.policy), 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
			cfg, err := Load(machine, nil)
			if row.want == "" {
				if cfg == nil || err != nil {
					t.Fatalf("valid policy: %v", err)
				}
			} else {
				want := row.want
				if strings.Contains(want, "%s") {
					want = fmt.Sprintf(want, policy)
				}
				var diagnostic *Diagnostic
				if cfg != nil || err == nil || err.Error() != want || errors.As(err, &diagnostic) || errors.Is(err, ErrConfigNotFound) != (row.name == "missing") {
					t.Fatalf("untyped system-policy failure: %v", err)
				}
			}
			actual, readErr := os.ReadFile(machine)
			if readErr != nil || string(actual) != text {
				t.Fatal("policy read changed machine bytes")
			}
			actual, readErr = os.ReadFile(policy)
			if row.name == "missing" {
				if !os.IsNotExist(readErr) {
					t.Fatal("policy read created missing policy")
				}
			} else if readErr != nil || string(actual) != row.policy {
				t.Fatal("policy read changed policy bytes")
			}
		})
	}
}
