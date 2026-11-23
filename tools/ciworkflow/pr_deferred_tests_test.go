package ciworkflow

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/curatormutations"
	"gopkg.in/yaml.v3"
)

type pullRequestStep struct {
	Name             string    `yaml:"name"`
	If               string    `yaml:"if"`
	Run              string    `yaml:"run"`
	Uses             string    `yaml:"uses"`
	WorkingDirectory string    `yaml:"working-directory"`
	With             yaml.Node `yaml:"with"`
	Env              yaml.Node `yaml:"env"`
}

type pullRequestJob struct {
	Name     string            `yaml:"name"`
	If       string            `yaml:"if"`
	Uses     string            `yaml:"uses"`
	Needs    yaml.Node         `yaml:"needs"`
	Env      yaml.Node         `yaml:"env"`
	Strategy yaml.Node         `yaml:"strategy"`
	Steps    []pullRequestStep `yaml:"steps"`
}

type pullRequestWorkflow struct {
	On   yaml.Node                 `yaml:"on"`
	Env  yaml.Node                 `yaml:"env"`
	Jobs map[string]pullRequestJob `yaml:"jobs"`
}

const legacyMutationJobID = "curator-mutation-pr"

var mutationJobResultReference = regexp.MustCompile(`(?i)\bneeds\s*(?:\.\s*([A-Za-z_][A-Za-z0-9_-]*)|\[\s*['"]([^'"]+)['"]\s*\])\s*\.\s*result\b`)
var dynamicNeedsResultReference = regexp.MustCompile(`(?i)\bneeds\s*\[\s*[^'"][^\]]+\]\s*\.\s*result\b`)
var githubExpression = regexp.MustCompile(`\$\{\{(?s:(.*?))\}\}`)
var simpleMatrixExpression = regexp.MustCompile(`^matrix\s*\.\s*([A-Za-z_][A-Za-z0-9_-]*)$`)
var workspaceEnvironmentPath = regexp.MustCompile(`^\$\{\{\s*github\.workspace\s*\}\}(?:/[A-Za-z0-9._/-]*)?$`)
var pinnedPackageVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// The current PR workflow uses no job-level reusable workflows. Keep this
// exact-target allowlist empty until a reviewed target is deliberately admitted.
var allowedPRReusableWorkflowTargets = map[string]struct{}{}

// These are the complete static matrix values the PR workflow currently passes
// into run/env: Go versions, runner labels, and shard indexes.
var allowedPRMatrixValues = map[string]map[string]struct{}{
	"go": {
		"1.25.5": {}, "1.26.0": {}, "1.27.1": {},
	},
	"os": {
		"ubuntu-latest": {}, "macos-latest": {}, "windows-latest": {},
	},
	"shard": {
		"0": {}, "1": {}, "2": {}, "3": {}, "4": {}, "5": {}, "6": {}, "7": {}, "8": {},
	},
}

func checkPullRequestDefersPostLandingTests(data []byte) error {
	_, root, err := parseWorkflowDocument(data)
	if err != nil {
		return err
	}
	if err := validateWorkflowMappingKeys(root); err != nil {
		return err
	}
	var workflow pullRequestWorkflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return err
	}
	if err := requirePullRequestTrigger(&workflow.On); err != nil {
		return err
	}
	if workflow.Jobs == nil {
		return fmt.Errorf("workflow has no jobs mapping")
	}
	testJob, ok := workflow.Jobs["test"]
	if !ok {
		return fmt.Errorf("test job is missing")
	}
	wantNeeds := []string{"windows-test-shards", "unix-test-shards", "store-v1-compatibility"}
	needs, err := pullRequestNeedIDs(&testJob.Needs)
	if err != nil || testJob.Needs.Kind != yaml.SequenceNode || !reflect.DeepEqual(needs, wantNeeds) {
		return fmt.Errorf("test job has unexpected dependencies: %v", needs)
	}

	mutationJobIDs := map[string]struct{}{legacyMutationJobID: {}}
	for id, job := range workflow.Jobs {
		if isMutationSuiteJob(id, job) {
			mutationJobIDs[strings.ToLower(id)] = struct{}{}
		}
	}
	for id, job := range workflow.Jobs {
		if !canRunOnPullRequest(job.If) {
			continue
		}
		if uses := strings.TrimSpace(job.Uses); uses != "" {
			if _, ok := allowedPRReusableWorkflowTargets[uses]; !ok {
				return fmt.Errorf("pull-request job %q calls an unapproved reusable workflow %q", id, uses)
			}
		}
		if err := checkPRDeferredTestCommands(id, job, &workflow.Env); err != nil {
			return err
		}
		if err := checkPRLocalTargets(id, job, &workflow.Env); err != nil {
			return err
		}
		if err := checkPRDynamicInputs(id, job, &workflow.Env); err != nil {
			return err
		}
		if isMutationSuiteJob(id, job) {
			return fmt.Errorf("mutation suite job %q can run on pull-request CI", id)
		}
		needs, err := pullRequestNeedIDs(&job.Needs)
		if err != nil {
			return fmt.Errorf("job %q has unreadable needs: %w", id, err)
		}
		for _, dependency := range needs {
			if isMutationJobID(dependency, mutationJobIDs) {
				return fmt.Errorf("pull-request job %q depends on deferred mutation job %q", id, dependency)
			}
		}
		if hasMutationResultReference(job.If, mutationJobIDs) || hasMutationResultReferenceInNode(&job.Needs, mutationJobIDs) || hasMutationResultReferenceInNode(&job.Env, mutationJobIDs) {
			return fmt.Errorf("pull-request job %q consumes a deferred mutation result", id)
		}
		for _, step := range job.Steps {
			if hasMutationResultReference(step.If, mutationJobIDs) || hasMutationResultReference(step.Run, mutationJobIDs) || hasMutationResultReferenceInNode(&step.Env, mutationJobIDs) {
				return fmt.Errorf("pull-request job %q consumes a deferred mutation result", id)
			}
		}
	}
	return nil
}

func requirePullRequestTrigger(on *yaml.Node) error {
	on = resolveWorkflowAlias(on)
	if on == nil {
		return fmt.Errorf("workflow has no trigger")
	}
	switch on.Kind {
	case yaml.ScalarNode:
		if on.Tag == "!!str" && on.Value == "pull_request" {
			return nil
		}
		return fmt.Errorf("workflow scalar trigger has no pull_request event")
	case yaml.SequenceNode:
		seen := make(map[string]struct{}, len(on.Content))
		found := false
		for _, event := range on.Content {
			if event.Kind != yaml.ScalarNode || event.Tag != "!!str" {
				return fmt.Errorf("workflow trigger has an unreadable event")
			}
			if _, ok := seen[event.Value]; ok {
				return fmt.Errorf("workflow trigger repeats event %q", event.Value)
			}
			seen[event.Value] = struct{}{}
			if event.Value == "pull_request" {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("workflow has no pull_request trigger")
		}
		return nil
	case yaml.MappingNode:
		// Continue below with the event-filter mapping form.
	default:
		return fmt.Errorf("workflow trigger must be a scalar, sequence, or mapping")
	}
	seen := make(map[string]struct{}, len(on.Content)/2)
	found := false
	for i := 0; i+1 < len(on.Content); i += 2 {
		key := on.Content[i]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
			return fmt.Errorf("workflow trigger has an unreadable event key")
		}
		if _, ok := seen[key.Value]; ok {
			return fmt.Errorf("workflow trigger repeats event %q", key.Value)
		}
		seen[key.Value] = struct{}{}
		if key.Value != "pull_request" {
			continue
		}
		event := resolveWorkflowAlias(on.Content[i+1])
		if event == nil || (event.Kind != yaml.MappingNode && (event.Kind != yaml.ScalarNode || event.Tag != "!!null")) {
			return fmt.Errorf("pull_request trigger has an unsupported value")
		}
		found = true
	}
	if !found {
		return fmt.Errorf("workflow has no pull_request trigger")
	}
	return nil
}

func resolveWorkflowAlias(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		if node.Alias == nil {
			return nil
		}
		node = node.Alias
	}
	return node
}

func validateWorkflowMappingKeys(root *yaml.Node) error {
	seenNodes := make(map[*yaml.Node]struct{})
	activeNodes := make(map[*yaml.Node]struct{})
	var visit func(*yaml.Node) error
	visit = func(node *yaml.Node) error {
		if node == nil {
			return nil
		}
		if node.Kind == yaml.AliasNode {
			if node.Alias == nil {
				return fmt.Errorf("workflow has an unresolved YAML alias")
			}
			node = node.Alias
		}
		if _, ok := activeNodes[node]; ok {
			return fmt.Errorf("workflow has a cyclic YAML alias")
		}
		if _, ok := seenNodes[node]; ok {
			return nil
		}
		activeNodes[node] = struct{}{}
		defer delete(activeNodes, node)
		if node.Kind == yaml.MappingNode {
			if len(node.Content)%2 != 0 {
				return fmt.Errorf("workflow has an incomplete mapping")
			}
			keys := make(map[string]struct{}, len(node.Content)/2)
			for i := 0; i+1 < len(node.Content); i += 2 {
				key := node.Content[i]
				if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
					return fmt.Errorf("workflow has a non-string mapping key")
				}
				if _, ok := keys[key.Value]; ok {
					return fmt.Errorf("workflow mapping repeats key %q", key.Value)
				}
				keys[key.Value] = struct{}{}
				if err := visit(node.Content[i+1]); err != nil {
					return err
				}
			}
		} else {
			for _, child := range node.Content {
				if err := visit(child); err != nil {
					return err
				}
			}
		}
		seenNodes[node] = struct{}{}
		return nil
	}
	return visit(root)
}

func checkPRDynamicInputs(jobID string, job pullRequestJob, workflowEnv *yaml.Node) error {
	matrixValues, err := pullRequestMatrixValues(&job.Strategy)
	if err != nil {
		return fmt.Errorf("pull-request job %q has an unclassifiable matrix: %w", jobID, err)
	}
	checkText := func(text, location string) error {
		if strings.Count(text, "${{") != strings.Count(text, "}}") {
			return fmt.Errorf("pull-request job %q has an unbalanced dynamic input in %s", jobID, location)
		}
		for _, match := range githubExpression.FindAllStringSubmatch(text, -1) {
			expression := strings.TrimSpace(match[1])
			if strings.Contains(strings.ToLower(expression), "matrix") {
				fieldMatch := simpleMatrixExpression.FindStringSubmatch(expression)
				if len(fieldMatch) != 2 {
					return fmt.Errorf("pull-request job %q has an unclassifiable matrix input %q in %s", jobID, expression, location)
				}
				if err := validatePRMatrixField(fieldMatch[1], matrixValues); err != nil {
					return fmt.Errorf("pull-request job %q: %w", jobID, err)
				}
				continue
			}
			switch expression {
			case "github.workspace":
				if location != "env" || !workspaceEnvironmentPath.MatchString(text) {
					return fmt.Errorf("pull-request job %q has an unclassifiable workspace input in %s", jobID, location)
				}
			case "env.PNPM_PIN":
				pin, ok := workflowStringValue(workflowEnv, "PNPM_PIN")
				if location != "run" || !ok || !pinnedPackageVersion.MatchString(pin) {
					return fmt.Errorf("pull-request job %q has an unclassifiable package-pin input in %s", jobID, location)
				}
			case "runner.os == 'Windows' && '120m' || '60m'":
				if location != "env" || strings.TrimSpace(text) != match[0] {
					return fmt.Errorf("pull-request job %q has an unclassifiable runner input in %s", jobID, location)
				}
			default:
				return fmt.Errorf("pull-request job %q has an unclassifiable dynamic input %q in %s", jobID, expression, location)
			}
		}
		return nil
	}
	for _, value := range workflowScalarValues(&job.Env) {
		if err := checkText(value, "env"); err != nil {
			return err
		}
	}
	for _, step := range job.Steps {
		if err := checkText(step.Run, "run"); err != nil {
			return err
		}
		for _, value := range workflowScalarValues(&step.Env) {
			if err := checkText(value, "env"); err != nil {
				return err
			}
		}
	}
	return nil
}

type workflowShellWord struct {
	value    string
	quote    byte
	operator bool
}

var workflowShellVariable = regexp.MustCompile(`^\$(?:\{([A-Za-z_][A-Za-z0-9_]*)\}|([A-Za-z_][A-Za-z0-9_]*))$`)
var prPythonInterpreterVersion = regexp.MustCompile(`(?i)^python(?:2|3)(?:\.\d+)*$`)
var prRubyInterpreterVersion = regexp.MustCompile(`(?i)^ruby(?:\d+)(?:\.\d+)*$`)
var prNodeInterpreterVersion = regexp.MustCompile(`(?i)^(?:node|nodejs)\d+(?:\.\d+)*$`)
var prPerlInterpreterVersion = regexp.MustCompile(`(?i)^perl\d+(?:\.\d+)*$`)
var prPHPInterpreterVersion = regexp.MustCompile(`(?i)^php\d+(?:\.\d+)*$`)
var prVersionedShellInterpreter = regexp.MustCompile(`(?i)^(?:bash|sh|dash|ash|zsh|ksh|mksh|yash|fish|csh|tcsh)\d+(?:\.\d+)*$`)
var prEnvSplitString = regexp.MustCompile(`(?i)(?:^|[\s;&|])env\s+(?:-S|--split-string)(?:\s|$)`)
var prShellFunctionDefinition = regexp.MustCompile(`(?im)(?:^|[;&|]\s*)(?:function\s+)?[A-Za-z_][A-Za-z0-9_]*\s*\(\s*\)\s*\{`)
var localWorkflowScriptInvocation = regexp.MustCompile(`(?i)\b(?:bash|sh|python3?|node|pwsh|powershell)\s+["']?((?:\./|\.\./|\.github/|[A-Za-z0-9_.-]+/|/)[^"' \t\r\n;&|]+|[A-Za-z0-9_][A-Za-z0-9_.-]*)`)
var localWorkflowScriptToken = regexp.MustCompile(`(?:^|[\s;&|])((?:(?:\./|\.\./|\.github/|[A-Za-z0-9_.-]+/|/)[^"' \t\r\n;&|]+|[A-Za-z0-9_.-]+)\.(?:sh|bash|zsh|py|js|ps1|cmd|bat|rb|pl))(?:$|[\s;&|])`)
var localWorkflowExecutable = regexp.MustCompile(`(?m)(?:^|[;&|]\s*)((?:\./|\.\./|[A-Za-z0-9_.-]+/)[^\s;&|]+)(?:$|[\s;&|])`)
var dynamicWorkflowScriptInvocation = regexp.MustCompile(`(?i)\b(?:bash|sh|python3?|node|pwsh|powershell)\s+["']?\$(?:\{|[A-Za-z_(])`)
var dynamicWorkflowSourceInvocation = regexp.MustCompile(`(?im)(?:^|[;&|]\s*)(?:source|\.)\s+["']?\$(?:\{|[A-Za-z_(])`)
var inlineWorkflowScriptBody = regexp.MustCompile(`(?i)\b(?:bash|sh|zsh)\s+-[A-Za-z]*c(?:\s|$)|\bpython3?\s+-c(?:\s|$)|\bnode\s+-e(?:\s|$)|\beval(?:\s|$)`)
var inlineWorkflowHeredoc = regexp.MustCompile(`(?im)^\s*(?:python3?|bash|sh|zsh|node|pwsh|powershell)\s+[^\n]*<<-?\s*["']?[A-Za-z_][A-Za-z0-9_]*["']?\s*$`)
var testGateSkipOverride = regexp.MustCompile(`(?i)(?:^|[\s;&|])CI_SKIP_POST_LANDING_TESTS=([^\s;&|]+)`)
var testGateSkipUnset = regexp.MustCompile(`(?i)(?:^|[\s;&|])(?:-u\s+|unset\s+)CI_SKIP_POST_LANDING_TESTS(?:[\s;&|]|$)`)
var isolatedWorkflowEnvironment = regexp.MustCompile(`(?i)\benv\s+-i(?:\s|$)`)

