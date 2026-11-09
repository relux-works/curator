package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLauncherClaudeEnvDiagnosticScope(t *testing.T) {
	for _, schema := range []int{1, 2, 3, 5, 7, 8, 9} {
		t.Run(fmt.Sprintf("malformed-schema-%d", schema), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "selected.json")
			text := fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{},}`, schema)
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path, nil)
			if err == nil || LauncherDiagnostic(err).Error() != "malformed JSON in config "+path || IsLauncherSelectionRefusal(err) {
				t.Fatalf("unrelated malformed file lost its path or legacy discovery: %v", err)
			}
		})
	}
	for _, member := range []string{"", `,"claude_env":` + claudeEnvControl} {
		for _, tc := range []struct{ name, extra, extensions, want string }{
			{"notification", "", `"notification":{"backend":"bogus"}`, "extensions: invalid launcher selection"},
			{"proxy", "", `"proxy":{"profile":42}`, "extensions: invalid launcher selection"},
			{"extension-key", "", `"unknown":"DO_NOT_ECHO"`, "extensions: invalid launcher selection"},
			{"locale", `,"preferred_locale":42`, `"notification":{"backend":"off"}`, "preferred_locale: must be null or a 1-64 character ASCII locale selector"},
			{"top-level", `,"unknown":"DO_NOT_ECHO"`, `"notification":{"backend":"off"}`, "config: unsupported top-level configuration field"},
		} {
			t.Run(tc.name+fmt.Sprint(member != ""), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "selected.json")
				text := `{"schema_version":9,"skills_root":"x","projects":{}` + tc.extra + `,"extensions":{` + tc.extensions + member + `}}`
				if err := os.WriteFile(path, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
				_, err := Load(path, nil)
				if err == nil || LauncherDiagnostic(err).Error() != tc.want {
					t.Fatalf("unrelated diagnostic changed: %v, want %s", err, tc.want)
				}
				if IsLauncherSelectionRefusal(err) != (member != "") {
					t.Fatal("discovery flag must follow explicit member independently of rendering")
				}
			})
		}
	}
	t.Run("non-object", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "selected.json")
		if err := os.WriteFile(path, []byte(`[]`), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := Load(path, nil)
		if err == nil || err.Error() != "config "+path+" must be a JSON object" || IsLauncherSelectionRefusal(err) {
			t.Fatalf("non-object diagnostic: %v", err)
		}
	})
	t.Run("missing-system", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "selected.json")
		missing := filepath.Join(t.TempDir(), "missing-system.json")
		t.Setenv("CURATOR_SYSTEM_CONFIG", missing)
		for _, text := range []string{`{"schema_version":9,"skills_root":"x","projects":{}}`, claudeEnvCarrier(claudeEnvControl)} {
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path, nil)
			if !errors.Is(err, ErrConfigNotFound) || LauncherDiagnostic(err).Error() != "global config not found: "+missing {
				t.Fatalf("missing-system identity/path lost: %v", err)
			}
		}
	})
}

func TestLauncherClaudeEnvMalformedSelectionSignal(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		explicit   bool
	}{
		{"legacy", `{"schema_version":5,"projects":{},}`, false},
		{"name-in-value", `{"schema_version":5,"skills_root":"extensions.claude_env",}`, false},
		{"nested-lookalike", `{"projects":{"sample":{"extensions":{"claude_env":{}}}},}`, false},
		{"other-extension", `{"extensions":{"proxy":{"profile":"claude_env"}},}`, false},
		{"member-malformed", claudeEnvCarrier(`{"BASH_MAX_OUTPUT_LENGTH":`), true},
		{"member-escaped", `{"extensions":{"\u0063laude_env":`, true},
		{"member-duplicate", claudeEnvCarrier(`{"BASH_MAX_OUTPUT_LENGTH":"8000","BASH_MAX_OUTPUT_LENGTH":"1"}`), true},
		{"duplicate-extension-after-null", `{"extensions":null,"extensions":{"claude_env":{}}}`, true},
		{"duplicate-extension-after-array", `{"extensions":[{},"claude_env"],"extensions":{"claude_env":{}}}`, true},
		{"member-number-overflow", claudeEnvCarrier(`{"BASH_MAX_OUTPUT_LENGTH":1e10000}`), true},
		{"unrelated-malformation-after-member", claudeEnvCarrier(claudeEnvControl) + ` {}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "selected.json")
			if err := os.WriteFile(path, []byte(tc.text), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path, nil)
			if err == nil || LauncherDiagnostic(err).Error() != "malformed JSON in config "+path || IsLauncherSelectionRefusal(err) != tc.explicit {
				t.Fatalf("wrong neutral diagnostic or discovery property: %v", err)
			}
		})
	}
}

