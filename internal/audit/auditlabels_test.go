package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/skillspec"
)

// scriptSubject lays out a snapshot carrying one benign script and binds the
// given parsed commands and capabilities, so the gate's verdict comes only
// from the script warning classes and never from a detector finding.
func scriptSubject(t *testing.T, commands map[string]skillspec.Command, caps capabilities.Manifest, schema int) Subject {
	t.Helper()
	snapshot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(snapshot, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshot, "scripts", "tool"), []byte("echo ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return Subject{
		Name: "skill-a", Source: "skill-a", Git: "git@git.example.com:skills/skill-a.git",
		Commit: "abc", Snapshot: snapshot, SchemaVersion: schema, Capabilities: caps,
		Commands: commands,
	}
}

func declaredOnlyCommands() map[string]skillspec.Command {
	return map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", UnixPath: "scripts/tool"},
	}
}

func enforcedCommands() map[string]skillspec.Command {
	return map[string]skillspec.Command{
		"tool": {
			Name: "tool", Type: "script", UnixPath: "scripts/tool",
			ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
		},
	}
}

// TestGateEmitsScriptAuditLabels drives the four audit-label shapes through
// the production gate and its read-only twin. The label text is asserted as
// a literal, exactly as the conformance vector names it, so a rename in the
// implementation fails here rather than following a constant.
func TestGateEmitsScriptAuditLabels(t *testing.T) {
	cases := []struct {
		name     string
		commands map[string]skillspec.Command
		caps     capabilities.Manifest
		schema   int
		want     []string
	}{
		{"schema7-script", declaredOnlyCommands(), capabilities.ImplicitNone(), 7,
			[]string{"script-command-declared-only"}},
		{"schema8-declared-only-script", declaredOnlyCommands(), capabilities.ImplicitNone(), 8,
			[]string{"script-command-declared-only"}},
		{"schema8-enforced-script", enforcedCommands(), capabilities.ImplicitNone(), 8, nil},
		{"schema8-enforced-unfiltered-network", enforcedCommands(),
			capabilities.Manifest{Network: []string{"api.example.com"}}, 8,
			[]string{"script-command-unfiltered-declared-network"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			for _, gate := range []struct {
				name string
				call func() ([]string, []string)
			}{
				{"gate", func() ([]string, []string) {
					return Gate(newCfg(t, "advisory", "high"),
						[]Subject{scriptSubject(t, testCase.commands, testCase.caps, testCase.schema)})
				}},
				{"read-only", func() ([]string, []string) {
					return GateReadOnly(newCfg(t, "advisory", "high"),
						[]Subject{scriptSubject(t, testCase.commands, testCase.caps, testCase.schema)})
				}},
			} {
				t.Run(gate.name, func(t *testing.T) {
					warnings, errs := gate.call()
					if len(errs) != 0 {
						t.Fatalf("script labels must never error: errs=%v", errs)
					}
					joined := strings.Join(warnings, "\n")
					for _, label := range testCase.want {
						if !strings.Contains(joined, label) {
							t.Fatalf("warnings do not carry %q: %v", label, warnings)
						}
					}
					if len(testCase.want) == 0 && len(warnings) != 0 {
						t.Fatalf("an enforced script with no declared hosts must be silent: %v", warnings)
					}
					for _, warning := range warnings {
						if !strings.HasPrefix(warning, "audit warning: ") {
							t.Fatalf("script label is not an audit warning: %q", warning)
						}
					}
				})
			}
		})
	}
}

