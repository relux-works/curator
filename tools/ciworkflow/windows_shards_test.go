package ciworkflow

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func checkWindowsShards(data []byte) error {
	// runs-on has a different shape on other jobs; decode only our two jobs.
	var raw struct{ Jobs map[string]yaml.Node }
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, name := range []string{"test", "windows-test-shards"} {
		node := raw.Jobs[name]
		var job struct {
			If, Needs string
			RunsOn    any `yaml:"runs-on"`
			Strategy  struct {
				FailFast bool `yaml:"fail-fast"`
				Matrix   map[string]any
			}
			Steps []step
		}
		if err := node.Decode(&job); err != nil {
			return err
		}
		if name == "test" {
			if job.If != "${{ !cancelled() }}" || job.Needs != "windows-test-shards" {
				return fmt.Errorf("Test must run after Windows shard failure unless cancelled")
			}
			download, consume, ordinary := 0, 0, 0
			for _, s := range job.Steps {
				if s.Uses == "actions/download-artifact@v4" && s.If == "runner.os == 'Windows'" && s.With["pattern"] == "windows-test-shard-*" && s.With["path"] == ".temp/windows-test-shards" && s.With["merge-multiple"] != "true" {
					download++
				}
				if strings.Contains(s.Run, "test-gate.sh") {
					if s.If == "runner.os == 'Windows'" && s.Env["CI_WINDOWS_TEST_SHARDS"] == ".temp/windows-test-shards" && s.Env["GO_TEST_TIMEOUT"] == "120m" && s.Env["GO_TEST_FLAGS"] == "" && download == 1 && strings.Contains(s.Run, "-u GIT_CONFIG_COUNT -u GIT_CONFIG_PARAMETERS") {
						consume++
					}
					if s.If == "runner.os != 'Windows'" && s.Env["CI_WINDOWS_TEST_SHARDS"] == "" {
						ordinary++
					}
				}
			}
			if download != 1 || consume != 1 || ordinary != 1 {
				return fmt.Errorf("Windows download/admission/ordinary split: %d/%d/%d", download, consume, ordinary)
			}
			continue
		}
		if job.If != "" || job.RunsOn != "windows-latest" || job.Strategy.FailFast || len(job.Strategy.Matrix) != 1 || !reflect.DeepEqual(job.Strategy.Matrix["shard"], []any{0, 1, 2, 3}) {
			return fmt.Errorf("require four unconditional Windows jobs, without an OS/package cross product")
		}
		run, upload, pin, nodeSetup, pnpm, rust := 0, 0, 0, 0, 0, 0
		for _, s := range job.Steps {
			if strings.Contains(s.Run, "go run ./tools/envprofilerace -lane windows-test -shard ${{ matrix.shard }}") && strings.Contains(s.Run, "-u GIT_CONFIG_COUNT -u GIT_CONFIG_PARAMETERS") && strings.Contains(s.Run, "HOME=\"$(mktemp -d)\"") && s.If == "" && s.Env["CURATOR_CONFORMANCE_ROOT"] == "${{ github.workspace }}/protocol-spec/conformance/v1" {
				run++
			}
			if s.Uses == "actions/upload-artifact@v4" && s.If == "always()" && s.With["name"] == "windows-test-shard-${{ matrix.shard }}" && s.With["path"] == ".temp/ci-evidence/windows-shard/" && s.With["if-no-files-found"] == "error" {
				upload++
			}
			if s.With["repository"] == "relux-works/curator-spec" && s.With["ref"] == "${{ env.SPEC_PIN }}" {
				pin++
			}
			if s.Uses == "actions/setup-node@v4" && s.With["node-version"] == "22" {
				nodeSetup++
			}
			if strings.Contains(s.Run, "npm install -g") && strings.Contains(s.Run, "pnpm@${{ env.PNPM_PIN }}") {
				pnpm++
			}
			if s.Run == "bash .github/ci/install-rust-toolchain.sh" {
				rust++
			}
		}
		if run != 1 || upload != 1 || pin != 1 || nodeSetup != 1 || pnpm != 1 || rust != 1 {
			return fmt.Errorf("Windows run/upload/pin/tool provisioning: %d/%d/%d/%d/%d/%d", run, upload, pin, nodeSetup, pnpm, rust)
		}
	}
	return nil
}

// Real workflow wiring must preserve all four shards on main/PR, required
// failure admission, native prerequisites and full Windows platform-gate input.
func TestWindowsShardsWiring(t *testing.T) {
	good := workflowBytes(t)
	if err := checkWindowsShards(good); err != nil {
		t.Fatal(err)
	}
	for name, pair := range map[string][2]string{
		"lost-shard":       {"shard: [0, 1, 2, 3]", "shard: [0, 1, 2]"},
		"success-only":     {"  test:\n    if: ${{ !cancelled() }}", "  test:\n    if: success()"},
		"wrong-dependency": {"needs: windows-test-shards", "needs: envprofile-race"},
		"lost-admission":   {"CI_WINDOWS_TEST_SHARDS: .temp/windows-test-shards", "CI_OTHER: .temp/windows-test-shards"},
		"lost-upload":      {"name: windows-test-shard-${{ matrix.shard }}", "name: windows-test-shard-0"},
		"lost-selector":    {"-lane windows-test -shard ${{ matrix.shard }}", "-lane windows-test -shard 0"},
		"only-main":        {"  windows-test-shards:\n", "  windows-test-shards:\n    if: github.event_name == 'push'\n"},
	} {
		t.Run(name, func(t *testing.T) {
			changed := strings.Replace(string(good), pair[0], pair[1], 1)
			if changed == string(good) {
				t.Fatal("plant not applied")
			}
			if err := checkWindowsShards([]byte(changed)); err == nil {
				t.Fatal("Windows wiring violation admitted")
			}
		})
	}
}

// Windows checkout line endings must leave every positive and negative shard
// wiring control executable, including the existing Race controls.
func TestShardControlsWithCRLFCheckout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ci.yml")
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(workflowBytes(t)), "\n", "\r\n")), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CI_WORKFLOW_UNDER_TEST", path)
	t.Run("race", TestEnvprofileRaceWiring)
	t.Run("windows", TestWindowsShardsWiring)
}
