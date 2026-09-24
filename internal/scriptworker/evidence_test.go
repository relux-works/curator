package scriptworker

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/skillspec"
)

// evidenceMutationCase is one published capability_evidence_cases vector
// case driven at the production session entry: the forge worker proves
// the true identities but attaches the mutated record, so only the
// evidence gate refuses.
type evidenceMutationCase struct {
	// name is the vector case name.
	name string
	// mutation is the forge runtime-entry suffix selecting the lie.
	mutation string
	// code is the expected stable diagnostic.
	code string
	// detail is a fragment the refusal must name, pinning the row to
	// its own validation rule rather than the identity-comparison
	// backstop, which refuses every contradiction with the same code.
	detail string
	// forcePlatform overrides the probed inventory on unix hosts when the
	// host inventory offers no candidate for the mutation. Windows rows
	// always run host-native: unix inventories do not exist there.
	forcePlatform string
	// skipWindows skips the subtest on Windows with the ledger
	// platform-control vocabulary when the Windows inventory offers no
	// candidate.
	skipWindows bool
}

var evidenceMutationCases = []evidenceMutationCase{
	{name: "available-control-reported-unavailable", mutation: "available-status-unavailable", code: CodeCapabilityEvidenceInvalid, detail: "reports status"},
	{name: "unavailable-control-reported-applied", mutation: "unavailable-status-applied", code: CodeCapabilityEvidenceInvalid, detail: "reports status", forcePlatform: ScriptPlatformMacOS},
	{name: "missing-control-entry", mutation: "remove-control", code: CodeCapabilityEvidenceInvalid, detail: "missing exactly one entry"},
	{name: "duplicate-control-entry", mutation: "duplicate-control", code: CodeCapabilityEvidenceInvalid, detail: "duplicates control"},
	{name: "extra-control-entry", mutation: "add-unknown-control", code: CodeCapabilityEvidenceInvalid, detail: "outside script-worker-v1-native-control-inventory-v1"},
	{name: "unknown-record-version", mutation: "record-version-v2", code: CodeCapabilityEvidenceInvalid, detail: "record_version"},
	{name: "host-conditional-status-contradicts-probe", mutation: "probe-unavailable-status-applied", code: CodeCapabilityEvidenceInvalid, detail: "reports status", forcePlatform: ScriptPlatformLinux, skipWindows: true},
	{name: "cached-probe-result", mutation: "probed-at-install", code: CodeCapabilityEvidenceInvalid, detail: "probed_at"},
	{name: "foreign-build-record-version", mutation: "capability-evidence-v1", code: CodeCapabilityEvidenceInvalid, detail: "record_version"},
	{name: "foreign-build-execution-policy", mutation: "manager-worker-v1", code: CodeHardenedClaimForbidden, detail: "is not the enforced script identity"},
	{name: "deferred-script-guarantee-entry", mutation: "script-total-network-denial", code: CodeHardenedClaimForbidden, detail: "deferred guarantee"},
	{name: "deferred-build-guarantee-entry", mutation: "total-network-denial", code: CodeHardenedClaimForbidden, detail: "deferred guarantee"},
}

// TestScriptEvidenceMutationsRefuseBeforePermit drives the twelve forged
// capability_evidence_cases through the production session: the worker
// proves the true identities, the parent validates the mutated record
// against this invocation's own probes before the permit, refuses with the
// vector's diagnostic, and never permits the run.
func TestScriptEvidenceMutationsRefuseBeforePermit(t *testing.T) {
	forge := mustPhysical(t, forgeWorkerBinary(t))
	for _, testCase := range evidenceMutationCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.skipWindows && runtime.GOOS == "windows" {
				t.Skip("host-conditional probe contradiction is a Linux-only control in script-worker-v1-native-control-inventory-v1")
			}
			// Sequential by construction when forcing: the platform
			// override is process-global. Never add t.Parallel here.
			if testCase.forcePlatform != "" && runtime.GOOS != "windows" {
				restore := OverrideInventoryPlatformForTest(testCase.forcePlatform)
				defer restore()
			}
			fixture := newLaunchFixture(t, forge)
			entry := fixture.entry + ".forge-evidence-" + testCase.mutation
			writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
			fixture.request.RuntimeEntry = entry
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_, err := runSession(ctx, fixture.request, nil)
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("%s error = %v, want %s", testCase.name, err, testCase.code)
			}
			if testCase.detail != "" && (err == nil || !strings.Contains(err.Error(), testCase.detail)) {
				t.Fatalf("%s error = %v, want it to name %q", testCase.name, err, testCase.detail)
			}
			waitForForgeRecord(t, entry+".forge-shutdown")
			if _, statErr := os.Stat(entry + ".forge-permit"); statErr == nil {
				t.Fatalf("%s: the parent permitted a forged record", testCase.name)
			}
		})
	}
}

