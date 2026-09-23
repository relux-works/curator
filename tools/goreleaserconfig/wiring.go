package goreleaserconfig

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// GateRun is the exact lint-step command that enforces the rc channel values
// on every push. Both this package's wiring test and gate-selftest.sh read
// it back out of .github/workflows/ci.yml so the step cannot silently
// disappear or be neutered.
const GateRun = "go test -count=1 ./tools/goreleaserconfig/"

// CheckWiring parses data as the CI workflow and verifies the lint job
// carries exactly one live step running GateRun. A step counts as live only
// when its parsed run value equals GateRun and it carries no if key: a
// commented-out run line parses as no key at all, and any condition (even
// `if: true`) needs human review before it may carry a must-run gate.
// Returns one failure message per violation; nil means the wiring is
// exactly as required.
func CheckWiring(data []byte) []string {
	// Strict decode first, as in Check: duplicate mapping keys must
	// fail the pin rather than resolve by first/last wins.
	var strict any
	if err := yaml.Unmarshal(data, &strict); err != nil {
		return []string{"invalid YAML: " + err.Error()}
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return []string{"invalid YAML: " + err.Error()}
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return []string{fmt.Sprintf("workflow is not a mapping, want jobs.lint.steps to run %q", GateRun)}
	}
	jobs := mappingValue(doc.Content[0], "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return []string{fmt.Sprintf("section %q is absent, want jobs.lint.steps to run %q", "jobs", GateRun)}
	}
	lint := mappingValue(jobs, "lint")
	if lint == nil || lint.Kind != yaml.MappingNode {
		return []string{fmt.Sprintf("job %q is absent, want lint to run %q", "lint", GateRun)}
	}
	steps := mappingValue(lint, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return []string{fmt.Sprintf("job %q has no steps, want one step running %q", "lint", GateRun)}
	}
	live := 0
	for _, step := range steps.Content {
		if step.Kind != yaml.MappingNode {
			continue
		}
		run := mappingValue(step, "run")
		if run == nil || run.Kind != yaml.ScalarNode || run.Value != GateRun {
			continue
		}
		if mappingValue(step, "if") != nil {
			continue
		}
		live++
	}
	switch live {
	case 1:
		return nil
	case 0:
		return []string{fmt.Sprintf("no live lint step runs %q, want exactly one (a commented run: line or an if: condition does not count)", GateRun)}
	default:
		return []string{fmt.Sprintf("%d live lint steps run %q, want exactly one", live, GateRun)}
	}
}
