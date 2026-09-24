package scriptpolicy

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

// scriptHostExecutionPolicyVector is the published script-worker-v1 behavioural
// family. Only the sections curator's surface answers are decoded; the rest of
// the file is checked by name in TestScriptHostExecutionPolicySectionsAreAllClassified,
// so a section this build does not read can never arrive unnoticed.
type scriptHostExecutionPolicyVector struct {
	SchemaVersion             int      `json:"schema_version"`
	ProtocolVersion           string   `json:"protocol_version"`
	ExecutionPolicy           string   `json:"execution_policy"`
	Interpreters              []string `json:"interpreters"`
	MandatoryControl          []string `json:"mandatory_controls"`
	CapabilityDerivationCases []struct {
		Name string `json:"name"`
	} `json:"capability_derivation_cases"`
	NativeControlInventory struct {
		Version  string `json:"version"`
		Controls []struct {
			Name string `json:"name"`
		} `json:"controls"`
	} `json:"native_control_inventory"`
	CapabilityEvidenceRecord struct {
		RecordVersion      string   `json:"record_version"`
		InventoryVersion   string   `json:"inventory_version"`
		RecordFields       []string `json:"record_fields"`
		ControlEntryFields []string `json:"control_entry_fields"`
	} `json:"capability_evidence_record"`
	CapabilityEvidenceCases []struct {
		Name string `json:"name"`
	} `json:"capability_evidence_cases"`
	OptInCases []struct {
		Name            string  `json:"name"`
		ManifestSchema  int     `json:"manifest_schema"`
		ExecutionPolicy *string `json:"execution_policy"`
		Interpreter     *string `json:"interpreter"`
		Mode            *string `json:"mode"`
		Accepted        bool    `json:"accepted"`
	} `json:"opt_in_cases"`
	PreflightCases []struct {
		Name                  string  `json:"name"`
		Operation             string  `json:"operation"`
		ExpectedError         *string `json:"expected_error"`
		InvocationSucceeds    bool    `json:"invocation_succeeds"`
		WorkerStarted         bool    `json:"worker_started"`
		MandatoryControlAvail *bool   `json:"mandatory_control_available"`
		InventoryAvailability *string `json:"inventory_availability"`
		ProbeResult           *string `json:"probe_result"`
		EvidenceStatus        *string `json:"evidence_status"`
	} `json:"preflight_cases"`
	AuditLabelCases []struct {
		Name            string   `json:"name"`
		ManifestSchema  int      `json:"manifest_schema"`
		ExecutionPolicy *string  `json:"execution_policy"`
		Labels          []string `json:"labels"`
	} `json:"audit_label_cases"`
}

type platformCaseRow struct {
	Package string
	Test    string
}

type scriptSectionClassification struct {
	Status string
	Rows   []platformCaseRow
}

func productionRow(pkg, test string) platformCaseRow {
	return platformCaseRow{Package: pkg, Test: test}
}

func (row platformCaseRow) key() string {
	return row.Package + "::" + row.Test
}

// Section classifications name the registered production-entry rows that
// consume each published section. The conformance test verifies both the
// vector key set and every row against the CI platform-case ledger.
const (
	consumedAtProductionEntry = "consumed by registered production-entry rows"
)

var scriptHostExecutionPolicySections = map[string]scriptSectionClassification{
	"schema_version": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	}},
	"protocol_version": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/scriptworker", "TestBuiltCuratorWorkerHandshake"),
	}},
	"execution_policy": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/install", "TestEnforcedInstallAndLaunchAtCLIEntry"),
	}},
	"interpreters": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/scriptworker", "TestProductionBinaryLaunchesWhenHostProvides"),
		productionRow("internal/scriptworker", "TestWindowsRealInterpretersRunDeclaredExec"),
	}},
	"opt_in_cases": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	}},
	"mandatory_controls": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/scriptworker", "TestHostProbeReportsClosedInventory"),
		productionRow("internal/install", "TestEnforcedScriptCommandIsRefusedAtInstall"),
		productionRow("internal/scriptworker", "TestPreflightRefusesUnavailableControlAtInvocation"),
	}},
	"preflight_cases": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/install", "TestEnforcedScriptCommandIsRefusedAtInstall"),
		productionRow("internal/scriptworker", "TestPreflightRefusesUnavailableControlAtInvocation"),
		productionRow("internal/scriptworker", "TestLinuxPidsMaxProbeAvailableApplies"),
		productionRow("internal/scriptworker", "TestLinuxPidsMaxProbeUnavailableSucceeds"),
		productionRow("internal/scriptworker", "TestFixedUnavailableControlDoesNotReject"),
	}},
	"audit_label_cases": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/install", "TestScriptAuditLabelsAtInstallEntry"),
		productionRow("internal/install", "TestScriptAuditLabelsAtGlobalInstallEntry"),
		productionRow("cmd/curator", "TestCLIAuditEmitsScriptLabels"),
	}},
	"capability_derivation_cases": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/scriptworker", "TestCapabilityDerivationAllFieldsAbsentDenyByDefault"),
		productionRow("internal/scriptworker", "TestCapabilityDerivationNetworkHostsReportingOnly"),
		productionRow("internal/scriptworker", "TestCapabilityDerivationExecIsManagerResolved"),
		productionRow("internal/scriptworker", "TestCapabilityDerivationSecretsRemainIdentifiers"),
	}},
	"capability_evidence_cases": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/scriptworker", "TestLinuxHostConditionalUnavailableSucceeds"),
		productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
		productionRow("internal/scriptworker", "TestScriptEvidenceSecondRecordRefuses"),
	}},
	"capability_evidence_record": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/scriptworker", "TestScriptEvidenceValidRecordSucceeds"),
		productionRow("internal/scriptworker", "TestRunShimWritesDiagnosticsRecord"),
	}},
	"native_control_inventory": {consumedAtProductionEntry, []platformCaseRow{
		productionRow("internal/scriptworker", "TestHostProbeReportsClosedInventory"),
		productionRow("internal/scriptworker", "TestLinuxHostProbeAndEvidenceAreConsistent"),
		productionRow("internal/scriptworker", "TestMacOSNativeControlsAppliedAtInvocation"),
		productionRow("internal/scriptworker", "TestWindowsJobLimitsAppliedAndConfirmed"),
	}},
}

