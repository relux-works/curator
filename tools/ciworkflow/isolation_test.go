// Package ciworkflow tests the Linux CI isolation contract at its workflow entry point.
package ciworkflow

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const isolationCommand = "bash .github/ci/isolate-go-state.sh"
const cacheCondition = "${{ runner.environment != 'self-hosted' || runner.os != 'Linux' }}"

type step struct {
	Uses string            `yaml:"uses"`
	Run  string            `yaml:"run"`
	If   string            `yaml:"if"`
	Env  map[string]string `yaml:"env"`
	With map[string]string `yaml:"with"`
}

type workflow struct {
	Concurrency struct {
		Group  string `yaml:"group"`
		Cancel string `yaml:"cancel-in-progress"`
	} `yaml:"concurrency"`
	Jobs map[string]struct {
		RunsOn any               `yaml:"runs-on"`
		Env    map[string]string `yaml:"env"`
		Steps  []step            `yaml:"steps"`
	} `yaml:"jobs"`
}

func workflowBytes(t *testing.T) []byte {
	t.Helper()
	path := os.Getenv("CI_WORKFLOW_UNDER_TEST")
	if path == "" {
		path = "../../.github/workflows/ci.yml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func checkIsolation(data []byte) error {
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return err
	}
	count := 0
	for name, job := range wf.Jobs {
		route, _ := job.RunsOn.(string)
		if !strings.Contains(route, "LINUX_SELF_HOSTED") && !strings.Contains(route, "relux-linux") {
			continue
		}
		count++
		// Only checkout may precede initialization: even an action may invoke Go.
		if len(job.Steps) < 2 || job.Steps[0].Uses != "actions/checkout@v4" || job.Steps[1].Run != isolationCommand || job.Steps[1].If != "runner.os == 'Linux'" {
			return fmt.Errorf("%s: private Go state must precede every Go consumer", name)
		}
		for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH"} {
			if _, exists := job.Env[key]; exists {
				return fmt.Errorf("%s: job overrides %s", name, key)
			}
			for _, s := range job.Steps {
				if _, exists := s.Env[key]; exists {
					return fmt.Errorf("%s: step overrides %s", name, key)
				}
			}
		}
		for _, s := range job.Steps {
			if strings.HasPrefix(s.Uses, "actions/setup-go@") && s.With["cache"] != cacheCondition {
				return fmt.Errorf("%s: setup-go may restore shared Go cache", name)
			}
			if strings.HasPrefix(s.Uses, "actions/cache") {
				return fmt.Errorf("%s: unqualified cache action", name)
			}
			if strings.HasPrefix(s.Uses, "golangci/golangci-lint-action@") && s.With["skip-cache"] != "${{ runner.environment == 'self-hosted' && runner.os == 'Linux' }}" {
				return fmt.Errorf("%s: lint action may restore shared state", name)
			}
		}
	}
	if count < 6 {
		return fmt.Errorf("found %d Linux-routable jobs, want at least six", count)
	}
	if wf.Concurrency.Group != "${{ github.workflow }}-${{ github.ref }}" || wf.Concurrency.Cancel != "${{ github.event_name == 'pull_request' }}" {
		return fmt.Errorf("workflow concurrency must serialize each workflow/ref and cancel only pull requests")
	}
	live := 0
	for _, s := range wf.Jobs["lint"].Steps {
		if s.Run == "go test -count=1 ./tools/ciworkflow/" && s.If == "" {
			live++
		}
	}
	if live != 1 {
		return fmt.Errorf("lint must run the isolation gate unconditionally exactly once")
	}
	return nil
}

// This production-file gate rejects ordering/cache/concurrency regressions; it does not reproduce an ABI linker race.
func TestLinuxCIWorkflow(t *testing.T) {
	if err := checkIsolation(workflowBytes(t)); err != nil {
		t.Fatal(err)
	}
}

// Each narrowing plant changes one job or one concurrency boundary while retaining the surrounding gates.
func TestLinuxCIWorkflowRejectsNarrowingPlants(t *testing.T) {
	data := workflowBytes(t)
	for _, tc := range []struct{ name, old, replacement, finding string }{
		{"one_job_skips_initialization", "run: " + isolationCommand, "run: echo '" + isolationCommand + "'", "private Go state"},
		{"one_setup_restores_shared_cache", "cache: " + cacheCondition, "cache: true # " + cacheCondition, "shared Go cache"},
		{"cancel_main", "cancel-in-progress: ${{ github.event_name == 'pull_request' }}", "cancel-in-progress: ${{ github.event_name != 'workflow_dispatch' }}", "concurrency"},
		{"split_same_ref", "group: ${{ github.workflow }}-${{ github.ref }}", "group: ${{ github.workflow }}-${{ github.ref }}-${{ github.run_id }}", "concurrency"},
		{"skip_gate_on_push", "run: go test -count=1 ./tools/ciworkflow/", "if: github.event_name == 'pull_request'\n        run: go test -count=1 ./tools/ciworkflow/", "unconditionally"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutant := strings.Replace(string(data), tc.old, tc.replacement, 1)
			if mutant == string(data) {
				t.Fatal("plant was not applied")
			}
			if err := checkIsolation([]byte(mutant)); err == nil || !strings.Contains(err.Error(), tc.finding) {
				t.Fatalf("want %q refusal, got %v", tc.finding, err)
			}
		})
	}
}

