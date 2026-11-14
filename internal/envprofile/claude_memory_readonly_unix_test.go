//go:build unix

package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// In both home forms, native-memory repair retains unrelated settings and the
// selected settings mode. Verification refuses drift without writing; ordinary
// native mode remains a writable positive control for the read-only opt-in.
func TestClaudeNativeMemoryReadOnlySettings(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, mode := range []string{"native", "read-only"} {
			t.Run(form+"/"+mode, func(t *testing.T) {
				_, req, seed := claudeRepairFixture(t, form)
				req.Machine.ClaudeSettingsMode = map[string]string{req.Profile: mode}
				path := filepath.Join(filepath.Dir(seed), "settings.json")
				before := []byte(`{"operator":true,"claudeMdExcludes":["**/vendor/CLAUDE.md"]}`)
				if err := os.WriteFile(path, before, 0o600); err != nil {
					t.Fatal(err)
				}
				req.Repair = false
				if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), "native memory exclusion") {
					t.Fatalf("missing native exclusion was not refused: %v", err)
				}
				if !bytes.Equal(readClaudeState(t, path), before) {
					t.Fatal("verification wrote settings")
				}
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				wantMode := os.FileMode(0o600)
				if mode == "read-only" {
					wantMode = 0o400
				}
				assertManagedMode(t, path, wantMode)
				var got struct {
					Operator bool     `json:"operator"`
					Excludes []string `json:"claudeMdExcludes"`
				}
				if err := json.Unmarshal(readClaudeState(t, path), &got); err != nil {
					t.Fatal(err)
				}
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatal(err)
				}
				want := filepath.ToSlash(filepath.Join(home, ".claude", "CLAUDE.md"))
				if !got.Operator || len(got.Excludes) != 3 || got.Excludes[0] != "**/vendor/CLAUDE.md" || got.Excludes[1] != want || got.Excludes[2] != filepath.ToSlash(filepath.Join(home, ".claude", "rules"))+"/**" {
					t.Fatalf("repair lost unrelated settings or the exact native exclusion: %+v", got)
				}
				req.Repair = false
				if _, err := Resolve(req); err != nil {
					t.Fatalf("repaired settings are stale: %v", err)
				}
			})
		}
	}
}

// The exclusion publisher preserves every tested existing mode, including
// group-readable and read-only files, and never changes the held old inode.
// An absent file has the ordinary private authored-file positive control.
func TestClaudeNativeMemoryPublicationModes(t *testing.T) {
	for _, mode := range []os.FileMode{0o400, 0o444, 0o600, 0o640} {
		t.Run(fmt.Sprintf("%04o", mode), func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "settings.json")
			before := []byte(`{"operator":true}`)
			if err := os.WriteFile(path, before, mode); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, mode); err != nil {
				t.Fatal(err)
			}
			old, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = old.Close() }()
			if err := publishClaudeSettings(root, "."); err != nil {
				t.Fatal(err)
			}
			assertManagedMode(t, path, mode)
			info, err := old.Stat()
			if err != nil || info.Mode().Perm() != mode {
				t.Fatalf("publication changed old inode mode: %v, %v", info, err)
			}
			data := make([]byte, len(before))
			if _, err := old.ReadAt(data, 0); err != nil || !bytes.Equal(data, before) {
				t.Fatalf("publication changed old inode bytes: %q, %v", data, err)
			}
			plan, err := prepareClaudeSettings(root, ".")
			if err != nil || plan.payload != nil {
				t.Fatalf("published exclusion remains stale: %v", err)
			}
		})
	}
	t.Run("absent", func(t *testing.T) {
		root := t.TempDir()
		if err := publishClaudeSettings(root, "."); err != nil {
			t.Fatal(err)
		}
		assertManagedMode(t, filepath.Join(root, "settings.json"), 0o600)
	})
}

var claudeSettingsModeProbe func(string) error