// Local scripts are admitted by exact path. The test gate is special: its
// workflow step must also set CI_SKIP_POST_LANDING_TESTS=1.
var allowedPRLocalScripts = map[string]struct{}{
	".github/ci/candidate-suite.sh":                 {},
	".github/ci/default-build-transport-control.py": {},
	".github/ci/gate-selftest.sh":                   {},
	".github/ci/install-rust-toolchain.sh":          {},
	".github/ci/isolate-go-state.sh":                {},
	".github/ci/isolated-test.sh":                   {},
	".github/ci/ledger-consistency.sh":              {},
	".github/ci/naming-gate.sh":                     {},
	".github/ci/no-broad-suppression.sh":            {},
	".github/ci/pnpm-pin-guard.sh":                  {},
	".github/ci/release-source-gate.sh":             {},
	".github/ci/rust-pin-guard.sh":                  {},
	".github/ci/test-gate.sh":                       {},
	".github/ci/toolchain-identity.sh":              {},
	"scripts/restore-runner-tmp.sh":                 {},
}

// These existing workflow bodies use shell function or trap forms outside the
// static shell model. An edit changes the digest and must satisfy the model.
var allowedPRUnmodeledShellRuns = map[string]struct{}{
	"race\x00Record macOS race capacity\x006132d13dc9aea819d1cd3aac9107a2742c6e83ea9e87807709fe545d91ab8780":                                                          {},
	"project-bin-windows-read-controls\x00Prove native read denial and legacy ownership controls\x001aee247d6840a83336f962d8e1f70aa48e5261993e7a2bf49fe036babf9029f1": {},
	"credential-windows-controls\x00Portable selection and native descriptor-policy controls\x005e49c924dc7d98df0d7b92056a339b89397962e0f2a55c8afce0ee6c13853370":     {},
	"default-build-transport\x00Default-lane credential isolation and refusal controls\x00ca1a083d029d462f33d93a16ab9f7efa121f1ee9d919c3df7000229d99b6915b":           {},
	"default-build-transport\x00Detect a dropped default-lane SSH policy\x0083bf56421a73f944f9ebcf8e3df6c56448975f7de2b98a0140eb1d48d07879f9":                         {},
	"openssh89-fetch\x00Real product fetch with key and policy refusals\x00fa25f69e82fd616751be0bdd623c5c8cb7387a301448f41773b92dc85dc6f9c9":                          {},
	"test\x00Windows fixture compatibility preflight\x003df993df26fed5d667e9099e1c5cd28936509a95e226558e865c8f31b7331207":                                             {},
	"linux-unprivileged-sweep\x00Real non-root process and unreadable-image controls\x006715c9b09642a4e2462ad36048c2353bfd7f010a49c0684f9f389118e1cb14ed":             {},
}

// No workflow directly executes a repository-local file. Keep this explicit
// allowlist empty until a reviewed invocation form is added.
var allowedPRLocalExecutables = map[string]struct{}{}

// Absolute host tools are admitted only within their exact current workflow
// step body, so new arguments cannot inherit a path-only exception.
var allowedPRAbsoluteCommandRuns = map[string]struct{}{
	"openssh89-fetch\x00Install trusted Git and image OpenSSH\x00/usr/bin/ssh\x001cefb3f7f37cd15379a368f7bc90f93f2828f40dca97f10949f2a3e041ec5903":                  {},
	"openssh89-fetch\x00Real product fetch with key and policy refusals\x00/usr/bin/ssh-keygen\x00fa25f69e82fd616751be0bdd623c5c8cb7387a301448f41773b92dc85dc6f9c9": {},
	"openssh89-fetch\x00Real product fetch with key and policy refusals\x00/usr/sbin/sshd\x00fa25f69e82fd616751be0bdd623c5c8cb7387a301448f41773b92dc85dc6f9c9":      {},
}

var allowedPRAbsoluteGoCommands = map[string]struct{}{
	"/opt/go/bin/go": {},
}

// No local action is currently admitted in PR workflows. A future local action
// must add an exact path and exact input map here after its target is reviewed.
var allowedPRLocalActions = map[string]map[string]string{}

var allowedPRLocalGoTools = map[string]struct{}{
	"./tools/check-store-compatibility": {},
	"./tools/envprofilerace":            {},
	"./tools/sanitize-ci-evidence":      {},
}

func checkPRDeferredTestCommands(jobID string, job pullRequestJob, workflowEnv *yaml.Node) error {
	for _, step := range job.Steps {
		stepDirectory, err := prStepWorkingDirectory(jobID, step, workflowEnv, &job.Env, &step.Env)
		if err != nil {
			return err
		}
		words := workflowShellWords(step.Run)
		workingDirectories, err := prWorkflowGoWorkingDirectories(jobID, words, stepDirectory, workflowEnv, &job.Env, &step.Env)
		if err != nil {
			return err
		}
		var goFlags []workflowShellWord
		goFlagsLoaded := false
		for i := 0; i < len(words); i++ {
			if words[i].operator || !isWorkflowGoCommand(words[i].value) {
				continue
			}
			directories := workingDirectories[i]
			if len(directories) == 0 {
				directories = []string{stepDirectory}
			}
			for _, baseDirectory := range directories {
				command, commandIndex, workingDirectory, err := parsePRGoCommand(jobID, words, i, baseDirectory, workflowEnv, &job.Env, &step.Env)
				if err != nil {
					return err
				}
				if command != "test" {
					continue
				}
				if !goFlagsLoaded {
					value, present, readable := workflowEnvironmentLookup("GOFLAGS", workflowEnv, &job.Env, &step.Env)
					if !readable {
						return fmt.Errorf("pull-request job %q has unclassifiable GOFLAGS", jobID)
					}
					if present {
						goFlags = workflowShellWords(value)
					}
					goFlagsLoaded = true
				}
				end := commandIndex + 1
				for end < len(words) && !words[end].operator {
					end++
				}
				if err := checkPRGoTestArguments(jobID, words[commandIndex+1:end], goFlags, workingDirectory, workflowEnv, &job.Env, &step.Env); err != nil {
					return err
				}
			}
			end := i + 1
			for end < len(words) && !words[end].operator {
				end++
			}
			i = end - 1
		}
	}
	return nil
}

type prShellDirectoryState struct {
	directory string
	stack     []string
}

type prShellDirectoryEffect struct {
	success []prShellDirectoryState
	failure []prShellDirectoryState
}

type prShellDirectoryFrame struct {
	states []prShellDirectoryState
}

var workflowShellAssignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.*$`)

func prWorkflowGoWorkingDirectories(jobID string, words []workflowShellWord, baseDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node) (map[int][]string, error) {
	states := []prShellDirectoryState{{directory: baseDirectory}}
	frames := []prShellDirectoryFrame{}
	workingDirectories := make(map[int][]string)
	var pending *prShellDirectoryEffect
	commandStart := true
	for i := 0; i < len(words); i++ {
		word := words[i]
		if word.operator {
			switch word.value {
			case "(":
				states = clonePRShellDirectoryStates(states)
				frames = append(frames, prShellDirectoryFrame{states: states})
				pending = nil
				commandStart = true
			case ")":
				if len(frames) > 0 {
					states = frames[len(frames)-1].states
					frames = frames[:len(frames)-1]
				}
				pending = nil
				commandStart = true
			case "&&":
				if pending != nil {
					states = clonePRShellDirectoryStates(pending.success)
				}
				pending = nil
				commandStart = true
			case "||":
				if pending != nil {
					states = clonePRShellDirectoryStates(pending.failure)
				}
				pending = nil
				commandStart = true
			case "|":
				if pending != nil {
					states = clonePRShellDirectoryStates(pending.failure)
				}
				pending = nil
				commandStart = true
			default:
				if pending != nil {
					states = unionPRShellDirectoryStates(pending.success, pending.failure)
					pending = nil
				}
				commandStart = true
			}
			continue
		}
		if word.quote == 0 && isPRShellControlKeyword(word.value) {
			commandStart = true
			continue
		}
		if commandStart && word.quote == 0 && strings.HasPrefix(word.value, "#") {
			for i+1 < len(words) && (!words[i+1].operator || words[i+1].value != ";") {
				i++
			}
			continue
		}
		if commandStart {
			if word.quote == 0 {
				if name, value, ok := strings.Cut(word.value, "="); ok && name == "CDPATH" && value != "" {
					return nil, fmt.Errorf("pull-request job %q sets an unmodeled CDPATH value", jobID)
				}
			}
			commandIndex := prWorkflowCommandTarget(words, i)
			if commandIndex < 0 {
				commandStart = false
				continue
			}
			command := words[commandIndex].value
			if command == "export" || command == "declare" || command == "typeset" {
				end := commandIndex + 1
				for end < len(words) && !words[end].operator {
					name, value, ok := strings.Cut(words[end].value, "=")
					if name == "CDPATH" && (!ok || value != "") {
						return nil, fmt.Errorf("pull-request job %q sets an unmodeled CDPATH value", jobID)
					}
					end++
				}
			}
			if command == "cd" || command == "pushd" || command == "popd" {
				end := commandIndex + 1
				for end < len(words) && !words[end].operator {
					end++
				}
				args := words[commandIndex+1 : end]
				success := make([]prShellDirectoryState, 0, len(states))
				failure := clonePRShellDirectoryStates(states)
				switch command {
				case "cd", "pushd":
					if len(args) != 1 || !isLiteralPRDirectory(args[0]) {
						return nil, fmt.Errorf("pull-request job %q has an unclassifiable shell %s directory", jobID, command)
					}
					cdPath, present, readable := workflowEnvironmentLookup("CDPATH", workflowEnv, jobEnv, stepEnv)
					if !readable || (present && cdPath != "") {
						return nil, fmt.Errorf("pull-request job %q has an unclassifiable shell CDPATH", jobID)
					}
					for _, prefix := range words[i:commandIndex] {
						if name, value, ok := strings.Cut(prefix.value, "="); ok && name == "CDPATH" && value != "" {
							return nil, fmt.Errorf("pull-request job %q sets an unmodeled CDPATH value", jobID)
						}
					}
					for _, state := range states {
						directory, err := resolvePRRepoDirectory(jobID, args[0], state.directory, workflowEnv, jobEnv, stepEnv)
						if err != nil {
							return nil, err
						}
						changed := clonePRShellDirectoryState(state)
						if command == "pushd" {
							changed.stack = append(changed.stack, state.directory)
						}
						changed.directory = directory
						success = append(success, changed)
					}
				case "popd":
					if len(args) != 0 {
						return nil, fmt.Errorf("pull-request job %q has an unclassifiable shell popd command", jobID)
					}
					for _, state := range states {
						if len(state.stack) == 0 {
							return nil, fmt.Errorf("pull-request job %q has a popd without a known directory stack", jobID)
						}
						changed := clonePRShellDirectoryState(state)
						changed.directory = changed.stack[len(changed.stack)-1]
						changed.stack = changed.stack[:len(changed.stack)-1]
						success = append(success, changed)
					}
				}
				pending = &prShellDirectoryEffect{success: success, failure: failure}
				states = unionPRShellDirectoryStates(success, failure)
				i = end - 1
				commandStart = false
				continue
			}
			commandStart = false
		}
		if isWorkflowGoCommand(word.value) {
			for _, state := range states {
				workingDirectories[i] = appendUniquePRString(workingDirectories[i], state.directory)
			}
		}
	}
	if len(frames) != 0 {
		return nil, fmt.Errorf("pull-request job %q has an unclosed shell subshell", jobID)
	}
	return workingDirectories, nil
}

func isLiteralPRDirectory(word workflowShellWord) bool {
	value := word.value
	return value != "" && !strings.HasPrefix(value, "-") && !strings.HasPrefix(value, "+") && !path.IsAbs(value) && !strings.HasPrefix(value, "~") &&
		!strings.ContainsAny(value, "$`*?[]{}") && !strings.Contains(value, "${{")
}

func clonePRShellDirectoryState(state prShellDirectoryState) prShellDirectoryState {
	state.stack = append([]string(nil), state.stack...)
	return state
}

func clonePRShellDirectoryStates(states []prShellDirectoryState) []prShellDirectoryState {
	cloned := make([]prShellDirectoryState, len(states))
	for i, state := range states {
		cloned[i] = clonePRShellDirectoryState(state)
	}
	return cloned
}

func unionPRShellDirectoryStates(groups ...[]prShellDirectoryState) []prShellDirectoryState {
	seen := make(map[string]struct{})
	var states []prShellDirectoryState
	for _, group := range groups {
		for _, state := range group {
			key := state.directory + "\x00" + strings.Join(state.stack, "\x00")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			states = append(states, clonePRShellDirectoryState(state))
		}
	}
	return states
}

func appendUniquePRString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func isPRShellControlKeyword(value string) bool {
	switch value {
	case "if", "then", "elif", "else", "fi", "while", "until", "for", "select", "do", "done", "case", "esac", "{", "}", "!":
		return true
	default:
		return false
	}
}

func prStepWorkingDirectory(jobID string, step pullRequestStep, workflowEnv, jobEnv, stepEnv *yaml.Node) (string, error) {
	value := strings.TrimSpace(step.WorkingDirectory)
	if value == "" {
		return ".", nil
	}
	if workspaceEnvironmentPath.MatchString(value) {
		end := strings.Index(value, "}}")
		if end < 0 {
			return ".", nil
		}
		value = strings.TrimSpace(value[end+2:])
		value = strings.TrimPrefix(value, "/")
		if value == "" {
			return ".", nil
		}
	}
	return resolvePRRepoDirectory(jobID, workflowShellWord{value: value}, ".", workflowEnv, jobEnv, stepEnv)
}

func resolvePRRepoDirectory(jobID string, word workflowShellWord, baseDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node) (string, error) {
	resolved, err := resolvePRRepoPath(jobID, word, baseDirectory, "Go working directory", workflowEnv, jobEnv, stepEnv)
	if err != nil {
		return "", err
	}
	root, err := prRepositoryRoot()
	if err != nil {
		return "", fmt.Errorf("pull-request job %q cannot establish the checkout root: %w", jobID, err)
	}
	target, err := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(resolved)))
	if err != nil {
		return "", fmt.Errorf("pull-request job %q has an unclassifiable Go working directory %q", jobID, resolved)
	}
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("pull-request job %q has an unclassifiable Go working directory %q", jobID, resolved)
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("pull-request job %q uses a Go working directory outside the checkout", jobID)
	}
	return filepath.ToSlash(relative), nil
}

func prRepositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return filepath.EvalSymlinks(current)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod was not found in the working directory ancestry")
		}
		current = parent
	}
}

func resolvePRRepoPath(jobID string, word workflowShellWord, baseDirectory, targetKind string, workflowEnv, jobEnv, stepEnv *yaml.Node) (string, error) {
	resolved, err := resolvePRTestSelector(word, workflowEnv, jobEnv, stepEnv)
	if err != nil {
		return "", fmt.Errorf("pull-request job %q has an unclassifiable %s: %w", jobID, targetKind, err)
	}
	if resolved == "" || path.IsAbs(resolved) || strings.ContainsAny(resolved, `\\:`) ||
		(word.quote == 0 && strings.ContainsAny(resolved, " \t\r\n")) {
		return "", fmt.Errorf("pull-request job %q has an unclassifiable %s %q", jobID, targetKind, resolved)
	}
	cleaned := path.Clean(path.Join(baseDirectory, resolved))
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("pull-request job %q uses a %s outside the checkout", jobID, targetKind)
	}
	return cleaned, nil
}

func parsePRGoCommand(jobID string, words []workflowShellWord, goIndex int, baseDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node) (string, int, string, error) {
	workingDirectory := baseDirectory
	seenChangeDirectory := false
	for i := goIndex + 1; i < len(words); {
		word := words[i]
		if word.operator {
			return "", -1, workingDirectory, nil
		}
		value := word.value
		switch {
		case value == "-C":
			if seenChangeDirectory || i+1 >= len(words) || words[i+1].operator {
				return "", -1, workingDirectory, fmt.Errorf("pull-request job %q has an unreadable Go -C directory", jobID)
			}
			directory, err := resolvePRRepoDirectory(jobID, words[i+1], workingDirectory, workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return "", -1, workingDirectory, err
			}
			workingDirectory = directory
			seenChangeDirectory = true
			i += 2
		case strings.HasPrefix(value, "-C="):
			if seenChangeDirectory {
				return "", -1, workingDirectory, fmt.Errorf("pull-request job %q repeats Go -C", jobID)
			}
			pathValue := value[strings.IndexByte(value, '=')+1:]
			directory, err := resolvePRRepoDirectory(jobID, workflowShellWord{value: pathValue, quote: word.quote}, workingDirectory, workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return "", -1, workingDirectory, err
			}
			workingDirectory = directory
			seenChangeDirectory = true
			i++
		case strings.HasPrefix(value, "-"):
			return "", -1, workingDirectory, fmt.Errorf("pull-request job %q uses an unmodeled Go global flag %q", jobID, value)
		default:
			return value, i, workingDirectory, nil
		}
	}
	return "", -1, workingDirectory, nil
}

func workflowShellWords(text string) []workflowShellWord {
	text = strings.ReplaceAll(text, "\\\r\n", " ")
	text = strings.ReplaceAll(text, "\\\n", " ")
	var words []workflowShellWord
	var current strings.Builder
	var quote byte
	var wordQuote byte
	started := false
	flush := func() {
		if started {
			words = append(words, workflowShellWord{value: current.String(), quote: wordQuote})
			current.Reset()
			wordQuote = 0
			started = false
		}
	}
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if quote == '\'' {
			if ch == '\'' {
				quote = 0
			} else {
				current.WriteByte(ch)
			}
			continue
		}
		if ch == '\\' && i+1 < len(text) {
			if text[i+1] == '\n' {
				i++
				continue
			}
			if isWindowsDrivePathPrefix(current.String()) {
				current.WriteByte('/')
				started = true
				continue
			}
			current.WriteByte(text[i+1])
			started = true
			i++
			continue
		}
		if quote == '"' {
			if ch == '"' {
				quote = 0
			} else {
				current.WriteByte(ch)
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			if started {
				wordQuote = 0
			} else {
				wordQuote = ch
			}
			started = true
			quote = ch
			continue
		}
		if ch == '\n' {
			flush()
			words = append(words, workflowShellWord{value: ";", operator: true})
			continue
		}
		if ch == ' ' || ch == '\t' || ch == '\r' {
			flush()
			continue
		}
		if strings.ContainsRune(";|&()<> ", rune(ch)) {
			flush()
			operator := string(ch)
			if i+1 < len(text) && ((ch == '&' && text[i+1] == '&') || (ch == '|' && text[i+1] == '|')) {
				operator += string(text[i+1])
				i++
			}
			words = append(words, workflowShellWord{value: operator, operator: true})
			continue
		}
		current.WriteByte(ch)
		started = true
	}
	flush()
	return words
}

func isWindowsDrivePathPrefix(value string) bool {
	return len(value) >= 2 && ((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) && value[1] == ':'
}

func isWorkflowGoCommand(value string) bool {
	if value == "go" || value == "$GO" || value == "${GO}" || strings.EqualFold(value, "go.exe") {
		return true
	}
	path := strings.ToLower(strings.ReplaceAll(value, "\\", "/"))
	return strings.HasSuffix(path, "/go") || strings.HasSuffix(path, "/go.exe")
}

func checkPRGoTestArguments(jobID string, args, goFlags []workflowShellWord, workingDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node) error {
	var runSelector, skipSelector string
	hasRun, hasSkip := false, false
	deferredPackage := false
	goTestArguments := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		value := arg.value
		if value == "-args" {
			goTestArguments = true
			continue
		}
		switch {
		case value == "-run" || value == "--run" || value == "-test.run" || value == "--test.run":
			if hasRun || i+1 >= len(args) {
				return fmt.Errorf("pull-request job %q has an unreadable go test -run selector", jobID)
			}
			i++
			resolved, err := resolvePRTestSelector(args[i], workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -run selector: %w", jobID, err)
			}
			runSelector, hasRun = resolved, true
		case strings.HasPrefix(value, "-run=") || strings.HasPrefix(value, "--run=") || strings.HasPrefix(value, "-test.run=") || strings.HasPrefix(value, "--test.run="):
			if hasRun {
				return fmt.Errorf("pull-request job %q repeats go test -run", jobID)
			}
			resolved, err := resolvePRTestSelector(workflowShellWord{value: value[strings.IndexByte(value, '=')+1:], quote: arg.quote}, workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -run selector: %w", jobID, err)
			}
			runSelector, hasRun = resolved, true
		case value == "-skip" || value == "--skip" || value == "-test.skip" || value == "--test.skip":
			if hasSkip || i+1 >= len(args) {
				return fmt.Errorf("pull-request job %q has an unreadable go test -skip selector", jobID)
			}
			i++
			resolved, err := resolvePRTestSelector(args[i], workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -skip selector: %w", jobID, err)
			}
			skipSelector, hasSkip = resolved, true
		case strings.HasPrefix(value, "-skip=") || strings.HasPrefix(value, "--skip=") || strings.HasPrefix(value, "-test.skip=") || strings.HasPrefix(value, "--test.skip="):
			if hasSkip {
				return fmt.Errorf("pull-request job %q repeats go test -skip", jobID)
			}
			resolved, err := resolvePRTestSelector(workflowShellWord{value: value[strings.IndexByte(value, '=')+1:], quote: arg.quote}, workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -skip selector: %w", jobID, err)
			}
			skipSelector, hasSkip = resolved, true
		case value == "-exec" || value == "--exec" || strings.HasPrefix(value, "-exec=") || strings.HasPrefix(value, "--exec="):
			return fmt.Errorf("pull-request job %q uses a custom go test executor", jobID)
		default:
			if goTestArguments || strings.HasPrefix(value, "-") {
				continue
			}
			packagePath := value
			if hasDynamicShellReference(value) {
				resolved, err := resolvePRTestSelector(arg, workflowEnv, jobEnv, stepEnv)
				if err != nil {
					return fmt.Errorf("pull-request job %q has an unclassifiable go test package target: %w", jobID, err)
				}
				packagePath = resolved
			}
			if packageReachesPRDeferredRootFromDirectory(packagePath, workingDirectory) {
				deferredPackage = true
			}
		}
	}
	goRun, goSkip, hasGoRun, hasGoSkip, err := parsePRGoFlags(jobID, goFlags)
	if err != nil {
		return err
	}
	if !hasRun && hasGoRun {
		runSelector, hasRun = goRun, true
	}
	if !hasSkip && hasGoSkip {
		skipSelector, hasSkip = goSkip, true
	}
	if hasRun && prSelectorMatchesDeferredTests(runSelector) && (!hasSkip || !prSelectorSkipsAllDeferredTests(skipSelector)) {
		return fmt.Errorf("pull-request job %q selects deferred tests with go test -run", jobID)
	}
	if deferredPackage && (!hasRun || prSelectorMatchesDeferredTests(runSelector)) && (!hasSkip || !prSelectorSkipsAllDeferredTests(skipSelector)) {
		return fmt.Errorf("pull-request job %q runs an unfiltered deferred test root", jobID)
	}
	return nil
}

func parsePRGoFlags(jobID string, words []workflowShellWord) (string, string, bool, bool, error) {
	var runSelector, skipSelector string
	hasRun, hasSkip := false, false
	for _, word := range words {
		if word.operator {
			return "", "", false, false, fmt.Errorf("pull-request job %q has unclassifiable GOFLAGS syntax", jobID)
		}
		value := word.value
		switch {
		case value == "-run" || value == "--run" || value == "-test.run" || value == "--test.run":
			return "", "", false, false, fmt.Errorf("pull-request job %q has a GOFLAGS selector without a static value", jobID)
		case strings.HasPrefix(value, "-run=") || strings.HasPrefix(value, "--run=") || strings.HasPrefix(value, "-test.run=") || strings.HasPrefix(value, "--test.run="):
			if hasRun {
				return "", "", false, false, fmt.Errorf("pull-request job %q repeats GOFLAGS -run", jobID)
			}
			runSelector = value[strings.IndexByte(value, '=')+1:]
			if runSelector == "" || strings.Contains(runSelector, "${{") {
				return "", "", false, false, fmt.Errorf("pull-request job %q has an unclassifiable GOFLAGS -run selector", jobID)
			}
			hasRun = true
		case value == "-skip" || value == "--skip" || value == "-test.skip" || value == "--test.skip":
			return "", "", false, false, fmt.Errorf("pull-request job %q has a GOFLAGS skip selector without a static value", jobID)
		case strings.HasPrefix(value, "-skip=") || strings.HasPrefix(value, "--skip=") || strings.HasPrefix(value, "-test.skip=") || strings.HasPrefix(value, "--test.skip="):
			if hasSkip {
				return "", "", false, false, fmt.Errorf("pull-request job %q repeats GOFLAGS -skip", jobID)
			}
			skipSelector = value[strings.IndexByte(value, '=')+1:]
			if skipSelector == "" || strings.Contains(skipSelector, "${{") {
				return "", "", false, false, fmt.Errorf("pull-request job %q has an unclassifiable GOFLAGS -skip selector", jobID)
			}
			hasSkip = true
		case value == "-exec" || value == "--exec" || strings.HasPrefix(value, "-exec=") || strings.HasPrefix(value, "--exec="):
			return "", "", false, false, fmt.Errorf("pull-request job %q uses a custom GOFLAGS test executor", jobID)
		default:
			if !strings.HasPrefix(value, "-") {
				return "", "", false, false, fmt.Errorf("pull-request job %q has an unclassifiable GOFLAGS entry", jobID)
			}
		}
	}
	return runSelector, skipSelector, hasRun, hasSkip, nil
}

func resolvePRTestSelector(word workflowShellWord, workflowEnv, jobEnv, stepEnv *yaml.Node) (string, error) {
	selector := strings.TrimSpace(word.value)
	if word.quote == '\'' {
		return selector, nil
	}
	if selector == "" || strings.Contains(selector, "${{") || strings.ContainsRune(selector, '`') {
		return "", fmt.Errorf("selector is empty or dynamic")
	}
	if match := workflowShellVariable.FindStringSubmatch(selector); len(match) == 3 {
		name := match[1]
		if name == "" {
			name = match[2]
		}
		value, ok := workflowEnvironmentValue(name, workflowEnv, jobEnv, stepEnv)
		if !ok {
			return "", fmt.Errorf("shell variable %q is not statically defined", name)
		}
		return value, nil
	}
	if hasDynamicShellReference(selector) {
		return "", fmt.Errorf("selector contains an unresolved shell expansion")
	}
	return selector, nil
}

func hasDynamicShellReference(value string) bool {
	if value == "$" {
		return true
	}
	for i := 0; i < len(value); i++ {
		if value[i] != '$' || i+1 >= len(value) {
			continue
		}
		next := value[i+1]
		if next == '{' || next == '(' || next == '?' || next == '!' || next == '$' || next == '-' || next >= '0' && next <= '9' || next == '_' || next >= 'A' && next <= 'Z' || next >= 'a' && next <= 'z' {
			return true
		}
		if next != '|' && next != ')' && next != '/' && next != ']' {
			return true
		}
	}
	return false
}

func workflowEnvironmentValue(name string, levels ...*yaml.Node) (string, bool) {
	value, present, readable := workflowEnvironmentLookup(name, levels...)
	return value, present && readable
}

func workflowEnvironmentLookup(name string, levels ...*yaml.Node) (string, bool, bool) {
	var value *yaml.Node
	for _, level := range levels {
		level = resolveWorkflowAlias(level)
		if current := workflowMappingValue(level, name); current != nil {
			value = current
		}
	}
	if value == nil {
		return "", false, true
	}
	value = resolveWorkflowAlias(value)
	if value == nil || value.Kind != yaml.ScalarNode || value.Tag != "!!str" || strings.Contains(value.Value, "${{") {
		return "", true, false
	}
	return value.Value, true, true
}

