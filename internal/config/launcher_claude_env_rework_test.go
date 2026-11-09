package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The landed carrier requires integer token 9, while its integer spelling is
// admitted. Read-only validation reports the fixed code/path and changes no bytes.
func TestLauncherClaudeEnvVersionTokens(t *testing.T) {
	for _, token := range []string{"9", "9.0", "9e0"} {
		t.Run(token, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "machine.json")
			text := `{"schema_version":` + token + `,"skills_root":"x","projects":{}}`
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path, nil)
			if token == "9" {
				if cfg == nil || cfg.Schema != 9 || err != nil {
					t.Fatalf("integer positive control: %v", err)
				}
			} else {
				var diagnostic *Diagnostic
				if cfg != nil || !errors.As(err, &diagnostic) || diagnostic.Code != "config_schema_unsupported" || diagnostic.Path != "schema_version" || err.Error() != "schema_version: must be an integer" {
					t.Fatalf("integer-token diagnostic: %v", err)
				}
			}
			actual, err := os.ReadFile(path)
			if err != nil || string(actual) != text {
				t.Fatal("token validation changed selected bytes")
			}
		})
	}
}

// Forbidden claude_env on schema8 wins independently of map order, even beside
// an unknown extension. Valid schema8/9 neighbours prevent reject-all success.
func TestLauncherClaudeEnvDiagnosticOrder(t *testing.T) {
	for _, row := range []struct {
		name, text string
		invalid    bool
	}{
		{"schema8-valid", `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{}}`, false},
		{"schema9-valid", `{"schema_version":9,"skills_root":"x","projects":{},"extensions":{"claude_env":{}}}`, false},
		{"schema8-mixed", `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"UNLISTED":"DO_NOT_ECHO","claude_env":{}}}`, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "machine.json")
			if err := os.WriteFile(path, []byte(row.text), 0600); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 200; i++ {
				cfg, err := Load(path, nil)
				if row.invalid {
					var member launcherClaudeEnvError
					var diagnostic *Diagnostic
					if cfg != nil || !errors.As(err, &member) || member.code != ClaudeEnvInvalid || !errors.As(err, &diagnostic) || diagnostic.Code != "config_shape_invalid" || diagnostic.Path != "$" || err.Error() != "extensions.claude_env: claude_env_invalid" {
						t.Fatalf("repeat %d: unstable member diagnostic: %v", i, err)
					}
				} else if err != nil || cfg == nil {
					t.Fatalf("valid neighbour: %v", err)
				}
			}
			actual, err := os.ReadFile(path)
			if err != nil || string(actual) != row.text {
				t.Fatal("diagnostic reads changed selected bytes")
			}
		})
	}
}