// TestGateScriptLabelsWarnUnderStrictPolicy proves the classes are always
// warnings, never blocks, and never subject to `fail_on`: under the
// strictest policy a labelled subject with no findings still warns.
func TestGateScriptLabelsWarnUnderStrictPolicy(t *testing.T) {
	subjects := []Subject{
		scriptSubject(t, declaredOnlyCommands(), capabilities.ImplicitNone(), 8),
		scriptSubject(t, enforcedCommands(), capabilities.Manifest{Network: []string{"api.example.com"}}, 8),
	}
	warnings, errs := Gate(newCfg(t, "strict", "low"), subjects)
	if len(errs) != 0 {
		t.Fatalf("script labels must never block, even strict/low: errs=%v", errs)
	}
	joined := strings.Join(warnings, "\n")
	for _, label := range []string{"script-command-declared-only", "script-command-unfiltered-declared-network"} {
		if !strings.Contains(joined, label) {
			t.Fatalf("strict warnings do not carry %q: %v", label, warnings)
		}
	}
}

// TestGateScriptLabelsStayWarningsWhenBlocked proves the classes stay
// warnings on the block path: a subject that blocks on a detector finding
// still reports its script label as a warning, never as a second error.
func TestGateScriptLabelsStayWarningsWhenBlocked(t *testing.T) {
	subject := subjectWith(t, "curl https://exfil.example.net/x\n", capabilities.ImplicitNone(), 3)
	subject.Commands = declaredOnlyCommands()
	warnings, errs := Gate(newCfg(t, "strict", "high"), []Subject{subject})
	if len(errs) == 0 || !strings.Contains(strings.Join(errs, "\n"), "network-undeclared") {
		t.Fatalf("the finding must still block: warnings=%v errs=%v", warnings, errs)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "script-command-declared-only") {
		t.Fatalf("the block path dropped the script warning: warnings=%v", warnings)
	}
	for _, err := range errs {
		if strings.Contains(err, "script-command-declared-only") {
			t.Fatalf("the script label escaped into the errors: %v", errs)
		}
	}
}

// TestGateScriptLabelsStayWarningsWhenPinned proves the require-pin path
// keeps the script warning: a pre-capability schema still warns about its
// declared-only command while the pin refusal stays an error.
func TestGateScriptLabelsStayWarningsWhenPinned(t *testing.T) {
	subject := subjectWith(t, "echo ok\n", capabilities.ImplicitNone(), 2)
	subject.Commands = declaredOnlyCommands()
	_, errs := Gate(newCfg(t, "strict", "high"), []Subject{subject})
	if len(errs) != 1 || !strings.Contains(errs[0], "requires pin") {
		t.Fatalf("errs: %v", errs)
	}
	warnings, _ := Gate(newCfg(t, "strict", "high"), []Subject{subject})
	if !strings.Contains(strings.Join(warnings, "\n"), "script-command-declared-only") {
		t.Fatalf("the require-pin path dropped the script warning: warnings=%v", warnings)
	}
}

// TestGateScriptLabelsStayWarningsWhenRevoked proves revocation keeps the
// script warning alongside its unconditional block.
func TestGateScriptLabelsStayWarningsWhenRevoked(t *testing.T) {
	subject := scriptSubject(t, declaredOnlyCommands(), capabilities.ImplicitNone(), 8)
	cfg := newCfg(t, "advisory", "high")
	cfg.Audit.Revocations = []string{"source:git@git.example.com:skills/*"}
	warnings, errs := Gate(cfg, []Subject{subject})
	if len(errs) != 1 || !strings.Contains(errs[0], "revoked") {
		t.Fatalf("revocation must still block: warnings=%v errs=%v", warnings, errs)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "script-command-declared-only") {
		t.Fatalf("the revoked path dropped the script warning: warnings=%v", warnings)
	}
}

// TestGateEnforcedIsNeverDeclaredOnly is the gate-level negative row for
// the declared-only class: enforced commands never render it, with or
// without declared hosts.
func TestGateEnforcedIsNeverDeclaredOnly(t *testing.T) {
	for _, caps := range []capabilities.Manifest{
		capabilities.ImplicitNone(),
		{Network: []string{"api.example.com"}},
	} {
		warnings, errs := Gate(newCfg(t, "advisory", "high"),
			[]Subject{scriptSubject(t, enforcedCommands(), caps, 8)})
		if len(errs) != 0 {
			t.Fatalf("errs=%v", errs)
		}
		for _, warning := range warnings {
			if strings.Contains(warning, "script-command-declared-only") {
				t.Fatalf("an enforced command was labelled declared-only: %v", warnings)
			}
		}
	}
}