func packageReachesPRDeferredRootFromDirectory(value, workingDirectory string) bool {
	if value == "..." || value == "./..." {
		return prPackageTreeContainsDeferredRoot(workingDirectory)
	}
	if strings.HasSuffix(value, "/...") {
		if !strings.HasPrefix(value, "./") && !strings.HasPrefix(value, "../") {
			return packageReachesPRDeferredImport(value)
		}
		base := strings.TrimPrefix(strings.TrimSuffix(value, "/..."), "./")
		return prPackageTreeContainsDeferredRoot(path.Clean(path.Join(workingDirectory, base)))
	}
	if value == "." || value == ".." || strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") {
		target := path.Clean(path.Join(workingDirectory, value))
		return target == "cmd/curator" || strings.HasPrefix(target, "cmd/curator/")
	}
	return packageReachesPRDeferredImport(value)
}

func prPackageTreeContainsDeferredRoot(directory string) bool {
	return directory == "." || directory == "cmd/curator" ||
		strings.HasPrefix("cmd/curator", directory+"/") || strings.HasPrefix(directory, "cmd/curator/")
}

func packageReachesPRDeferredImport(value string) bool {
	return value == "cmd/curator" || value == "./cmd/curator" ||
		value == "github.com/relux-works/curator/cmd/curator" || strings.HasSuffix(value, "/cmd/curator") ||
		strings.HasPrefix(value, "./cmd/curator/") || strings.HasPrefix(value, "cmd/curator/") ||
		strings.HasSuffix(value, "/cmd/curator/...")
}

func prSelectorMatchesDeferredTests(selector string) bool {
	patterns, ok := prCompileTestNamePattern(selector)
	if !ok {
		return true
	}
	for _, name := range curatormutations.PostLandingNames() {
		matched, _ := prTestNamePatternMatches(patterns, name)
		if matched {
			return true
		}
	}
	return false
}

func prSelectorSkipsAllDeferredTests(selector string) bool {
	patterns, ok := prCompileTestNamePattern(selector)
	if !ok {
		return false
	}
	for _, name := range curatormutations.PostLandingNames() {
		matched, partial := prTestNamePatternMatches(patterns, name)
		if !matched || partial {
			return false
		}
	}
	return true
}

func prCompileTestNamePattern(selector string) ([][]*regexp.Regexp, bool) {
	parts := prSplitTestNamePattern(selector)
	compiled := make([][]*regexp.Regexp, len(parts))
	for i, alternative := range parts {
		compiled[i] = make([]*regexp.Regexp, len(alternative))
		for j, component := range alternative {
			pattern, err := regexp.Compile(component)
			if err != nil {
				return nil, false
			}
			compiled[i][j] = pattern
		}
	}
	return compiled, len(compiled) != 0
}

func prSplitTestNamePattern(selector string) [][]string {
	var alternatives [][]string
	var components []string
	start := 0
	parenDepth := 0
	charClassDepth := 0
	for i := 0; i < len(selector); i++ {
		switch selector[i] {
		case '[':
			charClassDepth++
		case ']':
			charClassDepth--
			if charClassDepth < 0 {
				charClassDepth = 0
			}
		case '(':
			if charClassDepth == 0 {
				parenDepth++
			}
		case ')':
			if charClassDepth == 0 {
				parenDepth--
			}
		case '\\':
			i++
		case '/':
			if charClassDepth == 0 && parenDepth == 0 {
				components = append(components, selector[start:i])
				start = i + 1
			}
		case '|':
			if charClassDepth == 0 && parenDepth == 0 {
				components = append(components, selector[start:i])
				alternatives = append(alternatives, components)
				components = nil
				start = i + 1
			}
		}
	}
	components = append(components, selector[start:])
	return append(alternatives, components)
}

func prTestNamePatternMatches(patterns [][]*regexp.Regexp, name string) (bool, bool) {
	nameParts := strings.Split(name, "/")
	for _, alternative := range patterns {
		matched := true
		for i, namePart := range nameParts {
			if i >= len(alternative) {
				break
			}
			if !alternative[i].MatchString(namePart) {
				matched = false
				break
			}
		}
		if matched {
			return true, len(nameParts) < len(alternative)
		}
	}
	return false, false
}

func checkPRLocalTargets(jobID string, job pullRequestJob, workflowEnv *yaml.Node) error {
	for _, step := range job.Steps {
		stepDirectory, err := prStepWorkingDirectory(jobID, step, workflowEnv, &job.Env, &step.Env)
		if err != nil {
			return err
		}
		if err := checkPRLocalExecutableCommands(jobID, step, stepDirectory, workflowEnv, &job.Env, &step.Env); err != nil {
			return err
		}
		if err := checkPRWorkflowInterpreters(jobID, step, stepDirectory, workflowEnv, &job.Env, &step.Env); err != nil {
			return err
		}
		words := workflowShellWords(step.Run)
		for i := 0; i < len(words); i++ {
			if words[i].operator || !isWorkflowGoCommand(words[i].value) {
				continue
			}
			command, commandIndex, workingDirectory, err := parsePRGoCommand(jobID, words, i, stepDirectory, workflowEnv, &job.Env, &step.Env)
			if err != nil {
				return err
			}
			if command != "run" {
				continue
			}
			for j := commandIndex + 1; j < len(words) && !words[j].operator; j++ {
				candidate := words[j].value
				if strings.HasPrefix(candidate, "-") {
					continue
				}
				if strings.HasPrefix(candidate, "$") {
					if !isAllowedPRDynamicGoRun(step) {
						return fmt.Errorf("pull-request job %q uses a dynamic local go run target", jobID)
					}
					break
				}
				if candidate == "." || strings.HasPrefix(candidate, "./") || strings.HasPrefix(candidate, "../") || strings.HasSuffix(candidate, ".go") {
					target, err := resolvePRRepoPath(jobID, workflowShellWord{value: candidate}, workingDirectory, "local Go tool target", workflowEnv, &job.Env, &step.Env)
					if err != nil {
						return err
					}
					allowlistPath := "./" + strings.TrimPrefix(target, "./")
					if _, ok := allowedPRLocalGoTools[allowlistPath]; !ok {
						return fmt.Errorf("pull-request job %q calls an unapproved local Go tool %q", jobID, allowlistPath)
					}
					break
				}
			}
		}
		uses := strings.TrimSpace(step.Uses)
		if isPRLocalPath(uses) {
			allowedInputs, ok := allowedPRLocalActions[uses]
			actualInputs, readable := workflowStringMap(&step.With)
			if !ok || !readable || !reflect.DeepEqual(actualInputs, allowedInputs) {
				return fmt.Errorf("pull-request job %q calls an unapproved local action %q or changes its inputs", jobID, uses)
			}
		}
		if dynamicWorkflowScriptInvocation.MatchString(step.Run) {
			return fmt.Errorf("pull-request job %q calls a local script through a dynamic path", jobID)
		}
		if dynamicWorkflowSourceInvocation.MatchString(step.Run) {
			return fmt.Errorf("pull-request job %q sources a local script through a dynamic path", jobID)
		}
		if inlineWorkflowScriptBody.MatchString(step.Run) {
			return fmt.Errorf("pull-request job %q embeds an unclassifiable shell script body", jobID)
		}
		if inlineWorkflowHeredoc.MatchString(step.Run) && !isAllowedPRInlineHeredoc(step) {
			return fmt.Errorf("pull-request job %q embeds an unapproved interpreter heredoc", jobID)
		}
		for _, script := range prLocalScripts(step.Run) {
			target, err := resolvePRRepoPath(jobID, workflowShellWord{value: script}, stepDirectory, "local script target", workflowEnv, &job.Env, &step.Env)
			if err != nil {
				return err
			}
			if _, ok := allowedPRLocalScripts[target]; !ok {
				return fmt.Errorf("pull-request job %q calls an unapproved local script %q", jobID, target)
			}
			if target == ".github/ci/test-gate.sh" {
				value, ok := workflowEnvironmentValue("CI_SKIP_POST_LANDING_TESTS", workflowEnv, &job.Env, &step.Env)
				if !ok || value != "1" {
					return fmt.Errorf("pull-request job %q calls test-gate.sh without the deferred-suite skip", jobID)
				}
				if testGateSkipUnset.MatchString(step.Run) || isolatedWorkflowEnvironment.MatchString(step.Run) {
					return fmt.Errorf("pull-request job %q removes the test-gate deferred-suite skip", jobID)
				}
				for _, match := range testGateSkipOverride.FindAllStringSubmatch(step.Run, -1) {
					if strings.Trim(match[1], "\"'") != "1" {
						return fmt.Errorf("pull-request job %q overrides the test-gate deferred-suite skip", jobID)
					}
				}
			}
		}
	}
	return nil
}

func checkPRLocalExecutableCommands(jobID string, step pullRequestStep, stepDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node) error {
	words := workflowShellWords(step.Run)
	for _, index := range prWorkflowCommandIndices(words) {
		candidate := words[index]
		if path.IsAbs(candidate.value) {
			if _, ok := allowedPRAbsoluteGoCommands[candidate.value]; ok {
				continue
			}
			if isAllowedPRAbsoluteCommandRun(jobID, step, candidate.value) {
				continue
			}
		}
		if !path.IsAbs(candidate.value) && !strings.Contains(candidate.value, "/") {
			continue
		}
		target, err := resolvePRRepoPath(jobID, candidate, stepDirectory, "local executable target", workflowEnv, jobEnv, stepEnv)
		if err != nil {
			return err
		}
		if _, ok := allowedPRLocalExecutables[target]; !ok {
			return fmt.Errorf("pull-request job %q directly executes an unapproved local file %q", jobID, target)
		}
	}
	return nil
}

func isAllowedPRAbsoluteCommandRun(jobID string, step pullRequestStep, command string) bool {
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(step.Run)))
	_, ok := allowedPRAbsoluteCommandRuns[jobID+"\x00"+step.Name+"\x00"+command+"\x00"+digest]
	return ok
}

func checkPRWorkflowInterpreters(jobID string, step pullRequestStep, stepDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node) error {
	if inlineWorkflowHeredoc.MatchString(step.Run) {
		if !isAllowedPRInlineHeredoc(step) {
			return fmt.Errorf("pull-request job %q embeds an unapproved interpreter heredoc", jobID)
		}
		return nil
	}
	if err := checkPRWorkflowNestedExecution(jobID, step, stepDirectory, workflowEnv, jobEnv, stepEnv, step.Run, 0); err != nil {
		return err
	}
	unmodeledShell := prEnvSplitString.MatchString(step.Run) || prShellFunctionDefinition.MatchString(step.Run)
	words := workflowShellWords(step.Run)
	for _, index := range prWorkflowCommandIndices(words) {
		command := strings.ToLower(path.Base(strings.ReplaceAll(words[index].value, "\\", "/")))
		if command == "." || command == "source" {
			return fmt.Errorf("pull-request job %q uses an unapproved source invocation", jobID)
		}
		switch command {
		case "alias", "eval", "function", "shopt", "trap":
			unmodeledShell = true
		}
	}
	if unmodeledShell && !isAllowedPRUnmodeledShellRun(jobID, step) {
		return fmt.Errorf("pull-request job %q uses an unmodeled shell command form", jobID)
	}
	for _, index := range prWorkflowInterpreterCommandIndices(words) {
		if err := checkPRInterpreterInvocation(jobID, words, index, stepDirectory, workflowEnv, jobEnv, stepEnv); err != nil {
			return err
		}
	}
	return nil
}

