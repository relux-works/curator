package ciworkflow

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func checkAbsoluteTestHome(run string) bool {
	return strings.HasPrefix(run, "bash .github/ci/isolated-test.sh ") && !strings.Contains(run, "\n")
}

func checkUnixShards(data []byte) error {
	type job struct {
		If       string
		Needs    []string
		RunsOn   string `yaml:"runs-on"`
		Strategy struct {
			FailFast bool `yaml:"fail-fast"`
			Matrix   map[string]any
		}
		Steps []step
	}
	// Other jobs have heterogeneous needs/runs-on shapes; decode only these two.
	var raw struct{ Jobs map[string]yaml.Node }
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}
	jobs := make(map[string]job)
	for _, name := range []string{"test", "unix-test-shards"} {
		job := jobs[name]
		node := raw.Jobs[name]
		if err := node.Decode(&job); err != nil {
			return err
		}
		jobs[name] = job
	}
	producer := jobs["unix-test-shards"]
	if producer.If != "" || producer.Strategy.FailFast || len(producer.Strategy.Matrix) != 3 ||
		!reflect.DeepEqual(producer.Strategy.Matrix["os"], []any{"ubuntu-latest", "macos-latest"}) ||
		!reflect.DeepEqual(producer.Strategy.Matrix["shard"], []any{0, 1, 2}) ||
		producer.Strategy.Matrix["exclude"] != unixMacOSExclude {
		return fmt.Errorf("require three Linux shards and two macOS shards selected by landing policy")
	}
	if producer.RunsOn != jobs["test"].RunsOn {
		return fmt.Errorf("native shard runner routing differs from Test")
	}
	run, upload, pin, node, pnpm, rust, identity := 0, 0, 0, 0, 0, 0, 0
	for _, s := range producer.Steps {
		if strings.Contains(s.Run, "go run ./tools/envprofilerace -lane unix-test -shard ${{ matrix.shard }} -dir .temp/ci-evidence/unix-shard") && s.If == "" && checkAbsoluteTestHome(s.Run) && s.Env["CURATOR_CONFORMANCE_ROOT"] == "${{ github.workspace }}/protocol-spec/conformance/v1" {
			run++
		}
		if s.Uses == "actions/upload-artifact@v4" && s.If == "always()" && s.With["name"] == "unix-test-shard-${{ matrix.os }}-${{ matrix.shard }}" && s.With["path"] == ".temp/ci-evidence/unix-shard/" && s.With["if-no-files-found"] == "error" {
			upload++
		}
		if s.With["repository"] == "relux-works/curator-spec" && s.With["ref"] == "${{ env.SPEC_PIN }}" {
			pin++
		}
		if s.Uses == "actions/setup-node@v4" && s.With["node-version"] == "22" {
			node++
		}
		if strings.Contains(s.Run, "npm install -g") && strings.Contains(s.Run, "pnpm@${{ env.PNPM_PIN }}") {
			pnpm++
		}
		if s.Run == "bash .github/ci/install-rust-toolchain.sh" {
			rust++
		}
		if s.Run == "bash .github/ci/toolchain-identity.sh" {
			identity++
		}
	}
	if run != 1 || upload != 1 || pin != 1 || node != 1 || pnpm != 1 || rust != 1 || identity != 1 {
		return fmt.Errorf("Unix run/upload/pin/tool provisioning incomplete")
	}
	consumer := jobs["test"]
	if consumer.If != "${{ !cancelled() }}" || !reflect.DeepEqual(consumer.Needs, []string{"windows-test-shards", "unix-test-shards"}) {
		return fmt.Errorf("Test must admit even after a shard dependency failure")
	}
	download, consume := 0, 0
	for _, s := range consumer.Steps {
		if s.Uses == "actions/download-artifact@v4" && s.If == "runner.os != 'Windows'" && s.With["pattern"] == "unix-test-shard-${{ matrix.os }}-*" && s.With["path"] == ".temp/unix-test-shards" && s.With["merge-multiple"] != "true" {
			download++
		}
		if strings.Contains(s.Run, "bash .github/ci/test-gate.sh .temp/ci-evidence/test") && s.If == "runner.os != 'Windows'" && s.Env["CI_UNIX_TEST_SHARDS"] == ".temp/unix-test-shards" && s.Env["GO_TEST_TIMEOUT"] == "${{ runner.os == 'Windows' && '120m' || '60m' }}" && s.Env["GO_TEST_FLAGS"] == "" && download == 1 && checkAbsoluteTestHome(s.Run) {
			consume++
		}
	}
	if download != 1 || consume != 1 {
		return fmt.Errorf("Unix download/admission/full platform gate missing")
	}
	return nil
}

// The real workflow preserves Linux shards on every run and macOS on landing candidates, original deadlines,
// pinned prerequisites, absolute isolated HOME and full final gate. Each bounded
// plant must trigger rejection; fixture controls do not measure native runtime.
func TestUnixShardsWiring(t *testing.T) {
	good := workflowBytes(t)
	if err := checkUnixShards(good); err != nil {
		t.Fatal(err)
	}
	for name, pair := range map[string][2]string{
		"lost-linux-shard":         {"shard: [0, 1, 2]", "shard: [0, 1]"},
		"lost-macos-shard":         {"os: [ubuntu-latest, macos-latest]", "os: [ubuntu-latest]"},
		"package-cross-product":    {"shard: [0, 1, 2]", "shard: [0, 1, 2]\n        package: [envprofile, install]"},
		"only-main":                {"  unix-test-shards:\n", "  unix-test-shards:\n    if: github.event_name == 'push'\n"},
		"lost-admission":           {"CI_UNIX_TEST_SHARDS: .temp/unix-test-shards", "CI_OTHER: .temp/unix-test-shards"},
		"lost-upload":              {"name: unix-test-shard-${{ matrix.os }}-${{ matrix.shard }}", "name: unix-test-shard-fixed"},
		"wrong-platform-artifacts": {"pattern: unix-test-shard-${{ matrix.os }}-*", "pattern: unix-test-shard-*"},
		"lost-selector":            {"-lane unix-test -shard ${{ matrix.shard }}", "-lane unix-test -shard 0"},
		"lost-boundary":            {"bash .github/ci/isolated-test.sh go run", "go run"},
		"lost-consumer-boundary":   {"bash .github/ci/isolated-test.sh bash .github/ci/test-gate.sh", "bash .github/ci/test-gate.sh"},
		"relative-home":            {"bash .github/ci/isolated-test.sh go run", "HOME=.temp/home go run"},
	} {
		t.Run(name, func(t *testing.T) {
			changed := strings.Replace(string(good), pair[0], pair[1], 1)
			if changed == string(good) {
				t.Fatal("plant not applied")
			}
			if err := checkUnixShards([]byte(changed)); err == nil {
				t.Fatal("Unix wiring violation admitted")
			}
		})
	}
}