// Each named vector case maps to a registered production-entry test. The
// case names are derived from the pinned vector in the test below; a new or
// removed vector row therefore changes the measured 33/33 ratio.
var scriptVectorCaseConsumers = map[string]platformCaseRow{
	"opt_in_cases/schema8-explicit-opt-in":                                                      productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	"opt_in_cases/schema8-absent-policy":                                                        productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	"opt_in_cases/legacy-schema7-script":                                                        productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	"opt_in_cases/interpreter-without-policy":                                                   productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	"opt_in_cases/policy-without-interpreter":                                                   productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	"opt_in_cases/unknown-policy":                                                               productionRow("internal/install", "TestScriptOptInCasesAtInstallEntry"),
	"capability_derivation_cases/all-fields-absent-deny-by-default":                             productionRow("internal/scriptworker", "TestCapabilityDerivationAllFieldsAbsentDenyByDefault"),
	"capability_derivation_cases/declared-network-hosts-are-reporting-only":                     productionRow("internal/scriptworker", "TestCapabilityDerivationNetworkHostsReportingOnly"),
	"capability_derivation_cases/declared-exec-is-manager-resolved":                             productionRow("internal/scriptworker", "TestCapabilityDerivationExecIsManagerResolved"),
	"capability_derivation_cases/declared-secrets-remain-identifiers":                           productionRow("internal/scriptworker", "TestCapabilityDerivationSecretsRemainIdentifiers"),
	"capability_evidence_cases/valid-linux-host-conditional-unavailable":                        productionRow("internal/scriptworker", "TestLinuxHostConditionalUnavailableSucceeds"),
	"capability_evidence_cases/available-control-reported-unavailable":                          productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/unavailable-control-reported-applied":                            productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/missing-control-entry":                                           productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/duplicate-control-entry":                                         productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/extra-control-entry":                                             productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/unknown-record-version":                                          productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/host-conditional-status-contradicts-probe":                       productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/cached-probe-result":                                             productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/second-record-for-invocation":                                    productionRow("internal/scriptworker", "TestScriptEvidenceSecondRecordRefuses"),
	"capability_evidence_cases/foreign-build-record-version":                                    productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/foreign-build-execution-policy":                                  productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/deferred-script-guarantee-entry":                                 productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"capability_evidence_cases/deferred-build-guarantee-entry":                                  productionRow("internal/scriptworker", "TestScriptEvidenceMutationsRefuseBeforePermit"),
	"preflight_cases/mandatory-control-unavailable-at-install":                                  productionRow("internal/install", "TestEnforcedScriptCommandIsRefusedAtInstall"),
	"preflight_cases/mandatory-control-unavailable-at-invocation":                               productionRow("internal/scriptworker", "TestPreflightRefusesUnavailableControlAtInvocation"),
	"preflight_cases/linux-pids-max-probe-available-evidence-applied-invocation-succeeds":       productionRow("internal/scriptworker", "TestLinuxPidsMaxProbeAvailableApplies"),
	"preflight_cases/linux-pids-max-probe-unavailable-evidence-unavailable-invocation-succeeds": productionRow("internal/scriptworker", "TestLinuxPidsMaxProbeUnavailableSucceeds"),
	"preflight_cases/fixed-unavailable-control-does-not-reject":                                 productionRow("internal/scriptworker", "TestFixedUnavailableControlDoesNotReject"),
	"audit_label_cases/schema7-script":                                                          productionRow("internal/install", "TestScriptAuditLabelsAtInstallEntry"),
	"audit_label_cases/schema8-declared-only-script":                                            productionRow("internal/install", "TestScriptAuditLabelsAtInstallEntry"),
	"audit_label_cases/schema8-enforced-script":                                                 productionRow("internal/install", "TestScriptAuditLabelsAtInstallEntry"),
	"audit_label_cases/schema8-enforced-unfiltered-network":                                     productionRow("internal/install", "TestScriptAuditLabelsAtInstallEntry"),
}

