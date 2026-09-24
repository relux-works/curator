package scriptpolicy

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

func TestEffectiveLabels(t *testing.T) {
	labels := EffectiveLabels()
	if len(labels) == 0 {
		t.Fatalf("no effective labels")
	}
	joined := strings.Join(labels, "\n")
	if !strings.Contains(joined, "script-worker-v1") {
		t.Fatalf("labels %q miss the script-worker-v1 posture", labels)
	}
	seen := map[string]bool{}
	for _, label := range labels {
		if label == "" || seen[label] {
			t.Fatalf("labels must be non-empty and unique: %q", labels)
		}
		seen[label] = true
	}
}

// TestAuditLabelConstantsPinTheVectorText pins the closed warning-class
// names as string literals, independent of the conformance vector read.
// The vector consumer (TestScriptAuditLabelCases) binds these same names
// to the suite's own bytes; a rename on either side fails exactly one of
// the two tests.
func TestAuditLabelConstantsPinTheVectorText(t *testing.T) {
	if LabelDeclaredOnly != "script-command-declared-only" {
		t.Fatalf("LabelDeclaredOnly = %q", LabelDeclaredOnly)
	}
	if LabelUnfilteredDeclaredNetwork != "script-command-unfiltered-declared-network" {
		t.Fatalf("LabelUnfilteredDeclaredNetwork = %q", LabelUnfilteredDeclaredNetwork)
	}
}