// TestGateDeclaredOnlyWithHostsCarriesNoNetworkLabel is the gate-level
// negative row for the network class: a declared-only command with hosts
// renders only declared-only.
func TestGateDeclaredOnlyWithHostsCarriesNoNetworkLabel(t *testing.T) {
	warnings, errs := Gate(newCfg(t, "advisory", "high"), []Subject{
		scriptSubject(t, declaredOnlyCommands(), capabilities.Manifest{Network: []string{"api.example.com"}}, 8),
	})
	if len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "script-command-declared-only") {
		t.Fatalf("warnings=%v", warnings)
	}
	if strings.Contains(joined, "script-command-unfiltered-declared-network") {
		t.Fatalf("a declared-only command earned the network label: %v", warnings)
	}
}

// TestScriptPolicyRecordsCoverEveryCommand proves the production audit
// record carries every script command's execution-policy identity or its
// explicit absence, independent of warning eligibility: the enforced
// no-host command keeps its exact `script-worker-v1` entry while the
// declared-only command records an empty policy, and the mixed order is
// lexical across subjects.
func TestScriptPolicyRecordsCoverEveryCommand(t *testing.T) {
	subjects := []Subject{
		scriptSubject(t, map[string]skillspec.Command{
			"tool": {Name: "tool", Type: "script", UnixPath: "scripts/tool"},
			"guarded": {
				Name: "guarded", Type: "script", UnixPath: "scripts/guarded",
				ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
			},
			"sys": {Name: "sys", Type: "system", Command: "git"},
		}, capabilities.ImplicitNone(), 8),
	}
	got := ScriptPolicyRecords(subjects)
	want := []ScriptPolicyRecord{
		{Skill: "skill-a", Command: "guarded", Policy: "script-worker-v1"},
		{Skill: "skill-a", Command: "tool", Policy: ""},
	}
	if len(got) != len(want) {
		t.Fatalf("policy records = %+v, want %+v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("policy records = %+v, want %+v", got, want)
		}
	}
}

// TestFormatScriptPolicyRendersIdentityOrExplicitAbsence pins the
// terminal rendering: the exact policy identity for enforced commands
// and an explicit `(none)` for declared-only ones, always as `audit
// info`, never as a warning.
func TestFormatScriptPolicyRendersIdentityOrExplicitAbsence(t *testing.T) {
	enforced := FormatScriptPolicy(ScriptPolicyRecord{Skill: "skill-a", Command: "guarded", Policy: "script-worker-v1"})
	if enforced != "audit info: skill-a: command 'guarded' execution_policy=script-worker-v1" {
		t.Fatalf("enforced line = %q", enforced)
	}
	absent := FormatScriptPolicy(ScriptPolicyRecord{Skill: "skill-a", Command: "tool", Policy: ""})
	if absent != "audit info: skill-a: command 'tool' execution_policy=(none)" {
		t.Fatalf("absence line = %q", absent)
	}
	for _, line := range []string{enforced, absent} {
		if strings.Contains(line, "warning") {
			t.Fatalf("the record entry escaped into a warning: %q", line)
		}
	}
}

