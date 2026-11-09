//go:build unix

package config

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

// The admitted successor carrier follows the same auth-editor preservation
// rule as schema8. An invalid neighbouring Claude selection cannot be edited
// into apparent validity, and refusal leaves source/recovery state untouched.
func TestClaudeAuthClaudeEnvRoundTrip(t *testing.T) {
	for _, value := range []string{"8000", "0"} {
		t.Run(value, func(t *testing.T) {
			original := `{"schema_version":9,"skills_root":"x","projects":{},"extensions":{"proxy":{"profile":"private"},"notification":` + notificationSelectionFixture + `,"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"` + value + `"}}}`
			path := authEditFixture(t, original)
			plan, err := PlanClaudeAuth(path, authEditSelection())
			if value == "0" {
				actual, _ := os.ReadFile(path)
				if !errors.Is(err, ErrClaudeAuthEditRefused) || plan != nil || string(actual) != original || len(authEditBackups(t, path)) != 0 {
					t.Fatalf("invalid selection changed state: %v", err)
				}
				return
			}
			if err != nil || plan.SchemaFrom != 9 || plan.SchemaTo != 9 {
				t.Fatalf("schema9 plan: %+v %v", plan, err)
			}
			before, err := Load(path, nil)
			if err != nil {
				t.Fatal(err)
			}
			result, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
			if err != nil {
				t.Fatal(err)
			}
			after, err := Load(path, nil)
			if err != nil || after.Schema != 9 || after.Extensions.ClaudeAuth == nil || !reflect.DeepEqual(before.Extensions.ClaudeEnv, after.Extensions.ClaudeEnv) || !reflect.DeepEqual(before.Extensions.Notification, after.Extensions.Notification) || !reflect.DeepEqual(before.Extensions.Proxy, after.Extensions.Proxy) {
				t.Fatalf("schema9 preservation: %+v %v", after, err)
			}
			// A later environment selection, like notification, cannot be overwritten
			// by restoring an older auth backup even when the current digest is known.
			actual, _ := os.ReadFile(path)
			changed := bytes.Replace(actual, []byte("8000"), []byte("8001"), 1)
			if err := os.WriteFile(path, changed, 0600); err != nil {
				t.Fatal(err)
			}
			_, err = RestoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: result.BackupPath, ExpectCurrentSHA256: claudeAuthSHA(changed), ExpectBackupSHA256: result.SourceSHA256})
			current, _ := os.ReadFile(path)
			backup, _ := os.ReadFile(result.BackupPath)
			if !errors.Is(err, ErrClaudeAuthEditRefused) || !bytes.Equal(current, changed) || string(backup) != original {
				t.Fatalf("changed selection overwritten: %v", err)
			}
			if err := os.WriteFile(path, actual, 0600); err != nil {
				t.Fatal(err)
			}
			_, err = RestoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: result.BackupPath, ExpectCurrentSHA256: result.ResultSHA256, ExpectBackupSHA256: result.SourceSHA256})
			current, _ = os.ReadFile(path)
			backup, _ = os.ReadFile(result.BackupPath)
			if err != nil || string(current) != original || string(backup) != original {
				t.Fatalf("schema9 restore: %v", err)
			}
			if strings.Contains(string(current), "fixture-service") {
				t.Fatal("auth change survived restore")
			}
		})
	}
}