// TestAuditLabelsVectorShapes drives the four audit-label shapes through
// the label helper: schema-7 and schema-8 declared-only scripts carry the
// declared-only class, an enforced script with no declared hosts carries
// nothing, and an enforced script with declared hosts carries only the
// unfiltered-network class.
func TestAuditLabelsVectorShapes(t *testing.T) {
	declared := skillspec.Command{Name: "tool", Type: "script", UnixPath: "scripts/tool"}
	enforced := skillspec.Command{
		Name: "tool", Type: "script", UnixPath: "scripts/tool",
		ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
	}
	hosts := []string{"api.example.com"}
	for _, testCase := range []struct {
		name    string
		command skillspec.Command
		network []string
		want    []string
	}{
		{"schema7-script", declared, nil, []string{LabelDeclaredOnly}},
		{"schema8-declared-only-script", declared, nil, []string{LabelDeclaredOnly}},
		{"schema8-enforced-script", enforced, nil, nil},
		{"schema8-enforced-unfiltered-network", enforced, hosts, []string{LabelUnfilteredDeclaredNetwork}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got := AuditLabels(testCase.command, testCase.network)
			if strings.Join(got, ",") != strings.Join(testCase.want, ",") {
				t.Fatalf("AuditLabels = %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestAuditLabelsEnforcedIsNeverDeclaredOnly is the negative row for the
// declared-only class: an enforced command carries no declared-only label
// whether or not it declares network hosts. Labelling enforced commands
// declared-only fails here.
func TestAuditLabelsEnforcedIsNeverDeclaredOnly(t *testing.T) {
	enforced := skillspec.Command{
		Name: "tool", Type: "script", UnixPath: "scripts/tool",
		ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
	}
	for _, network := range [][]string{nil, {"api.example.com"}} {
		for _, label := range AuditLabels(enforced, network) {
			if label == LabelDeclaredOnly {
				t.Fatalf("an enforced command was labelled declared-only (network %q)", network)
			}
		}
	}
}

// TestAuditLabelsDeclaredOnlyWithHostsCarriesNoNetworkLabel is the
// negative row for the network class: a declared-only command with
// network hosts carries only declared-only. The network label is about
// enforcement, not declaration.
func TestAuditLabelsDeclaredOnlyWithHostsCarriesNoNetworkLabel(t *testing.T) {
	declared := skillspec.Command{Name: "tool", Type: "script", UnixPath: "scripts/tool"}
	got := AuditLabels(declared, []string{"api.example.com"})
	if len(got) != 1 || got[0] != LabelDeclaredOnly {
		t.Fatalf("a declared-only command with hosts was labelled %q, want only %q", got, LabelDeclaredOnly)
	}
}

// TestAuditLabelsIgnoreNonScriptCommands proves build and system commands
// never carry either script class, even when a manifest declares network
// hosts.
func TestAuditLabelsIgnoreNonScriptCommands(t *testing.T) {
	commands := map[string]skillspec.Command{
		"build": {Name: "build", Type: "build", Driver: "go-v1", SourceDir: "cmd/build"},
		"sys":   {Name: "sys", Type: "system", Command: "git"},
	}
	if got := AuditLabelsForCommands(commands, []string{"api.example.com"}); len(got) != 0 {
		t.Fatalf("non-script commands were labelled: %+v", got)
	}
}

// TestAuditLabelsUnknownPolicyCarriesNoLabel proves a command selecting an
// unknown execution policy carries no audit label: it is refused at
// admission, so neither the declared-only nor the enforced-network class
// describes it.
func TestAuditLabelsUnknownPolicyCarriesNoLabel(t *testing.T) {
	future := skillspec.Command{
		Name: "tool", Type: "script", UnixPath: "scripts/tool",
		ExecutionPolicy: "script-worker-v9", Interpreter: "python3-v1",
	}
	for _, network := range [][]string{nil, {"api.example.com"}} {
		if got := AuditLabels(future, network); len(got) != 0 {
			t.Fatalf("an unknown policy was labelled %q (network %q)", got, network)
		}
	}
}

// TestAuditPoliciesForCommandsCoversEveryScriptCommand proves the audit
// record carries every script command's execution-policy identity or its
// explicit absence, independent of warning eligibility: the enforced
// no-host command earns no warning class yet keeps its exact
// `script-worker-v1` entry, the declared-only command records an empty
// policy, and build/system commands are omitted.
func TestAuditPoliciesForCommandsCoversEveryScriptCommand(t *testing.T) {
	commands := map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", UnixPath: "scripts/tool"},
		"guarded": {
			Name: "guarded", Type: "script", UnixPath: "scripts/guarded",
			ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
		},
		"net": {
			Name: "net", Type: "script", UnixPath: "scripts/net",
			ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
		},
		"build": {Name: "build", Type: "build", Driver: "go-v1", SourceDir: "cmd/build"},
		"sys":   {Name: "sys", Type: "system", Command: "git"},
	}
	got := AuditPoliciesForCommands(commands)
	want := []CommandAuditPolicy{
		{Command: "guarded", Policy: "script-worker-v1"},
		{Command: "net", Policy: "script-worker-v1"},
		{Command: "tool", Policy: ""},
	}
	if len(got) != len(want) {
		t.Fatalf("policy record = %+v, want %+v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("policy record = %+v, want %+v", got, want)
		}
	}
	// The no-warning command keeps its entry while earning no label.
	if labelled := AuditLabelsForCommands(commands, nil); len(labelled) != 1 || labelled[0].Command != "tool" {
		t.Fatalf("labelled commands = %+v, want only tool", labelled)
	}
}

// TestAuditPoliciesForCommandsRecordsUnknownPolicyVerbatim proves a
// command selecting an unknown execution policy records its declared
// identity verbatim rather than collapsing to absence: it is refused at
// admission, not declared-only, and the record must not suggest otherwise.
func TestAuditPoliciesForCommandsRecordsUnknownPolicyVerbatim(t *testing.T) {
	commands := map[string]skillspec.Command{
		"tool": {
			Name: "tool", Type: "script", UnixPath: "scripts/tool",
			ExecutionPolicy: "script-worker-v9", Interpreter: "python3-v1",
		},
	}
	got := AuditPoliciesForCommands(commands)
	if len(got) != 1 || got[0].Command != "tool" || got[0].Policy != "script-worker-v9" {
		t.Fatalf("policy record = %+v, want the verbatim unknown identity", got)
	}
}

// TestAuditLabelsForCommandsIsLexical proves the multi-command report
// lists labelled commands in bytewise lexical order and omits unlabelled
// ones, so the audit gate and the validation output agree on every host.
func TestAuditLabelsForCommandsIsLexical(t *testing.T) {
	commands := map[string]skillspec.Command{
		"zulu":  {Name: "zulu", Type: "script", UnixPath: "scripts/zulu"},
		"alpha": {Name: "alpha", Type: "script", UnixPath: "scripts/alpha"},
		"mike": {
			Name: "mike", Type: "script", UnixPath: "scripts/mike",
			ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
		},
		"sys": {Name: "sys", Type: "system", Command: "git"},
	}
	got := AuditLabelsForCommands(commands, nil)
	if len(got) != 2 || got[0].Command != "alpha" || got[1].Command != "zulu" {
		t.Fatalf("labelled commands = %+v, want alpha then zulu", got)
	}
	for _, entry := range got {
		if len(entry.Labels) != 1 || entry.Labels[0] != LabelDeclaredOnly {
			t.Fatalf("command %q labels = %q, want only %q", entry.Command, entry.Labels, LabelDeclaredOnly)
		}
	}
}