// TestScriptEvidenceOrderMismatchRefuses bounds the parent/worker identity
// comparison: a record whose entries are individually valid but ordered
// differently from the parent's own derivation refuses with
// capability-evidence-invalid and earns no permit. It is a bound, not a
// vector case: it kills the mutant that skips the identity comparison.
func TestScriptEvidenceOrderMismatchRefuses(t *testing.T) {
	forge := mustPhysical(t, forgeWorkerBinary(t))
	fixture := newLaunchFixture(t, forge)
	entry := fixture.entry + ".forge-evidence-reorder-controls"
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	fixture.request.RuntimeEntry = entry
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := runSession(ctx, fixture.request, nil)
	if DiagnosticCode(err) != CodeCapabilityEvidenceInvalid {
		t.Fatalf("reordered-record error = %v, want %s", err, CodeCapabilityEvidenceInvalid)
	}
	if !strings.Contains(err.Error(), "contradicts this invocation's probe") {
		t.Fatalf("reordered-record error = %v, want the identity comparison to refuse", err)
	}
	waitForForgeRecord(t, entry+".forge-shutdown")
	if _, statErr := os.Stat(entry + ".forge-permit"); statErr == nil {
		t.Fatal("the parent permitted a reordered record")
	}
}

// TestFixedUnavailableControlsAreNeverProbed proves fixed-unavailable
// controls are never probed: probe faults injected for every one of them
// never fire, and the invocation succeeds.
func TestFixedUnavailableControlsAreNeverProbed(t *testing.T) {
	// Sequential by construction: the test forces the process-global probe
	// fault. Never add t.Parallel here.
	platform := ScriptInventoryPlatform(runtime.GOOS)
	if platform == ScriptPlatformLinux {
		// The Linux inventory defines no fixed-unavailable control;
		// drive the macOS case through the platform override.
		defer OverrideInventoryPlatformForTest(ScriptPlatformMacOS)()
		platform = ScriptPlatformMacOS
	}
	fixed := map[string]bool{}
	for _, name := range scriptInventoryOrder {
		if scriptInventoryPlatforms[platform][name].Availability == ScriptAvailabilityUnavailable {
			fixed[name] = true
		}
	}
	if len(fixed) == 0 {
		t.Fatalf("no fixed-unavailable control to fault on platform %q", platform)
	}
	defer OverrideScriptProbeFaultForTest(func(control string) error {
		if fixed[control] {
			return errInjectedProbeFailureForTest
		}
		return nil
	})()
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	assertValidScriptEvidence(t, result.Evidence, platform)
}

// TestScriptEvidenceSecondRecordRefuses drives vector case
// `second-record-for-invocation` through the production session: the
// acknowledgement carries the true record so the parent permits, then the
// result carries a second record and the invocation refuses with
// capability-evidence-invalid.
func TestScriptEvidenceSecondRecordRefuses(t *testing.T) {
	forge := mustPhysical(t, forgeWorkerBinary(t))
	fixture := newLaunchFixture(t, forge)
	entry := fixture.entry + ".forge-evidence-second-record"
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	fixture.request.RuntimeEntry = entry
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := runSession(ctx, fixture.request, nil)
	if DiagnosticCode(err) != CodeCapabilityEvidenceInvalid {
		t.Fatalf("second-record error = %v, want %s", err, CodeCapabilityEvidenceInvalid)
	}
	// The permit was earned (the acknowledgement was truthful) and the
	// refusal names the second record, so the row cannot pass on the
	// wrong gate.
	waitForForgeRecord(t, entry+".forge-permit")
	if !strings.Contains(err.Error(), "second") {
		t.Fatalf("second-record error = %v, want it to name the second record", err)
	}
}