// Shell substitutions are inspected recursively. Process substitutions,
// execution dispatchers and find execution actions are refused because their
// payloads are outside the direct command model.
func checkPRWorkflowNestedExecution(jobID string, step pullRequestStep, stepDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node, run string, depth int) error {
	if depth > 12 {
		return fmt.Errorf("pull-request job %q has shell command substitutions nested too deeply", jobID)
	}
	words := workflowShellWords(run)
	for _, index := range prWorkflowCommandIndices(words) {
		command := strings.ToLower(path.Base(strings.ReplaceAll(words[index].value, "\\", "/")))
		switch command {
		case "find":
			for i := index + 1; i < len(words) && !words[i].operator; i++ {
				switch words[i].value {
				case "-exec", "-execdir", "-ok", "-okdir":
					return fmt.Errorf("pull-request job %q uses an unmodeled find execution action", jobID)
				}
			}
		case "parallel", "sem", "make", "gmake", "ninja", "coproc":
			return fmt.Errorf("pull-request job %q uses an unmodeled execution dispatcher %q", jobID, command)
		}
	}

	bodies, complete := prWorkflowCommandSubstitutionBodies(run)
	if !complete {
		return fmt.Errorf("pull-request job %q has an unclassifiable shell command substitution", jobID)
	}
	for _, body := range bodies {
		nestedStep := step
		nestedStep.Run = body
		nestedWords := workflowShellWords(body)
		if err := checkPRLocalExecutableCommands(jobID, nestedStep, stepDirectory, workflowEnv, jobEnv, stepEnv); err != nil {
			return err
		}
		for _, index := range prWorkflowCommandIndices(nestedWords) {
			command := strings.ToLower(path.Base(strings.ReplaceAll(nestedWords[index].value, "\\", "/")))
			switch command {
			case ".", "source", "alias", "eval", "function", "shopt", "trap":
				return fmt.Errorf("pull-request job %q uses an unmodeled command inside a shell substitution", jobID)
			}
			if prWorkflowInterpreter(nestedWords[index].value) != "" {
				return fmt.Errorf("pull-request job %q invokes an interpreter inside a shell substitution", jobID)
			}
			if !isWorkflowGoCommand(nestedWords[index].value) {
				continue
			}
			goCommand, _, _, err := parsePRGoCommand(jobID, nestedWords, index, stepDirectory, workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return err
			}
			if goCommand == "test" || goCommand == "run" {
				return fmt.Errorf("pull-request job %q runs an unmodeled Go test or tool inside a shell substitution", jobID)
			}
		}
		if err := checkPRWorkflowNestedExecution(jobID, nestedStep, stepDirectory, workflowEnv, jobEnv, stepEnv, body, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func prWorkflowCommandSubstitutionBodies(run string) ([]string, bool) {
	var bodies []string
	var quote byte
	for i := 0; i < len(run); i++ {
		ch := run[i]
		if quote == '\'' {
			if ch == '\'' {
				quote = 0
			}
			continue
		}
		if ch == '\\' && quote != '\'' {
			i++
			continue
		}
		if quote == '"' {
			if ch == '"' {
				quote = 0
				continue
			}
		} else {
			if ch == '\'' || ch == '"' {
				quote = ch
				continue
			}
			if ch == '#' && (i == 0 || strings.ContainsRune(" \t\r\n;&|()", rune(run[i-1]))) {
				for i < len(run) && run[i] != '\n' {
					i++
				}
				continue
			}
		}
		if ch == '$' && i+1 < len(run) && run[i+1] == '(' {
			end, ok := prWorkflowCommandSubstitutionEnd(run, i+1)
			if !ok {
				return nil, false
			}
			bodies = append(bodies, run[i+2:end])
			i = end
			continue
		}
		if quote == 0 && (ch == '<' || ch == '>') && i+1 < len(run) && run[i+1] == '(' {
			return nil, false
		}
		if ch == '`' {
			end := i + 1
			for end < len(run) {
				if run[end] == '\\' {
					end += 2
					continue
				}
				if run[end] == '`' {
					break
				}
				end++
			}
			if end >= len(run) {
				return nil, false
			}
			bodies = append(bodies, run[i+1:end])
			i = end
		}
	}
	return bodies, true
}

func prWorkflowCommandSubstitutionEnd(run string, open int) (int, bool) {
	depth := 1
	var quote byte
	for i := open + 1; i < len(run); i++ {
		ch := run[i]
		if quote == '\'' {
			if ch == '\'' {
				quote = 0
			}
			continue
		}
		if ch == '\\' && quote != '\'' {
			i++
			continue
		}
		if quote == '"' {
			if ch == '"' {
				quote = 0
			}
			if ch == '$' && i+1 < len(run) && run[i+1] == '(' {
				end, ok := prWorkflowCommandSubstitutionEnd(run, i+1)
				if !ok {
					return 0, false
				}
				i = end
			}
			continue
		}
		if ch == '$' && i+1 < len(run) && run[i+1] == '(' {
			end, ok := prWorkflowCommandSubstitutionEnd(run, i+1)
			if !ok {
				return 0, false
			}
			i = end
			continue
		}
		if (ch == '<' || ch == '>') && i+1 < len(run) && run[i+1] == '(' {
			return 0, false
		}
		if ch == '`' {
			for i++; i < len(run); i++ {
				if run[i] == '\\' {
					i++
					continue
				}
				if run[i] == '`' {
					break
				}
			}
			if i >= len(run) {
				return 0, false
			}
			continue
		}
		switch ch {
		case '\'', '"':
			quote = ch
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func isAllowedPRUnmodeledShellRun(jobID string, step pullRequestStep) bool {
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(step.Run)))
	_, ok := allowedPRUnmodeledShellRuns[jobID+"\x00"+step.Name+"\x00"+digest]
	return ok
}

func checkPRInterpreterInvocation(jobID string, words []workflowShellWord, index int, stepDirectory string, workflowEnv, jobEnv, stepEnv *yaml.Node) error {
	interpreter := prWorkflowInterpreter(words[index].value)
	end := index + 1
	for end < len(words) && !words[end].operator {
		end++
	}
	args := words[index+1 : end]
	switch interpreter {
	case "bash":
		if len(args) == 0 || strings.HasPrefix(args[0].value, "-") || strings.Contains(args[0].value, "$") {
			return fmt.Errorf("pull-request job %q uses an unapproved bash invocation form", jobID)
		}
		target, err := resolvePRRepoPath(jobID, args[0], stepDirectory, "local script target", workflowEnv, jobEnv, stepEnv)
		if err != nil {
			return err
		}
		if _, ok := allowedPRLocalScripts[target]; !ok || !strings.HasSuffix(target, ".sh") {
			return fmt.Errorf("pull-request job %q calls an unapproved bash script %q", jobID, target)
		}
		if args[0].value != target {
			return fmt.Errorf("pull-request job %q uses an unapproved bash script path form %q", jobID, args[0].value)
		}
		if target == ".github/ci/isolated-test.sh" {
			nested := prWorkflowInterpreterCommandIndices(args[1:])
			for _, nestedIndex := range nested {
				if err := checkPRInterpreterInvocation(jobID, args[1:], nestedIndex, stepDirectory, workflowEnv, jobEnv, stepEnv); err != nil {
					return err
				}
			}
		}
	case "python3":
		if len(args) == 0 || strings.HasPrefix(args[0].value, "-") || strings.Contains(args[0].value, "$") {
			return fmt.Errorf("pull-request job %q uses an unapproved python3 invocation form", jobID)
		}
		target, err := resolvePRRepoPath(jobID, args[0], stepDirectory, "local script target", workflowEnv, jobEnv, stepEnv)
		if err != nil {
			return err
		}
		if _, ok := allowedPRLocalScripts[target]; !ok || !strings.HasSuffix(target, ".py") {
			return fmt.Errorf("pull-request job %q calls an unapproved python3 script %q", jobID, target)
		}
		if args[0].value != target {
			return fmt.Errorf("pull-request job %q uses an unapproved python3 script path form %q", jobID, args[0].value)
		}
	default:
		return fmt.Errorf("pull-request job %q uses an unapproved interpreter %q", jobID, words[index].value)
	}
	return nil
}

func prWorkflowCommandIndices(words []workflowShellWord) []int {
	var indices []int
	commandStart := true
	for i := 0; i < len(words); i++ {
		word := words[i]
		if word.operator {
			if word.value == ">" || word.value == "<" {
				if i+1 < len(words) && !words[i+1].operator {
					i++
				}
				continue
			}
			commandStart = true
			continue
		}
		if word.quote == 0 && strings.HasPrefix(word.value, "#") {
			for i+1 < len(words) && (!words[i+1].operator || words[i+1].value != ";") {
				i++
			}
			commandStart = true
			continue
		}
		if word.quote == 0 && isPRShellControlKeyword(word.value) {
			commandStart = true
			continue
		}
		if !commandStart {
			continue
		}
		commandIndex := prWorkflowCommandTarget(words, i)
		if commandIndex >= 0 {
			indices = append(indices, commandIndex)
		}
		commandStart = false
	}
	return indices
}

func prWorkflowInterpreterCommandIndices(words []workflowShellWord) []int {
	var indices []int
	for _, index := range prWorkflowCommandIndices(words) {
		if prWorkflowInterpreter(words[index].value) != "" {
			indices = append(indices, index)
		}
	}
	return indices
}

func prWorkflowCommandTarget(words []workflowShellWord, start int) int {
	index := start
	for index < len(words) && !words[index].operator && workflowShellAssignment.MatchString(words[index].value) {
		index++
	}
	for index < len(words) && !words[index].operator {
		command := strings.ToLower(path.Base(strings.ReplaceAll(words[index].value, "\\", "/")))
		switch command {
		case "env":
			index++
			for index < len(words) && !words[index].operator {
				value := words[index].value
				if value == "--" {
					index++
					break
				}
				if value == "-u" || value == "--unset" || value == "-C" || value == "--chdir" {
					index += 2
					continue
				}
				if strings.HasPrefix(value, "-") || workflowShellAssignment.MatchString(value) {
					index++
					continue
				}
				break
			}
		case "command", "builtin", "exec", "nohup", "nice", "time":
			index++
			for index < len(words) && !words[index].operator && strings.HasPrefix(words[index].value, "-") {
				index++
			}
		case "sudo":
			index++
			for index < len(words) && !words[index].operator {
				value := words[index].value
				if value == "-u" || value == "-g" || value == "-h" || value == "-p" || value == "-C" || value == "-D" {
					index += 2
					continue
				}
				if strings.HasPrefix(value, "-") || workflowShellAssignment.MatchString(value) {
					index++
					continue
				}
				break
			}
		case "timeout":
			index++
			for index < len(words) && !words[index].operator && strings.HasPrefix(words[index].value, "-") {
				value := words[index].value
				index++
				if value == "-k" || value == "--kill-after" {
					index++
				}
			}
			if index < len(words) && !words[index].operator {
				index++
			}
		case "xargs":
			index++
			for index < len(words) && !words[index].operator && strings.HasPrefix(words[index].value, "-") {
				value := words[index].value
				index++
				if value == "-a" || value == "-d" || value == "-E" || value == "-I" || value == "-L" || value == "-n" || value == "-P" || value == "-s" {
					index++
				}
			}
		default:
			if index < len(words) && !words[index].operator && prWorkflowInterpreter(words[index].value) == "" && workflowShellAssignment.MatchString(words[index].value) {
				index++
				continue
			}
			if index >= len(words) || words[index].operator {
				return -1
			}
			return index
		}
	}
	return -1
}

func prWorkflowInterpreter(command string) string {
	base := strings.ToLower(path.Base(strings.ReplaceAll(command, "\\", "/")))
	if base == "python" || base == "py" {
		return base
	}
	if strings.HasSuffix(base, ".exe") {
		if prWorkflowInterpreter(strings.TrimSuffix(base, ".exe")) != "" {
			return base
		}
		return ""
	}
	if base == "python3" {
		return "python3"
	}
	if base == "bash" {
		return "bash"
	}
	if prPythonInterpreterVersion.MatchString(base) || prNodeInterpreterVersion.MatchString(base) || prPerlInterpreterVersion.MatchString(base) || prPHPInterpreterVersion.MatchString(base) || prVersionedShellInterpreter.MatchString(base) {
		return base
	}
	if prRubyInterpreterVersion.MatchString(base) || base == "ruby" || base == "jruby" || base == "truffleruby" {
		return "ruby"
	}
	switch base {
	case "sh", "dash", "ash", "zsh", "ksh", "mksh", "yash", "fish", "csh", "tcsh",
		"pypy", "pypy3", "perl", "perl5", "node", "nodejs", "deno", "bun", "bunx", "tsx", "ts-node", "npx",
		"js", "jsc", "qjs", "quickjs", "pwsh", "powershell", "php", "lua", "luajit", "r", "rscript", "julia",
		"awk", "gawk", "mawk", "nawk", "tclsh", "wish", "java", "jshell", "groovy", "scala", "swift", "csi",
		"dotnet-script", "scriptcs", "cmd", "cmd.exe", "command.com", "cscript", "wscript", "ipython", "ipython3",
		"busybox", "toybox", "pythonw", "pythonw3",
		"irb", "pry", "elixir", "iex", "erl", "escript", "ghci", "runghc", "ocaml", "ocamlrun", "osascript":
		return base
	default:
		return ""
	}
}

func isAllowedPRDynamicGoRun(step pullRequestStep) bool {
	return step.Name == "Record macOS race capacity" && step.If == "runner.os == 'macOS'" &&
		fmt.Sprintf("%x", sha256.Sum256([]byte(step.Run))) == "6132d13dc9aea819d1cd3aac9107a2742c6e83ea9e87807709fe545d91ab8780"
}

func isAllowedPRInlineHeredoc(step pullRequestStep) bool {
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(step.Run)))
	switch step.Name {
	case "Driver qualification and reproducibility":
		return digest == "34dd145d07394a40b8159c102e5665a1363050b74e6623a802a8a57eb92ff0d8" ||
			digest == "dfeb880111ffb7b4a3b48aa2adb757ae2d344042a4a0cc13564b5c66522d1502"
	case "Compare thirty isolated counts at each revision":
		return digest == "34e2488b9fba5fbb5a0d64c4ca1fac95a431be59499ba12b02271c4ae08d9fd1"
	default:
		return false
	}
}

func isPRLocalPath(value string) bool {
	return strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") || strings.HasPrefix(value, ".github/")
}

func prLocalScripts(run string) []string {
	found := make(map[string]struct{})
	for _, pattern := range []*regexp.Regexp{localWorkflowScriptInvocation, localWorkflowScriptToken, localWorkflowExecutable} {
		for _, match := range pattern.FindAllStringSubmatch(run, -1) {
			if len(match) < 2 {
				continue
			}
			value := strings.Trim(match[1], "\"'")
			value = strings.TrimPrefix(value, "./")
			found[value] = struct{}{}
		}
	}
	paths := make([]string, 0, len(found))
	for value := range found {
		paths = append(paths, value)
	}
	sort.Strings(paths)
	return paths
}

func workflowStringMap(node *yaml.Node) (map[string]string, bool) {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind == 0 || node.Tag == "!!null" {
		return map[string]string{}, true
	}
	if node.Kind != yaml.MappingNode {
		return nil, false
	}
	values := make(map[string]string, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		value := resolveWorkflowAlias(node.Content[i+1])
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || value == nil || value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
			return nil, false
		}
		values[key.Value] = value.Value
	}
	return values, true
}

func pullRequestMatrixValues(strategy *yaml.Node) (map[string][]string, error) {
	values := make(map[string][]string)
	strategy = resolveWorkflowAlias(strategy)
	if strategy == nil || strategy.Kind == 0 || strategy.Tag == "!!null" {
		return values, nil
	}
	if strategy.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("strategy is not a mapping")
	}
	matrix := resolveWorkflowAlias(workflowMappingValue(strategy, "matrix"))
	if matrix == nil || matrix.Kind == 0 || matrix.Tag == "!!null" {
		return values, nil
	}
	if matrix.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("matrix is not a static mapping")
	}
	seen := make(map[string]struct{}, len(matrix.Content)/2)
	for i := 0; i+1 < len(matrix.Content); i += 2 {
		key := matrix.Content[i]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
			return nil, fmt.Errorf("matrix has a non-string key")
		}
		if _, ok := seen[key.Value]; ok {
			return nil, fmt.Errorf("matrix repeats key %q", key.Value)
		}
		seen[key.Value] = struct{}{}
		switch key.Value {
		case "include":
			if err := collectPRMatrixRows(matrix.Content[i+1], values, true); err != nil {
				return nil, fmt.Errorf("matrix include: %w", err)
			}
		case "exclude":
			if err := collectPRMatrixRows(matrix.Content[i+1], values, false); err != nil {
				return nil, fmt.Errorf("matrix exclude: %w", err)
			}
		default:
			items, err := prMatrixDimensionValues(matrix.Content[i+1])
			if err != nil {
				return nil, fmt.Errorf("matrix field %q: %w", key.Value, err)
			}
			values[key.Value] = append(values[key.Value], items...)
		}
	}
	return values, nil
}

func collectPRMatrixRows(node *yaml.Node, values map[string][]string, include bool) error {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind != yaml.SequenceNode {
		return fmt.Errorf("rows are not a sequence")
	}
	for _, row := range node.Content {
		row = resolveWorkflowAlias(row)
		if row == nil || row.Kind != yaml.MappingNode {
			return fmt.Errorf("row is not a mapping")
		}
		seen := make(map[string]struct{}, len(row.Content)/2)
		for i := 0; i+1 < len(row.Content); i += 2 {
			key := row.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return fmt.Errorf("row has a non-string key")
			}
			if _, ok := seen[key.Value]; ok {
				return fmt.Errorf("row repeats key %q", key.Value)
			}
			seen[key.Value] = struct{}{}
			value, err := prMatrixScalar(row.Content[i+1])
			if err != nil {
				return fmt.Errorf("field %q: %w", key.Value, err)
			}
			if include {
				values[key.Value] = append(values[key.Value], value)
			}
		}
	}
	return nil
}