var scriptMandatoryControlConsumers = map[string][]platformCaseRow{
	"fixed-process-graph":                              {productionRow("internal/scriptworker", "TestRunSessionHappyPath")},
	"worker-identity-verification":                     {productionRow("internal/scriptworker", "TestParentWithholdsPermitOnProofMismatch")},
	"interpreter-resolution-and-identity-verification": {productionRow("internal/scriptworker", "TestProductionBinaryLaunchesWhenHostProvides")},
	"manager-built-environment":                        {productionRow("internal/scriptworker", "TestManagerBuiltEnvironmentWithholdsReserved")},
	"manager-built-path":                               {productionRow("internal/scriptworker", "TestCapabilityDerivationExecIsManagerResolved")},
	"offline-network-configuration":                    {productionRow("internal/scriptworker", "TestOfflineScrubWhenNetworkNone")},
	"operation-private-runtime-area":                   {productionRow("internal/scriptworker", "TestOperationPrivateRuntimeAreaApplied")},
	"explicit-standard-stream-binding":                 {productionRow("internal/scriptworker", "TestScriptWorkerBindsStandardInput")},
	"inventory-controls-applied": {
		productionRow("internal/scriptworker", "TestHostProbeReportsClosedInventory"),
		productionRow("internal/scriptworker", "TestWindowsJobLimitsAppliedAndConfirmed"),
		productionRow("internal/scriptworker", "TestLinuxLandlockConfinementMatchesProbe"),
	},
	"closed-script-capability-evidence-record": {productionRow("internal/scriptworker", "TestScriptEvidenceValidRecordSucceeds")},
	"worker-domain-teardown":                   {productionRow("internal/scriptworker", "TestRunSessionTerminatesDescendants")},
}

// loadScriptHostExecutionPolicyVector reads the published family. The read is
// unguarded on purpose: `.github/ci/root-artifacts.tsv` declares this file for
// this package, so a root that stops publishing it defers the package and is
// fatal in any lane that requires a fully serving root. Skipping here instead
// would let a candidate be qualified against a suite this build never read.
func loadScriptHostExecutionPolicyVector(t *testing.T) (string, []byte) {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "script-host-execution-policy.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	return root, payload
}

// TestScriptHostExecutionPolicySectionsAreAllClassified proves the file this
// build reads is the file the suite publishes. A section the protocol adds --
// a new case family, a new record shape -- fails here on its first run rather
// than being silently ignored by a decoder that only names what it wants.
func TestScriptHostExecutionPolicySectionsAreAllClassified(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 {
		t.Fatal("the script-host-execution-policy family published no sections")
	}
	var unclassified, stale []string
	for section := range raw {
		if _, ok := scriptHostExecutionPolicySections[section]; !ok {
			unclassified = append(unclassified, section)
		}
	}
	for section := range scriptHostExecutionPolicySections {
		if _, ok := raw[section]; !ok {
			stale = append(stale, section)
		}
	}
	sort.Strings(unclassified)
	sort.Strings(stale)
	if len(unclassified) != 0 {
		t.Errorf("the root publishes sections this build does not classify: %s\n"+
			"classify each section and map it to registered production-entry rows",
			strings.Join(unclassified, ", "))
	}
	if len(stale) != 0 {
		t.Errorf("this build classifies sections the root no longer publishes: %s\n"+
			"a classification that names nothing proves nothing; delete it or fix the name",
			strings.Join(stale, ", "))
	}
	ledgerRows := platformCaseLedgerRows(t)
	for section, classification := range scriptHostExecutionPolicySections {
		if classification.Status != consumedAtProductionEntry || len(classification.Rows) == 0 {
			t.Errorf("section %q has no production-entry consumer: %+v", section, classification)
			continue
		}
		for _, row := range classification.Rows {
			if !ledgerRows[row.key()] {
				t.Errorf("section %q names unregistered production row %s", section, row.key())
			}
		}
	}
}

func platformCaseLedgerRows(t *testing.T) map[string]bool {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var ledgerPath string
	for current := directory; ; current = filepath.Dir(current) {
		candidate := filepath.Join(current, ".github", "ci", "platform-cases.tsv")
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			ledgerPath = candidate
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatalf("cannot find .github/ci/platform-cases.tsv from %q", directory)
		}
	}
	payload, err := os.ReadFile(ledgerPath) // #nosec G304 -- repository-owned CI ledger
	if err != nil {
		t.Fatal(err)
	}
	return platformCaseLedgerRowsFromPayload(t, payload)
}