func initialize(t *testing.T, script, temp string) map[string]string {
	t.Helper()
	envFile := filepath.Join(t.TempDir(), "github-env")
	cmd := exec.Command("bash", script)
	cmd.Env = append(os.Environ(), "RUNNER_TEMP="+temp, "GITHUB_ENV="+envFile,
		"GOCACHE=/shared/go-build", "GOMODCACHE=/shared/gomod", "GOPATH=/shared/gopath")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("initializer: %v: %s", err, out)
	}
	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("invalid GITHUB_ENV record %q", line)
		}
		values[key] = value
	}
	return values
}

func assertPrivate(t *testing.T, first, second map[string]string, temp string) {
	t.Helper()
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH"} {
		for _, values := range []map[string]string{first, second} {
			rel, err := filepath.Rel(temp, values[key])
			if err != nil || rel == "." || strings.HasPrefix(rel, "..") || values[key] == "" {
				t.Fatalf("%s is outside private runner temp", key)
			}
			if info, err := os.Stat(values[key]); err != nil || !info.IsDir() {
				t.Fatalf("%s directory missing", key)
			}
		}
		if first[key] == second[key] {
			t.Fatalf("%s is shared by two jobs", key)
		}
	}
}

// Real shell entry point: two jobs sharing RUNNER_TEMP still get disjoint state despite poisoned ambient cache variables.
func TestGoStateInitialization(t *testing.T) {
	script := os.Getenv("CI_ISOLATION_SCRIPT_UNDER_TEST")
	if script == "" {
		script = "../../.github/ci/isolate-go-state.sh"
	}
	temp := t.TempDir()
	first, second := initialize(t, script, temp), initialize(t, script, temp)
	assertPrivate(t, first, second, temp)
	cmd := exec.Command("go", "env", "GOCACHE", "GOMODCACHE", "GOPATH")
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOENV=off")
	for key, value := range first {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go env: %v: %s", err, out)
	}
	got := strings.Fields(string(out))
	want := []string{first["GOCACHE"], first["GOMODCACHE"], first["GOPATH"]}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Go did not consume the published private state: %s", out)
	}
}

// Missing runner context fails before publishing partial environment state; no ambient-cache fallback is allowed.
func TestGoStateMissingRunnerContext(t *testing.T) {
	for _, missing := range []string{"RUNNER_TEMP", "GITHUB_ENV"} {
		t.Run(missing, func(t *testing.T) {
			temp := t.TempDir()
			envFile := filepath.Join(temp, "github-env")
			cmd := exec.Command("bash", "../../.github/ci/isolate-go-state.sh")
			cmd.Env = append(os.Environ(), "RUNNER_TEMP="+temp, "GITHUB_ENV="+envFile, missing+"=")
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), missing+" is required") {
				t.Fatalf("want missing %s refusal, got %v: %s", missing, err, out)
			}
			entries, readErr := os.ReadDir(temp)
			if readErr != nil || len(entries) != 0 {
				t.Fatalf("partial state after refusal: %v, %v", entries, readErr)
			}
		})
	}
}

// Execute the actual routing expressions in their JS-compatible subset; covers forks, trusted PRs, push and dispatch on every matrix OS.
// GitHub scheduling itself is evidenced by the remote gate, not by this expression test.
func TestLinuxRoutingKeepsForksHosted(t *testing.T) {
	var wf workflow
	if err := yaml.Unmarshal(workflowBytes(t), &wf); err != nil {
		t.Fatal(err)
	}
	for name, job := range wf.Jobs {
		route, _ := job.RunsOn.(string)
		if !strings.Contains(route, "LINUX_SELF_HOSTED") {
			continue
		}
		for _, event := range []string{"pull_request", "push", "workflow_dispatch"} {
			for _, enabled := range []string{"true", "false"} {
				for _, fork := range []bool{true, false} {
					for _, platform := range []string{"ubuntu-latest", "macos-latest", "windows-latest"} {
						osName := platform
						if !strings.Contains(route, "matrix.os") {
							osName = "ubuntu-latest"
						}
						head := "example/repository"
						if fork {
							head = "fork/repository"
						}
						ctx, err := json.Marshal(map[string]any{"vars": map[string]string{"LINUX_SELF_HOSTED": enabled}, "matrix": map[string]string{"os": osName}, "github": map[string]any{"event_name": event, "repository": "example/repository", "event": map[string]any{"pull_request": map[string]any{"head": map[string]any{"repo": map[string]string{"full_name": head}}}}}})
						if err != nil {
							t.Fatal(err)
						}
						expr := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(route, "${{"), "}}"))
						program := "const {vars,matrix,github}=JSON.parse(process.argv[1]); const fromJSON=JSON.parse; console.log(JSON.stringify(" + expr + "));"
						out, err := exec.Command("node", "-e", program, string(ctx)).CombinedOutput()
						if err != nil {
							t.Fatalf("%s routing: %v: %s", name, err, out)
						}
						want := `"` + osName + `"`
						if enabled == "true" && osName == "ubuntu-latest" && (event != "pull_request" || !fork) {
							want = `["self-hosted","relux-linux","curator"]`
						}
						if strings.TrimSpace(string(out)) != want {
							t.Fatalf("%s %s enabled=%s fork=%v %s: got %s want %s", name, event, enabled, fork, platform, out, want)
						}
					}
				}
			}
		}
	}
}
