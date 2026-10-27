package ciworkflow

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func checkEnvprofileRace(data []byte) error {
	var wf struct {
		Jobs map[string]struct {
			If       string `yaml:"if"`
			Needs    any    `yaml:"needs"`
			RunsOn   any    `yaml:"runs-on"`
			Strategy struct {
				FailFast bool `yaml:"fail-fast"`
				Matrix   struct {
					OS    []string `yaml:"os"`
					Shard []int    `yaml:"shard"`
				} `yaml:"matrix"`
			} `yaml:"strategy"`
			Steps []step `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return err
	}
	shards, ok := wf.Jobs["envprofile-race"]
	if !ok {
		return fmt.Errorf("missing envprofile-race job")
	}
	race := wf.Jobs["race"]
	if shards.If != "" || shards.Strategy.FailFast || !reflect.DeepEqual(shards.Strategy.Matrix.Shard, []int{0, 1, 2, 3, 4, 5, 6, 7}) || !reflect.DeepEqual(shards.Strategy.Matrix.OS, race.Strategy.Matrix.OS) || !reflect.DeepEqual(shards.RunsOn, race.RunsOn) {
		return fmt.Errorf("shards must run every index on both Race platforms with identical routing, on main and PRs")
	}
	if race.Needs != "envprofile-race" {
		return fmt.Errorf("Race must require the completed envprofile shards")
	}
	run, upload, download, consume := 0, 0, 0, 0
	for _, s := range shards.Steps {
		if s.If == "" && s.Run == "env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS -u CODEX_HOME -u CLAUDE_CONFIG_DIR -u PI_CODING_AGENT_DIR -u XDG_DATA_HOME -u XDG_STATE_HOME -u XDG_CACHE_HOME HOME=\"$(mktemp -d)\" go run ./tools/envprofilerace -shard ${{ matrix.shard }} -dir .temp/ci-evidence/envprofile" && s.Env["CURATOR_CONFORMANCE_ROOT"] == "${{ github.workspace }}/protocol-spec/conformance/v1" {
			run++
		}
		if s.Uses == "actions/upload-artifact@v4" && s.If == "always()" && s.With["name"] == "envprofile-race-${{ matrix.os }}-${{ matrix.shard }}" && s.With["path"] == ".temp/ci-evidence/envprofile/" && s.With["if-no-files-found"] == "error" {
			upload++
		}
	}
	for _, s := range race.Steps {
		if s.Uses == "actions/download-artifact@v4" && s.If == "" && s.With["pattern"] == "envprofile-race-${{ matrix.os }}-*" && s.With["path"] == ".temp/envprofile-shards" && s.With["merge-multiple"] != "true" {
			download++
		}
		if strings.Contains(s.Run, "test-gate.sh") && s.If == "" && s.Env["CI_ENVPROFILE_SHARDS"] == ".temp/envprofile-shards" && s.Env["GO_TEST_FLAGS"] == "-race" && s.Env["GO_TEST_TIMEOUT"] == "60m" && download == 1 {
			consume++
		}
	}
	if run != 1 || upload != 1 || download != 1 || consume != 1 {
		return fmt.Errorf("shard run/upload/download/union wiring: %d/%d/%d/%d", run, upload, download, consume)
	}
	return nil
}

// The actual workflow must preserve all shards on main and PRs and make the
// existing platform gate consume their same-platform union before it can pass.
// Each plant retains valid YAML and must fail this wiring contract.
func TestEnvprofileRaceWiring(t *testing.T) {
	good := workflowBytes(t)
	if err := checkEnvprofileRace(good); err != nil {
		t.Fatal(err)
	}
	for name, pair := range map[string][2]string{
		"dropped-shard":  {"shard: [0, 1, 2, 3, 4, 5, 6, 7]", "shard: [0, 1, 2, 3, 4, 5, 6]"},
		"no-dependency":  {"needs: envprofile-race", "needs: test"},
		"only-main":      {"  envprofile-race:\n", "  envprofile-race:\n    if: github.event_name == 'push'\n"},
		"wrong-platform": {"pattern: envprofile-race-${{ matrix.os }}-*", "pattern: envprofile-race-ubuntu-latest-*"},
		"no-consumption": {"CI_ENVPROFILE_SHARDS: .temp/envprofile-shards", "CI_OTHER_SHARDS: .temp/envprofile-shards"},
		"lost-selector":  {"-shard ${{ matrix.shard }}", "-shard 0"},
		"lost-upload":    {"name: envprofile-race-${{ matrix.os }}-${{ matrix.shard }}", "name: envprofile-race-${{ matrix.os }}"},
	} {
		t.Run(name, func(t *testing.T) {
			planted := strings.Replace(string(good), pair[0], pair[1], 1)
			if planted == string(good) {
				t.Fatal("plant not applied")
			}
			if err := checkEnvprofileRace([]byte(planted)); err == nil {
				t.Fatal("wiring plant survived")
			}
		})
	}
}