func platformCaseLedgerRowsFromPayload(t *testing.T, payload []byte) map[string]bool {
	t.Helper()
	rows := map[string]bool{}
	for lineNumber, line := range strings.Split(string(payload), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 6 {
			t.Fatalf("platform-case ledger line %d has %d columns, want at least 6", lineNumber+1, len(fields))
		}
		if fields[0] == "package" {
			continue
		}
		rows[fields[0]+"::"+fields[1]] = true
	}
	return rows
}

func TestPlatformCaseLedgerRowsAcceptCRLFBlankLinesAndComments(t *testing.T) {
	ledger := []byte("# ledger comment\r\n\r\n  # indented comment\r\n" +
		"package\tname\tmust_run_on\tskip_allowed_on\tskip_class\treason\r\n" +
		"internal/scriptworker\tTestCRLFRow\tlinux\tdarwin,windows\tplatform-control\tfixture\r\n\r\n")
	rows := platformCaseLedgerRowsFromPayload(t, ledger)
	if len(rows) != 1 || !rows["internal/scriptworker::TestCRLFRow"] {
		t.Fatalf("CRLF ledger rows = %#v, want only internal/scriptworker::TestCRLFRow", rows)
	}
}

// TestScriptHostExecutionPolicyProductionConsumersCoverAllCases binds every
// named behavioral vector case and mandatory control to a real row in the
// platform-case ledger. This is the measured 33/33 case and 11/11 control
// qualification; helper-only conformance rows cannot satisfy the mapping.
func TestScriptHostExecutionPolicyProductionConsumersCoverAllCases(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	ledgerRows := platformCaseLedgerRows(t)
	expectedCases := map[string]bool{}
	addCases := func(family string, names []string) {
		t.Helper()
		for _, name := range names {
			key := family + "/" + name
			if expectedCases[key] {
				t.Fatalf("the suite repeats behavioral case %q", key)
			}
			expectedCases[key] = true
			row, ok := scriptVectorCaseConsumers[key]
			if !ok {
				t.Errorf("vector case %q has no production-entry consumer", key)
				continue
			}
			if !ledgerRows[row.key()] {
				t.Errorf("vector case %q maps to unregistered production row %s", key, row.key())
			}
		}
	}
	caseNames := func(cases []struct {
		Name string `json:"name"`
	}) []string {
		names := make([]string, 0, len(cases))
		for _, testCase := range cases {
			names = append(names, testCase.Name)
		}
		return names
	}
	optInNames := make([]string, 0, len(vector.OptInCases))
	for _, testCase := range vector.OptInCases {
		optInNames = append(optInNames, testCase.Name)
	}
	preflightNames := make([]string, 0, len(vector.PreflightCases))
	for _, testCase := range vector.PreflightCases {
		preflightNames = append(preflightNames, testCase.Name)
	}
	auditNames := make([]string, 0, len(vector.AuditLabelCases))
	for _, testCase := range vector.AuditLabelCases {
		auditNames = append(auditNames, testCase.Name)
	}
	addCases("opt_in_cases", optInNames)
	addCases("capability_derivation_cases", caseNames(vector.CapabilityDerivationCases))
	addCases("capability_evidence_cases", caseNames(vector.CapabilityEvidenceCases))
	addCases("preflight_cases", preflightNames)
	addCases("audit_label_cases", auditNames)
	for key := range scriptVectorCaseConsumers {
		if !expectedCases[key] {
			t.Errorf("production consumer mapping %q names no case in the pinned vector", key)
		}
	}
	if len(expectedCases) != 33 || len(scriptVectorCaseConsumers) != 33 {
		t.Errorf("named behavioral case coverage = %d vector cases / %d mapped rows, want 33/33", len(expectedCases), len(scriptVectorCaseConsumers))
	}

	controls := append([]string(nil), vector.MandatoryControl...)
	if len(controls) != 11 || len(scriptMandatoryControlConsumers) != 11 {
		t.Fatalf("mandatory control coverage = %d vector controls / %d mapped controls, want 11/11", len(controls), len(scriptMandatoryControlConsumers))
	}
	for _, control := range controls {
		rows, ok := scriptMandatoryControlConsumers[control]
		if !ok || len(rows) == 0 {
			t.Errorf("mandatory control %q has no production-entry row", control)
			continue
		}
		for _, row := range rows {
			if !ledgerRows[row.key()] {
				t.Errorf("mandatory control %q maps to unregistered production row %s", control, row.key())
			}
		}
	}
	for control := range scriptMandatoryControlConsumers {
		found := false
		for _, published := range controls {
			if control == published {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mandatory control consumer map names unpublished control %q", control)
		}
	}
	if len(vector.NativeControlInventory.Controls) != 8 {
		t.Errorf("native control inventory has %d controls, want the published eight", len(vector.NativeControlInventory.Controls))
	}
	if vector.CapabilityEvidenceRecord.RecordVersion != "script-capability-evidence-v1" ||
		vector.CapabilityEvidenceRecord.InventoryVersion != vector.NativeControlInventory.Version {
		t.Errorf("evidence record identities do not bind the published inventory: record=%q inventory=%q native=%q",
			vector.CapabilityEvidenceRecord.RecordVersion, vector.CapabilityEvidenceRecord.InventoryVersion, vector.NativeControlInventory.Version)
	}
	wantRecordFields := []string{"record_version", "execution_policy", "platform", "controls"}
	recordFields := append([]string(nil), vector.CapabilityEvidenceRecord.RecordFields...)
	sort.Strings(recordFields)
	sort.Strings(wantRecordFields)
	if strings.Join(recordFields, ",") != strings.Join(wantRecordFields, ",") {
		t.Errorf("evidence record fields = %q, want the fields checked by TestScriptEvidenceValidRecordSucceeds: %q",
			recordFields, wantRecordFields)
	}
	wantControlEntryFields := []string{"name", "availability", "status", "probed_at"}
	controlEntryFields := append([]string(nil), vector.CapabilityEvidenceRecord.ControlEntryFields...)
	sort.Strings(controlEntryFields)
	sort.Strings(wantControlEntryFields)
	if strings.Join(controlEntryFields, ",") != strings.Join(wantControlEntryFields, ",") {
		t.Errorf("evidence control-entry fields = %q, want the fields checked by TestScriptEvidenceValidRecordSucceeds: %q",
			controlEntryFields, wantControlEntryFields)
	}
}

// TestScriptExecutionPolicyIdentityMatchesTheSuite binds the two closed
// identities this build hard-codes to the suite's own bytes. A protocol that
// renamed the policy or widened the interpreter set would otherwise leave
// curator silently accepting the old spelling and rejecting the new one.
func TestScriptExecutionPolicyIdentityMatchesTheSuite(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if vector.SchemaVersion != 1 {
		t.Fatalf("the suite schema_version = %d, want 1", vector.SchemaVersion)
	}
	if vector.ProtocolVersion != "1.0.0-rc.9" {
		t.Fatalf("the pinned suite protocol_version = %q, want the rc.12-pinned script-worker-v1 protocol 1.0.0-rc.9", vector.ProtocolVersion)
	}
	if vector.ExecutionPolicy != skillspec.ScriptExecutionPolicy {
		t.Fatalf("the suite names execution policy %q; this build hard-codes %q",
			vector.ExecutionPolicy, skillspec.ScriptExecutionPolicy)
	}
	if len(vector.Interpreters) == 0 {
		t.Fatal("the suite published no interpreter identities")
	}
	published := append([]string(nil), vector.Interpreters...)
	sort.Strings(published)
	accepted := make([]string, 0, len(skillspec.ScriptInterpreters))
	for name := range skillspec.ScriptInterpreters {
		accepted = append(accepted, name)
	}
	sort.Strings(accepted)
	if strings.Join(published, ",") != strings.Join(accepted, ",") {
		t.Fatalf("the suite publishes interpreters [%s]; this build accepts [%s]",
			strings.Join(published, ", "), strings.Join(accepted, ", "))
	}
}

// TestScriptExecutionOptInCases runs every published opt-in case through the
// two decisions curator owns: whether the manifest parses, and -- when it does
// -- whether the command is enforced or declared-only. An enforced command
// must additionally be admitted by the preflight: the control table is
// complete, so admission proceeds and only a host that cannot provide a
// mandatory control refuses, at the install and invoke entries owned by
// internal/install and internal/scriptworker.
func TestScriptExecutionOptInCases(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.OptInCases) == 0 {
		t.Fatal("the script-host-execution-policy family published no opt-in cases")
	}
	for _, testCase := range vector.OptInCases {
		t.Run(testCase.Name, func(t *testing.T) {
			command := map[string]any{"type": "script", "unix_path": "scripts/tool"}
			if testCase.ExecutionPolicy != nil {
				command["execution_policy"] = *testCase.ExecutionPolicy
			}
			if testCase.Interpreter != nil {
				command["interpreter"] = *testCase.Interpreter
			}
			spec, err := skillspec.Load(materializeScriptSkill(t, testCase.ManifestSchema, command))
			if (err == nil) != testCase.Accepted {
				t.Fatalf("accepted = %v, want %v (error %v)", err == nil, testCase.Accepted, err)
			}
			if !testCase.Accepted {
				// A rejected manifest has no command to classify, and the case
				// records no mode for it.
				if testCase.Mode != nil {
					t.Fatalf("the suite records mode %q for a rejected case", *testCase.Mode)
				}
				return
			}
			if testCase.Mode == nil {
				t.Fatal("the suite accepts this case but records no mode for it")
			}
			parsed, ok := spec.Commands["tool"]
			if !ok {
				t.Fatalf("the parsed manifest carries no command: %+v", spec.Commands)
			}
			wantEnforced := *testCase.Mode == "enforced"
			if Enforced(parsed) != wantEnforced {
				t.Fatalf("Enforced = %v, want %v for mode %q (policy %q)",
					Enforced(parsed), wantEnforced, *testCase.Mode, parsed.ExecutionPolicy)
			}
			admission := Admit(spec.Commands)
			if !wantEnforced {
				if admission != nil {
					t.Fatalf("a declared-only command was refused: %v", admission)
				}
				return
			}
			if admission != nil {
				t.Fatalf("an enforced command was refused at admission: %v", admission)
			}
			if parsed.Interpreter == "" {
				t.Fatal("an enforced command parsed without its bound interpreter")
			}
		})
	}
}