func prMatrixDimensionValues(node *yaml.Node) ([]string, error) {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("dimension is not a sequence")
	}
	values := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		value, err := prMatrixScalar(item)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("dimension is empty")
	}
	return values, nil
}

func prMatrixScalar(node *yaml.Node) (string, error) {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind != yaml.ScalarNode || (node.Tag != "!!str" && node.Tag != "!!int" && node.Tag != "!!bool") {
		return "", fmt.Errorf("value is not a static string, integer, or boolean")
	}
	if strings.Contains(node.Value, "${{") || strings.Contains(node.Value, "}}") {
		return "", fmt.Errorf("value contains an unclassifiable dynamic expression")
	}
	return node.Value, nil
}

func validatePRMatrixField(field string, values map[string][]string) error {
	allowed, ok := allowedPRMatrixValues[field]
	if !ok {
		return fmt.Errorf("matrix field %q is not allowlisted", field)
	}
	fieldValues, ok := values[field]
	if !ok || len(fieldValues) == 0 {
		return fmt.Errorf("matrix field %q has no classifiable values", field)
	}
	for _, value := range fieldValues {
		if isDeferredSuiteMatrixValue(value) {
			return fmt.Errorf("matrix field %q carries deferred-suite selector or command %q", field, value)
		}
		if _, ok := allowed[value]; !ok {
			return fmt.Errorf("matrix field %q has unapproved value %q", field, value)
		}
	}
	return nil
}

func isDeferredSuiteMatrixValue(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, legacyMutationJobID) ||
		strings.Contains(value, "tools/mutationinventory") ||
		strings.Contains(value, "mutationinventory") ||
		hasMutationSuiteLabel(value) ||
		strings.Contains(value, "exhaustive") ||
		strings.Contains(value, "post-landing")
}

func workflowStringValue(mapping *yaml.Node, key string) (string, bool) {
	mapping = resolveWorkflowAlias(mapping)
	value := resolveWorkflowAlias(workflowMappingValue(mapping, key))
	if value == nil || value.Kind != yaml.ScalarNode || value.Tag != "!!str" || strings.Contains(value.Value, "${{") {
		return "", false
	}
	return value.Value, true
}

func canRunOnPullRequest(condition string) bool {
	condition = strings.TrimSpace(condition)
	if strings.HasPrefix(condition, "${{") && strings.HasSuffix(condition, "}}") {
		condition = strings.TrimSpace(condition[3 : len(condition)-2])
	}
	condition = strings.Join(strings.Fields(condition), " ")
	switch condition {
	case "vars.ROSE_AIR_RUNNER == 'true' && github.event_name == 'push' && github.ref == 'refs/heads/main'":
		return false
	case "github.event_name == 'workflow_dispatch' && (inputs.candidate_ref != '' || inputs.candidate_root != '')":
		return false
	case "github.event_name == 'workflow_dispatch' && inputs.fixture_isolation_diagnostics":
		return false
	default:
		// Unknown conditions are included so a new or changed condition cannot
		// silently remove a job from the pull-request scan.
		return true
	}
}

func isMutationSuiteJob(id string, job pullRequestJob) bool {
	identity := strings.ToLower(id + " " + job.Name)
	if strings.Contains(identity, legacyMutationJobID) || hasMutationSuiteLabel(identity) || strings.Contains(strings.ToLower(job.Uses), "mutation") {
		return true
	}
	for _, step := range job.Steps {
		run := strings.ToLower(step.Run)
		if hasMutationSuiteLabel(strings.ToLower(step.Name)) || hasMutationSuiteLabel(run) || strings.Contains(run, "tools/mutationinventory") {
			return true
		}
	}
	return false
}

func hasMutationSuiteLabel(identity string) bool {
	if !strings.Contains(identity, "mutation") {
		return false
	}
	return strings.Contains(identity, "suite") || strings.Contains(identity, "control") || strings.Contains(identity, "inventory")
}

func pullRequestNeedIDs(needs *yaml.Node) ([]string, error) {
	if needs == nil || needs.Kind == 0 || needs.Tag == "!!null" {
		return nil, nil
	}
	if needs.Kind == yaml.AliasNode && needs.Alias != nil {
		needs = needs.Alias
	}
	if needs.Kind == yaml.ScalarNode {
		return []string{needs.Value}, nil
	}
	if needs.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected scalar or sequence, got %s", needs.ShortTag())
	}
	ids := make([]string, 0, len(needs.Content))
	for _, dependency := range needs.Content {
		if dependency.Kind != yaml.ScalarNode || dependency.Tag != "!!str" {
			return nil, fmt.Errorf("dependency is not a string scalar")
		}
		ids = append(ids, dependency.Value)
	}
	return ids, nil
}

func isMutationJobID(id string, mutationJobIDs map[string]struct{}) bool {
	_, ok := mutationJobIDs[strings.ToLower(id)]
	return ok
}

func hasMutationResultReference(text string, mutationJobIDs map[string]struct{}) bool {
	if dynamicNeedsResultReference.MatchString(text) {
		// The current PR workflow has no dynamic needs.result selectors. Their
		// target cannot be proven non-mutation from the workflow text alone.
		return true
	}
	for _, match := range mutationJobResultReference.FindAllStringSubmatch(text, -1) {
		id := match[1]
		if id == "" {
			id = match[2]
		}
		if isMutationJobID(id, mutationJobIDs) {
			return true
		}
	}
	return false
}

func hasMutationResultReferenceInNode(node *yaml.Node, mutationJobIDs map[string]struct{}) bool {
	for _, value := range workflowScalarValues(node) {
		if hasMutationResultReference(value, mutationJobIDs) {
			return true
		}
	}
	return false
}

func workflowScalarValues(node *yaml.Node) []string {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.AliasNode && node.Alias != nil {
		return workflowScalarValues(node.Alias)
	}
	if node.Kind == yaml.ScalarNode {
		return []string{node.Value}
	}
	var values []string
	for _, child := range node.Content {
		values = append(values, workflowScalarValues(child)...)
	}
	return values
}

type workflowStepAddition struct {
	Name             string
	Run              string
	Uses             string
	WorkingDirectory string
	With             map[string]string
	Env              map[string]string
}

func addStepToWorkflowJob(data []byte, jobName string, step workflowStepAddition) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	jobs := workflowMappingValue(root, "jobs")
	job := workflowMappingValue(jobs, jobName)
	if job == nil {
		return nil, fmt.Errorf("workflow job %q is missing", jobName)
	}
	steps := workflowMappingValue(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("workflow job %q has no steps sequence", jobName)
	}
	steps.Content = append(steps.Content, workflowStepNode(step))
	return yaml.Marshal(document)
}

func addWorkflowJob(data []byte, jobID, name, run, uses, condition string) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	jobs := workflowMappingValue(root, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("workflow has no jobs mapping")
	}
	if workflowMappingValue(jobs, jobID) != nil {
		return nil, fmt.Errorf("workflow job %q already exists", jobID)
	}
	job := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendWorkflowMappingPair(job, "name", workflowStringNode(name))
	if condition != "" {
		appendWorkflowMappingPair(job, "if", workflowStringNode(condition))
	}
	if uses != "" {
		appendWorkflowMappingPair(job, "uses", workflowStringNode(uses))
	} else {
		appendWorkflowMappingPair(job, "runs-on", workflowStringNode("ubuntu-latest"))
		steps := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		if run != "" {
			steps.Content = append(steps.Content, workflowStepNode(workflowStepAddition{Run: run}))
		}
		appendWorkflowMappingPair(job, "steps", steps)
	}
	appendWorkflowMappingPair(jobs, jobID, job)
	return yaml.Marshal(document)
}

func addWorkflowMatrixInclude(data []byte, jobID string, rows []map[string]string) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	job := workflowMappingValue(workflowMappingValue(root, "jobs"), jobID)
	if job == nil {
		return nil, fmt.Errorf("workflow job %q is missing", jobID)
	}
	if workflowMappingValue(job, "strategy") != nil {
		return nil, fmt.Errorf("workflow job %q already has a strategy", jobID)
	}
	include := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, values := range rows {
		row := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendWorkflowMappingPair(row, key, workflowStringNode(values[key]))
		}
		include.Content = append(include.Content, row)
	}
	matrix := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendWorkflowMappingPair(matrix, "include", include)
	strategy := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendWorkflowMappingPair(strategy, "matrix", matrix)
	appendWorkflowMappingPair(job, "strategy", strategy)
	return yaml.Marshal(document)
}

func setWorkflowTriggerEvents(data []byte, events []string) ([]byte, error) {
	on := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, event := range events {
		appendWorkflowMappingPair(on, event, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"})
	}
	return setWorkflowTriggerNode(data, on)
}

func setWorkflowTriggerNode(data []byte, on *yaml.Node) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	workflowSetMappingValue(root, "on", on)
	return yaml.Marshal(document)
}

func appendWorkflowJobNeed(data []byte, jobID, dependency string) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	job := workflowMappingValue(workflowMappingValue(root, "jobs"), jobID)
	if job == nil {
		return nil, fmt.Errorf("workflow job %q is missing", jobID)
	}
	needs := workflowMappingValue(job, "needs")
	if needs == nil {
		needs = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		appendWorkflowMappingPair(job, "needs", needs)
	} else if needs.Kind == yaml.ScalarNode {
		needs = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode(needs.Value)}}
		workflowSetMappingValue(job, "needs", needs)
	} else if needs.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("workflow job %q needs is not a sequence", jobID)
	}
	needs.Content = append(needs.Content, workflowStringNode(dependency))
	return yaml.Marshal(document)
}

func workflowStepNode(step workflowStepAddition) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if step.Name != "" {
		appendWorkflowMappingPair(node, "name", workflowStringNode(step.Name))
	}
	if step.WorkingDirectory != "" {
		appendWorkflowMappingPair(node, "working-directory", workflowStringNode(step.WorkingDirectory))
	}
	if step.Run != "" {
		appendWorkflowMappingPair(node, "run", workflowStringNode(step.Run))
	}
	if step.Uses != "" {
		appendWorkflowMappingPair(node, "uses", workflowStringNode(step.Uses))
	}
	if len(step.With) != 0 {
		with := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keys := make([]string, 0, len(step.With))
		for key := range step.With {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendWorkflowMappingPair(with, key, workflowStringNode(step.With[key]))
		}
		appendWorkflowMappingPair(node, "with", with)
	}
	if len(step.Env) != 0 {
		env := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keys := make([]string, 0, len(step.Env))
		for key := range step.Env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendWorkflowMappingPair(env, key, workflowStringNode(step.Env[key]))
		}
		appendWorkflowMappingPair(node, "env", env)
	}
	return node
}

func parseWorkflowDocument(data []byte) (*yaml.Node, *yaml.Node, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, nil, err
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("workflow document has no root mapping")
	}
	return &document, document.Content[0], nil
}

func workflowMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func appendWorkflowMappingPair(mapping *yaml.Node, key string, value *yaml.Node) {
	mapping.Content = append(mapping.Content, workflowStringNode(key), value)
}

func workflowSetMappingValue(mapping *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	appendWorkflowMappingPair(mapping, key, value)
}