// A real external mode change between staging and the final read triggers a
// retry. Three changes refuse without publishing content or leaving residue;
// zero/one-change controls retain the observed mode and publish the exclusion.
func TestClaudeNativeMemoryModeRecheck(t *testing.T) {
	if os.Getenv("CURATOR_PUBLICATION_PROBE") != "1" {
		output, code := claudePublicationOverlay(t, "claude_settings.go",
			"input, present, mode, readErr := readClaudeSettingsInput(root, homeRel)",
			"if claudeSettingsModeProbe != nil { if err := claudeSettingsModeProbe(full); err != nil { release(); _ = os.Remove(temp); return err } }; input, present, mode, readErr := readClaudeSettingsInput(root, homeRel)",
			"^TestClaudeNativeMemoryModeRecheck$", true)
		if code != 0 || !strings.Contains(output, "--- PASS: TestClaudeNativeMemoryModeRecheck/changes-3 ") {
			t.Fatalf("mode retry witness failed: exit %d", code)
		}
		return
	}
	for _, changes := range []int{0, 1, 3} {
		t.Run(fmt.Sprintf("changes-%d", changes), func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "settings.json")
			before := []byte(`{"operator":true}`)
			if err := os.WriteFile(path, before, 0o600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			wantMode := os.FileMode(0o600)
			claudeSettingsModeProbe = func(full string) error {
				calls++
				if calls <= changes {
					wantMode ^= 0o200
					return os.Chmod(full, wantMode)
				}
				return nil
			}
			defer func() { claudeSettingsModeProbe = nil }()
			err := publishClaudeSettings(root, ".")
			if changes == 3 {
				if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) || !strings.Contains(err.Error(), "all 3 publication checks") {
					t.Fatalf("mode changes did not exhaust the bounded publication: %v", err)
				}
				if !bytes.Equal(readClaudeState(t, path), before) {
					t.Fatal("mode-change refusal published content")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != wantMode {
				t.Fatalf("publication overwrote external mode: want %04o, got %v, %v", wantMode, info, err)
			}
			wantCalls := changes + 1
			if wantCalls > 3 {
				wantCalls = 3
			}
			if calls != wantCalls {
				t.Fatalf("mode check count %d, want %d", calls, wantCalls)
			}
			if changes != 3 {
				plan, err := prepareClaudeSettings(root, ".")
				if err != nil || plan.payload != nil {
					t.Fatalf("successful mode retry lacks exclusion: %v", err)
				}
			}
			if staged, err := claudeSettingsStagingEntries(root); err != nil || len(staged) != 0 {
				t.Fatalf("mode-change publication left staging files: %v, %v", staged, err)
			}
		})
	}
}

// Narrowing either mode preservation or its final recheck must reach the
// matching behavioral failure; the native/unchanged positive controls survive.
func TestClaudeNativeMemoryModeNarrowingMutants(t *testing.T) {
	const read = "input, present, mode, readErr := readClaudeSettingsInput(root, homeRel)"
	const probe = "if claudeSettingsModeProbe != nil { if err := claudeSettingsModeProbe(full); err != nil { release(); _ = os.Remove(temp); return err } }; "
	for _, mutant := range []struct {
		name, before, after, run, fail, pass, diagnostic string
		child                                            bool
	}{
		{"replace-mode", "file.Chmod(plan.mode)", "file.Chmod(managedAuthoredFileMode)", "^TestClaudeNativeMemoryReadOnlySettings$", "TestClaudeNativeMemoryReadOnlySettings/referenced/read-only", "TestClaudeNativeMemoryReadOnlySettings/referenced/native", "has write bits (0600)", false},
		{"omit-mode-recheck", read + "\n\t\tif readErr != nil || present != plan.present || mode != plan.mode || !bytes.Equal(input, plan.input) {", probe + read + "; _ = mode\n\t\tif readErr != nil || present != plan.present || !bytes.Equal(input, plan.input) {", "^TestClaudeNativeMemoryModeRecheck$", "TestClaudeNativeMemoryModeRecheck/changes-1", "TestClaudeNativeMemoryModeRecheck/changes-0", "publication overwrote external mode", true},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			output, code := claudePublicationOverlay(t, "claude_settings.go", mutant.before, mutant.after, mutant.run, mutant.child)
			if code != 1 || !strings.Contains(output, "--- FAIL: "+mutant.fail+" ") || !strings.Contains(output, "--- PASS: "+mutant.pass+" ") || !strings.Contains(output, mutant.diagnostic) {
				t.Fatalf("mode mutant did not reach its named failure/control: exit %d", code)
			}
			t.Logf("expected-red exit 1; %s: %s", mutant.name, mutant.diagnostic)
		})
	}
}
