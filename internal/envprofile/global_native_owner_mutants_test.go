package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each overlay removes one ownership condition for a bounded witness. A kill
// requires the named assertion and exit 1; setup, build failures, crashes and
// missing tests do not count. Children execute one package at a time.
func TestGlobalNativeOwnershipNarrowingControls(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, plant := range []struct{ name, file, before, after, pkg, run, failure, message string }{
		{"omit-direct-declarations", "internal/envprofile/global.go", "if global != nil {", "if global != nil && len(global.Skills) == 0 {", "./cmd/curator", "^TestGlobalNativeOwnerNewDeclarationAfterProfile$", "TestGlobalNativeOwnerNewDeclarationAfterProfile", "profile native owner lost"},
		{"retain-old-direct-pin", "internal/envprofile/global.go", "if declared[member.Name] || (addition != nil && member.Name == addition.Name) {", "if addition != nil && member.Name == addition.Name {", "./cmd/curator", "^TestGlobalNativeOwnerNewDeclarationAfterProfile$", "TestGlobalNativeOwnerNewDeclarationAfterProfile", "want 0"},
		{"second-publisher", "internal/adapters/stage.go", "if preserved[name] {", "if preserved[name] && name != \"shared-skill\" {", "./cmd/curator", "^TestGlobalNativeOwnerCleanBootstrapInstall$", "TestGlobalNativeOwnerCleanBootstrapInstall", "want 0"},
		{"double-claim", "internal/adapters/stage.go", "if expected[name] && !preserved[name] {", "if expected[name] {", "./internal/adapters", "^TestGlobalProfileOwnerPreservesEntries$/^selected$", "TestGlobalProfileOwnerPreservesEntries/selected", "profile entry claimed by legacy ledger"},
		{"retired-owner-removal", "internal/adapters/stage.go", "if !expected[name] && !preserved[name] {", "if !expected[name] {", "./internal/adapters", "^TestGlobalProfileOwnerPreservesEntries$/^retired$", "TestGlobalProfileOwnerPreservesEntries/retired", "profile entry was replaced or removed"},
		{"ownership-read-error", "internal/adapters/stage.go", "if err != nil {\n\t\t\t\treturn err\n\t\t\t}\n\t\t\tpreserved[name]", "if err != nil && err.Error() != \"ownership state unreadable\" {\n\t\t\t\treturn err\n\t\t\t}\n\t\t\tpreserved[name]", "./internal/adapters", "^TestGlobalProfileOwnerPreservesEntries$/^selected$", "TestGlobalProfileOwnerPreservesEntries/selected", "ownership error was ignored"},
		{"source-disagreement", "internal/envprofile/native_skill_owner.go", "member.Source != requested.Source", "(member.Source != requested.Source && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^source$", "TestNativeSkillOwnerProof/source", "invalid native ownership admitted"},
		{"pin-disagreement", "internal/envprofile/native_skill_owner.go", "member.Commit != requested.Commit", "(member.Commit != requested.Commit && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^commit$", "TestNativeSkillOwnerProof/commit", "invalid native ownership admitted"},
		{"directory-disagreement", "internal/envprofile/native_skill_owner.go", "member.Directory != requested.Directory", "(member.Directory != requested.Directory && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^directory$", "TestNativeSkillOwnerProof/directory", "invalid native ownership admitted"},
		{"marker-lock", "internal/envprofile/native_skill_owner.go", "if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, \"sha256:\") {", "if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, \"sha256:\") && name != \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^lock$", "TestNativeSkillOwnerProof/lock", "invalid native ownership admitted"},
		{"marker-content", "internal/envprofile/native_skill_owner.go", "digest != surface.ContentSHA256", "(digest != surface.ContentSHA256 && name != \"myskill\")", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^surface$", "TestNativeSkillOwnerProof/surface", "invalid native ownership admitted"},
		{"marker-entry", "internal/envprofile/native_skill_owner.go", "if !recorded {", "if !recorded && name != \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^unrecorded$", "TestNativeSkillOwnerProof/unrecorded", "invalid native ownership admitted"},
		{"live-link", "internal/envprofile/native_skill_owner.go", "if link == selected.root {", "if link == selected.root || name == \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerProof$/^live-link$", "TestNativeSkillOwnerProof/live-link", "invalid native ownership admitted"},
		{"copy-record", "internal/envprofile/native_skill_owner.go", "if copied {", "if copied || name == \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerCopyProof$/^fallback$", "TestNativeSkillOwnerCopyProof/fallback", "unrecorded copied entry admitted"},
		{"copy-content", "internal/envprofile/native_skill_owner.go", "if hash == selected.hash {", "if hash == selected.hash || name == \"myskill\" {", "./internal/envprofile", "^TestNativeSkillOwnerCopyProof$/^fallback$", "TestNativeSkillOwnerCopyProof/fallback", "drifted copied entry admitted"},
		{"ignore-ledgers", "internal/adapters/adapters.go", "func unmanagedConflict(target string, inLedger bool, source string) (bool, error) {", "func unmanagedConflict(target string, inLedger bool, source string) (bool, error) {\nif filepath.Base(target) == \"shared-skill\" { return false, nil }", "./internal/adapters", "^TestGlobalProfileOwnerForeignRefuses$", "TestGlobalProfileOwnerForeignRefuses", "foreign adapter entry admitted"},
	} {
		t.Run(plant.name, func(t *testing.T) {
			original := filepath.Join(root, filepath.FromSlash(plant.file))
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if count := strings.Count(source, plant.before); count != 1 {
				t.Fatalf("plant anchor count = %d, want 1", count)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, plant.before, plant.after, 1)), 0600); err != nil {
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
			command := exec.Command("env", "-u", "GIT_DIR", "-u", "GIT_WORK_TREE", "go", "test", "-p=2", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", plant.pkg, "-run", plant.run)
			command.Dir = root
			if plant.pkg == "./cmd/curator" {
				command.Env = append(os.Environ(), "CURATOR_TEST_CLI_HELPER=1")
			}
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.failure+" ") || !strings.Contains(string(output), plant.message) || strings.Contains(string(output), "panic:") {
				t.Fatalf("plant survived or lacked intended assertion: %v\n%s", err, output)
			}
			t.Logf("expected-red exit 1; killed by %s: %s", plant.failure, plant.message)
		})
	}
}
