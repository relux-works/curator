package godriver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
)

// hostExecutionVector is the accepted rc.5 authority for the portable
// manager-worker-v1 execution policy.
type hostExecutionVector struct {
	ProtocolVersion                 string   `json:"protocol_version"`
	ExecutionPolicy                 string   `json:"execution_policy"`
	ReservedHardenedExecutionPolicy string   `json:"reserved_hardened_execution_policy"`
	HardenedProfileOwner            string   `json:"hardened_profile_owner"`
	Drivers                         []string `json:"drivers"`
	ProcessGraph                    []string `json:"process_graph"`
	SessionStates                   []string `json:"session_states"`

	MandatoryControls []struct {
		Name              string `json:"name"`
		Portable          bool   `json:"portable"`
		HardenedGuarantee bool   `json:"hardened_guarantee"`
		Enforced          string `json:"enforced"`
	} `json:"mandatory_controls"`

	NativeControlInventory struct {
		Version           string   `json:"version"`
		Exhaustive        bool     `json:"exhaustive"`
		Platforms         []string `json:"platforms"`
		ProbeTiming       string   `json:"probe_timing"`
		ProbeScope        string   `json:"probe_scope"`
		AvailabilityState []string `json:"availability_states"`
		UnavailableReason []string `json:"unavailable_reasons"`
		Controls          []struct {
			Name                 string `json:"name"`
			AppliedWhenAvailable bool   `json:"applied_when_available"`
			HardenedGuarantee    bool   `json:"hardened_guarantee"`
			Platforms            map[string]struct {
				Availability      string `json:"availability"`
				Mechanism         string `json:"mechanism"`
				UnavailableReason string `json:"unavailable_reason"`
			} `json:"platforms"`
		} `json:"controls"`
	} `json:"native_control_inventory"`

	CapabilityEvidenceRecord struct {
		RecordVersion      string   `json:"record_version"`
		InventoryVersion   string   `json:"inventory_version"`
		RecordFields       []string `json:"record_fields"`
		ControlEntryFields []string `json:"control_entry_fields"`
		AvailabilityStates []string `json:"availability_states"`
		StatusStates       []string `json:"status_states"`
		ProbeTimings       []string `json:"probe_timings"`
		EntryCardinality   string   `json:"entry_cardinality"`
		ResultOnly         bool     `json:"result_only"`
		ExcludedFrom       []string `json:"excluded_from"`
		ExposedIn          []string `json:"exposed_in"`
		Examples           map[string]struct {
			RecordVersion   string `json:"record_version"`
			ExecutionPolicy string `json:"execution_policy"`
			Platform        string `json:"platform"`
			Controls        []struct {
				Name         string `json:"name"`
				Availability string `json:"availability"`
				Status       string `json:"status"`
				ProbedAt     string `json:"probed_at"`
			} `json:"controls"`
		} `json:"examples"`
	} `json:"capability_evidence_record"`

	CapabilityEvidenceCases []struct {
		Name                   string `json:"name"`
		Control                string `json:"control"`
		Availability           string `json:"availability"`
		Status                 string `json:"status"`
		EntryCount             int    `json:"entry_count"`
		InInventory            bool   `json:"in_inventory"`
		RecordVersion          string `json:"record_version"`
		RecordExecutionPolicy  string `json:"record_execution_policy"`
		HardenedGuaranteeClaim bool   `json:"hardened_guarantee_claimed"`
		RecordValid            bool   `json:"record_valid"`
		BuildPermitted         bool   `json:"build_permitted"`
		ChangesCacheKey        bool   `json:"changes_cache_key"`
		ExpectedError          string `json:"expected_error"`
	} `json:"capability_evidence_cases"`

	DeferredHardenedGuarantees []struct {
		Name                  string `json:"name"`
		DeferredTo            string `json:"deferred_to"`
		PortableProfileClaims bool   `json:"portable_profile_claims"`
		RejectsPortableBuild  bool   `json:"rejects_portable_build"`
	} `json:"deferred_hardened_guarantees"`

	DeferredCapabilityRejectionGuards []struct {
		Name                     string `json:"name"`
		InMandatoryControls      bool   `json:"in_mandatory_controls"`
		InNativeControlInventory bool   `json:"in_native_control_inventory"`
		InCapabilityEvidence     bool   `json:"in_capability_evidence_record"`
		PortableRejectionCode    string `json:"portable_rejection_code"`
		BuildPermittedWhenAbsent bool   `json:"build_permitted_when_absent"`
	} `json:"deferred_capability_rejection_guards"`

	FailureBoundary map[string]struct {
		ExpectedError string `json:"expected_error"`
		FailsBefore   string `json:"fails_before"`
		Published     bool   `json:"published"`
		RejectsBuild  bool   `json:"rejects_build"`
	} `json:"failure_boundary"`

	IdentityAndProtocolCases []struct {
		Name            string `json:"name"`
		ExpectedError   string `json:"expected_error"`
		WorkerStarted   bool   `json:"worker_started"`
		CompilerStarted bool   `json:"compiler_started"`
		Published       bool   `json:"published"`
	} `json:"identity_and_protocol_cases"`

	PackageInfluenceCases []struct {
		Name            string `json:"name"`
		ExpectedError   string `json:"expected_error"`
		WorkerStarted   bool   `json:"worker_started"`
		CompilerStarted bool   `json:"compiler_started"`
		Published       bool   `json:"published"`
	} `json:"package_influence_cases"`

	CacheIdentity struct {
		Aliases  bool `json:"aliases"`
		Portable struct {
			CacheKey        string `json:"cache_key"`
			ExecutionPolicy string `json:"execution_policy"`
			SchemaValid     bool   `json:"schema_valid"`
		} `json:"portable"`
		ReservedHardened struct {
			CacheKey        string `json:"cache_key"`
			ExecutionPolicy string `json:"execution_policy"`
			SchemaValid     bool   `json:"schema_valid"`
		} `json:"reserved_hardened"`
		Legacy struct {
			CacheKey    string `json:"cache_key"`
			SchemaValid bool   `json:"schema_valid"`
		} `json:"legacy_rc4_without_execution_policy"`
	} `json:"cache_identity"`
}

