package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each plant narrows one behavior to a named exception while retaining the
// neighboring gates. A kill requires exit 1 and the named behavioral failure;
// compilation, setup errors, and missing tests are harness failures.
func TestNativeSkillsNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct{ name, file, before, after, pkg, run, fails, message string }{
		{"pi-native-only", "native_skills.go", "return adapter.SkillsDir", "if id == envregistry.Pi { return profileSkillsSurfaceDir }; return adapter.SkillsDir", "envprofile", "TestNativeSkillsUseAndSync", "TestNativeSkillsUseAndSync", "native skill link missing"},
		{"codex-link-drift", "native_skills.go", "link.Target != skill.root", "(link.Target != skill.root && id != envregistry.CodexCLI)", "envprofile", "TestNativeSkillsRepair", "TestNativeSkillsRepair", "drift status"},
		{"dangling-foreign-consent", "switch.go", "if !takeover {\n\t\t\t\t\tresult.OK", "if !takeover && !strings.HasPrefix(filepath.ToSlash(link.Target), \"../../retired/\") {\n\t\t\t\t\tresult.OK", "envprofile", "TestNativeSkillsTakeover/foreign", "TestNativeSkillsTakeover/foreign", "environment_foreign_manager_detected"},
		{"plain-file-consent", "switch.go", "if !takeover {\n\t\t\tresult.OK", "if !takeover && !info.Mode().IsRegular() {\n\t\t\tresult.OK", "envprofile", "TestNativeSkillsTakeover/file", "TestNativeSkillsTakeover/file", "environment_surface_unmanaged_conflict"},
		{"skills-parent-traversal", "nofollow.go", "if info.Mode()&os.ModeSymlink != 0 {", "if info.Mode()&os.ModeSymlink != 0 { if component == \"skills\" { continue };", "envprofile", "TestNativeSkillsLinkedParent", "TestNativeSkillsLinkedParent", "environment_write_would_follow_link"},
		{"unnamed-takeover-directory", "switch.go", "// Remove recorded files the new profile no longer wants.", "if takeover { if info, err := os.Lstat(filepath.Join(native, \"skills\", \"unnamed\")); err == nil && info.IsDir() { recorded[\"skills/unnamed\"] = true } }\n// Remove recorded files the new profile no longer wants.", "envprofile", "TestNativeSkillsConvergenceAndSwitch", "TestNativeSkillsConvergenceAndSwitch", "unnamed skill changed"},
		{"one-ledger-name", "native_skills.go", "return dir != profileSkillsSurfaceDir && ledger[", "return path != \"skills/myskill\" && dir != profileSkillsSurfaceDir && ledger[", "envprofile", "TestNativeSkillsLedger", "TestNativeSkillsLedger", "ledger entry not reported"},
		{"one-old-staging-name", "switch.go", "if !want[path] {", "if !want[path] && path != \".agent-context/skills/myskill\" {", "envprofile", "TestNativeSkillsConvergenceAndSwitch", "TestNativeSkillsConvergenceAndSwitch", "old staging entry retained"},
		{"one-retired-name", "switch.go", "if !want[path] {", "if !want[path] && path != \"skills/myskill\" {", "envprofile", "TestNativeSkillsConvergenceAndSwitch", "TestNativeSkillsConvergenceAndSwitch", "retired recorded skill retained"},
		{"backup-directory-collision", "switch.go", "} else if generationState.Kind == stateread.KindPresent {", "} else if generationState.Kind == stateread.KindPresent && !generationState.Info.IsDir() {", "envprofile", "TestNativeSkillsBackupCollision", "TestNativeSkillsBackupCollision", "environment_backup_exists"},
		{"directory-marker-is-absence", "switch.go", "prior, err := envmarker.Read(native)\n\tif err != nil {\n\t\treturn EntryResult{Adapter: adapter.ID, Home: native, OK: false, Detail: err.Error()}\n\t}", "prior, err := envmarker.Read(native)\n\tif err != nil {\n if info, e := os.Lstat(filepath.Join(native,envmarker.Name)); e != nil || !info.IsDir() { return EntryResult{Adapter:adapter.ID,Home:native,Detail:err.Error()} }\n }", "envprofile", "TestNativeSkillsUnreadableMarker", "TestNativeSkillsUnreadableMarker", "environment_marker_unreadable"},
		{"claude-takeover-without-repair", "../../cmd/curator/command_help.go", "case \"env resolve\":\n\t\t\treturn !repair", "case \"env resolve\":\n\t\t\treturn !repair && (len(args) < 3 || args[2] != \"claude_code\")", "curator", "TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code", "TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code", "takeover admitted or unrelated failure"},
	} {
		t.Run(m.name, func(t *testing.T) {
			original := filepath.Join(root, "internal", "envprofile", m.file)
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			// nofollow has the same guard in both walkers; weaken only managedPath.
			count := strings.Count(source, m.before)
			if count != 1 && (m.name != "skills-parent-traversal" || count != 4) {
				t.Fatalf("mutation anchor count = %d", count)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, m.before, m.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			pkg := "./internal/envprofile"
			if m.pkg == "curator" {
				pkg = "./cmd/curator"
			}
			run := "^" + strings.ReplaceAll(m.run, "/", "$/^") + "$"
			cmd := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", run, pkg)
			cmd.Dir = root
			if m.pkg == "curator" {
				// These flag-only cases do not fingerprint GOROOT. Use the CLI's
				// existing helper lane so a full CLI package run cannot block them.
				cmd.Env = append(os.Environ(), "CURATOR_TEST_CLI_HELPER=1")
			}
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+m.fails+" ") || !strings.Contains(string(output), m.message) {
				t.Fatalf("mutant survived or lacked intended behavioral failure: %v\n%s", err, output)
			}
			t.Logf("expected-red exit %d; killed by %s:\n%s", exit.ExitCode(), m.fails, output)
		})
	}
}
