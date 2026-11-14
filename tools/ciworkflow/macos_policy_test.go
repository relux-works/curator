package ciworkflow

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const fullMacOSCondition = "github.event_name == 'workflow_dispatch' || github.event_name == 'merge_group' || (github.event_name == 'push' && github.ref == 'refs/heads/main') || (github.event_name == 'pull_request' && contains(github.event.pull_request.labels.*.name, 'landing-candidate'))"
const macOSExclude = "${{ fromJSON((" + fullMacOSCondition + ") && '[]' || '[{\"os\":\"macos-latest\"}]') }}"
const unixMacOSExclude = "${{ fromJSON((" + fullMacOSCondition + ") && '[{\"os\":\"macos-latest\",\"shard\":2}]' || '[{\"os\":\"macos-latest\"}]') }}"

var macOSMatrixJobs = []string{"go-family-qualification", "unix-test-shards", "test", "envprofile-race", "race", "gate-selftest"}

type macOSPolicyJob struct {
	If       string
	Needs    []string
	RunsOn   any `yaml:"runs-on"`
	Timeout  int `yaml:"timeout-minutes"`
	Strategy struct{ Matrix map[string]any }
	Steps    []step
}

func checkMacOSWiring(data []byte) error {
	var wf struct {
		On struct {
			PullRequest struct{ Types []string } `yaml:"pull_request"`
		}
		Jobs map[string]macOSPolicyJob
	}
	// Other jobs use a scalar needs; inspect only the policy's own jobs.
	var raw struct {
		On   yaml.Node
		Jobs map[string]yaml.Node
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}
	if err := raw.On.Decode(&wf.On); err != nil {
		return err
	}
	wf.Jobs = make(map[string]macOSPolicyJob)
	seen := map[string]bool{}
	for name, node := range raw.Jobs {
		if name != "macos-policy-gate" && name != "macos-smoke" {
			// Matrix jobs need no additional dependency or job-level skip.
			var job struct {
				Strategy struct{ Matrix map[string]any }
			}
			if err := node.Decode(&job); err != nil {
				return err
			}
			if len(job.Strategy.Matrix) == 0 {
				continue
			}
			matrix := job.Strategy.Matrix
			platforms, _ := matrix["os"].([]any)
			hasMac := false
			for _, platform := range platforms {
				hasMac = hasMac || platform == "macos-latest"
			}
			if !hasMac {
				continue
			}
			if name == "candidate-conformance" {
				continue
			} // dispatch-only, pinned below
			found := false
			for _, expected := range macOSMatrixJobs {
				found = found || name == expected
			}
			want := macOSExclude
			if name == "unix-test-shards" {
				want = unixMacOSExclude
			}
			wantPlatforms := []any{"ubuntu-latest", "macos-latest", "windows-latest"}
			if name == "unix-test-shards" || name == "race" || name == "envprofile-race" {
				wantPlatforms = wantPlatforms[:2]
			}
			if !reflect.DeepEqual(platforms, wantPlatforms) {
				return fmt.Errorf("%s: platform set changed", name)
			}
			seen[name] = true
			if !found || matrix["exclude"] != want {
				return fmt.Errorf("%s: missing landing selection or aggregate coverage", name)
			}
			continue
		}
		job := wf.Jobs[name]
		if err := node.Decode(&job); err != nil {
			return err
		}
		wf.Jobs[name] = job
	}
	for _, name := range macOSMatrixJobs {
		if !seen[name] {
			return fmt.Errorf("%s: native matrix missing", name)
		}
	}
	if !reflect.DeepEqual(wf.On.PullRequest.Types, []string{"opened", "synchronize", "reopened", "labeled", "unlabeled"}) {
		return fmt.Errorf("PR label changes must trigger policy reevaluation")
	}
	gate := wf.Jobs["macos-policy-gate"]
	wantNeeds := append([]string{"macos-smoke"}, macOSMatrixJobs...)
	wantNeeds = append(wantNeeds, "candidate-conformance")
	if gate.If != "always()" || !reflect.DeepEqual(gate.Needs, wantNeeds) || len(gate.Steps) != 2 ||
		gate.Steps[1].Run != "python3 .github/ci/macos-policy-gate.py" || gate.Steps[1].Env["CI_NEEDS"] != "${{ toJSON(needs) }}" {
		return fmt.Errorf("gate must inspect every real result even on failure or skip")
	}
	smoke := wf.Jobs["macos-smoke"]
	if smoke.If != "github.event_name == 'pull_request'" || smoke.RunsOn != "macos-latest" || smoke.Timeout != 10 {
		return fmt.Errorf("one bounded native smoke must run on every PR")
	}
	build, tests := false, false
	for _, s := range smoke.Steps {
		if s.If != "" || s.Continue {
			continue
		}
		build = build || s.Run == "set -euo pipefail\nbash .github/ci/isolated-test.sh go build ./...\nbash .github/ci/isolated-test.sh go vet ./...\n"
		tests = tests || s.Run == "bash .github/ci/macos-smoke.sh"
	}
	if !build || !tests {
		return fmt.Errorf("smoke must build, vet and require exact test witnesses")
	}
	var candidate struct{ If string }
	node := raw.Jobs["candidate-conformance"]
	if err := node.Decode(&candidate); err != nil {
		return err
	}
	if candidate.If != "github.event_name == 'workflow_dispatch' && (inputs.candidate_ref != '' || inputs.candidate_root != '')" {
		return fmt.Errorf("optional candidate matrix must remain dispatch-only")
	}
	return nil
}

