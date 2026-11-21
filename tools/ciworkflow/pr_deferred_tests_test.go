package ciworkflow

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type pullRequestStep struct {
	Name string    `yaml:"name"`
	If   string    `yaml:"if"`
	Run  string    `yaml:"run"`
	Env  yaml.Node `yaml:"env"`
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
	Name string
	Run  string
	Env  map[string]string
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
	if step.Run != "" {
		appendWorkflowMappingPair(node, "run", workflowStringNode(step.Run))
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
