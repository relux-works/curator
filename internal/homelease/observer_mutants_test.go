//go:build linux || darwin

package homelease

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
func TestObserverPlants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool := testtoolchain.Go(t)
	cases := []struct{ name, file, before, after, pkg, run, message string }{
		// Skipping the birth-identity comparison must let a reused pid bind.
		{"skip-start-time-comparison", "internal/homelease/state/process.go",
			"\treturn *a == *b", "\treturn true",
			"./internal/homelease", "TestProductionBindObserverRows", "wanted environment_lease_holder_mismatch/process_incarnation_mismatch"},
		// Trusting the manager's assertion instead of reading the kernel
		// must let a tuple the kernel disagrees with bind.
		{"trust-manager-assertion", "internal/homelease/observer.go",
			"observed, err := kernel(requested.Process.PID)", "observed, err := *requested.Process, error(nil)",
			"./internal/homelease", "TestProductionBindObserverRows", "wanted environment_lease_holder_mismatch/process_incarnation_mismatch"},
		// Returning a historical bind outcome without fresh observation
		// must let a stale retry succeed against a changed incarnation.
		{"retry-skips-observer", "internal/homelease/lifecycle.go",
			"if err := m.verifyBind(ctx, e.Initial.Holder, holder); err != nil {", "if err := error(nil); err != nil {",
			"./internal/homelease", "TestProductionBindRetryReverifies", "wanted environment_lease_holder_mismatch/process_incarnation_mismatch"},
		// Dropping the uid ceiling must let the (uid_t)-1 sentinel through.
		{"drop-uid-ceiling", "internal/homelease/state/process.go",
			"if p.UID < 0 || p.UID > 4294967294 {", "if p.UID < 0 {",
			"./internal/homelease/state", "TestProcessSemanticCeilings", "value above its ceiling accepted"},
	}
	controls := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := func(overlay string) ([]byte, error) {
				args := []string{"test", "-count=1", "-timeout=5m", "-v", tc.pkg, "-run", "^" + tc.run + "$"}
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
			if !controls[tc.pkg+"/"+tc.run] {
				out, err := run("")
				if err != nil || !strings.Contains(string(out), "--- PASS: "+tc.run) {
					t.Fatalf("positive control failed: %v\n%s", err, out)
				}
				controls[tc.pkg+"/"+tc.run] = true
			}
			path := filepath.Join(root, tc.file)
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