// TestEnforcedShapesProceedPastAdmission proves admission honesty: every
// enforced shape the suite itself declares -- the closed policy identity
// against each published interpreter -- is admitted now that the control
// table is complete, so the probe, derivation, and evidence sections of
// this family are reachable from this build at the production entries.
// An unknown policy keeps the unconditional
// `script_execution_policy_unsupported`.
func TestEnforcedShapesProceedPastAdmission(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.Interpreters) == 0 {
		t.Fatal("the suite published no interpreter identities")
	}
	for _, interpreter := range vector.Interpreters {
		t.Run(interpreter, func(t *testing.T) {
			spec, err := skillspec.Load(materializeScriptSkill(t, 8, map[string]any{
				"type":             "script",
				"unix_path":        "scripts/tool",
				"execution_policy": vector.ExecutionPolicy,
				"interpreter":      interpreter,
			}))
			if err != nil {
				t.Fatalf("the suite's own enforced shape did not parse: %v", err)
			}
			if err := Admit(spec.Commands); err != nil {
				t.Fatalf("an enforced %s command was refused at admission: %v", interpreter, err)
			}
		})
	}
	t.Run("unknown-policy", func(t *testing.T) {
		err := Admit(map[string]skillspec.Command{
			"tool": {Name: "tool", Type: "script", ExecutionPolicy: "script-worker-v9", Interpreter: "python3-v1"},
		})
		if Code(err) != PolicyUnsupported {
			t.Fatalf("an unknown policy was not refused with %s: %v", PolicyUnsupported, err)
		}
	})
}

