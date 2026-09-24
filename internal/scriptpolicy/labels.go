package scriptpolicy

import (
	"sort"

	"github.com/relux-works/curator/internal/skillspec"
)

// Closed script audit warning classes (manager profile §7). Both are always
// `warn` in every mode and never block: neither is a finding about source
// content, neither is subject to `fail_on`, and neither may be reported as
// an applied control.
const (
	// LabelDeclaredOnly marks a script command that does not declare
	// `execution_policy: "script-worker-v1"`: every schema-7 script
	// command and every schema-8 script command without the field. Its
	// capability declaration is documentation and bounds nothing at run
	// time.
	LabelDeclaredOnly = "script-command-declared-only"
	// LabelUnfilteredDeclaredNetwork marks an enforced script command
	// whose declared `network` hosts are non-empty. The globs are
	// recorded and reported; no portable filtering is applied and none
	// is claimed. A declared-only command with network hosts carries
	// only LabelDeclaredOnly: the network label is about enforcement,
	// not declaration.
	LabelUnfilteredDeclaredNetwork = "script-command-unfiltered-declared-network"
)

// EffectiveLabels reports the closed script-policy labels a source-audit
// evidence report binds. This manager does not implement script-worker-v1,
// so every enforced command is refused at admission; the label records that
// posture so a future worker ships a new label and ages stored bindings out
// through the policy digest instead of inheriting verdicts taken under a
// different containment promise.
func EffectiveLabels() []string {
	return []string{
		"script-worker-v1:" + StateUnsupported,
	}
}

// AuditLabels reports the audit warning classes for one command. Only
// script commands are labelled; build and system commands never carry
// either class. An enforced command is never labelled declared-only, and
// the unfiltered-network label applies only to enforced commands whose
// declared network hosts are non-empty. An unknown execution policy is
// refused at admission, so it carries no audit label here.
func AuditLabels(command skillspec.Command, networkHosts []string) []string {
	if command.Type != "script" {
		return nil
	}
	if command.ExecutionPolicy == "" {
		return []string{LabelDeclaredOnly}
	}
	if command.ExecutionPolicy == skillspec.ScriptExecutionPolicy && len(networkHosts) != 0 {
		return []string{LabelUnfilteredDeclaredNetwork}
	}
	return nil
}

// CommandAuditLabels is one command's audit warning classes.
type CommandAuditLabels struct {
	Command string
	Labels  []string
}

// CommandAuditPolicy is one script command's entry in the audit record
// (manager profile §7): the effective execution-policy identity or its
// explicit absence. Policy carries the declared identity verbatim — the
// closed `script-worker-v1` for an enforced command — and is empty when
// the command declares no policy. Unlike CommandAuditLabels, every script
// command has an entry whether or not it earns a warning class, so a
// reviewer can separate enforced commands from declared-only ones without
// consulting the warnings.
type CommandAuditPolicy struct {
	Command string `json:"command"`
	Policy  string `json:"policy"`
}

// AuditPoliciesForCommands reports the audit-record entry for every script
// command in command-lexical order, so the CLI, the stored verdict, and
// every host list the same commands in the same order. Commands without a
// warning are included: warning eligibility never drops a record entry.
// Build and system commands carry no execution policy and are omitted.
func AuditPoliciesForCommands(commands map[string]skillspec.Command) []CommandAuditPolicy {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	var policies []CommandAuditPolicy
	for _, name := range names {
		command := commands[name]
		if command.Type != "script" {
			continue
		}
		policies = append(policies, CommandAuditPolicy{Command: name, Policy: command.ExecutionPolicy})
	}
	return policies
}

// AuditLabelsForCommands reports the audit warning classes for every
// labelled command in command-lexical order, so the audit gate and the
// validation output list the same commands in the same order on every
// host and every run. Commands without a label are omitted.
func AuditLabelsForCommands(commands map[string]skillspec.Command, networkHosts []string) []CommandAuditLabels {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	var labelled []CommandAuditLabels
	for _, name := range names {
		if labels := AuditLabels(commands[name], networkHosts); len(labels) != 0 {
			labelled = append(labelled, CommandAuditLabels{Command: name, Labels: labels})
		}
	}
	return labelled
}