// TestReportCarriesScriptPoliciesOnEveryPath proves the audit outcome
// carries the per-command record on the fresh path and on the cache-hit
// path alike: the record is computed from the subject, never from the
// cached findings.
func TestReportCarriesScriptPoliciesOnEveryPath(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	commands := map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", UnixPath: "scripts/tool"},
		"guarded": {
			Name: "guarded", Type: "script", UnixPath: "scripts/guarded",
			ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
		},
	}
	subject := scriptSubject(t, commands, capabilities.ImplicitNone(), 8)
	first, err := auditSubject(cfg, subject, true)
	if err != nil {
		t.Fatal(err)
	}
	if first.CacheHit {
		t.Fatal("the first audit must not be a cache hit")
	}
	second, err := auditSubject(cfg, subject, true)
	if err != nil {
		t.Fatal(err)
	}
	if !second.CacheHit {
		t.Fatal("the second audit must be a cache hit")
	}
	for _, report := range []Report{first, second} {
		if len(report.ScriptPolicies) != 2 {
			t.Fatalf("report policies = %+v, want both commands", report.ScriptPolicies)
		}
		byCommand := map[string]string{}
		for _, entry := range report.ScriptPolicies {
			byCommand[entry.Command] = entry.Policy
		}
		if byCommand["guarded"] != "script-worker-v1" {
			t.Fatalf("report policies = %+v, want the exact enforced identity", report.ScriptPolicies)
		}
		if policy, ok := byCommand["tool"]; !ok || policy != "" {
			t.Fatalf("report policies = %+v, want the explicit absence for tool", report.ScriptPolicies)
		}
	}
}

// TestStoredVerdictPersistsScriptPolicies proves the persisted verdict
// file records the per-command entries: the enforced no-host command
// keeps its exact identity and the declared-only command its explicit
// absence.
func TestStoredVerdictPersistsScriptPolicies(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	commands := map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", UnixPath: "scripts/tool"},
		"guarded": {
			Name: "guarded", Type: "script", UnixPath: "scripts/guarded",
			ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
		},
	}
	if _, errs := Gate(cfg, []Subject{scriptSubject(t, commands, capabilities.ImplicitNone(), 8)}); len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	matches, err := filepath.Glob(filepath.Join(cfg.Home(), "audit", "*", "verdict-*.json"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("verdict files = %v, err = %v", matches, err)
	}
	payload, err := os.ReadFile(matches[0]) // #nosec G304 -- test-owned verdict path
	if err != nil {
		t.Fatal(err)
	}
	var verdict struct {
		Skill          string `json:"skill"`
		ScriptPolicies []struct {
			Command string `json:"command"`
			Policy  string `json:"policy"`
		} `json:"script_policies"`
	}
	if err := json.Unmarshal(payload, &verdict); err != nil {
		t.Fatalf("the stored verdict is not machine-readable: %v\n%s", err, payload)
	}
	if len(verdict.ScriptPolicies) != 2 {
		t.Fatalf("stored script_policies = %+v, want both commands", verdict.ScriptPolicies)
	}
	byCommand := map[string]string{}
	for _, entry := range verdict.ScriptPolicies {
		byCommand[entry.Command] = entry.Policy
	}
	if byCommand["guarded"] != "script-worker-v1" {
		t.Fatalf("stored script_policies = %+v, want the exact enforced identity", verdict.ScriptPolicies)
	}
	if policy, ok := byCommand["tool"]; !ok || policy != "" {
		t.Fatalf("stored script_policies = %+v, want the explicit absence for tool", verdict.ScriptPolicies)
	}
}

// TestNoWarningCommandKeepsItsRecordWithoutWarnings proves the record and
// the warnings stay independent: an enforced no-host skill audits with no
// warnings at all while its policy record still names the command and its
// exact identity.
func TestNoWarningCommandKeepsItsRecordWithoutWarnings(t *testing.T) {
	subject := scriptSubject(t, enforcedCommands(), capabilities.ImplicitNone(), 8)
	warnings, errs := Gate(newCfg(t, "advisory", "high"), []Subject{subject})
	if len(errs) != 0 || len(warnings) != 0 {
		t.Fatalf("an enforced no-host skill must audit silently: warnings=%v errs=%v", warnings, errs)
	}
	records := ScriptPolicyRecords([]Subject{subject})
	if len(records) != 1 || records[0].Command != "tool" || records[0].Policy != "script-worker-v1" {
		t.Fatalf("policy records = %+v, want the silent command's exact identity", records)
	}
}