func TestLauncherClaudeEnvInvalidCorpusControls(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "machine-claude-env-v1", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vectors []claudeEnvVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, v := range vectors {
		if v.Entry != "declaration" || v.Want != ClaudeEnvInvalid {
			continue
		}
		count++
		t.Run(v.ID, func(t *testing.T) {
			if got := claudeEnvDeclarationVerdict(t, claudeEnvControl, claudeEnvCarrier); got != "accept" {
				t.Fatalf("adjacent control: %s", got)
			}
			if got := claudeEnvDeclarationVerdict(t, v.Input, claudeEnvCarrier); got != v.Want {
				t.Fatalf("intended defect: %s", got)
			}
			for _, mutation := range []struct {
				name    string
				carrier func(string) string
			}{
				{"unknown-top-level", func(member string) string {
					return strings.Replace(claudeEnvCarrier(member), `"projects":{}`, `"projects":{},"unknown":"DO_NOT_ECHO"`, 1)
				}},
				{"bad-notification", func(member string) string {
					return strings.Replace(claudeEnvCarrier(member), `"extensions":{`, `"extensions":{"notification":{"backend":"bogus"},`, 1)
				}},
			} {
				t.Run(mutation.name, func(t *testing.T) {
					for _, member := range []string{claudeEnvControl, v.Input} {
						if got := claudeEnvDeclarationVerdict(t, member, mutation.carrier); got != "invalid-containing-file" {
							t.Fatalf("unrelated defect counted as claude_env success: %s", got)
						}
					}
				})
			}
		})
	}
	if count != 8 {
		t.Fatalf("expected exactly eight declaration-invalid vectors, got %d", count)
	}
}

// Transactional editors use loadObject directly: it must retain the same
// explicit-selection discovery marker and wrapped missing-policy identity as
// Load. A valid policy is the nearby control; failed reads create no policy.
func TestLauncherClaudeEnvSharedLoader(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			dir := t.TempDir()
			policy := filepath.Join(dir, "policy.json")
			bytes := []byte(`{"schema_version":2,"locked":[]}`)
			if err := os.WriteFile(policy, bytes, 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CURATOR_SYSTEM_CONFIG", policy)
			text := `{"schema_version":9,"skills_root":"x","projects":{}}`
			if explicit {
				text = claudeEnvCarrier(claudeEnvControl)
			}
			var object map[string]any
			if err := json.Unmarshal([]byte(text), &object); err != nil {
				t.Fatal(err)
			}
			if cfg, err := loadObject(filepath.Join(dir, "machine.json"), object, nil); err != nil || cfg == nil {
				t.Fatalf("positive common-loader control: %v", err)
			}
			actual, err := os.ReadFile(policy)
			if err != nil || string(actual) != string(bytes) {
				t.Fatal("common loader changed policy")
			}
			missing := filepath.Join(dir, "missing.json")
			t.Setenv("CURATOR_SYSTEM_CONFIG", missing)
			cfg, err := loadObject(filepath.Join(dir, "machine.json"), object, nil)
			if cfg != nil || !errors.Is(err, ErrConfigNotFound) || IsLauncherSelectionRefusal(err) != explicit {
				t.Fatalf("shared loader lost error identity or signal: %v", err)
			}
			if _, err := os.Lstat(missing); !os.IsNotExist(err) {
				t.Fatal("failed read created policy")
			}
			actual, err = os.ReadFile(policy)
			if err != nil || string(actual) != string(bytes) {
				t.Fatal("failed read changed policy")
			}
		})
	}
}