// TestMandatoryControlsMatchTheSuite binds the 11-control implementation
// table this build preflights to the suite's own bytes, in order. A protocol
// that renamed, reordered, added, or removed a control would otherwise leave
// curator preflighting a different set than the suite names.
func TestMandatoryControlsMatchTheSuite(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.MandatoryControl) == 0 {
		t.Fatal("the suite published no mandatory controls")
	}
	controls := MandatoryControls()
	table := make([]string, 0, len(controls))
	for _, control := range controls {
		table = append(table, control.Name)
	}
	if strings.Join(vector.MandatoryControl, ",") != strings.Join(table, ",") {
		t.Fatalf("the suite publishes mandatory controls [%s]; this build preflights [%s]",
			strings.Join(vector.MandatoryControl, ", "), strings.Join(table, ", "))
	}
	if missing := MissingControls(); len(missing) != 0 {
		t.Fatalf("the preflight reports missing mandatory controls %q; the R3 table is complete", missing)
	}
}

// TestPreflightTableCompleteAdmitsEnforced is the R3 admission-honesty row:
// every mandatory control is implemented, no owner remains, nothing is
// missing, and an enforced command is admitted — only a host that cannot
// provide a mandatory control refuses, at the install and invoke entries.
func TestPreflightTableCompleteAdmitsEnforced(t *testing.T) {
	controls := MandatoryControls()
	if len(controls) != 11 {
		t.Fatalf("the preflight table carries %d controls, want 11", len(controls))
	}
	for _, control := range controls {
		if !control.Implemented {
			t.Fatalf("control %q is not implemented; the R3 table is complete", control.Name)
		}
		if control.Owner != "" {
			t.Fatalf("control %q still names owner %q", control.Name, control.Owner)
		}
	}
	if missing := MissingControls(); len(missing) != 0 {
		t.Fatalf("MissingControls = %q, want empty", missing)
	}
	if err := Admit(map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1"},
	}); err != nil {
		t.Fatalf("an enforced command was refused at admission: %v", err)
	}
}

// TestPreflightTableAccessorIsACopy proves the accessor cannot move the
// preflight: mutating the returned table leaves the enforced table
// unchanged. The R2 injection seam is removed — there is no override to
// set — so this copy property is the whole seam surface.
func TestPreflightTableAccessorIsACopy(t *testing.T) {
	mutated := MandatoryControls()
	for index := range mutated {
		mutated[index].Implemented = false
		mutated[index].Owner = "mutant"
	}
	if missing := MissingControls(); len(missing) != 0 {
		t.Fatalf("mutating the accessor copy moved the preflight table: %q", missing)
	}
	if err := Admit(map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1"},
	}); err != nil {
		t.Fatalf("an enforced command was refused after mutating the accessor copy: %v", err)
	}
}