// soleVerdictPath locates the single stored verdict the gate wrote.
func soleVerdictPath(t *testing.T, cfg *config.Config) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(cfg.Home(), "audit", "*", "verdict-*.json"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("verdict files = %v, err = %v", matches, err)
	}
	return matches[0]
}

func readStoredVerdict(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- test-owned verdict path
	if err != nil {
		t.Fatal(err)
	}
	var stored map[string]json.RawMessage
	if err := json.Unmarshal(payload, &stored); err != nil {
		t.Fatal(err)
	}
	return stored
}

func writeStoredVerdict(t *testing.T, path string, stored map[string]json.RawMessage) {
	t.Helper()
	payload, err := json.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func storedPolicyEntries(t *testing.T, stored map[string]json.RawMessage) map[string]string {
	t.Helper()
	raw, ok := stored["script_policies"]
	if !ok {
		t.Fatalf("production Gate left existing verdict without per-command identity: %v", stored)
	}
	var policies []struct {
		Command string `json:"command"`
		Policy  string `json:"policy"`
	}
	if err := json.Unmarshal(raw, &policies); err != nil {
		t.Fatalf("stored script_policies is not machine-readable: %v", err)
	}
	byCommand := map[string]string{}
	for _, entry := range policies {
		byCommand[entry.Command] = entry.Policy
	}
	return byCommand
}

// TestReviewerExistingVerdictGetsPolicyRecord is the revision-2 review
// regression, committed verbatim in flow: a valid pre-R4 cached verdict (no
// `script_policies` member) must gain the per-command identity on the next
// writable production Gate. The enforced no-host command earns no warning,
// so only the stored record proves its identity survived the upgrade.
func TestReviewerExistingVerdictGetsPolicyRecord(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	subject := scriptSubject(t, enforcedCommands(), capabilities.ImplicitNone(), 8)
	if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	path := soleVerdictPath(t, cfg)
	stored := readStoredVerdict(t, path)
	delete(stored, "script_policies") // valid pre-R4 cached verdict
	writeStoredVerdict(t, path, stored)
	if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	byCommand := storedPolicyEntries(t, readStoredVerdict(t, path))
	if len(byCommand) != 1 || byCommand["tool"] != "script-worker-v1" {
		t.Fatalf("stored script_policies = %v, want the silent command's exact identity", byCommand)
	}
}

// TestExistingVerdictGainsMixedPolicyRecord proves the upgrade backfill
// records every script command with the exact entries a fresh audit would
// store, while the cached findings are preserved entry-for-entry.
func TestExistingVerdictGainsMixedPolicyRecord(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	commands := map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", UnixPath: "scripts/tool"},
		"guarded": {
			Name: "guarded", Type: "script", UnixPath: "scripts/guarded",
			ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
		},
	}
	subject := scriptSubject(t, commands, capabilities.ImplicitNone(), 8)
	if err := os.WriteFile(filepath.Join(subject.Snapshot, "scripts", "tool"), []byte("curl https://unlisted.example.net/x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	path := soleVerdictPath(t, cfg)
	before := readStoredVerdict(t, path)
	var cachedFindings []Finding
	if err := json.Unmarshal(before["findings"], &cachedFindings); err != nil || len(cachedFindings) == 0 {
		t.Fatalf("the fixture must cache a detector finding: %v", before)
	}
	delete(before, "script_policies") // valid pre-R4 cached verdict
	writeStoredVerdict(t, path, before)
	warnings, errs := Gate(cfg, []Subject{subject})
	if len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "script-command-declared-only") {
		t.Fatalf("the cache-hit path dropped the script warning: warnings=%v", warnings)
	}
	after := readStoredVerdict(t, path)
	var keptFindings []Finding
	if err := json.Unmarshal(after["findings"], &keptFindings); err != nil {
		t.Fatalf("backfilled verdict findings are not machine-readable: %v", err)
	}
	if !reflect.DeepEqual(keptFindings, cachedFindings) {
		t.Fatalf("backfill rewrote cached findings: before=%+v after=%+v", cachedFindings, keptFindings)
	}
	byCommand := storedPolicyEntries(t, after)
	if len(byCommand) != 2 || byCommand["guarded"] != "script-worker-v1" {
		t.Fatalf("stored script_policies = %v, want both commands with the exact enforced identity", byCommand)
	}
	if policy, ok := byCommand["tool"]; !ok || policy != "" {
		t.Fatalf("stored script_policies = %v, want the explicit absence for tool", byCommand)
	}
}

// TestGateReadOnlyCacheHitReportsPoliciesWithoutWriting proves the no-write
// contract on the cache-hit path: a read-only audit of an old-format verdict
// leaves the stored file byte-identical while the in-memory report still
// carries the per-command policies computed from the current manifest.
func TestGateReadOnlyCacheHitReportsPoliciesWithoutWriting(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	subject := scriptSubject(t, enforcedCommands(), capabilities.ImplicitNone(), 8)
	if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	path := soleVerdictPath(t, cfg)
	stored := readStoredVerdict(t, path)
	delete(stored, "script_policies") // valid pre-R4 cached verdict
	writeStoredVerdict(t, path, stored)
	before, err := os.ReadFile(path) // #nosec G304 -- test-owned verdict path
	if err != nil {
		t.Fatal(err)
	}
	if _, errs := GateReadOnly(cfg, []Subject{subject}); len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	after, err := os.ReadFile(path) // #nosec G304 -- test-owned verdict path
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("the read-only gate rewrote the cached verdict:\nbefore=%s\nafter=%s", before, after)
	}
	report, err := auditSubject(cfg, subject, false)
	if err != nil {
		t.Fatal(err)
	}
	if !report.CacheHit {
		t.Fatal("the read-only audit must still be a cache hit")
	}
	if len(report.ScriptPolicies) != 1 || report.ScriptPolicies[0].Command != "tool" || report.ScriptPolicies[0].Policy != "script-worker-v1" {
		t.Fatalf("read-only report policies = %+v, want the current manifest's exact identity", report.ScriptPolicies)
	}
}