func workflowStringNode(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

// TestPullRequestDefersPostLandingTests keeps the slow mutation and exhaustive
// set out of the PR package gate. It also pins the PR trigger and the static
// matrix inputs that may reach commands or environment values.
func TestPullRequestDefersPostLandingTests(t *testing.T) {
	good := workflowBytes(t)
	if err := checkPullRequestDefersPostLandingTests(good); err != nil {
		t.Fatalf("valid PR deferral rejected: %v", err)
	}
	triggerForms := []struct {
		name  string
		plant func([]byte) ([]byte, error)
	}{
		{
			name: "scalar_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				return setWorkflowTriggerNode(data, workflowStringNode("pull_request"))
			},
		},
		{
			name: "sequence_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				on := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode("push"), workflowStringNode("pull_request")}}
				return setWorkflowTriggerNode(data, on)
			},
		},
		{
			name: "mapping_null_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				on := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(on, "pull_request", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"})
				return setWorkflowTriggerNode(data, on)
			},
		},
		{
			name: "mapping_filtered_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				branches := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode("main")}}
				filter := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(filter, "branches", branches)
				on := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(on, "pull_request", filter)
				return setWorkflowTriggerNode(data, on)
			},
		},
	}
	for _, test := range triggerForms {
		t.Run("valid_"+test.name, func(t *testing.T) {
			workflow, err := test.plant(good)
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPullRequestDefersPostLandingTests(workflow); err != nil {
				t.Fatalf("supported trigger form rejected: %v", err)
			}
		})
	}
	tests := []struct {
		name  string
		plant func([]byte) ([]byte, error)
	}{
		{
			name: "restored_pr_job",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, legacyMutationJobID, "Curator mutation controls", "go run ./tools/mutationinventory -mode inventory", "", "")
			},
		},
		{
			name: "renamed_pr_mutation_job",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "deferred-suite", "Curator mutation controls", "go run ./tools/mutationinventory -mode inventory", "", "")
			},
		},
		{
			name: "mutation_job_with_unknown_condition",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "suite-lane", "Scheduled analysis", "go run ./tools/mutationinventory -mode inventory", "", "github.event_name == 'pull_request' || inputs.run_suite")
			},
		},
		{
			name: "neutral_local_reusable_workflow_calls_deferred_suite",
			plant: func(data []byte) ([]byte, error) {
				target := []byte("name: Quality\non: workflow_call\njobs:\n  routine:\n    runs-on: ubuntu-latest\n    steps:\n      - run: go run ./tools/mutationinventory -mode inventory\n")
				var targetWorkflow pullRequestWorkflow
				if err := yaml.Unmarshal(target, &targetWorkflow); err != nil {
					return nil, err
				}
				if !isMutationSuiteJob("routine", targetWorkflow.Jobs["routine"]) {
					return nil, fmt.Errorf("reusable-workflow target fixture does not call the deferred suite")
				}
				return addWorkflowJob(data, "quality-check", "Quality check", "", "./.github/workflows/quality.yml", "")
			},
		},
		{
			name: "renamed_reusable_suite_job",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "quality-check", "Quality check", "", "./.github/workflows/post-landing-mutation-tests.yml", "")
			},
		},
		{
			name: "matrix_include_supplies_deferred_suite_selector",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ matrix.suite_selector }}'`, "", "")
				if err != nil {
					return nil, err
				}
				return addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "TestMutationInventory|TestExhaustive"}})
			},
		},
		{
			name: "matrix_include_supplies_deferred_suite_selector_through_env",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run "$SUITE_SELECTOR"`, "", "")
				if err != nil {
					return nil, err
				}
				mutant, err = addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "TestMutationInventory|TestExhaustive"}})
				if err != nil {
					return nil, err
				}
				document, root, err := parseWorkflowDocument(mutant)
				if err != nil {
					return nil, err
				}
				job := workflowMappingValue(workflowMappingValue(root, "jobs"), "routine-check")
				env := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(env, "SUITE_SELECTOR", workflowStringNode("${{ matrix.suite_selector }}"))
				appendWorkflowMappingPair(job, "env", env)
				return yaml.Marshal(document)
			},
		},
		{
			name: "matrix_include_has_unclassifiable_dynamic_input",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ matrix.suite_selector }}'`, "", "")
				if err != nil {
					return nil, err
				}
				return addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "${{ inputs.suite_filter }}"}})
			},
		},
		{
			name: "matrix_lookup_is_unclassifiable",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ matrix[inputs.selector] }}'`, "", "")
				if err != nil {
					return nil, err
				}
				return addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "TestRoutine"}})
			},
		},
		{
			name: "unclassified_dynamic_run_selector",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ inputs.suite_filter }}'`, "", "")
			},
		},
		{
			name: "scalar_push_only_trigger",
			plant: func(data []byte) ([]byte, error) {
				return setWorkflowTriggerNode(data, workflowStringNode("push"))
			},
		},
		{
			name: "sequence_push_only_trigger",
			plant: func(data []byte) ([]byte, error) {
				on := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode("push")}}
				return setWorkflowTriggerNode(data, on)
			},
		},
		{
			name: "push_only_trigger",
			plant: func(data []byte) ([]byte, error) {
				return setWorkflowTriggerEvents(data, []string{"push"})
			},
		},
		{
			name: "duplicate_trigger_mapping_is_rejected",
			plant: func(data []byte) ([]byte, error) {
				document, root, err := parseWorkflowDocument(data)
				if err != nil {
					return nil, err
				}
				push := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(push, "push", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"})
				appendWorkflowMappingPair(root, "on", push)
				return yaml.Marshal(document)
			},
		},
		{
			name: "restored_pr_dependency",
			plant: func(data []byte) ([]byte, error) {
				return appendWorkflowJobNeed(data, "test", legacyMutationJobID)
			},
		},
		{
			name: "legacy_bridge_in_linux_build_controls",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Require PR mutation gate success",
					Run:  `test "${{ needs.curator-mutation-pr.result }}" = success`,
				})
			},
		},
		{
			name: "renamed_bridge_run",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Run:  `echo "${{ needs.curator-mutation-pr.result }}"`,
				})
			},
		},
		{
			name: "renamed_bridge_env",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Env:  map[string]string{"MUTATION_RESULT": "${{ needs.curator-mutation-pr.result }}"},
				})
			},
		},
		{
			name: "dynamic_mutation_result_reference",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Run:  `echo "${{ needs[vars.MUTATION_JOB].result }}"`,
				})
			},
		},
		{
			name: "renamed_mutation_result_reference",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "deferred-suite", "Curator mutation controls", "go run ./tools/mutationinventory -mode inventory", "", "vars.ROSE_AIR_RUNNER == 'true' && github.event_name == 'push' && github.ref == 'refs/heads/main'")
				if err != nil {
					return nil, err
				}
				return addStepToWorkflowJob(mutant, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Run:  `echo "${{ needs.deferred-suite.result }}"`,
				})
			},
		},
		{
			name: "restored_pr_test_bridge",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "test", workflowStepAddition{
					Name: "Require PR mutation gate success",
					Run:  `test "${{ needs.curator-mutation-pr.result }}" = success`,
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := checkPullRequestDefersPostLandingTests(good); err != nil {
				t.Fatalf("unmodified workflow is not the passing control: %v", err)
			}
			mutant, err := test.plant(good)
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPullRequestDefersPostLandingTests(mutant); err == nil {
				t.Fatal("pull-request deferred-suite violation admitted")
			}
		})
	}
}

func TestAddStepToWorkflowJobRequiresExistingStepsSequence(t *testing.T) {
	good := workflowBytes(t)
	if err := checkPullRequestDefersPostLandingTests(good); err != nil {
		t.Fatalf("unmodified workflow is not the passing control: %v", err)
	}
	if _, err := addStepToWorkflowJob(good, "linux-build-controls", workflowStepAddition{Name: "control", Run: "true"}); err != nil {
		t.Fatalf("valid workflow helper use failed: %v", err)
	}
	for _, test := range []struct {
		name     string
		workflow string
		job      string
	}{
		{name: "missing_job", workflow: "jobs:\n  present:\n    steps: []\n", job: "missing"},
		{name: "missing_steps", workflow: "jobs:\n  present:\n    runs-on: ubuntu-latest\n", job: "present"},
		{name: "non_sequence_steps", workflow: "jobs:\n  present:\n    steps:\n      run: echo\n", job: "present"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := addStepToWorkflowJob([]byte(test.workflow), test.job, workflowStepAddition{Name: "control"}); err == nil {
				t.Fatal("helper accepted a missing job or invalid steps field")
			}
		})
	}
}

// Claim: the PR guard rejects deferred test selectors resolved from workflow,
// job, step, package, and GOFLAGS inputs while admitting safe neighbors.
// Limit: this is a static workflow check, not hosted-runner evidence.
func TestPullRequestGuardResolvesDeferredSelectorsFromEnvironment(t *testing.T) {
	good := workflowBytes(t)
	if err := checkPullRequestDefersPostLandingTests(good); err != nil {
		t.Fatalf("unmodified workflow is not the passing control: %v", err)
	}
	makeWorkflow := func(scope, selector string) ([]byte, error) {
		data, err := addWorkflowJob(good, "selector-control", "Selector control", `go test ./... -run "$SUITE_SELECTOR"`, "", "")
		if err != nil {
			return nil, err
		}
		document, root, err := parseWorkflowDocument(data)
		if err != nil {
			return nil, err
		}
		target := root
		if scope != "workflow" {
			target = workflowMappingValue(workflowMappingValue(root, "jobs"), "selector-control")
			if scope == "step" {
				steps := workflowMappingValue(target, "steps")
				if steps == nil || len(steps.Content) != 1 {
					return nil, fmt.Errorf("selector-control job has no single step")
				}
				target = steps.Content[0]
			}
		}
		env := workflowMappingValue(target, "env")
		if env == nil {
			env = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			appendWorkflowMappingPair(target, "env", env)
		}
		appendWorkflowMappingPair(env, "SUITE_SELECTOR", workflowStringNode(selector))
		return yaml.Marshal(document)
	}
	for _, scope := range []string{"workflow", "job", "step"} {
		t.Run(scope+"_environment_refuses_deferred_selector", func(t *testing.T) {
			workflow, err := makeWorkflow(scope, curatormutations.PostLandingSelector())
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPullRequestDefersPostLandingTests(workflow); err == nil {
				t.Fatalf("deferred selector from %s environment was admitted", scope)
			}
			workflow, err = makeWorkflow(scope, "^TestRoutine$")
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPullRequestDefersPostLandingTests(workflow); err != nil {
				t.Fatalf("ordinary selector from %s environment rejected: %v", scope, err)
			}
		})
	}
	for _, test := range []struct {
		name      string
		packageID string
		wantError bool
	}{
		{name: "deferred_package_from_environment_refuses", packageID: "./cmd/curator", wantError: true},
		{name: "ordinary_package_from_environment_admits", packageID: "./tools/ciworkflow"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addWorkflowJob(good, "package-control", "Package control", `go test "$TEST_PACKAGE"`, "", "")
			if err != nil {
				t.Fatal(err)
			}
			document, root, err := parseWorkflowDocument(workflow)
			if err != nil {
				t.Fatal(err)
			}
			job := workflowMappingValue(workflowMappingValue(root, "jobs"), "package-control")
			steps := workflowMappingValue(job, "steps")
			if steps == nil || len(steps.Content) != 1 {
				t.Fatal("package-control job has no single step")
			}
			step := steps.Content[0]
			env := workflowMappingValue(step, "env")
			if env == nil {
				env = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(step, "env", env)
			}
			appendWorkflowMappingPair(env, "TEST_PACKAGE", workflowStringNode(test.packageID))
			workflow, err = yaml.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			err = checkPullRequestDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("deferred package target from the step environment was admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("ordinary package target from the step environment was rejected: %v", err)
			}
		})
	}
	workflow, err := addWorkflowJob(good, "unknown-package-control", "Unknown package control", `go test "$MISSING_PACKAGE"`, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPullRequestDefersPostLandingTests(workflow); err == nil {
		t.Fatal("unresolved dynamic package target was admitted")
	}
	for _, test := range []struct {
		name      string
		goPath    string
		selector  string
		wantError bool
	}{
		{name: "absolute_go_path_refuses_deferred_selector", goPath: "/opt/go/bin/go", selector: curatormutations.PostLandingSelector(), wantError: true},
		{name: "absolute_go_exe_path_refuses_deferred_selector", goPath: "C:/Go/bin/go.exe", selector: curatormutations.PostLandingSelector(), wantError: true},
		{name: "absolute_windows_backslash_go_path_refuses_deferred_selector", goPath: `C:\Go\bin\go.exe`, selector: curatormutations.PostLandingSelector(), wantError: true},
		{name: "absolute_go_path_admits_ordinary_selector", goPath: "/opt/go/bin/go", selector: "^TestRoutine$"},
		{name: "unapproved_absolute_go_path_refuses", goPath: "/tmp/go", selector: "^TestRoutine$", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := test.goPath + " test ./... -run '" + test.selector + "'"
			workflow, err := addWorkflowJob(good, "absolute-go-control", "Absolute Go control", run, "", "")
			if err != nil {
				t.Fatal(err)
			}
			err = checkPullRequestDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("absolute Go invocation admitted a deferred selector")
			}
			if !test.wantError && err != nil {
				t.Fatalf("absolute Go invocation rejected a safe selector: %v", err)
			}
		})
	}
	makeGOFlagsWorkflow := func(scope, goFlags, run string) ([]byte, error) {
		data, err := addWorkflowJob(good, "goflags-control", "GOFLAGS control", run, "", "")
		if err != nil {
			return nil, err
		}
		document, root, err := parseWorkflowDocument(data)
		if err != nil {
			return nil, err
		}
		target := root
		if scope != "workflow" {
			target = workflowMappingValue(workflowMappingValue(root, "jobs"), "goflags-control")
			if scope == "step" {
				steps := workflowMappingValue(target, "steps")
				if steps == nil || len(steps.Content) != 1 {
					return nil, fmt.Errorf("GOFLAGS-control job has no single step")
				}
				target = steps.Content[0]
			}
		}
		env := workflowMappingValue(target, "env")
		if env == nil {
			env = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			appendWorkflowMappingPair(target, "env", env)
		}
		appendWorkflowMappingPair(env, "GOFLAGS", workflowStringNode(goFlags))
		return yaml.Marshal(document)
	}
	goFlagsCases := []struct {
		name      string
		scope     string
		goFlags   string
		run       string
		wantError bool
	}{
		{name: "workflow_goflags_deferred_refuses", scope: "workflow", goFlags: `-run="` + curatormutations.PostLandingSelector() + `"`, run: "go test ./tools/ciworkflow", wantError: true},
		{name: "workflow_goflags_ordinary_admits", scope: "workflow", goFlags: `-run="^TestRoutine$"`, run: "go test ./tools/ciworkflow"},
		{name: "job_goflags_test_run_alias_refuses", scope: "job", goFlags: `-test.run="` + curatormutations.PostLandingSelector() + `"`, run: "go test ./tools/ciworkflow", wantError: true},
		{name: "step_goflags_complete_skip_admits", scope: "step", goFlags: `-skip="` + curatormutations.PostLandingSelector() + `"`, run: "go test ./..."},
		{name: "command_run_overrides_goflags_selector", scope: "step", goFlags: `-run="` + curatormutations.PostLandingSelector() + `"`, run: `go test ./tools/ciworkflow -run '^TestRoutine$'`},
		{name: "dynamic_goflags_refuses", scope: "step", goFlags: "${{ env.SUITE_SELECTOR }}", run: "go test ./tools/ciworkflow", wantError: true},
		{name: "step_goflags_custom_executor_refuses", scope: "step", goFlags: "-exec=./tools/deferred-wrapper", run: "go test ./tools/ciworkflow", wantError: true},
	}
	for _, test := range goFlagsCases {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := makeGOFlagsWorkflow(test.scope, test.goFlags, test.run)
			if err != nil {
				t.Fatal(err)
			}
			err = checkPullRequestDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("GOFLAGS-selected deferred tests were admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("safe GOFLAGS selection rejected: %v", err)
			}
		})
	}
	for _, test := range []struct {
		name      string
		run       string
		wantError bool
	}{
		{name: "go_global_c_path_refuses_deferred_package", run: "go -C ./cmd/curator test .", wantError: true},
		{name: "go_global_c_equals_path_refuses_deferred_package", run: "go -C=./cmd/curator test .", wantError: true},
		{name: "go_global_c_ancestor_recursive_target_refuses", run: "go -C ./cmd test ./...", wantError: true},
		{name: "go_global_c_path_admits_ordinary_package", run: `go -C ./tools/ciworkflow test . -run '^TestRoutine$'`},
		{name: "go_global_c_equals_path_admits_ordinary_package", run: `go -C=./tools/ciworkflow test . -run '^TestRoutine$'`},
		{name: "go_global_c_unapproved_local_tool_refuses", run: "go -C ./tools run ./deferred-suite", wantError: true},
		{name: "go_global_c_allowlisted_local_tool_admits", run: "go -C ./tools run ./sanitize-ci-evidence -source input -destination output"},
		{name: "go_unmodeled_global_flag_refuses", run: "go -x test ./tools/ciworkflow", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addWorkflowJob(good, "global-go-flag-control", "Global Go flag control", test.run, "", "")
			if err != nil {
				t.Fatal(err)
			}
			err = checkPullRequestDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("Go global flags admitted a deferred package target")
			}
			if !test.wantError && err != nil {
				t.Fatalf("safe Go global flag form was rejected: %v", err)
			}
		})
	}
	if !isWorkflowGoCommand(`C:\Go\bin\go.exe`) {
		t.Fatal("Windows absolute Go executable path was not recognized")
	}
	if isWorkflowGoCommand("gopher") {
		t.Fatal("non-Go command was recognized as the Go executable")
	}
	workflow, err = addWorkflowJob(good, "selector-control", "Selector control", `go test ./... -run "$MISSING_SELECTOR"`, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPullRequestDefersPostLandingTests(workflow); err == nil {
		t.Fatal("unresolved shell selector was admitted")
	}
	workflow, err = addWorkflowJob(good, "literal-selector", "Routine", "go test ./... -run '"+curatormutations.PostLandingSelector()+"'", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPullRequestDefersPostLandingTests(workflow); err == nil {
		t.Fatal("literal deferred selector was admitted")
	}
	for _, test := range []struct {
		name      string
		run       string
		wantError bool
	}{
		{name: "unfiltered_recursive_packages", run: "go test ./...", wantError: true},
		{name: "unfiltered_deferred_root", run: "go test ./cmd/curator", wantError: true},
		{name: "ordinary_recursive_selector", run: `go test ./... -run '^TestRoutine$'`},
		{name: "ordinary_deferred_root_selector", run: `go test ./cmd/curator -run '^TestRoutine$'`},
		{name: "recursive_with_complete_deferred_skip", run: "go test ./... -skip '" + curatormutations.PostLandingSelector() + "'"},
		{name: "recursive_with_partial_deferred_skip_refuses", run: "go test ./... -skip '" + curatormutations.PostLandingSelector() + "/init-bypasses-help'", wantError: true},
		{name: "ordinary_run_with_partial_skip_admitted", run: `go test ./... -run '^TestRoutine$' -skip '^TestRoutine$/init-bypasses-help'`},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addWorkflowJob(good, "selector-shape", "Selector shape", test.run, "", "")
			if err != nil {
				t.Fatal(err)
			}
			err = checkPullRequestDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("deferred test execution was admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("safe test selection rejected: %v", err)
			}
		})
	}
}