// The production wiring must select all Darwin jobs together, preserve label
// reevaluation, and report skipped/failed dependencies through the actual gate.
// Plants are workflow contract checks; GitHub runner timing is measured separately.
func TestMacOSPolicyWiring(t *testing.T) {
	good := workflowBytes(t)
	if err := checkMacOSWiring(good); err != nil {
		t.Fatal(err)
	}
	for name, pair := range map[string][2]string{
		"lost-label-trigger":   {"reopened, labeled, unlabeled", "reopened, unlabeled"},
		"lost-removal-trigger": {"reopened, labeled, unlabeled", "reopened, labeled"},
		"skipped-gate":         {"    if: always()\n    needs: [macos-smoke", "    if: success()\n    needs: [macos-smoke"},
		"lost-dependency":      {"macos-smoke, go-family-qualification,", "macos-smoke,"},
		"fake-results":         {"CI_NEEDS: ${{ toJSON(needs) }}", "CI_NEEDS: '{}'"},
		"lost-policy":          {"        exclude: " + macOSExclude, "        exclude: []"},
		"always-excluded-unix": {"        exclude: " + unixMacOSExclude, "        exclude: [{os: macos-latest}]"},
		"missing-smoke":        {"if: github.event_name == 'pull_request'", "if: false"},
		"lost-macos-platform":  {"os: [ubuntu-latest, macos-latest, windows-latest]", "os: [ubuntu-latest, windows-latest]"},
		"ignored-smoke":        {"run: bash .github/ci/macos-smoke.sh", "run: 'true'"},
	} {
		t.Run(name, func(t *testing.T) {
			changed := strings.Replace(string(good), pair[0], pair[1], 1)
			if changed == string(good) {
				t.Fatal("plant not applied")
			}
			if err := checkMacOSWiring([]byte(changed)); err == nil {
				t.Fatal("policy violation admitted")
			}
		})
	}
}

// Execute the real exclude expressions with independent event/count expectations.
// Linux/Windows entries stay identical; ordinary PRs select zero full Darwin jobs.
func TestMacOSMatrixSelection(t *testing.T) {
	var wf struct {
		Jobs map[string]struct {
			Strategy struct{ Matrix map[string]any }
		}
	}
	if err := yaml.Unmarshal(workflowBytes(t), &wf); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, event, ref, label string
		full                    bool
	}{
		{"pr", "pull_request", "refs/pull/1/merge", "", false},
		{"unrelated-label", "pull_request", "refs/pull/1/merge", "other", false},
		{"landing", "pull_request", "refs/pull/1/merge", "landing-candidate", true},
		{"case-insensitive-label", "pull_request", "refs/pull/1/merge", "LANDING-CANDIDATE", true},
		{"main", "push", "refs/heads/main", "", true},
		{"gate-push", "push", "refs/heads/gate/control", "", false},
		{"dispatch", "workflow_dispatch", "refs/heads/control", "", true},
		{"train", "merge_group", "refs/heads/gh-readonly-queue/main/control", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range macOSMatrixJobs {
				expr, ok := wf.Jobs[name].Strategy.Matrix["exclude"].(string)
				if !ok {
					t.Fatalf("%s: missing expression", name)
				}
				expr = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expr, "${{"), "}}"))
				expr = strings.ReplaceAll(expr, "github.event.pull_request.labels.*.name", "github.event.pull_request.labels.map(x => x.name)")
				ctx, _ := json.Marshal(map[string]any{"event_name": tc.event, "ref": tc.ref, "event": map[string]any{"pull_request": map[string]any{"labels": []any{map[string]string{"name": tc.label}}}}})
				program := `const github=JSON.parse(process.argv[1]); const fromJSON=JSON.parse; const contains=(a,b)=>a.some(x=>x.toLowerCase()===b.toLowerCase()); console.log(JSON.stringify(` + expr + `));`
				out, err := exec.Command("node", "-e", program, string(ctx)).CombinedOutput()
				if err != nil {
					t.Fatalf("%s: %v: %s", name, err, out)
				}
				want := `[{"os":"macos-latest"}]`
				if tc.full {
					want = `[]`
					if name == "unix-test-shards" {
						want = `[{"os":"macos-latest","shard":2}]`
					}
				}
				if strings.TrimSpace(string(out)) != want {
					t.Fatalf("%s: got %s want %s", name, out, want)
				}
			}
		})
	}
}