func loadHostExecutionVector(t *testing.T) hostExecutionVector {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "go-host-execution-policy.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s is a pre-revision root and publishes no portable execution-policy vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vector hostExecutionVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	return vector
}

func TestPortableExecutionPolicyMatchesTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	if vector.ExecutionPolicy != ExecutionPolicy {
		t.Fatalf("vector execution policy = %q, implementation = %q", vector.ExecutionPolicy, ExecutionPolicy)
	}
	if vector.ReservedHardenedExecutionPolicy != buildmeta.ReservedHardenedExecutionPolicy {
		t.Fatalf("vector reserved hardened policy = %q", vector.ReservedHardenedExecutionPolicy)
	}
	wantGraph := []string{
		"manager-parent", "identity-verified-manager-owned-worker",
		"fingerprinted-goroot-bin-go", "fingerprinted-goroot-pkg-tool-child",
	}
	if !reflect.DeepEqual(vector.ProcessGraph, wantGraph) {
		t.Fatalf("vector process graph = %q", vector.ProcessGraph)
	}
	wantStates := []string{
		"parent-package-independent-toolchain-probe",
		"parent-native-control-availability-probe",
		"parent-worker-identity-verification",
		"worker-launch",
		"worker-identity-proof-and-nonce-acknowledgement",
		"worker-control-application-and-evidence",
		"worker-fixed-go-list",
		"parent-complete-package-graph-validation",
		"parent-authenticated-build-permit",
		"worker-fixed-go-build",
		"parent-artifact-verification",
		"parent-post-exec-identity-reverification",
		"worker-domain-teardown",
	}
	if !reflect.DeepEqual(vector.SessionStates, wantStates) {
		t.Fatalf("vector session states = %q", vector.SessionStates)
	}
}

func TestMandatoryControlsMatchTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	names := make([]string, 0, len(vector.MandatoryControls))
	for _, control := range vector.MandatoryControls {
		if !control.Portable || control.HardenedGuarantee || control.Enforced != "always" {
			t.Fatalf("mandatory control %+v is not an always-enforced portable control", control)
		}
		names = append(names, control.Name)
	}
	if !reflect.DeepEqual(names, MandatoryControls()) {
		t.Fatalf("vector mandatory controls = %q\nimplementation = %q", names, MandatoryControls())
	}
}

func TestNativeControlInventoryMatchesTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	inventory := vector.NativeControlInventory
	if inventory.Version != NativeControlInventoryVersion || !inventory.Exhaustive {
		t.Fatalf("inventory version/exhaustive = %q/%v", inventory.Version, inventory.Exhaustive)
	}
	if inventory.ProbeTiming != ProbeTiming || inventory.ProbeScope != "per-operation" {
		t.Fatalf("inventory probe timing/scope = %q/%q", inventory.ProbeTiming, inventory.ProbeScope)
	}
	if !reflect.DeepEqual(inventory.AvailabilityState, []string{AvailabilityAvailable, AvailabilityUnavailable}) {
		t.Fatalf("inventory availability states = %q", inventory.AvailabilityState)
	}
	if !reflect.DeepEqual(inventory.UnavailableReason, []string{UnavailableReasonNoPrivateAggregateDomain}) {
		t.Fatalf("inventory unavailable reasons = %q", inventory.UnavailableReason)
	}
	if !reflect.DeepEqual(inventory.Platforms, []string{PlatformMacOS, PlatformWindows}) {
		t.Fatalf("inventory platforms = %q", inventory.Platforms)
	}
	names := make([]string, 0, len(inventory.Controls))
	for _, control := range inventory.Controls {
		if control.HardenedGuarantee || !control.AppliedWhenAvailable {
			t.Fatalf("inventory control %+v", control)
		}
		names = append(names, control.Name)
		if len(control.Platforms) != 2 {
			t.Fatalf("control %q covers %d platforms", control.Name, len(control.Platforms))
		}
		for platform, record := range control.Platforms {
			got := nativeControlPlatforms[platform][control.Name]
			if got.Availability != record.Availability || got.Mechanism != record.Mechanism || got.UnavailableReason != record.UnavailableReason {
				t.Fatalf("%s/%s implementation record = %+v, vector = %+v", platform, control.Name, got, record)
			}
		}
	}
	if !reflect.DeepEqual(names, NativeControlInventory()) {
		t.Fatalf("vector inventory = %q\nimplementation = %q", names, NativeControlInventory())
	}
}

func TestCapabilityEvidenceRecordMatchesTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	record := vector.CapabilityEvidenceRecord
	if record.RecordVersion != CapabilityEvidenceVersion || record.InventoryVersion != NativeControlInventoryVersion {
		t.Fatalf("record identity = %q/%q", record.RecordVersion, record.InventoryVersion)
	}
	if !record.ResultOnly || record.EntryCardinality != "exactly-one-per-inventory-control" {
		t.Fatalf("record cardinality/result-only = %q/%v", record.EntryCardinality, record.ResultOnly)
	}
	if !reflect.DeepEqual(record.RecordFields, []string{"controls", "execution_policy", "platform", "record_version"}) {
		t.Fatalf("record fields = %q", record.RecordFields)
	}
	if !reflect.DeepEqual(record.ControlEntryFields, []string{"availability", "name", "probed_at", "status"}) {
		t.Fatalf("entry fields = %q", record.ControlEntryFields)
	}
	if !reflect.DeepEqual(record.ProbeTimings, []string{ProbeTiming}) {
		t.Fatalf("probe timings = %q", record.ProbeTimings)
	}
	if !reflect.DeepEqual(record.StatusStates, []string{StatusApplied, StatusUnavailable}) {
		t.Fatalf("status states = %q", record.StatusStates)
	}
	if !reflect.DeepEqual(record.ExcludedFrom, []string{"cache-key", "conformance-claim", "install-marker", "receipt"}) {
		t.Fatalf("excluded from = %q", record.ExcludedFrom)
	}
	if !reflect.DeepEqual(record.ExposedIn, []string{"dry-run-plan-result", "install-result", "status-result"}) {
		t.Fatalf("exposed in = %q", record.ExposedIn)
	}
	for platform, example := range record.Examples {
		want := validEvidence(platform)
		if example.RecordVersion != want.RecordVersion || example.ExecutionPolicy != want.ExecutionPolicy || example.Platform != want.Platform {
			t.Fatalf("%s example identity = %+v", platform, example)
		}
		if len(example.Controls) != len(want.Controls) {
			t.Fatalf("%s example has %d entries", platform, len(example.Controls))
		}
		for index, entry := range example.Controls {
			got := want.Controls[index]
			if entry.Name != got.Name || entry.Availability != got.Availability || entry.Status != got.Status || entry.ProbedAt != got.ProbedAt {
				t.Fatalf("%s example entry %d = %+v, implementation = %+v", platform, index, entry, got)
			}
		}
		if err := validateCapabilityEvidence(want, platform, syntheticProbes(platform)); err != nil {
			t.Fatalf("%s example rejected by the implementation: %v", platform, err)
		}
	}
}

func TestCapabilityEvidenceCasesMatchTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	if len(vector.CapabilityEvidenceCases) != 11 {
		t.Fatalf("evidence cases = %d, want the exact rc.5 inventory of 11", len(vector.CapabilityEvidenceCases))
	}
	for _, testCase := range vector.CapabilityEvidenceCases {
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.ChangesCacheKey {
				t.Fatal("capability evidence must never change a cache key")
			}
			if testCase.BuildPermitted != (testCase.ExpectedError == "") || testCase.RecordValid != (testCase.ExpectedError == "") {
				t.Fatalf("case %+v is internally inconsistent", testCase)
			}
			if testCase.ExpectedError == CodeControlUnavailable {
				t.Fatal("a reporting fault must never become the mandatory-control rejection")
			}
			platform := PlatformMacOS
			record := validEvidence(platform)
			record.RecordVersion = testCase.RecordVersion
			record.ExecutionPolicy = testCase.RecordExecutionPolicy
			switch {
			case testCase.EntryCount == 0:
				kept := record.Controls[:0]
				for _, entry := range record.Controls {
					if entry.Name != testCase.Control {
						kept = append(kept, entry)
					}
				}
				record.Controls = kept
			case !testCase.InInventory:
				record.Controls[0].Name = testCase.Control
				record.Controls[0].Availability = testCase.Availability
				record.Controls[0].Status = testCase.Status
			default:
				for index := range record.Controls {
					if record.Controls[index].Name == testCase.Control {
						record.Controls[index].Availability = testCase.Availability
						record.Controls[index].Status = testCase.Status
						if testCase.EntryCount == 2 {
							record.Controls = append(record.Controls, record.Controls[index])
						}
						break
					}
				}
			}
			err := validateCapabilityEvidence(record, platform, syntheticProbes(platform))
			if DiagnosticCode(err) != testCase.ExpectedError {
				t.Fatalf("error = %v, want %q", err, testCase.ExpectedError)
			}
			if testCase.HardenedGuaranteeClaim && DiagnosticCode(err) != CodeHardenedClaimForbidden {
				t.Fatalf("hardened claim %q was not rejected as a hardened claim", testCase.Name)
			}
		})
	}
}

