//go:build unix

package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Real plan/apply/restore preserves the carrier and every non-auth selection;
// absent, off and host are distinct valid neighbours. No credential is opened.
func TestClaudeAuthNotificationRoundTrip(t *testing.T) {
	for _, notification := range []string{"", `{"backend":"off"}`, notificationSelectionFixture} {
		for _, schema := range []int{7, 8} {
			if schema == 7 && notification != "" {
				continue
			}
			t.Run(fmt.Sprintf("schema%d/%s", schema, notification), func(t *testing.T) {
				member := ""
				if notification != "" {
					member = `,"notification":` + notification
				}
				original := fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{},"extensions":{"proxy":{"profile":"private"},"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"previous","account":"previous"}}%s}}`, schema, member)
				path := authEditFixture(t, original)
				before, err := Load(path, nil)
				if err != nil {
					t.Fatal(err)
				}
				plan, err := PlanClaudeAuth(path, authEditSelection())
				if err != nil || plan.SchemaFrom != schema || plan.SchemaTo != schema {
					t.Fatalf("plan carrier: %+v %v", plan, err)
				}
				result, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
				if err != nil {
					t.Fatal(err)
				}
				after, err := Load(path, nil)
				if err != nil || after.Schema != schema || !reflect.DeepEqual(before.Extensions.Notification, after.Extensions.Notification) || !reflect.DeepEqual(before.Extensions.Proxy, after.Extensions.Proxy) || after.Extensions.ClaudeAuth.Store.Service != "fixture-service" {
					t.Fatalf("selection preservation: %+v %v", after, err)
				}
				_, err = RestoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: result.BackupPath, ExpectCurrentSHA256: result.ResultSHA256, ExpectBackupSHA256: result.SourceSHA256})
				if err != nil {
					t.Fatal(err)
				}
				actual, _ := os.ReadFile(path)
				if string(actual) != original {
					t.Fatal("restore changed original bytes")
				}
				backup, _ := os.ReadFile(result.BackupPath)
				if string(backup) != original {
					t.Fatal("restore changed recovery copy")
				}
			})
		}
	}
}

// A stale apply or restore cannot overwrite later notification choices or
// downgrade schema8. The unchanged round trips above are positive controls.
func TestClaudeAuthNotificationRefusals(t *testing.T) {
	for _, kind := range []string{"stale-apply", "later-notification", "backup-downgrade"} {
		t.Run(kind, func(t *testing.T) {
			original := `{"schema_version":8,"skills_root":"x","projects":{},"extensions":{"notification":` + notificationSelectionFixture + `}}`
			path := authEditFixture(t, original)
			request := authEditRequest(t, path)
			if kind == "stale-apply" {
				changed := strings.Replace(original, "host-notification-inbox-v1", "later", 1)
				if err := os.WriteFile(path, []byte(changed), 0600); err != nil {
					t.Fatal(err)
				}
				_, err := ApplyClaudeAuth(context.Background(), request)
				actual, _ := os.ReadFile(path)
				if !errors.Is(err, ErrClaudeAuthStale) || string(actual) != changed || len(authEditBackups(t, path)) != 0 {
					t.Fatalf("stale apply changed state: %v", err)
				}
				return
			}
			result, err := ApplyClaudeAuth(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			current, _ := os.ReadFile(path)
			backup, _ := os.ReadFile(result.BackupPath)
			if kind == "later-notification" {
				current = bytes.Replace(current, []byte("host-notification-inbox-v1"), []byte("later"), 1)
			} else {
				backup = []byte(`{"schema_version":7,"skills_root":"x","projects":{}}`)
			}
			if err := os.WriteFile(path, current, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(result.BackupPath, backup, 0600); err != nil {
				t.Fatal(err)
			}
			_, err = RestoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: result.BackupPath, ExpectCurrentSHA256: claudeAuthSHA(current), ExpectBackupSHA256: claudeAuthSHA(backup)})
			actual, _ := os.ReadFile(path)
			saved, _ := os.ReadFile(result.BackupPath)
			if !errors.Is(err, ErrClaudeAuthEditRefused) || !bytes.Equal(actual, current) || !bytes.Equal(saved, backup) {
				t.Fatalf("restore refusal changed state: %v", err)
			}
		})
	}
}

// With no non-auth selections to mask the schema guard, an equivalent older
// backup must not downgrade the admitted carrier. Same-carrier restore works.
func TestClaudeAuthNotificationRestoreCarrierGuard(t *testing.T) {
	assertClaudeAuthRestoreCarrierGuard(t, 8, []int{8, 7})
}

func assertClaudeAuthRestoreCarrierGuard(t *testing.T, currentSchema int, backupSchemas []int) {
	t.Helper()
	for _, savedSchema := range backupSchemas {
		t.Run(fmt.Sprintf("schema%d/backup%d", currentSchema, savedSchema), func(t *testing.T) {
			original := fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{}}`, currentSchema)
			path := authEditFixture(t, original)
			result, err := ApplyClaudeAuth(context.Background(), authEditRequest(t, path))
			if err != nil {
				t.Fatal(err)
			}
			current, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			saved := []byte(fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{}}`, savedSchema))
			if err := os.WriteFile(result.BackupPath, saved, 0600); err != nil {
				t.Fatal(err)
			}
			before, err := claudeAuthObject(current)
			if err != nil {
				t.Fatal(err)
			}
			after, err := claudeAuthObject(saved)
			if err != nil {
				t.Fatal(err)
			}
			if !claudeAuthEquivalent(path, before, after) {
				t.Fatal("fixture must reach the carrier guard with equivalent non-auth state")
			}
			restored, err := RestoreClaudeAuth(context.Background(), ClaudeAuthRestoreRequest{Path: path, BackupPath: result.BackupPath, ExpectCurrentSHA256: claudeAuthSHA(current), ExpectBackupSHA256: claudeAuthSHA(saved)})
			actual, readErr := os.ReadFile(path)
			backup, backupErr := os.ReadFile(result.BackupPath)
			if savedSchema == currentSchema {
				if err != nil || restored == nil || string(actual) != original {
					t.Fatalf("same-carrier positive restore: %v", err)
				}
			} else if !errors.Is(err, ErrClaudeAuthEditRefused) || restored != nil || !bytes.Equal(actual, current) {
				t.Fatalf("carrier downgrade must refuse without writes: %v", err)
			}
			if readErr != nil || backupErr != nil || !bytes.Equal(backup, saved) || len(authEditBackups(t, path)) != 1 {
				t.Fatal("restore changed recovery state")
			}
		})
	}
}
