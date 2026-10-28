package envprofile

import (
	"bytes"
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
	isolateDetectorParent(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, file, before, after, run, fail, message string }{
		{"native-raw-root", "switch.go", "skillLinks[path] = skill.root", "skillLinks[path] = skill.source", "^TestProfileSkillProjectionPublication$/^commands-false$", "TestProfileSkillProjectionPublication/commands-false/", "skill publication differs from projection"},
		{"managed-raw-root", "managed.go", "p.links[path] = skill.root", "p.links[path] = skill.source", "^TestProfileSkillProjectionPublication$/^commands-false$", "TestProfileSkillProjectionPublication/commands-false/", "skill publication differs from projection"},
		{"exported-scripts", "skill_projection.go", "len(spec.Commands) == 0", "true", "^TestProfileSkillProjectionPublication$/^commands-true$", "TestProfileSkillProjectionPublication/commands-true/", "skill publication differs from projection"},
		{"omit-runtime-build-exclusions", "skill_projection.go", "whitelist.ContextExcludedRoots(spec.RuntimeRoots, spec.BuildRoots)", "nil", "^TestProfileSkillProjectionPublication$/^commands-true$", "TestProfileSkillProjectionPublication/commands-true/", "skill publication differs from projection"},
		{"legacy-plan-framing", "skill_projection.go", "if err := opaquescan.RefuseNULV1(source, version); err != nil {", "if err := opaquescan.RefuseNULV1(source, version); false && err != nil {", "^TestSkillProjectionLegacyFraming$/^plan$", "TestSkillProjectionLegacyFraming/plan", "legacy projection hashed or admitted NUL"},
		{"legacy-tree-framing", "skill_projection.go", "if err := opaquescan.RefuseNULV1(root, version); err != nil {", "if err := opaquescan.RefuseNULV1(root, version); false && err != nil {", "^TestSkillProjectionLegacyFraming$/^(stage|verify)$", "TestSkillProjectionLegacyFraming/", "legacy projection hashed or admitted NUL"},
		// Both link and copy are supported today, so merely choosing a link is
		// no violation. Reuse the stale owned tree instead of the filtered
		// store projection at the actual copy publisher to plant the defect.
		{"old-tree-copy", "home_publication.go", "pub.tree(filepath.Join(plan.homeDir, filepath.FromSlash(path)), target)", "pub.tree(filepath.Join(plan.homeDir, filepath.FromSlash(path)), filepath.Join(plan.homeDir, filepath.FromSlash(path)))", "^TestProfileSkillProjectionRepair$/./^raw-copy$", "TestProfileSkillProjectionRepair/", "repair failed to publish filtered projection: environment_repair_failed: surface file skills/myskill is drifted: content hash differs"},
		{"old-surface-identity", "skill_projection.go", `skillProjectionVersion + ":" + skill.hash`, `skill.hash`, "^TestProfileSkillProjectionRepair$/./^old-marker$", "TestProfileSkillProjectionRepair/", "old native surface current"},
		{"missing-managed-projection", "managed.go", "if err := verifySkillProjection(req.Home, skill); err != nil {", "if err := verifySkillProjection(req.Home, skill); false && err != nil {", "^TestProfileSkillProjectionMissingStore$", "TestProfileSkillProjectionMissingStore", "missing projection current"},
		{"missing-native-projection", "native_skills.go", "if err := verifySkillProjection(home, skill); err != nil {", "if err := verifySkillProjection(home, skill); false && err != nil {", "^TestProfileSkillProjectionMissingStore$", "TestProfileSkillProjectionMissingStore", "missing native projection current"},
		{"omit-projection-rebuild", "skill_projection.go", "if err := verifySkillProjection(home, skill); !errors.Is(err, ErrSkillProjectionIdentity) {\n\t\t\treturn err\n\t\t}", "return verifySkillProjection(home, skill)", "^TestProfileSkillProjectionDamagedStore$/./^(missing-file|extra-file|changed-file)$", "TestProfileSkillProjectionDamagedStore/", "projection repair did not converge"},
		{"rebuild-source-identity", "skill_projection.go", `return fmt.Errorf("%w: skill context changed while projecting", ErrSkillProjectionIdentity)`, `return nil`, "^TestSkillProjectionRebuildSourceRefusal$", "TestSkillProjectionRebuildSourceRefusal", "wrong staged identity refusal"},
		{"changed-projection-bytes", "skill_projection.go", `return ErrSkillProjectionIdentity`, `return nil`, "^TestSkillProjectionStoreReuse$/^bytes$", "TestSkillProjectionStoreReuse/bytes", "altered projection reused"},
		{"extra-projection-link", "skill_projection.go", `if err := pathboundary.ValidateWithin(contextstore.Root(home), skill.root); err != nil {`, `if err := pathboundary.ValidateWithin(contextstore.Root(home), skill.root); false && err != nil {`, "^TestSkillProjectionStoreReuse$/^link$", "TestSkillProjectionStoreReuse/link", "altered projection reused"},
		{"changed-source", "skill_projection.go", `return fmt.Errorf("%w: skill context changed while projecting", ErrSkillProjectionIdentity)`, `return nil`, "^TestSkillProjectionStoreReuse$/^source-changed$", "TestSkillProjectionStoreReuse/source-changed", "changed source published"},
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
			command.Env = detectorChildEnv()
			var buffer bytes.Buffer
			command.Stdout = &buffer
			command.Stderr = &buffer
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			pid := command.Process.Pid
			err = command.Wait()
			t.Logf("detector child PID %d reaped; exit %d", pid, command.ProcessState.ExitCode())
			output := buffer.Bytes()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+tc.fail) || !strings.Contains(string(output), tc.message) {
				t.Fatalf("mutant survived or lacked intended failure: %v\n%s", err, output)
			}
			if tc.name == "omit-projection-rebuild" && strings.Count(string(output), "--- FAIL: TestProfileSkillProjectionDamagedStore/") != 6 {
				t.Fatalf("rebuild mutant did not fail all six cases:\n%s", output)
			}
			t.Logf("expected-red exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