func TestDeferredGuaranteesAndFailureBoundaryMatchTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	names := make([]string, 0, len(vector.DeferredHardenedGuarantees))
	for _, guarantee := range vector.DeferredHardenedGuarantees {
		if guarantee.PortableProfileClaims || guarantee.RejectsPortableBuild {
			t.Fatalf("deferred guarantee %+v", guarantee)
		}
		if guarantee.DeferredTo != vector.HardenedProfileOwner {
			t.Fatalf("guarantee %q is deferred to %q", guarantee.Name, guarantee.DeferredTo)
		}
		names = append(names, guarantee.Name)
	}
	if !reflect.DeepEqual(names, DeferredHardenedGuarantees()) {
		t.Fatalf("vector deferred guarantees = %q\nimplementation = %q", names, DeferredHardenedGuarantees())
	}
	for _, guard := range vector.DeferredCapabilityRejectionGuards {
		if guard.InMandatoryControls || guard.InNativeControlInventory || guard.InCapabilityEvidence {
			t.Fatalf("rejection guard %+v", guard)
		}
		if guard.PortableRejectionCode != "" || !guard.BuildPermittedWhenAbsent {
			t.Fatalf("rejection guard %+v carries a portable rejection", guard)
		}
		if inInventory(guard.Name) || !isDeferredHardenedGuarantee(guard.Name) {
			t.Fatalf("rejection guard %q is not a deferred hardened guarantee", guard.Name)
		}
	}
	boundary := vector.FailureBoundary
	if len(boundary) != 3 {
		t.Fatalf("failure boundary has %d keys, want exactly three", len(boundary))
	}
	mandatory := boundary["missing_mandatory_portable_control"]
	if mandatory.ExpectedError != CodeControlUnavailable || mandatory.FailsBefore != "worker-launch" || mandatory.Published || !mandatory.RejectsBuild {
		t.Fatalf("mandatory-control boundary = %+v", mandatory)
	}
	for _, key := range []string{"unavailable_inventory_native_control", "missing_deferred_hardened_capability"} {
		entry := boundary[key]
		if entry.ExpectedError != "" || entry.FailsBefore != "" || !entry.Published || entry.RejectsBuild {
			t.Fatalf("%s boundary = %+v", key, entry)
		}
	}
}

func TestIdentityProtocolAndPackageInfluenceCodesMatchTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	implemented := map[string]bool{
		CodeWorkerIdentityInvalid: true, CodeWorkerProtocolInvalid: true,
		CodeControlUnavailable: true, CodeCapabilityEvidenceInvalid: true,
		CodeHardenedClaimForbidden: true, CodePackageInfluenceForbidden: true,
	}
	for _, testCase := range vector.IdentityAndProtocolCases {
		if !implemented[testCase.ExpectedError] {
			t.Fatalf("case %q expects unimplemented diagnostic %q", testCase.Name, testCase.ExpectedError)
		}
		if testCase.Published {
			t.Fatalf("case %q publishes", testCase.Name)
		}
	}
	if len(vector.PackageInfluenceCases) != 8 {
		t.Fatalf("package influence cases = %d, want eight", len(vector.PackageInfluenceCases))
	}
	for _, testCase := range vector.PackageInfluenceCases {
		if testCase.ExpectedError != CodePackageInfluenceForbidden || testCase.WorkerStarted || testCase.CompilerStarted || testCase.Published {
			t.Fatalf("package influence case %+v", testCase)
		}
	}
}

func TestCacheIdentityMatchesTheAcceptedVector(t *testing.T) {
	vector := loadHostExecutionVector(t)
	if vector.CacheIdentity.Aliases {
		t.Fatal("vector declares aliasing cache identities")
	}
	portable := vector.CacheIdentity.Portable
	if portable.ExecutionPolicy != ExecutionPolicy || !portable.SchemaValid {
		t.Fatalf("portable entry = %+v", portable)
	}
	key, err := portableVectorInput().CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	if string(key) != portable.CacheKey {
		t.Fatalf("portable cache key = %s, vector = %s", key, portable.CacheKey)
	}
	if vector.CacheIdentity.ReservedHardened.SchemaValid || vector.CacheIdentity.Legacy.SchemaValid {
		t.Fatal("the reserved hardened and pre-revision entries must not be schema-valid for go-v1")
	}
	for _, other := range []string{vector.CacheIdentity.ReservedHardened.CacheKey, vector.CacheIdentity.Legacy.CacheKey} {
		if other == portable.CacheKey {
			t.Fatalf("cache identity %s aliases the portable key", other)
		}
	}
}