// Claim: local actions and scripts are refused unless their exact target is
// admitted, while supported remote actions and guarded CI scripts remain valid.
// Limit: an allowlisted target remains trusted by path; this does not run it.
func TestPullRequestGuardTracksShellDirectoryChanges(t *testing.T) {
	good := workflowBytes(t)
	for _, test := range []struct {
		name      string
		run       string
		wantError bool
	}{
		{name: "cd_to_deferred_package_refuses", run: "cd cmd/curator && go test .", wantError: true},
		{name: "command_cd_to_deferred_package_refuses", run: "command cd cmd/curator && go test .", wantError: true},
		{name: "pushd_to_deferred_package_refuses", run: "pushd cmd/curator; go test .", wantError: true},
		{name: "pushd_and_popd_to_safe_tree_admits", run: "pushd tools && go test ./... && popd"},
		{name: "popd_to_deferred_tree_refuses", run: "pushd tools && popd && go test ./...", wantError: true},
		{name: "dynamic_cd_target_refuses", run: `cd "$X" && go test ./...`, wantError: true},
		{name: "pushd_stack_option_refuses", run: "pushd -0 && go test .", wantError: true},
		{name: "pushd_plus_stack_option_refuses", run: "pushd +0 && go test .", wantError: true},
		{name: "cdpath_prefix_refuses", run: "CDPATH=cmd cd curator && go test .", wantError: true},
		{name: "cdpath_export_refuses", run: "export CDPATH=cmd; cd curator && go test .", wantError: true},
		{name: "empty_cdpath_safe_neighbor", run: "CDPATH= cd tools && go test ./..."},
		{name: "cd_to_safe_package_tree_admits", run: "cd tools && go test ./..."},
		{name: "subshell_cd_to_safe_package_tree_admits", run: "(cd tools && go test ./...)"},
		{name: "cd_dash_refuses", run: "cd - && go test ./...", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addWorkflowJob(good, "directory-control", "Directory control", test.run, "", "")
			if err != nil {
				t.Fatal(err)
			}
			err = checkPullRequestDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("shell directory change reached a deferred test package")
			}
			if !test.wantError && err != nil {
				t.Fatalf("safe shell directory change was rejected: %v", err)
			}
		})
	}
}

func TestPullRequestGuardResolvesShellDirectorySymlinks(t *testing.T) {
	root, err := prRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	tempRoot := filepath.Join(root, ".temp")
	if err := os.MkdirAll(tempRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	tempDir, err := os.MkdirTemp(tempRoot, "deferred-directory-symlink-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })
	link := filepath.Join(tempDir, "curator")
	if err := os.Symlink(filepath.Join(root, "cmd", "curator"), link); err != nil {
		t.Skipf("symlink creation is unavailable: %v", err)
	}
	relative, err := filepath.Rel(root, link)
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := addWorkflowJob(workflowBytes(t), "directory-symlink-control", "Directory symlink control", "cd "+filepath.ToSlash(relative)+" && go test .", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPullRequestDefersPostLandingTests(workflow); err == nil {
		t.Fatal("shell directory symlink reached a deferred test package")
	}
}

func TestPullRequestGuardChecksLocalActionsAndScripts(t *testing.T) {
	good := workflowBytes(t)
	document, root, err := parseWorkflowDocument(good)
	if err != nil {
		t.Fatal(err)
	}
	jobs := workflowMappingValue(root, "jobs")
	changedHeredoc := false
	for i := 1; i < len(jobs.Content) && !changedHeredoc; i += 2 {
		steps := workflowMappingValue(jobs.Content[i], "steps")
		if steps == nil {
			continue
		}
		for _, step := range steps.Content {
			name := workflowMappingValue(step, "name")
			if name == nil || name.Value != "Driver qualification and reproducibility" {
				continue
			}
			run := workflowMappingValue(step, "run")
			run.Value += "\npython3 - <<'MUTATION'\nimport subprocess\nsubprocess.run(['go', 'test', './...'])\nMUTATION\n"
			changedHeredoc = true
			break
		}
	}
	if !changedHeredoc {
		t.Fatal("did not find the allowlisted Python heredoc to mutate")
	}
	mutatedWorkflow, err := yaml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPullRequestDefersPostLandingTests(mutatedWorkflow); err == nil {
		t.Fatal("modified allowlisted interpreter heredoc was admitted")
	}
	document, root, err = parseWorkflowDocument(good)
	if err != nil {
		t.Fatal(err)
	}
	jobs = workflowMappingValue(root, "jobs")
	changedGeneratedGoRun := false
	for i := 1; i < len(jobs.Content) && !changedGeneratedGoRun; i += 2 {
		steps := workflowMappingValue(jobs.Content[i], "steps")
		if steps == nil {
			continue
		}
		for _, step := range steps.Content {
			name := workflowMappingValue(step, "name")
			if name == nil || name.Value != "Record macOS race capacity" {
				continue
			}
			run := workflowMappingValue(step, "run")
			run.Value += "\necho changed\n"
			changedGeneratedGoRun = true
			break
		}
	}
	if !changedGeneratedGoRun {
		t.Fatal("did not find the allowlisted generated Go command to mutate")
	}
	mutatedWorkflow, err = yaml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPullRequestDefersPostLandingTests(mutatedWorkflow); err == nil {
		t.Fatal("modified allowlisted generated Go command was admitted")
	}
	addStep := func(uses, run string, with, env map[string]string, workingDirectory ...string) ([]byte, error) {
		data, err := addWorkflowJob(good, "local-target-control", "Local target control", "", "", "")
		if err != nil {
			return nil, err
		}
		step := workflowStepAddition{Uses: uses, Run: run, With: with, Env: env}
		if len(workingDirectory) != 0 {
			step.WorkingDirectory = workingDirectory[0]
		}
		return addStepToWorkflowJob(data, "local-target-control", step)
	}
	for _, test := range []struct {
		name             string
		uses             string
		run              string
		with             map[string]string
		env              map[string]string
		workingDirectory string
		wantError        bool
	}{
		{name: "unapproved_local_action_with_selector", uses: "./.github/actions/deferred-suite", with: map[string]string{"selector": curatormutations.PostLandingSelector()}, wantError: true},
		{name: "remote_action_neighbor", uses: "actions/checkout@v4", with: map[string]string{"fetch-depth": "1"}},
		{name: "unapproved_local_script", run: "bash .github/ci/deferred-suite.sh", wantError: true},
		{name: "unapproved_bare_bash_script", run: "bash deferred-suite.sh", wantError: true},
		{name: "unapproved_bare_sh_script", run: "sh x.sh", wantError: true},
		{name: "allowlisted_script_basename_from_working_directory_refuses", run: "bash isolate-go-state.sh", workingDirectory: ".github/ci", wantError: true},
		{name: "unapproved_local_go_tool", run: "go run ./tools/deferred-suite", wantError: true},
		{name: "allowlisted_local_go_tool_neighbor", run: "go run ./tools/sanitize-ci-evidence -source input -destination output"},
		{name: "go_test_custom_executor", run: `go test ./tools/ciworkflow -run '^TestRoutine$' -exec ./tools/deferred-wrapper`, wantError: true},
		{name: "go_test_without_custom_executor", run: `go test ./tools/ciworkflow -run '^TestRoutine$'`},
		{name: "dynamic_local_go_target", run: `go run "$LOCAL_GO_FILE"`, wantError: true},
		{name: "dynamic_interpreter_script", run: `bash "$LOCAL_SCRIPT"`, wantError: true},
		{name: "inline_shell_body", run: `bash -c 'go test ./... -run "${SUITE_SELECTOR}"'`, wantError: true},
		{name: "bash_inline_command", run: `bash -c 'go test ./...'`, wantError: true},
		{name: "ruby_inline_eval", run: `ruby -e 'require "open3"; Open3.capture3("go", "test", "./...")'`, wantError: true},
		{name: "ruby_extensionless_script", run: "ruby deferred-suite", wantError: true},
		{name: "quoted_command_substitution_interpreter", run: `echo "$(ruby -e 'system("go", "test", "./...")')"`, wantError: true},
		{name: "command_substitution_go_test", run: `echo "$(go test ./... -run TestCommandHelpNarrowingMutants)"`, wantError: true},
		{name: "command_substitution_go_env_neighbor", run: `echo "$(go env GOTOOLCHAIN)"`},
		{name: "process_substitution_refuses", run: "cat <(printf safe)", wantError: true},
		{name: "find_exec_interpreter", run: `find . -exec ruby -e 'puts 1' \;`, wantError: true},
		{name: "parallel_execution_dispatcher", run: `parallel ruby -e 'puts 1'`, wantError: true},
		{name: "make_execution_dispatcher", run: "make test", wantError: true},
		{name: "coproc_execution_dispatcher", run: `coproc ruby -e 'puts 1'`, wantError: true},
		{name: "find_print_neighbor", run: "find . -type f -print"},
		{name: "perl_inline_eval", run: `perl -e 'system("go", "test", "./...")'`, wantError: true},
		{name: "node_inline_eval", run: `node -e 'require("child_process").spawnSync("go", ["test", "./..."])'`, wantError: true},
		{name: "node_inline_print", run: `node -p 'require("child_process").spawnSync("go", ["test", "./..."])'`, wantError: true},
		{name: "node_long_inline_eval", run: `node --eval 'require("child_process").spawnSync("go", ["test", "./..."])'`, wantError: true},
		{name: "python_launcher_inline_eval", run: `py -c 'import subprocess; subprocess.run(["go", "test", "./..."])'`, wantError: true},
		{name: "powershell_inline_command", run: `pwsh -Command 'go test ./...'`, wantError: true},
		{name: "windows_command_shell", run: `cmd /C "go test ./..."`, wantError: true},
		{name: "versioned_perl_interpreter", run: `perl5.36 -e 'print 1'`, wantError: true},
		{name: "versioned_node_interpreter", run: `node20 -p 'console.log("ok")'`, wantError: true},
		{name: "versioned_php_interpreter", run: `php8.3 -r 'echo "ok";'`, wantError: true},
		{name: "versioned_shell_interpreter", run: `bash5.2 -c 'printf ok'`, wantError: true},
		{name: "windows_executable_interpreter", run: `python3.exe -c 'print(1)'`, wantError: true},
		{name: "busybox_shell_interpreter", run: `busybox ash -c 'echo ok'`, wantError: true},
		{name: "env_split_string_interpreter", run: `env -S "ruby -e 'puts 1'"`, wantError: true},
		{name: "extensionless_sourced_script", run: "source ./x", wantError: true},
		{name: "extensionless_dot_script", run: ". ./x", wantError: true},
		{name: "unmodeled_shell_function", run: "noop() { echo ok; }", wantError: true},
		{name: "shell_alias_directory_change", run: "alias cdir='cd cmd/curator'; cdir; go test .", wantError: true},
		{name: "shell_trap_directory_change", run: "trap 'cd cmd/curator' DEBUG; go test .", wantError: true},
		{name: "extensionless_local_executable", run: "./x", wantError: true},
		{name: "absolute_local_executable", run: "/tmp/x", wantError: true},
		{name: "extensionless_local_executable_through_env", run: "env ./x", wantError: true},
		{name: "extensionless_local_executable_in_subshell", run: "(./x)", wantError: true},
		{name: "allowlisted_python_script_neighbor", run: "python3 .github/ci/default-build-transport-control.py"},
		{name: "unapproved_interpreter_heredoc", run: "python3 - <<'PY'\nimport subprocess\nsubprocess.run(['go', 'test', './...'])\nPY", wantError: true},
		{name: "dynamic_sourced_script", run: `source "$LOCAL_SCRIPT"`, wantError: true},
		{name: "allowlisted_helper_script_neighbor", run: "bash .github/ci/isolate-go-state.sh"},
		{name: "sourcing_allowlisted_script_refuses", run: "source .github/ci/isolate-go-state.sh", wantError: true},
		{name: "test_gate_without_skip", run: "bash .github/ci/test-gate.sh", wantError: true},
		{name: "test_gate_with_skip", run: "bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}},
		{name: "test_gate_inline_skip_override", run: "CI_SKIP_POST_LANDING_TESTS=0 bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
		{name: "test_gate_case_insensitive_skip_override", run: "ci_skip_post_landing_tests=0 bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
		{name: "test_gate_skip_unset", run: "env -u CI_SKIP_POST_LANDING_TESTS bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
		{name: "test_gate_shell_unset", run: "unset CI_SKIP_POST_LANDING_TESTS; bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addStep(test.uses, test.run, test.with, test.env, test.workingDirectory)
			if err != nil {
				t.Fatal(err)
			}
			err = checkPullRequestDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("unapproved local action or deferred script was admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("supported local or remote action was rejected: %v", err)
			}
		})
	}
}
