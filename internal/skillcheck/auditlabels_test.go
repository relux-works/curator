package skillcheck

import (
	"testing"
)

// auditLabelSkill lays out a one-command skill at the requested manifest
// schema: a declared-only or enforced script command with the given declared
// network hosts. The prompt text carries the shell-neutral resolver, so the
// only warnings the fixture can produce are the script audit classes.
func auditLabelSkill(t *testing.T, schema int, enforced bool, network []string) string {
	t.Helper()
	dir := t.TempDir()
	writeSkillFile(t, dir, "SKILL.md", "---\nname: skill\ndescription: d\n---\n"+
		"Resolve the command from project .agents/bin (tool.cmd on Windows), then the manager global/bin fallback, "+
		"then a validated bare command with command -v or Get-Command.\n")
	writeSkillFile(t, dir, "scripts/tool", "#!/bin/sh\necho ok\n")
	command := map[string]any{"type": "script", "unix_path": "scripts/tool"}
	if enforced {
		command["execution_policy"] = "script-worker-v1"
		command["interpreter"] = "python3-v1"
	}
	caps := map[string]any{}
	if network != nil {
		caps["network"] = network
	}
	writeSkillFile(t, dir, "agent-skill.json", marshal(t, map[string]any{
		"schema_version": schema,
		"capabilities":   caps,
		"runtime_roots":  []string{"scripts"},
		"commands":       map[string]any{"tool": command},
	}))
	return dir
}

// scriptLabelIssues selects the script audit warning classes from a
// validation report.
func scriptLabelIssues(issues []Issue) []Issue {
	var selected []Issue
	for _, issue := range issues {
		if issue.Code == "script-command-declared-only" ||
			issue.Code == "script-command-unfiltered-declared-network" {
			selected = append(selected, issue)
		}
	}
	return selected
}

// TestValidateEmitsScriptAuditLabels drives the four audit-label shapes
// through the validation entry. The issue code is asserted as a literal,
// exactly as the conformance vector names it; every label is a warning and
// none is an error, so declared-only skills keep validating.
func TestValidateEmitsScriptAuditLabels(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		schema   int
		enforced bool
		network  []string
		wantCode string
		wantPath string
	}{
		{"schema7-script", 7, false, nil, "script-command-declared-only", "commands.tool.execution_policy"},
		{"schema8-declared-only-script", 8, false, nil, "script-command-declared-only", "commands.tool.execution_policy"},
		{"schema8-enforced-script", 8, true, nil, "", ""},
		{"schema8-enforced-unfiltered-network", 8, true, []string{"api.example.com"},
			"script-command-unfiltered-declared-network", "commands.tool"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			issues := Validate(auditLabelSkill(t, testCase.schema, testCase.enforced, testCase.network), "")
			if HasErrors(issues) {
				t.Fatalf("a labelled skill reported errors: %+v", issues)
			}
			selected := scriptLabelIssues(issues)
			if testCase.wantCode == "" {
				if len(selected) != 0 {
					t.Fatalf("an enforced script with no declared hosts was labelled: %+v", selected)
				}
				return
			}
			if len(selected) != 1 {
				t.Fatalf("script labels = %+v, want exactly %q (all issues: %+v)",
					selected, testCase.wantCode, issues)
			}
			if selected[0].Code != testCase.wantCode {
				t.Fatalf("code = %q, want %q", selected[0].Code, testCase.wantCode)
			}
			if selected[0].Severity != "warning" {
				t.Fatalf("severity = %q, want warning: %+v", selected[0].Severity, selected[0])
			}
			if selected[0].Path != testCase.wantPath {
				t.Fatalf("path = %q, want %q: %+v", selected[0].Path, testCase.wantPath, selected[0])
			}
		})
	}
}

// TestValidateEnforcedIsNeverDeclaredOnly is the validation-level negative
// row for the declared-only class.
func TestValidateEnforcedIsNeverDeclaredOnly(t *testing.T) {
	for _, network := range [][]string{nil, {"api.example.com"}} {
		issues := Validate(auditLabelSkill(t, 8, true, network), "")
		if HasErrors(issues) {
			t.Fatalf("an enforced skill reported errors: %+v", issues)
		}
		for _, issue := range scriptLabelIssues(issues) {
			if issue.Code == "script-command-declared-only" {
				t.Fatalf("an enforced command was labelled declared-only: %+v", issues)
			}
		}
	}
}

// TestValidateDeclaredOnlyWithHostsCarriesNoNetworkLabel is the
// validation-level negative row for the network class.
func TestValidateDeclaredOnlyWithHostsCarriesNoNetworkLabel(t *testing.T) {
	issues := Validate(auditLabelSkill(t, 8, false, []string{"api.example.com"}), "")
	if HasErrors(issues) {
		t.Fatalf("a declared-only skill reported errors: %+v", issues)
	}
	selected := scriptLabelIssues(issues)
	if len(selected) != 1 || selected[0].Code != "script-command-declared-only" {
		t.Fatalf("a declared-only command with hosts was labelled %+v", selected)
	}
}