// TestCandidateGoV1SourceAwareContract consumes the source-aware driver vector
// when the selected conformance root publishes one. The accepted rc.5 candidate
// does not, so the check runs only against a root that does.
func TestCandidateGoV1SourceAwareContract(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "build-drivers.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-drivers vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Driver           string            `json:"driver"`
		FixedEnvironment map[string]string `json:"fixed_environment"`
		Argv             []struct {
			Name string   `json:"name"`
			Argv []string `json:"argv"`
		} `json:"argv"`
		Rejections []struct {
			Name     string `json:"name"`
			Expected struct {
				Error string `json:"error"`
			} `json:"expected"`
		} `json:"rejection_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if vectors.Driver != buildmeta.DriverGoV1 || len(vectors.Argv) != 5 {
		t.Fatalf("candidate driver/argv = %q/%d", vectors.Driver, len(vectors.Argv))
	}
	if !reflect.DeepEqual(vectors.Argv[3].Argv[1:], listArguments) {
		t.Fatalf("candidate list argv = %q, implementation = %q", vectors.Argv[3].Argv[1:], listArguments)
	}
	if len(vectors.Argv[4].Argv) != 12 || !reflect.DeepEqual(vectors.Argv[4].Argv[1:10], buildArgumentPrefix) || vectors.Argv[4].Argv[11] != "." {
		t.Fatalf("candidate build argv = %q", vectors.Argv[4].Argv)
	}
	for key, want := range map[string]string{
		"GOENV": "off", "GOTOOLCHAIN": "local", "GOFLAGS": "", "GOPROXY": "off", "GOSUMDB": "off",
		"GOVCS": "*:off", "GOWORK": "off", "CGO_ENABLED": "0", "GO_EXTLINK_ENABLED": "0", "GOEXPERIMENT": "",
	} {
		if vectors.FixedEnvironment[key] != want {
			t.Fatalf("candidate fixed environment %s = %q, want %q", key, vectors.FixedEnvironment[key], want)
		}
	}
	wantRejections := map[string]string{
		"non-main-package": "build_package_not_main", "multiple-packages": "build_package_ambiguous",
		"missing-vendored-dependency": "vendor_dependency_missing", "inconsistent-vendor-modules": "vendor_metadata_inconsistent",
		"workspace-only-dependency": "workspace_dependency_forbidden", "toolchain-switch-request": "toolchain_switch_forbidden",
		"cgo-only-package": "cgo_required", "native-c-input": "go_native_input_forbidden", "native-cxx-input": "go_native_input_forbidden",
		"native-swig-input": "go_native_input_forbidden", "root-syso": "go_syso_forbidden", "transitive-syso": "go_syso_forbidden",
		"escaped-embed-input": "go_embed_input_escape", "cgo-import-dynamic": "go_forbidden_compiler_directive",
		"attempted-go-generate": "go_generator_forbidden", "default-pgo": "go_pgo_forbidden",
		"external-link-required": "external_link_forbidden", "libgcc-fallback-attempt": "libgcc_fallback_forbidden",
	}
	seen := make(map[string]string, len(vectors.Rejections))
	for _, rejection := range vectors.Rejections {
		seen[rejection.Name] = rejection.Expected.Error
	}
	for name, code := range wantRejections {
		if seen[name] != code {
			t.Fatalf("candidate rejection %s = %q, want %q", name, seen[name], code)
		}
	}
}