// TestHostControlErrorNamesUnavailableControls pins the host-preflight
// refusal shape: it carries `script_execution_control_unavailable` and
// names exactly the controls the host cannot provide.
func TestHostControlErrorNamesUnavailableControls(t *testing.T) {
	err := HostControlError("commands.tool.execution_policy",
		[]string{"inventory-controls-applied"}, "native control is unavailable on this host")
	if Code(err) != ControlUnavailable {
		t.Fatalf("code = %q, want %q", Code(err), ControlUnavailable)
	}
	if !strings.Contains(err.Detail, "inventory-controls-applied") {
		t.Fatalf("detail %q does not name the unavailable control", err.Detail)
	}
	if err.Path != "commands.tool.execution_policy" {
		t.Fatalf("path = %q, want the manifest field path", err.Path)
	}
	var refusal *Error
	if !errors.As(error(err), &refusal) {
		t.Fatal("the host refusal is not a *scriptpolicy.Error")
	}
}

// TestPreflightRefusalCases pins the published preflight family's shape:
// exactly the five named cases, with the install/invoke
// mandatory-control-unavailable cases refusing with
// `script_execution_control_unavailable` and starting no worker, and the
// three success cases succeeding. The behavior rows live at the production
// entries: install.Project in internal/install and the native launcher and
// session in internal/scriptworker.
func TestPreflightRefusalCases(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.PreflightCases) == 0 {
		t.Fatal("the suite published no preflight cases")
	}
	byName := map[string]int{}
	for index, testCase := range vector.PreflightCases {
		byName[testCase.Name] = index
	}
	for _, name := range []string{
		"mandatory-control-unavailable-at-install",
		"mandatory-control-unavailable-at-invocation",
		"linux-pids-max-probe-available-evidence-applied-invocation-succeeds",
		"linux-pids-max-probe-unavailable-evidence-unavailable-invocation-succeeds",
		"fixed-unavailable-control-does-not-reject",
	} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("the suite no longer publishes preflight case %q", name)
		}
	}
	if len(vector.PreflightCases) != 5 {
		names := make([]string, 0, len(vector.PreflightCases))
		for _, testCase := range vector.PreflightCases {
			names = append(names, testCase.Name)
		}
		t.Fatalf("the suite publishes preflight cases [%s]; this build classifies exactly five", strings.Join(names, ", "))
	}
	for _, name := range []string{
		"mandatory-control-unavailable-at-install",
		"mandatory-control-unavailable-at-invocation",
	} {
		t.Run(name, func(t *testing.T) {
			testCase := vector.PreflightCases[byName[name]]
			if testCase.ExpectedError == nil || *testCase.ExpectedError != ControlUnavailable {
				t.Fatalf("case %q expects error %v, want %q", name, testCase.ExpectedError, ControlUnavailable)
			}
			if testCase.InvocationSucceeds || testCase.WorkerStarted {
				t.Fatalf("case %q succeeds or starts a worker, want refusal before any worker", name)
			}
			if testCase.MandatoryControlAvail == nil || *testCase.MandatoryControlAvail {
				t.Fatalf("case %q must record mandatory_control_available=false", name)
			}
		})
	}
	for _, name := range []string{
		"linux-pids-max-probe-available-evidence-applied-invocation-succeeds",
		"linux-pids-max-probe-unavailable-evidence-unavailable-invocation-succeeds",
		"fixed-unavailable-control-does-not-reject",
	} {
		t.Run(name, func(t *testing.T) {
			testCase := vector.PreflightCases[byName[name]]
			if testCase.ExpectedError != nil {
				t.Fatalf("case %q expects error %v, want success", name, *testCase.ExpectedError)
			}
			if !testCase.InvocationSucceeds || !testCase.WorkerStarted {
				t.Fatalf("case %q must succeed with a started worker", name)
			}
		})
	}
}

// capabilitiesRequiredFrom is the first manifest schema that requires a
// `capabilities` object. The fixtures below carry an empty one from there on so
// a case fails for its execution-policy rule and not for a missing field.
const capabilitiesRequiredFrom = 3

// materializeScriptSkill writes a one-command skill snapshot at the requested
// manifest schema and lays out the script the command names, so a case fails
// for its execution-policy rule rather than for a missing file.
func materializeScriptSkill(t *testing.T, schema int, command map[string]any) string {
	t.Helper()
	return materializeScriptSkillWithCaps(t, schema, command, map[string]any{})
}