// TestCacheHitWithoutCommandsKeepsStoredPolicies proves a writable audit of
// a subject without parsed commands (the source-audit path carries none)
// never erases a stored policy record: verdicts that already record
// policies are left byte-identical.
func TestCacheHitWithoutCommandsKeepsStoredPolicies(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	subject := scriptSubject(t, enforcedCommands(), capabilities.ImplicitNone(), 8)
	if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	path := soleVerdictPath(t, cfg)
	before, err := os.ReadFile(path) // #nosec G304 -- test-owned verdict path
	if err != nil {
		t.Fatal(err)
	}
	bare := subject
	bare.Commands = nil
	if _, err := auditSubject(cfg, bare, true); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path) // #nosec G304 -- test-owned verdict path
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("a commands-less audit rewrote the stored record:\nbefore=%s\nafter=%s", before, after)
	}
	byCommand := storedPolicyEntries(t, readStoredVerdict(t, path))
	if len(byCommand) != 1 || byCommand["tool"] != "script-worker-v1" {
		t.Fatalf("stored script_policies = %v, want the untouched exact identity", byCommand)
	}
}

// TestGateReadOnlyEmitsScriptLabelsWithoutWritingState proves the dry-run
// gate renders the classes while leaving verdict and trust state alone.
func TestGateReadOnlyEmitsScriptLabelsWithoutWritingState(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	warnings, errs := GateReadOnly(cfg, []Subject{
		scriptSubject(t, declaredOnlyCommands(), capabilities.ImplicitNone(), 8),
	})
	if len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "script-command-declared-only") {
		t.Fatalf("warnings=%v", warnings)
	}
	if _, err := os.Stat(filepath.Join(cfg.Home(), "audit")); !os.IsNotExist(err) {
		t.Fatalf("read-only gate wrote audit state: %v", err)
	}
}
