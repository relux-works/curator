package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

// The exact status command keeps a recorded default/custom native origin under
// a managed caller. Missing/mismatched authority and retargeted auth still fail
// without modifying fixture state. Other status rows (including absent external
// providers in a test binary) retain their verdicts: this qualifies the origin
// row and caller invariance, not an all-green matrix or host adoption.
func TestEnvStatusManagedCallerRecordedOrigin(t *testing.T) {
	for _, mode := range []string{"default", "custom", "missing-origin", "wrong-lock", "wrong-target", "retarget"} {
		t.Run(mode, func(t *testing.T) {
			source, manager := profileHome(t)
			operator := t.TempDir()
			t.Setenv("HOME", operator)
			t.Setenv("USERPROFILE", operator)
			t.Setenv("CURATOR_CONFIG", source.path)
			t.Chdir(t.TempDir())
			if mode == "default" {
				t.Setenv("CODEX_HOME", filepath.Join(operator, ".codex"))
			}
			native := os.Getenv("CODEX_HOME")
			installCLIEnvProfile(t, source)
			for _, adapter := range envregistry.Registry {
				if adapter.ID == envregistry.Muse {
					continue // An absent optional Muse home is informational.
				}
				if code, _, stderr := runProfile(t, source, "env", "resolve", adapter.ID, "--repair"); code != exitOK {
					t.Fatalf("provision %s: code=%d stderr=%s", adapter.ID, code, stderr)
				}
			}
			managed := envprofile.ManagedHomeDir(manager, "acme", envregistry.CodexCLI)
			baselineCode, baselineJSON, baselineStderr := runProfile(t, source, "env", "status", "--check", "--json")
			var baseline envprofile.Status
			if err := json.Unmarshal([]byte(baselineJSON), &baseline); err != nil {
				t.Fatalf("baseline status: %v stderr=%s", err, baselineStderr)
			}
			baselineCurrent := false
			for _, row := range baseline.Homes {
				if row.Profile == "acme" && row.Environment == envregistry.CodexCLI {
					baselineCurrent = row.Current
				}
			}
			if !baselineCurrent {
				t.Fatalf("native-caller positive control is not current: %s", baselineJSON)
			}
			originPath := filepath.Join(managed, envorigin.Name)
			if mode == "missing-origin" {
				if err := os.Remove(originPath); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "wrong-lock" || mode == "wrong-target" {
				raw, err := os.ReadFile(originPath)
				if err != nil {
					t.Fatal(err)
				}
				record, err := envorigin.Parse(raw)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "wrong-lock" {
					record.Binding.LockSHA256 = "sha256:" + strings.Repeat("b", 64)
				} else {
					record.Binding.ManagedTarget = filepath.Join(managed, "other")
				}
				raw, err = record.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(originPath, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "retarget" {
				link := filepath.Join(managed, "auth.json")
				if err := os.Remove(link); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(t.TempDir(), "auth.json"), link); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("CODEX_HOME", managed)
			before := []map[string]helpTreeEntry{nativeOwnerSnapshot(t, manager), nativeOwnerSnapshot(t, native)}
			code, stdout, stderr := runProfile(t, source, "env", "status", "--check", "--json")
			if after := []map[string]helpTreeEntry{nativeOwnerSnapshot(t, manager), nativeOwnerSnapshot(t, native)}; !reflect.DeepEqual(before, after) {
				t.Fatal("status modified fixture manager or native state")
			}
			if got := os.Getenv("CODEX_HOME"); got != managed {
				t.Fatal("status changed the admitted caller selector")
			}
			var status envprofile.Status
			if err := json.Unmarshal([]byte(stdout), &status); err != nil {
				t.Fatalf("status JSON: %v, code=%d stderr=%s", err, code, stderr)
			}
			var selected *envprofile.HomeState
			for i := range status.Homes {
				row := &status.Homes[i]
				if row.Profile == "acme" && row.Environment == envregistry.CodexCLI {
					selected = row
				}
			}
			if selected == nil || !selected.Provisioned {
				t.Fatalf("missing provisioned Codex row: %+v", selected)
			}
			if mode == "default" || mode == "custom" {
				if code != baselineCode || status.NonCurrent != baseline.NonCurrent || !selected.Current {
					t.Fatalf("valid origin: code=%d selected=%+v status=%s stderr=%s", code, selected, stdout, stderr)
				}
				return
			}
			want := envorigin.DiagInvalid
			if mode == "missing-origin" {
				want = envorigin.DiagRequired
			} else if mode == "retarget" {
				want = envregistry.DiagCredentialConflict
			}
			if code != exitFail || !status.NonCurrent || selected.Current || !strings.Contains(strings.Join(selected.Findings, "\n"), want) {
				t.Fatalf("expected %s refusal: code=%d selected=%+v stderr=%s", want, code, selected, stderr)
			}
		})
	}
}
