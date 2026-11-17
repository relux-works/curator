package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Every plant below must turn its named behavioral test red while the
// unchanged positive control passes first. A crash, setup error, missing
// anchor or unapplied plant never counts as a killed mutant.
func TestHomeLeaseBatchNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool := testtoolchain.Go(t)
	cases := []struct{ name, before, after, run, message string }{
		// Dropping the batch repin compensation must leave the repinned
		// records and re-rendered scopes behind after a late refusal.
		{"drop-batch-repin-compensation",
			"\t\t\top.undo = &globalUndo{entries: map[string]*globalImage{}, managerHome: home}",
			"\t\t\top.undo = nil",
			"TestRepinBatchTakeoverRefusalPreservesState", "late batch refusal changed mutable profile state"},
		// Dropping the per-name pre-lock probe must let the batch reach a
		// per-profile visit instead of refusing the v5 plan at entry.
		{"drop-batch-prepared-plan-probe",
			"\t\tif err := refusePreparedSkillPlanAdoption(home, name); err != nil {",
			"\t\tif err := error(nil); err != nil {",
			"TestPreparedPlanProbeOnBatchUpdateAndSync", "batch update probe:"},
		// Dropping the sync entry probe must let sync proceed past its
		// entry -- into recovery and ensureDefault, whose
		// migrateGlobalSkills fetch of the fixture's unreachable global
		// skill fails first, or which creates the default profile --
		// instead of refusing the v5 plan at the probe. Either way the
		// test fails under its "sync" stage prefix ("sync probe:" for the
		// wrong error, "sync: pre-recovery probe did not run first" for a
		// created default), so the plant keys on that stage prefix; the
		// ordering itself is asserted by the test's defaultAbsent check.
		{"drop-sync-prepared-plan-probe",
			"\tif err := refusePreparedSkillPlanAdoption(home, \"\"); err != nil {\n\t\treturn nil, nil, err, nil",
			"\tif err := error(nil); err != nil {\n\t\treturn nil, nil, err, nil",
			"TestPreparedPlanProbeOnBatchUpdateAndSync", "sync"},
	}
	controls := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := func(overlay string) ([]byte, error) {
				args := []string{"test", "-count=1", "-timeout=10m", "-v", "./internal/envprofile", "-run", "^" + tc.run + "$"}
				if overlay != "" {
					args = append([]string{"test", "-overlay", overlay}, args[1:]...)
				}
				cmd := exec.Command(goTool, args...)
				cmd.Dir = root
				for _, kv := range os.Environ() {
					if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") {
						cmd.Env = append(cmd.Env, kv)
					}
				}
				testtoolchain.InheritHostGOROOTLock(t, cmd)
				return cmd.CombinedOutput()
			}
			if !controls[tc.run] {
				out, err := run("")
				if err != nil || !strings.Contains(string(out), "--- PASS: "+tc.run) {
					t.Fatalf("positive control failed: %v\n%s", err, out)
				}
				controls[tc.run] = true
			}
			path := filepath.Join(root, "internal/envprofile/homelease.go")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(raw), "\r\n", "\n")
			if strings.Count(source, tc.before) != 1 {
				t.Fatal("mutation anchor not unique")
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, tc.before, tc.after, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			raw, err = json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlay, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			out, err := run(overlay)
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: "+tc.run) ||
				!strings.Contains(string(out), tc.message) || strings.Contains(string(out), "panic:") {
				t.Fatalf("plant survived or failed for another reason: %v\n%s", err, out)
			}
			t.Logf("intended expected-red exit 1:\n%s", out)
		})
	}
}