// materializeScriptSkillWithCaps is materializeScriptSkill with an explicit
// declared `capabilities` object, for cases whose labels depend on the
// declaration. A nil object keeps the schema default (empty from schema 3).
func materializeScriptSkillWithCaps(t *testing.T, schema int, command map[string]any, caps map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	if path, ok := command["unix_path"].(string); ok {
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("fixture"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := map[string]any{
		"schema_version": schema,
		"commands":       map[string]any{"tool": command},
	}
	if schema >= capabilitiesRequiredFrom {
		if caps == nil {
			caps = map[string]any{}
		}
		manifest["capabilities"] = caps
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, skillspec.CanonicalManifestName), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestScriptAuditLabelCases pins the published audit-label family and drives
// every case through the label helper against a parsed manifest: schema-7
// and schema-8 declared-only scripts carry `script-command-declared-only`,
// an enforced script with no declared hosts carries nothing, and an
// enforced script with declared hosts carries only
// `script-command-unfiltered-declared-network`. The vector names the policy
// but not the interpreter or the declared hosts, so the enforced rows bind
// the closed `python3-v1` interpreter the suite itself publishes, and the
// unfiltered row declares one host glob while its twin declares none — the
// case name is the distinguishing input, and the suite's two enforced rows
// would be identical manifests without it. The warning rendering lives at
// the production entries: internal/audit, internal/skillcheck,
// internal/install, and cmd/curator. Each case also pins its
// per-command audit-record entry here (identity or explicit absence).
func TestScriptAuditLabelCases(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.AuditLabelCases) == 0 {
		t.Fatal("the suite published no audit-label cases")
	}
	byName := map[string]int{}
	for index, testCase := range vector.AuditLabelCases {
		byName[testCase.Name] = index
	}
	for _, name := range []string{
		"schema7-script",
		"schema8-declared-only-script",
		"schema8-enforced-script",
		"schema8-enforced-unfiltered-network",
	} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("the suite no longer publishes audit-label case %q", name)
		}
	}
	if len(vector.AuditLabelCases) != 4 {
		names := make([]string, 0, len(vector.AuditLabelCases))
		for _, testCase := range vector.AuditLabelCases {
			names = append(names, testCase.Name)
		}
		t.Fatalf("the suite publishes audit-label cases [%s]; this build classifies exactly four", strings.Join(names, ", "))
	}
	for _, testCase := range vector.AuditLabelCases {
		t.Run(testCase.Name, func(t *testing.T) {
			for _, label := range testCase.Labels {
				if label != LabelDeclaredOnly && label != LabelUnfilteredDeclaredNetwork {
					t.Fatalf("case %q names unknown label %q", testCase.Name, label)
				}
			}
			command := map[string]any{"type": "script", "unix_path": "scripts/tool"}
			if testCase.ExecutionPolicy != nil {
				if *testCase.ExecutionPolicy != skillspec.ScriptExecutionPolicy {
					t.Fatalf("case %q selects policy %q, want %q or null",
						testCase.Name, *testCase.ExecutionPolicy, skillspec.ScriptExecutionPolicy)
				}
				command["execution_policy"] = *testCase.ExecutionPolicy
				command["interpreter"] = "python3-v1"
			}
			caps := map[string]any{}
			if testCase.Name == "schema8-enforced-unfiltered-network" {
				caps["network"] = []string{"audit-label.example.com"}
			}
			spec, err := skillspec.Load(materializeScriptSkillWithCaps(t, testCase.ManifestSchema, command, caps))
			if err != nil {
				t.Fatalf("the suite's own audit-label shape did not parse: %v", err)
			}
			labelled := AuditLabelsForCommands(spec.Commands, spec.Capabilities.Network)
			var got []string
			if len(labelled) == 1 && labelled[0].Command == "tool" {
				got = labelled[0].Labels
			} else if len(labelled) != 0 {
				t.Fatalf("labelled commands = %+v, want only tool", labelled)
			}
			if strings.Join(got, ",") != strings.Join(testCase.Labels, ",") {
				t.Fatalf("labels = %q, want the suite's %q", got, testCase.Labels)
			}
			// The audit record carries the case's command whether or not
			// it earns a warning: the exact policy identity for enforced
			// rows, the explicit absence for declared-only rows.
			policies := AuditPoliciesForCommands(spec.Commands)
			if len(policies) != 1 || policies[0].Command != "tool" {
				t.Fatalf("policy record = %+v, want the tool entry", policies)
			}
			var wantPolicy string
			if testCase.ExecutionPolicy != nil {
				wantPolicy = *testCase.ExecutionPolicy
			}
			if policies[0].Policy != wantPolicy {
				t.Fatalf("policy record = %+v, want policy %q", policies, wantPolicy)
			}
			if testCase.ExecutionPolicy != nil {
				for _, label := range got {
					if label == LabelDeclaredOnly {
						t.Fatalf("an enforced command was labelled declared-only")
					}
				}
			}
		})
	}
}