// TestScriptEvidenceValidRecordSucceeds drives the valid record shape
// through the production entry on the host inventory: the invocation
// succeeds and carries exactly one closed result-only record with one
// entry per inventory control, probe-consistent statuses, and no
// command output inside it. On Linux the vector case
// `valid-linux-host-conditional-unavailable` is pinned explicitly by
// TestLinuxHostConditionalUnavailableSucceeds.
func TestScriptEvidenceValidRecordSucceeds(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	assertValidScriptEvidence(t, result.Evidence, ScriptInventoryPlatform(runtime.GOOS))
	if len(result.Evidence.Controls) != len(scriptInventoryOrder) {
		t.Fatalf("the record carries %d entries, want exactly one per inventory control", len(result.Evidence.Controls))
	}
	// Closed and result-only: the marshalled record carries exactly the
	// vector's record_fields, each entry exactly the control_entry_fields,
	// and never the command's own output.
	payload, err := json.Marshal(result.Evidence)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"record_version", "execution_policy", "platform", "controls"} {
		if _, ok := decoded[field]; !ok {
			t.Fatalf("the record omits field %q: %s", field, payload)
		}
		delete(decoded, field)
	}
	if len(decoded) != 0 {
		t.Fatalf("the record carries extra fields %v: %s", decoded, payload)
	}
	for _, entry := range result.Evidence.Controls {
		if entry.Name == "" || entry.Availability == "" || entry.Status == "" || entry.ProbedAt == "" {
			t.Fatalf("the record carries an empty entry field: %+v", entry)
		}
		fieldPayload, err := json.Marshal(entry)
		if err != nil {
			t.Fatal(err)
		}
		var fieldDecoded map[string]any
		if err := json.Unmarshal(fieldPayload, &fieldDecoded); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"name", "availability", "status", "probed_at"} {
			if _, ok := fieldDecoded[field]; !ok {
				t.Fatalf("entry %+v omits field %q", entry, field)
			}
			delete(fieldDecoded, field)
		}
		if len(fieldDecoded) != 0 {
			t.Fatalf("entry %+v carries extra fields %v", entry, fieldDecoded)
		}
	}
}

// assertValidScriptEvidence requires the closed record shape and
// probe-consistent statuses for the host platform.
func assertValidScriptEvidence(t *testing.T, record ScriptEvidence, platform string) {
	t.Helper()
	if record.RecordVersion != ScriptEvidenceVersion {
		t.Fatalf("record_version = %q, want %q", record.RecordVersion, ScriptEvidenceVersion)
	}
	if record.ExecutionPolicy != skillspec.ScriptExecutionPolicy {
		t.Fatalf("execution_policy = %q, want %q", record.ExecutionPolicy, skillspec.ScriptExecutionPolicy)
	}
	if record.Platform != platform {
		t.Fatalf("platform = %q, want the probed platform %q", record.Platform, platform)
	}
	seen := map[string]bool{}
	for _, entry := range record.Controls {
		if !inScriptInventory(entry.Name) {
			t.Fatalf("the record names %q outside the inventory", entry.Name)
		}
		if seen[entry.Name] {
			t.Fatalf("the record duplicates %q", entry.Name)
		}
		seen[entry.Name] = true
		if entry.ProbedAt != ScriptProbeTiming {
			t.Fatalf("the record reports probed_at %q, want %q", entry.ProbedAt, ScriptProbeTiming)
		}
		wantAvailability := scriptInventoryPlatforms[platform][entry.Name].Availability
		if entry.Availability != wantAvailability {
			t.Fatalf("control %q reports availability %q, want the probed %q",
				entry.Name, entry.Availability, wantAvailability)
		}
		switch entry.Availability {
		case ScriptAvailabilityAvailable:
			if entry.Status != ScriptStatusApplied {
				t.Fatalf("available control %q reports status %q", entry.Name, entry.Status)
			}
		case ScriptAvailabilityUnavailable:
			if entry.Status != ScriptStatusUnavailable {
				t.Fatalf("unavailable control %q reports status %q", entry.Name, entry.Status)
			}
		case ScriptAvailabilityHostConditional:
			if entry.Status != ScriptStatusApplied && entry.Status != ScriptStatusUnavailable {
				t.Fatalf("host-conditional control %q reports status %q", entry.Name, entry.Status)
			}
		}
	}
	if len(seen) != len(scriptInventoryOrder) {
		t.Fatalf("the record carries %d controls, want exactly one per inventory control", len(seen))
	}
}
