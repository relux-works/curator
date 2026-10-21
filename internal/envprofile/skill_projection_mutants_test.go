package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each bounded overlay restores one forbidden publication or weakens one new
// store/currentness gate. Only exit 1 and the intended behavioral assertion
// count; unrelated setup failures and crashes cannot satisfy these controls.
func TestSkillProjectionNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, file, before, after, run, fail, message string }{
		{"native-raw-root", "switch.go", "skillLinks[path] = skill.root", "skillLinks[path] = skill.source", "^TestProfileSkillProjectionPublication$/^commands-false$", "TestProfileSkillProjectionPublication/commands-false/", "skill publication differs from projection"},
		{"managed-raw-root", "managed.go", "p.links[path] = skill.root", "p.links[path] = skill.source", "^TestProfileSkillProjectionPublication$/^commands-false$", "TestProfileSkillProjectionPublication/commands-false/", "skill publication differs from projection"},
		{"exported-scripts", "skill_projection.go", "len(spec.Commands) == 0", "true", "^TestProfileSkillProjectionPublication$/^commands-true$", "TestProfileSkillProjectionPublication/commands-true/", "skill publication differs from projection"},
		{"omit-runtime-build-exclusions", "skill_projection.go", "whitelist.ContextExcludedRoots(spec.RuntimeRoots, spec.BuildRoots)", "nil", "^TestProfileSkillProjectionPublication$/^commands-true$", "TestProfileSkillProjectionPublication/commands-true/", "skill publication differs from projection"},
		{"old-surface-identity", "skill_projection.go", `skillProjectionVersion + ":" + skill.hash`, `skill.hash`, "^TestProfileSkillProjectionRepair$/./^raw-link-false$", "TestProfileSkillProjectionRepair/", "old native surface current"},
		{"missing-managed-projection", "managed.go", "if err := verifySkillProjection(req.Home, skill); err != nil {", "if err := verifySkillProjection(req.Home, skill); false && err != nil {", "^TestProfileSkillProjectionMissingStore$", "TestProfileSkillProjectionMissingStore", "missing projection current"},
		{"missing-native-projection", "native_skills.go", "if err := verifySkillProjection(home, skill); err != nil {", "if err := verifySkillProjection(home, skill); false && err != nil {", "^TestProfileSkillProjectionMissingStore$", "TestProfileSkillProjectionMissingStore", "missing native projection current"},
		{"changed-projection-bytes", "skill_projection.go", `return fmt.Errorf("skill context projection differs from its identity")`, `return nil`, "^TestSkillProjectionStoreReuse$/^bytes$", "TestSkillProjectionStoreReuse/bytes", "altered projection reused"},
		{"extra-projection-link", "skill_projection.go", `if err := pathboundary.ValidateWithin(contextstore.Root(home), skill.root); err != nil {`, `if err := pathboundary.ValidateWithin(contextstore.Root(home), skill.root); false && err != nil {`, "^TestSkillProjectionStoreReuse$/^link$", "TestSkillProjectionStoreReuse/link", "altered projection reused"},
		{"changed-source", "skill_projection.go", `return fmt.Errorf("skill context changed while projecting")`, `return nil`, "^TestSkillProjectionStoreReuse$/^source-changed$", "TestSkillProjectionStoreReuse/source-changed", "changed source published"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, "internal/envprofile", tc.file)
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(body), "\r\n", "\n")
			if strings.Count(source, tc.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, tc.before, tc.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("env", "-u", "GIT_DIR", "-u", "GIT_WORK_TREE", "go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "./internal/envprofile", "-run", tc.run)
			command.Dir = root
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+tc.fail) || !strings.Contains(string(output), tc.message) {
				t.Fatalf("mutant survived or lacked intended failure: %v\n%s", err, output)
			}
			t.Logf("expected-red exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
